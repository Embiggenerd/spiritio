package sfu

import (
	"encoding/json"
	"log"
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
	ListLock        sync.RWMutex
	PeerConnections []PeerConnectionState
}

func NewSelectiveForwardingUnit(cfg *config.Config) *SFU {
	s := &SFU{}
	s.trackLocals = map[string]*webrtc.TrackLocalStaticRTP{}
	// do this in the handler
	// go func() {
	// 	for range time.NewTicker(time.Second * 3).C {
	// 		s.DispatchKeyFrame()
	// 	}
	// }()
	return s
}

func (s *SFU) DispatchKeyFrame() {
	s.ListLock.Lock()
	defer s.ListLock.Unlock()
	log.Println("$$$$$$$ dispatchKeyFrame(), $$$$$$$$", len(s.PeerConnections))
	for i := range s.PeerConnections {
		// log.Println("peerConnection.ConnectionState().String() in dispatch", s.PeerConnections[i].PeerConnection.ConnectionState().String())

		for _, receiver := range s.PeerConnections[i].PeerConnection.GetReceivers() {
			if receiver.Track() == nil {
				log.Println("receiver.Tracl == nil")
				continue
			}
			log.Println("we never get past this in dispatch")
			err := s.PeerConnections[i].PeerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
			log.Println(err)
		}
	}
}

func (s *SFU) SignalPeerConnections() {
	s.ListLock.Lock()
	defer func() {
		s.ListLock.Unlock()
		s.DispatchKeyFrame()
	}()

	attemptSync := func() (tryAgain bool) {
		for i := range s.PeerConnections {
			if s.PeerConnections[i].PeerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				s.PeerConnections = append(s.PeerConnections[:i], s.PeerConnections[i+1:]...)
				return true // We modified the slice, start from the beginning
			}

			// map of sender we already are seanding, so we don't double send
			existingSenders := map[string]bool{}

			for _, sender := range s.PeerConnections[i].PeerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true

				// If we have a RTPSender that doesn't map to a existing track remove and signal
				if _, ok := s.trackLocals[sender.Track().ID()]; !ok {
					if err := s.PeerConnections[i].PeerConnection.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			// Don't receive videos we are sending, make sure we don't have loopback
			for _, receiver := range s.PeerConnections[i].PeerConnection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}

			// Add all track we aren't sending yet to the PeerConnection
			for trackID := range s.trackLocals {
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := s.PeerConnections[i].PeerConnection.AddTrack(s.trackLocals[trackID]); err != nil {
						return true
					}
				}
			}

			offer, err := s.PeerConnections[i].PeerConnection.CreateOffer(nil)
			// log.Println("### offer in signal ###", offer, err)
			if err != nil {
				return true
			}

			if err = s.PeerConnections[i].PeerConnection.SetLocalDescription(offer); err != nil {
				return true
			}

			offerString, err := json.Marshal(offer)
			if err != nil {
				return true
			}

			if err = s.PeerConnections[i].Websocket.WriteJSON(&websocketMessage{
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
				s.SignalPeerConnections()
			}()
			return
		}

		if !attemptSync() {
			break
		}
	}
}

// Add to list of tracks and fire renegotation for all PeerConnections
func (s *SFU) AddTrack(t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	s.ListLock.Lock()
	defer func() {
		s.ListLock.Unlock()
		s.SignalPeerConnections()
	}()

	// Create a new TrackLocal with the same codec as our incoming
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		panic(err)
	}

	s.trackLocals[t.ID()] = trackLocal
	return trackLocal
}

func (s *SFU) RemoveTrack(t *webrtc.TrackLocalStaticRTP) {
	s.ListLock.Lock()
	defer func() {
		s.ListLock.Unlock()
		s.SignalPeerConnections()
	}()

	delete(s.trackLocals, t.ID())
}

func (t *ThreadSafeWriter) WriteJSON(v interface{}) error {
	t.Lock()
	defer t.Unlock()

	return t.Conn.WriteJSON(v)
}

type ThreadSafeWriter struct {
	*websocket.Conn
	sync.Mutex
}

type websocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type PeerConnectionState struct {
	PeerConnection *webrtc.PeerConnection
	Websocket      *ThreadSafeWriter
}
