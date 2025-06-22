package server

import (
	"context"
	"log"

	"com.itis.apps/gotermchat/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageMgr struct {
	conn *database.MongoDB
}

func NewMessageMgr(pool *database.MongoDB) *MessageMgr {
	mgr := new(MessageMgr)
	mgr.conn = pool
	return mgr
}

// GetMessages - Returns all the Message in the Messages Collection
func (mm *MessageMgr) GetMessages() []Message {

	// Select database and collection
	collection := mm.conn.GetCollection("MessageService", "Messages")
	var messages []Message
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return messages
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var message Message
		if err := cursor.Decode(&message); err != nil {
			log.Fatal(err)
		}
		messages = append(messages, message)
	}

	// Check for any errors during iteration
	if err := cursor.Err(); err != nil {
		return messages
	}

	return messages

}

// CreateOrUpdateMessage - Creates or Updates (Upsert) the Message in the Messages Collection with id parameter
func (mm *MessageMgr) CreateOrUpdateMessage(msg Message) bool {

	// Select database and collection
	collection := mm.conn.GetCollection("MessageService", "Messages")
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateByID(context.Background(), msg.ID, msg, opts)
	if err != nil {
		log.Println("Error creating Message: ", err.Error())
		return false
	}

	return true
}

// ShowMessage - Returns the Message in the Messages Collection with name equal to the id parameter (id == name)
func (mm *MessageMgr) ShowMessage(id string) Message {
	// Select database and collection
	collection := mm.conn.GetCollection("MessageService", "Messages")
	filter := bson.M{"id": id}

	var message Message

	collection.FindOne(context.Background(), filter).Decode(&message)

	return message
}

// DeleteMessage - Deletes the Message in the Messages Collection with name equal to the id parameter (id == name)
func (mm *MessageMgr) DeleteMessage(id string) bool {

	// Select database and collection
	collection := mm.conn.GetCollection("MessageService", "Messages")

	filter := bson.D{{"id", id}}

	res, err := collection.DeleteOne(context.Background(), filter)
	return err == nil && res.DeletedCount > 0
}
