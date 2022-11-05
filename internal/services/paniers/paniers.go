package paniers

import (
	"bytes"
	"io/ioutil"
	"os"
	"time"

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
	Type        string             `bson:"type"`
	Category    string             `bson:"category"`
	Quantity    int                `bson:"quantity"`
	Price       float64            `bson:"price"`
	Reduction   float64            `bons:"reduction"`
	EndingDate  *time.Time         `bson:"endingDate"`
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
		Type:        panier.Type,
		Category:    panier.Category,
		Quantity:    panier.Quantity,
		EndingDate:  panier.EndingDate,
		Price:       panier.Price,
		Reduction:   panier.Reduction,
	}
}

func (panier Panier) IsLast(paniersService PaniersService) bool {
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

	lastPanier, err := paniersService.GetFiltered(filter, &opts)

	if err != nil || len(lastPanier) <= 0 {
		return false
	}

	return lastPanier[0].ID == panier.ID
}

// Service

type paniersService struct {
	ProductsService products.ProductsService
}

type PaniersService interface {
	Create(commerceID primitive.ObjectID, input model.NewPanier) (*Panier, error)
	Update(changes *Panier, image *graphql.Upload) error
	GetById(id string) (*Panier, error)
	GetPaginated(commerceID string, startValue *string, first int, filters *model.PanierFilter) ([]Panier, error)
	GetFiltered(filter interface{}, opts *options.FindOptions) ([]Panier, error)
	GetProducts(panierID string) ([]*model.PanierProduct, error)
}

func NewPaniersService(productsService products.ProductsService) *paniersService {
	return &paniersService{
		ProductsService: productsService,
	}
}

// Créateur de base de données

func (p *paniersService) Create(commerceID primitive.ObjectID, input model.NewPanier) (*Panier, error) {
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
		Type:        input.Type,
		Category:    input.Category,
		Quantity:    input.Quantity,
		Price:       input.Price,
		Reduction:   input.Reduction,
		EndingDate:  input.EndingDate,
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

func (p *paniersService) Update(changes *Panier, image *graphql.Upload) error {
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

func (p *paniersService) GetById(id string) (*Panier, error) {
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

	paniers, err := p.GetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(paniers) == 0 {
		return nil, nil
	}

	return &paniers[0], nil
}

func (p *paniersService) GetPaginated(commerceID string, startValue *string, first int, filters *model.PanierFilter) ([]Panier, error) {
	commerceObjectID, err := primitive.ObjectIDFromHex(commerceID)

	if err != nil {
		return nil, err
	}

	// On doit faire un filtre spécifique si on veut commencer
	// le curseur à l'ID de départ
	var finalFilter bson.M

	if startValue != nil {
		objectID, err := primitive.ObjectIDFromHex(*startValue)

		if err != nil {
			return nil, err
		}

		finalFilter = bson.M{
			"commerceID": commerceObjectID,
			"_id": bson.M{
				"$gt": objectID,
			},
			"$or": []bson.M{
				{
					"endingDate": bson.M{
						"$gte": time.Now(),
					},
				},
				{
					"type": "PERMANENT",
				},
			},
		}
	} else {
		finalFilter = bson.M{
			"commerceID": commerceObjectID,
			"$or": []bson.M{
				{
					"endingDate": bson.M{
						"$gte": time.Now(),
					},
				},
				{
					"type": "PERMANENT",
				},
			},
		}
	}

	if filters != nil {
		if filters.Type != nil {
			finalFilter = bson.M{
				"$and": []bson.M{
					finalFilter,
					{
						"type": filters.Type,
					},
				},
			}
		}
	}

	opts := options.Find()
	opts.SetLimit(int64(first))

	return p.GetFiltered(finalFilter, opts)
}

func (p *paniersService) GetFiltered(filter interface{}, opts *options.FindOptions) ([]Panier, error) {
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

func (p *paniersService) GetProducts(panierID string) ([]*model.PanierProduct, error) {
	databasePanier, err := p.GetById(panierID)

	if err != nil {
		return nil, err
	}

	if databasePanier == nil {
		return nil, &PanierNotFoundError{}
	}

	modelProducts := []*model.PanierProduct{}

	for _, product := range databasePanier.Products {
		databaseProduct, err := p.ProductsService.GetById(product.ProductID.Hex())

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
