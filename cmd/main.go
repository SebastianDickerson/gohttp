package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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

		lines := getLinesChannel(conn)

		for line := range lines {
			fmt.Println(line)
		}
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		var currentLine string
		for {
			buffer := make([]byte, 8)

			n, err := f.Read(buffer)
			if err != nil {
				if currentLine != "" {
					lines <- currentLine
					currentLine = ""
				}
				if errors.Is(err, io.EOF) {
					break
				}
				log.Fatalf("failed to read file: %v", err)
				break
			}
			str := string(buffer[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				lines <- currentLine + parts[i]
				currentLine = ""
			}
			currentLine += parts[len(parts)-1]
		}
	}()
	return lines
}
