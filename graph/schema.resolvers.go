package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/auth"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/helper"
	"chemin-du-local.bzh/graphql/internal/products"
	"chemin-du-local.bzh/graphql/internal/services/clickandcollect"
	"chemin-du-local.bzh/graphql/internal/services/paniers"
	"chemin-du-local.bzh/graphql/internal/users"
	"chemin-du-local.bzh/graphql/pkg/jwt"
	"github.com/99designs/gqlgen/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *cCCommandResolver) Products(ctx context.Context, obj *model.CCCommand) ([]*model.CCProduct, error) {
	return clickandcollect.GetProducts(obj.ID)
}

func (r *cCCommandResolver) User(ctx context.Context, obj *model.CCCommand) (*model.User, error) {
	databaseCCCommand, err := clickandcollect.GetById(obj.ID)

	if err != nil {
		return nil, err
	}

	if databaseCCCommand == nil {
		return nil, &clickandcollect.CCCommandNotFoundError{}
	}

	databaseUser, err := users.GetUserById(databaseCCCommand.UserID.Hex())

	if err != nil {
		return nil, err
	}

	if databaseUser == nil {
		return nil, &users.UserNotFoundError{}
	}

	return databaseUser.ToModel(), nil
}

func (r *commerceResolver) Storekeeper(ctx context.Context, obj *model.Commerce) (*model.User, error) {
	storekeeper, err := users.GetUserById(obj.StorekeeperID)

	if err != nil {
		return nil, err
	}

	return storekeeper.ToModel(), nil
}

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

func (r *commerceResolver) Cccommands(ctx context.Context, obj *model.Commerce, first *int, after *string, filters *model.CCCommandFilter) (*model.CCCommandConnection, error) {
	var decodedCursor *string

	if after != nil {
		bytes, err := base64.StdEncoding.DecodeString(*after)

		if err != nil {
			return nil, err
		}

		decodedCursorString := string(bytes)
		decodedCursor = &decodedCursorString
	}

	databaseCCCommands, err := clickandcollect.GetPaginated(obj.ID, decodedCursor, *first, filters)

	if err != nil {
		return nil, err
	}

	// On construit les edges
	edges := []*model.CCCommandEdge{}

	for _, datadatabaseCCCommand := range databaseCCCommands {
		cccommandEdge := model.CCCommandEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(datadatabaseCCCommand.ID.Hex())),
			Node:   datadatabaseCCCommand.ToModel(),
		}

		edges = append(edges, &cccommandEdge)
	}

	itemCount := len(edges)

	// Si jamais il n'y a pas de produits, on veut quand même renvoyer un
	// tableau vide
	if itemCount == 0 {
		return &model.CCCommandConnection{
			Edges:    edges,
			PageInfo: &model.CCCommandPageInfo{},
		}, nil
	}

	hasNextPage := !databaseCCCommands[itemCount-1].IsLast()

	pageInfo := model.CCCommandPageInfo{
		StartCursor: base64.StdEncoding.EncodeToString([]byte(edges[0].Node.ID)),
		EndCursor:   base64.StdEncoding.EncodeToString([]byte(edges[itemCount-1].Node.ID)),
		HasNextPage: hasNextPage,
	}

	connection := model.CCCommandConnection{
		Edges:    edges[:itemCount],
		PageInfo: &pageInfo,
	}

	return &connection, nil
}

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

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	// On doit d'abord vérifier que l'email n'est pas déjà prise
	existingUser, err := users.GetUserByEmail(input.Email)

	if existingUser != nil {
		return nil, &users.UserEmailAlreadyExistsError{}
	}

	if err != nil {
		return nil, err
	}

	databaseUser := users.Create(input)

	return databaseUser.ToModel(), nil
}

