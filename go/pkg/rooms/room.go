package rooms

import (
	"time"

	"github.com/Embiggenerd/spiritio/pkg/sfu"
	"github.com/Embiggenerd/spiritio/pkg/websocketClient"
	"github.com/pion/webrtc/v3"
)

type Room interface {
	Build()
	AddPeerConnection(pc *webrtc.PeerConnection, w *websocketClient.ThreadSafeWriter)
}

type ChatRoom struct {
	ID      uint            `gorm:"primaryKey"`
	SFU     *sfu.SFUService `gorm:"-:all"`
	ChatLog *[]ChatRoomLog  `gorm:"-:all"`
}

func (r *ChatRoom) AddPeerConnection(pc *webrtc.PeerConnection, w *websocketClient.ThreadSafeWriter) {
	r.SFU.ListLock.Lock()
	r.SFU.PeerConnections = append(r.SFU.PeerConnections, sfu.PeerConnectionState{PeerConnection: pc, Websocket: w})
	r.SFU.ListLock.Unlock()
}

func (r *ChatRoom) Build() {
	if r.ChatLog == nil {
		r.ChatLog = &[]ChatRoomLog{}
	}

	r.SFU = sfu.NewSelectiveForwardingUnit()
	go func() {
		for range time.NewTicker(time.Second * 3).C {
			r.SFU.DispatchKeyFrame()
		}
	}()
}
