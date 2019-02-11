package clients

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func connect(conf *Config) {

	url := conf.URLBuilder()

	fmt.Println("Connecting to ", url, "\n\nPlease wait...")
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
	fmt.Println("Connected to ", url, ":\n\n")
	fmt.Println("-------------------------------------------------------------------------")
	fmt.Println(" Welcome to GoTermyChat Chat")
	fmt.Println("-------------------------------------------------------------------------")

	client := new(Client)

	client.Conn = conn

	user := new(User)
	user.Name = conf.Username
	user.Phone = conf.Phone
	user.Password = conf.Password

	client.Member = user
	client.MsgCHAN = make(chan *Message, 10)
	client.Messages = make(map[string]*Message)

	client.receiver()

}

//StartConn -- Starts the connection.
func StartConn(conf *Config) {

	conf.URLBuilder = func() string {

		var buffer bytes.Buffer

		buffer.WriteString("ws://")
		buffer.WriteString(conf.Host)
		buffer.WriteString(":")
		buffer.WriteString(conf.Port)
		buffer.WriteString("/ws/imaxine-that?name=")

		buffer.WriteString(conf.Username)
		buffer.WriteString("&phone=")
		buffer.WriteString(conf.Phone)

		buffer.WriteString("&reg=")

		if len(flag.Args()) == 1 && flag.Arg(0) == "reg" {
			buffer.WriteString(strconv.FormatBool(true))
		} else {
			buffer.WriteString(strconv.FormatBool(false))
		}

		url := buffer.String()

		return url
	}

	connect(conf)

}

func (client *Client) messenger(msgChan chan *Message) {

	//time.Sleep(time.Second * 2)
	//conn.WriteMessage(websocket.TextMessage, []byte(time.Now().Format("2006-01-02 15:04:05")))
	message := <-msgChan

	log.Println("Broadcast detected:", message)

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
			fmt.Println("read:", err)
			return
		}

		if strings.HasPrefix(string(message), "...") {
			fmt.Println(string(message))
		} else {

			var msg Message

			err = json.Unmarshal(message, &msg)
			if err != nil {
				fmt.Println("Couldn't decode the message:", string(message), err)
				return
			}

			printMessage(msg)

			client.Messages[msg.ID] = &msg
		}

	}
}

func (client *Client) acceptInput() {

	fmt.Println("Enter your messages here: ")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		chatMsg := scanner.Text()
		message := createMessage(chatMsg, time.Now(), client.Member.Phone, client.Member.Name)

		if strings.HasPrefix(chatMsg, ExitCommand) { //@exit
			client.normalClose()
			fmt.Println("Bye!")
			break
		} else if strings.HasPrefix(chatMsg, HistoryCommand) { //<hist=20>
			message.Type = HistoryRetriever
			fmt.Println("Not yet implemented! This will allow you to view past messages on the command line.\n The format is <hist=number>\n e.g\n <hist=12> This will fetch 12 messages from your message history.")
			continue
		} else if strings.HasPrefix(chatMsg, PrivateCommand) { // syntax is: <pr=08176765555>

			startIndex := strings.Index(chatMsg, "<")
			endIndex := strings.Index(chatMsg, ">") + 1
			cmd := chatMsg[startIndex:endIndex]
			_, _, syntax := client.parseCommand(cmd)

			if syntax {
				message.Type = PrivateMessage
			} else {
				fmt.Println("Please check the syntax to use near: ", cmd, " Message not sent! \nThe syntax is <private=phone-number>. \nIt allows you to send a private message to the user that has that phone number.")
				continue
			}

		} else {
			message.Type = BroadcastMessage
			fmt.Println("Sending a broadcast!!")
		}

		client.MsgCHAN <- message
		go client.messenger(client.MsgCHAN)
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
	fmt.Println(msg.Msg, "\n")
	fmt.Println()
	fmt.Println(msg.Phone)
	fmt.Println("_________________________________________________________")

}

/**
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
