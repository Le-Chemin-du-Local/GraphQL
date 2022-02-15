package products

import (
	"log"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Product struct {
	ID          primitive.ObjectID `bson:"_id"`
	CommerceID  primitive.ObjectID `bson:"commerceID"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	Price       float64            `bson:"price"`
	Unit        string             `bson:"unit"`
	IsBreton    bool               `bson:"isBreton"`
	Tags        []string           `bson:"tags"`
	Categories  []string           `bson:"categories"`
}

func (product *Product) ToModel() *model.Product {
	return &model.Product{
		ID:          product.ID.Hex(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Unit:        product.Unit,
		IsBreton:    product.IsBreton,
		Tags:        product.Tags,
		Categories:  product.Categories,
	}
}

func (product Product) IsLast() bool {
	filter := bson.D{
		primitive.E{
			Key:   "commerceID",
			Value: product.CommerceID,
		},
	}

	opts := options.FindOptions{}
	opts.SetLimit(1)
	opts.SetSort(bson.D{
		primitive.E{
			Key: "_id", Value: -1,
		},
	})

	lastProduct, err := GetFiltered(filter, &opts)

	if err != nil || len(lastProduct) <= 0 {
		return false
	}

	return lastProduct[0].ID == product.ID
}

// Créateur de base de données

func Create(input model.NewProduct) *Product {
	if input.CommerceID == nil {
		log.Fatal("aucun commerce n'a été fourni")
	}

	commerceObjectID, err := primitive.ObjectIDFromHex(*input.CommerceID)

	if err != nil {
		log.Fatal(err)
	}

	databaseProduct := Product{
		ID:          primitive.NewObjectID(),
		CommerceID:  commerceObjectID,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Unit:        input.Unit,
		IsBreton:    input.IsBreton,
		Tags:        input.Tags,
		Categories:  input.Categories,
	}

	_, err = database.CollectionProducts.InsertOne(database.MongoContext, databaseProduct)

	if err != nil {
		log.Fatal(err)
	}

	return &databaseProduct
}

// Getter de base de données

func GetById(id string) (*Product, error) {
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

	products, err := GetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, nil
	}

	return &products[0], nil
}

func GetForCommerce(commerceID string) ([]Product, error) {
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

func GetPaginated(commerceID string, startValue *string, first int) ([]Product, error) {
	commerceObjectID, err := primitive.ObjectIDFromHex(commerceID)

	if err != nil {
		return nil, err
	}

	// On doit faire un filtre spécifique si on veut commencer
	// le curseur à l'ID de départ
	var filter interface{}

	if startValue != nil {
		objectID, err := primitive.ObjectIDFromHex(*startValue)

		if err != nil {
			return nil, err
		}

		filter = bson.M{
			"commerceID": commerceObjectID,
			"_id": bson.M{
				"$gt": objectID,
			},
		}
	} else {
		filter = bson.D{
			primitive.E{
				Key:   "commerceID",
				Value: commerceObjectID,
			},
		}
	}

	opts := options.Find()
	opts.SetLimit(int64(first))

	return GetFiltered(filter, opts)
}

func GetFiltered(filter interface{}, opts *options.FindOptions) ([]Product, error) {
	products := []Product{}

	cursor, err := database.CollectionProducts.Find(database.MongoContext, filter, opts)

	if err != nil {
		return products, err
	}

	for cursor.Next(database.MongoContext) {
		var product Product

		err := cursor.Decode(&product)

		if err != nil {
			return products, err
		}

		products = append(products, product)
	}

	if err := cursor.Err(); err != nil {
		return products, err
	}

	return products, nil
}
