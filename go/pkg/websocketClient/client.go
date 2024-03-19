package websocketClient

import (
	"context"
	"net/http"
	"sync"

	"github.com/Embiggenerd/spiritio/pkg/logger"
	"github.com/Embiggenerd/spiritio/types"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WebsocketClient exposes a websocket connection and a custom writer to write to the conneciton
type WebsocketClient struct {
	Conn   *websocket.Conn
	Writer *ThreadSafeWriter
}

// New upgrades an http connection to ws and
func New(ctx context.Context, log logger.Logger, w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*WebsocketClient, error) {
	unsafeConn, err := upgrader.Upgrade(w, r, responseHeader)
	writer := &ThreadSafeWriter{
		unsafeConn,
		sync.Mutex{},
		ctx,
		log,
	}
	client := &WebsocketClient{
		Conn:   writer.Conn,
		Writer: writer,
	}
	return client, err
}

func (t *ThreadSafeWriter) WriteJSON(v interface{}) error {
	t.Lock()
	defer t.Unlock()

	message := &types.WebsocketMessage{
		Data: v,
	}
	switch v.(type) {
	case *types.Event:
		message.Type = "event"
	case *types.Question:
		message.Type = "question"
	}
	t.log.LogMessageSent(t.ctx, message)

	return t.Conn.WriteJSON(message)
}

type ThreadSafeWriter struct {
	*websocket.Conn
	sync.Mutex
	ctx context.Context
	log logger.Logger
}

type JoinRoomWebsocketMessage struct {
	Event string       `json:"event"`
	Data  JoinRoomData `json:"data"`
}

type JoinRoomData struct {
	RoomID  uint      `json:"room_id,omitempty"`
	ChatLog []ChatLog `json:"chat_log"`
	Name    string    `json:"name,omitempty"`
}

type ChatLog struct {
	Text string
	From string
}
