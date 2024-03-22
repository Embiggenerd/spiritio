package rooms

import (
	"context"
	"time"

	"github.com/Embiggenerd/spiritio/pkg/config"
	"github.com/Embiggenerd/spiritio/pkg/sfu"
	"github.com/Embiggenerd/spiritio/pkg/utils"
	"github.com/Embiggenerd/spiritio/pkg/websocketClient"
	"github.com/Embiggenerd/spiritio/types"
	"github.com/pion/webrtc/v4"
)

type Room interface {
	Build(ctx context.Context, cfg *config.Config)
	AddPeerConnection(pc *webrtc.PeerConnection, w *websocketClient.ThreadSafeWriter)
	BroadcastEvent(event *types.Event)
	AddVisitor(visitor *Visitor)
}

type ChatRoom struct {
	Service  *ChatRoomsService `gorm:"-:all"`
	ID       uint              `gorm:"primaryKey"`
	SFU      sfu.SFU           `gorm:"-:all"`
	ChatLog  []ChatRoomLog     `gorm:"-:all"`
	Visitors []*Visitor        `gorm:"-:all"`
}

func (r *ChatRoom) AddPeerConnection(pc *webrtc.PeerConnection, w *websocketClient.ThreadSafeWriter) {
	r.SFU.AddPeerConnection(pc, w)
}

func (r *ChatRoom) Build(ctx context.Context, service *ChatRoomsService) {
	if r.ChatLog == nil {
		r.ChatLog = []ChatRoomLog{}
	}

	r.Service = service

	r.SFU = sfu.NewSelectiveForwardingUnit()
	go func() {
		for range time.NewTicker(time.Second * 3).C {
			r.SFU.DispatchKeyFrame()
		}
	}()
}

func (r *ChatRoom) BroadcastEvent(event *types.Event) {
	for _, v := range r.Visitors {
		v.Client.Writer.WriteJSON(event)
	}
}

func (r *ChatRoom) AddVisitor(visitor *Visitor) {

	r.Visitors = append(r.Visitors, visitor)
}

func (r *ChatRoom) CreateUniqueDisplayName() string {
	return r.untilUnique(utils.RandName())
}

func (r *ChatRoom) untilUnique(name string) string {
	unique := true
	for _, v := range r.Visitors {
		if v.User != nil {
			if name == v.User.Name {
				unique = false
				break
			}
		}
	}

	if name == "" || !unique {
		return r.untilUnique(utils.RandName())
	}
	return name
}
