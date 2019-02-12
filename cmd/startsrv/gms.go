package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"bytes"
	"flag"

	serv "github.com/gbenroscience/gotermchat/server"
	"github.com/gbenroscience/gotermchat/server/utils"
	"github.com/gorilla/websocket"
)

// SERVER ... The Server
var server *serv.Server

func main() {

	levComp := &utils.StringCompare{
		Source: "Gbemiro",
		Target: "Awele",
	}

	dist := levComp.ComputeDistance()

	fmt.Println("distance: ", dist)

	var port int

	flag.IntVar(&port, "p", 8080, "The application server will be started on this port")

	flag.Parse()

	var buffer bytes.Buffer

	buffer.WriteString(":")
	buffer.WriteString(strconv.Itoa(port))

	//buffer now contains someting like ":8080"

	server = serv.NewServer("/ws")
	http.HandleFunc("/ws/imaxine-that", wsHandler)
	http.HandleFunc("/", rootHandler)

	go server.StartListening()

	panic(http.ListenAndServe(buffer.String(), nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("index.html")
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
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	// websocket handler
	go onConnect(conn, r, w)

}

func onConnect(ws *websocket.Conn, req *http.Request, response http.ResponseWriter) {
	phone := req.FormValue("phone")

	regVal := req.FormValue("reg")

	reg := regVal != ""

	defer func() {
		err := ws.Close()
		if err != nil {
			server.ErrCh <- err
		}
	}()

	/**
	Register this person
	*/
	if reg {
		userName := req.FormValue("name")

		user := new(serv.User)
		user.Phone = phone
		user.Name = userName
		user.RegTime = time.Now()
		user.CreateOrUpdateUser()
		ws.WriteMessage(websocket.TextMessage, []byte("...Connected!"))

	} else {

		user, err := serv.ShowUser(phone)
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte("...Login Failed. Incorrect  credentials."))
		} else {
			//valid user---allow login
			if user.Phone == phone {
				ws.WriteMessage(websocket.TextMessage, []byte("...Login Successful"))
			}
		}

	}

	client := serv.NewClient(phone, ws, server)
	client.Conn = ws
	client.MsgCHAN = make(chan *serv.Message, serv.ChannelBufSize)

	server.Add(client)
	client.Listen()
}
