package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/base64"
	"fmt"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/helper"
	"chemin-du-local.bzh/graphql/internal/products"
	"chemin-du-local.bzh/graphql/internal/registeredpaymentmethod"
	"chemin-du-local.bzh/graphql/internal/services/clickandcollect"
	"chemin-du-local.bzh/graphql/internal/services/commands"
	"chemin-du-local.bzh/graphql/internal/services/paniers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Products is the resolver for the products field.
func (r *cCCommandResolver) Products(ctx context.Context, obj *model.CCCommand) ([]*model.CCProduct, error) {
	return clickandcollect.GetProducts(obj.ID)
}

// User is the resolver for the user field.
func (r *commandResolver) User(ctx context.Context, obj *model.Command) (*model.User, error) {
	user, err := r.CommandsService.GetUser(obj.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Commerces is the resolver for the commerces field.
func (r *commandResolver) Commerces(ctx context.Context, obj *model.Command) ([]*model.CommerceCommand, error) {
	databaseCommerceCommands, err := commands.CommerceGetForCommand(obj.ID)

	if err != nil {
		return nil, err
	}

	commerceCommands := []*model.CommerceCommand{}

	for _, databaseCommerceCommand := range databaseCommerceCommands {
		commerceCommands = append(commerceCommands, databaseCommerceCommand.ToModel())
	}

	return commerceCommands, nil
}

// Status is the resolver for the status field.
func (r *commandResolver) Status(ctx context.Context, obj *model.Command) (string, error) {
	status, err := commands.GetStatus(obj.ID)

	if err != nil {
		return "", err
	}

	return *status, nil
}

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

	databaseProducts, err := products.GetFiltered(filter, nil)
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

	databaseProducts, err := products.GetPaginated(obj.ID, decodedCursor, *first, filters)

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

	hasNextPage := !databaseProducts[itemCount-1].IsLast()

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
	databaseCommere, err := commerces.GetById(obj.ID)

	if err != nil {
		return nil, err
	}

	if databaseCommere == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	productsResult := []*model.Product{}
	for _, productId := range databaseCommere.ProductsAvailableForClickAndCollect {
		databaseProduct, err := products.GetById(productId)

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
	databaseCommerce, err := commerces.GetById(obj.ID)

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

	databasePaniers, err := paniers.GetPaginated(obj.ID, decodedCursor, *first, filters)

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

	hasNextPage := !databasePaniers[itemCount-1].IsLast()

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

// Commerce is the resolver for the commerce field.
func (r *commerceCommandResolver) Commerce(ctx context.Context, obj *model.CommerceCommand) (*model.Commerce, error) {
	commerce, err := commands.CommerceGetCommerce(obj.ID)

	if err != nil {
		return nil, err
	}

	return commerce, nil
}

// Cccommands is the resolver for the cccommands field.
func (r *commerceCommandResolver) Cccommands(ctx context.Context, obj *model.CommerceCommand) ([]*model.CCCommand, error) {
	databaseCCCommands, err := clickandcollect.GetForCommmerceCommand(obj.ID)

	if err != nil {
		return nil, err
	}

	cccommands := []*model.CCCommand{}

	for _, databaseCCCommand := range databaseCCCommands {
		cccommands = append(cccommands, databaseCCCommand.ToModel())
	}

	return cccommands, nil
}

// Paniers is the resolver for the paniers field.
func (r *commerceCommandResolver) Paniers(ctx context.Context, obj *model.CommerceCommand) ([]*model.PanierCommand, error) {
	databasePanierCommands, err := paniers.GetCommandsForCommerceCommand(obj.ID)

	if err != nil {
		return nil, err
	}

	panierCommands := []*model.PanierCommand{}

	for _, databasePanierCommand := range databasePanierCommands {
		panierCommands = append(panierCommands, databasePanierCommand.ToModel())
	}

	return panierCommands, nil
}

// User is the resolver for the user field.
func (r *commerceCommandResolver) User(ctx context.Context, obj *model.CommerceCommand) (*model.User, error) {
	user, err := r.CommerceCommandsService.GetUser(obj.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Products is the resolver for the products field.
func (r *panierResolver) Products(ctx context.Context, obj *model.Panier) ([]*model.PanierProduct, error) {
	return paniers.GetProducts(obj.ID)
}

// Panier is the resolver for the panier field.
func (r *panierCommandResolver) Panier(ctx context.Context, obj *model.PanierCommand) (*model.Panier, error) {
	databasePanierCommand, err := paniers.GetCommandById(obj.ID)

	if err != nil {
		return nil, err
	}

	if databasePanierCommand == nil {
		return nil, &paniers.PanierNotFoundError{}
	}

	databasePanier, err := paniers.GetById(databasePanierCommand.PanierID.Hex())

	if err != nil {
		return nil, err
	}

	if databasePanier == nil {
		return nil, &paniers.PanierNotFoundError{}
	}

	return databasePanier.ToModel(), nil
}

// CCCommand returns generated.CCCommandResolver implementation.
func (r *Resolver) CCCommand() generated.CCCommandResolver { return &cCCommandResolver{r} }

// Command returns generated.CommandResolver implementation.
func (r *Resolver) Command() generated.CommandResolver { return &commandResolver{r} }

// Commerce returns generated.CommerceResolver implementation.
func (r *Resolver) Commerce() generated.CommerceResolver { return &commerceResolver{r} }

// CommerceCommand returns generated.CommerceCommandResolver implementation.
func (r *Resolver) CommerceCommand() generated.CommerceCommandResolver {
	return &commerceCommandResolver{r}
}

// Panier returns generated.PanierResolver implementation.
func (r *Resolver) Panier() generated.PanierResolver { return &panierResolver{r} }

// PanierCommand returns generated.PanierCommandResolver implementation.
func (r *Resolver) PanierCommand() generated.PanierCommandResolver { return &panierCommandResolver{r} }

type cCCommandResolver struct{ *Resolver }
type commandResolver struct{ *Resolver }
type commerceResolver struct{ *Resolver }
type commerceCommandResolver struct{ *Resolver }
type panierResolver struct{ *Resolver }
type panierCommandResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *commerceResolver) AddressDetailed(ctx context.Context, obj *model.Commerce) (*model.Address, error) {
	panic(fmt.Errorf("not implemented"))
}
