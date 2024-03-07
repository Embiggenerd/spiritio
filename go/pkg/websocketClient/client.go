package websocketClient

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
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
	message := &types.WebsocketMessage{
		Data: v,
	}
	switch v.(type) {
	case *types.Event:
		message.Type = "event"
	case *types.Question:
		message.Type = "question"
	default:
		// fmt.Println("type&&", ty)
	}
	fmt.Println("type&", reflect.TypeOf(v))
	t.log.LogMessageSent(reqID.(string), mdJSON, message)

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
	RoomID  uint     `json:"room_id,omitempty"`
	ChatLog []string `json:"chat_log,omitempty"`
	Name    string   `json:"name,omitempty"`
}
