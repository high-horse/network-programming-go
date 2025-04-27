package main

import (
	"bufio"
	"net"
	"reflect"
	"testing"
)

const payload  = "The bigger the interface, the weaker the abstraction."

func TestScanner(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		
		_, err = conn.Write([]byte(payload))
		if err != nil {
			t.Error(err)
		}
	}()
	
	conn, err := net.Dial(listener.Addr().Network(), listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	
	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanWords)
	
	var words []string
	for scanner.Scan(){
		words = append(words, scanner.Text())
	}
	
	expected := []string{"The", "bigger", "the", "interface,", "the", "weaker", "the", "abstraction."}
	
	if !reflect.DeepEqual(words, expected) {
		t.Fatal("inaccurate  scanned word list  ", words)
	}
	
	t.Logf("scanned word: %#v", words)
}