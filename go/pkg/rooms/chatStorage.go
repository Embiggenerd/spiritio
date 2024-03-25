package rooms

import (
	"github.com/Embiggenerd/spiritio/pkg/db"
)

type ChatLogStore interface {
	SaveChatlog(chatLog *ChatRoomLog) (*ChatRoomLog, error)
	GetChatLogsByRoomID(roomID uint) ([]ChatRoomLog, error)
}

type ChatLogStorage struct {
	db *db.Database
}

func (s *ChatLogStorage) SaveChatlog(chatLog *ChatRoomLog) (*ChatRoomLog, error) {
	result := s.db.DB.Create(chatLog)
	return chatLog, result.Error
}

func (s *ChatLogStorage) GetChatLogsByRoomID(roomID uint) ([]ChatRoomLog, error) {
	var err error
	chatLogs := []ChatRoomLog{}

	result := s.db.DB.Where(ChatRoomLog{RoomID: roomID}).Find(&chatLogs)
	err = result.Error
	return chatLogs, err
}
