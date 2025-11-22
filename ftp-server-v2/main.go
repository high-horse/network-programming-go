package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var serverAddr *string
var port *string

func main() {
	mode := flag.String("mode", "client", "server or client")
	serverAddr = flag.String("addr", "localhost:2121", "Ip:port of server hosting the file")
	port = flag.String("port", ":2121", "Port to host")
	sharedDir := flag.String("dir", "./", "Directory you want to share vis FTP")
	flag.Parse()

	if *mode == "server" {
		runServer(*sharedDir)
	} else {
		runClient()
	}
}

func getLANIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return strings.ReplaceAll(ipnet.IP.String(), ".", ",")
		}
	}
	return "127,0,0,1" // fallback
}

func runServer(sharedDir string) {
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
			sendLine(writer, "214-Commands:")
			cmds := []string{
				"USER username",
				"PASS password",
				"PASV upgrade_connection",
				"PWD current-dir",
				"LIST list",
				"CWD change_working_directory",
				"CDUP move_cd_to_parent_dir",
				"RETR file_name_to_retrieve",
				"STOR upload_file",
				"QUIT quit",
			}
			for _, c := range cmds {
				sendLine(writer, "214-"+c)
			}
			sendLine(writer, "214 End of HELP")
		case "USER":
			sendLine(writer, "331 User name okay, need password")
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
			if !authenticated {
				sendLine(writer, "530 Not logged in")
				continue
			}
			if arg == "" {
				sendLine(writer, "501 Missing directory")
				continue
			}

			// Build new path
			newPath := filepath.Join(currentDir, arg)

			// Normalize path
			newPath, err := filepath.Abs(newPath)
			if err != nil {
				sendLine(writer, "550 Invalid path")
				continue
			}

			// Ensure user cannot leave rootDir
			if !strings.HasPrefix(newPath, rootDir) {
				sendLine(writer, "550 Access denied")
				continue
			}

			// Check if directory exists
			info, err := os.Stat(newPath)
			if err != nil || !info.IsDir() {
				sendLine(writer, "550 NhandleConot a directory")
				continue
			}

			currentDir = newPath
			sendLine(writer, "250 Directory successfully changed")

		case "CDUP":
			if !authenticated {
				sendLine(writer, "530 Not logged in")
				continue
			}

			parent := filepath.Dir(currentDir)

			// Normalize
			parent, err := filepath.Abs(parent)
			if err != nil {
				sendLine(writer, "550 Invalid path")
				continue
			}

			// Prevent escape above root
			if !strings.HasPrefix(parent, rootDir) {
				sendLine(writer, "550 Access denied")
				continue
			}

			currentDir = parent
			sendLine(writer, "200 Command okay")

		case "LIST":
			if !authenticated {
				sendLine(writer, "530 Not logged in")
				continue
			}

			if dataListener == nil {
				sendLine(writer, "425 Use PASV first")
				continue
			}

			sendLine(writer, "150 Here comes the directory listing")

			dataConn, err := dataListener.Accept()
			if err != nil {
				sendLine(writer, "425 Can't open data connection")
				dataListener.Close()
				dataListener = nil
				continue
			}

			// Now send the directory listing over dataConn
			files, err := os.ReadDir(currentDir)
			if err != nil {
				sendLine(writer, "550 Failed to list directory")
				dataConn.Close()
				dataListener.Close()
				dataListener = nil
				continue
			}

			for _, f := range files {
				info, err := f.Info()
				if err != nil {
					continue
				}
				modTime := info.ModTime().Format("Jan _2 15:04")
				line := fmt.Sprintf("%s %12d %s %s\r\n", fileModeToStr(info.Mode()), info.Size(), modTime, f.Name())
				dataConn.Write([]byte(line))
			}

			dataConn.Close()
			dataListener.Close()
			dataListener = nil

			sendLine(writer, "226 Directory send OK")

		case "RETR":
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

			filepath := filepath.Join(currentDir, arg)
			f, err := os.Open(filepath)
			if err != nil {
				sendLine(writer, "550 File not found")
				continue
			}
			defer f.Close()

			sendLine(writer, "150 Opening data connection for file transfer")

			dataConn, err := dataListener.Accept()
			if err != nil {
				sendLine(writer, "425 Can't open data connection")
				dataListener.Close()
				dataListener = nil
				continue
			}

			io.Copy(dataConn, f)
			dataConn.Close()
			dataListener.Close()
			dataListener = nil

			sendLine(writer, "226 Transfer complete")

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

func handleLIST(writer *bufio.Writer, dir string) {
	sendLine(writer, "Searching in directory: "+dir)
	files, err := os.ReadDir(dir)
	if err != nil {
		sendLine(writer, "550 Failed to list directory")
		return
	}

	sendLine(writer, "150 Here comes the directory listing")

	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			sendLine(writer, fmt.Sprintf("550 Could not stat file: %s", f.Name()))
			continue
		}
		modTime := info.ModTime().Format("Jan _2 15:04")
		line := fmt.Sprintf("%s %12d %s %s", fileModeToStr(info.Mode()), info.Size(), modTime, f.Name())
		sendLine(writer, line)
	}

	sendLine(writer, "226 Directory send OK")
}

