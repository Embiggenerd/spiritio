package rooms

import (
	"github.com/Embiggenerd/spiritio/pkg/db"
	"github.com/Embiggenerd/spiritio/pkg/users"
)

type ChatLogStore interface {
	SaveChatlog(text string, room *ChatRoom, from *users.User) (*ChatRoomLog, error)
	GetChatLogsByRoomID(roomID uint) ([]ChatRoomLog, error)
}

type ChatLogStorage struct {
	db *db.Database
}

func (s *ChatLogStorage) SaveChatlog(text string, room *ChatRoom, user *users.User) (*ChatRoomLog, error) {
	var err error
	newChatLog := &ChatRoomLog{
		Text:         text,
		Room:         room,
		UserName:     user.Name,
		UserID:       user.ID,
		UserVerified: user.Verified,
	}

	result := s.db.DB.Create(newChatLog)
	err = result.Error

	return newChatLog, err
}

func (s *ChatLogStorage) GetChatLogsByRoomID(roomID uint) ([]ChatRoomLog, error) {
	var err error
	chatLogs := []ChatRoomLog{}

	result := s.db.DB.Where(ChatRoomLog{RoomID: roomID}).Find(&chatLogs)
	err = result.Error
	return chatLogs, err
}
