package server

import (
	"context"
	"log"

	"com.itis.apps/gotermchat/cmd"
	"com.itis.apps/gotermchat/database"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type UserMgr struct {
	conn *database.MongoDB
}

func NewUserMgr(pool *database.MongoDB) *UserMgr {
	mgr := new(UserMgr)
	mgr.conn = pool
	return mgr
}

// Getuser - Returns all the users in the users Collection
func (um *UserMgr) GetUsers() []cmd.User {

	// Select database and collection
	collection := um.conn.GetCollection("UserService", "Users")
	var users []cmd.User
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return users
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user cmd.User
		if err := cursor.Decode(&user); err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}

	// Check for any errors during iteration
	if err := cursor.Err(); err != nil {
		return users
	}

	return users

}

// CreateOrUpdateUser - Creates or Updates (Upsert) the User in the Users Collection with id parameter
func (um *UserMgr) CreateOrUpdateUser(user cmd.User) bool {

	// Select database and collection
	collection := um.conn.GetCollection("UserService", "Users")
	filter := bson.M{"phone": user.Phone}

	opts := options.Update().SetUpsert(true)

	u := bson.M{
		"$setOnInsert": user,
	}

	_, err := collection.UpdateOne(context.Background(), filter, u, opts)
	if err != nil {
		log.Println("Error creating User: ", err.Error())
		return false
	}

	return true
}

// ShowUser - Returns the User in the Users Collection with name equal to the id parameter (id == name)
func (um *UserMgr) ShowUser(phone string) (cmd.User, error) {
	// Select database and collection
	collection := um.conn.GetCollection("UserService", "Users")
	filter := bson.M{"phone": phone}
	var u cmd.User
	err := collection.FindOne(context.Background(), filter).Decode(&u)

	return u, err
}

// ShowUserByUserName - Returns the User in the Users Collection with name equal to the id parameter (id == name)
func (um *UserMgr) ShowUserByUserName(userName string) (cmd.User, error) {
	// Select database and collection
	collection := um.conn.GetCollection("UserService", "Users")
	filter := bson.M{"name": userName}
	var u cmd.User
	err := collection.FindOne(context.Background(), filter).Decode(&u)

	return u, err
}

// DeleteUser - Deletes the User in the Users Collection with name equal to the id parameter (id == phone)
func (um *UserMgr) DeleteUser(phone string) bool {

	// Select database and collection
	collection := um.conn.GetCollection("UserService", "Users")

	filter := bson.D{{"phone", phone}}

	res, err := collection.DeleteOne(context.Background(), filter)
	return err == nil && res.DeletedCount > 0
}
