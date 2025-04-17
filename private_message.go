package main
import (
	"fmt"
	"net"
	"strings"
)

type PrivateMessage struct {
	sender    string // Username of the sender
	recipient string // Username of the recipient
	message   string // The actual message content
}

var (
	privateMsg = make(chan PrivateMessage)
)

// Format: /private <username> <message>
func handlePrivateMessage(conn net.Conn, message string) {
	// Split the message into parts: command, recipient, and content
	parts := strings.SplitN(message, " ", 3)
	if len(parts) != 3 {
		conn.Write([]byte("Usage: /private <username> <message>\n"))
		return
	}

	// Extract recipient and message content
	recipient := parts[1]
	content := parts[2]

	// Create and send the private message
	privateMsg <- PrivateMessage{
		sender:    clients[conn],
		recipient: recipient,
		message:   content,
	}
}


func processPrivateMessages() {
	for msg := range privateMsg {
		mutex.Lock()
		conn, ok := nameToConn[msg.recipient]
		senderConn := nameToConn[msg.sender]
		mutex.Unlock()

		if ok {
			mutex.Lock()
			lastPrivateSender[msg.recipient] = msg.sender
			mutex.Unlock()
			conn.Write([]byte(fmt.Sprintf("\033[34m[Private from %s] %s\033[0m\n", msg.sender, msg.message)))
		} else {
			senderConn.Write([]byte(fmt.Sprintf("User %s not found\n", msg.recipient)))
		}
	}
}
