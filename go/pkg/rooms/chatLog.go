package rooms

import (
	"gorm.io/gorm"
)

type ChatRoomLog struct {
	gorm.Model
	Text         string
	RoomID       uint
	Room         *ChatRoom `gorm:"foreignKey:RoomID"`
	UserID       uint
	UserName     string
	UserVerified int `gorm:"default:0"`
}
