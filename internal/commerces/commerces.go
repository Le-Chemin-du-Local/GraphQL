package commerces

import (
	"log"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Commerce struct {
	ID              primitive.ObjectID `bson:"_id"`
	StorekeeperID   primitive.ObjectID `bson:"storekeeperID"`
	Name            string             `bson:"name"`
	Description     string             `bson:"description"`
	StorekeeperWord string             `bson:"storekeeperWord"`
	Address         string             `bson:"address"`
	Phone           string             `bson:"phone"`
	Email           string             `bson:"email"`
}

func (commerce *Commerce) ToModel() *model.Commerce {
	return &model.Commerce{
		ID:              commerce.ID.Hex(),
		StorekeeperID:   commerce.StorekeeperID.Hex(),
		Name:            commerce.Name,
		Description:     commerce.Description,
		StorekeeperWord: commerce.StorekeeperWord,
		Address:         commerce.Address,
		Phone:           commerce.Phone,
		Email:           commerce.Email,
	}
}

// Créateur de base de données

func Create(input model.NewCommerce, storekeeperID primitive.ObjectID) *Commerce {
	databaseCommerce := Commerce{
		ID:              primitive.NewObjectID(),
		StorekeeperID:   storekeeperID,
		Name:            input.Name,
		Description:     input.Description,
		StorekeeperWord: input.StorekeeperWord,
		Address:         input.Address,
		Phone:           input.Phone,
		Email:           input.Email,
	}

	_, err := database.CollectionCommerces.InsertOne(database.MongoContext, databaseCommerce)

	if err != nil {
		log.Fatal(err)
	}

	return &databaseCommerce
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
