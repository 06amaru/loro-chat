package models

const (
	Forward  = 0
	Incoming = 1
)

type Message struct {
	ID       *int    `json:"id,omitempty"`
	Body     *string `json:"body,omitempty"`
	Sender   *string `json:"sender,omitempty"`
	Receiver *string `json:"receiver,omitempty"`
	ChatID   *int    `json:"chatId,omitempty"`
}

type MessageEvent struct {
	Type int
	*Message
}
