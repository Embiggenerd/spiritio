package rooms

import (
	"context"

	"github.com/Embiggenerd/spiritio/pkg/config"
	"github.com/Embiggenerd/spiritio/pkg/db"
	"github.com/Embiggenerd/spiritio/pkg/logger"
)

func NewRoomsService(ctx context.Context, cfg *config.Config, log logger.Logger, db *db.Database) RoomsService {
	db.DB.AutoMigrate(&ChatRoom{})
	db.DB.AutoMigrate(&ChatRoomLog{})
	roomsTable := make(RoomsTable)
	rooms := &ChatRoomsService{
		cache:       &RoomsCache{table: roomsTable},
		DB:          db,
		RoomStorage: &RoomStorage{db: db},
		ChatStorage: &ChatLogStorage{db: db},
	}
	return rooms
}

type RoomsService interface {
	CreateRoom(ctx context.Context) (*ChatRoom, error)
	GetRoomByID(roomID uint) (*ChatRoom, error)
	SaveChatLog(text string, room *ChatRoom) error
}

type ChatRoomsService struct {
	cache       Cache
	DB          *db.Database
	RoomStorage RoomStore
	ChatStorage ChatLogStore
}

// CreateRoom creates a new room
func (r *ChatRoomsService) CreateRoom(ctx context.Context) (*ChatRoom, error) {
	newRoom, err := r.RoomStorage.CreateRoom(ctx)
	if err != nil {
		return newRoom, err
	}

	newRoom.Build()

	r.cache.AddRoom(newRoom)
	return newRoom, err
}

func (r *ChatRoomsService) GetRoomByID(roomID uint) (*ChatRoom, error) {
	room, err := r.cache.GetRoom(roomID)
	if err != nil {
		room, err = r.RoomStorage.FindRoomByID(roomID)
		if err != nil {
			return room, err
		}
		chatLog, _ := r.ChatStorage.GetChatLogsByRoomID(roomID)
		room.ChatLog = chatLog
		room.Build()
		r.cache.AddRoom(room)
	}
	return room, err
}

func (s *ChatRoomsService) SaveChatLog(text string, room *ChatRoom) error {
	chatRoomLog, err := s.ChatStorage.SaveChatlog(text, room)
	s.cache.UpdateChatLogs(room.ID, chatRoomLog)
	return err
}
