package rooms

import (
	"time"

	"github.com/Embiggenerd/spiritio/pkg/sfu"
	"github.com/Embiggenerd/spiritio/pkg/users"
	"github.com/Embiggenerd/spiritio/pkg/utils"
	"github.com/Embiggenerd/spiritio/pkg/websocketClient"
	"github.com/Embiggenerd/spiritio/types"
	"github.com/pion/webrtc/v3"
)

type Room interface {
	Build()
	AddPeerConnection(pc *webrtc.PeerConnection, w *websocketClient.ThreadSafeWriter)
	BroadcastEvent(event *types.Event)
	AddVisitor(visitor *Visitor)
}

type ChatRoom struct {
	ID       uint           `gorm:"primaryKey"`
	SFU      sfu.SFU        `gorm:"-:all"`
	ChatLog  *[]ChatRoomLog `gorm:"-:all"`
	Visitors []*Visitor     `gorm:"-:all"`
}

func (r *ChatRoom) AddPeerConnection(pc *webrtc.PeerConnection, w *websocketClient.ThreadSafeWriter) {
	r.SFU.AddPeerConnection(pc, w)
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

func (r *ChatRoom) BroadcastEvent(event *types.Event) {
	for _, v := range r.Visitors {
		v.Client.Writer.WriteJSON(event)
	}
}

func (r *ChatRoom) AddVisitor(visitor *Visitor) {

	r.Visitors = append(r.Visitors, visitor)
}

type Visitor struct {
	User           *users.User
	Host           bool
	Client         *websocketClient.WebsocketClient
	PeerConnection *webrtc.PeerConnection
	DisplayName    string
}

func (r *ChatRoom) CreateUniqueDisplayName() string {
	return r.untilUnique(utils.RandName())
}

func (r *ChatRoom) untilUnique(name string) string {
	unique := true
	for _, v := range r.Visitors {
		if name == v.DisplayName {
			unique = false
			break
		}
	}

	if name == "" || !unique {
		return r.untilUnique(utils.RandName())
	}
	return name
}

func (v *Visitor) Clarify(ask string) error {
	question := &types.Question{
		Ask: ask,
	}
	return v.Client.Writer.WriteJSON(question)
}
