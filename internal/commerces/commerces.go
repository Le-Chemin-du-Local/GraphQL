package commerces

import (
	"bytes"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/address"
	"chemin-du-local.bzh/graphql/internal/config"
	"chemin-du-local.bzh/graphql/internal/database"
	"chemin-du-local.bzh/graphql/internal/services/servicesinfo"
	"chemin-du-local.bzh/graphql/pkg/geojson"
	"github.com/99designs/gqlgen/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Commerce struct {
	ID                                  primitive.ObjectID  `bson:"_id"`
	StorekeeperID                       primitive.ObjectID  `bson:"storekeeperID"`
	Siret                               string              `bson:"siret"`
	Name                                string              `bson:"name"`
	Description                         string              `bson:"description"`
	StorekeeperWord                     string              `bson:"storekeeperWord"`
	Address                             address.Address     `bson:"address"`
	AddressGeo                          geojson.GeoJSON     `bson:"addressGeo"`
	Phone                               string              `bson:"phone"`
	Email                               string              `bson:"email"`
	IBANOwner                           *string             `bson:"ibanOwner"`
	IBAN                                *string             `bson:"iban"`
	BIC                                 *string             `bson:"bic"`
	Facebook                            *string             `bson:"facebook"`
	Twitter                             *string             `bson:"twitter"`
	Instagram                           *string             `bson:"instagram"`
	BusinessHours                       model.BusinessHours `bson:"businesHours"`
	ClickAndCollectHours                model.BusinessHours `bson:"clickAndCollectHours"`
	Services                            []string            `bson:"services"`
	ProductsAvailableForClickAndCollect []string            `bson:"productsAvailableForClickAndCollect"`
	StripID                             *string             `bson:"stripeID"`
	DefaultPaymentMethodID              *string             `bson:"defaultPaymentMethodID"`
	LastBilledDate                      *time.Time          `bson:"lastBilledDate"`
	Balance                             float64             `bson:"balance"`
	DueBalance                          float64             `bson:"dueBalance"`
	Transferts                          []model.Transfert   `bson:"transferts"`
}

func (commerce *Commerce) ToModel() *model.Commerce {
	return &model.Commerce{
		ID:                   commerce.ID.Hex(),
		StorekeeperID:        commerce.StorekeeperID.Hex(),
		Siret:                commerce.Siret,
		Name:                 commerce.Name,
		Description:          commerce.Description,
		StorekeeperWord:      commerce.StorekeeperWord,
		Address:              *commerce.Address.ToModel(),
		Latitude:             commerce.AddressGeo.Coordinates[1],
		Longitude:            commerce.AddressGeo.Coordinates[0],
		Phone:                commerce.Phone,
		Email:                commerce.Email,
		IBANOwner:            commerce.IBANOwner,
		IBAN:                 commerce.IBAN,
		BIC:                  commerce.BIC,
		Facebook:             commerce.Facebook,
		Twitter:              commerce.Twitter,
		Instagram:            commerce.Instagram,
		BusinessHours:        commerce.BusinessHours,
		ClickAndCollectHours: commerce.ClickAndCollectHours,
		Services:             commerce.Services,
		LastBilledDate:       commerce.LastBilledDate,
		Balance:              commerce.Balance,
		DueBalance:           commerce.DueBalance,
		Transferts:           commerce.Transferts,
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
	var clickAndCollectHours model.BusinessHours

	if input.BusinessHours != nil {
		businessHours = *input.BusinessHours.ToModel()
	}

	if input.ClickAndCollectHours != nil {
		clickAndCollectHours = *input.ClickAndCollectHours.ToModel()
	}

	databaseCommerce := Commerce{
		ID:              commerceObjectID,
		StorekeeperID:   storekeeperID,
		Name:            input.Name,
		Siret:           input.Siret,
		Description:     description,
		StorekeeperWord: storekeeperWord,
		Address: address.Address{
			ID:            primitive.NewObjectID(),
			Number:        input.Address.Number,
			Route:         input.Address.Route,
			OptionalRoute: input.Address.OptionalRoute,
			PostalCode:    input.Address.PostalCode,
			City:          input.Address.City,
		},
		AddressGeo: geojson.GeoJSON{
			Type:        "Point",
			Coordinates: []float64{input.Latitude, input.Longitude},
		},
		Phone:                input.Phone,
		Email:                input.Email,
		Facebook:             input.Facebook,
		Twitter:              input.Twitter,
		Instagram:            input.Instagram,
		BusinessHours:        businessHours,
		ClickAndCollectHours: clickAndCollectHours,
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

func UpdateBalancesForOrder(commerceID string, price int, priceClickAndCollect float64, pricePaniers float64) error {
	commerce, err := GetById(commerceID)

	if err != nil {
		return err
	}

	if commerce == nil {
		return &CommerceErrorNotFound{}
	}

	for _, service := range commerce.Services {
		if strings.Contains(service, "CLICKANDCOLLECT_T") {
			clickandcollectInfo := servicesinfo.ClickAndCollect()
			priceToAdd := math.Round(clickandcollectInfo.TransactionPercentage*priceClickAndCollect) / 100

			commerce.DueBalance = commerce.DueBalance + priceToAdd
		} else if strings.Contains(service, "PANIERS_T") {
			paniersInfo := servicesinfo.Paniers()
			priceToAdd := math.Round(paniersInfo.TransactionPercentage*pricePaniers) / 100

			commerce.DueBalance = commerce.DueBalance + priceToAdd
		}
	}

	commerce.Balance = commerce.Balance + (float64(price) / 100)

	err = Update(commerce, nil, nil)

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

func GetPaginated(startValue *string, first int, filter *model.CommerceFilter) ([]Commerce, int, error) {
	// On doit faire un filtre spécifique si on veut commencer
	// le curseur à l'ID de départ
	var finalFilter bson.M
	var matchStage bson.D

	skip := 0

	if startValue != nil {
		if strings.Split(*startValue, ":")[0] == "offset" {
			skipValue, err := strconv.Atoi(strings.Split(*startValue, ":")[1])

			if err != nil {
				return nil, 0, err
			}

			skip = skipValue

		} else {
			objectID, err := primitive.ObjectIDFromHex(*startValue)

			if err != nil {
				return nil, 0, err
			}

			finalFilter = bson.M{
				"_id": bson.M{
					"$gt": objectID,
				},
			}
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

		matchStage = bson.D{{Key: "$geoNear", Value: bson.M{
			"near": bson.M{
				"type": "Point",
				"coordinates": []float64{
					*filter.NearLongitude,
					*filter.NearLatitude,
				},
			},
			"distanceField": "dist.calculated",
			"maxDistance":   maxDistance,
		}}}
	}

	opts := options.Find()
	opts.SetLimit(int64(first))
	opts.SetSkip(int64(skip))

	result, err := GetFiltered(finalFilter, opts)

	if err != nil {
		return nil, 0, err
	}

	countStage := bson.D{{Key: "$count", Value: "total_documents"}}

	pipeline := mongo.Pipeline{}

	if matchStage != nil {
		pipeline = append(pipeline, matchStage)
	}

	pipeline = append(pipeline, countStage)

	cursor, err := database.CollectionCommerces.Aggregate(database.MongoContext, pipeline)

	if err != nil {
		return nil, 0, err
	}

	var results []bson.D
	if err = cursor.All(database.MongoContext, &results); err != nil {
		return nil, 0, err
	}

	count := results[0][0].Value.(int32)

	return result, int(count), err
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
