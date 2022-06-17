package address

import (
	"chemin-du-local.bzh/graphql/graph/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Address struct {
	ID            primitive.ObjectID `bson:"_id"`
	Number        *string            `bson:"number"`
	Route         *string            `bson:"route"`
	OptionalRoute *string            `bson:"optionalRoute"`
	PostalCode    *string            `bson:"postalCode"`
	City          *string            `bson:"city"`
}

func (address *Address) ToModel() *model.Address {
	return &model.Address{
		ID:            address.ID.Hex(),
		Number:        address.Number,
		Route:         address.Route,
		OptionalRoute: address.OptionalRoute,
		PostalCode:    address.PostalCode,
		City:          address.City,
	}
}
