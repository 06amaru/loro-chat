package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"loro-tui/internal/models"
	"net/http"
	"time"

	ws "loro-tui/internal/web_socket"
)

type NetworkClient struct {
	socketClient  *ws.SocketClient
	httpClient    *http.Client
	url           string
	MessageEvents chan *models.MessageEvent
	ChatEvents    chan *models.ChatEvent
	token         string
}

func NewNetworkClient(url string) (*NetworkClient, error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	response, err := client.Get(url + "/health-check")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf(response.Status)
	}

	return &NetworkClient{
		httpClient:    &client,
		url:           url,
		MessageEvents: make(chan *models.MessageEvent),
		ChatEvents:    make(chan *models.ChatEvent),
	}, nil
}

func (c *NetworkClient) AddListener() {
	defer func() {
		c.socketClient.Close()
	}()

	for {
		msg, err := c.socketClient.Listen()
		if err != nil {
			log.Print(err)
		}
		msgSerialized := &models.Message{}
		err = json.Unmarshal(msg, msgSerialized)
		if err != nil {
			log.Print(err)
		}

		c.MessageEvents <- &models.MessageEvent{Type: models.Incoming, Message: msgSerialized}
	}
}

func (c *NetworkClient) Login(payload models.LoginRequest) (*models.LoginResponse, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	response, err := c.doRequest("POST", c.url+"/login", bytes, headers)
	if err != nil {
		return nil, err
	}
	loginResponse := new(models.LoginResponse)
	err = json.Unmarshal(response, loginResponse)
	if err != nil {
		return nil, err
	}
	c.token = loginResponse.Token

	ws, err := ws.NewWSocketClient(c.url, loginResponse.Token)
	if err != nil {
		return nil, err
	}
	c.socketClient = ws

	return loginResponse, nil
}

func (c *NetworkClient) GetChats() ([]*models.Chat, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
		"Cookie":       fmt.Sprintf("token=%s", c.token),
	}

	response, err := c.doRequest("GET", c.url+"/api/chats", nil, headers)
	if err != nil {
		return nil, err
	}
	chatsResponse := make([]*models.Chat, 0)
	err = json.Unmarshal(response, &chatsResponse)
	if err != nil {
		return nil, err
	}

	return chatsResponse, nil
}

func (c *NetworkClient) GetMessages(chatID, limit, offset int) ([]*models.Message, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
		"Cookie":       fmt.Sprintf("token=%s", c.token),
	}

	path := fmt.Sprintf("/api/%d/messages?limit=%d&offset=%d", chatID, limit, offset)
	response, err := c.doRequest("GET", c.url+path, nil, headers)
	if err != nil {
		return nil, err
	}
	msgResponse := make([]*models.Message, 0)
	err = json.Unmarshal(response, &msgResponse)
	if err != nil {
		return nil, err
	}

	return msgResponse, nil
}

func (c *NetworkClient) doRequest(method, url string, body []byte, headers map[string]string) ([]byte, error) {
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return respBody, nil
}
