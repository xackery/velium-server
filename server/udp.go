package server

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// udpClient represents a connected UDP clien	t.
type udpClient struct {
	conn       *net.UDPConn
	addr       *net.UDPAddr
	identifier string
	lastSeen   time.Time
	clientKey  string
	session    uuid.UUID
}

// udpServer represents the UDP server.
type udpServer struct {
	clients     map[string]*udpClient
	clientsLock sync.Mutex
}

func newUDPServer() (*udpServer, error) {
	return &udpServer{
		clients: make(map[string]*udpClient),
	}, nil
}

func (c *udpClient) Identity() string {
	if c.identifier == "" {
		return c.clientKey
	}

	return c.identifier + "@" + c.clientKey
}

// handleClient handles incoming client messages and manages client sessions.
func (s *udpServer) handleClient(clientAddr *net.UDPAddr, data []byte) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	clientKey := clientAddr.String()
	client, exists := s.clients[clientKey]

	if !exists {
		// New client connection
		fmt.Printf("[udp] New client connected: %s\n", clientKey)
		conn, err := net.DialUDP("udp", nil, clientAddr)
		if err != nil {
			fmt.Printf("[udp] Error creating client connection: %v\n", err)
			return
		}

		client = &udpClient{
			conn:      conn,
			addr:      clientAddr,
			lastSeen:  time.Now(),
			clientKey: clientKey,
		}
		s.clients[clientKey] = client
	}

	fmt.Printf("[udp] From %s: 0x%x %s\n", client.Identity(), data[0], string(data[1:]))
	s.parseMessage(client, data)
	client.lastSeen = time.Now()
}

// cleanupDisconnectedClients removes clients that haven't sent messages in a while.
func (s *udpServer) cleanupDisconnectedClients() {
	for {
		time.Sleep(10 * time.Second)

		s.clientsLock.Lock()
		now := time.Now()
		for clientKey, client := range s.clients {
			if now.Sub(client.lastSeen) > 30*time.Second {
				fmt.Printf("[udp] Client %s disconnected (inactive)\n", clientKey)
				client.conn.Close()
				delete(s.clients, clientKey)
			}
		}
		s.clientsLock.Unlock()
	}
}

// parseMessage parses a message from a client.
func (s *udpServer) parseMessage(client *udpClient, data []byte) {
	err := s.parseMessageInternal(client, data)
	if err != nil {
		fmt.Printf("[udp] Error parsing message from %s: %v\n", client.Identity(), err)
	}
}

// parseMessageInternal parses a message from a client.
func (s *udpServer) parseMessageInternal(client *udpClient, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("invalid message (empty))")
	}

	args := strings.Split(string(data[1:]), " ")
	if len(args) == 0 {
		return fmt.Errorf("invalid message")
	}

	switch data[0] {
	case opIdentify:
		fmt.Printf("[udp] Ignoring opIdentify from %s\n", client.Identity())
		return nil
	case opSession:
		var err error
		if len(data) != 17 {
			return fmt.Errorf("invalid opSession message (%d bytes, wanted 17)", len(data))
		}

		session := uuid.UUID{}
		copy(session[:], data[1:17])

		shim := Shim(session)
		if shim == nil {
			fmt.Printf("[udp] Error finding session: %v\n", err)
			_, err = client.conn.Write([]byte{opQuit, '\n'})
			if err != nil {
				fmt.Printf("[udp] Error writing quit: %v\n", err)
			}
		}

		client.session = session
		fmt.Printf("[udp] Client %s identified, session UUID: %s\n", client.Identity(), client.session.String())
		return nil
	case opPing:
		fmt.Printf("[udp] Ignoring opPing from %s\n", client.Identity())
		return nil
	case opCommand:
		fmt.Printf("[udp] Ignoring opCommand from %s\n", client.Identity())
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
