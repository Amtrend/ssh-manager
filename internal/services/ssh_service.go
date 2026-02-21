package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"ssh_manager/internal/encryption"
	"ssh_manager/internal/models"
	"ssh_manager/internal/repository"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHService manages the lifecycle of active SSH connections.
type SSHService struct {
	HostRepo        *repository.HostRepository
	KeyRepo         *repository.KeyRepository
	Sessions        map[int]map[int]*models.ActiveSession // [userID][hostID]
	Mu              sync.RWMutex
	CleanupInterval time.Duration
	SessionTimeout  time.Duration
}

// NewSSHService creates a new instance of SSHService and starts it.
func NewSSHService(hRepo *repository.HostRepository, kRepo *repository.KeyRepository, cleanupInterval, sessionTimeout time.Duration) *SSHService {
	s := &SSHService{
		HostRepo:        hRepo,
		KeyRepo:         kRepo,
		Sessions:        make(map[int]map[int]*models.ActiveSession),
		CleanupInterval: cleanupInterval,
		SessionTimeout:  sessionTimeout,
	}
	go s.startCleaner()
	return s
}

// GetSession searches for an existing session or creates a new one.
func (s *SSHService) GetSession(userID, hostID int, ctx context.Context) (*models.ActiveSession, error) {
	s.Mu.Lock()
	if s.Sessions[userID] == nil {
		s.Sessions[userID] = make(map[int]*models.ActiveSession)
	}

	as, exists := s.Sessions[userID][hostID]
	s.Mu.Unlock()

	if exists {
		return as, nil
	}

	// If there is no session, create a new one
	return s.initNewSession(userID, hostID, ctx)
}

// initNewSession creates a new ssh session.
func (s *SSHService) initNewSession(userID, hostID int, ctx context.Context) (*models.ActiveSession, error) {
	host, err := s.HostRepo.GetByID(ctx, hostID, userID)
	if err != nil {
		return nil, err
	}

	var authMethods []ssh.AuthMethod

	if host.AuthType == "password" {
		if host.Password == "" {
			return nil, err
		}

		decryptedPassword, err := encryption.Decrypt(host.Password)
		if err != nil {
			return nil, err
		}
		authMethods = append(authMethods, ssh.Password(decryptedPassword))
	} else {
		if host.KeyID == nil {
			return nil, err
		}

		key, err := s.KeyRepo.GetByID(ctx, *host.KeyID, userID)
		if err != nil {
			return nil, err
		}

		decryptedKey, err := encryption.Decrypt(key.KeyData)
		if err != nil {
			return nil, err
		}

		signer, err := ssh.ParsePrivateKey([]byte(decryptedKey))
		if err != nil {
			return nil, err
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.Address, host.Port), &ssh.ClientConfig{
		User:            host.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("Connection failed: %v", err)
	}

	sshSess, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, err
	}

	stdin, _ := sshSess.StdinPipe()
	stdout, _ := sshSess.StdoutPipe()
	stderr, _ := sshSess.StderrPipe()

	sshSess.RequestPty("xterm-256color", 40, 80, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	})
	sshSess.Shell()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		log.Printf("SFTP sub-protocol failed for host %d: %v", hostID, err)
		sftpClient = nil
	}

	as := &models.ActiveSession{
		HostID:       hostID,
		SSHClient:    client,
		SSHSession:   sshSess,
		SFTPClient:   sftpClient,
		Stdin:        stdin,
		Clients:      make(map[chan []byte]bool),
		LastActivity: time.Now(),
		OutputBuffer: make([]byte, 0),
	}

	// Background reading of SSH output
	go func() {
		combined := io.MultiReader(stdout, stderr)
		buf := make([]byte, 4096)
		for {
			n, err := combined.Read(buf)
			if err != nil {
				s.TerminateSession(userID, hostID)
				return
			}

			data := make([]byte, n)
			copy(data, buf[:n])

			as.Mu.Lock()
			as.OutputBuffer = append(as.OutputBuffer, data...)
			if len(as.OutputBuffer) > 51200 {
				as.OutputBuffer = as.OutputBuffer[len(as.OutputBuffer)-51200:]
			}

			for ch := range as.Clients {
				select {
				case ch <- data:
				default:
				}
			}
			as.Mu.Unlock()
		}
	}()

	s.Mu.Lock()
	s.Sessions[userID][hostID] = as
	s.Mu.Unlock()

	return as, nil
}

// TerminateSession kills a session.
func (s *SSHService) TerminateSession(userID, hostID int) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.terminateSessionUnsafe(userID, hostID)
}

// terminateSessionUnsafe does the dirty work without locking the mutex s.Mu.
func (s *SSHService) terminateSessionUnsafe(userID, hostID int) {
	userMap, ok := s.Sessions[userID]
	if !ok {
		return
	}

	as, ok := userMap[hostID]
	if !ok {
		return
	}

	// Closing everything
	if as.SFTPClient != nil {
		as.SFTPClient.Close()
	}
	if as.SSHSession != nil {
		as.SSHSession.Close()
	}
	if as.SSHClient != nil {
		as.SSHClient.Close()
	}
	if as.Stdin != nil {
		as.Stdin.Close()
	}

	// Notifying clients
	as.Mu.Lock()
	for ch := range as.Clients {
		select {
		case ch <- []byte("[STOPSESSION]"):
		default:
		}
	}
	as.Mu.Unlock()

	delete(userMap, hostID)
}

// startCleaner cleaner of abandoned SSH sessions.
func (s *SSHService) startCleaner() {
	ticker := time.NewTicker(s.CleanupInterval)
	for range ticker.C {
		s.Mu.Lock()
		for userID, userMap := range s.Sessions {
			for hostID, as := range userMap {
				as.Mu.Lock()
				isExpired := as.RefCount <= 0 && time.Since(as.LastActivity) > s.SessionTimeout
				as.Mu.Unlock()

				if isExpired {
					log.Printf("[CLEANER] Removing session: User %d, Host %d", userID, hostID)
					// We call the version without a lock
					s.terminateSessionUnsafe(userID, hostID)
				}
			}
		}
		s.Mu.Unlock()
	}
}

// GetActiveHostIDs returns a list of host IDs that have an active session for the user
func (s *SSHService) GetActiveHostIDs(userID int) map[int]bool {
	activeIDs := make(map[int]bool)

	s.Mu.RLock()
	defer s.Mu.RUnlock()

	if userMap, ok := s.Sessions[userID]; ok {
		for hID := range userMap {
			activeIDs[hID] = true
		}
	}
	return activeIDs
}
