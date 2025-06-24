package server

import (
	"fmt"
	"log"
	"strings"
	"time"

	"com.itis.apps/gotermchat/cmd"
	comd "com.itis.apps/gotermchat/cmd"
	"com.itis.apps/gotermchat/database"
	"go.mongodb.org/mongo-driver/mongo"
)

// NewServer ... Create a new chat server
func NewServer(pattern string, pool *database.MongoDB) *Server {
	messageMgr := NewMessageMgr(pool)
	userMgr := NewUserMgr(pool)
	groupMgr := NewGroupMgr(pool)
	messages := messageMgr.GetMessages()
	clients := make(map[string]*Client)
	groups := make(map[string]*Group)
	addCh := make(chan *Client)
	delCh := make(chan *Client)
	sendAllCh := make(chan *Message)
	doneCh := make(chan bool)
	errCh := make(chan error)

	grps, err := groupMgr.GetGroups()
	if err != nil {
		fmt.Printf("Error loading groups: %v\n", err)
	}
	for _, group := range grps {
		groups[group.ID] = &group
	}

	return &Server{
		pattern,
		messageMgr,
		userMgr,
		groupMgr,
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

func createErrorMessage(errMsg string, timeT time.Time) *Message {

	message := new(Message)

	message.Phone = AppPhone
	message.Time = timeT
	message.Msg = errMsg
	message.SenderName = AppName

	message.ID = comd.GenUlid()
	message.Type = NotificationErr
	return message
}
func createSuccessMessage(succMsg string, timeT time.Time) *Message {

	message := new(Message)

	message.Phone = AppPhone
	message.Time = timeT
	message.Msg = succMsg
	message.SenderName = AppName

	message.ID = comd.GenUlid()
	message.Type = NotificationSucc
	return message
}

// listGroups - Lists all Groups that belong to a certain user
func (s *Server) listGroups(phone string) *[]Group {

	groups := make([]Group, 0)

	for _, v := range s.groups {
		if v.AdminPhone == phone {
			groups = append(groups, *v)
		}
	}

	return &groups
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
		c.Write(&msg)
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
			log.Println("Deleted client. ", len(s.clients), "clients connected.")
			delete(s.clients, c.Member.Phone)

			//Handle messages
		case msg := <-s.sendAllCh:
			s.messageMgr.CreateOrUpdateMessage(*msg)
			log.Println("Send all-->>", msg)
			s.messages = append(s.messages, *msg)
			if msg.Type == BroadcastMessage {
				log.Println("Broadcast detected:", *msg)
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
			} else if msg.Type == GroupMessage {
				log.Println("Group message detected:", msg)
				senderPhone := msg.Phone
				sender := s.clients[senderPhone]
				var groupMessage comd.GroupMessage
				err := cmd.DecodeItem(msg.Msg, &groupMessage)
				if err != nil {
					sender.Write(createErrorMessage(err.Error(), time.Now()))
					continue
				}

				group := s.groups[groupMessage.NameOrAlias] //NameOrAlias is same as the id
				if group != nil {
					for _, v := range group.Members {
						if v == senderPhone {
							continue //skip the sender
						}
						destClient := s.clients[v]
						if destClient != nil {
							destClient.Write(msg)
						} else {
							log.Println("User", v, "is not online")
						}
					}
				}
			} else if msg.Type == GroupMake {
				//<grpmk:grpName>
				log.Println("Group create command detected:", msg)

				senderPhone := msg.Phone
				sender := s.clients[senderPhone]
				var groupMake comd.GroupMake
				err := cmd.DecodeItem(msg.Msg, &groupMake)
				if err != nil {
					sender.Write(createErrorMessage(err.Error(), time.Now()))
					continue
				}

				_, err = s.groupMgr.ShowGroup(groupMake.Alias)
				if err == mongo.ErrNoDocuments { //no group by that alias exists yet
					grp := Group{
						ID:         groupMake.Alias, //not a mistake, aliases are unique
						Name:       groupMake.Name,
						Alias:      groupMake.Alias,
						AdminPhone: senderPhone,
						Members:    make([]string, 0),
					}
					admin := s.clients[senderPhone]
					s.groups[grp.ID] = &grp
					s.groupMgr.CreateOrUpdateGroup(grp)
					fmt.Printf("admin: %v\n", admin)
					admin.Write(createSuccessMessage("The group, `"+grp.Name+"` was created successfully. \nStart adding members with"+
						GroupAddCommandSyntax+"\nSend a message to the group with: "+
						GroupMessageCommandSyntax+"\n Delete the group with: "+
						GroupDelCommandSyntax+"\n Remove a member with: "+
						GroupRemoveMemberCommandSyntax+"\n List the groups you created with: "+
						GroupListCommandSyntax+"\n List the groups someone created with: "+
						GroupsForListCommandSyntax, time.Now()))
					continue
				}

				if err == nil {
					sender.Write(createErrorMessage("The group: "+groupMake.Alias+" exists already.", time.Now()))
				} else {
					sender.Write(createErrorMessage("Error in creating the group: "+groupMake.Alias, time.Now()))
				}
			} else if msg.Type == GroupAdd {
				//<grpadd:08165779034:grpName>
				log.Println("Group add member command detected:", msg)

				senderPhone := msg.Phone
				sender := s.clients[senderPhone]
				var groupAdd comd.GroupAdd
				err := cmd.DecodeItem(msg.Msg, &groupAdd)
				if err != nil {
					sender.Write(createErrorMessage(err.Error(), time.Now()))
					continue
				}

				grp := s.groups[groupAdd.NameOrAlias]
				if grp.AdminPhone != senderPhone {
					sender.Write(createErrorMessage("Only the group admin can add members", time.Now()))
					continue
				}
				grp.Members = append(grp.Members, groupAdd.Phone)
				s.groupMgr.CreateOrUpdateGroup(*grp)
				adminPhone := msg.Phone
				fmt.Println("s.clients.len: ", len(s.clients))
				admin := s.clients[adminPhone]
				member := s.clients[groupAdd.Phone]

				if err != nil {
					admin.Write(createErrorMessage(err.Error(), time.Now()))
					continue
				} else {
					fmt.Println("sender-phone: "+adminPhone+", s.clients: ", s.clients)
					admin.Write(createSuccessMessage("You have added "+member.Member.Name+" to "+grp.Alias, time.Now()))
				}
			} else if msg.Type == GroupRemoveMember {
				//<grprem:08165779034:grpName>
				log.Println("Group removed member command detected:", msg)

				senderPhone := msg.Phone
				sender := s.clients[senderPhone]
				var groupRem comd.GroupRemoveMember
				err := cmd.DecodeItem(msg.Msg, &groupRem)
				if err != nil {
					sender.Write(createErrorMessage(err.Error(), time.Now()))
					continue
				}

				grp := s.groups[groupRem.NameOrAlias]
				if grp.AdminPhone != senderPhone {
					sender.Write(createErrorMessage("Only the group admin can remove members", time.Now()))
					continue
				}
				newMembers := make([]string, 0)
				for _, anon := range grp.Members { //delete specifiedd member
					if anon != groupRem.MemberPhone {
						newMembers = append(newMembers, anon)
					}
				}
				grp.Members = newMembers
				s.groupMgr.CreateOrUpdateGroup(*grp)
				adminPhone := msg.Phone
				admin := s.clients[adminPhone]
				member := s.clients[groupRem.MemberPhone]

				if err != nil {
					admin.Write(createErrorMessage(err.Error(), time.Now()))
				} else {
					admin.Write(createSuccessMessage("You have removed "+member.Member.Name+" to "+grp.Alias, time.Now()))
				}
			} else if msg.Type == GroupDel {
				//<grpdel:alias>
				log.Println("Group delete command detected:", msg)

				senderPhone := msg.Phone
				sender := s.clients[senderPhone]
				var groupDel comd.GroupDelete
				err := cmd.DecodeItem(msg.Msg, &groupDel)
				if err != nil {
					sender.Write(createErrorMessage(err.Error(), time.Now()))
					continue
				}

				grp := s.groups[groupDel.NameOrAlias]
				if grp.AdminPhone == senderPhone {
					s.groupMgr.DeleteGroup(grp.ID)
					delete(s.groups, groupDel.NameOrAlias)
					sender.Write(createSuccessMessage("You have deleted the group: "+grp.Alias, time.Now()))
				} else {
					sender.Write(createErrorMessage("You are not the admin of this group: "+grp.Alias, time.Now()))
				}

			} else if msg.Type == GroupList {
				//<grpdel:alias>
				log.Println("Retrieve current User's Groups command detected:", msg)

				senderPhone := msg.Phone
				sender := s.clients[senderPhone]

				var groups []Group = make([]Group, 0)
				for _, v := range s.groups {
					if v.AdminPhone == senderPhone {
						groups = append(groups, *v)
					}
				}

				groupsJsn, err := cmd.EncodeStruct(groups)
				if err != nil {
					sender.Write(createErrorMessage("Error occurred while listing groups for: "+sender.Member.Name, time.Now()))
					continue
				}
				sender.Write(createSuccessMessage("Your groups: \n"+groupsJsn, time.Now()))
			} else if msg.Type == GroupsForList {
				//<grpdel:alias>
				log.Println("Retrieve Groups created by user command detected:", msg)

				senderPhone := msg.Phone
				sender := s.clients[senderPhone]
				var groupLs comd.GroupsListForUser
				err := cmd.DecodeItem(msg.Msg, &groupLs)
				if err != nil {
					sender.Write(createErrorMessage(err.Error(), time.Now()))
					continue
				}

				var groups []Group = make([]Group, 0)
				for _, v := range s.groups {
					if v.AdminPhone == groupLs.Phone {
						groups = append(groups, *v)
					}
				}

				groupsJsn, err := cmd.EncodeStruct(groups)
				if err != nil {
					sender.Write(createErrorMessage("Error occurred while listing groups for: "+sender.Member.Name, time.Now()))
					continue
				}
				sender.Write(createSuccessMessage(sender.Member.Name+"'s groups: \n"+groupsJsn, time.Now()))

			} else if msg.Type == ListCmds {
				log.Println("List Commands command detected:", msg)

				senderPhone := msg.Phone
				sender := s.clients[senderPhone]

				cmdsJsn, err := cmd.EncodeStruct(Commands)
				if err != nil {
					sender.Write(createErrorMessage("Error occurred while listing commands", time.Now()))
					continue
				}
				sender.Write(createSuccessMessage("Available commands: \n]n"+cmdsJsn+"\n\n", time.Now()))
			}

		case err := <-s.ErrCh:
			log.Println("Error:", err.Error())

		case <-s.doneCh:
			return
		}
	}
}

func (s Server) Shutdown() {
	s.messageMgr.conn.Close()
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
