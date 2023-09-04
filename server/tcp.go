package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/xackery/velium-server/def"
)

// tcpClient represents a connected TCP client.
type tcpClient struct {
	conn       net.Conn
	lastSeen   time.Time
	createdAt  time.Time
	identifier string
	clientKey  string
	session    uuid.UUID
}

// tcpServer represents the TCP server.
type tcpServer struct {
	clients     map[string]*tcpClient
	clientsLock sync.Mutex
}

func newTCPServer() (*tcpServer, error) {
	return &tcpServer{
		clients: make(map[string]*tcpClient),
	}, nil
}

func (c *tcpClient) Identity() string {
	if c.identifier == "" {
		return c.clientKey
	}

	return c.identifier + "@" + c.clientKey
}

// handleClient handles incoming client messages and manages client sessions.
func (s *tcpServer) handleClient(conn net.Conn) {
	clientKey := conn.RemoteAddr().String()
	client := &tcpClient{
		conn:      conn,
		lastSeen:  time.Now(),
		createdAt: time.Now(),
		clientKey: clientKey,
	}

	s.clientsLock.Lock()
	s.clients[clientKey] = client
	s.clientsLock.Unlock()

	defer func() {
		s.clientsLock.Lock()
		delete(s.clients, clientKey)
		s.clientsLock.Unlock()
		conn.Close()
		ShimRemove(client.session)
		fmt.Printf("[tcp] Client %s disconnected (%d remaining shims)\n", client.Identity(), len(s.clients))
	}()

	reader := bufio.NewReader(conn)

	for {
		data, isPrefix, err := reader.ReadLine()
		//n, err := conn.Read(buf)
		if err != nil {
			if err == net.ErrClosed {
				return
			}
			if err == io.EOF {
				return
			}
			fmt.Printf("[tcp] Error reading from client %s: %v\n", client.Identity(), err)
			return
		}
		if isPrefix {
			fmt.Printf("[tcp] Client %s sent a message that was too long\n", client.Identity())
			return
		}

		fmt.Printf("[tcp] From %s: 0x%x %s\n", client.Identity(), data[0], string(data[1:]))
		if client.identifier == "" && time.Since(client.createdAt) > 5*time.Second {
			fmt.Printf("[tcp] Client %s timed out without identifier\n", client.Identity())
			return
		}

		s.parseMessage(client, data)
		client.lastSeen = time.Now()
	}
}

// cleanupDisconnectedClients removes clients that haven't sent messages in a while.
func (s *tcpServer) cleanupDisconnectedClients() {
	for {
		time.Sleep(10 * time.Second)

		s.clientsLock.Lock()
		now := time.Now()
		for clientKey, client := range s.clients {
			if now.Sub(client.lastSeen) > 30*time.Second {
				fmt.Printf("[tcp] Client %s disconnected (inactive)\n", client.Identity())
				client.conn.Close()
				delete(s.clients, clientKey)
			}
		}
		s.clientsLock.Unlock()
	}
}

// parseMessage parses a message from a client.
func (s *tcpServer) parseMessage(client *tcpClient, data []byte) {
	err := s.parseMessageInternal(client, data)
	if err != nil {
		fmt.Printf("[tcp] Error parsing message from %s: %v\n", client.Identity(), err)
	}
}

// parseMessageInternal parses a message from a client.
func (s *tcpServer) parseMessageInternal(client *tcpClient, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("invalid message (empty))")
	}

	args := strings.Split(string(data[1:]), " ")
	if len(args) == 0 {
		return fmt.Errorf("invalid message")
	}

	switch data[0] {
	case opIdentify:
		if len(args) != 1 {
			return fmt.Errorf("invalid opIdentify message (%d args)", len(args)-1)
		}

		if client.identifier != "" {
			fmt.Printf("[tcp] Client %s renamed to %s\n", client.Identity(), args[0])
			client.identifier = args[0]
			return nil
		}

		s.clientsLock.Lock()
		client.identifier = args[0]
		client.session = uuid.New()
		s.clientsLock.Unlock()
		ShimAdd(&def.Shim{
			UUID: client.session,
			Name: client.identifier,
		})

		fmt.Printf("[tcp] Client %s identified, session UUID: %s\n", client.Identity(), client.session.String())
		data := []byte{opSession}
		for _, id := range client.session {
			data = append(data, byte(id))
		}
		data = append(data, '\n')

		_, err := client.conn.Write(data)
		if err != nil {
			return fmt.Errorf("write opIdentify: %w", err)
		}
		return nil
	case opSession:
		fmt.Printf("[tcp] Ignoring opSession from %s\n", client.Identity())
		return nil
	case opPing:
		if len(args) != 1 {
			return fmt.Errorf("invalid PING message")
		}
		_, err := client.conn.Write([]byte("PONG\n"))
		return err
	case opCommand:
		if len(args) < 2 {
			return fmt.Errorf("invalid CMD message")
		}
		_, err := client.conn.Write([]byte("\x04" + strings.Join(args[1:], " ") + "\n"))
		return err
	case opEcho:
		if len(args) < 3 {
			return fmt.Errorf("invalid MSG message")
		}
		recipient := args[1]
		message := strings.Join(args[2:], " ")
		return s.sendMessage(client, recipient, message)
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

// sendMessage sends a message from a client to another client.
func (s *tcpServer) sendMessage(sender *tcpClient, recipient, message string) error {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	for _, client := range s.clients {
		if client.identifier == recipient {
			_, err := client.conn.Write([]byte(fmt.Sprintf("MSG %s %s\n", sender.identifier, message)))
			return err
		}
	}

	return fmt.Errorf("recipient not found")
}
