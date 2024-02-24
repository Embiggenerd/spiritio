package websocketClient

import (
	"context"
	"net/http"
	"sync"

	"github.com/Embiggenerd/spiritio/pkg/logger"
	"github.com/Embiggenerd/spiritio/pkg/utils"
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

	md := utils.ExposeContextMetadata(t.ctx)
	mdJSON := md.ToJSON()
	reqID, _ := md.Get("requestID")
	t.log.LogEventSent(reqID.(string), mdJSON, v.(*types.WebsocketMessage))

	return t.Conn.WriteJSON(v)
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
	RoomID  uint     `json:"room_id,omitempty"`
	ChatLog []string `json:"chat_log,omitempty"`
}
