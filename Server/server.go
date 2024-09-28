package netclient

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

func Tcp_request(port string) {
	clients := make(map[net.Conn]string)
	var mu sync.Mutex
	var messages []string
	shut := false
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()
	err = os.Remove("log.txt")
	if !os.IsNotExist(err) && err != nil {
		log.Fatalf("Error deleting log message file:%v", err)
	}
	fmt.Printf("Listening on the port: %s\n", port)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil || shut {
				if shut {
					fmt.Println("\nExiting server...")
					break
				}
				log.Printf("Error accepting connection: %v", err)
				continue
			}
			go handleConnection(conn, &clients, &messages, &mu, &shut)
		}
	}()
	<-sigs
	shut = true
	shutdown(&clients, "\nserver interrupted : Press enter to shut down")
}

func shutdown(clie *map[net.Conn]string, message string) {
	for conn := range *clie {
		conn.Write([]byte(message))
		conn.Close()
		delete(*clie, conn)
	}
}
