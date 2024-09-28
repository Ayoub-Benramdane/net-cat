package netclient

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func welcoming(conn net.Conn, clients *map[net.Conn]string, messages *[]string, mu *sync.Mutex, lenMessages *int) string {
	conn.Write([]byte("Welcome to TCP-Chat!\n" +
		"         _nnnn_\n" +
		"        dGGGGMMb\n" +
		"       @p~qp~~qMb\n" +
		"       M|@||@) M|\n" +
		"       @,----.JM|\n" +
		"      JS^\\__/  qKL\n" +
		"     dZP        qKRb\n" +
		"    dZP          qKKb\n" +
		"   fZP            SMMb\n" +
		"   HZM            MMMM\n" +
		"   FqM            MMMM\n" +
		" __| \".        |\\dS\"qML\n" +
		" |    `.       | `' \\Zq\n" +
		"_)      \\.___.,|     .'\n" +
		"\\____   )MMMMMP|   .'\n" +
		"     `-'       `--'\n" +
		"[ENTER YOUR NAME]: "))

	name := getName(conn, clients)
	if name == "" {
		return ""
	}
	mu.Lock()
	if len(*clients) >= 10 {
		fmt.Fprintf(conn, "Maximum connections reached. Try later.\n")
		conn.Close()
		mu.Unlock()
		return ""
	}
	(*clients)[conn] = name
	mu.Unlock()
	broadcast(fmt.Sprintf("\n%s has joined the chat...", name), *clients, conn, mu)
	for _, msg := range *messages {
		conn.Write([]byte(msg))
	}
	logMessage(fmt.Sprintf("%s has joined the chat...", name), messages, mu)
	*lenMessages = len(*messages)
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
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

func getName(conn net.Conn, clients *map[net.Conn]string) string {
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		for _, na := range name {
			if (na < 48 || na > 57) && (na < 65 || na > 90) && (na < 97 || na > 122) {
				conn.Write([]byte("invalid name please print alphanumeric one.\n[ENTER YOUR NAME]: "))
				return getName(conn, clients)
			}
		}
		if name == "" {
			conn.Write([]byte("invalid name please print alphanumeric one.\n[ENTER YOUR NAME]: "))
			return getName(conn, clients)
		}
		for _, n := range *clients {
			if strings.Compare(n, strings.TrimSpace(scanner.Text())) == 0 {
				conn.Write([]byte("this name is already registred.\n[ENTER YOUR NAME]:"))
				return getName(conn, clients)
			}
		}
		return name
	}
	return ""
}

func AddMsg(msg string) {
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	if _, err := file.WriteString(msg); err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func logMessage(msg string, messages *[]string, mu *sync.Mutex) {
	mu.Lock()
	*messages = append(*messages, msg+"\n")
	AddMsg(msg + "\n")
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

func handleConnection(conn net.Conn, clients *map[net.Conn]string, messages *[]string, mu *sync.Mutex, shut *bool) {
	defer conn.Close()
	lenMessages := 0
	name := welcoming(conn, clients, messages, mu, &lenMessages)
	if name == "" {
		delete(*clients, conn)
		return
	}
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
			logMessage(fmt.Sprintf("[%s][%s]:%s", time.Now().Format("2006-01-02 15:04:05"), name, msg), messages, mu)
		} else {
			break
		}
	}
	if !*shut {
		mu.Lock()
		delete(*clients, conn)
		mu.Unlock()
		lenMessages++
		broadcast(fmt.Sprintf("\n%s has left the chat...", name), *clients, nil, mu)
		logMessage(fmt.Sprintf("%s has left the chat...", name), messages, mu)
	} else {
		logMessage(fmt.Sprintf("%s has left the chat...", name), messages, mu)
	}
}
