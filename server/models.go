package server

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

	GroupAdd
	GroupMake
	GroupDel
	GroupRemoveMember
)

// Special Commands used to control messages
const (
	PrivateCommand = "<pr"   // Sends a private message: e.g <pr:08165779034> message body...
	GroupCommand   = "<grp"  // Sends a group message: e.g <grp:grpName> message body...
	ExitCommand    = "@exit" // Tells the server that this user is disconnecting from chat
	HistoryCommand = "<hist" // Loads the last ...n... messages. <hist:n>

	GroupAddCommand          = "<grpadd" //The admin adds a user: e.g <grpadd:08165779034:grpName>
	GroupMakeCommand         = "<grpmk"  // Creates a group: e.g  <grpmk:grpName>
	GroupDelCommand          = "<grpdel" // Deletes a group: e.g  <grpdel:grpName>
	GroupRemoveMemberCommand = "<grpprg" // Removes a member from a group: e.g <grpprg:08165779034:grpName>
)

// Server Constants
const (
	MongoURL       = "localhost:27017"
	ChannelBufSize = 100
)

// User ... Models information for the User
type User struct {
	Name     string    `json:"name"`
	RegTime  time.Time `json:"regTime"`
	Phone    string    `json:"phone"`
	Password string    `json:"password"`
}

// Client ... Models information for the User and its connection
type Client struct {
	Member  *User `json:"member"`
	Conn    *websocket.Conn
	server  *Server
	MsgCHAN chan *Message
	doneCh  chan bool
}

//Group A group chat model
type Group struct {
	ID      string
	Name    string
	members []*string
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

// Server ... The chat server
type Server struct {
	pattern   string
	messages  []*Message
	clients   map[string]*Client
	groups    map[string]*Group
	addCh     chan *Client
	delCh     chan *Client
	sendAllCh chan *Message
	doneCh    chan bool
	ErrCh     chan error
}

func createMessage(msg string, timeT time.Time, phone string, senderName string) {
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
}
