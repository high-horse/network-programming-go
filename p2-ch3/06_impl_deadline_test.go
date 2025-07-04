// Listing 3-9: A server-enforced deadline terminates the network connection (deadline_test.go).
package ch0301

import (
	"io"
	"net"
	_ "sync"
	"testing"
	"time"
)


func TestDeadline(t *testing.T) {
	sync := make(chan struct{})
	
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		
		defer func(){
			conn.Close()
			close(sync) 	// read from sync shouldn't block due to early return
		}()
		
		// implement deadline
		err = conn.SetDeadline(time.Now().Add(time.Second * 5))
		if err != nil {
			t.Error(err)
			return
		}
		
		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		nErr, ok := err.(net.Error)
		if !ok || !nErr.Timeout() {
			t.Errorf("expected timeout error ; got %v \n", err)
		}
		
		sync <- struct{}{}
		
		// extend deadline by 5 second
		err = conn.SetDeadline(time.Now().Add(time.Second * 5))
		if err != nil {
			t.Error(err)
			return
		}
		
		_, err = conn.Read(buf)
		if err != nil {
			t.Error(err)
		}
	}()
	
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	
	
	<-sync
	_, err = conn.Write([]byte("1"))
	if err != nil {
		t.Fatal(err)
	}
	
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err != nil{
		if err != io.EOF {
			t.Errorf("expected server timeout ; got %v \n", err)
		}
	}
}