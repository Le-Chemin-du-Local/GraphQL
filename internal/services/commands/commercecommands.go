package commands

import (
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/database"
	"chemin-du-local.bzh/graphql/internal/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const COMMERCE_COMMAND_STATUS_IN_PROGRESS = "INPROGRESS"
const COMMERCE_COMMAND_STATUS_READY = "READY"
const COMMERCE_COMMAND_STATUS_DONE = "DONE"

type CommerceCommand struct {
	ID         primitive.ObjectID `bson:"_id"`
	CommandID  primitive.ObjectID `bson:"commandID"`
	CommerceID primitive.ObjectID `bson:"commerceID"`
	PickupDate time.Time          `bson:"pickupDate"`
	Status     string             `bson:"status"`
}

func (command *CommerceCommand) ToModel() *model.CommerceCommand {
	return &model.CommerceCommand{
		ID:         command.ID.Hex(),
		PickupDate: command.PickupDate,
		Status:     command.Status,
	}
}

func (commerceCommand CommerceCommand) IsLast() bool {
	filter := bson.D{{}}

	opts := options.FindOptions{}
	opts.SetLimit(1)
	opts.SetSort(bson.D{
		primitive.E{
			Key: "_id", Value: -1,
		},
	})

	lastCommerceCommand, err := GetFiltered(filter, &opts)

	if err != nil || len(lastCommerceCommand) <= 0 {
		return false
	}

	return lastCommerceCommand[0].ID == commerceCommand.ID
}

func CommerceCreate(input model.NewCommerceCommand, commandID primitive.ObjectID) (*CommerceCommand, error) {
	commerceObjectID, err := primitive.ObjectIDFromHex(input.CommerceID)

	if err != nil {
		return nil, err
	}

	databaseCommerceCommand := CommerceCommand{
		ID:         primitive.NewObjectID(),
		CommandID:  commandID,
		CommerceID: commerceObjectID,
		PickupDate: input.PickupDate,
		Status:     COMMERCE_COMMAND_STATUS_IN_PROGRESS,
	}

	_, err = database.CollectionCommerceCommand.InsertOne(database.MongoContext, databaseCommerceCommand)

	if err != nil {
		return nil, err
	}

	return &databaseCommerceCommand, nil
}

// Mise à jour de la base de données

func CommerceUpdate(changes *CommerceCommand) error {
	databaseCommand, err := GetById(changes.CommandID.Hex())

	if err != nil {
		return err
	}

	if databaseCommand == nil {
		return &CommandNotFoundError{}
	}

	filter := bson.D{
		primitive.E{
			Key:   "_id",
			Value: changes.ID,
		},
	}

	_, err = database.CollectionCommerceCommand.ReplaceOne(database.MongoContext, filter, changes)

	if err != nil {
		return err
	}

	commandStatus, err := GetStatus(databaseCommand.ID.Hex())

	if err != nil {
		return err
	}

	databaseCommand.Status = *commandStatus

	err = Update(databaseCommand)

	return err
}

// Getters

func CommerceGetAll() ([]CommerceCommand, error) {
	filter := bson.D{{}}

	return CommerceGetFiltered(filter, nil)
}

func CommerceGetById(id string) (*CommerceCommand, error) {
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

	commerceCommands, err := CommerceGetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(commerceCommands) == 0 {
		return nil, nil
	}

	return &commerceCommands[0], nil
}

func CommerceGetForCommand(commandID string) ([]CommerceCommand, error) {
	commandObjectId, err := primitive.ObjectIDFromHex(commandID)

	if err != nil {
		return nil, err
	}

	filter := bson.D{
		primitive.E{
			Key:   "commandID",
			Value: commandObjectId,
		},
	}

	commerceCommands, err := CommerceGetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	return commerceCommands, nil
}

func CommerceGetPaginated(startValue *string, first int, commerceID *string) ([]CommerceCommand, error) {
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

	if commerceID != nil {
		commerceObjectID, err := primitive.ObjectIDFromHex(*commerceID)

		if err != nil {
			return nil, err
		}

		finalFilter = bson.M{
			"$and": []bson.M{
				finalFilter,
				{
					"commerceID": bson.M{
						"$eq": commerceObjectID,
					},
				},
			},
		}
	}

	opts := options.Find()
	opts.SetLimit(int64(first))

	return CommerceGetFiltered(finalFilter, opts)
}

func CommerceGetFiltered(filter interface{}, opts *options.FindOptions) ([]CommerceCommand, error) {
	commerceCommands := []CommerceCommand{}

	cursor, err := database.CollectionCommerceCommand.Find(database.MongoContext, filter, opts)

	if err != nil {
		return commerceCommands, err
	}

	for cursor.Next(database.MongoContext) {
		var commerceCommand CommerceCommand

		err := cursor.Decode(&commerceCommand)

		if err != nil {
			return commerceCommands, err
		}

		commerceCommands = append(commerceCommands, commerceCommand)
	}

	if err := cursor.Err(); err != nil {
		return commerceCommands, err
	}

	return commerceCommands, nil
}

func CommerceGetCommerce(commerceCommandID string) (*model.Commerce, error) {
	databaseCommerceCommand, err := CommerceGetById(commerceCommandID)

	if err != nil {
		return nil, err
	}

	if databaseCommerceCommand == nil {
		return nil, &CommerceCommandNotFoundError{}
	}

	databaseCommerce, err := commerces.GetById(databaseCommerceCommand.CommerceID.Hex())

	if err != nil {
		return nil, err
	}

	if databaseCommerce == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	return databaseCommerce.ToModel(), nil
}

func CommerceGetUser(commerceCommandID string) (*model.User, error) {
	databaseCommerceCommand, err := CommerceGetById(commerceCommandID)

	if err != nil {
		return nil, err
	}

	if databaseCommerceCommand == nil {
		return nil, &CommerceCommandNotFoundError{}
	}

	databaseCommand, err := GetById(databaseCommerceCommand.CommandID.Hex())

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
