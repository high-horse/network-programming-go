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

    reader := make([]byte, 1024)
    conn.Read(reader) // Read welcome message

    sendCommand(conn, "USER anonymous")
    readResponse(conn)

    sendCommand(conn, "PASS anonymous")
    readResponse(conn)

    sendCommand(conn, "TYPE I")
    readResponse(conn)

    sendCommand(conn, "RETR " + *filePath)

    dest, err := os.Create(*destPath)
    if err != nil {
        log.Fatalf("Cannot create destination file: %v", err)
    }
    defer dest.Close()

    io.Copy(dest, conn)
    fmt.Println("File downloaded successfully to", *destPath)

    sendCommand(conn, "QUIT")
}

func sendCommand(conn net.Conn, cmd string) {
    conn.Write([]byte(cmd + "\r\n"))
}

func readResponse(conn net.Conn) {
    buf := make([]byte, 1024)
    conn.Read(buf)
    fmt.Println("Server:", strings.TrimSpace(string(buf)))
}
