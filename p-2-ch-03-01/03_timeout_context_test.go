package ch0301_test

import (
	"context"
	"net"
	"testing"
	"time"
	"syscall"
)

func TestDialContext(t *testing.T){
	dl := time.Now().Add(time.Second * 5)
	ctx, cancel := context.WithDeadline(context.Background(), dl)
	defer cancel()
	
	var d net.Dialer // DialContext is a method on a Dialer
	d.Control = func(_ , _ string, _ syscall.RawConn) error {
		// Sleep long enough to reach the context's deadline.
		time.Sleep(5*time.Second + time.Millisecond)
		return nil
	}
	
	conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")
	if err == nil {
		conn.Close()
		t.Fatal("connection did not timed out ")
	}
	
	nErr, ok := err.(net.Error)
	if !ok {
		t.Error(err)
	} else {
		if !nErr.Timeout(){
			t.Errorf("err is not timeout : %v \n", err)
		}
	}
	
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
	}
}