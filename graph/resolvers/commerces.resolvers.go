package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/base64"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/helper"
	"chemin-du-local.bzh/graphql/internal/registeredpaymentmethod"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Storekeeper is the resolver for the storekeeper field.
func (r *commerceResolver) Storekeeper(ctx context.Context, obj *model.Commerce) (*model.User, error) {
	storekeeper, err := r.UsersService.GetUserById(obj.StorekeeperID)

	if err != nil {
		return nil, err
	}

	return storekeeper.ToModel(), nil
}

// Categories is the resolver for the categories field.
func (r *commerceResolver) Categories(ctx context.Context, obj *model.Commerce) ([]string, error) {
	commerceObjectID, err := primitive.ObjectIDFromHex(obj.ID)

	if err != nil {
		return nil, err
	}

	filter := bson.D{
		primitive.E{
			Key:   "commerceID",
			Value: commerceObjectID,
		},
	}

	databaseProducts, err := r.ProductsService.GetFiltered(filter, nil)
	categories := []string{}

	if err != nil {
		return nil, err
	}

	for _, databaseProduct := range databaseProducts {
		for _, category := range databaseProduct.Categories {
			if !helper.Contains(categories, category) {
				categories = append(categories, category)
			}
		}
	}

	return categories, nil
}

// Products is the resolver for the products field.
func (r *commerceResolver) Products(ctx context.Context, obj *model.Commerce, first *int, after *string, filters *model.ProductFilter) (*model.ProductConnection, error) {
	var decodedCursor *string

	if after != nil {
		bytes, err := base64.StdEncoding.DecodeString(*after)

		if err != nil {
			return nil, err
		}

		decodedCursorString := string(bytes)
		decodedCursor = &decodedCursorString
	}

	databaseProducts, err := r.ProductsService.GetPaginated(obj.ID, decodedCursor, *first, filters)

	if err != nil {
		return nil, err
	}

	// On construit les edges
	edges := []*model.ProductEdge{}

	for _, datadatabaseProduct := range databaseProducts {
		productEdge := model.ProductEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(datadatabaseProduct.ID.Hex())),
			Node:   datadatabaseProduct.ToModel(),
		}

		edges = append(edges, &productEdge)
	}

	itemCount := len(edges)

	// Si jamais il n'y a pas de produits, on veut quand même renvoyer un
	// tableau vide
	if itemCount == 0 {
		return &model.ProductConnection{
			Edges:    edges,
			PageInfo: &model.ProductPageInfo{},
		}, nil
	}

	hasNextPage := !databaseProducts[itemCount-1].IsLast(r.ProductsService)

	pageInfo := model.ProductPageInfo{
		StartCursor: base64.StdEncoding.EncodeToString([]byte(edges[0].Node.ID)),
		EndCursor:   base64.StdEncoding.EncodeToString([]byte(edges[itemCount-1].Node.ID)),
		HasNextPage: hasNextPage,
	}

	connection := model.ProductConnection{
		Edges:    edges[:itemCount],
		PageInfo: &pageInfo,
	}

	return &connection, nil
}

// ProductsAvailableForClickAndCollect is the resolver for the productsAvailableForClickAndCollect field.
func (r *commerceResolver) ProductsAvailableForClickAndCollect(ctx context.Context, obj *model.Commerce) ([]*model.Product, error) {
	databaseCommere, err := r.CommercesService.GetById(obj.ID)

	if err != nil {
		return nil, err
	}

	if databaseCommere == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	productsResult := []*model.Product{}
	for _, productId := range databaseCommere.ProductsAvailableForClickAndCollect {
		databaseProduct, err := r.ProductsService.GetById(productId)

		if err != nil {
			return nil, err
		}

		if databaseProduct != nil {
			productsResult = append(productsResult, databaseProduct.ToModel())
		}
	}

	return productsResult, nil
}

// DefaultPaymentMethod is the resolver for the defaultPaymentMethod field.
func (r *commerceResolver) DefaultPaymentMethod(ctx context.Context, obj *model.Commerce) (*model.RegisteredPaymentMethod, error) {
	databaseCommerce, err := r.CommercesService.GetById(obj.ID)

	if err != nil {
		return nil, err
	}

	if databaseCommerce == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	if databaseCommerce.DefaultPaymentMethodID == nil {
		return nil, nil
	}

	details, err := registeredpaymentmethod.GetPaymentMethodDetails(*databaseCommerce.DefaultPaymentMethodID)

	if err != nil {
		return nil, err
	}

	return details, nil
}

// Paniers is the resolver for the paniers field.
func (r *commerceResolver) Paniers(ctx context.Context, obj *model.Commerce, first *int, after *string, filters *model.PanierFilter) (*model.PanierConnection, error) {
	var decodedCursor *string

	if after != nil {
		bytes, err := base64.StdEncoding.DecodeString(*after)

		if err != nil {
			return nil, err
		}

		decodedCursorString := string(bytes)
		decodedCursor = &decodedCursorString
	}

	databasePaniers, err := r.PaniersService.GetPaginated(obj.ID, decodedCursor, *first, filters)

	if err != nil {
		return nil, err
	}

	// On construit les edges
	edges := []*model.PanierEdge{}

	for _, databasePanier := range databasePaniers {
		panierEdge := model.PanierEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(databasePanier.ID.Hex())),
			Node:   databasePanier.ToModel(),
		}

		edges = append(edges, &panierEdge)
	}

	itemCount := len(edges)

	// Si jamais il n'y a pas de paniers, on veut quand même renvoyer un
	// tableau vide
	if itemCount == 0 {
		return &model.PanierConnection{
			Edges:    edges,
			PageInfo: &model.PanierPageInfo{},
		}, nil
	}

	hasNextPage := !databasePaniers[itemCount-1].IsLast(r.PaniersService)

	pageInfo := model.PanierPageInfo{
		StartCursor: base64.StdEncoding.EncodeToString([]byte(edges[0].Node.ID)),
		EndCursor:   base64.StdEncoding.EncodeToString([]byte(edges[itemCount-1].Node.ID)),
		HasNextPage: hasNextPage,
	}

	connection := model.PanierConnection{
		Edges:    edges[:itemCount],
		PageInfo: &pageInfo,
	}

	return &connection, nil
}

// Commerce returns generated.CommerceResolver implementation.
func (r *Resolver) Commerce() generated.CommerceResolver { return &commerceResolver{r} }

type commerceResolver struct{ *Resolver }
