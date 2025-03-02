package internal

import "loro-tui/internal/models"

type ChatMessages struct {
	offset   int
	messages []*models.Message
}

type ChatHandler struct {
	selectedChat *models.Chat
	limit        int
	messagesMap  map[int]*ChatMessages
	chatsMap     map[int]*models.Chat
	chatList     []int
}

func NewChatHandler(limit int) *ChatHandler {
	return &ChatHandler{
		limit:        limit,
		messagesMap:  make(map[int]*ChatMessages),
		chatsMap:     make(map[int]*models.Chat),
		chatList:     make([]int, 0),
		selectedChat: nil,
	}
}

func (c *ChatHandler) saveChats(chats []*models.Chat) {
	for _, chat := range chats {
		c.chatsMap[chat.ChatID] = chat
		c.chatList = append(c.chatList, chat.ChatID)
	}
}

func (c *ChatHandler) setChatFirst(chatID int) {
	newList := make([]int, 0)
	for _, id := range c.chatList {
		if chatID == id {
			newList = append(newList, id)
		}
	}

	for _, id := range c.chatList {
		if chatID != id {
			newList = append(newList, id)
		}
	}

	c.chatList = newList
}

func (c *ChatHandler) saveMessages(chatID int, messages []*models.Message) *ChatMessages {
	if cmsgs, ok := c.messagesMap[chatID]; ok {
		cmsgs.offset += c.limit
		cmsgs.messages = append(cmsgs.messages, messages...)
	} else {
		cmsgs := &ChatMessages{
			offset:   c.limit,
			messages: make([]*models.Message, 0),
		}
		cmsgs.messages = append(cmsgs.messages, messages...)
		c.messagesMap[chatID] = cmsgs
	}

	return c.messagesMap[chatID]
}
