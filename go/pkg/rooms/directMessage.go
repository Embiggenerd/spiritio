package rooms

import "gorm.io/gorm"

type DirectMessage struct {
	gorm.Model
	Text       string
	ToUserID   uint
	FromUserID uint
}
