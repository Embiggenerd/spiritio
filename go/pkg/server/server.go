package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
	"unicode"

	"github.com/Embiggenerd/spiritio/pkg/config"
	"github.com/Embiggenerd/spiritio/pkg/logger"
	"github.com/Embiggenerd/spiritio/pkg/rooms"
	"github.com/Embiggenerd/spiritio/pkg/users"
	"github.com/Embiggenerd/spiritio/pkg/utils"
	"github.com/Embiggenerd/spiritio/pkg/websocketClient"
	"github.com/Embiggenerd/spiritio/types"

	"github.com/pion/webrtc/v4"
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
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

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

func (s *APIServer) serveWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	wsClient, err := websocketClient.New(ctx, s.log, w, r, nil)
	if err != nil {
		s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, nil)
		return
	}
	defer wsClient.Conn.Close()

	var room *rooms.ChatRoom
	roomIDStr := r.URL.Query().Get("room")

	// visitor will be used throughout to gain access to user info and write to connection
	visitor := rooms.NewVisitor(wsClient, nil, room)

	closeHandler := wsClient.Conn.CloseHandler()
	wsClient.Conn.SetCloseHandler(func(code int, text string) error {

		// Remove visitor from room on websocket close message
		for i, v := range room.Visitors {
			if visitor.SocketID == v.SocketID {
				room.Visitors = append(room.Visitors[:i], room.Visitors[i+1:]...)
				break
			}
		}

		if visitor.User != nil {
			event := &types.Event{Event: "user_exited_chat", Data: types.UserExitedChatData{
				Name: visitor.User.Name, ID: visitor.User.ID,
			}}
			visitor.Room.BroadcastEvent(event)
		}
		return closeHandler(code, text)
	})

	if roomIDStr == "" {
		// If the user does not have roomID in search params, create new room
		room, err = s.roomsService.CreateRoom(ctx)
		if err != nil {
			s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, nil)
			return
		}
		// Create and send message with appropriate event
		event := &types.Event{}
		event.Event = "created_room"
		event.Data = strconv.FormatUint(uint64(room.ID), 10)
		visitor.Notify(event)
		return
	}
	// If there is a room specified, they will be redirected here
	roomID, err := utils.StringToUint(roomIDStr)
	if err != nil {
		// If the room ID can't be parsed, create a new room
		room, err = s.roomsService.CreateRoom(ctx)
		if err != nil {
			s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, nil)
			return
		}
		// Create and send message with appropriate event
		event := &types.Event{}
		event.Event = "created_room"
		event.Data = strconv.FormatUint(uint64(room.ID), 10)
		visitor.Notify(event)
		return

	}
	room, err = s.roomsService.GetRoomByID(roomID)
	if err != nil {
		// If room ID can't be found, create a new room
		room, err = s.roomsService.CreateRoom(ctx)
		if err != nil {
			s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, nil)
			return
		}
		// Create and send message with appropriate event
		event := &types.Event{}
		event.Event = "created_room"
		event.Data = strconv.FormatUint(uint64(room.ID), 10)
		wsClient.Writer.WriteJSON(event)
		return
	}

	visitor.Room = room
	room.AddVisitor(visitor)

	event := &types.Event{}
	event.Event = "joined_room"
	// create chat events from room's chatlog cache
	var chats []types.UserMessageData
	for i := 0; i < len(room.ChatLog); i++ {
		chat := room.ChatLog[i].UserMessageData
		chats = append(chats, chat)
	}

	visitors := []types.Visitor{}
	for _, v := range room.Visitors {
		utils.PrintStruct(v)
		if v.User != nil {
			visitors = append(visitors, types.Visitor{ID: v.User.ID, Name: v.User.Name})
		}
	}

	event.Data = types.JoinedRoomData{
		ChatLog:  chats,
		RoomID:   room.ID,
		Visitors: visitors,
	}
	visitor.Notify(event)

	// ask for authentication
	visitor.Clarify("access_token")

	var peerConnection *webrtc.PeerConnection
	workOrder := &types.WorkOrder{}
	for {
		_, raw, err := wsClient.Conn.ReadMessage()
		if err != nil {
			s.log.Error(err.Error())
			return
		} else if err := json.Unmarshal(raw, &workOrder); err != nil {
			s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, nil)
			return
		}
		s.log.LogWorkOrderReceived(ctx, workOrder)

		switch workOrder.Order {
		case "media_request":
			peerConnection, err = room.SFU.CreatePeerConnection()
			if err != nil {
				s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, visitor)
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

				if writeErr := wsClient.Writer.WriteJSON(&types.Event{
					Event: "candidate",
					Data:  string(candidateString),
				}); writeErr != nil {
					s.log.Error(writeErr.Error())
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
				}
			})

			peerConnection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
				// Create a track to fan out our incoming video to all peers
				trackLocal := room.SFU.AddTrack(t)
				visitor.StreamID = trackLocal.StreamID()
				defer room.SFU.RemoveTrack(trackLocal)

				buf := make([]byte, 1500)
				for {
					i, _, err := t.Read(buf)
					if err != nil {
						s.log.Error(err.Error())
						return
					}

					if _, err = trackLocal.Write(buf[:i]); err != nil {
						s.log.Error(err.Error())
						return
					}
				}
			})

			room.SFU.SignalPeerConnections()

		case "validate_access_token":
			token, err := s.userService.ValidateAccessToken(workOrder.Details.(string))
			if err != nil {
				user, accessToken, err := s.userService.CreateUser(false)
				if err != nil {
					s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, visitor)
					break
				}
				visitor.AddUser(user)

				data := types.UserLoggedInData{
					Name:        user.Name,
					ID:          user.ID,
					AccessToken: accessToken,
				}

				event := &types.Event{
					Event: "user_logged_in",
					Data:  data,
				}

				visitor.Notify(event)
				visitor.Clarify("credentials")

				event = &types.Event{
					Event: "user_entered_chat",
					Data:  user.Name,
				}
				visitor.Room.BroadcastEvent(event)

			} else {
				user, err := s.userService.GetUserFromAccessToken(token)
				if err != nil {
					s.handleError(ctx, "failed parsing user token", http.StatusInternalServerError, err, visitor)
					// below code is untested and seems to be broken
					user, accessToken, err := s.userService.CreateUser(false)

					event := &types.Event{
						Event: "user_entered_chat",
						Data:  user.Name,
					}
					visitor.Room.BroadcastEvent(event)

					visitor.AddUser(user)
					if err != nil {
						s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, visitor)
						return
					}

					data := types.UserLoggedInData{
						Name:        user.Name,
						ID:          user.ID,
						AccessToken: accessToken,
					}

					event = &types.Event{
						Event: "user_logged_in",
						Data:  data,
					}

					visitor.Notify(event)
					visitor.Clarify("credentials")
				}

				event := &types.Event{
					Event: "user_entered_chat",
					Data:  user.Name,
				}
				visitor.Room.BroadcastEvent(event)

				visitor.User = user

				data := types.UserLoggedInData{
					Name:        user.Name,
					ID:          user.ID,
					AccessToken: workOrder.Details.(string),
				}
				event = &types.Event{
					Event: "user_logged_in",
					Data:  data,
				}
				visitor.Notify(event)

			}

		case "candidate":
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(workOrder.Details.(string)), &candidate); err != nil {
				s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, visitor)
				break
			}

			if err := peerConnection.AddICECandidate(candidate); err != nil {
				s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, visitor)
			}

		case "answer":
			answer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(workOrder.Details.(string)), &answer); err != nil {
				s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, visitor)
				break
			}

			if err := peerConnection.SetRemoteDescription(answer); err != nil {
				s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, visitor)
			}

		case "user_message":
			wo := &types.UserMessageWorkOrder{}
			err = json.Unmarshal(raw, wo)
			if err != nil {
				s.handleError(ctx, "user message did not go through", 400, err, visitor)
				break
			}

			data := types.UserMessageData{
				Text:         wo.Details.Text,
				ToUserID:     wo.Details.ToUserID,
				FromUserName: visitor.User.Name,
				UserVerified: visitor.User.Verified != 0,
				FromUserID:   visitor.User.ID,
			}

			event := &types.Event{
				Event: "user_message",
				Data:  data,
			}

			isDirectMessage := data.ToUserID != 0
			if isDirectMessage {
				userPresent := false
				for _, v := range visitor.Room.Visitors {
					if v.User.ID == data.ToUserID {
						userPresent = true
						v.Notify(event)
					}
				}
				if !userPresent {
					s.handleError(ctx, "user is not present", 400, nil, visitor)
				}
			} else {
				room.BroadcastEvent(event)
			}

			// Write new chatlog to DB with this room's ID as foreign key
			err := s.roomsService.SaveChatLog(data, visitor)
			if err != nil {
				s.log.Error(err.Error())
			}

		case "set_user_password":
			password := workOrder.Details.(map[string]interface{})["password"]
			valid := validateUserPassword(password.(string))
			if !valid {
				s.handleError(ctx, "password must be at least 8 characters long, and contain a number, letter, and special character", http.StatusBadRequest, err, visitor)
				break
			}

			err = s.userService.UpdateUserPassword(visitor.User.ID, password.(string))
			if err != nil {
				s.handleError(ctx, "", http.StatusInternalServerError, err, visitor)
			}

			data := types.UserMessageData{
				Text:         "Password changed",
				FromUserName: "ADMIN (to you)",
				UserVerified: false,
				FromUserID:   0,
			}

			event := &types.Event{
				Event: "user_message",
				Data:  data,
			}
			visitor.Notify(event)

		case "set_user_name":
			// Check if user has set a password
			userID := visitor.User.ID
			user, err := s.userService.GetUserByID(userID)
			if err != nil {
				s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, visitor)
			}
			if user.Password == "" {
				s.handleError(ctx, "please set a password to create a permanent user name", http.StatusForbidden, err, visitor)
				break
			}

			name := workOrder.Details.(map[string]interface{})["name"].(string)
			err = s.userService.UpdateUserName(userID, name)
			if err != nil {
				s.handleError(ctx, "", http.StatusBadRequest, err, visitor)
				break
			}

			visitor.User.Name = name

			event := &types.Event{
				Event: "user_name_change",
				Data:  name,
			}
			visitor.Notify(event)

			data := &types.StreamIDUserNameData{
				StreamID: visitor.StreamID,
				Name:     name,
			}

			event = &types.Event{
				Event: "streamid_user_name",
				Data:  data,
			}
			visitor.Room.BroadcastEvent(event)

		case "validate_user_name_password":
			password := workOrder.Details.(map[string]interface{})["password"]
			name := workOrder.Details.(map[string]interface{})["name"]

			user, err := s.userService.ValidateNamePassword(name.(string), password.(string))
			if err != nil {
				s.handleError(ctx, "failed login", http.StatusBadRequest, err, visitor)
			} else {
				accessToken, err := s.userService.CreateAccessToken(user)
				if err != nil {
					s.handleError(ctx, "internal server error", http.StatusInternalServerError, err, visitor)
					break
				}

				visitor.AddUser(user)
				data := types.UserLoggedInData{
					Name:        user.Name,
					ID:          user.ID,
					AccessToken: accessToken,
				}

				event := &types.Event{
					Event: "user_logged_in",
					Data:  data,
				}

				visitor.Notify(event)

				event = &types.Event{
					Event: "user_entered_chat",
					Data:  user.Name,
				}
				visitor.Room.BroadcastEvent(event)
			}

		case "identify_streamid":
			for _, v := range visitor.Room.Visitors {
				if v.StreamID == workOrder.Details.(string) {
					data := &types.StreamIDUserNameData{
						StreamID: workOrder.Details.(string),
						Name:     v.User.Name,
					}

					event := &types.Event{
						Event: "streamid_user_name",
						Data:  data,
					}
					visitor.Room.BroadcastEvent(event)
					break
				}
			}

		case "get_current_guests":
			guests := types.CurrentGuestsData{}
			// Remove duplicates and own visitor
			for _, v := range visitor.Room.Visitors {
				if v.User != nil && visitor.User != nil {
					guests = append(guests, types.CurrentGuest{
						Name: v.User.Name, ID: v.User.ID,
					})
				}
			}
			deduped := utils.RemoveDuplicate(guests)
			visitor.Notify(&types.Event{Event: "current_guests", Data: deduped})
		}

	}
}

func (s *APIServer) handleError(ctx context.Context, message string, statusCode int, err error, visitor *rooms.Visitor) {
	reqID, _ := utils.ExposeContextMetadata(ctx).Get("requestID")

	if err == nil {
		err = fmt.Errorf(message)
	}

	if message == "" {
		message = err.Error()
	}

	s.log.LogRequestError(reqID.(string), err.Error(), http.StatusInternalServerError)
	data := &types.ErrorData{
		StatusCode: statusCode,
		Message:    message,
		Public:     true,
	}
	event := &types.Event{
		Event: "error",
		Data:  data,
	}
	if visitor != nil {
		visitor.Notify(event)
	}
}

func validateUserPassword(password string) bool {
	var (
		hasCorrectLen = false
		hasLetter     = false
		hasNumber     = false
		hasSpecial    = false
	)
	if len(password) >= 8 {
		hasCorrectLen = true
	}
	for _, char := range password {
		switch {
		case unicode.IsLetter(char):
			hasLetter = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasCorrectLen && hasLetter && hasNumber && hasSpecial

}
