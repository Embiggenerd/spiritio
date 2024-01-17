package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/Embiggenerd/spiritio/pkg/chat"
	"github.com/Embiggenerd/spiritio/pkg/sfu"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// type Client struct {
// 	websocketService *chat.WebsocketService
// 	sfu              *sfu.SFU
// 	conn             *websocket.Conn
// 	send             chan []byte
// }

// type threadSafeWriter struct {
// 	*websocket.Conn
// 	sync.Mutex
// }

// ServeWs handles websocket requests from the peer.
func ServeWs(websocketService *chat.WebsocketService, sfuService *sfu.SFU, w http.ResponseWriter, r *http.Request) {
	unsafeConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	conn := &sfu.ThreadSafeWriter{unsafeConn, sync.Mutex{}}
	defer conn.Close()

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	log.Println("peerConnection.ConnectionState().String()", peerConnection.ConnectionState().String())
	if err != nil {
		log.Print(err)
		return
	}
	defer peerConnection.Close() //nolint

	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := peerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			log.Print(err)
			return
		}
	}
	// client := &chat.Client{WebsocketService: websocketService, SFU: sfuService, Conn: conn, Send: make(chan []byte, 256)}
	// client.WebsocketService.Register <- client
	// log.Println("we got past Rigester <- client ^^^^^****")
	// go client.WritePump()
	// go client.ReadPump()

	// Add our new PeerConnection to global list
	sfuService.ListLock.Lock()
	sfuService.PeerConnections = append(sfuService.PeerConnections, sfu.PeerConnectionState{PeerConnection: peerConnection, Websocket: conn})
	sfuService.ListLock.Unlock()
	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}

		candidateString, err := json.Marshal(i.ToJSON())
		if err != nil {
			log.Println(err)
			return
		}

		if writeErr := conn.WriteJSON(&chat.WebsocketMessage{
			Event: "candidate",
			Data:  string(candidateString),
		}); writeErr != nil {
			log.Println(writeErr)
		}
	})

	// If PeerConnection is closed remove it from global list
	peerConnection.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		switch p {
		case webrtc.PeerConnectionStateFailed:
			if err := peerConnection.Close(); err != nil {
				log.Print(err)
			}
		case webrtc.PeerConnectionStateClosed:
			sfuService.SignalPeerConnections()
		default:
		}
	})

	peerConnection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		// Create a track to fan out our incoming video to all peers
		trackLocal := sfuService.AddTrack(t)
		defer sfuService.RemoveTrack(trackLocal)

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

	sfuService.SignalPeerConnections()
	log.Println("^^sfuService.ListLock.Unlock()")

	message := &chat.WebsocketMessage{}
	for {
		_, raw, err := conn.ReadMessage()
		log.Println("new message, * ", string(raw[:]), err)
		if err != nil {
			log.Println(err)
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			log.Println(err)
			return
		}
		log.Println("msg", message)

		switch message.Event {
		case "candidate":
			candidate := webrtc.ICECandidateInit{}
			log.Printf(`Sending cadidate %v`, candidate)
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
			log.Printf(`Sending answer %v`, answer)

			if err := json.Unmarshal([]byte(message.Data), &answer); err != nil {
				log.Println(err)
				return
			}

			if err := peerConnection.SetRemoteDescription(answer); err != nil {
				log.Println(err)
				return
			}

		case "user-message":
			log.Printf("writing message: %s", message.Data)
			// if err := c.WriteJSON(message); err != nil {
			// 	log.Println(err)
			// 	return
			// }
			for i := range sfuService.PeerConnections {
				if err = sfuService.PeerConnections[i].Websocket.WriteJSON(message); err != nil {
					log.Println(err)
				}
			}
		}
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.

}
