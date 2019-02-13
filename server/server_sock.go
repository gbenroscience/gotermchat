package server

import (
	"log"
	"strings"
	"time"

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
func (s *Server) makeGroup(cmd string, phone string) (Group, error) {

	startIndex := strings.Index(cmd, "<")
	endIndex := strings.Index(cmd, ">") + 1
	_, commandVal, validSyntax := parseCommand(cmd[startIndex:endIndex])

	if validSyntax {

		grp := &Group{
			ID:         utils.GenUlid(),
			Name:       commandVal,
			AdminPhone: phone,
			Members:    []*string{},
		}

		if s.userHasGroupByName(phone, grp.Name) {
			return Group{}, errors.New("Error: The Group, " + grp.Name + " is already amongst your groups!")
		}

		return *grp, nil
	}

	err := errors.New("The syntax of your command i.e `" + cmd + "` is wrong!\n Please use `<grpmk:grpName>` to create a new group")

	return Group{}, err
}

func createErrorMessage(errMsg string, timeT time.Time) *Message {

	message := new(Message)

	message.Phone = AppPhone
	message.Time = timeT
	message.Msg = errMsg
	message.SenderName = AppName

	message.ID = utils.GenUlid()
	message.Type = NotificationErr
	return message
}
func createSuccessMessage(succMsg string, timeT time.Time) *Message {

	message := new(Message)

	message.Phone = AppPhone
	message.Time = timeT
	message.Msg = succMsg
	message.SenderName = AppName

	message.ID = utils.GenUlid()
	message.Type = NotificationSucc
	return message
}

//listGroups - Lists all Groups that belong to a certain user
func (s *Server) listGroups(phone string) *[]Group {

	groups := make([]Group, 0)

	for _, v := range s.groups {
		if v.AdminPhone == phone {
			groups = append(groups, *v)
		}
	}

	return &groups
}

//userHasGroupByName - Checks that no group of the user having the phone number has the supplied name.
// phone is the phone number of the group creator and grpName is the name to check for
func (s *Server) userHasGroupByName(phone string, grpName string) bool {

	for _, v := range s.groups {
		if v.Name == grpName {
			return true
		}
	}

	return false
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
			msg.CreateOrUpdateMessage()
			log.Println("Send all-->>", msg)
			s.messages = append(s.messages, msg)
			if msg.Type == BroadcastMessage {
				log.Println("Broadcast detected:", msg)
				s.sendAll(msg)
			} else if msg.Type == PrivateMessage {
				log.Println("Private message detected:", msg)
				text := msg.Msg
				//<pr=08176765555>

				startIndex := strings.Index(text, "<")
				endIndex := strings.Index(text, ">") + 1
				cmd := text[startIndex:endIndex]
				_, userPhone, _ := parseCommand(cmd)

				destClient := s.clients[userPhone]
				if destClient != nil {
					destClient.Write(msg)
				}
			} else if msg.Type == GroupMake {
				//<grpmk:grpName>
				log.Println("Group create command detected:", msg)

				text := msg.Msg

				adminPhone := msg.Phone
				startIndex := strings.Index(text, "<")
				endIndex := strings.Index(text, ">") + 1
				cmd := text[startIndex:endIndex]
				grp, err := s.makeGroup(cmd, adminPhone)

				admin := s.clients[adminPhone]
				if err != nil {
					admin.Write(createErrorMessage(err.Error(), time.Now()))
				} else {
					s.groups[grp.ID] = &grp
					grp.CreateOrUpdateGroup()
					admin.Write(createSuccessMessage("The group, `"+grp.Name+"` was created successfully. \nStart adding members with"+
						GroupAddCommandSyntax+"\nSend a message to the group with: "+
						GroupMessageCommandSyntax+"\n Delete the group with: "+
						GroupDelCommandSyntax+"\n Remove a member with: "+
						GroupRemoveMemberCommandSyntax+"\n List the groups you created with: "+
						GroupListCommandSyntax+"\n List the groups someone created with: "+
						GroupsForListCommandSyntax, time.Now()))
				}
			} else if msg.Type == GroupAdd {
				//<grpadd:08165779034:grpName>
				log.Println("Group create command detected:", msg)

				text := msg.Msg

				s.groups[]

				adminPhone := msg.Phone
				startIndex := strings.Index(text, "<")
				endIndex := strings.Index(text, ">") + 1
				cmd := text[startIndex:endIndex]
				grp, err := s.makeGroup(cmd, adminPhone)

				admin := s.clients[adminPhone]
				if err != nil {
					admin.Write(createErrorMessage(err.Error(), time.Now()))
				} else {
					s.groups[grp.ID] = &grp
					grp.CreateOrUpdateGroup()
					admin.Write(createSuccessMessage("The group, `"+grp.Name+"` was created successfully. \nStart adding members with"+
						GroupAddCommandSyntax+"\nSend a message to the group with: "+
						GroupMessageCommandSyntax+"\n Delete the group with: "+
						GroupDelCommandSyntax+"\n Remove a member with: "+
						GroupRemoveMemberCommandSyntax+"\n List the groups you created with: "+
						GroupListCommandSyntax+"\n List the groups someone created with: "+
						GroupsForListCommandSyntax, time.Now()))
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
