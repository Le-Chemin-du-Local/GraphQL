package clickandcollect

import (
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/database"
	"chemin-du-local.bzh/graphql/internal/products"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CCCOMMAND_STATUS_IN_PRGRESS = "INPROGRESS"

type CCCommand struct {
	ID         primitive.ObjectID `bson:"_id"`
	CommerceID primitive.ObjectID `bson:"commerceID"`
	UserID     primitive.ObjectID `bson:"userID"`
	Status     string             `bson:"status"`
	PickupDate time.Time          `bson:"pickupDate"`
	Products   []CCProduct        `bson:"products"`
}

type CCProduct struct {
	ID        primitive.ObjectID `bson:"_id"`
	ProductID primitive.ObjectID `bson:"productID"`
	Quantity  int                `bson:"quantity"`
}

func (cccommand *CCCommand) ToModel() *model.CCCommand {
	return &model.CCCommand{
		ID:         cccommand.ID.Hex(),
		Status:     cccommand.Status,
		PickupDate: cccommand.PickupDate,
	}
}

func (cccommand CCCommand) IsLast() bool {
	filter := bson.D{
		primitive.E{
			Key:   "commerceID",
			Value: cccommand.CommerceID,
		},
	}

	opts := options.FindOptions{}
	opts.SetLimit(1)
	opts.SetSort(bson.D{
		primitive.E{
			Key: "_id", Value: -1,
		},
	})

	lastCCCommand, err := GetFiltered(filter, &opts)

	if err != nil || len(lastCCCommand) <= 0 {
		return false
	}

	return lastCCCommand[0].ID == cccommand.ID
}

// Créateur de base de données

func Create(userID primitive.ObjectID, commerceID string, input model.NewCCCommand) (*CCCommand, error) {
	commerceObjectID, err := primitive.ObjectIDFromHex(commerceID)

	if err != nil {
		return nil, err
	}

	products := []CCProduct{}
	for _, product := range input.ProductsID {
		productObjectID, err := primitive.ObjectIDFromHex(product.ProductID)

		if err != nil {
			return nil, err
		}

		products = append(products, CCProduct{
			ID:        primitive.NewObjectID(),
			ProductID: productObjectID,
			Quantity:  product.Quantity,
		})
	}

	databaseCCCommand := CCCommand{
		ID:         primitive.NewObjectID(),
		CommerceID: commerceObjectID,
		UserID:     userID,
		Status:     CCCOMMAND_STATUS_IN_PRGRESS,
		PickupDate: input.PickupDate,
		Products:   products,
	}

	_, err = database.CollectionCCCommand.InsertOne(database.MongoContext, databaseCCCommand)

	if err != nil {
		return nil, err
	}

	return &databaseCCCommand, nil
}

// Mise à jour de la base de données

func Update(changes *CCCommand) error {
	filter := bson.D{
		primitive.E{
			Key:   "_id",
			Value: changes.ID,
		},
	}

	_, err := database.CollectionCCCommand.ReplaceOne(database.MongoContext, filter, changes)

	return err
}

// Getter de base de données

func GetById(id string) (*CCCommand, error) {
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

	cccommands, err := GetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(cccommands) == 0 {
		return nil, nil
	}

	return &cccommands[0], nil
}

func GetForCommerce(commerceID string) ([]CCCommand, error) {
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

	return GetFiltered(filter, nil)
}

func GetPaginated(commerceID string, startValue *string, first int, filters *model.CCCommandFilter) ([]CCCommand, error) {
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

	return GetFiltered(finalFilter, opts)
}

func GetFiltered(filter interface{}, opts *options.FindOptions) ([]CCCommand, error) {
	cccommands := []CCCommand{}

	cursor, err := database.CollectionCCCommand.Find(database.MongoContext, filter, opts)

	if err != nil {
		return cccommands, err
	}

	for cursor.Next(database.MongoContext) {
		var cccommand CCCommand

		err := cursor.Decode(&cccommand)

		if err != nil {
			return cccommands, err
		}

		cccommands = append(cccommands, cccommand)
	}

	if err := cursor.Err(); err != nil {
		return cccommands, err
	}

	return cccommands, nil
}

// Getter pour les produits

func GetProducts(cccommandID string) ([]*model.CCProduct, error) {
	databaseCCCommand, err := GetById(cccommandID)

	if err != nil {
		return nil, err
	}

	if databaseCCCommand == nil {
		return nil, &CCCommandNotFoundError{}
	}

	modelProducts := []*model.CCProduct{}

	for _, product := range databaseCCCommand.Products {
		databaseProduct, err := products.GetById(product.ProductID.Hex())

		if err != nil {
			return nil, err
		}

		if databaseProduct != nil {
			modelProducts = append(modelProducts, &model.CCProduct{
				Quantity: product.Quantity,
				Product:  databaseProduct.ToModel(),
			})
		}
	}

	return modelProducts, nil
}
