package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "strings"
)

func runClient() {
    conn, err := net.Dial("tcp", *addr)
    if err != nil {
        log.Fatalf("Connection error: %v", err)
    }
    defer conn.Close()

    // Read welcome message
    readResponse(conn)

    sendCommand(conn, "USER anonymous")
    readResponse(conn)

    sendCommand(conn, "PASS anonymous")
    readResponse(conn)

    sendCommand(conn, "TYPE I")
    readResponse(conn)

    sendCommand(conn, "RETR " + *filePath)
    readResponse(conn) // <-- Important: Read the "150 Opening data connection" response

    destFilePath := *destPath
    if destFilePath == "." {
        // If destination is ".", save with the same name as source
        parts := strings.Split(*filePath, "/")
        destFilePath = parts[len(parts)-1]
    }

    dest, err := os.Create(destFilePath)
    if err != nil {
        log.Fatalf("Cannot create destination file: %v", err)
    }
    defer dest.Close()

    progressWriter := &ProgressWriter{Writer: dest}
    _, err = io.Copy(progressWriter, conn)
    if err != nil {
        log.Fatalf("Download error: %v", err)
    }

    fmt.Println("\nFile downloaded successfully to", destFilePath)

    sendCommand(conn, "QUIT")
    readResponse(conn)
}

type ProgressWriter struct {
    Writer io.Writer
    total  int64
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
    n, err := pw.Writer.Write(p)
    if err == nil {
        pw.total += int64(n)
        mb := float64(pw.total) / (1024 * 1024)
        fmt.Printf("\rDownloaded: %.2f MB", mb)
    }
    return n, err
}

func sendCommand(conn net.Conn, cmd string) {
    _, err := conn.Write([]byte(cmd + "\r\n"))
    if err != nil {
        log.Fatalf("Send command error: %v", err)
    }
}

func readResponse(conn net.Conn) {
    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err != nil {
        log.Fatalf("Read response error: %v", err)
    }
    response := strings.TrimSpace(string(buf[:n]))
    fmt.Println("Server:", response)
}
