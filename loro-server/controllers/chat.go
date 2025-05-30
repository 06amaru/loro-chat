package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"server/db"
	"server/services"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  512,
		WriteBufferSize: 512,
		CheckOrigin: func(r *http.Request) bool {
			return r.Method == http.MethodGet
		},
	}
	// Get the JWT from the Authorization header
	//jwt := r.Header.Get("Authorization")

	// Validate the JWT
	// If the JWT is invalid, return an error
	//if !validateJWT(jwt) {
	//	return nil, fmt.Errorf("invalid JWT")
	//}

	return upgrader.Upgrade(w, r, nil)
}

type ChatController struct {
	svc services.ChatService
}

func NewChatController(repo *db.PostgresPool) ChatController {

	return ChatController{
		svc: services.NewChatService(repo),
	}
}

func (ctrl ChatController) GetMembers(c echo.Context) error {
	chatID, err := strconv.Atoi(c.QueryParam("chatID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	members, err := ctrl.svc.GetMembers(chatID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, members)
}

func (ctrl ChatController) GetMessages(c echo.Context) error {
	chatID := c.Param("chatID")
	if chatID == "" {
		return c.JSON(http.StatusBadRequest, fmt.Errorf("there is no chat id found"))
	}

	chatIDInt, err := strconv.Atoi(chatID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Errorf("chat id is not a number"))
	}

	limit := c.QueryParam("limit")
	if limit == "" {
		return c.JSON(http.StatusBadRequest, fmt.Errorf("limit required"))
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Errorf("limit is not a number"))
	}

	offset := c.QueryParam("offset")
	if offset == "" {
		return c.JSON(http.StatusBadRequest, fmt.Errorf("offset required"))
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Errorf("offset is not a number"))
	}

	messages, err := ctrl.svc.GetMessages(chatIDInt, limitInt, offsetInt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, messages)
}

func (ctrl ChatController) GetChats(c echo.Context) error {
	//context has a map where "user" is the default key for jwt
	token := c.Get("user").(*jwt.Token)
	username := token.Claims.(jwt.MapClaims)["username"].(string)

	chats, err := ctrl.svc.GetChats(username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, chats)
}

func (ctrl ChatController) JoinChat(c echo.Context) error {
	ws, err := Upgrade(c.Response(), c.Request())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer ws.Close()

	token := c.Get("user").(*jwt.Token)
	username := token.Claims.(jwt.MapClaims)["username"].(string)

	err = ctrl.svc.Subscribe(username, ws)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return nil
}
