package server

import (
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// GetGroups - Returns all the Groups in the Groups Collection
func GetGroups() []Group {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Println("Could not connect to mongo: ", err.Error())
		return nil
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("GroupsService").C("Groups")
	var Groups []Group
	err = c.Find(bson.M{}).All(&Groups)

	return Groups
}

// CreateOrUpdateGroup - Creates or Updates (Upsert) the Group in the Groups Collection with id parameter
func (msg *Group) CreateOrUpdateGroup() bool {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Println("Could not connect to mongo: ", err.Error())
		return false
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("GroupsService").C("Groups")
	_, err = c.UpsertId(msg.ID, msg)
	if err != nil {
		log.Println("Error creating Group: ", err.Error())
		return false
	}
	return true
}

// ShowGroup - Returns the Group in the Groups Collection with name equal to the id parameter (id == name)
func ShowGroup(id string) Group {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Println("Could not connect to mongo: ", err.Error())
		return Group{}
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("GroupsService").C("Groups")
	group := Group{}
	err = c.Find(bson.M{"id": id}).One(&group)

	return group
}

// DeleteGroup - Deletes the Group in the Groups Collection with name equal to the id parameter (id == name)
func DeleteGroup(id string) bool {
	session, err := mgo.Dial(MongoURL)
	if err != nil {
		log.Println("Could not connect to mongo: ", err.Error())
		return false
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("GroupsService").C("Groups")
	err = c.RemoveId(id)

	return true
}
