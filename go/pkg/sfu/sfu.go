package sfu

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/Embiggenerd/spiritio/pkg/config"
	"github.com/gorilla/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type SFU struct {
	trackLocals map[string]*webrtc.TrackLocalStaticRTP
	// lock for peerConnections and trackLocals
	listLock        *sync.RWMutex
	peerConnections []peerConnectionState
}

func NewSelectiveForwardingUnit(cfg *config.Config) *SFU {
	s := &SFU{}
	s.trackLocals = map[string]*webrtc.TrackLocalStaticRTP{}
	// do this in the handler
	// go func() {
	// 	for range time.NewTicker(time.Second * 3).C {
	// 		s.dispatchKeyFrame()
	// 	}
	// }()
	return s
}

func (s SFU) dispatchKeyFrame() {
	s.listLock.Lock()
	defer s.listLock.Unlock()

	for i := range s.peerConnections {
		for _, receiver := range s.peerConnections[i].peerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = s.peerConnections[i].peerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}

func (s SFU) signalPeerConnections() {
	s.listLock.Lock()
	defer func() {
		s.listLock.Unlock()
		s.dispatchKeyFrame()
	}()

	attemptSync := func() (tryAgain bool) {
		for i := range s.peerConnections {
			if s.peerConnections[i].peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				s.peerConnections = append(s.peerConnections[:i], s.peerConnections[i+1:]...)
				return true // We modified the slice, start from the beginning
			}

			// map of sender we already are seanding, so we don't double send
			existingSenders := map[string]bool{}

			for _, sender := range s.peerConnections[i].peerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true

				// If we have a RTPSender that doesn't map to a existing track remove and signal
				if _, ok := s.trackLocals[sender.Track().ID()]; !ok {
					if err := s.peerConnections[i].peerConnection.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			// Don't receive videos we are sending, make sure we don't have loopback
			for _, receiver := range s.peerConnections[i].peerConnection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}

			// Add all track we aren't sending yet to the PeerConnection
			for trackID := range s.trackLocals {
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := s.peerConnections[i].peerConnection.AddTrack(s.trackLocals[trackID]); err != nil {
						return true
					}
				}
			}

			offer, err := s.peerConnections[i].peerConnection.CreateOffer(nil)
			if err != nil {
				return true
			}

			if err = s.peerConnections[i].peerConnection.SetLocalDescription(offer); err != nil {
				return true
			}

			offerString, err := json.Marshal(offer)
			if err != nil {
				return true
			}

			if err = s.peerConnections[i].websocket.WriteJSON(&websocketMessage{
				Event: "offer",
				Data:  string(offerString),
			}); err != nil {
				return true
			}
		}

		return
	}

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			// Release the lock and attempt a sync in 3 seconds. We might be blocking a RemoveTrack or AddTrack
			go func() {
				time.Sleep(time.Second * 3)
				s.signalPeerConnections()
			}()
			return
		}

		if !attemptSync() {
			break
		}
	}
}

func (t *threadSafeWriter) WriteJSON(v interface{}) error {
	t.Lock()
	defer t.Unlock()

	return t.Conn.WriteJSON(v)
}

type threadSafeWriter struct {
	*websocket.Conn
	sync.Mutex
}

type websocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type peerConnectionState struct {
	peerConnection *webrtc.PeerConnection
	websocket      *threadSafeWriter
}