func parseCmd(line string) (string, string) {
	parts := strings.SplitN(line, " ", 2)
	cmd := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = parts[1]
	}
	return cmd, arg
}

func sendLine(w *bufio.Writer, line string) {
	// fmt.Fprintln(w, line)
	w.WriteString(line + "\r\n")
	w.Flush()
}

func fileModeToStr(mode os.FileMode) string {
	// Simplified version of ls -l mode string
	var str strings.Builder
	if mode.IsDir() {
		str.WriteByte('d')
	} else {
		str.WriteByte('-')
	}
	perms := []struct {
		bit  os.FileMode
		char byte
	}{
		{0400, 'r'}, {0200, 'w'}, {0100, 'x'},
		{0040, 'r'}, {0020, 'w'}, {0010, 'x'},
		{0004, 'r'}, {0002, 'w'}, {0001, 'x'},
	}
	for _, p := range perms {
		if mode&p.bit != 0 {
			str.WriteByte(p.char)
		} else {
			str.WriteByte('-')
		}
	}
	return str.String()
}

func runClient() {
	conn, err := net.Dial("tcp", *serverAddr)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	console := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(conn)

	// Read welcome message
	line, _ := reader.ReadString('\n')
	fmt.Print("Server: " + line)

	var dataConn net.Conn = nil

	for {
		fmt.Print("ftp> ")
		cmdLine, err := console.ReadString('\n')
		if err != nil {
			break
		}
		cmdLine = strings.TrimSpace(cmdLine)
		if cmdLine == "" {
			continue
		}

		cmdUpper := strings.ToUpper(cmdLine)

		// Special handling for PASV command to get data port info
		if cmdUpper == "PASV" {
			writer.WriteString(cmdLine + "\r\n")
			writer.Flush()

			resp, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Connection closed")
				return
			}
			fmt.Print("Server: " + resp)

			// Parse PASV response e.g. 227 Entering Passive Mode (127,0,0,1,168,161)
			start := strings.Index(resp, "(")
			end := strings.Index(resp, ")")
			if start == -1 || end == -1 || end <= start {
				fmt.Println("Failed to parse PASV response")
				continue
			}

			addrParts := strings.Split(resp[start+1:end], ",")
			if len(addrParts) != 6 {
				fmt.Println("Unexpected PASV address format")
				continue
			}

			ip := strings.Join(addrParts[0:4], ".")
			p1 := atoi(addrParts[4])
			p2 := atoi(addrParts[5])
			port := p1*256 + p2

			dataConn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
			if err != nil {
				fmt.Println("Failed to connect to data port:", err)
				dataConn = nil
			} else {
				fmt.Printf("Data connection established to %s:%d\n", ip, port)
			}

			continue
		}

		// Handle LIST command
		if cmdUpper == "LIST" {
			if dataConn == nil {
				fmt.Println("No data connection established. Use PASV first.")
				continue
			}

			writer.WriteString(cmdLine + "\r\n")
			writer.Flush()

			// Read control responses until 150
			for {
				resp, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Connection closed")
					return
				}
				fmt.Print("Server: " + resp)
				if strings.HasPrefix(resp, "150") {
					break
				}
				if strings.HasPrefix(resp, "425") || strings.HasPrefix(resp, "530") {
					break
				}
			}

			// Now read directory listing from dataConn
			dataReader := bufio.NewReader(dataConn)
			for {
				line, err := dataReader.ReadString('\n')
				if err != nil {
					break
				}
				fmt.Print(line)
			}
			dataConn.Close()
			dataConn = nil

			// Read final confirmation after data transfer
			for {
				resp, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Connection closed")
					return
				}
				fmt.Print("Server: " + resp)
				if len(resp) >= 4 && resp[3] == ' ' {
					break
				}
			}

			continue
		}

		if strings.HasPrefix(cmdUpper, "STOR") {
			if dataConn == nil {
				fmt.Println("No data connection established. Use PASV first.")
				continue
			}

			parts := strings.SplitN(cmdLine, " ", 2)
			if len(parts) < 2 {
				fmt.Println("No filename specified for STOR")
				continue
			}
			filename := parts[1]

			// Open local file
			file, err := os.Open(filename)
			if err != nil {
				fmt.Println("Failed to open file:", err)
				dataConn.Close()
				dataConn = nil
				continue
			}

			// Send STOR command
			writer.WriteString(cmdLine + "\r\n")
			writer.Flush()

			// Wait for 150
			transferOkay := false
			for {
				resp, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Connection closed")
					return
				}
				fmt.Print("Server: " + resp)
				if strings.HasPrefix(resp, "150") {
					transferOkay = true
					break
				}
				if strings.HasPrefix(resp, "425") || strings.HasPrefix(resp, "530") || strings.HasPrefix(resp, "550") {
					dataConn.Close()
					dataConn = nil
					transferOkay = false
					break
				}
			}

			if !transferOkay {
				file.Close()
				continue
			}

			// Upload file bytes
			_, err = io.Copy(dataConn, file)
			file.Close()
			dataConn.Close()
			dataConn = nil

			if err != nil {
				fmt.Println("Error uploading file:", err)
				continue
			}

			// Wait for final response (226)
			for {
				resp, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Connection closed")
					return
				}
				fmt.Print("Server: " + resp)
				if len(resp) >= 4 && resp[3] == ' ' {
					break
				}
			}

			continue
		}

		// Handle RETR (download) command
		if strings.HasPrefix(cmdUpper, "RETR") {
			if dataConn == nil {
				fmt.Println("No data connection established. Use PASV first.")
				continue
			}

			// Send RETR command
			writer.WriteString(cmdLine + "\r\n")
			writer.Flush()

			// Wait for "150" reply from server before reading data
			transferStarted := false
			for {
				resp, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Connection closed")
					return
				}
				fmt.Print("Server: " + resp)
				if strings.HasPrefix(resp, "150") {
					transferStarted = true
					break
				}
				if strings.HasPrefix(resp, "425") || strings.HasPrefix(resp, "530") || strings.HasPrefix(resp, "550") {
					// Error replies
					dataConn.Close()
					dataConn = nil
					transferStarted = false
					break
				}
			}

			if !transferStarted {
				continue
			}

			// Parse filename from command line for local saving
			parts := strings.SplitN(cmdLine, " ", 2)
			if len(parts) < 2 {
				fmt.Println("No filename specified for RETR")
				dataConn.Close()
				dataConn = nil
				continue
			}
			filename := parts[1]

			// Open local file to save
			file, err := os.Create(filename)
			if err != nil {
				fmt.Println("Failed to create local file:", err)
				dataConn.Close()
				dataConn = nil
				continue
			}

			// Copy data from data connection to file
			_, err = io.Copy(file, dataConn)
			file.Close()
			dataConn.Close()
			dataConn = nil

			if err != nil {
				fmt.Println("Error downloading file:", err)
				continue
			}

			// Read final server response after data transfer
			for {
				resp, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Connection closed")
					return
				}
				fmt.Print("Server: " + resp)
				if len(resp) >= 4 && resp[3] == ' ' {
					break
				}
			}

			continue
		}

		// For other commands, just send and print response normally
		writer.WriteString(cmdLine + "\r\n")
		writer.Flush()

		// Read server response(s)
		for {
			resp, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Connection closed")
				return
			}
			fmt.Print("Server: " + resp)
			if len(resp) < 4 {
				continue
			}
			// Responses start with 3-digit code and a space means last line
			if resp[3] == ' ' {
				break
			}
		}

		if cmdUpper == "QUIT" {
			break
		}
	}
}

