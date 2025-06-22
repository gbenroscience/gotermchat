package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"bytes"
	"flag"

	"com.itis.apps/gotermchat/cmd"
	"com.itis.apps/gotermchat/database"
	serv "com.itis.apps/gotermchat/server"
	"github.com/gorilla/websocket"
)

// SERVER ... The Server
var server *serv.Server

// Config ... Models information used to start the client connection
type ClientConfig struct {
	Phone    string
	Host     string
	Username string
	Password string
	Port     string
	Reg      bool
}

func main() {

	var port int
	var serverIP string

	flag.IntVar(&port, "p", 8080, "The application server will be started on this port")
	flag.StringVar(&serverIP, "h", "localhost", "The MongoDB URL to connect to")

	flag.Parse()

	var buffer bytes.Buffer

	//buffer.WriteString("")
	buffer.WriteString(":")
	buffer.WriteString(strconv.Itoa(port))

	//buffer now contains something like ":8080"

	fmt.Println("Creating MongoDB connection pool")
	pool, err := database.NewMongoDB(serv.MongoURL)
	if err != nil {
		fmt.Printf("error connecting to MongoDB: %v\n Exiting", err)
		return
	}

	fmt.Println("Created MongoDB connection pool... Now creating server")
	server = serv.NewServer("/ws", pool)
	http.HandleFunc("/ws/imaxine-that", wsHandler)
	http.HandleFunc("/", rootHandler)

	defer func() {
		if err := pool.Close(); err != nil {
			fmt.Printf("Couldn't close Mongo connection %v\n", err)
		}
	}()

	go server.StartListening()

	fmt.Printf("Server is listening @ %s\n", buffer.String())
	panic(http.ListenAndServe(buffer.String(), nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("index.html")
	if err != nil {
		fmt.Println("Could not open file.", err)
	}
	fmt.Fprintf(w, "%s", content)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	/*log.Print(r.Host)
	log.Println("----------------------")
	log.Println(r.Header.Get("Origin"))
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}*/

	var upgrader = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		msg := fmt.Sprintf("Could not open websocket connection: %v", err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	socketTimeOut := 30 * time.Minute
	err = conn.SetReadDeadline(time.Now().Add(socketTimeOut))
	if err != nil {
		msg := fmt.Sprintf("Could not set read deadline on socket %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	err = conn.SetWriteDeadline(time.Now().Add(socketTimeOut))
	if err != nil {
		msg := fmt.Sprintf("Could not set write deadline on socket %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// websocket handler
	go onConnect(conn, r, w)

}

func onConnect(ws *websocket.Conn, req *http.Request, response http.ResponseWriter) {
	data := req.FormValue("data")

	k, err := cmd.NewKryptik(serv.ExchangeKeysSecret, cmd.ModeCBC) //base64.RawURLEncoding.DecodeString(base64Str)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("...Error loading password decryptor!"))
		ws.Close()
		return
	}

	jsonData, err := k.Decrypt(data)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("...Error decrypting credentials!..."+fmt.Errorf("...err: %v\n", err).Error()))
		ws.Close()
		return
	}

	var config ClientConfig
	err = cmd.DecodeItem(jsonData, &config)

	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("...Error decoding client credentials"))
		ws.Close()
		return
	}

	pwd, err := k.Decrypt(config.Password)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("...Error decrypting password!..."+fmt.Errorf("...err: %v\n", err).Error()))
		ws.Close()
		return
	}

	if len(pwd) < 6 {
		ws.WriteMessage(websocket.TextMessage, []byte("...Decrypted Password too short!"))
		ws.Close()
		return
	}

	defer func() {
		err := ws.Close()
		if err != nil {
			server.ErrCh <- err
		}
	}()

	var u *serv.User
	/**
	Register this person
	*/
	if config.Reg {

		user := new(serv.User)
		user.Phone = config.Phone
		user.Name = config.Username
		user.Password = config.Password
		user.RegTime = time.Now()
		server.GetUserManager().CreateOrUpdateUser(*user)
		ws.WriteMessage(websocket.TextMessage, []byte("...Connected!"))
		u = user

	} else {

		if len(strings.Trim(config.Phone, " ")) >= 7 {
			user, err := server.GetUserManager().ShowUser(config.Phone)
			if err != nil {
				ws.WriteMessage(websocket.TextMessage, []byte("...Login Failed. Bad credentials."))
				ws.Close()
				return
			}

			if user.Password != config.Password { //do passwords match?
				ws.WriteMessage(websocket.TextMessage, []byte("...Login Failed. Incorrect credentials."))
				ws.Close()
				return
			}
			//valid user---allow login via phone
			ws.WriteMessage(websocket.TextMessage, []byte("...Login Successful...via phone"))
			u = &user

		} else if len(strings.Trim(config.Username, " ")) >= 4 { //usernames should be at least 4 digits long
			user, err := server.GetUserManager().ShowUserByUserName(config.Username)
			if err != nil {
				ws.WriteMessage(websocket.TextMessage, []byte("...Login Failed!! Bad credentials."))
				ws.Close()
				return
			}
			if user.Password != config.Password {
				ws.WriteMessage(websocket.TextMessage, []byte("...Login Failed!! Incorrect  credentials."))
				ws.Close()
				return
			}
			//valid user---allow login via username
			ws.WriteMessage(websocket.TextMessage, []byte("...Login Successful!! via username"))
			u = &user
		}
	}

	client := serv.NewClient(u, ws, server)
	client.Conn = ws
	client.MsgChan = make(chan *serv.Message, serv.ChannelBufSize)

	server.Add(client)
	client.Listen()
}
