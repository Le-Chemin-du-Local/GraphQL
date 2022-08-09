package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/auth"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/helper"
	"chemin-du-local.bzh/graphql/internal/products"
	"chemin-du-local.bzh/graphql/internal/registeredpaymentmethod"
	"chemin-du-local.bzh/graphql/internal/services/clickandcollect"
	"chemin-du-local.bzh/graphql/internal/services/commands"
	"chemin-du-local.bzh/graphql/internal/services/paniers"
	"chemin-du-local.bzh/graphql/internal/services/servicesinfo"
	"chemin-du-local.bzh/graphql/internal/users"
	"chemin-du-local.bzh/graphql/pkg/geojson"
	"chemin-du-local.bzh/graphql/pkg/jwt"
	"chemin-du-local.bzh/graphql/pkg/stripehandler"
	"github.com/99designs/gqlgen/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *cCCommandResolver) Products(ctx context.Context, obj *model.CCCommand) ([]*model.CCProduct, error) {
	return clickandcollect.GetProducts(obj.ID)
}

func (r *commandResolver) User(ctx context.Context, obj *model.Command) (*model.User, error) {
	user, err := commands.GetUser(obj.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

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

func (r *commandResolver) Status(ctx context.Context, obj *model.Command) (string, error) {
	status, err := commands.GetStatus(obj.ID)

	if err != nil {
		return "", err
	}

	return *status, nil
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

func (r *commerceCommandResolver) Commerce(ctx context.Context, obj *model.CommerceCommand) (*model.Commerce, error) {
	commerce, err := commands.CommerceGetCommerce(obj.ID)

	if err != nil {
		return nil, err
	}

	return commerce, nil
}

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

func (r *commerceCommandResolver) User(ctx context.Context, obj *model.CommerceCommand) (*model.User, error) {
	user, err := commands.CommerceGetUser(obj.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	// On doit d'abord vérifier que l'email n'est pas déjà prise
	existingUser, err := users.GetUserByEmail(strings.ToLower(input.Email))

	if existingUser != nil {
		return nil, &users.UserEmailAlreadyExistsError{}
	}

	if err != nil {
		return nil, err
	}

	databaseUser, err := users.Create(input)

	if err != nil {
		return nil, err
	}

	return databaseUser.ToModel(), nil
}

func (r *mutationResolver) Login(ctx context.Context, input model.Login) (string, error) {
	// On check d'abord le mot de passe
	isPasswordCorrect := users.Authenticate(input)

	if !isPasswordCorrect {
		return "", &users.UserPasswordIncorrect{}
	}

	// Puis on génère le token
	user, err := users.GetUserByEmail(strings.ToLower(input.Email))

	if user == nil || err != nil {
		return "", err
	}

	token, err := jwt.GenerateToken(user.ID.Hex())

	if err != nil {
		return "", err
	}

	return token, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, id *string, input map[string]interface{}) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateCommerce(ctx context.Context, userID string, input model.NewCommerce) (*model.Commerce, error) {
	// TODO: s'assurer de n'avoir qu'un seul commerce par commerçant

	databaseUser, err := users.GetUserById(userID)

	if err != nil {
		return nil, err
	}

	if databaseUser == nil {
		return nil, &users.UserNotFoundError{}
	}

	databaseCommerce, err := commerces.Create(input, databaseUser.ID)

	if err != nil {
		return nil, err
	}

	databaseUser.Role = users.USERROLE_STOREKEEPER
	err = users.Update(databaseUser)

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

	if changes["latitude"] != nil && changes["longitude"] != nil {
		castedLatitude := changes["latitude"].(json.Number)
		castedLongitude := changes["longitude"].(json.Number)

		latitude, err1 := castedLatitude.Float64()
		longitude, err2 := castedLongitude.Float64()

		if err1 != nil || err2 != nil {
			return nil, err1
		}

		databaseCommerce.AddressGeo = geojson.GeoJSON{
			Type:        "Point",
			Coordinates: []float64{longitude, latitude},
		}

	}

	err = commerces.Update(databaseCommerce, image, profilePicture)

	if err != nil {
		return nil, err
	}

	return databaseCommerce.ToModel(), nil
}

func (r *mutationResolver) CreateProduct(ctx context.Context, commerceID *string, input model.NewProduct) (*model.Product, error) {
	user := auth.ForContext(ctx)

	if user.Role == users.USERROLE_ADMIN && commerceID == nil {
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

		databaseCommerceID := databaseCommerce.ID.Hex()
		commerceID = &databaseCommerceID
	}

	databaseProduct, err := products.Create(*commerceID, input)

	if err != nil {
		return nil, err
	}

	return databaseProduct.ToModel(), nil
}

func (r *mutationResolver) CreateProducts(ctx context.Context, commerceID *string, input []*model.NewProduct) ([]*model.Product, error) {
	user := auth.ForContext(ctx)

	if user.Role == users.USERROLE_ADMIN && commerceID == nil {
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

		databaseCommerceID := databaseCommerce.ID.Hex()
		commerceID = &databaseCommerceID
	}

	result := []*model.Product{}

	for _, produdct := range input {
		databaseProduct, err := products.Create(*commerceID, *produdct)

		if err != nil {
			return nil, err
		}

		result = append(result, databaseProduct.ToModel())
	}

	return result, nil
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

func (r *mutationResolver) UpdateProducts(ctx context.Context, changes []*model.BulkChangesProduct) ([]*model.Product, error) {
	result := []*model.Product{}

	for _, change := range changes {
		databaseProduct, err := products.GetById(change.ID)

		if err != nil {
			return nil, err
		}

		if databaseProduct == nil {
			return nil, &products.ProductNotFoundError{}
		}

		helper.ApplyChanges(change.Changes, databaseProduct)

		var image *graphql.Upload

		if change.Changes["image"] != nil {
			castedImage := change.Changes["image"].(graphql.Upload)
			image = &castedImage
		}

		err = products.Update(databaseProduct, image)

		if err != nil {
			return nil, err
		}

		result = append(result, databaseProduct.ToModel())
	}

	return result, nil
}

func (r *mutationResolver) UpdateCommerceCommand(ctx context.Context, id string, changes map[string]interface{}) (*model.CommerceCommand, error) {
	databaseCommerceCommand, err := commands.CommerceGetById(id)

	if err != nil {
		return nil, err
	}

	if databaseCommerceCommand == nil {
		return nil, &commands.CommerceCommandNotFoundError{}
	}

	helper.ApplyChanges(changes, databaseCommerceCommand)

	err = commands.CommerceUpdate(databaseCommerceCommand)

	if err != nil {
		return nil, err
	}

	return databaseCommerceCommand.ToModel(), nil
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
		if auth.ForContext(ctx) == nil {
			return nil, &users.UserNotFoundError{}
		}

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

func (r *queryResolver) Commerces(ctx context.Context, first *int, after *string, filter *model.CommerceFilter) (*model.CommerceConnection, error) {
	var decodedCursor *string

	if after != nil {
		bytes, err := base64.StdEncoding.DecodeString(*after)

		if err != nil {
			return nil, err
		}

		decodedCursorString := string(bytes)
		decodedCursor = &decodedCursorString
	}

	databaseCommerces, err := commerces.GetPaginated(decodedCursor, *first, filter)

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

func (r *queryResolver) Commands(ctx context.Context, first *int, after *string, filter *model.CommandsFilter) (*model.CommandConnection, error) {
	if filter == nil {
		filter = &model.CommandsFilter{}
	}

	user := auth.ForContext(ctx)
	if filter.UserID == nil {
		userIDValue := user.ID.Hex()
		filter.UserID = &userIDValue
	}

	var decodedCursor *string

	if after != nil {
		bytes, err := base64.StdEncoding.DecodeString(*after)

		if err != nil {
			return nil, err
		}

		decodedCursorString := string(bytes)
		decodedCursor = &decodedCursorString
	}

	databaseCommands, err := commands.GetPaginated(decodedCursor, *first, filter)

	if err != nil {
		return nil, err
	}

	// On construit les edges
	edges := []*model.CommandEdge{}

	for _, datadatabaseCommand := range databaseCommands {
		commandEdge := model.CommandEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(datadatabaseCommand.ID.Hex())),
			Node:   datadatabaseCommand.ToModel(),
		}

		edges = append(edges, &commandEdge)
	}

	itemCount := len(edges)

	// Si jamais il n'y a pas de command, on veut quand même renvoyer un
	// tableau vide
	if itemCount == 0 {
		return &model.CommandConnection{
			Edges: edges,
			PageInfo: &model.CommandPageInfo{
				HasNextPage: false,
			},
		}, nil
	}

	hasNextPage := !databaseCommands[itemCount-1].IsLast()

	pageInfo := model.CommandPageInfo{
		StartCursor: base64.StdEncoding.EncodeToString([]byte(edges[0].Node.ID)),
		EndCursor:   base64.StdEncoding.EncodeToString([]byte(edges[itemCount-1].Node.ID)),
		HasNextPage: hasNextPage,
	}

	connection := model.CommandConnection{
		Edges:    edges[:itemCount],
		PageInfo: &pageInfo,
	}

	return &connection, nil
}

func (r *queryResolver) CommerceCommands(ctx context.Context, first *int, after *string, filter *model.CommerceCommandsFilter) (*model.CommerceCommandConnection, error) {
	if filter == nil {
		filter = &model.CommerceCommandsFilter{}
	}

	user := auth.ForContext(ctx)
	if user == nil {
		return nil, &users.UserAccessDenied{}
	}

	if filter.CommerceID == nil {
		userIDValue := user.ID.Hex()

		databaseCommerce, err := commerces.GetForUser(userIDValue)

		if err != nil {
			return nil, err
		}

		if databaseCommerce == nil {
			return nil, &commerces.CommerceErrorNotFound{}
		}

		commerceID := databaseCommerce.ID.Hex()
		filter.CommerceID = &commerceID
	}

	var decodedCursor *string

	if after != nil {
		bytes, err := base64.StdEncoding.DecodeString(*after)

		if err != nil {
			return nil, err
		}

		decodedCursorString := string(bytes)
		decodedCursor = &decodedCursorString
	}

	databaseCommands, err := commands.CommerceGetPaginated(decodedCursor, *first, filter)

	if err != nil {
		return nil, err
	}

	// On construit les edges
	edges := []*model.CommerceCommandEdge{}

	for _, datadatabaseCommand := range databaseCommands {
		commandEdge := model.CommerceCommandEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(datadatabaseCommand.ID.Hex())),
			Node:   datadatabaseCommand.ToModel(),
		}

		edges = append(edges, &commandEdge)
	}

	itemCount := len(edges)

	// Si jamais il n'y a pas de command, on veut quand même renvoyer un
	// tableau vide
	if itemCount == 0 {
		return &model.CommerceCommandConnection{
			Edges: edges,
			PageInfo: &model.CommerceCommandPageInfo{
				HasNextPage: false,
			},
		}, nil
	}

	hasNextPage := !databaseCommands[itemCount-1].IsLast()

	pageInfo := model.CommerceCommandPageInfo{
		StartCursor: base64.StdEncoding.EncodeToString([]byte(edges[0].Node.ID)),
		EndCursor:   base64.StdEncoding.EncodeToString([]byte(edges[itemCount-1].Node.ID)),
		HasNextPage: hasNextPage,
	}

	connection := model.CommerceCommandConnection{
		Edges:    edges[:itemCount],
		PageInfo: &pageInfo,
	}

	return &connection, nil
}

func (r *queryResolver) Command(ctx context.Context, id string) (*model.Command, error) {
	databaseCommand, err := commands.GetById(id)

	if err != nil {
		return nil, err
	}

	if databaseCommand == nil {
		return nil, &commands.CommandNotFoundError{}
	}

	return databaseCommand.ToModel(), nil
}

func (r *queryResolver) AllServicesInfo(ctx context.Context) ([]*model.ServiceInfo, error) {
	clickandcollect := servicesinfo.ClickAndCollect()
	paniers := servicesinfo.Paniers()

	return []*model.ServiceInfo{
		&clickandcollect,
		&paniers,
	}, nil
}

func (r *queryResolver) ServiceInfo(ctx context.Context, id string) (*model.ServiceInfo, error) {
	clickandcollect := servicesinfo.ClickAndCollect()
	paniers := servicesinfo.Paniers()

	if strings.Contains(id, "CLICKANDCOLLECT") {
		return &clickandcollect, nil
	} else if strings.Contains(id, "PANIERS") {
		return &paniers, nil
	}

	return nil, &servicesinfo.ServiceNotFoundError{}
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

func (r *userResolver) Basket(ctx context.Context, obj *model.User) (*model.Basket, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userResolver) RegisteredPaymentMethods(ctx context.Context, obj *model.User) ([]*model.RegisteredPaymentMethod, error) {
	if obj == nil {
		return nil, &users.UserAccessDenied{}
	}

	databaseUser, err := users.GetUserById(obj.ID)

	if err != nil {
		return nil, err
	}

	if databaseUser == nil {
		return nil, &users.UserNotFoundError{}
	}

	if databaseUser.StripID == nil {
		return nil, nil
	}

	return stripehandler.GetPaymentMethods(*databaseUser.StripID), nil
}

func (r *userResolver) DefaultPaymentMethod(ctx context.Context, obj *model.User) (*model.RegisteredPaymentMethod, error) {
	databaseUser, err := users.GetUserById(obj.ID)

	if err != nil {
		return nil, err
	}

	if databaseUser == nil {
		return nil, &users.UserNotFoundError{}
	}

	if databaseUser.DefaultPaymentMethod == nil {
		return nil, nil
	}

	for _, paymentMethod := range databaseUser.RegisteredPaymentMethods {
		if paymentMethod.StripeID != *databaseUser.DefaultPaymentMethod {
			continue
		}

		details, err := registeredpaymentmethod.GetPaymentMethodDetails(paymentMethod.StripeID)

		if err != nil {
			return nil, err
		}

		return &model.RegisteredPaymentMethod{
			Name:            paymentMethod.Name,
			StripeID:        paymentMethod.StripeID,
			CardBrand:       details.CardBrand,
			CardLast4Digits: details.CardLast4Digits,
		}, nil
	}

	return nil, nil
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

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Panier returns generated.PanierResolver implementation.
func (r *Resolver) Panier() generated.PanierResolver { return &panierResolver{r} }

// PanierCommand returns generated.PanierCommandResolver implementation.
func (r *Resolver) PanierCommand() generated.PanierCommandResolver { return &panierCommandResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type cCCommandResolver struct{ *Resolver }
type commandResolver struct{ *Resolver }
type commerceResolver struct{ *Resolver }
type commerceCommandResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type panierResolver struct{ *Resolver }
type panierCommandResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *commerceResolver) AddressDetailed(ctx context.Context, obj *model.Commerce) (*model.Address, error) {
	panic(fmt.Errorf("not implemented"))
}
