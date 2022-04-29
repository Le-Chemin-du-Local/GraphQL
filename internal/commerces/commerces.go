package commerces

import (
	"bytes"
	"io/ioutil"
	"os"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/config"
	"chemin-du-local.bzh/graphql/internal/database"
	"chemin-du-local.bzh/graphql/pkg/geojson"
	"github.com/99designs/gqlgen/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Commerce struct {
	ID                                  primitive.ObjectID  `bson:"_id"`
	StorekeeperID                       primitive.ObjectID  `bson:"storekeeperID"`
	Name                                string              `bson:"name"`
	Description                         string              `bson:"description"`
	StorekeeperWord                     string              `bson:"storekeeperWord"`
	Address                             string              `bson:"address"`
	AddressGeo                          geojson.GeoJSON     `bson:"addressGeo"`
	Phone                               string              `bson:"phone"`
	Email                               string              `bson:"email"`
	Facebook                            *string             `bson:"facebook"`
	Twitter                             *string             `bson:"twitter"`
	Instagram                           *string             `bson:"instagram"`
	BusinessHours                       model.BusinessHours `bson:"businesHours"`
	Services                            []string            `bson:"services"`
	ProductsAvailableForClickAndCollect []string            `bson:"productsAvailableForClickAndCollect"`
}

func (commerce *Commerce) ToModel() *model.Commerce {
	return &model.Commerce{
		ID:              commerce.ID.Hex(),
		StorekeeperID:   commerce.StorekeeperID.Hex(),
		Name:            commerce.Name,
		Description:     commerce.Description,
		StorekeeperWord: commerce.StorekeeperWord,
		Address:         commerce.Address,
		Latitude:        commerce.AddressGeo.Coordinates[1],
		Longitude:       commerce.AddressGeo.Coordinates[0],
		Phone:           commerce.Phone,
		Email:           commerce.Email,
		Facebook:        commerce.Facebook,
		Twitter:         commerce.Twitter,
		Instagram:       commerce.Instagram,
		BusinessHours:   commerce.BusinessHours,
		Services:        commerce.Services,
	}
}

func (commerce Commerce) IsLast() bool {
	filter := bson.D{{}}

	opts := options.FindOptions{}
	opts.SetLimit(1)
	opts.SetSort(bson.D{
		primitive.E{
			Key: "_id", Value: -1,
		},
	})

	lastCommerce, err := GetFiltered(filter, &opts)

	if err != nil || len(lastCommerce) <= 0 {
		return false
	}

	return lastCommerce[0].ID == commerce.ID
}

// Créateur de base de données

func Create(input model.NewCommerce, storekeeperID primitive.ObjectID) (*Commerce, error) {
	commerceObjectID := primitive.NewObjectID()

	description := ""
	storekeeperWord := ""

	if input.Description != nil {
		description = *input.Description
	}

	if input.StorekeeperWord != nil {
		storekeeperWord = *input.StorekeeperWord
	}

	var businessHours model.BusinessHours

	if input.BusinessHours != nil {
		businessHours = *input.BusinessHours.ToModel()
	}

	databaseCommerce := Commerce{
		ID:              commerceObjectID,
		StorekeeperID:   storekeeperID,
		Name:            input.Name,
		Description:     description,
		StorekeeperWord: storekeeperWord,
		Address:         input.Address,
		AddressGeo: geojson.GeoJSON{
			Type:        "Point",
			Coordinates: []float64{input.Latitude, input.Longitude},
		},
		Phone:         input.Phone,
		Email:         input.Email,
		Facebook:      input.Facebook,
		Twitter:       input.Twitter,
		Instagram:     input.Instagram,
		BusinessHours: businessHours,
	}

	_, err := database.CollectionCommerces.InsertOne(database.MongoContext, databaseCommerce)

	if err != nil {
		return nil, err
	}

	// Le header
	if input.Image != nil {
		fileData := input.Image.File

		buffer := &bytes.Buffer{}
		buffer.ReadFrom(fileData)

		data := buffer.Bytes()

		folderPath := config.Cfg.Paths.Static + "/commerces/" + commerceObjectID.Hex()
		os.MkdirAll(folderPath, os.ModePerm)
		err := ioutil.WriteFile(folderPath+"/header.jpg", data, 0644)

		if err != nil {
			return &databaseCommerce, err
		}
	}

	// La photo de profil
	if input.ProfilePicture != nil {
		fileData := input.Image.File

		buffer := &bytes.Buffer{}
		buffer.ReadFrom(fileData)

		data := buffer.Bytes()

		folderPath := config.Cfg.Paths.Static + "/commerces/" + commerceObjectID.Hex()
		os.MkdirAll(folderPath, os.ModePerm)
		err := ioutil.WriteFile(folderPath+"/profile.jpg", data, 0644)

		if err != nil {
			return &databaseCommerce, err
		}
	}

	return &databaseCommerce, nil
}

