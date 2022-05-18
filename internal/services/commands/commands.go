package commands

import (
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Command struct {
	ID           primitive.ObjectID `bson:"_id"`
	CreationDate time.Time          `bson:"creationDate"`
	UserID       primitive.ObjectID `bson:"userID"`
}

func (command *Command) ToModel() *model.Command {
	return &model.Command{
		ID:           command.ID.Hex(),
		CreationDate: command.CreationDate,
		User:         command.UserID.Hex(),
	}
}

func Create(input model.NewCommand) (*Command, error) {
	userObjectID, err := primitive.ObjectIDFromHex(input.User)

	if err != nil {
		return nil, err
	}

	databaseCommand := Command{
		ID:           primitive.NewObjectID(),
		CreationDate: time.Now(),
		UserID:       userObjectID,
	}

	_, err = database.CollectionCommand.InsertOne(database.MongoContext, databaseCommand)

	if err != nil {
		return nil, err
	}

	return &databaseCommand, nil
}

func GetAll() ([]Command, error) {
	filter := bson.D{{}}

	return GetFiltered(filter, nil)
}

func GetById(id string) (*Command, error) {
	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	filter := bson.D{
		primitive.E{
			Key:   "_id",
			Value: objectId,
		},
	}

	commands, err := GetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(commands) == 0 {
		return nil, nil
	}

	return &commands[0], nil
}

func GetFiltered(filter interface{}, opts *options.FindOptions) ([]Command, error) {
	commands := []Command{}

	cursor, err := database.CollectionCommand.Find(database.MongoContext, filter, opts)

	if err != nil {
		return commands, err
	}

	for cursor.Next(database.MongoContext) {
		var command Command

		err := cursor.Decode(&command)

		if err != nil {
			return commands, err
		}

		commands = append(commands, command)
	}

	if err := cursor.Err(); err != nil {
		return commands, err
	}

	return commands, nil
}
