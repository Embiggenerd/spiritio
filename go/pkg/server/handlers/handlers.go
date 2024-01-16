package handlers

import (
	"log"
	"net/http"

	"github.com/Embiggenerd/spiritio/pkg/chat"
	"github.com/Embiggenerd/spiritio/pkg/sfu"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// type Client struct {
// 	websocketService *chat.WebsocketService
// 	sfu              *sfu.SFU
// 	conn             *websocket.Conn
// 	send             chan []byte
// }

// ServeWs handles websocket requests from the peer.
func ServeWs(websocketService *chat.WebsocketService, sfu *sfu.SFU, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &chat.Client{WebsocketService: websocketService, SFU: sfu, Conn: conn, Send: make(chan []byte, 256)}
	client.WebsocketService.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}
