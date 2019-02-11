package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/gorilla/websocket"
)

// NewClient ... Creates a new Client
func NewClient(phone string, ws *websocket.Conn, server *Server) *Client {

	if ws == nil {
		panic("ws cannot be nil")
	}

	if server == nil {
		panic("server cannot be nil")
	}

	ch := make(chan *Message, ChannelBufSize)
	doneCh := make(chan bool)
	user, err := ShowUser(phone)
	if err != nil {
		log.Println("BAD CODE HERE")
		return &Client{}
	}

	return &Client{&user, ws, server, ch, doneCh}
}

func (c *Client) Write(msg *Message) {
	select {
	case c.MsgCHAN <- msg:
	default:
		c.server.Del(c)
		err := fmt.Errorf("client %s is disconnected", c.Member.Phone)
		c.server.Err(err)
	}
}

// Done ... Sends a done signal
func (c *Client) Done() {
	c.doneCh <- true
}

// Listen for Write and Read request via channel
func (c *Client) Listen() {
	go c.listenWrite()
	c.listenRead()
}

// Listen for write request via channel
func (c *Client) listenWrite() {
	log.Println("Listening write to client")
	for {
		select {

		// send message to the client
		case msg := <-c.MsgCHAN:
			log.Println("Send:", msg.Msg)

			byteArr, err := json.Marshal(msg)

			if err != nil {

			} else {
				c.Conn.WriteMessage(websocket.TextMessage, []byte(byteArr))
			}

			// receive done request
		case <-c.doneCh:
			c.server.Del(c)
			c.doneCh <- true // for listenRead method
			return
		}
	}
}

// Listen read request via chanel
func (c *Client) listenRead() {
	log.Println("Listening read from client")
	for {
		select {

		// receive done request
		case <-c.doneCh:
			c.server.Del(c)
			c.doneCh <- true // for listenWrite method
			return

			// read data from websocket connection
		default:
			var msg Message

			msgType, byteArr, err := c.Conn.ReadMessage()

			if err != nil {
				if err == io.EOF {
					c.doneCh <- true
				} else if websocket.IsCloseError(err, websocket.CloseGoingAway) {
					//log.Printf("error: %v", err)
					break
				} else {
					//log.Printf("error: %v-------------------------444------------------------------------", err)
					return
				}

				c.server.Err(err)
			} else {
				if msgType == websocket.TextMessage {
					//log.Printf("MESSAGE: %v-------------------------------555------------------------------", string(byteArr))
					json.Unmarshal(byteArr, &msg)
					//log.Printf("msg: %v-------------------------------555------------------------------", msg.Msg)
					c.server.SendAll(&msg)
				} else {
					json.Unmarshal(byteArr, &msg)
					c.server.SendAll(&msg)
				}
			}
		}
	}

}
