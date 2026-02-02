package server

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var waiters = map[string]chan *websocket.Conn{}
var waitersMutex = &sync.Mutex{}

var clients = map[string]*websocket.Conn{}
var clientsMutex = &sync.Mutex{}

// HTTP handler for WebSocket server
func WebsocketHandler(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}

	clientID := Register(req, socket)
	if clientID == "" {
		return
	}

	defer func() {
		clientsMutex.Lock()

		delete(clients, clientID)

		clientsMutex.Unlock()
		socket.Close()
	}()

	keepAlive(socket)

	for {
		msgType, msg, err := socket.ReadMessage()
		if err != nil {
			return
		}
		_ = socket.WriteMessage(msgType, msg)
	}
}

func GetClient(id string) *websocket.Conn {
	return clients[id]
}

// Register client by using the HTTP request with `?client_id=abc`
func Register(req *http.Request, socket *websocket.Conn) string {
	clientID := req.URL.Query().Get("client_id")
	
	if clientID == "" {
		socket.Close()
		return ""
	}

	clientsMutex.Lock()
	clients[clientID] = socket
	clientsMutex.Unlock()

	waitersMutex.Lock()
	ch, ok := waiters[clientID]
	if ok {
		select {
		case ch <- socket:
		default:
		}

		close(ch)
		delete(waiters, clientID)
	}
	waitersMutex.Unlock()

	return clientID
}

// Wait for client with id to connect
func WaitForClient(id string, timeout time.Duration) (*websocket.Conn, error) {
    clientsMutex.Lock()
	conn, ok := clients[id]
    if ok {
        clientsMutex.Unlock()
        return conn, nil
    }
    clientsMutex.Unlock()

    waitCh := make(chan *websocket.Conn, 1)

    waitersMutex.Lock()
    waiters[id] = waitCh
    waitersMutex.Unlock()

    select {
    case conn := <-waitCh:
        return conn, nil
    case <-time.After(timeout):
        waitersMutex.Lock()
        delete(waiters, id)
        waitersMutex.Unlock()

        return nil, errors.New("Timed out waiting for client")
    }
}

func GenerateID() string {
	return uuid.New().String()
}

func keepAlive(socket *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)

	go func() {
		for range ticker.C {
			err := socket.WriteControl(
				websocket.PingMessage,
				[]byte("keepalive"),
				time.Now().Add(5 * time.Second),
			)
			if err != nil {
				ticker.Stop()
				socket.Close()

				return
			}
		}
	}()
}