package internal

import (
	"loro-tui/internal/models"
	"loro-tui/internal/style"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	chatList      *tview.Table
	chatInput     *tview.InputField
	chatMesssages *tview.Table
	Pages         *tview.Pages
	usernameTV    *tview.TextView
	buttonNewChat *tview.Button
)

type Loro struct {
	*tview.Application
	*NetworkClient
	*ChatHandler
	username  string
	indexPage int
}

func focusInput(app *tview.Application, inputs []tview.Primitive) {
	for index, v := range inputs {
		if v.HasFocus() && index == len(inputs)-1 {
			app.SetFocus(inputs[0])
			break
		}
		if v.HasFocus() {
			app.SetFocus(inputs[index+1])
			break
		}
	}
}

func (l *Loro) setMessagesInTable(messages []*models.Message) {
	for i := len(messages) - 1; i >= 0; i-- {
		row := len(messages) - i - 1
		msg := messages[i]
		newCell := tview.NewTableCell(*msg.Body).SetExpansion(1)
		newCell.SetTextColor(style.LoroTheme.SecondaryTextColor)
		if *msg.Sender == l.username {
			newCell.SetAlign(tview.AlignRight)
		}
		chatMesssages.SetCell(row, 0, newCell)
	}
}

func createLoginPage(l *Loro) tview.Primitive {

	form := tview.NewForm()
	form.SetLabelColor(style.LoroTheme.SecondaryTextColor)
	form.SetFieldBackgroundColor(style.LoroTheme.MoreContrastBackgroundColor)
	form.SetFieldTextColor(style.LoroTheme.PrimitiveBackgroundColor)
	form.SetButtonStyle(style.ButtonStyle)
	form.SetButtonActivatedStyle(style.BtnActivatedStyle)
	form.AddInputField("Username", "", 20, nil, nil).
		AddPasswordField("Password", "", 20, '*', nil).
		AddButton("Login", func() {
			username := form.GetFormItem(0).(*tview.InputField).GetText()
			password := form.GetFormItem(1).(*tview.InputField).GetText()
			loginRequest := &models.LoginRequest{
				Username: username,
				Password: password,
			}
			_, err := l.NetworkClient.Login(*loginRequest)
			if err != nil {
				panic(err)
			}
			go l.AddListener()

			l.username = username
			l.ChatEvents <- &models.ChatEvent{Type: models.FetchChats}
			l.indexPage = 1
			usernameTV.SetText("Welcome " + username)
			Pages.SwitchToPage("chat")
		})

	return tview.NewGrid().
		SetColumns(0, 40, 0).
		SetRows(0, 9, 0).
		AddItem(form, 1, 1, 1, 1, 0, 0, true)
}

func (l *Loro) fetchChats() {
	chats, err := l.GetChats()
	if err != nil {
		panic(err)
	}

	l.saveChats(chats)

	for i, v := range chats {
		newCell := tview.NewTableCell(v.Username).SetExpansion(1)
		newCell.SetTextColor(style.LoroTheme.SecondaryTextColor)
		chatList.SetCell(i, 0, newCell)
	}
}

func (l *Loro) getMessages(chatID int, loadChat bool) {
	offset := 0
	chatMsg := l.messagesMap[chatID]
	if chatMsg != nil && loadChat {
		// if there are messages and key enter was trigger
		l.setMessagesInTable(chatMsg.messages)
		chatMesssages.ScrollToEnd()
		return
	}
	if chatMsg != nil {
		// set offset to get older messages from this
		offset = chatMsg.offset
	}
	msg, err := l.GetMessages(chatID, l.limit, offset)
	if err != nil {
		panic(err)
	}

	if len(msg) != 0 {
		chatMsg = l.saveMessages(chatID, msg)
	}
	l.setMessagesInTable(chatMsg.messages)
}

func (l *Loro) eventLoop() {
	for {
		select {
		case event := <-l.MessageEvents:
			l.handleMessageEvents(event)
		case event := <-l.ChatEvents:
			l.handleChatEvents(event)
		}
	}
}

