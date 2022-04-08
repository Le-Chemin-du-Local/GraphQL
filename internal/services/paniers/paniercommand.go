package paniers

import (
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PANIERCOMMAND_STATUS_IN_PROGRESS = "INPROGRESS"

type PanierCommand struct {
	ID         primitive.ObjectID `bson:"_id"`
	CommerceID primitive.ObjectID `bson:"commerceID"`
	UserID     primitive.ObjectID `bson:"userID"`
	PanierID   primitive.ObjectID `bson:"panierID"`
	Status     string             `bson:"status"`
	PickupDate time.Time          `bson:"pickupDate"`
}

func (panierCommand *PanierCommand) ToModel() *model.PanierCommand {
	return &model.PanierCommand{
		ID:         panierCommand.ID.Hex(),
		Status:     panierCommand.Status,
		PickupDate: panierCommand.PickupDate,
	}
}

func (panierCommand PanierCommand) IsLast() bool {
	filter := bson.D{
		primitive.E{
			Key:   "commerceID",
			Value: panierCommand.CommerceID,
		},
	}

	opts := options.FindOptions{}
	opts.SetLimit(1)
	opts.SetSort(bson.D{
		primitive.E{
			Key: "_id", Value: -1,
		},
	})

	lastPanierCommand, err := GetFilteredCommands(filter, &opts)

	if err != nil || len(lastPanierCommand) <= 0 {
		return false
	}

	return lastPanierCommand[0].ID == panierCommand.ID
}

// Créateur de base de données

func CreateCommand(userID primitive.ObjectID, commerceID string, input model.NewPanierCommand) (*PanierCommand, error) {
	commerceObjectID, err := primitive.ObjectIDFromHex(commerceID)

	if err != nil {
		return nil, err
	}

	panierObjectID, err := primitive.ObjectIDFromHex(input.PanierID)

	if err != nil {
		return nil, err
	}

	databasePanierCommand := PanierCommand{
		ID:         primitive.NewObjectID(),
		CommerceID: commerceObjectID,
		UserID:     userID,
		PanierID:   panierObjectID,
		Status:     PANIERCOMMAND_STATUS_IN_PROGRESS,
		PickupDate: input.PickupDate,
	}

	_, err = database.CollectionPanierCommands.InsertOne(database.MongoContext, databasePanierCommand)

	if err != nil {
		return nil, err
	}

	return &databasePanierCommand, nil
}

// Mise à jour de la base de données

func UpdateCommand(changes *PanierCommand) error {
	filter := bson.D{
		primitive.E{
			Key:   "_id",
			Value: changes.ID,
		},
	}

	_, err := database.CollectionPanierCommands.ReplaceOne(database.MongoContext, filter, changes)

	return err
}

// Getter de base de données

func GetCommandById(id string) (*PanierCommand, error) {
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

	panierCommands, err := GetFilteredCommands(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(panierCommands) == 0 {
		return nil, nil
	}

	return &panierCommands[0], nil
}

func GetCommandsForCommerce(commerceID string) ([]PanierCommand, error) {
	commerceObjectId, err := primitive.ObjectIDFromHex(commerceID)

	if err != nil {
		return nil, err
	}

	filter := bson.D{
		primitive.E{
			Key:   "commerceID",
			Value: commerceObjectId,
		},
	}

	return GetFilteredCommands(filter, nil)
}

func GetCommandsPaginated(commerceID string, startValue *string, first int, filters *model.PanierCommandFilter) ([]PanierCommand, error) {
	commerceObjectID, err := primitive.ObjectIDFromHex(commerceID)

	if err != nil {
		return nil, err
	}

	// On doit faire un filtre spécifique si on veut commencer
	// le curseur à l'ID de départ
	var finalFilter bson.D

	if startValue != nil {
		objectID, err := primitive.ObjectIDFromHex(*startValue)

		if err != nil {
			return nil, err
		}

		finalFilter = bson.D{
			primitive.E{
				Key:   "commerceID",
				Value: commerceObjectID,
			},
			primitive.E{
				Key: "_id",
				Value: bson.M{
					"$gt": objectID,
				},
			},
		}
	} else {
		finalFilter = bson.D{
			primitive.E{
				Key:   "commerceID",
				Value: commerceObjectID,
			},
		}
	}

	if filters != nil {
		if filters.Status != nil {
			finalFilter = append(finalFilter, primitive.E{
				Key:   "status",
				Value: filters.Status,
			})
		}
	}

	opts := options.Find()
	opts.SetLimit(int64(first))

	return GetFilteredCommands(finalFilter, opts)
}

func GetFilteredCommands(filter interface{}, opts *options.FindOptions) ([]PanierCommand, error) {
	panierCommands := []PanierCommand{}

	cursor, err := database.CollectionPanierCommands.Find(database.MongoContext, filter, opts)

	if err != nil {
		return panierCommands, err
	}

	for cursor.Next(database.MongoContext) {
		var panierCommand PanierCommand

		err := cursor.Decode(&panierCommand)

		if err != nil {
			return panierCommands, err
		}

		panierCommands = append(panierCommands, panierCommand)
	}

	if err := cursor.Err(); err != nil {
		return panierCommands, err
	}

	return panierCommands, nil
}
