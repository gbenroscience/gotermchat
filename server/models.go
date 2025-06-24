package server

import (
	"time"

	"com.itis.apps/gotermchat/cmd"
	"github.com/gorilla/websocket"
)

const GroupAliasMaxLen = 15

// Type constants for the Message struct
const (
	PrivateMessage = iota + 1
	GroupMessage
	BroadcastMessage
	ExitMessage
	HistoryRetriever
	NotificationErr
	NotificationSucc
	NotificationWarning
	NotificationInfo

	GroupAdd
	GroupMake
	GroupDel
	GroupRemoveMember
	GroupList
	GroupsForList
	ListCmds
)

// Special Commands used to control messages
const (
	PrivateCommand      = "<pr"   // Sends a private message: e.g <pr:08165779034> message body...
	GroupMessageCommand = "<grp"  // Sends a group message: e.g <grp:alias> message-body...
	ExitCommand         = "@exit" // Tells the server that this user is disconnecting from chat
	HistoryCommand      = "<hist" // Loads the last ...n... messages. <hist:n>

	GroupAddCommand          = "<grpadd"  //The admin adds a user: e.g <grpadd:08165779034:grpName>
	GroupMakeCommand         = "<grpmk"   // Creates a group: e.g  <grpmk:grpName:alias>
	GroupDelCommand          = "<grpdel"  // Deletes a group: e.g  <grpdel:alias>
	GroupRemoveMemberCommand = "<grprem"  // Removes a member from a group: e.g <grprem:08165779034:alias>
	GroupListCommand         = "<grpsls>" // Lists all your groups: e.g <grpsls> to list all groups created
	GroupsForListCommand     = "<grpsls"  // Lists all groups created by 0816577904: <grpsls:0816577904>
	ListCommands             = "<cmdls>"  // lists all available commands

)

var Commands = []string{PrivateCommandSyntax, GroupMessageCommandSyntax, ExitCommandSyntax,
	HistoryCommandSyntax, GroupAddCommandSyntax, GroupMakeCommandSyntax, GroupDelCommandSyntax, GroupRemoveMemberCommandSyntax, GroupListCommandSyntax,
	GroupsForListCommandSyntax, ListCommandsSyntax}

// The syntax for using the commands
const (
	PrivateCommandSyntax      = "<pr:08165779034>  message..."
	GroupMessageCommandSyntax = "<grp:grpAlias> message... Sends a message to a group of users. The grpAlias is the alias given to the group by the admin."
	ExitCommandSyntax         = "Type @exit. Disconnects you from the server normally"
	HistoryCommandSyntax      = " <hist:n> Loads n messages from your message history"

	GroupMakeCommandSyntax         = "<grpmk:grpName:alias> Creates a new group...e.g <grpmk:Days of our lives:days_group>"
	GroupAddCommandSyntax          = "<grpadd:08165779034:grpName> Adds the user with the given phone number to the group"
	GroupDelCommandSyntax          = "<grpdel:grpName> Deletes a group"
	GroupRemoveMemberCommandSyntax = "<grprem:08165779034:grpName> Removes a member from a group"
	GroupListCommandSyntax         = "<lsgrps> Lists all your groups"
	GroupsForListCommandSyntax     = "<lsgrps-for:08165779034> Lists all groups created by 08165779034"
	ListCommandsSyntax             = "@cmd. Lists all commands usable here"
)

// Server Constants
const (
	AppPhone           = "080-GTC-000"
	AppName            = "GoTermyChat"
	MongoURL           = "mongodb://localhost:27017"
	ExchangeKeysSecret = "Hast thou known not? hast thou.." // must be 32 bytes
	ChannelBufSize     = 100
)

// Client ... Models information for the User and its connection
type Client struct {
	Member  *cmd.User `json:"member"`
	Conn    *websocket.Conn
	server  *Server
	MsgChan chan *Message
	doneCh  chan bool
}

// Group A group chat model
type Group struct {
	ID         string   `bson:"_id" json:"id"`                  //ID - The unique ID for the group on the platform
	Name       string   `bson:"name" json:"name"`               //Name - A human friendly name for the group
	Alias      string   `bson:"alias" json:"alias"`             //Alias - A shorter name for the group... Must be unique within the group
	AdminPhone string   `bson:"admin_phone" json:"admin_phone"` //AdminPhone - The phone number of the group's admin
	Members    []string `bson:"members" json:"members"`         //Members - pointer reference to array storing the phone numbers of all group members
}

// Message ... Models information for the message payload
type Message struct {
	Msg        string    `bson:"msg" json:"msg"`
	ID         string    `bson:"_id" json:"id"`
	Time       time.Time `bson:"time" json:"time"`
	Phone      string    `bson:"phone" json:"phone"`
	SenderName string    `bson:"sender_name"  json:"sender_name"`
	Type       int       `bson:"msg_type" json:"msg_type"`
	GroupID    string    `bson:"group_id" json:"group_id"`
}

// Server ... The chat server
type Server struct {
	pattern    string
	messageMgr *MessageMgr
	userMgr    *UserMgr
	groupMgr   *GroupMgr
	messages   []Message
	clients    map[string]*Client
	groups     map[string]*Group
	addCh      chan *Client
	delCh      chan *Client
	sendAllCh  chan *Message
	doneCh     chan bool
	ErrCh      chan error
}

func (s Server) GetUserManager() *UserMgr {
	return s.userMgr
}

func (s Server) GetMessageManager() *MessageMgr {
	return s.messageMgr
}

func (s Server) GetGroupManager() *GroupMgr {
	return s.groupMgr
}
