package rooms

import (
	"github.com/Embiggenerd/spiritio/types"
	"gorm.io/gorm"
)

type ChatRoomLog struct {
	gorm.Model
	types.UserMessageData
	RoomID uint
}
