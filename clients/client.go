package clients

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"com.itis.apps/gotermchat/cmd"
	comd "com.itis.apps/gotermchat/cmd"
	"github.com/gbenroscience/gscanner/scanner"
	"github.com/gorilla/websocket"
)

func connect(conf *Config) {

	url := conf.URLBuilder()

	fmt.Println("Connecting.\nPlease wait...")
	var dialer *websocket.Dialer
	//dialer.HandshakeTimeout = time.Second * 1

	header := http.Header{}
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	header.Set("Access-Control-Allow-Headers", "Content-Type")
	header.Set("Origin", "http://localhost:8080")

	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		fmt.Println("Couldn't connect to ", url)
		fmt.Println(err)
		log.Fatal("Please try again. Tips: Check your network connections and your credentials.")
	}
	fmt.Println("Connected.")
	fmt.Println("-------------------------------------------------------------------------")
	fmt.Println(" Welcome to GoTermChat Chat")
	fmt.Println("-------------------------------------------------------------------------")

	client := new(Client)

	client.Conn = conn

	user := new(cmd.User)
	user.Name = conf.Username
	user.Phone = conf.Phone
	user.Password = conf.Password

	client.Member = user
	client.MsgChan = make(chan *Message, 10)
	client.Messages = make(map[string]*Message)

	client.receiver()

}

// StartConn -- Starts the connection.
func StartConn(conf *Config) {

	k, err := cmd.NewKryptik(ExchangeKeysSecret, cmd.ModeCBC) //base64.RawURLEncoding.DecodeString(base64Str)
	if err != nil {
		fmt.Println("...Error loading password encryptor!")
		return
	}

	jsonData, err := cmd.EncodeStruct(conf)
	if err != nil {
		fmt.Printf("...Error dumping config struct to JSON: %v\n", err)
		return
	}
	data, err := k.Encrypt(jsonData)
	if err != nil {
		fmt.Printf("...Error encrypting config: %v\n", err)
		return
	}
	conf.URLBuilder = func() string {

		var buffer bytes.Buffer

		buffer.WriteString("ws://")
		buffer.WriteString(conf.Host)
		buffer.WriteString(":")
		buffer.WriteString(conf.Port)
		buffer.WriteString("/ws/imaxine-that?data=")
		buffer.WriteString(data)

		url := buffer.String()

		return url
	}

	connect(conf)

}

func (client *Client) messenger(msgChan chan *Message) {

	//time.Sleep(time.Second * 2)
	//conn.WriteMessage(websocket.TextMessage, []byte(time.Now().Format("2006-01-02 15:04:05")))
	message := <-msgChan

	buf, err := json.Marshal(*message)
	if err != nil {
		fmt.Println("Unsupported Message")
	} else {
		err := client.Conn.WriteMessage(websocket.TextMessage, buf)
		if err == nil {
			client.Messages[message.ID] = message
			printMessage(*message)
		} else {
			fmt.Println("\nMessage not sent..." + err.Error())
		}

	}

}

func (client *Client) receiver() {

	go client.acceptInput()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			fmt.Println("read-err:", err)
			return
		}

		if strings.HasPrefix(string(message), "...") {
			fmt.Println(string(message))
		} else {
			var msg Message
			decoder := json.NewDecoder(bytes.NewBuffer(message))
			decoder.DisallowUnknownFields()
			err = decoder.Decode(&msg)
			if err != nil { //cant decode as Message
				var resp cmd.LoginResponse
				decoder := json.NewDecoder(bytes.NewBuffer(message))
				decoder.DisallowUnknownFields()
				err = decoder.Decode(&resp)
				if err != nil { //cant decode as login respone
					fmt.Printf("Error decoding message: %v\n", err)
					fmt.Println("Message causing error:", string(message))
					continue
				}
				client.Member = &resp.User
				fmt.Printf("Logged in User\n%s,\n%s\n____________________________________________________________\n", resp.User.Name, resp.User.Phone)
			}

			printMessage(msg)

			client.Messages[msg.ID] = &msg
		}
	} //end for loop
}

