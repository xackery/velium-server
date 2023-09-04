package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUDP(t *testing.T) {
	if os.Getenv("SINGLE_TEST") != "1" {
		t.Skip("Skipping single test")
	}
	tcpConn, err := net.Dial("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}
	defer tcpConn.Close()

	_, err = tcpConn.Write([]byte("\x01xackery\n"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Sent identify")

	dataChan := make(chan []byte, 1000)
	go func() {
		buf := make([]byte, 1024)
		n, err := tcpConn.Read(buf)
		if err != nil {
			t.Logf("Error reading TCP data: %s", err)
		}
		dataChan <- buf[:n]
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	var session uuid.UUID
	select {
	case <-ctx.Done():
		t.Fatal("timeout")
	case data := <-dataChan:
		if data[0] != opSession {
			t.Fatalf("expected opSession, got %d", data[0])
		}

		session, err = uuid.FromBytes(data[1:17])
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("Received UUID: %s", session.String())
		break
	}

	conn, err := net.Dial("udp", "127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	data := []byte{opSession}
	for _, id := range session {
		data = append(data, byte(id))
	}

	fmt.Printf("Sending session: %v\n", data)

	_, err = conn.Write(data)
	if err != nil {
		t.Fatal(err)
	}

}
