package rooms

import (
	"errors"
	"sync"
)

type Cache interface {
	AddRoom(room *ChatRoom)
	GetRoom(roomID uint) (*ChatRoom, error)
	UpdateChatLogs(roomID uint, chatRoomLog *ChatRoomLog)
}

type RoomsCache struct {
	table RoomsTable
	mu    sync.Mutex
}

func (c *RoomsCache) AddRoom(room *ChatRoom) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.table[room.ID] = room
}

func (c *RoomsCache) GetRoom(roomID uint) (*ChatRoom, error) {
	var err error
	val, ok := c.table[roomID]
	if !ok {
		err = errors.New("room could not be found")
	}
	return val, err
}

func (c *RoomsCache) UpdateChatLogs(roomID uint, chatRoomLog *ChatRoomLog) {
	c.mu.Lock()
	defer c.mu.Unlock()
	currentChatLogs := *c.table[roomID].ChatLog
	newChatLogs := append(currentChatLogs, *chatRoomLog)
	c.table[roomID].ChatLog = &newChatLogs
}

type RoomsTable map[uint]*ChatRoom
