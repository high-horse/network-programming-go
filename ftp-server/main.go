package main

import (
    "flag"
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
