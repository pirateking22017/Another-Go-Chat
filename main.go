package main
import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)
var (
	clients = make(map[net.Conn]string)
	nameToConn = make(map[string]net.Conn)
	nameToPass = make(map[string]string)
	broadcast = make(chan string)
	status = make(map[string]string)
	mutex             = &sync.Mutex{}
	lastPrivateSender = make(map[string]string) 
)

func main() {
	if data, err := os.ReadFile("users.json"); err == nil {
		if err := json.Unmarshal(data, &nameToPass); err != nil {
			fmt.Println("Error loading users:", err)
		}
	}

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer ln.Close()

	go startUDPBroadcast()

	go handleBroadcasting()     
	go processPrivateMessages() 

	fmt.Println("Server is running on port 8080")
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	reader := bufio.NewReader(conn)
	var username string
	var authenticated bool
	conn.Write([]byte("\033[1;36mWelcome to the Chat Server!\033[0m\n"))
	conn.Write([]byte("\033[1;32mPlease register or login:\033[0m\n"))
	conn.Write([]byte("\033[1;33m1. To register: /register <username> <password>\033[0m\n"))
	conn.Write([]byte("\033[1;33m2. To login: /login <username> <password>\033[0m\n"))

	for !authenticated {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}
		message = strings.TrimSpace(message)

		if strings.HasPrefix(message, "/register") {
			username = handleRegisterCommand(conn, message)
			if username != "" {
				authenticated = true
			}
		} else if strings.HasPrefix(message, "/login") {
			username = handleLoginCommand(conn, message)
			if username != "" {
				authenticated = true
			}
		} else if strings.HasPrefix(message, "/exit") {
			handleExitCommand(conn)
			return
		} else {
			conn.Write([]byte("\033[1;31mPlease register or login first.\033[0m\n"))
		}
	}

	conn.Write([]byte("\033[1;33mEnter your display name: \033[0m"))
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading name:", err)
		return
	}
	name = strings.TrimSpace(name)
	mutex.Lock()
	clients[conn] = name
	nameToConn[name] = conn
	mutex.Unlock()
	broadcast <- fmt.Sprintf("\033[33m%s has joined the chat\033[0m\n", name)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		message = strings.TrimSpace(message)
		if handleCommand(conn, message) {
			continue
		}
		broadcast <- fmt.Sprintf("\033[34m%s: %s\033[0m\n", name, message)
	}
	mutex.Lock()
	delete(clients, conn)
	delete(nameToConn, name)
	mutex.Unlock()
	broadcast <- fmt.Sprintf("\033[33m%s has left the chat\033[0m\n", name)
	conn.Close()
}
func handleRegisterCommand(conn net.Conn, message string) string {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) != 3 {
		conn.Write([]byte("Usage: /register <username> <password>\n"))
		return ""
	}
	username := strings.TrimSpace(parts[1])
	password := strings.TrimSpace(parts[2])
	mutex.Lock()
	_, exists := nameToPass[username]
	mutex.Unlock()
	if exists {
		conn.Write([]byte("\033[1;31mUsername already exists. Please choose another.\033[0m\n"))
		return ""
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		conn.Write([]byte("\033[1;31mError registering user. Please try again.\033[0m\n"))
		return ""
	}
	mutex.Lock()
	nameToPass[username] = string(hashedPassword)
	mutex.Unlock()
	if err := saveUsersToFile(); err != nil {
		fmt.Println("Error saving users:", err)
		conn.Write([]byte("\033[1;31mError saving registration. Please try again.\033[0m\n"))
		return ""
	}
	conn.Write([]byte(fmt.Sprintf("\033[1;32mWelcome, %s! You can now start chatting.\033[0m\n", username)))
	return username
}
func saveUsersToFile() error {
	mutex.Lock()
	data, err := json.Marshal(nameToPass)
	mutex.Unlock()
	if err != nil {
		return err
	}
	return os.WriteFile("users.json", data, 0644)
}

