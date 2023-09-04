package server

import (
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/xackery/velium-server/def"
)

var (
	// mu protects the server singleton
	mu sync.RWMutex
	// server is a singleton for all server instances
	serverInstance *server
)

// server wraps both tcp and udp server instances
type server struct {
	tcpServer *tcpServer
	udpServer *udpServer
	shims     map[uuid.UUID]*def.Shim
}

// Start initializes the server singleton
func Start() error {
	var err error
	if serverInstance != nil {
		return fmt.Errorf("server already loaded")
	}

	mu.Lock()
	serverInstance = &server{
		shims: make(map[uuid.UUID]*def.Shim),
	}
	mu.Unlock()

	serverInstance.udpServer, err = newUDPServer()
	if err != nil {
		return fmt.Errorf("new udp server: %w", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ":12345")
	if err != nil {
		return fmt.Errorf("resolve udp address: %w", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer conn.Close()

	fmt.Println("Listening on udp :12345")

	go serverInstance.udpServer.cleanupDisconnectedClients()

	go func() {
		// Handle incoming UDP messages
		buf := make([]byte, 1024)
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Printf("Error reading UDP data: %v\n", err)
				continue
			}

			go serverInstance.udpServer.handleClient(addr, buf[:n])
		}
	}()

	serverInstance.tcpServer, err = newTCPServer()
	if err != nil {
		return fmt.Errorf("new tcp server: %w", err)
	}

	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		return fmt.Errorf("tcp net listen: %w", err)
	}
	defer listener.Close()

	fmt.Println("Listening on tcp :12345")

	go serverInstance.tcpServer.cleanupDisconnectedClients()

	// Accept incoming client connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting client connection: %v\n", err)
			continue
		}

		go serverInstance.tcpServer.handleClient(conn)
	}

}

func ShimAdd(shim *def.Shim) {
	mu.Lock()
	defer mu.Unlock()

	serverInstance.shims[shim.UUID] = shim
}

func Shim(uuid uuid.UUID) *def.Shim {
	mu.RLock()
	defer mu.RUnlock()

	return serverInstance.shims[uuid]
}

func ShimRemove(uuid uuid.UUID) {
	mu.Lock()
	defer mu.Unlock()

	delete(serverInstance.shims, uuid)
}

func ShimCount() int {
	mu.RLock()
	defer mu.RUnlock()

	return len(serverInstance.shims)
}
