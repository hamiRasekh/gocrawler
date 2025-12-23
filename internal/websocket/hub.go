package websocket

import (
	"encoding/json"
	"sync"
)

type Message struct {
	Type    string      `json:"type"`
	TaskID  *int64      `json:"task_id,omitempty"`
	Level   string      `json:"level,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type Client struct {
	ID       string
	TaskID   *int64 // Filter logs by task ID (nil = all tasks)
	UserID   int64  // User ID for authentication
	Username string // Username for authentication
	Send     chan []byte
	Hub      *Hub
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Filter by task ID if client has a filter
				if client.TaskID != nil {
					var msg Message
					if err := json.Unmarshal(message, &msg); err == nil {
						if msg.TaskID == nil || *msg.TaskID != *client.TaskID {
							continue
						}
					}
				}
				
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Broadcast(message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		return
	}
	
	select {
	case h.broadcast <- data:
	default:
	}
}

func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

