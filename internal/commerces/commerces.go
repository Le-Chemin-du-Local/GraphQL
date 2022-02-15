package commerces

import (
	"log"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
