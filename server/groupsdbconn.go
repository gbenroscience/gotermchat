package server

import (
	"context"
	"log"

	"com.itis.apps/gotermchat/database"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type GroupMgr struct {
	conn *database.MongoDB
}

func NewGroupMgr(pool *database.MongoDB) *GroupMgr {
	mgr := new(GroupMgr)
	mgr.conn = pool
	return mgr
}

// GetGroups - Returns all the Groups in the Groups Collection
func (gm *GroupMgr) GetGroups() ([]Group, error) {

	// Select database and collection
	collection := gm.conn.GetCollection("GroupsService", "Groups")
	var groups []Group
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var group Group
		if err := cursor.Decode(&group); err != nil {
			log.Fatal(err)
		}
		groups = append(groups, group)
	}

	// Check for any errors during iteration
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return groups, nil

}

// CreateOrUpdateGroup - Creates or Updates (Upsert) the Group in the Groups Collection with id parameter
func (gm *GroupMgr) CreateOrUpdateGroup(grp Group) bool {

	// Select database and collection
	collection := gm.conn.GetCollection("GroupsService", "Groups")
	filter := bson.M{"id": grp.ID}
	opts := options.Update().SetUpsert(true)
	g := bson.M{
		"$setOnInsert": grp,
	}
	_, err := collection.UpdateOne(context.Background(), filter, g, opts)
	if err != nil {
		log.Println("Error creating Group: ", err.Error())
		return false
	}

	return true
}

// ShowGroup - Returns the Group in the Groups Collection with name equal to the id parameter (id == name)
func (gm *GroupMgr) ShowGroup(id string) (Group, error) {
	// Select database and collection
	collection := gm.conn.GetCollection("GroupsService", "Groups")
	filter := bson.M{"id": id}
	var group Group
	err := collection.FindOne(context.Background(), filter).Decode(&group)
	return group, err
}

// DeleteGroup - Deletes the Group in the Groups Collection with same id-value as the id parameter
func (gm *GroupMgr) DeleteGroup(id string) bool {

	// Select database and collection
	collection := gm.conn.GetCollection("GroupsService", "Groups")

	filter := bson.D{{"id", id}}

	res, err := collection.DeleteOne(context.Background(), filter)
	return err == nil && res.DeletedCount > 0
}
