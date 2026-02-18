package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"ssh_manager/internal/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// upgrader socket connection parameters
var upgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	CheckOrigin:      func(r *http.Request) bool { return true },
	HandshakeTimeout: 10 * time.Second,
}

// ResizeMessage Describes the structure of the JSON packet received from the frontend
// for resizing the terminal (PTY) on the fly.
type ResizeMessage struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

// SSHWebsocketHandler Handles the upgrade of an HTTP connection to a WebSocket.
// It connects the browser terminal's I/O stream (xterm.js) to the SSH session.
func (h *Handlers) SSHWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	hostID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	session, _ := h.Store.Get(r, utils.SessionName)
	userID, _ := session.Values[utils.UserIDKey].(int)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// We request a session from the service
	as, err := h.SSHService.GetSession(userID, hostID, r.Context())
	if err != nil {
		utils.LogErrorf("SSH Connection failed", err, "host_id", hostID)

		displayErr := "SSH Connection Error"
		if strings.Contains(err.Error(), "unable to authenticate") {
			displayErr = "Authentication failed: Please check your username and key/password"
		} else if strings.Contains(err.Error(), "i/o timeout") {
			displayErr = "Connection timeout: Server is unreachable"
		}

		errMsg := fmt.Sprintf("\r\n\x1b[31m[SSH Error] %v\x1b[0m\r\n", displayErr)
		_ = conn.WriteMessage(websocket.TextMessage, []byte(errMsg))
		// We give a short pause so that the frontend has time to read the message before closing the socket.
		time.Sleep(100 * time.Millisecond)

		return
	}

	messageChan := make(chan []byte, 256)

	as.Mu.Lock()
	as.Clients[messageChan] = true
	as.RefCount++
	if len(as.OutputBuffer) > 0 {
		_ = conn.WriteMessage(websocket.TextMessage, as.OutputBuffer)
	}
	as.Mu.Unlock()

	// WebSocket Send Goroutine
	go func() {
		for msg := range messageChan {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	// Reading from a websocket
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var rm ResizeMessage
		if err := json.Unmarshal(msg, &rm); err == nil && rm.Type == "resize" {
			as.Mu.Lock()
			if as.SSHSession != nil {
				as.SSHSession.WindowChange(rm.Rows, rm.Cols)
			}
			as.Mu.Unlock()
			continue
		}

		as.Mu.Lock()
		if as.Stdin != nil {
			as.Stdin.Write(msg)
		}
		as.LastActivity = time.Now()
		as.Mu.Unlock()
	}

	as.Mu.Lock()
	delete(as.Clients, messageChan)
	as.RefCount--
	as.Mu.Unlock()
	close(messageChan)
}

// TerminateSessionHandler Handles a request to immediately close an SSH connection.
func (h *Handlers) TerminateSessionHandler(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	hostID, _ := strconv.Atoi(idParam)

	session, _ := h.Store.Get(r, utils.SessionName)
	userID, _ := session.Values[utils.UserIDKey].(int)

	h.SSHService.TerminateSession(userID, hostID)

	w.WriteHeader(http.StatusOK)
}