func (l *Loro) handleMessageEvents(msg *models.MessageEvent) {
	switch msg.Type {
	case models.Incoming:
		// if chatID is nil then it is a offline/online notification
		// if chatID is not nil then is a new message
		if msg.ChatID != nil {
			if chat, ok := l.chatsMap[*msg.ChatID]; ok {
				chatMsg := l.messagesMap[chat.ChatID]
				if chat == l.selectedChat {
					// incoming message belongs to current chat
					chatMsg.offset++
					chatMsg.messages = append([]*models.Message{msg.Message}, chatMsg.messages...)
					l.setMessagesInTable(chatMsg.messages)
				} else {
					if chatMsg == nil {
						// there is no messages record
						newChatMsg := &ChatMessages{
							offset:   1, // storing the first incoming message
							messages: make([]*models.Message, 0),
						}
						newChatMsg.messages = append(newChatMsg.messages, msg.Message)
						l.messagesMap[chat.ChatID] = newChatMsg
					} else {
						chatMsg.offset++
						chatMsg.messages = append([]*models.Message{msg.Message}, chatMsg.messages...)
					}
				}
				chatList.Clear()
				l.setChatFirst(chat.ChatID)

			} else { // new chat was created
				username := *msg.Receiver
				if *msg.Receiver == l.username { // current user cannot be receiver
					username = *msg.Sender
				}
				l.chatsMap[*msg.ChatID] = &models.Chat{
					ChatID:   *msg.ChatID,
					Username: username,
				}
				newChatMsg := &ChatMessages{
					offset:   1, // storing the first incoming message
					messages: make([]*models.Message, 0),
				}
				newChatMsg.messages = append(newChatMsg.messages, msg.Message)
				l.messagesMap[*msg.ChatID] = newChatMsg
				l.selectedChat = l.chatsMap[*msg.ChatID]
				l.chatList = append([]int{*msg.ChatID}, l.chatList...)
				chatMesssages.Clear()
				l.setMessagesInTable(newChatMsg.messages)
				l.SetFocus(chatInput)
			}
			for i, chatID := range l.chatList {
				username := l.chatsMap[chatID].Username
				newCell := tview.NewTableCell(username).SetExpansion(1)
				newCell.SetTextColor(style.LoroTheme.SecondaryTextColor)
				chatList.SetCell(i, 0, newCell)
				if l.selectedChat != nil && chatID == l.selectedChat.ChatID {
					chatList.Select(i, 0)
				}
			}
		}
		l.Application.QueueUpdateDraw(func() {})
	case models.Forward:
		if err := l.socketClient.Send(msg.Message); err != nil {
			panic("ERROR SENDING MESSAGE")
		}
	}
}

func (l *Loro) handleChatEvents(event *models.ChatEvent) {
	switch event.Type {
	case models.FetchChats:
		l.Application.SetFocus(chatList)
		l.Application.QueueUpdateDraw(l.fetchChats)
	case models.GetMessages:
		// when you scroll up then fetch older messages
		l.getMessages(event.ChatID, false)
		l.Application.QueueUpdateDraw(func() {})
	case models.LoadChat:
		l.getMessages(event.ChatID, true)
		l.Application.SetFocus(chatInput)
		l.Application.QueueUpdateDraw(func() {})
	}
}

