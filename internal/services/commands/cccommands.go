package commands

import (
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/database"
	"chemin-du-local.bzh/graphql/internal/products"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CCCommand struct {
	ID                primitive.ObjectID `bson:"_id"`
	CommerceCommandID primitive.ObjectID `bson:"commerceCommandID"`
	Products          []CCProduct        `bson:"products"`
}

type CCProduct struct {
	ID        primitive.ObjectID `bson:"_id"`
	ProductID primitive.ObjectID `bson:"productID"`
	Quantity  int                `bson:"quantity"`
}

func (cccommand *CCCommand) ToModel() *model.CCCommand {
	return &model.CCCommand{
		ID: cccommand.ID.Hex(),
	}
}

// Service

type ccCommandsService struct {
	ProductsSerivce products.ProductsService
}

type CCCommandsService interface {
	Create(commerceCommandID primitive.ObjectID, input model.NewCCCommand) (*CCCommand, error)
	GetById(id string) (*CCCommand, error)
	GetForCommmerceCommand(commerceCommandID string) ([]CCCommand, error)
	GetFiltered(filter interface{}, opts *options.FindOptions) ([]CCCommand, error)
	GetProducts(cccommandID string) ([]*model.CCProduct, error)
}

func NewCCCommandsService(productsService products.ProductsService) *ccCommandsService {
	return &ccCommandsService{
		ProductsSerivce: productsService,
	}
}

// Créateur de base de données

func (c *ccCommandsService) Create(commerceCommandID primitive.ObjectID, input model.NewCCCommand) (*CCCommand, error) {
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
		ID:                primitive.NewObjectID(),
		CommerceCommandID: commerceCommandID,
		Products:          products,
	}

	_, err := database.CollectionCCCommand.InsertOne(database.MongoContext, databaseCCCommand)

	if err != nil {
		return nil, err
	}

	return &databaseCCCommand, nil
}

// Getter de base de données

func (c *ccCommandsService) GetById(id string) (*CCCommand, error) {
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

	cccommands, err := c.GetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(cccommands) == 0 {
		return nil, nil
	}

	return &cccommands[0], nil
}

func (c *ccCommandsService) GetForCommmerceCommand(commerceCommandID string) ([]CCCommand, error) {
	commerceCommandObjectId, err := primitive.ObjectIDFromHex(commerceCommandID)

	if err != nil {
		return nil, err
	}

	filter := bson.D{
		primitive.E{
			Key:   "commerceCommandID",
			Value: commerceCommandObjectId,
		},
	}

	return c.GetFiltered(filter, nil)
}

func (c *ccCommandsService) GetFiltered(filter interface{}, opts *options.FindOptions) ([]CCCommand, error) {
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

func (c *ccCommandsService) GetProducts(cccommandID string) ([]*model.CCProduct, error) {
	databaseCCCommand, err := c.GetById(cccommandID)

	if err != nil {
		return nil, err
	}

	if databaseCCCommand == nil {
		return nil, &CCCommandNotFoundError{}
	}

	modelProducts := []*model.CCProduct{}

	for _, product := range databaseCCCommand.Products {
		databaseProduct, err := c.ProductsSerivce.GetById(product.ProductID.Hex())

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
