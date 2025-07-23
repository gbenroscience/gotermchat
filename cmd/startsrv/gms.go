package main

import (
	"embed"
	"fmt"
	"os"
	"strconv"
	"time"

	"bytes"
	"flag"

	"com.itis.apps/gotermchat/database"
	serv "com.itis.apps/gotermchat/server"
	"github.com/apex/log"
	"github.com/apex/log/handlers/multi"
	"github.com/apex/log/handlers/text"
)

//go:embed html/*
var Assets embed.FS

func main() {

	var port int
	var serverIP string

	flag.IntVar(&port, "port", 8080, "The application server will be started on this port")
	flag.StringVar(&serverIP, "h", "localhost", "The MongoDB URL to connect to")

	flag.Parse()

	// Make handlers
	handlers := make([]log.Handler, 0)

	// Try to recover from a Crash
	defer func() {
		if err := recover(); err != nil {
			log.WithField("type", "CRASH").Error("System Crashed at " + time.Now().String())
		}
	}()

	fmt.Println()

	// Test Handler
	hCli := text.New(os.Stderr)
	handlers = append(handlers, hCli)

	// Test Handlers
	log.SetHandler(multi.New(handlers...))

	logger := log.NewEntry(&log.Logger{
		Handler: multi.New(handlers...), // Try to get the messages on time
	})

	// Get Hostname
	hostName, err := os.Hostname()
	if err != nil {
		log.WithError(err).Fatal("Can't get hostname")
	}

	logger = logger.WithFields(log.Fields{
		"app":     serv.AppName,
		"host":    hostName,
		"version": serv.Version,
	})

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
	server := serv.NewServer("/ws", pool)

	defer func() {
		if err := pool.Close(); err != nil {
			fmt.Printf("Couldn't close Mongo connection %v\n", err)
		}
	}()

	go server.StartListening()

	fmt.Printf("Server is listening @ %s\n", buffer.String())
	//panic(http.ListenAndServe(buffer.String(), nil))

	panic(serv.Start(server, logger, buffer.String(), Assets))
}