func (r *mutationResolver) Login(ctx context.Context, input model.Login) (string, error) {
	// On check d'abord le mot de passe
	isPasswordCorrect := users.Authenticate(input)

	if !isPasswordCorrect {
		return "", &users.UserPasswordIncorrect{}
	}

	// Puis on génère le token
	user, err := users.GetUserByEmail(input.Email)

	if user == nil || err != nil {
		return "", err
	}

	token, err := jwt.GenerateToken(user.ID.Hex())

	if err != nil {
		return "", err
	}

	return token, nil
}

func (r *mutationResolver) CreateCommerce(ctx context.Context, input model.NewCommerce) (*model.Commerce, error) {
	// TODO: s'assurer de n'avoir qu'un seul commerce par commerçant

	// Cas spécifique : seul les commerçant peuvent créer un commerce
	// pas même les admin
	user := auth.ForContext(ctx) // NOTE: pas besoin de vérifier le nil ici

	if user.Role != users.USERROLE_STOREKEEPER {
		return nil, &users.UserAccessDenied{}
	}

	databaseCommerce, err := commerces.Create(input, user.ID)

	if err != nil {
		return nil, err
	}

	return databaseCommerce.ToModel(), nil
}

func (r *mutationResolver) UpdateCommerce(ctx context.Context, id string, changes map[string]interface{}) (*model.Commerce, error) {
	databaseCommerce, err := commerces.GetById(id)

	if err != nil {
		return nil, err
	}

	if databaseCommerce == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	helper.ApplyChanges(changes, databaseCommerce)

	var image *graphql.Upload
	var profilePicture *graphql.Upload

	if changes["image"] != nil {
		castedImage := changes["image"].(graphql.Upload)
		image = &castedImage
	}

	if changes["profilePicture"] != nil {
		castedImage := changes["profilePicture"].(graphql.Upload)
		profilePicture = &castedImage
	}

	err = commerces.Update(databaseCommerce, image, profilePicture)

	if err != nil {
		return nil, err
	}

	return databaseCommerce.ToModel(), nil
}

func (r *mutationResolver) CreateProduct(ctx context.Context, input model.NewProduct) (*model.Product, error) {
	user := auth.ForContext(ctx)

	if user.Role == users.USERROLE_ADMIN && input.CommerceID == nil {
		return nil, &products.MustSpecifyCommerceIDError{}
	}

	// Si l'utilisateur est un commerçant, il doit créer des produits
	// pour son commerce
	if user.Role == users.USERROLE_STOREKEEPER {
		databaseCommerce, err := commerces.GetForUser(user.ID.Hex())

		// Cela permet aussi d'éviter qu'un commerçant créer un
		// produit sans commerce
		if err != nil {
			return nil, err
		}

		commerceID := databaseCommerce.ID.Hex()
		input.CommerceID = &commerceID
	}

	databaseProduct, err := products.Create(input)

	if err != nil {
		return nil, err
	}

	return databaseProduct.ToModel(), nil
}

func (r *mutationResolver) UpdateProduct(ctx context.Context, id string, changes map[string]interface{}) (*model.Product, error) {
	databaseProduct, err := products.GetById(id)

	if err != nil {
		return nil, err
	}

	if databaseProduct == nil {
		return nil, &products.ProductNotFoundError{}
	}

	helper.ApplyChanges(changes, databaseProduct)

	var image *graphql.Upload

	if changes["image"] != nil {
		castedImage := changes["image"].(graphql.Upload)
		image = &castedImage
	}

	err = products.Update(databaseProduct, image)

	if err != nil {
		return nil, err
	}

	return databaseProduct.ToModel(), nil
}

func (r *mutationResolver) Order(ctx context.Context, commerceID string, command model.NewCCCommand) (*model.CCCommand, error) {
	user := auth.ForContext(ctx)

	databaseCommand, err := clickandcollect.Create(user.ID, commerceID, command)

	if err != nil {
		return nil, err
	}

	return databaseCommand.ToModel(), nil
}

