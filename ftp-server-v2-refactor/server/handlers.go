package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"io"
)

func handleHelpCommand(writer *bufio.Writer) {
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
}

func handleUserCommand(writer *bufio.Writer, arg string) {

	sendLine(writer, fmt.Sprintf("331 User %s ok, need password", arg))
}

func handleCwdCommand(writer *bufio.Writer, arg string, rootDir string, currentDir *string) {
	if arg == "" {
		sendLine(writer, "501 Missing directory")
		return
	}

	// Normalize rootDir to absolute ONCE here:
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		sendLine(writer, "550 Server path error")
		return
	}

	// Build & normalize target path
	newPath, err := filepath.Abs(filepath.Join(*currentDir, arg))
	if err != nil {
		sendLine(writer, "550 Invalid path")
		return
	}

	// Prevent leaving root
	if !strings.HasPrefix(newPath, absRoot) {
		sendLine(writer, "550 Access denied")
		return
	}

	info, err := os.Stat(newPath)
	if err != nil || !info.IsDir() {
		sendLine(writer, "550 Not a directory")
		return
	}

	*currentDir = newPath
	sendLine(writer, "250 Directory successfully changed")
}

func handleCdupCommand(writer *bufio.Writer, rootDir string, currentDir *string) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		sendLine(writer, "550 Server path error")
		return
	}

	// Compute parent
	parent := filepath.Dir(*currentDir)

	parent, err = filepath.Abs(parent)
	if err != nil {
		sendLine(writer, "550 Invalid path")
		return
	}

	// Prevent escape above root
	if !strings.HasPrefix(parent, absRoot) {
		sendLine(writer, "550 Access denied")
		return
	}

	*currentDir = parent
	sendLine(writer, "200 Command okay")
}

func handleListCommand(writer *bufio.Writer, currentDir string, dataListener *net.Listener) {
	if *dataListener == nil {
		sendLine(writer, "425 Use PASV first")
		return
	}

	sendLine(writer, "150 Here comes the directory listing")

	dataConn, err := (*dataListener).Accept()
	if err != nil {
		sendLine(writer, "425 Can't open data connection")
		(*dataListener).Close()
		dataListener = nil
		return
	}

	// Now send the directory listing over dataConn
	files, err := os.ReadDir(currentDir)
	if err != nil {
		sendLine(writer, "550 Failed to list directory")
		dataConn.Close()
		(*dataListener).Close()
		*dataListener = nil
		return
	}

	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			continue
		}
		itemType := "File"
        if info.IsDir() {
            itemType = "Folder"
        }
		sizeStr := humanReadableSize(info.Size())


		modTime := info.ModTime().Format("Jan _2 15:04")
		line := fmt.Sprintf("%-6s %-10s %-20s %s\r\n", itemType, sizeStr, modTime, f.Name())
        dataConn.Write([]byte(line))

		// line := fmt.Sprintf("%s %12d %s %s\r\n", fileModeToStr(info.Mode()), info.Size(), modTime, f.Name())
		// dataConn.Write([]byte(line))
	}

	dataConn.Close()
	(*dataListener).Close()
	*dataListener = nil

	sendLine(writer, "226 Directory send OK")
}


func handleRetrCommand(writer *bufio.Writer, arg string, currentDir string, dataListener *net.Listener) {
	if arg == "" {
		sendLine(writer, "501 Syntax error in parameters or arguments")
		return
	}
	if (*dataListener) == nil {
		sendLine(writer, "425 Use PASV first")
		return
	}

	filepath := filepath.Join(currentDir, arg)
	f, err := os.Open(filepath)
	if err != nil {
		sendLine(writer, "550 File not found")
		return
	}
	defer f.Close()

	sendLine(writer, "150 Opening data connection for file transfer")

	dataConn, err := (*dataListener).Accept()
	if err != nil {
		sendLine(writer, "425 Can't open data connection")
		(*dataListener).Close()
		(*dataListener) = nil
		return
	}

	io.Copy(dataConn, f)
	dataConn.Close()
	(*dataListener).Close()
	(*dataListener) = nil

	sendLine(writer, "226 Transfer complete")
}