package main

import (
	"fmt"
	"log"
	"net"

	"httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error listening for TCP connection:", err.Error())
	}
	defer l.Close()
	log.Println("Listening on port", port)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err.Error())
		}
		log.Println("Accepted connection from", conn.RemoteAddr())

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Println("Error reading request:", err.Error())
			conn.Close()
			continue
		}
		fmt.Printf("Request line: \n")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HTTPVersion)

		fmt.Printf("Headers: \n")
		for k, v := range req.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}

		fmt.Printf("Body: \n")
		fmt.Printf("%s\n", req.Body)
	}
}
