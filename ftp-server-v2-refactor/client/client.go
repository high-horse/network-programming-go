package client

import (
	"bufio"
	"fmt"
	"ftp/common"
	"io"
	"net"
	"os"
	"strings"
)

func StartClient_(host string, port int, timeout int, verbose bool) {
	if verbose {
		println("Verbose logging enabled")
	}
	println("Starting client to connect to host:", host, "on port:", port)
	println("Timeout set to:", timeout, "seconds")
	// Placeholder for actual client logic
}


func StartClient(serverAddr *string) {
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
			p1 := common.Atoi(addrParts[4])
			p2 := common.Atoi(addrParts[5])
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
