package rooms

import (
	"github.com/Embiggenerd/spiritio/pkg/users"
	"github.com/Embiggenerd/spiritio/pkg/utils"
	"github.com/Embiggenerd/spiritio/pkg/websocketClient"
	"github.com/Embiggenerd/spiritio/types"
	"github.com/pion/webrtc/v3"
	"gorm.io/gorm"
)

func NewVisitor(client *websocketClient.WebsocketClient, user *users.User, room *ChatRoom) *Visitor {
	newVisitor := &Visitor{
		User:   user,
		Client: client,
		Room:   room,
	}
	return newVisitor
}

type Visitor struct {
	gorm.Model
	UserID         uint
	Room           *ChatRoom                        `gorm:"-:all"`
	User           *users.User                      `gorm:"foreignKey:UserID"`
	Host           bool                             `gorm:"-:all"`
	Client         *websocketClient.WebsocketClient `gorm:"-:all"`
	PeerConnection *webrtc.PeerConnection           `gorm:"-:all"`
	StreamID       string                           `gorm:"-:all"`
}

func (v *Visitor) AddUser(user *users.User) {
	v.User = user
}

func (v *Visitor) CreateUniqueDisplayName() {
	v.User.Name = v.Room.untilUnique(utils.RandName())
}

func (v *Visitor) Clarify(ask string) error {
	question := &types.Question{
		Ask: ask,
	}
	return v.Client.Writer.WriteJSON(question)
}

func (v *Visitor) Notify(event *types.Event) error {
	return v.Client.Writer.WriteJSON(event)
}