func runClient_v1() {
	conn, err := net.Dial("tcp", "localhost:2121")
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	console := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(conn)

	// Read welcome message
	line, _ := reader.ReadString('\n')
	fmt.Print("Server: " + line)

	var dataConn net.Conn = nil

	for {
		fmt.Print("ftp> ")
		cmdLine, err := console.ReadString('\n')
		if err != nil {
			break
		}
		cmdLine = strings.TrimSpace(cmdLine)
		if cmdLine == "" {
			continue
		}

		cmdUpper := strings.ToUpper(cmdLine)

		// Special handling for PASV command to get data port info
		if cmdUpper == "PASV" {
			writer.WriteString(cmdLine + "\r\n")
			writer.Flush()

			resp, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Connection closed")
				return
			}
			fmt.Print("Server: " + resp)

			// Parse PASV response e.g. 227 Entering Passive Mode (127,0,0,1,168,161)
			start := strings.Index(resp, "(")
			end := strings.Index(resp, ")")
			if start == -1 || end == -1 || end <= start {
				fmt.Println("Failed to parse PASV response")
				continue
			}

			addrParts := strings.Split(resp[start+1:end], ",")
			if len(addrParts) != 6 {
				fmt.Println("Unexpected PASV address format")
				continue
			}

			ip := strings.Join(addrParts[0:4], ".")
			p1 := atoi(addrParts[4])
			p2 := atoi(addrParts[5])
			port := p1*256 + p2

			dataConn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
			if err != nil {
				fmt.Println("Failed to connect to data port:", err)
				dataConn = nil
			} else {
				fmt.Printf("Data connection established to %s:%d\n", ip, port)
			}

			continue
		}

		// For LIST, send command, then read from data connection
		if cmdUpper == "LIST" {
			if dataConn == nil {
				fmt.Println("No data connection established. Use PASV first.")
				continue
			}

			writer.WriteString(cmdLine + "\r\n")
			writer.Flush()

			// Read control responses until 150
			for {
				resp, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Connection closed")
					return
				}
				fmt.Print("Server: " + resp)
				if strings.HasPrefix(resp, "150") {
					break
				}
				if strings.HasPrefix(resp, "425") || strings.HasPrefix(resp, "530") {
					break
				}
			}

			// Now read directory listing from dataConn
			dataReader := bufio.NewReader(dataConn)
			for {
				line, err := dataReader.ReadString('\n')
				if err != nil {
					break
				}
				fmt.Print(line)
			}
			dataConn.Close()
			dataConn = nil

			// Read final confirmation after data transfer
			for {
				resp, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Connection closed")
					return
				}
				fmt.Print("Server: " + resp)
				if len(resp) >= 4 && resp[3] == ' ' {
					break
				}
			}

			continue
		}

		// For other commands, just send and print response normally
		writer.WriteString(cmdLine + "\r\n")
		writer.Flush()

		// Read server response(s)
		for {
			resp, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Connection closed")
				return
			}
			fmt.Print("Server: " + resp)
			if len(resp) < 4 {
				continue
			}
			// Responses start with 3-digit code and a space means last line
			if resp[3] == ' ' {
				break
			}
		}

		if cmdUpper == "QUIT" {
			break
		}
	}
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}