func createChatPage(l *Loro) tview.Primitive {
	chatList = tview.NewTable()
	chatInput = tview.NewInputField()
	chatInput.SetFieldBackgroundColor(style.LoroTheme.MoreContrastBackgroundColor)
	chatInput.SetFieldTextColor(style.LoroTheme.PrimitiveBackgroundColor)
	chatInput.SetBlurFunc(func() {
		chatInput.SetFieldBackgroundColor(style.LoroTheme.MoreContrastBackgroundColor)
	})
	chatInput.SetFocusFunc(func() {
		chatInput.SetFieldBackgroundColor(style.LoroTheme.ContrastBackgroundColor)
	})
	chatMesssages = tview.NewTable()
	buttonNewChat = tview.NewButton("New Chat")
	inputs := []tview.Primitive{
		chatList,
		chatInput,
		chatMesssages,
		buttonNewChat,
	}

	chatInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if l.selectedChat != nil {
				input := chatInput.GetText()
				if len(input) > 0 {
					message := &models.Message{
						Body:     &input,
						Sender:   &l.username,
						ChatID:   &l.selectedChat.ChatID,
						Receiver: &l.selectedChat.Username,
					}
					l.MessageEvents <- &models.MessageEvent{Type: models.Forward, Message: message}
					chatInput.SetText("")
				}
			}

		case tcell.KeyTab:
			focusInput(l.Application, inputs)
		}
		return event
	})

	chatMesssages.SetBorder(true)
	chatMesssages.SetBorderColor(style.LoroTheme.MoreContrastBackgroundColor)
	chatMesssages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := chatMesssages.GetSelection()
		switch event.Key() {
		case tcell.KeyUp:
			if row == 0 && l.selectedChat != nil {
				l.ChatEvents <- &models.ChatEvent{Type: models.GetMessages, ChatID: l.selectedChat.ChatID}
			}
			return event
		case tcell.KeyTab:
			focusInput(l.Application, inputs)
		}

		return event
	})

	l.Application.SetFocus(chatList)
	chatList.SetBorder(true)
	chatList.SetBorderColor(style.LoroTheme.MoreContrastBackgroundColor)
	chatList.SetSelectable(true, true)
	chatList.SetSelectedStyle(style.CellSelectedtyle)
	chatList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			chatMesssages.Clear()
			row, _ := chatList.GetSelection()
			chatID := l.chatList[row]
			l.selectedChat = l.chatsMap[chatID]
			l.ChatEvents <- &models.ChatEvent{Type: models.LoadChat, ChatID: chatID}
		case tcell.KeyTab:
			focusInput(l.Application, inputs)
		}
		return event
	})

	usernameTV = tview.NewTextView()
	usernameTV.SetTextColor(style.LoroTheme.SecondaryTextColor)

	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow)

	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false), width, 1, true).
			AddItem(nil, 0, 1, false)
	}

	form := tview.NewForm()
	form.SetLabelColor(style.LoroTheme.SecondaryTextColor)
	form.SetFieldBackgroundColor(style.LoroTheme.MoreContrastBackgroundColor)
	form.SetFieldTextColor(style.LoroTheme.PrimitiveBackgroundColor)
	form.SetButtonStyle(style.ButtonStyle)
	form.SetButtonActivatedStyle(style.BtnActivatedStyle)
	form.AddInputField("To", "", 30, nil, nil).
		AddInputField("Message", "", 30, nil, nil).
		AddButton("Send", func() {
			toUsername := form.GetFormItem(0).(*tview.InputField).GetText()
			body := form.GetFormItem(1).(*tview.InputField).GetText()
			message := &models.Message{
				Receiver: &toUsername,
				Sender:   &l.username,
				Body:     &body,
			}

			l.MessageEvents <- &models.MessageEvent{Type: models.Forward, Message: message}

			Pages.RemovePage("modal")
			// l.Application.SetFocus(chatList)
		})

	buttonNewChat.SetStyle(style.ButtonStyle)
	buttonNewChat.SetActivatedStyle(style.BtnActivatedStyle)
	buttonNewChat.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			Pages.AddPage("modal", modal(form, 10, 10), true, true)
		case tcell.KeyTab:
			focusInput(l.Application, inputs)
		}
		return event
	})

	menuTitle := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(usernameTV, 0, 1, false).
		AddItem(buttonNewChat, 0, 1, false)

	chatLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(chatList,
			0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(chatMesssages, 0, 1, false).
			AddItem(chatInput, 1, 1, false),
			0, 4, false)
	chatLayout.SetBorder(false)

	mainLayout.
		AddItem(menuTitle, 0, 1, false).
		AddItem(chatLayout, 0, 12, false)
	mainLayout.SetBorder(false)

	return mainLayout
}

func NewLoro(url string) (*Loro, error) {

	networkClient, err := NewNetworkClient(url)
	if err != nil {
		return nil, err
	}

	loro := &Loro{
		Application:   tview.NewApplication(),
		NetworkClient: networkClient,
		indexPage:     0,
		ChatHandler:   NewChatHandler(5),
	}

	go loro.eventLoop()

	login := createLoginPage(loro)
	chat := createChatPage(loro)

	Pages = tview.NewPages()
	Pages.AddPage("login", login, true, loro.indexPage == 0)
	Pages.AddPage("chat", chat, true, loro.indexPage == 1)

	return loro, nil
}
