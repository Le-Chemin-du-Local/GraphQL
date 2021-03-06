package commands

import (
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/database"
	"chemin-du-local.bzh/graphql/internal/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const COMMAND_STATUS_IN_PROGRESS = "INPROGRESS"
const COMMAND_STATUS_READY = "READY"
const COMMAND_STATUS_DONE = "DONE"

type Command struct {
	ID           primitive.ObjectID `bson:"_id"`
	CreationDate time.Time          `bson:"creationDate"`
	UserID       primitive.ObjectID `bson:"userID"`
	Status       string             `bson:"status"`
}

func (command *Command) ToModel() *model.Command {
	return &model.Command{
		ID:           command.ID.Hex(),
		CreationDate: command.CreationDate,
		User:         command.UserID.Hex(),
	}
}

func (command Command) IsLast() bool {
	filter := bson.D{{}}

	opts := options.FindOptions{}
	opts.SetLimit(1)
	opts.SetSort(bson.D{
		primitive.E{
			Key: "_id", Value: -1,
		},
	})

	lastCommand, err := GetFiltered(filter, &opts)

	if err != nil || len(lastCommand) <= 0 {
		return false
	}

	return lastCommand[0].ID == command.ID
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
		Status:       COMMAND_STATUS_IN_PROGRESS,
	}

	_, err = database.CollectionCommand.InsertOne(database.MongoContext, databaseCommand)

	if err != nil {
		return nil, err
	}

	return &databaseCommand, nil
}

// Mise à jour de base de données

func Update(changes *Command) error {
	filter := bson.D{
		primitive.E{
			Key:   "_id",
			Value: changes.ID,
		},
	}

	_, err := database.CollectionCommand.ReplaceOne(database.MongoContext, filter, changes)

	return err
}

// Getters

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

func GetPaginated(startValue *string, first int, filter *model.CommandsFilter) ([]Command, error) {
	var finalFilter bson.M

	if startValue != nil {
		objectID, err := primitive.ObjectIDFromHex(*startValue)

		if err != nil {
			return nil, err
		}

		finalFilter = bson.M{
			"_id": bson.M{
				"$gt": objectID,
			},
		}
	} else {
		finalFilter = bson.M{}
	}

	if filter != nil && filter.UserID != nil {
		userObjectID, err := primitive.ObjectIDFromHex(*filter.UserID)

		if err != nil {
			return nil, err
		}

		finalFilter = bson.M{
			"$and": []bson.M{
				finalFilter,
				{
					"userID": bson.M{
						"$eq": userObjectID,
					},
				},
			},
		}
	}

	if filter != nil && filter.Status != nil {
		finalFilter = bson.M{
			"$and": []bson.M{
				finalFilter,
				{
					"status": bson.M{
						"$in": filter.Status,
					},
				},
			},
		}
	}

	opts := options.Find()
	opts.SetLimit(int64(first))

	return GetFiltered(finalFilter, opts)
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

func GetStatus(commandID string) (*string, error) {
	databaseCommerceCommands, err := CommerceGetForCommand(commandID)

	if err != nil {
		return nil, err
	}

	result := COMMAND_STATUS_DONE

	for _, databaseCommerceCommand := range databaseCommerceCommands {
		if databaseCommerceCommand.Status == COMMERCE_COMMAND_STATUS_IN_PROGRESS {
			result = COMMAND_STATUS_IN_PROGRESS
			break
		}

		if databaseCommerceCommand.Status == COMMAND_STATUS_READY {
			result = COMMAND_STATUS_READY
		}
	}

	return &result, nil
}

func GetUser(commandID string) (*model.User, error) {
	databaseCommand, err := GetById(commandID)

	if err != nil {
		return nil, err
	}

	if databaseCommand == nil {
		return nil, &CommandNotFoundError{}
	}

	databaseUser, err := users.GetUserById(databaseCommand.UserID.Hex())

	if err != nil {
		return nil, err
	}

	if databaseUser == nil {
		return nil, &users.UserNotFoundError{}
	}

	return databaseUser.ToModel(), nil
}
