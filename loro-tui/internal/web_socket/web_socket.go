package web_socket

import (
	"encoding/json"
	"fmt"
	"loro-tui/internal/models"
	"net/http"

	"github.com/gorilla/websocket"
)

type SocketClient struct {
	conn             *websocket.Conn
	IncomingMessages chan *models.Message
}

func NewWSocketClient(url, token string) (*SocketClient, error) {

	header := http.Header{}
	header.Set("Cookie", fmt.Sprintf("token=%s", token))

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8081/ws/join", header)
	if err != nil {
		return nil, fmt.Errorf("websocket new client [%v]", err)
	}

	return &SocketClient{conn: conn}, nil
}

func (ws *SocketClient) Send(message *models.Message) error {
	bytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("websocket fail writting bytes [%v]", err)
	}

	err = ws.conn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		return fmt.Errorf("websocket send [%v]", err)
	}
	return nil
}

func (ws *SocketClient) Listen() ([]byte, error) {
	_, message, err := ws.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("websocket listen [%v]", err)
	}

	return message, nil
}

func (ws *SocketClient) Close() {
	ws.conn.Close()
}
