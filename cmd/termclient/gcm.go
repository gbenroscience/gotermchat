package main

import (
	"flag"

	"github.com/gbenroscience/gotermchat/clients"
	clientele "github.com/gbenroscience/gotermchat/clients"
)

// StartConnection ... Initializes the connection to the server
func main() {

	//userNamePtr := flag.String("word", "foo", "a string")

	//registered := flag.Bool("reg", false, "Boolean flag checking if user is just registering.\n If not present or set to false, it means the user is logging in, having registered before.")

	var userName string
	flag.StringVar(&userName, "username", "---", "Your user name")

	var phone string
	flag.StringVar(&phone, "phone", "---", "Your phone number")

	var hostname string
	flag.StringVar(&hostname, "h", "192.168.43.145", "The ip address or host name of the host server")

	var port string
	flag.StringVar(&port, "p", "8080", "The port number of the host server")

	flag.Parse()

	conf := &clientele.Config{
		Phone:    phone,
		Host:     hostname,
		Username: userName,
		Password: "******",
		Port:     port,
	}

	/*
		fmt.Println("phone:", phone)
		fmt.Println("userName:", userName)
		fmt.Println("registering? :", *registered)
		fmt.Println("tail:", flag.Args())
	*/

	clients.StartConn(conf)

}