func (r *mutationResolver) UpdateCCCommand(ctx context.Context, id string, changes map[string]interface{}) (*model.CCCommand, error) {
	databaseCommand, err := clickandcollect.GetById(id)

	if err != nil {
		return nil, err
	}

	if databaseCommand == nil {
		return nil, &clickandcollect.CCCommandNotFoundError{}
	}

	helper.ApplyChanges(changes, databaseCommand)

	err = clickandcollect.Update(databaseCommand)

	if err != nil {
		return nil, err
	}

	return databaseCommand.ToModel(), nil
}

func (r *mutationResolver) CreatePanier(ctx context.Context, commerceID *string, input model.NewPanier) (*model.Panier, error) {
	user := auth.ForContext(ctx)

	var databaseCommerce *commerces.Commerce

	// On a d'abord besoin de trouver le commerce de l'utilisateur ou celui en paramètre
	if commerceID == nil {
		userDatabaseCommerce, err := commerces.GetForUser(user.ID.Hex())

		if err != nil {
			return nil, err
		}

		databaseCommerce = userDatabaseCommerce
	} else {
		commerceDatabaseCommerce, err := commerces.GetById(*commerceID)

		if err != nil {
			return nil, err
		}

		databaseCommerce = commerceDatabaseCommerce
	}

	if databaseCommerce == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	databasePanier, err := paniers.Create(databaseCommerce.ID, input)

	if err != nil {
		return nil, err
	}

	return databasePanier.ToModel(), nil
}

func (r *mutationResolver) UpdatePanier(ctx context.Context, id string, changes map[string]interface{}) (*model.Panier, error) {
	databasePanier, err := paniers.GetById(id)

	if err != nil {
		return nil, err
	}

	if databasePanier == nil {
		return nil, &paniers.PanierNotFoundError{}
	}

	// On doit faire un traitement spécial pour les produits
	products := []paniers.PanierProduct{}

	if changes["products"] == nil {
		products = databasePanier.Products
	} else {
		productsChange := changes["products"].([]interface{})
		for _, product := range productsChange {
			castedProduct := product.(map[string]interface{})

			productObjectID, err := primitive.ObjectIDFromHex(castedProduct["productID"].(string))

			if err != nil {
				return nil, err
			}

			castedQuantity := castedProduct["quantity"].(json.Number)
			quantity, err := castedQuantity.Int64()

			if err != nil {
				return nil, err
			}

			products = append(products, paniers.PanierProduct{
				ID:        primitive.NewObjectID(),
				ProductID: productObjectID,
				Quantity:  int(quantity),
			})
		}
	}

	helper.ApplyChanges(changes, databasePanier)
	databasePanier.Products = products

	var image *graphql.Upload

	if changes["image"] != nil {
		castedImage := changes["image"].(graphql.Upload)
		image = &castedImage
	}

	err = paniers.Update(databasePanier, image)

	if err != nil {
		return nil, err
	}

	return databasePanier.ToModel(), nil
}

func (r *panierResolver) Products(ctx context.Context, obj *model.Panier) ([]*model.PanierProduct, error) {
	return paniers.GetProducts(obj.ID)
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	databaseUsers, err := users.GetAllUser()

	if err != nil {
		return nil, err
	}

	users := []*model.User{}

	for _, databaseUser := range databaseUsers {
		user := databaseUser.ToModel()

		users = append(users, user)
	}

	return users, nil
}

func (r *queryResolver) User(ctx context.Context, id *string) (*model.User, error) {
	if id == nil {
		return auth.ForContext(ctx).ToModel(), nil
	}

	databaseUser, err := users.GetUserById(*id)

	if err != nil {
		return nil, err
	}

	if databaseUser == nil {
		return nil, &users.UserNotFoundError{}
	}

	return databaseUser.ToModel(), nil
}

