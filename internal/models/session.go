package models

import (
	"io"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// ActiveSession model of active SSH user connections.
type ActiveSession struct {
	HostID       int
	SSHClient    *ssh.Client
	SSHSession   *ssh.Session
	SFTPClient   *sftp.Client
	Stdin        io.WriteCloser
	OutputBuffer []byte
	LastActivity time.Time
	Mu           sync.Mutex
	RefCount     int
	Clients      map[chan []byte]bool
}
