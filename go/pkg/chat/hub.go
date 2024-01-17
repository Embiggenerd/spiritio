package chat

import (
	"log"

	"github.com/Embiggenerd/spiritio/pkg/config"
)

// WebsocketService maintains the set of active clients and broadcasts messages to the
// clients.
type WebsocketService struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewWebsocketService(cfg *config.Config) *WebsocketService {
	return &WebsocketService{
		broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (w *WebsocketService) Run() {
	for {
		select {
		case client := <-w.Register:
			w.clients[client] = true
		case client := <-w.unregister:
			if _, ok := w.clients[client]; ok {
				log.Println("client unregistered")

				delete(w.clients, client)
				close(client.Send)
			}
		case message := <-w.broadcast:
			log.Println("message was receiced from boradcast", string(message))
			for client := range w.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(w.clients, client)
				}
			}
		}
	}
}
