package sfu

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/Embiggenerd/spiritio/pkg/websocketClient"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type SFUService struct {
	trackLocals map[string]*webrtc.TrackLocalStaticRTP
	// lock for peerConnections and trackLocals
	ListLock        sync.RWMutex
	PeerConnections []PeerConnectionState
}

func NewSelectiveForwardingUnit() *SFUService {
	s := &SFUService{}
	s.trackLocals = map[string]*webrtc.TrackLocalStaticRTP{}
	return s
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
			log.Println(err)
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

			if err = s.PeerConnections[i].Websocket.WriteJSON(&websocketClient.WebsocketMessage{
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
