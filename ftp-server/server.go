package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"path/filepath"
)

func runServer() {
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	defer ln.Close()
	fmt.Println("Server listening on", *addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	file, err := os.Open(*filePath)
	if err != nil {
		log.Printf("Cannot open file: %v", err)
		return
	}
	defer file.Close()

	reader := make([]byte, 1024)
	conn.Write([]byte("220 Simple FTP Server Ready\r\n"))

	for {
		n, err := conn.Read(reader)
		if err != nil {
			log.Printf("Connection read error: %v", err)
			return
		}
		input := string(reader[:n])
		input = strings.TrimSpace(input)

		fmt.Println("Received:", input)

		switch {
		case strings.HasPrefix(input, "USER"):
			conn.Write([]byte("331 Username OK, need password\r\n"))
		case strings.HasPrefix(input, "PASS"):
			conn.Write([]byte("230 Login successful\r\n"))
		case strings.HasPrefix(input, "SYST"):
			conn.Write([]byte("215 UNIX Type: L8\r\n"))
		case strings.HasPrefix(input, "TYPE"):
			conn.Write([]byte("200 Type set to I.\r\n"))
		case strings.HasPrefix(input, "PASV"):
			conn.Write([]byte("502 Passive mode not implemented\r\n"))
		case strings.HasPrefix(input, "RETR"):
			conn.Write([]byte("150 Opening binary mode data connection\r\n"))
			io.Copy(conn, file)
			conn.Write([]byte("226 Transfer complete\r\n"))
			return
		case strings.HasPrefix(input, "INFO"):
			// conn.Write([]byte("200 Serving file: " + *filePath + "\r\n"))
			filename := filepath.Base(*filePath)
			conn.Write([]byte("200 Serving file: " + filename + "\r\n"))
			// return
		case strings.HasPrefix(input, "QUIT"):
			conn.Write([]byte("221 Goodbye\r\n"))
			return
		default:
			conn.Write([]byte("500 Unknown command\r\n"))
		}
	}
}
