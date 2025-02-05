package ch0301

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDialContextFanOut(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second *10))
	
	listener, err := net.Listen("tcp","127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	
	go func() {
		// only accepting a single connection
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()
	
	dial := func(ctx context.Context, address string, response chan int, id int, wg *sync.WaitGroup) {
		defer wg.Done()
		
		var d net.Dialer
		c, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			return
		}
		c.Close()
		
		select {
			case <- ctx.Done():
			case response <- id:
		}
	}
	
	res := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i<10; i++ {
		wg.Add(1)
		go dial(ctx, listener.Addr().String(), res, i+1, &wg)
	}
	
	response := <-res
	cancel()
	wg.Wait()
	close(res)
	
	if ctx.Err() != context.Canceled {
		t.Errorf("expected context cancelled ; got %v\n", ctx.Err())
	}
	
	t.Logf("dialer %d retrived the resource ", response)
}