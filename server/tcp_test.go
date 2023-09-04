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

func TestTCP(t *testing.T) {
	if os.Getenv("SINGLE_TEST") != "1" {
		t.Skip("Skipping single test")
	}
	conn, err := net.Dial("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("\x01xackery\n"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Sent identify")

	_, err = conn.Write([]byte("\x04peq.ecommons.bob /echo hi\n"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Sent comand")

	dataChan := make(chan []byte, 1000)
	go func() {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			t.Logf("Error reading TCP data: %s", err)
		}
		dataChan <- buf[:n]
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("timeout")
		case data := <-dataChan:
			if data[0] != opSession {
				t.Fatalf("expected opSession, got %d", data[0])
			}

			id, err := uuid.FromBytes(data[1:17])
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("Received UUID: %s", id.String())
		}
	}
}
