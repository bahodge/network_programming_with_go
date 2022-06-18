package ch03

import (
	"io"
	"net"
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

		// After the function is done close all the things
		defer func() {
			conn.Close()
			close(sync) // read from sync shouldn't block due to early return
		}()

		// Set a deadline for the connection
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		// make a new slice with a length of 1
		buf := make([]byte, 1)
		_, err = conn.Read(buf) // blocked until remote node sends data

		// Wait until timeout
		nErr, ok := err.(net.Error)
		if !ok || !nErr.Timeout() {
			t.Errorf("Expected timeout error; actual: %v", err)
		}

		sync <- struct{}{}

		// Extend the deadline by another 5 seconds
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		// Block with read
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
	if err != io.EOF {
		t.Errorf("Expected server termination; actual %v", err)
	}

}
