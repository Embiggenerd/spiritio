package sfu

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Embiggenerd/spiritio/pkg/websocketClient"
	"github.com/Embiggenerd/spiritio/types"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type SFU interface {
	AddPeerConnection(pc *webrtc.PeerConnection, w *websocketClient.ThreadSafeWriter)
	DispatchKeyFrame()
	SignalPeerConnections()
	AddTrack(t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP
	RemoveTrack(t *webrtc.TrackLocalStaticRTP)
	CreatePeerConnection() (*webrtc.PeerConnection, error)
	BroadcastMessage(message *types.WebsocketMessage)
}

type SFUService struct {
	trackLocals     map[string]*webrtc.TrackLocalStaticRTP
	ListLock        sync.RWMutex
	PeerConnections []PeerConnectionState
}

func NewSelectiveForwardingUnit() SFU {
	s := &SFUService{}
	s.trackLocals = map[string]*webrtc.TrackLocalStaticRTP{}
	return s
}

func (s *SFUService) AddPeerConnection(pc *webrtc.PeerConnection, w *websocketClient.ThreadSafeWriter) {
	s.ListLock.Lock()
	s.PeerConnections = append(s.PeerConnections, PeerConnectionState{PeerConnection: pc, Websocket: w})
	s.ListLock.Unlock()
}

func (s *SFUService) BroadcastMessage(message *types.WebsocketMessage) {
	// Send message to each client subbed to this peer's peerConnections
	for i := range s.PeerConnections {
		if err := s.PeerConnections[i].Websocket.WriteJSON(message); err != nil {
			// log.Error(err.Error())
		}
	}
}

func (s *SFUService) DispatchKeyFrame() {
	s.ListLock.Lock()
	defer s.ListLock.Unlock()
	for i := range s.PeerConnections {

		for _, receiver := range s.PeerConnections[i].PeerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}
			err := s.PeerConnections[i].PeerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (s *SFUService) SignalPeerConnections() {
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

			// map of sender we already are sending, so we don't double send
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

			if err = s.PeerConnections[i].Websocket.WriteJSON(&types.Event{
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
func (s *SFUService) AddTrack(t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
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

func (s *SFUService) RemoveTrack(t *webrtc.TrackLocalStaticRTP) {
	s.ListLock.Lock()
	defer func() {
		s.ListLock.Unlock()
		s.SignalPeerConnections()
	}()

	delete(s.trackLocals, t.ID())
}

type PeerConnectionState struct {
	PeerConnection *webrtc.PeerConnection
	Websocket      *websocketClient.ThreadSafeWriter
}

func (s *SFUService) CreatePeerConnection() (*webrtc.PeerConnection, error) {
	fmt.Print("CreatePeerConnection()")
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Print(err)
		return peerConnection, err
	}

	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := peerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			return peerConnection, err
		}
	}
	return peerConnection, err
}
