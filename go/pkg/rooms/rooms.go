package rooms

// Rooms map always exists
// If visited /ws with no room id, create room, redirect to ?room=abc

// Create unique room name
// UUID => check if exists

// Room has sfu service, which contains peers and their websocket conns
// Log of chats

// Host peer

// Sends a message to frontend :
// {"event":"joined room", "data": "abc"}
// Event forces frontend to change url + text

// When host leaves, room is closed and all connections severed, room deleted

import (
	"log"
	"time"

	"github.com/Embiggenerd/spiritio/pkg/sfu"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type Rooms struct {
	RoomsTable
}

func Init() *Rooms {
	roomsTable := make(RoomsTable)
	rooms := &Rooms{
		RoomsTable: roomsTable,
	}
	return rooms
}

type Room struct {
	ID      string
	Host    *webrtc.PeerConnection
	ChatLog []string
	SFU     *sfu.SFUService
}

func (r *Rooms) CreateRoom() *Room {
	id := uuid.New()
	room := &Room{
		ID:  id.String(),
		SFU: sfu.NewSelectiveForwardingUnit(),
	}
	r.RoomsTable[id.String()] = room
	peerConnection, err := room.SFU.CreatePeerConnection()
	if err != nil {
		log.Println(err)
	}
	room.Host = peerConnection
	defer peerConnection.Close() //nolint
	go func() {
		for range time.NewTicker(time.Second * 3).C {
			room.SFU.DispatchKeyFrame()
		}
	}()
	return room
}

type RoomsTable map[string]*Room
