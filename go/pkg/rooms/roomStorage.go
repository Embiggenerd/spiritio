package rooms

import (
	"context"

	"github.com/Embiggenerd/spiritio/pkg/db"
)

type RoomStore interface {
	CreateRoom(ctx context.Context) (*ChatRoom, error)
	// SaveChatlog(text string, room *Room) *ChatLog
	FindRoomByID(roomID uint) (*ChatRoom, error)
}

type RoomStorage struct {
	db *db.Database
}

// CreateRoom creates a new room
func (r *RoomStorage) CreateRoom(ctx context.Context) (*ChatRoom, error) {
	newRoom := &ChatRoom{}
	roomResult := r.db.DB.Create(newRoom)
	return newRoom, roomResult.Error
}

func (r *RoomStorage) FindRoomByID(roomID uint) (*ChatRoom, error) {
	foundRoom := &ChatRoom{}
	roomResult := r.db.DB.Where(ChatRoom{ID: roomID}).First(foundRoom)
	return foundRoom, roomResult.Error
}
