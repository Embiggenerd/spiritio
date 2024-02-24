package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/Embiggenerd/spiritio/pkg/config"
	"github.com/Embiggenerd/spiritio/pkg/logger"
	"github.com/Embiggenerd/spiritio/pkg/rooms"
	"github.com/Embiggenerd/spiritio/pkg/users"
	"github.com/Embiggenerd/spiritio/pkg/utils"
	"github.com/Embiggenerd/spiritio/pkg/websocketClient"
	"github.com/Embiggenerd/spiritio/types"

	"github.com/pion/webrtc/v3"
	"gorm.io/gorm"
)

type APIServer struct {
	server       *http.Server
	roomsService rooms.RoomsService
	userService  users.Users
	log          logger.Logger
}

func NewServer(ctx context.Context, cfg *config.Config, log logger.Logger, roomsService rooms.RoomsService, usersService users.Users) *APIServer {
	server := &http.Server{
		Addr:              cfg.Addr,
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Info("api server up")
	return &APIServer{
		server:       server,
		roomsService: roomsService,
		userService:  usersService,
		log:          log,
	}
}

func (s *APIServer) Run() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.serveHome)
	mux.HandleFunc("/ws", s.serveWS)

	withMW := s.log.LoggingMW(mux)

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		s.log.Fatal(err.Error())
	}

	s.log.Info("server listening on port " + s.server.Addr)

	if err := http.Serve(l, withMW); err != nil {
		s.log.Fatal(err.Error())
	}
}

