package types

type WebsocketMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}
type Event struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type Question struct {
	Type string `json:"type,omitempty"`
	Ask  string `json:"ask,omitempty"`
}

type WorkOrder struct {
	Order   string
	Details interface{}
}
