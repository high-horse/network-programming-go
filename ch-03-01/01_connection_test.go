package ch0301_test

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	
	done := make(chan struct{})
	
	go func() {
		defer func (){ done <- struct{}{}}()
		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}
			
			go func(c net.Conn) {
				defer func() {
					c.Close()
					done <- struct{}{}
				}()
				
				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF{
							t.Log(err)
						}
						t.Log(err)
						return
					}
					
					t.Logf("recieved %q \n", buf[:n])
				}
			}(conn)
		}
	}()
	
	conn, err := net.Dial(listener.Addr().Network(), listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("sending msg")
	conn.Write([]byte("hello"))
	conn.Close()
	<-done
	
	listener.Close()
	<-done
	
}