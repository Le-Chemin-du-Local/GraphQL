package paniers

import (
	"bytes"
	"io/ioutil"
	"os"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/config"
	"chemin-du-local.bzh/graphql/internal/database"
	"chemin-du-local.bzh/graphql/internal/products"
	"github.com/99designs/gqlgen/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Panier struct {
	ID          primitive.ObjectID `bson:"_id"`
	CommerceID  primitive.ObjectID `bson:"commerceID"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	Category    string             `bson:"category"`
	Quantity    int                `bson:"quantity"`
	Price       int                `bson:"price"`
	Products    []PanierProduct    `bson:"products"`
}

type PanierProduct struct {
	ID        primitive.ObjectID `bson:"_id"`
	ProductID primitive.ObjectID `bson:"productID"`
	Quantity  int                `bson:"quantity"`
}

func (panier *Panier) ToModel() *model.Panier {
	return &model.Panier{
		ID:          panier.ID.Hex(),
		Name:        panier.Name,
		Description: panier.Description,
		Category:    panier.Category,
		Quantity:    panier.Quantity,
		Price:       panier.Price,
	}
}

func (panier Panier) IsLast() bool {
	filter := bson.D{
		primitive.E{
			Key:   "commerceID",
			Value: panier.CommerceID,
		},
	}

	opts := options.FindOptions{}
	opts.SetLimit(1)
	opts.SetSort(bson.D{
		primitive.E{
			Key: "_id", Value: -1,
		},
	})

	lastPanier, err := GetFiltered(filter, &opts)

	if err != nil || len(lastPanier) <= 0 {
		return false
	}

	return lastPanier[0].ID == panier.ID
}

// Créateur de base de données

func Create(commerceID primitive.ObjectID, input model.NewPanier) (*Panier, error) {
	products := []PanierProduct{}
	for _, product := range input.Products {
		productObjectID, err := primitive.ObjectIDFromHex(product.ProductID)

		if err != nil {
			return nil, err
		}

		products = append(products, PanierProduct{
			ID:        primitive.NewObjectID(),
			ProductID: productObjectID,
			Quantity:  product.Quantity,
		})
	}

	panierObjectId := primitive.NewObjectID()
	databasePanier := Panier{
		ID:          panierObjectId,
		CommerceID:  commerceID,
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Quantity:    input.Quantity,
		Price:       input.Price,
		Products:    products,
	}

	_, err := database.CollectionPaniers.InsertOne(database.MongoContext, databasePanier)

	if err != nil {
		return nil, err
	}

	if input.Image != nil {
		fileData := input.Image.File

		buffer := &bytes.Buffer{}
		buffer.ReadFrom(fileData)

		data := buffer.Bytes()

		folderPath := config.Cfg.Paths.Static + "/paniers"
		os.MkdirAll(folderPath, os.ModePerm)
		err := ioutil.WriteFile(folderPath+"/"+panierObjectId.Hex()+".jpg", data, 0644)

		if err != nil {
			return &databasePanier, err
		}
	}

	return &databasePanier, nil
}

// Mise à jour de la base de données

func Update(changes *Panier, image *graphql.Upload) error {
	filter := bson.D{
		primitive.E{
			Key:   "_id",
			Value: changes.ID,
		},
	}

	_, err := database.CollectionPaniers.ReplaceOne(database.MongoContext, filter, changes)

	if image != nil {
		fileData := image.File

		buffer := &bytes.Buffer{}
		buffer.ReadFrom(fileData)

		data := buffer.Bytes()

		folderPath := config.Cfg.Paths.Static + "/paniers"
		os.MkdirAll(folderPath, os.ModePerm)
		err := ioutil.WriteFile(folderPath+"/"+changes.ID.Hex()+".jpg", data, 0644)

		if err != nil {
			return err
		}
	}

	return err
}

// Getter de base de données

func GetById(id string) (*Panier, error) {
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

	paniers, err := GetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(paniers) == 0 {
		return nil, nil
	}

	return &paniers[0], nil
}

func GetForCommerce(commerceID string) ([]Panier, error) {
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

func GetPaginated(commerceID string, startValue *string, first int, filters *model.PanierFilter) ([]Panier, error) {
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
		if filters.Category != nil {
			finalFilter = append(finalFilter, primitive.E{
				Key:   "category",
				Value: filters.Category,
			})
		}
	}

	opts := options.Find()
	opts.SetLimit(int64(first))

	return GetFiltered(finalFilter, opts)
}

func GetFiltered(filter interface{}, opts *options.FindOptions) ([]Panier, error) {
	paniers := []Panier{}

	cursor, err := database.CollectionPaniers.Find(database.MongoContext, filter, opts)

	if err != nil {
		return paniers, err
	}

	for cursor.Next(database.MongoContext) {
		var pannier Panier

		err := cursor.Decode(&pannier)

		if err != nil {
			return paniers, err
		}

		paniers = append(paniers, pannier)
	}

	if err := cursor.Err(); err != nil {
		return paniers, err
	}

	return paniers, nil
}

// Getter pour les produits

func GetProducts(panierID string) ([]*model.PanierProduct, error) {
	databasePanier, err := GetById(panierID)

	if err != nil {
		return nil, err
	}

	if databasePanier == nil {
		return nil, &PanierNotFoundError{}
	}

	modelProducts := []*model.PanierProduct{}

	for _, product := range databasePanier.Products {
		databaseProduct, err := products.GetById(product.ProductID.Hex())

		if err != nil {
			return nil, err
		}

		if databaseProduct != nil {
			modelProducts = append(modelProducts, &model.PanierProduct{
				Quantity: product.Quantity,
				Product:  databaseProduct.ToModel(),
			})
		}
	}

	return modelProducts, nil
}