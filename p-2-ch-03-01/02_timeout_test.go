package ch0301_test

import (
	"net"
	"testing"
	"time"
	"syscall"
)


func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer{
		Control: func(_, address string, _ syscall.RawConn) error {
			return &net.DNSError{
				Err: "connection timed out",
				Name: address,
				Server: "127.0.0.1",
				IsTimeout: true,
				IsTemporary: true,
			}
		},
		Timeout: timeout,
	}
	return d.Dial(network, address)
}

func TestDialTimeout(t *testing.T) {
	c, err := DialTimeout("tcp", "10.0.0.1:http", 5*time.Second)
	if err == nil {
		c.Close()
		t.Fatal("connection did not timed out")
	}
	
	nErr, ok := err.(net.Error)
	if !ok {
		t.Fatal(err)
	}
	
	if !nErr.Timeout() {
		t.Fatal("error is not timeout")
	}
}