package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Embiggenerd/spiritio/pkg/rooms"
	"github.com/Embiggenerd/spiritio/pkg/sfu"
	"github.com/Embiggenerd/spiritio/pkg/websocketClient"
	"github.com/pion/webrtc/v3"
)

// type Client struct {
// 	websocketService *websocketService.WebsocketService
// 	sfu              *sfu.SFU
// 	conn             *websocket.Conn
// 	send             chan []byte
// }

// type threadSafeWriter struct {
// 	*websocket.Conn
// 	sync.Mutex
// }

// ServeWs handles websocket requests from the peer.
func ServeWs(roomsService *rooms.Rooms, w http.ResponseWriter, r *http.Request) {
	// If user got here without roomID in param, we create a new room
	// and attach references to that room

	// Else, we use the roomID to get the room from RoomsTable, and use
	// its reference to SFU

	// Possible problem: when we broadcast to only peer connections in room,
	// not a problem. However, if we send to the connection from browser,
	// does everyone get it?

	wsClient, err := websocketClient.New(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer wsClient.Conn.Close()

	var room *rooms.Room
	var peerConnection *webrtc.PeerConnection
	roomID := r.URL.Query().Get("room")
	log.Println("******* ", roomID, " ********")
	if roomID == "" {
		// send create room message
		room = roomsService.CreateRoom()
		log.Print("created room", room.ID)
		peerConnection = room.Host
		message := &websocketClient.WebsocketMessage{}
		message.Event = "created-room"
		message.Data = room.ID
		wsClient.Writer.WriteJSON(message)
	} else {
		val, ok := roomsService.RoomsTable[roomID]
		// log.Print("joining room", room.ID)

		if ok {
			room = val
			peerConnection, err = room.SFU.CreatePeerConnection()
			message := &websocketClient.JoinRoomWebsocketMessage{}
			message.Event = "joined-room"
			message.Data = websocketClient.JoinRoomData{
				ChatLog: room.ChatLog,
				RoomID:  room.ID,
			}

			wsClient.Writer.WriteJSON(message)
			if err != nil {
				log.Println(err)
				// Handle error
			}
		} else {
			log.Printf("room with key %s doesn't exist", roomID)
			// Handle error
		}
	}

	room.SFU.ListLock.Lock()
	room.SFU.PeerConnections = append(room.SFU.PeerConnections, sfu.PeerConnectionState{PeerConnection: peerConnection, Websocket: wsClient.Writer})
	room.SFU.ListLock.Unlock()

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
	// End

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
			for i := range room.SFU.PeerConnections {
				if err = room.SFU.PeerConnections[i].Websocket.WriteJSON(message); err != nil {
					log.Println(err)
				}
			}
			room.ChatLog = append(room.ChatLog, message.Data)
		}
	}
}