func (client *Client) acceptInput() {

	fmt.Println("Enter your messages here: ")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		chatMsg := scanner.Text()
		message := createMessage(chatMsg, time.Now(), client.Member.Phone, client.Member.Name)
		fmt.Printf("client-phone: " + client.Member.Phone)

		if strings.HasPrefix(chatMsg, ExitCommand) { //@exit
			client.normalClose()
			fmt.Println("Bye!")
			break
		} else if strings.HasPrefix(chatMsg, HistoryCommand) { //<hist=20>
			message.Type = HistoryRetriever
			fmt.Println("Not yet implemented! This will allow you to view past messages on the command line.\n The format is " + HistoryCommandSyntax + "\n e.g\n <hist=12> This will fetch 12 messages from your message history.")
			continue
		} else if strings.HasPrefix(chatMsg, PrivateCommand) { // syntax is: <pr=08176765555>

			startIndex := strings.Index(chatMsg, "<")
			endIndex := strings.Index(chatMsg, ">") + 1
			cmd := chatMsg[startIndex:endIndex]
			_, _, syntax := client.parseCommand(cmd)

			if syntax {
				message.Type = PrivateMessage
			} else {
				fmt.Println("Please check the syntax to use near: ", cmd, " Message not sent! \nThe syntax is "+PrivateCommandSyntax+". \nIt allows you to send a private message to the user that has that phone number.")
				continue
			}

		} else if strings.HasPrefix(chatMsg, GroupDelCommand) { // syntax is: <grpdel:grpName>

			startIndex := strings.Index(chatMsg, "<")
			endIndex := strings.Index(chatMsg, ">") + 1
			cmd := chatMsg[startIndex:endIndex]
			_, cmdVal, syntax := client.parseCommand(cmd)

			if syntax {
				message.Type = GroupDel
				var cmdMsg = comd.GroupDelete{
					NameOrAlias: cmdVal,
				}
				message.Msg, _ = comd.EncodeStruct(cmdMsg)
			} else {
				fmt.Println("Please check the syntax to use near: ", cmd, " Message not sent! \nThe syntax is "+GroupMessageCommandSyntax+". \nIt allows you to send a message to users in the specified group")
				continue
			}

		} else if strings.HasPrefix(chatMsg, GroupMakeCommand) { //<grpMk:grpName>

			startIndex := strings.Index(chatMsg, "<")
			endIndex := strings.Index(chatMsg, ">") + 1
			cmd := chatMsg[startIndex:endIndex]

			_, grpName, alias, err := client.parse3ArgsCommand(cmd)

			if err != nil {
				fmt.Println("error! to create new groups the format is " + GroupMakeCommandSyntax)
				continue
			} else {
				message.Type = GroupMake
				var cmdMsg = comd.GroupMake{
					Name:  grpName,
					Alias: alias,
				}
				message.Msg, _ = comd.EncodeStruct(cmdMsg)
			}

		} else if strings.HasPrefix(chatMsg, GroupAddCommand) {

			startIndex := strings.Index(chatMsg, "<")
			endIndex := strings.Index(chatMsg, ">") + 1
			cmd := chatMsg[startIndex:endIndex]

			_, memPhone, alias, err := client.parse3ArgsCommand(cmd)

			if err != nil {
				fmt.Println(err.Error() + " To add a user to your group, the format is " + GroupAddCommandSyntax + "\n")
				continue
			} else {
				message.Type = GroupAdd
				var cmdMsg = comd.GroupAdd{
					NameOrAlias: alias,
					Phone:       memPhone,
				}
				message.Msg, _ = comd.EncodeStruct(cmdMsg)
			}

		} else if strings.HasPrefix(chatMsg, GroupRemoveMemberCommand) {

			startIndex := strings.Index(chatMsg, "<")
			endIndex := strings.Index(chatMsg, ">") + 1
			cmd := chatMsg[startIndex:endIndex]

			_, memPhone, alias, err := client.parse3ArgsCommand(cmd)

			if err != nil {
				fmt.Println(err.Error() + " To remove a user from your group, the format is " + GroupRemoveMemberCommandSyntax + "\n")
				continue
			} else {
				message.Type = GroupRemoveMember
				var cmdMsg = comd.GroupRemoveMember{
					NameOrAlias: alias,
					MemberPhone: memPhone,
				}
				message.Msg, _ = comd.EncodeStruct(cmdMsg)
			}

		} else if strings.HasPrefix(chatMsg, GroupMessageCommand) { // syntax is: <grp:grpAlias>

			startIndex := strings.Index(chatMsg, "<")
			endIndex := strings.Index(chatMsg, ">") + 1
			textMsg := chatMsg[endIndex:]
			cmd := chatMsg[startIndex:endIndex]
			_, cmdVal, syntax := client.parseCommand(cmd)

			if syntax {
				message.Type = GroupMessage
				var cmdMsg = comd.GroupMessage{
					NameOrAlias: cmdVal,
					TextMessage: textMsg,
				}
				message.Msg, _ = comd.EncodeStruct(cmdMsg)
			} else {
				fmt.Println("Please check the syntax to use near: ", cmd, " Message not sent! \nThe syntax is "+GroupMessageCommandSyntax+". \nIt allows you to send a message to users in the specified group")
				continue
			}

		} else if strings.HasPrefix(chatMsg, GroupsForListCommand) { // syntax is: <grpsls:phone>

			startIndex := strings.Index(chatMsg, "<")
			endIndex := strings.Index(chatMsg, ">") + 1
			cmd := chatMsg[startIndex:endIndex]
			_, cmdVal, syntax := client.parseCommand(cmd)
			if syntax {
				message.Type = GroupsForList
				var cmdMsg = comd.GroupsListForUser{
					Phone: cmdVal,
				}
				message.Msg, _ = comd.EncodeStruct(cmdMsg)
			} else {
				fmt.Println("Please check the syntax to use near: ", cmd, " Message not sent! \nThe syntax is "+GroupsForListCommandSyntax+". \nIt allows you to send a message to users in the specified group")
				continue
			}

		} else if chatMsg == GroupListCommand { // syntax is: <grpsls>
			message.Type = GroupList
			message.Msg = GroupListCommand
		} else if chatMsg == ListCommands { // syntax is <cmdls>
			message.Type = ListCmds
			message.Msg = ListCommands
		} else {
			message.Type = BroadcastMessage
			fmt.Println("Sending a broadcast!!")
		}

		client.MsgChan <- message
		go client.messenger(client.MsgChan)
	}
	scanner.Err()
}

