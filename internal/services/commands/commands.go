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
const COMMAND_STATUS_CANCELED = "CANCELED"

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

func (command Command) IsLast(commandsService CommandsService) bool {
	filter := bson.D{{}}

	opts := options.FindOptions{}
	opts.SetLimit(1)
	opts.SetSort(bson.D{
		primitive.E{
			Key: "_id", Value: -1,
		},
	})

	lastCommand, err := commandsService.GetFiltered(filter, &opts)

	if err != nil || len(lastCommand) <= 0 {
		return false
	}

	return lastCommand[0].ID == command.ID
}

// Le service
type commandsService struct {
	UsersService users.UsersService
}

type CommandsService interface {
	Create(input model.NewCommand) (*Command, error)
	Update(changes *Command) error
	GetAll() ([]Command, error)
	GetById(id string) (*Command, error)
	GetPaginated(startValue *string, first int, filter *model.CommandsFilter) ([]Command, error)
	GetFiltered(filter interface{}, opts *options.FindOptions) ([]Command, error)
	GetTheoricalStatus(commandID string, commerceCommandsService CommerceCommandsService) (*string, error)
	GetUser(commandID string) (*model.User, error)
}

func NewCommandsService(usersService users.UsersService) *commandsService {
	return &commandsService{
		UsersService: usersService,
	}
}

func (c *commandsService) Create(input model.NewCommand) (*Command, error) {
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

func (c *commandsService) Update(changes *Command) error {
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

func (c *commandsService) GetAll() ([]Command, error) {
	filter := bson.D{{}}

	return c.GetFiltered(filter, nil)
}

func (c *commandsService) GetById(id string) (*Command, error) {
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

	commands, err := c.GetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(commands) == 0 {
		return nil, nil
	}

	return &commands[0], nil
}

func (c *commandsService) GetPaginated(startValue *string, first int, filter *model.CommandsFilter) ([]Command, error) {
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

	return c.GetFiltered(finalFilter, opts)
}

func (c *commandsService) GetFiltered(filter interface{}, opts *options.FindOptions) ([]Command, error) {
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

func (c *commandsService) GetTheoricalStatus(commandID string, commerceCommandsService CommerceCommandsService) (*string, error) {
	databaseCommerceCommands, err := commerceCommandsService.GetForCommand(commandID)

	if err != nil {
		return nil, err
	}

	result := COMMAND_STATUS_CANCELED
	isCanceled := false

	// On doit vérifier qu'elle n'est pas annuler en premier
	for _, databaseCommerceCommand := range databaseCommerceCommands {
		if databaseCommerceCommand.Status != COMMERCE_COMMAND_STATUS_CANCELED {
			break
		}

		isCanceled = true
	}

	if isCanceled {
		return &result, nil
	}

	// Sinon on fait le classique
	result = COMMAND_STATUS_DONE

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

func (c *commandsService) GetUser(commandID string) (*model.User, error) {
	databaseCommand, err := c.GetById(commandID)

	if err != nil {
		return nil, err
	}

	if databaseCommand == nil {
		return nil, &CommandNotFoundError{}
	}

	databaseUser, err := c.UsersService.GetUserById(databaseCommand.UserID.Hex())

	if err != nil {
		return nil, err
	}

	if databaseUser == nil {
		return nil, &users.UserNotFoundError{}
	}

	return databaseUser.ToModel(), nil
}