// Mise à jour en base de données

func Update(
	changes *Commerce, image *graphql.Upload, profilePicture *graphql.Upload) error {
	filter := bson.D{
		primitive.E{
			Key:   "_id",
			Value: changes.ID,
		},
	}

	_, err := database.CollectionCommerces.ReplaceOne(database.MongoContext, filter, changes)

	// Le header
	if image != nil {
		fileData := image.File

		buffer := &bytes.Buffer{}
		buffer.ReadFrom(fileData)

		data := buffer.Bytes()

		folderPath := config.Cfg.Paths.Static + "/commerces/" + changes.ID.Hex()
		os.MkdirAll(folderPath, os.ModePerm)
		err := ioutil.WriteFile(folderPath+"/header.jpg", data, 0644)

		if err != nil {
			return err
		}
	}

	// La photo de profil
	if profilePicture != nil {
		fileData := profilePicture.File

		buffer := &bytes.Buffer{}
		buffer.ReadFrom(fileData)

		data := buffer.Bytes()

		folderPath := config.Cfg.Paths.Static + "/commerces/" + changes.ID.Hex()
		os.MkdirAll(folderPath, os.ModePerm)
		err := ioutil.WriteFile(folderPath+"/profile.jpg", data, 0644)

		if err != nil {
			return err
		}
	}

	return err
}

// Getter de base de données

func GetAll() ([]Commerce, error) {
	filter := bson.D{{}}

	return GetFiltered(filter, nil)
}

func GetById(id string) (*Commerce, error) {
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

	commerces, err := GetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(commerces) == 0 {
		return nil, nil
	}

	return &commerces[0], nil
}

func GetForUser(userID string) (*Commerce, error) {
	userObjectId, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		return nil, err
	}

	filter := bson.D{
		primitive.E{
			Key:   "storekeeperID",
			Value: userObjectId,
		},
	}

	commerces, err := GetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(commerces) == 0 {
		return nil, nil
	}

	return &commerces[0], nil
}

func GetPaginated(startValue *string, first int, filter *model.CommerceFilter) ([]Commerce, error) {
	// On doit faire un filtre spécifique si on veut commencer
	// le curseur à l'ID de départ
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

	if filter != nil && filter.NearLatitude != nil && filter.NearLongitude != nil {
		maxDistance := 20000.0

		if filter.Radius != nil {
			maxDistance = *filter.Radius
		}

		finalFilter = bson.M{
			"$and": []bson.M{
				finalFilter,
				{
					"addressGeo": bson.M{
						"$near": bson.M{
							"$geometry": bson.M{
								"type": "Point",
								"coordinates": []float64{
									*filter.NearLongitude,
									*filter.NearLatitude,
								},
							},
							"$maxDistance": maxDistance,
						},
					},
				},
			},
		}
	}

	opts := options.Find()
	opts.SetLimit(int64(first))

	return GetFiltered(finalFilter, opts)
}

func GetFiltered(filter interface{}, opts *options.FindOptions) ([]Commerce, error) {
	commerces := []Commerce{}

	cursor, err := database.CollectionCommerces.Find(database.MongoContext, filter, opts)

	if err != nil {
		return commerces, err
	}

	for cursor.Next(database.MongoContext) {
		var commerce Commerce

		err := cursor.Decode(&commerce)

		if err != nil {
			return commerces, err
		}

		commerces = append(commerces, commerce)
	}

	if err := cursor.Err(); err != nil {
		return commerces, err
	}

	return commerces, nil
}