func (r *queryResolver) Commerces(ctx context.Context, first *int, after *string) (*model.CommerceConnection, error) {
	var decodedCursor *string

	if after != nil {
		bytes, err := base64.StdEncoding.DecodeString(*after)

		if err != nil {
			return nil, err
		}

		decodedCursorString := string(bytes)
		decodedCursor = &decodedCursorString
	}

	databaseCommerces, err := commerces.GetPaginated(decodedCursor, *first)

	if err != nil {
		return nil, err
	}

	// On construit les edges
	edges := []*model.CommerceEdge{}

	for _, datadatabaseCommerce := range databaseCommerces {
		commerceEdge := model.CommerceEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(datadatabaseCommerce.ID.Hex())),
			Node:   datadatabaseCommerce.ToModel(),
		}

		edges = append(edges, &commerceEdge)
	}

	itemCount := len(edges)

	// Si jamais il n'y a pas de commerce, on veut quand même renvoyer un
	// tableau vide
	if itemCount == 0 {
		return &model.CommerceConnection{
			Edges:    edges,
			PageInfo: &model.CommercePageInfo{},
		}, nil
	}

	hasNextPage := !databaseCommerces[itemCount-1].IsLast()

	pageInfo := model.CommercePageInfo{
		StartCursor: base64.StdEncoding.EncodeToString([]byte(edges[0].Node.ID)),
		EndCursor:   base64.StdEncoding.EncodeToString([]byte(edges[itemCount-1].Node.ID)),
		HasNextPage: hasNextPage,
	}

	connection := model.CommerceConnection{
		Edges:    edges[:itemCount],
		PageInfo: &pageInfo,
	}

	return &connection, nil
}

func (r *queryResolver) Commerce(ctx context.Context, id *string) (*model.Commerce, error) {
	if id != nil {
		databaseCommerce, err := commerces.GetById(*id)

		if err != nil {
			return nil, err
		}

		if databaseCommerce == nil {
			return nil, &commerces.CommerceErrorNotFound{}
		}

		return databaseCommerce.ToModel(), nil
	}

	user := auth.ForContext(ctx)

	if user == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	databaseCommerce, err := commerces.GetForUser(user.ID.Hex())

	if err != nil {
		return nil, err
	}

	if databaseCommerce == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	return databaseCommerce.ToModel(), nil
}

func (r *queryResolver) Product(ctx context.Context, id string) (*model.Product, error) {
	databaseProduct, err := products.GetById(id)

	if err != nil {
		return nil, err
	}

	return databaseProduct.ToModel(), nil
}

func (r *queryResolver) Cccommand(ctx context.Context, id string) (*model.CCCommand, error) {
	databaseCCCommand, err := clickandcollect.GetById(id)

	if err != nil {
		return nil, err
	}

	if databaseCCCommand == nil {
		return nil, &clickandcollect.CCCommandNotFoundError{}
	}

	return databaseCCCommand.ToModel(), nil
}

func (r *queryResolver) Panier(ctx context.Context, id string) (*model.Panier, error) {
	databasePanier, err := paniers.GetById(id)

	if err != nil {
		return nil, err
	}

	if databasePanier == nil {
		return nil, &paniers.PanierNotFoundError{}
	}

	return databasePanier.ToModel(), nil
}

func (r *userResolver) Commerce(ctx context.Context, obj *model.User) (*model.Commerce, error) {
	databaseCommerce, err := commerces.GetForUser(obj.ID)

	if err != nil {
		return nil, err
	}

	if databaseCommerce == nil {
		return nil, nil
	}

	return databaseCommerce.ToModel(), nil
}

// CCCommand returns generated.CCCommandResolver implementation.
func (r *Resolver) CCCommand() generated.CCCommandResolver { return &cCCommandResolver{r} }

// Commerce returns generated.CommerceResolver implementation.
func (r *Resolver) Commerce() generated.CommerceResolver { return &commerceResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Panier returns generated.PanierResolver implementation.
func (r *Resolver) Panier() generated.PanierResolver { return &panierResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type cCCommandResolver struct{ *Resolver }
type commerceResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type panierResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *panierResolver) EndingDate(ctx context.Context, obj *model.Panier) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}
func (r *panierResolver) Price(ctx context.Context, obj *model.Panier) (float64, error) {
	panic(fmt.Errorf("not implemented"))
}
