package chat

import (
	"context"
	"encoding/json"
	"log"

	"loro-chat/server/models"

	"github.com/gorilla/websocket"
)

/*
User has own web socket connection, database client and needs to socket manager
in order to send and receive message from other users.
*/
type User struct {
	Username      string
	Conn          *websocket.Conn
	SocketManager *SocketManager
	Database      *ent.Client
	EntUser       *ent.User
}

/*
Listen() is a for-loop where the user gets incoming message by websocket.
The message is serialized, stored in DB and then sent to websocket manager.
*/
func (u *User) Listen() error {
	defer func() {
		//necesitamos avisar al Chat que user se fue
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
			msg, err := u.Database.Message.Create().SetBody(*msgSerialized.Body).SetFrom(u.EntUser).Save(context.Background())
			if err != nil {
				log.Print(err)
			}

			if msgSerialized.ChatID == nil {
				ids := make([]int, 0)
				ids = append(ids, u.EntUser.ID)
				receiver, err := u.Database.User.Query().Where(eu.UsernameEQ(*msgSerialized.Receiver)).First(context.Background())
				if err != nil {
					log.Print(err)
				} else {
					ids = append(ids, receiver.ID)
				}

				// if chat creation fails then update chat throws panic
				newChat, err := u.Database.Chat.Create().SetType("public").AddMemberIDs(ids...).Save(context.Background())
				if err != nil {
					log.Print(err)
				}
				if newChat != nil {
					msgSerialized.ChatID = &newChat.ID
				}
			}

			_, err = u.Database.Chat.UpdateOneID(int(*msgSerialized.ChatID)).AddMessageIDs(msg.ID).Save(context.Background())
			if err != nil {
				log.Print(err)
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

func (u *User) Send(message *models.Message) {
	b, _ := json.Marshal(message)

	if err := u.Conn.WriteMessage(websocket.TextMessage, b); err != nil {
		log.Println("Error on write message:", err.Error())
	}
}
