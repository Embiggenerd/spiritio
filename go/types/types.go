package types

type WebsocketMessage struct {
	Type string      `json:"type,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

type Event struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type Question struct {
	Ask string `json:"ask,omitempty"`
}

type WorkOrder struct {
	Order   string
	Details interface{}
}

type JoinedRoomData struct {
	RoomID   uint              `json:"room_id,omitempty"`
	ChatLog  []UserMessageData `json:"chat_log"`
	Name     string            `json:"name,omitempty"`
	Visitors []Visitor         `json:"visitors"`
}

type Visitor struct {
	ID   uint   `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type ErrorData struct {
	StatusCode int    `json:"status_code,omitempty"`
	Message    string `json:"message,omitempty"`
	Public     bool   `json:"public,omitempty"`
}

type UserLoggedInData struct {
	Name        string `json:"name,omitempty"`
	ID          uint   `json:"id,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

type StreamIDUserNameData struct {
	StreamID string `json:"stream_id,omitempty"`
	Name     string `json:"name,omitempty"`
}

type DirectMessageData struct {
	Text       string
	ToUserID   uint
	FromUserID uint
}

type UserMessageData struct {
	Text         string `json:"text,omitempty"`
	FromUserName string `json:"from_user_name,omitempty"`
	FromUserID   uint   `json:"from_user_id,omitempty"`
	UserVerified bool   `json:"user_verified"`
	ToUserID     uint   `json:"to_user_id,omitempty"`
}

type UserMessageWorkOrderDetail struct {
	Text     string
	ToUserID uint
}

type UserMessageWorkOrder struct {
	Order   string
	Details UserMessageWorkOrderDetail
}

type UserExitedChatData struct {
	Name string `json:"name,omitempty"`
	ID   uint   `json:"id,omitempty"`
}

type CurrentGuest struct {
	Name string `json:"name,omitempty"`
	ID   uint   `json:"id,omitempty"`
}

type CurrentGuestsData []CurrentGuest
