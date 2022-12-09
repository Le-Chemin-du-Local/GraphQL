package products

import (
	"bytes"
	"io/ioutil"
	"os"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/config"
	"chemin-du-local.bzh/graphql/internal/database"
	"github.com/99designs/gqlgen/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Product struct {
	ID                  primitive.ObjectID `bson:"_id"`
	CommerceID          primitive.ObjectID `bson:"commerceID"`
	Name                string             `bson:"name"`
	Description         string             `bson:"description"`
	Price               float64            `bson:"price"`
	Unit                string             `bson:"unit"`
	PerUnitQuantity     float64            `bson:"perUnitQuantity"`
	PerUnitQuantityUnit string             `bson:"perUnitQuantityUnit"`
	Tva                 float64            `bson:"tva"`
	IsBreton            bool               `bson:"isBreton"`
	Tags                []string           `bson:"tags"`
	HasGluten           bool               `bson:"hasGlutted"`
	Allergens           []string           `bson:"allergens"`
	Categories          []string           `bson:"categories"`
}

func (product *Product) ToModel() *model.Product {
	return &model.Product{
		ID:                  product.ID.Hex(),
		Name:                product.Name,
		Description:         product.Description,
		Price:               product.Price,
		Unit:                product.Unit,
		PerUnitQuantity:     product.PerUnitQuantity,
		PerUnitQuantityUnit: product.PerUnitQuantityUnit,
		Tva:                 product.Tva,
		IsBreton:            product.IsBreton,
		HasGluten:           product.HasGluten,
		Tags:                product.Tags,
		Allergens:           product.Allergens,
		Categories:          product.Categories,
	}
}

func (product Product) IsLast(productsService ProductsService) bool {
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

	lastProduct, err := productsService.GetFiltered(filter, &opts)

	if err != nil || len(lastProduct) <= 0 {
		return false
	}

	return lastProduct[0].ID == product.ID
}

// Service

type productsService struct{}

type ProductsService interface {
	Create(commerceID string, input model.NewProduct) (*Product, error)
	Update(changes *Product, image *graphql.Upload) error
	GetById(id string) (*Product, error)
	GetForCommerce(commerceID string) ([]Product, error)
	GetPaginated(commerceID string, startValue *string, first int, filters *model.ProductFilter) ([]Product, error)
	GetFiltered(filter interface{}, opts *options.FindOptions) ([]Product, error)
}

func NewProductsService() *productsService {
	return &productsService{}
}

// Créateur de base de données

func (p *productsService) Create(commerceID string, input model.NewProduct) (*Product, error) {
	commerceObjectID, err := primitive.ObjectIDFromHex(commerceID)

	if err != nil {
		return nil, err
	}

	productObjectID := primitive.NewObjectID()
	databaseProduct := Product{
		ID:                  productObjectID,
		CommerceID:          commerceObjectID,
		Name:                input.Name,
		Description:         input.Description,
		Price:               input.Price,
		Unit:                input.Unit,
		PerUnitQuantity:     input.PerUnitQuantity,
		PerUnitQuantityUnit: input.PerUnitQuantityUnit,
		Tva:                 input.Tva,
		IsBreton:            input.IsBreton,
		HasGluten:           input.HasGluten,
		Tags:                input.Tags,
		Allergens:           input.Allergens,
		Categories:          input.Categories,
	}

	_, err = database.CollectionProducts.InsertOne(database.MongoContext, databaseProduct)

	if err != nil {
		return nil, err
	}

	if input.Image != nil {
		fileData := input.Image.File

		buffer := &bytes.Buffer{}
		buffer.ReadFrom(fileData)

		data := buffer.Bytes()

		folderPath := config.Cfg.Paths.Static + "/products"
		os.MkdirAll(folderPath, os.ModePerm)
		err := ioutil.WriteFile(folderPath+"/"+productObjectID.Hex()+".jpg", data, 0644)

		if err != nil {
			return &databaseProduct, err
		}
	}

	return &databaseProduct, nil
}

// Mise à jour de la base de données

func (p *productsService) Update(changes *Product, image *graphql.Upload) error {
	filter := bson.D{
		primitive.E{
			Key:   "_id",
			Value: changes.ID,
		},
	}

	_, err := database.CollectionProducts.ReplaceOne(database.MongoContext, filter, changes)

	if image != nil {
		fileData := image.File

		buffer := &bytes.Buffer{}
		buffer.ReadFrom(fileData)

		data := buffer.Bytes()

		folderPath := config.Cfg.Paths.Static + "/products"
		os.MkdirAll(folderPath, os.ModePerm)
		err := ioutil.WriteFile(folderPath+"/"+changes.ID.Hex()+".jpg", data, 0644)

		if err != nil {
			return err
		}
	}

	return err
}

// Getter de base de données

func (p *productsService) GetById(id string) (*Product, error) {
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

	products, err := p.GetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, nil
	}

	return &products[0], nil
}

func (p *productsService) GetForCommerce(commerceID string) ([]Product, error) {
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

	return p.GetFiltered(filter, nil)
}

func (p *productsService) GetPaginated(commerceID string, startValue *string, first int, filters *model.ProductFilter) ([]Product, error) {
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

		// finalFilter = bson.D{
		// 	"commerceID": ,
		// 	"_id": bson.M{
		// 		"$gt": objectID,
		// 	},
		// }
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
				Key:   "categories",
				Value: filters.Category,
			})
		}
	}

	opts := options.Find()
	opts.SetLimit(int64(first))

	return p.GetFiltered(finalFilter, opts)
}

func (p *productsService) GetFiltered(filter interface{}, opts *options.FindOptions) ([]Product, error) {
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
