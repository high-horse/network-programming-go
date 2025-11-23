package main

import (
	"flag"
	"ftp/client"
	"ftp/server"
)


func main() {
	mode := flag.String("mode", "client", "server or client")
	serverAddr := flag.String("addr", "localhost:2121", "Ip:port of server hosting the file")
	port := flag.String("port", ":2121", "Port to host")
	sharedDir := flag.String("dir", "./", "Directory you want to share vis FTP")
	flag.Parse()

	if *mode == "server" {
		server.StartServer(*sharedDir, port)
	} else {
		client.StartClient(serverAddr)
	}
}
