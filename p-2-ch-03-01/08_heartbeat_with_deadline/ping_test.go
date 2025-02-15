// Listing 3-12: Receiving data advances the deadline (ping_test.go)
package main

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestPingerAdvancedDeadline(t *testing.T){
	done := make(chan struct{})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	
	begin := time.Now()
	go func(){
		defer func (){
			close(done)
		}()
		
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		
		ctx, cancel := context.WithCancel(context.Background())
		defer func(){
			cancel()
			conn.Close()
		}()
		
		resetTimer := make(chan time.Duration, 1)
		resetTimer <- time.Second
		
		go Pinger(ctx, conn, resetTimer)
		
		if err := conn.SetDeadline(time.Now().Add(time.Second * 5)); err != nil {
			t.Error(err)
			return
		}
		
		buf := make([]byte, 1024)
		for{
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			
			t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
			
			resetTimer <- 0
			if err = conn.SetDeadline(time.Now().Add(time.Second * 5)); err !=nil {
				t.Error(err)
				return
			}
		}
	}()
	
	conn, err := net.Dial(listener.Addr().Network(), listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	
	buf := make([]byte, 1024)
	for i := 0; i<4 ; i++ { // read up to four pings
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	_, err = conn.Write([]byte("PONG!!!"))
	if err != nil {
		t.Fatal(err)
	}
	
	for i := 1; i < 4; i++ {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}
	
	<-done
	end := time.Since(begin).Truncate(time.Second)
	t.Logf("[%s] done",end)
	
	if end != time.Second * 9 {
		t.Fatalf("expected EOF at 9 seconds; actual %s", end)
	}
}