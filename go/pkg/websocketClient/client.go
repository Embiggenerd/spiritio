package websocketClient

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebsocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

// WebsocketClient maintains the set of active clients and broadcasts messages to the
// clients.
type WebsocketClient struct {
	Conn   *websocket.Conn
	Writer *ThreadSafeWriter
}

func New(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*WebsocketClient, error) {
	unsafeConn, err := upgrader.Upgrade(w, r, responseHeader)
	writer := &ThreadSafeWriter{unsafeConn, sync.Mutex{}}
	client := &WebsocketClient{
		Conn:   writer.Conn,
		Writer: writer,
	}
	return client, err
}

// // CreateWebsocketConnectionWriter upgrades a connection, wraps it in a lockable
// func (wss *WebsocketClient) CreateWebsocketConnectionWriter(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*ThreadSafeWriter, error) {
// 	unsafeConn, err := upgrader.Upgrade(w, r, responseHeader)
// 	writer := &ThreadSafeWriter{unsafeConn, sync.Mutex{}}
// 	return writer, err
// }

func (t *ThreadSafeWriter) WriteJSON(v interface{}) error {
	t.Lock()
	defer t.Unlock()

	return t.Conn.WriteJSON(v)
}

type ThreadSafeWriter struct {
	*websocket.Conn
	sync.Mutex
}

type websocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}
