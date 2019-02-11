package server

import (
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// GetUsers - Returns all the User in the Users Collection
func GetUsers() []User {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Println("Could not connect to mongo: ", err.Error())
		return nil
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("UserService").C("Users")
	var Users []User
	err = c.Find(bson.M{}).All(&Users)

	return Users
}

// CreateOrUpdateUser - Creates or Updates (Upsert) the User in the Users Collection with id parameter
func (u *User) CreateOrUpdateUser() bool {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Println("Could not connect to mongo: ", err.Error())
		return false
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("UserService").C("Users")
	_, err = c.UpsertId(u.Name, u)
	if err != nil {
		log.Println("Error creating User: ", err.Error())
		return false
	}
	return true
}

// ShowUser - Returns the User in the Users Collection with name equal to the id parameter (id == name)
func ShowUser(id string) (User, error) {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Println("Could not connect to mongo: ", err.Error())
		return User{}, err
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("UserService").C("Users")
	user := User{}
	err = c.Find(bson.M{"phone": id}).One(&user)

	return user, err
}

// DeleteUser - Deletes the User in the Users Collection with name equal to the id parameter (id == name)
func DeleteUser(id string) bool {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Println("Could not connect to mongo: ", err.Error())
		return false
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("UserService").C("Users")
	err = c.RemoveId(id)

	return true
}
