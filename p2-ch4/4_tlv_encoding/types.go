// Listing 4-4: The message struct implements a simple protocol (types.go).
package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	BinaryType uint8 = iota + 1
	StringType
	
	MaxPayloadSize uint32 = 10<<20 // 10 mb
) 

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")

type Payload interface{
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
}

type Binary []byte


// Listing 4-5: Creating the Binary type (types.go)
func(m Binary)Bytes() []byte {return m}
func(m Binary)String() string {return string(m)}

func(m Binary)WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, BinaryType) // 1-byte type
	if err != nil {
		return 0, err
	}
	
	var n int64 = 1
	
	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4-byte size
	if err != nil {
		return 0, err
	}
	
	n+=4
	o, err := w.Write(m)
	
	return n+int64(o), nil
}

// Listing 4-5: Creating the Binary type (types.go)
func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ) // 1-type binary
	if err != nil {
		return 0, err
	}
	
	var n int64 = 1
	if typ != BinaryType {
		return n, errors.New("invalid Binary")
	}
	
	var size uint32
	err = binary.Read(r, binary.BigEndian, &size) // b-byte size
	if err != nil {
		return n, err
	}

	n += 4
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}
	
	*m = make([]byte, size)
	o, err := r.Read(*m)
	
	return n + int64(o), err
}