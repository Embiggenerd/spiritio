package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/Embiggenerd/spiritio/pkg/rooms"
	"github.com/Embiggenerd/spiritio/pkg/utils"
	"github.com/Embiggenerd/spiritio/pkg/websocketClient"
	"github.com/pion/webrtc/v3"
	"gorm.io/gorm"
)

// Create single source of truth for who is subbed to a room - right now we depend on peer connections, but should happen on 'join room'
// Peer connections may fail, but websockets should still work

// Each room should have a table with all websocket connections and all peer connections
// associated with it. Send messages down websockets only, peer connections are for rtc only
// Change sfu to rtc once mcu is incorporated

// SFU, MCU services, with RTC client that uses both

// ServeWs handles websocket requests from the peer.
func ServeWs(roomsService rooms.RoomsService, w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = utils.WithMetadata(ctx)
	metaData := utils.ExposeContextMetadata(ctx)

	defer cancel()
	userIP := r.RemoteAddr
	metaData.Set("userIP", userIP)

	wsClient, err := websocketClient.New(ctx, w, r, nil)
	defer wsClient.Conn.Close()
	if err != nil {
		log.Println(err)
		return
	}

	var room *rooms.ChatRoom
	roomIDStr := r.URL.Query().Get("room")
	if roomIDStr == "" {
		// If the user does not have roomID in search params, create new room
		room, err = roomsService.CreateRoom(ctx)
		if err != nil {
			log.Println(err)
		}
		// Create and send message with appropriate event
		message := &websocketClient.WebsocketMessage{}
		message.Event = "created-room"
		message.Data = strconv.FormatUint(uint64(room.ID), 10)
		wsClient.Writer.WriteJSON(message)
	} else {
		// If there is a room specified, or if the creator's URL was modified, they will be redirected here
		roomID, err := utils.StringToUint(roomIDStr)
		if err != nil {
			log.Println(err)
		}
		r, err := roomsService.GetRoomByID(roomID)
		if err == nil {
			// If found, designate the room and send appropriate message
			room = r
			message := &websocketClient.JoinRoomWebsocketMessage{}
			message.Event = "joined-room"
			var chats []string
			for i := 0; i < len(*room.ChatLog); i++ {
				chats = append(chats, (*room.ChatLog)[i].Text)
			}
			message.Data = websocketClient.JoinRoomData{
				ChatLog: chats,
				RoomID:  room.ID,
			}

			wsClient.Writer.WriteJSON(message)
			if err != nil {
				log.Println(err)
			}
		} else {
			room, err = roomsService.GetRoomByID(roomID)
			if errors.Is(err, gorm.ErrRecordNotFound) || room.ID == 0 {
				log.Printf("room with key %s doesn't exist", roomIDStr)
				// Return 400 error
				return
			} else {
				var chats []string
				for i := 0; i < len(*room.ChatLog); i++ {
					chats = append(chats, (*room.ChatLog)[i].Text)
				}
				message := &websocketClient.JoinRoomWebsocketMessage{}
				message.Event = "joined-room"
				message.Data = websocketClient.JoinRoomData{
					ChatLog: chats,
					RoomID:  room.ID,
				}

				wsClient.Writer.WriteJSON(message)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
	// Create peer connection
	peerConnection, err := room.SFU.CreatePeerConnection()
	defer peerConnection.Close()
	if err != nil {
		log.Println(err)
	}

	room.AddPeerConnection(peerConnection, wsClient.Writer)

	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}

		candidateString, err := json.Marshal(i.ToJSON())
		if err != nil {
			log.Println(err)
			return
		}

		if writeErr := wsClient.Writer.WriteJSON(&websocketClient.WebsocketMessage{
			Event: "candidate",
			Data:  string(candidateString),
		}); writeErr != nil {
			log.Println(writeErr)
		}
	})

	peerConnection.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		switch p {
		case webrtc.PeerConnectionStateFailed:
			if err := peerConnection.Close(); err != nil {
				log.Print(err)
			}
		case webrtc.PeerConnectionStateClosed:
			room.SFU.SignalPeerConnections()
		default:
		}
	})

	peerConnection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		// Create a track to fan out our incoming video to all peers
		trackLocal := room.SFU.AddTrack(t)
		defer room.SFU.RemoveTrack(trackLocal)

		buf := make([]byte, 1500)
		for {
			i, _, err := t.Read(buf)
			if err != nil {
				return
			}

			if _, err = trackLocal.Write(buf[:i]); err != nil {
				return
			}
		}
	})

	room.SFU.SignalPeerConnections()

	message := &websocketClient.WebsocketMessage{}
	for {
		_, raw, err := wsClient.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			log.Println(err)
			return
		}
		switch message.Event {
		case "candidate":
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(message.Data), &candidate); err != nil {
				log.Println(err)
				return
			}

			if err := peerConnection.AddICECandidate(candidate); err != nil {
				log.Println(err)
				return
			}
		case "answer":
			answer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(message.Data), &answer); err != nil {
				log.Println(err)
				return
			}

			if err := peerConnection.SetRemoteDescription(answer); err != nil {
				log.Println(err)
				return
			}
		case "user-message":
			// Send message to each client subbed to this peer's peerConnections
			for i := range room.SFU.PeerConnections {
				if err = room.SFU.PeerConnections[i].Websocket.WriteJSON(message); err != nil {
					log.Println(err)
				}
			}
			// Write new chatlog to DB with this room's ID as foreign key
			err := roomsService.SaveChatLog(message.Data, room)
			if err != nil {
				log.Println(err.Error())
			}
		}
	}
}
