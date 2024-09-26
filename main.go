package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func CheckPort(por string) bool {
	p := 0
	for _, c := range por {
		if c < 48 || c > 57 {
			fmt.Print("print a valid port number")
			return true
		}
		p = (p * 10) + (int(c) - 48)
	}
	if p < 1024 || p > 65535 {
		fmt.Print("port range >1024 && <65535")
		return true
	}
	return false
}

func welcoming(conn net.Conn, clients *map[net.Conn]string, messages *[]string, mu *sync.Mutex, lenMessages *int) string {
	conn.Write([]byte("Welcome to TCP-Chat!\n         _nnnn_\n        dGGGGMMb\n       @p~qp~~qMb\n       M|@||@) M|\n       @,----.JM|\n      JS^\\__/  qKL\n     dZP        qKRb\n    dZP          qKKb\n   fZP            SMMb\n   HZM            MMMM\n   FqM            MMMM\n __| \".        |\\dS\"qML\n |    `.       | `' \\Zq\n_)      \\.___.,|     .'\n\\____   )MMMMMP|   .'\n     `-'       `--'\n[ENTER YOUR NAME]: "))

	name := getName(conn)
	mu.Lock()
	(*clients)[conn] = name
	mu.Unlock()
	broadcast(fmt.Sprintf("\n%s has joined the chat...", name), *clients, conn, mu)

	// Send previous messages to the new client
	for _, msg := range *messages {
		conn.Write([]byte(msg))
	}
	*messages = append(*messages, fmt.Sprintf("%s has joined the chat...\n", name))
	*lenMessages = len(*messages)
	go func() {
		for {
			mu.Lock()
			if len(*messages) > *lenMessages {
				conn.Write([]byte(fmt.Sprintf("[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), name)))
				*lenMessages = len(*messages)
			}
			mu.Unlock()
		}
	}()
	return name
}

func handleConnection(conn net.Conn, clients *map[net.Conn]string, messages *[]string, mu *sync.Mutex) {
	defer conn.Close()
	lenMessages := 0
	name := welcoming(conn, clients, messages, mu, &lenMessages)
	scanner := bufio.NewScanner(conn)
	for {
		conn.Write([]byte(fmt.Sprintf("[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), name)))
		if scanner.Scan() {
			msg := scanner.Text()
			if strings.TrimSpace(msg) == "" {
				continue
			}
			lenMessages++
			broadcast(fmt.Sprintf("\n[%s][%s]:%s", time.Now().Format("2006-01-02 15:04:05"), name, msg), *clients, conn, mu)
			logMessage(name, msg, messages, mu)
		} else {
			break
		}
	}
	mu.Lock()
	delete(*clients, conn)
	mu.Unlock()
	lenMessages++
	broadcast(fmt.Sprintf("\n%s has left the chat...", name), *clients, nil, mu)
	logMessage(name, fmt.Sprintf("%s has left the chat...", name), messages, mu)
}

func getName(conn net.Conn) string {
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	name := strings.TrimSpace(scanner.Text())
	if name == "" {
		getName(conn)
	}
	return name
}

func AddMsg(msg string) {
	err := os.WriteFile("file.txt", []byte(msg), 0o644)
	if err != nil {
		return
	}
}

func logMessage(name, msg string, messages *[]string, mu *sync.Mutex) {
	mu.Lock()
	*messages = append(*messages, fmt.Sprintf("[%s][%s]: %s\n", time.Now().Format("2006-01-02 15:04:05"), name, msg))
	file, _ := os.Open("file.txt")
	_, err := fmt.Fprint(file, msg)
    if err != nil {
        fmt.Println("Error writing to file:", err)
    }
	mu.Unlock()
}

func broadcast(message string, clients map[net.Conn]string, conn net.Conn, mu *sync.Mutex) {
	mu.Lock()
	for client := range clients {
		if conn != client {
			client.Write([]byte(message + "\n"))
		}
	}
	mu.Unlock()
}

func Tcp_request(port string) {
	clients := make(map[net.Conn]string)
	var mu sync.Mutex
	var messages []string

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Listening on the port: %s\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		mu.Lock()
		if len(clients) >= 10 {
			fmt.Fprintf(conn, "Maximum connections reached. Try later.\n")
			conn.Close()
			mu.Unlock()
			continue
		}
		mu.Unlock()
		go handleConnection(conn, &clients, &messages, &mu)
	}
}

func main() {
	port := "8989"
	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	} else if len(os.Args) == 2 {
		if CheckPort(os.Args[1]) {
			return
		}
		port = os.Args[1]
	}
	Tcp_request(port)
}
