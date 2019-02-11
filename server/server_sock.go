package server

import (
	"log"
	"strings"

	"errors"

	"github.com/gbenroscience/gotermchat/server/utils"
)

// NewServer ... Create a new chat server
func NewServer(pattern string) *Server {
	messages := []*Message{}
	clients := make(map[string]*Client)
	groups := make(map[string]*Group)
	addCh := make(chan *Client)
	delCh := make(chan *Client)
	sendAllCh := make(chan *Message)
	doneCh := make(chan bool)
	errCh := make(chan error)

	return &Server{
		pattern,
		messages,
		clients,
		groups,
		addCh,
		delCh,
		sendAllCh,
		doneCh,
		errCh,
	}
}

//makeGroup Creates a group from a group command of the format: <grpmk:grpName>
//It returns the group name, a uid for the group or error if any
func (s *Server) makeGroup(cmd string) (Group, error) {

	startIndex := strings.Index(cmd, "<")
	endIndex := strings.Index(cmd, ">") + 1
	_, commandVal, validSyntax := parseCommand(cmd[startIndex:endIndex])

	if validSyntax {
		grp := &Group{
			ID:      utils.GenUlid(),
			Name:    commandVal,
			members: []*string{},
		}

		return *grp, nil
	}

	err := errors.New("The syntax of your command i.e `" + cmd + "` is wrong!\n Please use `<grpmk:grpName>` to create a new group")

	return Group{}, err
}

func createErrorMessage(errMsg string) Message {
	var msg Message = Message{}
}

// Add ... Adds a new client
func (s *Server) Add(c *Client) {
	s.addCh <- c
}

// Del ... Removes a client
func (s *Server) Del(c *Client) {
	s.delCh <- c
}

// SendAll ... Sends a message to all connected client
func (s *Server) SendAll(msg *Message) {
	s.sendAllCh <- msg
}

// Done ... Sends the done signal
func (s *Server) Done() {
	s.doneCh <- true
}

// Err ... Signals that an error occurred
func (s *Server) Err(err error) {
	s.ErrCh <- err
}

func (s *Server) sendPastMessages(c *Client) {
	for _, msg := range s.messages {
		c.Write(msg)
	}
}

func (s *Server) sendAll(msg *Message) {
	for _, c := range s.clients {
		if msg.Phone != c.Member.Phone {
			c.Write(msg)
		}
	}
}

//"ws://<ip>:port/x/y?token="qwerty"
// Listen and serve.
// It serves client connection and broadcast request.

// StartListening ... Makes the server to begin to listen for websocket connctions
func (s *Server) StartListening() {

	log.Println("Server listening")

	for {
		select {

		// Adds a new client
		case c := <-s.addCh:
			log.Println("New user connected")
			s.clients[c.Member.Phone] = c
			log.Println("Now", len(s.clients), "clients connected.")
			s.sendPastMessages(c)

		// del a client
		case c := <-s.delCh:
			log.Println("Delete client")
			delete(s.clients, c.Member.Phone)

		// broadcast message for all clients
		case msg := <-s.sendAllCh:
			log.Println("Send all-->>", msg)
			s.messages = append(s.messages, msg)
			if msg.Type == BroadcastMessage {
				log.Println("Broadcast detected:", msg)
				s.sendAll(msg)
			} else if msg.Type == PrivateMessage {
				log.Println("Private message detected:", msg)
				text := msg.Msg
				//<private=08176765555>

				startIndex := strings.Index(text, "<")
				endIndex := strings.Index(text, ">") + 1
				cmd := text[startIndex:endIndex]
				_, userPhone, _ := parseCommand(cmd)

				destClient := s.clients[userPhone]
				if destClient != nil {
					destClient.Write(msg)
				}
			} else if msg.Type == GroupMake {
				log.Println("Group create command detected:", msg)

				text := msg.Msg
				//<grpmk:grpName>
				startIndex := strings.Index(text, "<")
				endIndex := strings.Index(text, ">") + 1
				cmd := text[startIndex:endIndex]
				grp, err := s.makeGroup(cmd)

				if err != nil {
					adminPhone := msg.Phone
					admin := s.clients[adminPhone]
					admin.Write(err.Error())
				}
			}

		case err := <-s.ErrCh:
			log.Println("Error:", err.Error())

		case <-s.doneCh:
			return
		}
	}
}

/**
The name of the command is the first item returned.
The values of the command is the second item returned
The third item is a boolean value indicating if or not the
command's syntax is correct.

Parses commands of type <cmd:val> e.g <grpmk:grpName>, <pr:0816577904>
*/

func parseCommand(cmd string) (string, string, bool) {

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
