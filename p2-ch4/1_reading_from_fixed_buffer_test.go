// Listing 4-1: Receiving data over a network connection ( read_test.go)
// Reading Data into a Fixed Buffer
package main

import (
	"crypto/rand"
	"io"
	"net"
	"testing"
)

func TestReadIntoBuffer(t *testing.T) {
	payload := make([]byte, 1<<24) // 16mb
	_, err := rand.Read(payload) // generate random payload
	if err != nil {
		t.Fatal(err)
	}
	
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	
	go func(){
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		defer conn.Close()
		
		_, err = conn.Write(payload)
		if err != nil {
			t.Error(err)
		}
	}()
	
	conn, err := net.Dial(listener.Addr().Network(), listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	
	buf := make([]byte, 1<<19) // 512 kb
	for{
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF{
				t.Error(err)
			}
			break
		}
		
		t.Logf("Read %d bytes ", n)
		// t.Log(buf[:n])
	}
} 