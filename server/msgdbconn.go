package server

import (
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// GetMessages - Returns all the Message in the Messages Collection
func GetMessages() []Message {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Println("Could not connect to mongo: ", err.Error())
		return nil
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("MessageService").C("Messages")
	var Messages []Message
	err = c.Find(bson.M{}).All(&Messages)

	return Messages
}

// CreateOrUpdateMessage - Creates or Updates (Upsert) the Message in the Messages Collection with id parameter
func (msg *Message) CreateOrUpdateMessage() bool {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Println("Could not connect to mongo: ", err.Error())
		return false
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("MessageService").C("Messages")
	_, err = c.UpsertId(msg.ID, msg)
	if err != nil {
		log.Println("Error creating Message: ", err.Error())
		return false
	}
	return true
}

// ShowMessage - Returns the Message in the Messages Collection with name equal to the id parameter (id == name)
func ShowMessage(id string) Message {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Println("Could not connect to mongo: ", err.Error())
		return Message{}
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("MessageService").C("Messages")
	Message := Message{}
	err = c.Find(bson.M{"id": id}).One(&Message)

	return Message
}

// DeleteMessage - Deletes the Message in the Messages Collection with name equal to the id parameter (id == name)
func DeleteMessage(id string) bool {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Println("Could not connect to mongo: ", err.Error())
		return false
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("MessageService").C("Messages")
	err = c.RemoveId(id)

	return true
}
