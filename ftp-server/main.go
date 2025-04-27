package main

import (
    "flag"
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "strings"
)

var (
    isServer = flag.Bool("server", false, "Run as server")
    filePath = flag.String("file", "", "Path to file (send or receive)")
    destPath = flag.String("dest", "", "Destination path (client only)")
    addr     = flag.String("addr", "0.0.0.0:2121", "Address to listen or connect")
)

func main() {
    flag.Parse()

    if *isServer {
        runServer()
    } else {
        runClient()
    }
}

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
        case strings.HasPrefix(input, "QUIT"):
            conn.Write([]byte("221 Goodbye\r\n"))
            return
        default:
            conn.Write([]byte("500 Unknown command\r\n"))
        }
    }
}

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
    // Read file into dest
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
