package core

import (
	"fmt"

	"server/models"
)

/*
The socket manager stores all user connections. Each incoming message
will be forwarded by the socket manager to the corresponding user.
*/
type SocketManager struct {
	Messages    chan *models.Message
	Join        chan *Connection
	Leave       chan *Connection
	Id          int
	Connections map[string]*Connection
}

func NewSocketManager() *SocketManager {
	return &SocketManager{
		Messages:    make(chan *models.Message),
		Join:        make(chan *Connection),
		Leave:       make(chan *Connection),
		Connections: make(map[string]*Connection),
	}
}

func (c *SocketManager) Run() {
	fmt.Println("running chat ... ")
	for {
		select {
		case user := <-c.Join:
			c.add(user)
		case message := <-c.Messages:
			c.broadcast(message)
		case user := <-c.Leave:
			c.disconnect(user)
		}
	}
}
func (sm *SocketManager) add(con *Connection) {
	if _, ok := sm.Connections[*con.User.Username]; !ok {
		sm.Connections[*con.User.Username] = con

		body := fmt.Sprintf("%s is online", *con.User.Username)
		sender := con.User.Username
		sm.broadcast(&models.Message{
			Body:   &body,
			Sender: sender,
		})
	}
}

func (sm *SocketManager) broadcast(message *models.Message) {
	if message.Receiver == nil {
		// offline and online notification to all user
		for _, user := range sm.Connections {
			user.Send(message)
		}
		return
	}

	if user, ok := sm.Connections[*message.Sender]; ok {
		user.Send(message)
	}

	if user, ok := sm.Connections[*message.Receiver]; ok {
		user.Send(message)
	}
}

func (sm *SocketManager) disconnect(con *Connection) {
	if _, ok := sm.Connections[*con.User.Username]; ok {
		defer con.Conn.Close()
		delete(sm.Connections, *con.User.Username)

		body := fmt.Sprintf("%s is offline", *con.User.Username)
		sender := con.User.Username
		sm.broadcast(&models.Message{
			Body:   &body,
			Sender: sender,
		})
	}
}