func (s *APIServer) serveHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID, _ := utils.ExposeContextMetadata(ctx).Get("requestID")
	if r.URL.Path != "/" {
		s.log.LogRequestError(reqID.(string), "Not found", http.StatusNotFound)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		s.log.LogRequestError(reqID.(string), "Method not allowed", http.StatusMethodNotAllowed)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func (s *APIServer) serveWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metadata := utils.ExposeContextMetadata(ctx)
	reqID, _ := metadata.Get("requestID")

	wsClient, err := websocketClient.New(ctx, s.log, w, r, nil)

	if err != nil {
		s.log.LogRequestError(reqID.(string), err.Error(), http.StatusInternalServerError)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer wsClient.Conn.Close()
	// Block until user is either created or authenticated
	userAuth := make(chan *users.User, 1)
	user := <-userAuth
	s.log.Info("user name: " + user.Name)
	var room *rooms.ChatRoom
	roomIDStr := r.URL.Query().Get("room")
	if roomIDStr == "" {
		// If the user does not have roomID in search params, create new room
		room, err = s.roomsService.CreateRoom(ctx)
		if err != nil {
			s.log.Error(err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		metadata.Set("roomID", room.ID)
		// Create and send message with appropriate event
		message := &types.WebsocketMessage{}
		message.Event = "created-room"
		message.Data = strconv.FormatUint(uint64(room.ID), 10)
		wsClient.Writer.WriteJSON(message)
	} else {
		// If there is a room specified, or if the creator's URL was modified, they will be redirected here
		roomID, err := utils.StringToUint(roomIDStr)
		if err != nil {
			s.log.Error(err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		r, err := s.roomsService.GetRoomByID(roomID)
		if err == nil {
			// If found, send appropriate message
			room = r
			message := &types.WebsocketMessage{}
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
		} else {
			room, err = s.roomsService.GetRoomByID(roomID)
			if errors.Is(err, gorm.ErrRecordNotFound) || room.ID == 0 {
				s.log.Error("room with key %s doesn't exist", roomIDStr)
				http.Error(w, "Room with key %s doesn't exist", http.StatusBadRequest)
				return
			} else {
				var chats []string
				for i := 0; i < len(*room.ChatLog); i++ {
					chats = append(chats, (*room.ChatLog)[i].Text)
				}
				message := &types.WebsocketMessage{}
				message.Event = "joined-room"
				message.Data = websocketClient.JoinRoomData{
					ChatLog: chats,
					RoomID:  room.ID,
				}

				wsClient.Writer.WriteJSON(message)
			}
		}
	}
	// Create peer connection
	peerConnection, err := room.SFU.CreatePeerConnection()
	if err != nil {
		s.log.Error(err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer peerConnection.Close()

	room.AddPeerConnection(peerConnection, wsClient.Writer)

	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}

		candidateString, err := json.Marshal(i.ToJSON())
		if err != nil {
			s.log.Error(err.Error())
			return
		}

		if writeErr := wsClient.Writer.WriteJSON(&types.WebsocketMessage{
			Event: "candidate",
			Data:  string(candidateString),
		}); writeErr != nil {
			s.log.Error(writeErr.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	peerConnection.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		switch p {
		case webrtc.PeerConnectionStateFailed:
			if err := peerConnection.Close(); err != nil {
				s.log.Error(err.Error())
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
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if _, err = trackLocal.Write(buf[:i]); err != nil {
				s.log.Error(err.Error())
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}
	})

	room.SFU.SignalPeerConnections()

	message := &types.WebsocketMessage{}
	for {
		_, raw, err := wsClient.Conn.ReadMessage()
		if err != nil {
			s.log.Error(err.Error())
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			s.log.Error(err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		s.log.LogEventReceived(reqID.(string), metadata.ToJSON(), message)
		switch message.Event {
		case "authentication":
			if message.Data == nil || message.Data == "" {
				fmt.Println("*&*&", message.Data)
				message := &types.WebsocketMessage{
					Event: "error",
					Data:  http.StatusUnauthorized,
				}
				if err = wsClient.Writer.WriteJSON(message); err != nil {
					s.log.Error(err.Error())
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				// create new user or ask for userName / password
				user, token, err := s.userService.CreateUser(false)
				if err != nil {
					s.log.Error(err.Error())
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				// Once we have created a user, we can write to the channel blocking the function

				message = &types.WebsocketMessage{
					Event: "authorization",
					Data:  token,
				}

				if err = wsClient.Writer.WriteJSON(message); err != nil {
					s.log.Error(err.Error())
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				message = &types.WebsocketMessage{
					Event: "login",
					Data: map[string]string{
						"userName": user.Name,
						"userID":   utils.UintToString(user.ID),
					},
				}

				if err = wsClient.Writer.WriteJSON(message); err != nil {
					s.log.Error(err.Error())
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				userAuth <- user

			} else {
				token, err := s.userService.ValidateAccessToken(message.Data.(string))
				if err != nil {
					s.log.Info("*&^", err)
					message := &types.WebsocketMessage{
						Event: "error",
						Data:  401,
					}
					if err = wsClient.Writer.WriteJSON(message); err != nil {
						s.log.Error(err.Error())
						http.Error(w, "Internal server error", http.StatusInternalServerError)
						return
					}
				}

				user, err := s.userService.GetUserFromAccessToken(token)
				if err != nil {
					s.log.Error(err.Error())
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				newToken, err := s.userService.CreateAccessToken(user)
				if err != nil {
					s.log.Error(err.Error())
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				message := &types.WebsocketMessage{
					Event: "authorization",
					Data:  newToken,
				}

				if err = wsClient.Writer.WriteJSON(message); err != nil {
					s.log.Error(err.Error())
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				message = &types.WebsocketMessage{
					Event: "login",
					Data:  user,
				}
				if err = wsClient.Writer.WriteJSON(message); err != nil {
					s.log.Error(err.Error())
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				userAuth <- user
			}
			// data := &struct {
			// 	Bearer *jwt.Token
			// }{}

			// if err := json.Unmarshal(message.Data, &data); err != nil {
			// 	s.log.Error(err.Error())
			// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
			// 	return
			// }
			// s.log.Info(string(data))

		case "candidate":
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(message.Data.(string)), &candidate); err != nil {
				s.log.Error(err.Error())
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if err := peerConnection.AddICECandidate(candidate); err != nil {
				s.log.Error(err.Error())
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		case "answer":
			answer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(message.Data.(string)), &answer); err != nil {
				s.log.Error(err.Error())
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if err := peerConnection.SetRemoteDescription(answer); err != nil {
				s.log.Error(err.Error())
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		case "user-message":
			// Send message to each client subbed to this peer's peerConnections
			room.SFU.BroadcastMessage(message)
			// Write new chatlog to DB with this room's ID as foreign key
			err := s.roomsService.SaveChatLog(message.Data.(string), room)
			if err != nil {
				s.log.Error(err.Error())
			}
		}
	}
}
