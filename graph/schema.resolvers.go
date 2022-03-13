package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/base64"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/auth"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/helper"
	"chemin-du-local.bzh/graphql/internal/products"
	"chemin-du-local.bzh/graphql/internal/users"
	"chemin-du-local.bzh/graphql/pkg/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

	databaseCommerce := commerces.Create(input, user.ID)

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

	databaseProduct := products.Create(input)

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

	err = products.Update(databaseProduct)

	if err != nil {
		return nil, err
	}

	return databaseProduct.ToModel(), nil
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

	return databaseCommerce.ToModel(), nil
}

func (r *queryResolver) Product(ctx context.Context, id string) (*model.Product, error) {
	databaseProduct, err := products.GetById(id)

	if err != nil {
		return nil, err
	}

	return databaseProduct.ToModel(), nil
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

// Commerce returns generated.CommerceResolver implementation.
func (r *Resolver) Commerce() generated.CommerceResolver { return &commerceResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type commerceResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
