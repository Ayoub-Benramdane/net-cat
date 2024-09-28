package main

import (
	"fmt"
	C "netclient/Server"
	"os"
)

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
	C.Tcp_request(port)
}

func CheckPort(por string) bool {
	p := 0
	for _, c := range por {
		if c < 48 || c > 57 {
			fmt.Println("print a valid port number")
			return true
		}
		p = (p * 10) + (int(c) - 48)
	}
	if p < 1024 || p > 65535 {
		fmt.Println("port range > 1023 && < 65536")
		return true
	}
	return false
}
