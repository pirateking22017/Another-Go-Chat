package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func discoverServer() string {
	addr, _ := net.ResolveUDPAddr("udp", ":9999")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // avoid hanging forever

	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("No server found. Enter IP manually:")
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter server IP (e.g., 192.168.1.100:8080): ")
		ip, _ := reader.ReadString('\n')
		return strings.TrimSpace(ip)
	}

	serverAddr := string(buf[:n])
	fmt.Println("Discovered server:", serverAddr)
	return serverAddr
}


func main() {
	serverAddr := discoverServer()
	
	//
	//Bit of a doozy here, youre going to have to run ifconfig or ipconfig and edit this so that your client can connect to the network
	//in terminal or command prompt run
	//ipconfig (windows) / ifconfig (linux)
	

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()
	fmt.Println("\033[1;36mConnected to chat server.\033[0m")

	// Handle incoming messages
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		fmt.Println("\033[1;31mDisconnected from server.\033[0m")
		os.Exit(0)
	}()

	// Handle outgoing messages
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		text := input.Text()
		if strings.TrimSpace(text) == "/exit" {
			fmt.Fprintln(conn, "/exit")
			break
		}
		fmt.Fprintln(conn, text)
	}
}



