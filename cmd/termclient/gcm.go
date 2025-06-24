package main

import (
	"flag"
	"fmt"
	"os"

	// Import the term package
	"com.itis.apps/gotermchat/clients"
	clientele "com.itis.apps/gotermchat/clients"
	"com.itis.apps/gotermchat/cmd"
	"golang.org/x/term"
)

// StartConnection ... Initializes the connection to the server
func main() {

	//userNamePtr := flag.String("word", "foo", "a string")

	//registered := flag.Bool("reg", false, "Boolean flag checking if user is just registering.\n If not present or set to false, it means the user is logging in, having registered before.")

	var userName string
	flag.StringVar(&userName, "u", "", "Your user name")

	var phone string
	flag.StringVar(&phone, "ph", "", "Your phone number")

	var hostname string
	flag.StringVar(&hostname, "h", "localhost", "The ip address or host name of the host server")

	var port string
	flag.StringVar(&port, "p", "8080", "The port number of the host server")

	flag.Parse()

	isReg := len(flag.Args()) == 1 && flag.Arg(0) == "reg"

	// Get password securely
	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd())) // Read password from stdin
	if err != nil {
		fmt.Printf("Error reading password:%v\n", err)
		return
	}
	fmt.Println() // Add a newline after password input for better formatting
	password := string(bytePassword)

	k, err := cmd.NewKryptik(clientele.ExchangeKeysSecret, cmd.ModeCBC) //base64.RawURLEncoding.DecodeString(base64Str)
	if err != nil {
		fmt.Println("...Error loading password encryptor!")
		return
	}
	pwd, err := k.Encrypt(password)

	if err != nil {
		fmt.Println("...Error encrypting password!")
		return
	}

	conf := &clientele.Config{
		Phone:    phone,
		Host:     hostname,
		Username: userName,
		Password: pwd,
		Port:     port,
		Reg:      isReg,
	}

	clients.StartConn(conf)

}
