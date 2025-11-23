package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"os"
	"path/filepath"
	"io"
)

func StartServer_(port int, timeout int, verbose bool) {
	if verbose {
		println("Verbose logging enabled")
	}
	println("Starting server on port:", port)
	println("Timeout set to:", timeout, "seconds")
	// Placeholder for actual server logic
}

func StartServer(sharedDir string, port *string) {
	ln, err := net.Listen("tcp", *port)
	if err != nil {
		fmt.Print("Error listening:", err)
		return
	}
	defer ln.Close()
	fmt.Printf("Server listening on %s serving %s", *port, sharedDir)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		fmt.Println("Client connected")
		go handleConnection(conn, sharedDir)
	}
}

func handleConnection(conn net.Conn, sharedDir string) {
	var dataListener net.Listener

	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	sendLine(writer, "220 Simple FTP server ready")

	authenticated := false

	rootDir := sharedDir
	currentDir := sharedDir //"./"

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("Read error:", err)
			}
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		cmd, arg := parseCmd(line)

		switch strings.ToUpper(cmd) {
		case "HELP":
			handleHelpCommand(writer)

		case "USER":
			handleUserCommand(writer, arg)

		case "PASS":
			// For simplicity, accept any password
			authenticated = true
			sendLine(writer, "230 User logged in")

		case "PWD":
			if !authenticated {
				sendLine(writer, "530 Not logged in")
				continue
			}
			sendLine(writer, fmt.Sprintf("257 \"%s\"", currentDir))

		case "OLD_CWD":
			newDir := filepath.Join(currentDir, arg)
			if _, err := os.Stat(newDir); err != nil {
				sendLine(writer, "550 Directory not found")
			} else {
				currentDir = newDir
				sendLine(writer, "250 Directory changed")
			}

		case "CWD":
			// TODO: implement cwd handler 
			if !authenticated {
				sendLine(writer, "530 Not logged in")
				continue
			}

			handleCwdCommand(writer, arg, rootDir, &currentDir)

		case "CDUP":
			if !authenticated {
				sendLine(writer, "530 Not logged in")
				continue
			}
			handleCdupCommand(writer, rootDir, &currentDir)


		case "LIST":
			if !authenticated {
				sendLine(writer, "530 Not logged in")
				continue
			}
			handleListCommand(writer, currentDir, &dataListener)

		case "RETR":
			if !authenticated {
				sendLine(writer, "530 Not logged in")
				continue
			}
			handleRetrCommand(writer, arg, currentDir, &dataListener)

		case "STOR":
			if !authenticated {
				sendLine(writer, "530 Not logged in")
				continue
			}
			if arg == "" {
				sendLine(writer, "501 Syntax error in parameters or arguments")
				continue
			}
			if dataListener == nil {
				sendLine(writer, "425 Use PASV first")
				continue
			}

			// Path for uploaded file
			filePath := filepath.Join(currentDir, arg)

			// Create or overwrite the file
			f, err := os.Create(filePath)
			if err != nil {
				sendLine(writer, "550 Cannot create file")
				continue
			}

			sendLine(writer, "150 Opening data connection for file upload")

			// Accept incoming data connection
			dataConn, err := dataListener.Accept()
			if err != nil {
				sendLine(writer, "425 Can't open data connection")
				dataListener.Close()
				dataListener = nil
				f.Close()
				continue
			}

			// Copy data from client to file
			_, copyErr := io.Copy(f, dataConn)

			f.Close()
			dataConn.Close()
			dataListener.Close()
			dataListener = nil

			if copyErr != nil {
				sendLine(writer, "426 Connection closed; transfer aborted")
				continue
			}

			sendLine(writer, "226 Transfer complete")

		case "PASV":
			// Listen on any available port
			dataListener, err = net.Listen("tcp", "0.0.0.0:0")
			if err != nil {
				sendLine(writer, "425 Can't open data connection")
				continue
			}

			// Get the port
			addr := dataListener.Addr().(*net.TCPAddr)
			p1 := addr.Port / 256
			p2 := addr.Port % 256

			// Send PASV response with server IP and port
			hostIP := getLANIP() // "127,0,0,1" // You should get the actual IP of your server here
			sendLine(writer, fmt.Sprintf("227 Entering Passive Mode (%s,%d,%d)", hostIP, p1, p2))
		case "QUIT":
			sendLine(writer, "221 Goodbye")
			return
		default:
			sendLine(writer, "502 Command not implemented")
		}
	}
}
