package chat

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Embiggenerd/spiritio/pkg/sfu"
	"github.com/gorilla/websocket"
)

type WebsocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	WebsocketService *WebsocketService
	SFU              *sfu.SFU
	Conn             *sfu.ThreadSafeWriter
	Send             chan []byte
}

// Client is a middleman between the websocket connection and the hub.
// type Client struct {
// 	hub *WebsocketService

// 	// The websocket connection.
// 	conn *websocket.Conn

// 	// Buffered channel of outbound messages.
// 	Send chan []byte
// }

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.

// In pump, split based on events to different endpoints.
func (c *Client) ReadPump() {
	defer func() {
		log.Println("readpump deffered")
		c.WebsocketService.unregister <- c
		c.Conn.Close()
	}()
	// c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		message := &WebsocketMessage{}
		_, raw, err := c.Conn.ReadMessage()
		log.Println("reading message error", err)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("1error: %v", err)
			}
			break
		}
		log.Println("new message pushed to broadcast chan, ", string(raw[:]))
		if err != nil {
			log.Println("2error:", err)
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			log.Println("3error:", err)
			return
		}
		c.WebsocketService.broadcast <- raw
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	log.Println("we got to WritePump")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Println("write pump deferred")
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case raw, ok := <-c.Send:
			message := &WebsocketMessage{}
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				// c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				log.Println("this righ here is what happened")
				c.Conn.WriteJSON(websocket.CloseMessage)
				return
			}

			// w, err := c.Conn.NextWriter(websocket.TextMessage)
			// if err != nil {
			// 	return
			// }

			// if err != nil {
			// 	log.Println(err)
			// 	return
			if err := json.Unmarshal(raw, &message); err != nil {
				log.Println(err)
				return
			}
			// log.Println("new message sent from broadcast chan, ", message)
			// w.Write(raw)
			c.Conn.WriteJSON(message)
			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				c.Conn.WriteJSON(newline)
				c.Conn.WriteJSON(<-c.Send)
			}

			if err := c.Conn.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
// func ServeWs(hub *WebsocketService, w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	client := &Client{hub: hub, conn: conn, Send: make(chan []byte, 256)}
// 	client.hub.register <- client

// 	// Allow collection of memory referenced by the caller by doing all work in
// 	// new goroutines.
// 	go client.writePump()
// 	go client.readPump()
// }