func handleCommand(conn net.Conn, message string) bool {
	if strings.HasPrefix(message, "/register") {
		handleRegisterCommand(conn, message)
		return true
	}
	if strings.HasPrefix(message, "/users") {
		handleUsersCommand(conn)
		return true
	}
	if strings.HasPrefix(message, "/private") {
		handlePrivateMessage(conn, message)
		return true
	}
	if strings.HasPrefix(message, "/reply") {
		handleReplyCommand(conn, message)
		return true
	}
	if strings.HasPrefix(message, "/exit") {
		handleExitCommand(conn)
		return true
	}
	if strings.HasPrefix(message, "/help") {
		handleHelpCommand(conn)
		return true
	}
	if strings.HasPrefix(message, "/status") {
		handleStatusCommand(conn, message)
		return true
	}
	return false
}
func handleStatusCommand(conn net.Conn, message string) {
	parts := strings.SplitN(message, " ", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		conn.Write([]byte("\033[1;31mUsage: /status <set status>\033[0m\n"))
		return
	}
	newStatus := parts[1]
	mutex.Lock()
	username := clients[conn]
	status[username] = newStatus
	mutex.Unlock()
	conn.Write([]byte(fmt.Sprintf("\033[1;32mYour status has been set to: %s\033[0m\n", newStatus)))
}
func handleReplyCommand(conn net.Conn, message string) {
	mutex.Lock()
	username := clients[conn]
	lastSender, ok := lastPrivateSender[username]
	mutex.Unlock()
	if !ok {
		conn.Write([]byte("\033[1;31mNo private messages to reply to.\033[0m\n"))
		return
	}
	parts := strings.SplitN(message, " ", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		conn.Write([]byte("\033[1;31mUsage: /reply <message>\033[0m\n"))
		return
	}
	msg := parts[1]
	handlePrivateMessage(conn, fmt.Sprintf("/private %s %s", lastSender, msg))
}

func handleUsersCommand(conn net.Conn) {
	mutex.Lock()
	for _, name := range clients {
		status, ok := status[name]
		if ok {
			conn.Write([]byte(fmt.Sprintf("\033[90m%s (%s)\033[0m\n", name, status)))
		} else {
			conn.Write([]byte("\033[90m" + name + "\033[0m\n"))
		}
	}
	mutex.Unlock()
}

func handleBroadcasting() {
	for message := range broadcast {
		mutex.Lock()
		for conn, _ := range clients {
			conn.Write([]byte(message))
		}
		mutex.Unlock()
	}
}

func handleLoginCommand(conn net.Conn, message string) string {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) != 3 {
		conn.Write([]byte("Usage: /login <username> <password>\n"))
		return ""
	}
	username := strings.TrimSpace(parts[1])
	password := strings.TrimSpace(parts[2])

	mutex.Lock()
	hashedPassword, exists := nameToPass[username]
	mutex.Unlock()

	if !exists {
		conn.Write([]byte("\033[1;31mUser not found. Please register first.\033[0m\n"))
		return ""
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		conn.Write([]byte("\033[1;31mInvalid password.\033[0m\n"))
		return ""
	}

	conn.Write([]byte(fmt.Sprintf("\033[1;32mWelcome back, %s!\033[0m\n", username)))
	return username
}

func handleExitCommand(conn net.Conn) {
	mutex.Lock()
	name := clients[conn]
	delete(clients, conn)
	delete(nameToConn, name)
	mutex.Unlock()
	broadcast <- fmt.Sprintf("\033[33m%s has left the chat\033[0m\n", name)
	conn.Write([]byte("\033[1;32mGoodbye! Thanks for chatting.\033[0m\n"))
	conn.Close()
}

func handleHelpCommand(conn net.Conn) {
	helpMessage := "\033[1;36mAvailable Commands:\033[0m\n\n" +
		"\033[1;33m/register <username> <password>\033[0m\n" +
		"    Register a new user account\n\n" +
		"\033[1;33m/login <username> <password>\033[0m\n" +
		"    Login to your account\n\n" +
		"\033[1;33m/users\033[0m\n" +
		"    List all currently connected users\n\n" +
		"\033[1;33m/private <username> <message>\033[0m\n" +
		"    Send a private message to a specific user\n\n" +
		"\033[1;33m/reply <message>\033[0m\n" +
		"    Reply to the last private message you received\n\n" +
		"\033[1;33m/exit\033[0m\n" +
		"    Exit the chat server\n\n" +
		"\033[1;33m/help\033[0m\n" +
		"    Display this help message\n\n" +
		"\033[1;36mRegular Messages:\033[0m\n" +
		"    Type any message without a command to broadcast to all users\n"

	conn.Write([]byte(helpMessage))
}

func startUDPBroadcast() {
	addr, _ := net.ResolveUDPAddr("udp", "255.255.255.255:9999")
	conn, _ := net.DialUDP("udp", nil, addr)
	defer conn.Close()

	for {
		conn.Write([]byte("CHAT_SERVER_HERE:8080"))
		time.Sleep(3 * time.Second)
	}
}
