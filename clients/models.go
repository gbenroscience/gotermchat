package clients

import (
	"bytes"
	"math/rand"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// Type constants for the Message struct
const (
	PrivateMessage = iota + 1
	GroupMessage
	BroadcastMessage
	ExitMessage
	HistoryRetriever
)

// Special Commands used to control messages
const (
	PrivateCommand = "<pr"
	GroupCommand   = "<grp"
	ExitCommand    = "@exit"
	HistoryCommand = "<hist"
)

//User ...  Models user information
type User struct {
	Name     string    `json:"name"`
	RegTime  time.Time `json:"regTime"`
	Phone    string    `json:"phone"`
	Password string    `json:"password"`
}

// Client ... Models a user and its connection information
type Client struct {
	Member   *User `json:"member"`
	Conn     *websocket.Conn
	MsgCHAN  chan *Message
	doneCh   chan bool
	Messages map[string]*Message
}

// Message ... Models information for the message payload
type Message struct {
	Msg        string    `json:"msg"`
	ID         string    `json:"id"`
	Time       time.Time `json:"time"`
	Phone      string    `json:"phone"`
	SenderName string    `json:"sender_name"`
	Type       int       `json:"msg_type"`
}

// Config ... Models information used to start the client connection
type Config struct {
	Phone      string
	Host       string
	Username   string
	Password   string
	Port       string
	URLBuilder func() string
}

func createMessage(msg string, timeT time.Time, phone string, senderName string) *Message {

	rand.Seed(time.Now().Unix())

	message := new(Message)

	message.Phone = phone
	message.Time = timeT
	message.Msg = msg
	message.SenderName = senderName

	var buffer bytes.Buffer

	buffer.WriteString(phone)
	buffer.WriteString("-")
	randomVal := 10000 + rand.Intn(400000)
	buffer.WriteString(strconv.Itoa(randomVal))

	message.ID = buffer.String()
	message.Type = BroadcastMessage
	return message
}
