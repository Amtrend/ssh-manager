package models

import (
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// ActiveSession model of active SSH user connections.
type ActiveSession struct {
	HostID       int
	SSHClient    *ssh.Client
	SSHSession   *ssh.Session
	Stdin        io.WriteCloser
	OutputBuffer []byte
	LastActivity time.Time
	Mu           sync.Mutex
	RefCount     int
	Clients      map[chan []byte]bool
}
