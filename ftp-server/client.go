package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "os"
    "strings"
    "time"
)

func runClient() {
    conn, err := net.Dial("tcp", *addr)
    if err != nil {
        log.Fatalf("Connection error: %v", err)
    }
    defer conn.Close()

    readResponse(conn)

    sendCommand(conn, "USER anonymous")
    readResponse(conn)

    sendCommand(conn, "PASS anonymous")
    readResponse(conn)

    sendCommand(conn, "TYPE I")
    readResponse(conn)

    sendCommand(conn, "RETR " + *filePath)
    readResponse(conn)

    destFilePath := *destPath
    if destFilePath == "." {
        parts := strings.Split(*filePath, "/")
        destFilePath = parts[len(parts)-1]
    }

    dest, err := os.Create(destFilePath)
    if err != nil {
        log.Fatalf("Cannot create destination file: %v", err)
    }
    defer dest.Close()

    progressWriter := &ProgressWriter{
        Writer: dest,
        start:  time.Now(),
    }

    _, err = io.Copy(progressWriter, conn)
    if err != nil {
        log.Fatalf("Download error: %v", err)
    }

    fmt.Printf("\nFile downloaded successfully to %s\n", destFilePath)

    sendCommand(conn, "QUIT")
    readResponse(conn)
}

type ProgressWriter struct {
    Writer io.Writer
    total  int64
    start  time.Time
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
    n, err := pw.Writer.Write(p)
    if err == nil {
        pw.total += int64(n)
        elapsed := time.Since(pw.start).Seconds()
        speedMBs := float64(pw.total) / (1024 * 1024) / elapsed

        downloadedMB := float64(pw.total) / (1024 * 1024)
        eta := "-"
        if speedMBs > 0 {
            estimatedTotalSeconds := float64(pw.total) / (speedMBs * 1024 * 1024)
            etaDuration := time.Duration(estimatedTotalSeconds-float64(pw.start.Second())) * time.Second
            eta = etaDuration.Round(time.Second).String()
        }

        fmt.Printf("\rDownloaded: %.2f MB | Speed: %.2f MB/s | ETA: %s", downloadedMB, speedMBs, eta)
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