func (client *Client) normalClose() {
	err := client.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close: ", err)
		return
	}
}

func printMessage(msg Message) {

	fmt.Println("\n*****", msg.SenderName, "*****, ")

	fmt.Println(msg.Time.Format("Mon Jan _2 15:04:05 2006"))
	fmt.Println()
	fmt.Println(msg.Msg + "\n")
	fmt.Println()
	fmt.Println(msg.Phone)
	fmt.Println("_________________________________________________________")

}

/*
*
The name of the command is the first item returned.
The values of the command is the second item returned
The third item is a boolean value indicating if or not the
command's syntax is correct.
*/
func (client *Client) parseCommand(cmd string) (string, string, bool) {

	indexOfOpenTag := strings.Index(cmd, "<")
	indexOfColon := strings.Index(cmd, ":")
	indexOfCloseTag := strings.Index(cmd, ">")
	indexOfSpace := strings.Index(cmd, " ")
	countOpenTags := strings.Count(cmd, "<")
	countCloseTags := strings.Count(cmd, ">")
	countColons := strings.Count(cmd, ":")
	countOpenSpaces := strings.Count(cmd, " ")
	conditions := (indexOfOpenTag != -1) && (indexOfCloseTag != -1) && (indexOfColon != -1) && (indexOfSpace == -1) && (countOpenTags == 1) && (countCloseTags == 1) && (countColons == 1) && (countOpenSpaces == 0)

	arrangement := indexOfOpenTag < indexOfColon && indexOfColon < indexOfCloseTag

	validSyntax := conditions && arrangement

	if !validSyntax {
		return "", "", false
	}
	commandName := cmd[1:indexOfColon]
	commandVal := cmd[1+indexOfColon : indexOfCloseTag]

	return commandName, commandVal, validSyntax

}

// parse3ArgsCommand Parses commands of the form: <grpadd:08165779034:grpName>
func (client *Client) parse3ArgsCommand(cmd string) (string, string, string, error) {

	startIndex := strings.Index(cmd, "<")
	endIndex := strings.Index(cmd, ">") + 1

	scannerEngine := scanner.NewScanner(cmd[startIndex:endIndex], []string{"<", ">", ":"}, false)

	output := scannerEngine.Scan()

	if len(output) == 3 {

		command := output[0]

		if command == GroupMakeCommand[1:] { //<grpmk:grpName:alias>
			grpName := strings.Trim(output[1], " ")
			grpAlias := strings.Trim(output[2], " ")
			return command, grpName, grpAlias, nil
		}
		if command == GroupAddCommand[1:] { //<grpadd:08165779034:grpName>
			memPhone := strings.TrimSpace(output[1])
			alias := strings.TrimSpace(output[2])
			return command, memPhone, alias, nil
		}
		if command == GroupRemoveMemberCommand[1:] { //<grprem:08165779034:grpName>
			memPhone := strings.TrimSpace(output[1])
			alias := strings.TrimSpace(output[2])
			return command, memPhone, alias, nil
		}
		if command == GroupDelCommand[1:] { //<grpdel:grpName>
			memPhone := strings.TrimSpace(output[1])
			alias := strings.TrimSpace(output[2])
			return command, memPhone, alias, nil
		}
		if command == GroupListCommand[1:] { //<lsgrps> or <lsgrps:0816577904> to list all groups created

		}
		if command == GroupsForListCommand[1:] { //<lsgrps-for:0816577904>

		}

		return "", "", "", errors.New("Invalid command")

	}

	return "", "", "", errors.New("Invalid Command. Did you mean ")

}
