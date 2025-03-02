package models

const (
	SortChat    = 0
	NewChat     = 1
	FetchChats  = 2
	GetMessages = 3
	LoadChat    = 4
)

type ChatEvent struct {
	Type   int
	ChatID int
}

type Chat struct {
	ChatID   int    `json:"chatID"`
	Username string `json:"username"`
}
