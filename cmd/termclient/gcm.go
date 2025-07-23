package main

import (
	"flag"
	"fmt"
	"net"
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

	if !isPrivateIP(hostname) {
		port = ""
		//discountenance port if is public ip to enforce servers to use standard port 80|443
	}

	fmt.Println("hostname: ", hostname)

	conf := &clientele.Config{
		Phone:    phone,
		Host:     hostname,
		Username: userName,
		Password: pwd,
		Port:     port,
		Reg:      isReg,
	}

	fmt.Printf("conf: %v\n", conf)

	clients.StartConn(conf)

}

func isPrivateIP(ipStr string) bool {
	if ipStr == "localhost" || ipStr == "127.0.0.1" {
		return true
	}
	v, _ := isLocalhost(ipStr)

	if v {
		return true
	}
	ip := net.ParseIP(ipStr)
	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, block := range privateBlocks {
		_, cidr, _ := net.ParseCIDR(block)
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func isLocalhost(addr string) (bool, error) {
	ips, err := net.LookupIP(addr)
	if err != nil {
		return false, fmt.Errorf("failed to lookup IP for address %s: %w", addr, err)
	}

	for _, ip := range ips {
		if ip.IsLoopback() {
			return true, nil
		}
	}
	return false, nil
}
