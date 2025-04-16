package core

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"server/db"
	"server/db/utils"
	"server/models"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

/*
Connection has own web socket connection, database client. Connection needs a
socket manager to send and receive message from other connections(users).
*/
type Connection struct {
	User          *utils.User
	Conn          *websocket.Conn
	SocketManager *SocketManager
	Pool          *db.PostgresPool
}

/*
Listen() is a for-loop where connections get incoming message by websocket.
The message is serialized, stored in DB and then sent to websocket manager.
*/
func (u *Connection) Listen() error {
	defer func() {
		//notify connection dropped to socket manager
		u.SocketManager.Leave <- u
	}()
	for {
		if _, message, err := u.Conn.ReadMessage(); err != nil {
			log.Println("Error on read message =>\n", err.Error())
			return err
		} else {
			msgSerialized := &models.Message{}
			err := json.Unmarshal(message, msgSerialized)
			if err != nil {
				log.Print(err)
			}

			// if message creation fails then update chat throws panic
			err = u.Pool.Transaction(context.Background(), func(tx pgx.Tx) error {
				err := tx.QueryRow(context.Background(), `insert into messages(body, created_at, user_messages) values($1, $2, $3) returning id`,
					msgSerialized.Body, time.Now(), u.User.ID).Scan(&msgSerialized.ID)
				if err != nil {
					return err
				}

				if msgSerialized.ChatID == nil {
					// create chat
					err := tx.QueryRow(context.Background(), `insert into chats(type) values($1) returning id`, "public").Scan(&msgSerialized.ChatID)
					if err != nil {
						return err
					}

					_, err = tx.Exec(context.Background(), `insert into chat_members(chat_id, user_id) values($1, $2)`, *msgSerialized.ChatID, *u.User.ID)
					if err != nil {
						return err
					}

					// check if recipient exists
					var recipientID *uint
					err = tx.QueryRow(context.Background(), `select id from users where username = $1`, msgSerialized.Receiver).Scan(&recipientID)
					if err != nil {
						return err
					}

					_, err = tx.Exec(context.Background(), `insert into chat_members(chat_id, user_id) values($1, $2)`, *msgSerialized.ChatID, *recipientID)
					if err != nil {
						return err
					}
				}

				_, err = tx.Exec(context.Background(), `insert into chat_messages(chat_id, message_id) values($1, $2)`,
					*msgSerialized.ChatID, *msgSerialized.ID)
				if err != nil {
					return err
				}

				return nil

			})
			if err != nil {
				return err
			}

			u.SocketManager.Messages <- &models.Message{
				Body:     msgSerialized.Body,
				Sender:   msgSerialized.Sender,
				Receiver: msgSerialized.Receiver,
				ChatID:   msgSerialized.ChatID,
			}
		}
	}
}

func (u *Connection) Send(message *models.Message) {
	b, _ := json.Marshal(message)

	if err := u.Conn.WriteMessage(websocket.TextMessage, b); err != nil {
		log.Println("Error on write message:", err.Error())
	}
}
