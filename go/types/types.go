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
