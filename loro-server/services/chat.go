package services

import (
	"context"

	"server/core"
	"server/db"
	"server/db/utils"

	"github.com/gorilla/websocket"
)

type ChatService struct {
	pool          *db.PostgresPool
	socketManager *core.SocketManager
}

func NewChatService(pool *db.PostgresPool) ChatService {
	newSocketManager := core.NewSocketManager()
	go newSocketManager.Run()
	return ChatService{pool: pool, socketManager: newSocketManager}
}

func (svc ChatService) GetMembers(chatID int) ([]utils.User, error) {
	members := make([]utils.User, 0)

	rows, err := svc.pool.Query(context.Background(), `select u.username, u.public_key from users u 
	inner join chat_members cm on u.id = cm.user_id where cm.chat_id = $1`, chatID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		member := utils.User{}
		err := rows.Scan(&member.Username, &member.PublicKey)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	return members, nil
}

func (svc ChatService) GetMessages(chatID, limit, offset int) ([]utils.Message, error) {
	messages := make([]utils.Message, 0)
	rows, err := svc.pool.Query(context.Background(), `select distinct m.id, m.body, m.created_at, u.username 
		from messages m 
		inner join chat_messages cm on m.id = cm.message_id
		inner join users u on m.user_messages = u.id
		where cm.chat_id = $1
		order by m.created_at desc limit $2 offset $3`, chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		msg := utils.Message{}
		err := rows.Scan(&msg.ID, &msg.Body, &msg.CreatedAt, &msg.Sender)
		if err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

func (svc ChatService) GetChats(username string) ([]utils.Chat, error) {
	user := utils.User{}
	err := svc.pool.QueryRow(context.Background(), `select u.id from users u where username = $1`, username).
		Scan(&user.ID)
	if err != nil {
		return nil, err
	}

	chats := make([]utils.Chat, 0)
	rows, err := svc.pool.Query(context.Background(),
		`with user_chats as (
			select distinct on (c.id) c.id, m.created_at as last_message_time, m.body as last_message from chats c 
			inner join chat_messages cm on c.id = cm.chat_id
			inner join messages m on cm.message_id = m.id
			where c.id in (
				select distinct chat_id from chat_members
				where user_id = $1
			) order by c.id, m.created_at desc
		) select u.username, uc.* from user_chats uc 
		inner join chat_members cm on uc.id = cm.chat_id
		inner join users u on cm.user_id = u.id
		where cm.user_id != $1 
		order by uc.last_message_time desc`, *user.ID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		chat := utils.Chat{}
		err := rows.Scan(&chat.RecipientUsername, &chat.ID, &chat.LastMessageTime, &chat.LastMessage)
		if err != nil {
			return nil, err
		}

		chats = append(chats, chat)
	}

	return chats, nil
}

func (svc ChatService) Subscribe(username string, ws *websocket.Conn) error {
	user := utils.User{}
	err := svc.pool.QueryRow(context.Background(), `select u.id from users u where username = $1`, username).
		Scan(&user.ID)
	if err != nil {
		return err
	}
	user.Username = &username

	newConnection := &core.Connection{
		User:          &user,
		Conn:          ws,
		SocketManager: svc.socketManager,
		Pool:          svc.pool,
	}
	svc.socketManager.Join <- newConnection

	return newConnection.Listen()
}
