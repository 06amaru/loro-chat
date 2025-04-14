package services

import (
	"context"
	"errors"
	"fmt"

	"server/core"
	"server/db"
	"server/db/utils"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
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
	rows, err := svc.pool.Query(context.Background(), `select m.id, m.body, m.created_at 
		from messages m inner join chat_messages cm on m.id = cm.message_id where cm.chat_id = $1 
		order by m.created_at desc limit = $2 offset = $3`, chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		msg := utils.Message{}
		err := rows.Scan(&msg.ID, &msg.Body, &msg.CreatedAt)
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
		select distinct on (c.id) c.id, m.created_at as last_message_time, m.content as last_message from chats c 
		inner join chat_messages cm on c.id = cm.chat_id
		inner join messages m on cm.message_id = m.id
		where c.id in (
			select distinct chat_id from chat_messages cm2
			inner join messages m2 on cm2.messages_id = m2.id
			where user_messages = $1
		) order by c.id, m.created_at desc
	) select * from user_chats order by last_message_time desc`, *user.ID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		chat := utils.Chat{}
		err := rows.Scan(&chat.ID, &chat.LastMessageTime, &chat.LastMessage)
		if err != nil {
			return nil, err
		}

		chats = append(chats, chat)
	}

	return chats, nil
}

func (svc ChatService) VerifyChat(to, from string) error {
	recipient := utils.User{}
	err := svc.pool.QueryRow(context.Background(), `select id, username from user where username = $1`, to).
		Scan(&recipient.ID, &recipient.Username)
	if err != nil {
		return err
	}

	sender := utils.User{}
	err = svc.pool.QueryRow(context.Background(), `select id, username from user where username = $1`, from).
		Scan(&sender.ID, &sender.Username)
	if err != nil {
		return err
	}

	chat := utils.Chat{}
	err = svc.pool.QueryRow(context.Background(), `select chat_id from public.chat_members where user_id in ($1, $2)
			group by chat_id`, *recipient.ID, *sender.ID).Scan(&chat.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	return fmt.Errorf("chat with ID %d exists between %s and %s", *chat.ID, to, from)
}

func (svc ChatService) CreateChat(to, from string) error {
	recipient := utils.User{}
	err := svc.pool.QueryRow(context.Background(), `select id, username from user where username = $1`, to).
		Scan(&recipient.ID, &recipient.Username)
	if err != nil {
		return err
	}

	sender := utils.User{}
	err = svc.pool.QueryRow(context.Background(), `select id, username from user where username = $1`, from).
		Scan(&sender.ID, &sender.Username)
	if err != nil {
		return err
	}

	err = svc.pool.Transaction(context.Background(), func(tx pgx.Tx) error {
		var chatID *uint

		// public chat is not encrypted
		err := tx.QueryRow(context.Background(), `insert into chats(type) values("public") returning id`).Scan(&chatID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(context.Background(), `insert into chat_members(chat_id, user_id) values($1, $2)`, *chatID, *sender.ID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(context.Background(), `insert into chat_members(chat_id, user_id) values($1, $2)`, *chatID, *recipient.ID)
		if err != nil {
			return err
		}

		return nil
	})

	return nil
}

func (svc ChatService) Subscribe(username string, ws *websocket.Conn) error {
	user := utils.User{}
	err := svc.pool.QueryRow(context.Background(), `select u.id from users u where username = $1`, username).
		Scan(&user.ID)
	if err != nil {
		return err
	}

	newConnection := &core.Connection{
		User:          user,
		Conn:          ws,
		SocketManager: svc.socketManager,
		Pool:          svc.pool,
	}
	svc.socketManager.Join <- newConnection

	return newConnection.Listen()
}
