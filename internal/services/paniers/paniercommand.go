package paniers

import (
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PanierCommand struct {
	ID                primitive.ObjectID `bson:"_id"`
	CommerceCommandID primitive.ObjectID `bson:"commerceCommandID"`
	PanierID          primitive.ObjectID `bson:"panierID"`
}

func (panierCommand *PanierCommand) ToModel() *model.PanierCommand {
	return &model.PanierCommand{
		ID: panierCommand.ID.Hex(),
	}
}

// Créateur de base de données

func CreateCommand(commerceCommandID primitive.ObjectID, input model.NewPanierCommand) (*PanierCommand, error) {
	panierObjectID, err := primitive.ObjectIDFromHex(input.PanierID)

	if err != nil {
		return nil, err
	}

	databasePanierCommand := PanierCommand{
		ID:                primitive.NewObjectID(),
		CommerceCommandID: commerceCommandID,
		PanierID:          panierObjectID,
	}

	_, err = database.CollectionPanierCommands.InsertOne(database.MongoContext, databasePanierCommand)

	if err != nil {
		return nil, err
	}

	return &databasePanierCommand, nil
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

func GetCommandsForCommerceCommand(commerceCommandID string) ([]PanierCommand, error) {
	commerceCommandObjectID, err := primitive.ObjectIDFromHex(commerceCommandID)

	if err != nil {
		return nil, err
	}

	filter := bson.D{
		primitive.E{
			Key:   "commerceCommandID",
			Value: commerceCommandObjectID,
		},
	}

	return GetFilteredCommands(filter, nil)
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
