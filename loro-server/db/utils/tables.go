package utils

import "time"

type Message struct {
	ID        *uint      `json:"id"`
	Body      *string    `json:"body"`
	CreatedAt *time.Time `json:"created_at"`
}
type User struct {
	ID         *uint      `json:"id"`
	Username   *string    `json:"username"`
	CreatedAt  *time.Time `json:"created_at"`
	Password   *string    `json:"password"`
	PrivateKey []byte     `json:"private_key"`
	PublicKey  []byte     `json:"public_key"`
}

type Chat struct {
	ID              *uint      `json:"id"`
	LastMessage     *string    `json:"last_message"`
	LastMessageTime *time.Time `json:"last_message_time"`
}
