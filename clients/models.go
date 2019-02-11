package clients

import (
	"math/rand"
	"time"

	"github.com/gbenroscience/gotermchat/server/utils"
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

	message.ID = utils.GenUlid()
	message.Type = BroadcastMessage
	return message
}
