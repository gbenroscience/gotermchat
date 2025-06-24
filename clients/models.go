package clients

import (
	"time"

	"com.itis.apps/gotermchat/cmd"
	"github.com/gorilla/websocket"
)

// GroupAliasMaxLen - The max characters that the Group alias can have
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
	GroupMessageCommand = "<grp"  // Sends a group message: e.g <grp:alias> message body...
	ExitCommand         = "@exit" // Tells the server that this user is disconnecting from chat
	HistoryCommand      = "<hist" // Loads the last ...n... messages. <hist:n>

	GroupAddCommand          = "<grpadd"  //The admin adds a user: e.g <grpadd:08165779034:alias>
	GroupMakeCommand         = "<grpmk"   // Creates a group: e.g  <grpmk:grpName:alias>
	GroupDelCommand          = "<grpdel"  // Deletes a group: e.g  <grpdel:alias>
	GroupRemoveMemberCommand = "<grprem"  // Removes a member from a group: e.g <grprem:08165779034:alias>
	GroupListCommand         = "<grpsls>" // Lists all your groups: e.g <grpsls> to list all groups created
	GroupsForListCommand     = "<grpsls"  // Lists all groups created by 0816577904: <grpsls:0816577904>
	ListCommands             = "<cmdls>"  // lists all available commands
)

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

// Defines more constants
const (
	AppName            = "GoTermyChat"
	ExchangeKeysSecret = "Hast thou known not? hast thou.." // must be 32 bytes
)

// Client ... Models a user and its connection information
type Client struct {
	Member   *cmd.User `json:"member"`
	Conn     *websocket.Conn
	MsgChan  chan *Message
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
	GroupID    string    `json:"group_id"`
}

// Config ... Models information used to start the client connection
type Config struct {
	Phone      string        `json:"phone"`
	Host       string        `json:"host"`
	Username   string        `json:"user_name"`
	Password   string        `json:"password"`
	Port       string        `json:"port"`
	Reg        bool          `json:"reg"`
	URLBuilder func() string `json:"-"`
}

func createMessage(msg string, timeT time.Time, phone string, senderName string) *Message {

	message := new(Message)

	message.Phone = phone
	message.Time = timeT
	message.Msg = msg
	message.SenderName = senderName

	message.ID = cmd.GenUlid()
	message.Type = BroadcastMessage
	return message
}
