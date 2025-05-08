package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")

	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Error dialing UDP:", err)
		return
	}
	fmt.Println("Connected to UDP server at", addr)
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		message, _ := reader.ReadString('\n')
		_, err := fmt.Fprint(conn, message)
		if err != nil {
			fmt.Println("Error sending message:", err)
			break
		}
	}
}
