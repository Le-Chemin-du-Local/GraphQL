package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/mail"
	"strings"
	"time"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/auth"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/helper"
	"chemin-du-local.bzh/graphql/internal/products"
	"chemin-du-local.bzh/graphql/internal/services/commands"
	"chemin-du-local.bzh/graphql/internal/services/paniers"
	"chemin-du-local.bzh/graphql/internal/services/servicesinfo"
	"chemin-du-local.bzh/graphql/internal/users"
	"chemin-du-local.bzh/graphql/pkg/geojson"
	"chemin-du-local.bzh/graphql/pkg/jwt"
	"chemin-du-local.bzh/graphql/pkg/utils"
	"github.com/99designs/gqlgen/graphql"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	// On vérifie que l'adresse mail est valide
	_, err := mail.ParseAddress(input.Email)

	if err != nil {
		return nil, &users.UserEmailAddressInvalidError{}
	}

	// On doit d'abord vérifier que l'email n'est pas déjà prise
	existingUser, err := r.UsersService.GetUserByEmail(strings.ToLower(input.Email))

	if existingUser != nil {
		return nil, &users.UserEmailAlreadyExistsError{}
	}

	if err != nil {
		return nil, err
	}

	databaseUser, err := r.UsersService.Create(input)

	if err != nil {
		return nil, err
	}

	user := databaseUser.ToModel()

	// if input.Commerce != nil {
	// 	databaseCommerce, err := commerces.GetForUser(databaseUser.ID.Hex())

	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	commerceID := databaseCommerce.ID.Hex()
	// 	user.CommerceID = &commerceID
	// }

	return user, nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, input model.Login) (string, error) {
	// On check d'abord le mot de passe
	isPasswordCorrect := r.UsersService.Authenticate(input)

	if !isPasswordCorrect {
		return "", &users.UserPasswordIncorrect{}
	}

	// Puis on génère le token
	user, err := r.UsersService.GetUserByEmail(strings.ToLower(input.Email))

	if user == nil || err != nil {
		return "", err
	}

	token, err := jwt.GenerateToken(user.ID.Hex())

	if err != nil {
		return "", err
	}

	return token, nil
}

// UpdateUser is the resolver for the updateUser field.
func (r *mutationResolver) UpdateUser(ctx context.Context, id *string, input map[string]interface{}) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

// CreateCommerce is the resolver for the createCommerce field.
func (r *mutationResolver) CreateCommerce(ctx context.Context, userID string, input model.NewCommerce) (*model.Commerce, error) {
	// TODO: s'assurer de n'avoir qu'un seul commerce par commerçant

	databaseUser, err := r.UsersService.GetUserById(userID)

	if err != nil {
		return nil, err
	}

	if databaseUser == nil {
		return nil, &users.UserNotFoundError{}
	}

	databaseCommerce, err := r.CommercesService.Create(input, databaseUser.ID)

	if err != nil {
		return nil, err
	}

	databaseUser.Role = users.USERROLE_STOREKEEPER
	err = r.UsersService.Update(databaseUser)

	if err != nil {
		return nil, err
	}

	return databaseCommerce.ToModel(), nil
}

// UpdateCommerce is the resolver for the updateCommerce field.
func (r *mutationResolver) UpdateCommerce(ctx context.Context, id string, changes map[string]interface{}) (*model.Commerce, error) {
	databaseCommerce, err := r.CommercesService.GetById(id)

	if err != nil {
		return nil, err
	}

	if databaseCommerce == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	// On a besoin d'un workaround pour les services
	tempServices := databaseCommerce.Services
	helper.ApplyChanges(changes, databaseCommerce)
	databaseCommerce.Services = tempServices

	// Les changements de service
	if changes["services"] != nil {
		servicesChanges := changes["services"].([]interface{})

		for _, serviceChanges := range servicesChanges {
			// On doit convertir la map en service
			jsonString, _ := json.Marshal(serviceChanges)
			castedServiceChange := model.ChangesService{}
			json.Unmarshal(jsonString, &castedServiceChange)

			var serviceInfo *model.ServiceInfo
			nextBillingTime := databaseCommerce.LastBilledDate.AddDate(0, 0, 30)
			nextBillingTimeRounded := time.Date(
				nextBillingTime.Year(),
				nextBillingTime.Month(),
				nextBillingTime.Day(),
				1, 0, 0, 0, time.Local,
			)
			nowRounded := time.Date(
				time.Now().Year(),
				time.Now().Month(),
				time.Now().Day(),
				1, 0, 0, 0, time.Local,
			)

			if strings.Contains(castedServiceChange.ServiceID, "CLICKANDCOLLECT") {
				clickandcollectInfo := servicesinfo.ClickAndCollect()
				serviceInfo = &clickandcollectInfo
			}
			if strings.Contains(castedServiceChange.ServiceID, "PANIERS") {
				paniersInfo := servicesinfo.Paniers()
				serviceInfo = &paniersInfo
			}

			if serviceInfo == nil {
				return nil, &servicesinfo.ServiceNotFoundError{}
			}

			remainingDays := nextBillingTimeRounded.Sub(nowRounded).Hours() / 24
			remainingPrice := serviceInfo.MonthPrice / 30 * remainingDays
			remainingPriceRounded := math.Round(remainingPrice*100) / 100

			if castedServiceChange.UpdateType == "ADD" {
				if castedServiceChange.ServiceID[len(castedServiceChange.ServiceID)-1:] == "M" {
					databaseCommerce.DueBalance = databaseCommerce.DueBalance + remainingPriceRounded
				}

				databaseCommerce.Services = append(databaseCommerce.Services, castedServiceChange.ServiceID)

				// Pour le Click&Collect, on veut ajouter tous les produits par défaut
				productsOfCommerce, err := r.ProductsService.GetForCommerce(databaseCommerce.ID.Hex())

				if err != nil {
					productsId := []string{}

					for _, product := range productsOfCommerce {
						productsId = append(productsId, product.ID.Hex())
					}

					databaseCommerce.ProductsAvailableForClickAndCollect = productsId
				}
			}

			if castedServiceChange.UpdateType == "UPDATE" {
				for index, databaseService := range databaseCommerce.Services {
					if strings.Contains(databaseService, strings.Split(castedServiceChange.ServiceID, "_")[0]) {
						if strings.Split(castedServiceChange.ServiceID, "_")[1] == "M" {
							if !strings.Contains(databaseService, "_UPDATE") {
								databaseCommerce.DueBalance = databaseCommerce.DueBalance + remainingPriceRounded
							}

							databaseCommerce.Services[index] = castedServiceChange.ServiceID
						} else {
							databaseCommerce.Services[index] = databaseService + "_UPDATE"
						}
					}
				}
			}

			fmt.Println(castedServiceChange)
			if castedServiceChange.UpdateType == "REMOVE" {
				indexesToRemove := []int{}

				for index, databaseService := range databaseCommerce.Services {
					if strings.Contains(databaseService, strings.Split(castedServiceChange.ServiceID, "_")[0]) {
						if strings.Split(databaseService, "_")[1] == "M" {
							databaseCommerce.Services[index] = strings.Split(castedServiceChange.ServiceID, "_")[0] + "_M_REMOVE"
						} else {
							indexesToRemove = append(indexesToRemove, index)
						}
					}
				}

				for _, index := range indexesToRemove {
					databaseCommerce.Services = utils.RemoveString(databaseCommerce.Services, index)
				}
			}
		}
	}

	// Les changements d'images
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

	err = r.CommercesService.Update(databaseCommerce, image, profilePicture)

	if err != nil {
		return nil, err
	}

	return databaseCommerce.ToModel(), nil
}

// CreateProduct is the resolver for the createProduct field.
func (r *mutationResolver) CreateProduct(ctx context.Context, commerceID *string, input model.NewProduct) (*model.Product, error) {
	user := auth.ForContext(ctx)

	if user.Role == users.USERROLE_ADMIN && commerceID == nil {
		return nil, &products.MustSpecifyCommerceIDError{}
	}

	// Si l'utilisateur est un commerçant, il doit créer des produits
	// pour son commerce
	if user.Role == users.USERROLE_STOREKEEPER {
		databaseCommerce, err := r.CommercesService.GetForUser(user.ID.Hex())

		// Cela permet aussi d'éviter qu'un commerçant créer un
		// produit sans commerce
		if err != nil {
			return nil, err
		}

		databaseCommerceID := databaseCommerce.ID.Hex()
		commerceID = &databaseCommerceID
	}

	databaseProduct, err := r.ProductsService.Create(*commerceID, input)

	if err != nil {
		return nil, err
	}

	// Si le commerçant a souscrit au Click&Collect, le produit doit automatiquement y être ajouté
	databaseCommerce, err := r.CommercesService.GetForUser(user.ID.Hex())

	if err != nil {
		return databaseProduct.ToModel(), nil
	}

	commerceSubscribedToCC := false

	for _, service := range databaseCommerce.Services {
		if strings.Contains(service, "CLICKANDCOLLECT") {
			commerceSubscribedToCC = true
			break
		}
	}

	if commerceSubscribedToCC {
		databaseCommerce.ProductsAvailableForClickAndCollect = append(databaseCommerce.ProductsAvailableForClickAndCollect, databaseProduct.ID.Hex())
	}

	r.CommercesService.Update(databaseCommerce, nil, nil)

	return databaseProduct.ToModel(), nil
}

// CreateProducts is the resolver for the createProducts field.
func (r *mutationResolver) CreateProducts(ctx context.Context, commerceID *string, input []*model.NewProduct) ([]*model.Product, error) {
	user := auth.ForContext(ctx)

	if user.Role == users.USERROLE_ADMIN && commerceID == nil {
		return nil, &products.MustSpecifyCommerceIDError{}
	}

	// Si l'utilisateur est un commerçant, il doit créer des produits
	// pour son commerce
	if user.Role == users.USERROLE_STOREKEEPER {
		databaseCommerce, err := r.CommercesService.GetForUser(user.ID.Hex())

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
		databaseProduct, err := r.ProductsService.Create(*commerceID, *produdct)

		if err != nil {
			return nil, err
		}

		result = append(result, databaseProduct.ToModel())
	}

	return result, nil
}

// UpdateProduct is the resolver for the updateProduct field.
func (r *mutationResolver) UpdateProduct(ctx context.Context, id string, changes map[string]interface{}) (*model.Product, error) {
	databaseProduct, err := r.ProductsService.GetById(id)

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

	err = r.ProductsService.Update(databaseProduct, image)

	if err != nil {
		return nil, err
	}

	return databaseProduct.ToModel(), nil
}

// UpdateProducts is the resolver for the updateProducts field.
func (r *mutationResolver) UpdateProducts(ctx context.Context, changes []*model.BulkChangesProduct) ([]*model.Product, error) {
	result := []*model.Product{}

	for _, change := range changes {
		databaseProduct, err := r.ProductsService.GetById(change.ID)

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

		err = r.ProductsService.Update(databaseProduct, image)

		if err != nil {
			return nil, err
		}

		result = append(result, databaseProduct.ToModel())
	}

	return result, nil
}

// UpdateCommerceCommand is the resolver for the updateCommerceCommand field.
func (r *mutationResolver) UpdateCommerceCommand(ctx context.Context, id string, changes map[string]interface{}) (*model.CommerceCommand, error) {
	databaseCommerceCommand, err := r.CommerceCommandsService.GetById(id)

	if err != nil {
		return nil, err
	}

	if databaseCommerceCommand == nil {
		return nil, &commands.CommerceCommandNotFoundError{}
	}

	helper.ApplyChanges(changes, databaseCommerceCommand)

	err = r.CommerceCommandsService.Update(databaseCommerceCommand)

	if err != nil {
		return nil, err
	}

	return databaseCommerceCommand.ToModel(), nil
}

// CreatePanier is the resolver for the createPanier field.
func (r *mutationResolver) CreatePanier(ctx context.Context, commerceID *string, input model.NewPanier) (*model.Panier, error) {
	user := auth.ForContext(ctx)

	var databaseCommerce *commerces.Commerce

	// On a d'abord besoin de trouver le commerce de l'utilisateur ou celui en paramètre
	if commerceID == nil {
		userDatabaseCommerce, err := r.CommercesService.GetForUser(user.ID.Hex())

		if err != nil {
			return nil, err
		}

		databaseCommerce = userDatabaseCommerce
	} else {
		commerceDatabaseCommerce, err := r.CommercesService.GetById(*commerceID)

		if err != nil {
			return nil, err
		}

		databaseCommerce = commerceDatabaseCommerce
	}

	if databaseCommerce == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	databasePanier, err := r.PaniersService.Create(databaseCommerce.ID, input)

	if err != nil {
		return nil, err
	}

	return databasePanier.ToModel(), nil
}

// UpdatePanier is the resolver for the updatePanier field.
func (r *mutationResolver) UpdatePanier(ctx context.Context, id string, changes map[string]interface{}) (*model.Panier, error) {
	databasePanier, err := r.PaniersService.GetById(id)

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

	err = r.PaniersService.Update(databasePanier, image)

	if err != nil {
		return nil, err
	}

	return databasePanier.ToModel(), nil
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	databaseUsers, err := r.UsersService.GetAllUser()

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

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, id *string) (*model.User, error) {
	if id == nil {
		if auth.ForContext(ctx) == nil {
			return nil, &users.UserNotFoundError{}
		}

		return auth.ForContext(ctx).ToModel(), nil
	}

	databaseUser, err := r.UsersService.GetUserById(*id)

	if err != nil {
		return nil, err
	}

	if databaseUser == nil {
		return nil, &users.UserNotFoundError{}
	}

	return databaseUser.ToModel(), nil
}

// Commerces is the resolver for the commerces field.
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

	databaseCommerces, totalCount, err := r.CommercesService.GetPaginated(decodedCursor, *first, filter)

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
			TotalCount: totalCount,
			Edges:      edges,
			PageInfo:   &model.CommercePageInfo{},
		}, nil
	}

	hasNextPage := !databaseCommerces[itemCount-1].IsLast(r.CommercesService)

	pageInfo := model.CommercePageInfo{
		StartCursor: base64.StdEncoding.EncodeToString([]byte(edges[0].Node.ID)),
		EndCursor:   base64.StdEncoding.EncodeToString([]byte(edges[itemCount-1].Node.ID)),
		HasNextPage: hasNextPage,
	}

	connection := model.CommerceConnection{
		TotalCount: totalCount,
		Edges:      edges[:itemCount],
		PageInfo:   &pageInfo,
	}

	return &connection, nil
}

// Commerce is the resolver for the commerce field.
func (r *queryResolver) Commerce(ctx context.Context, id *string) (*model.Commerce, error) {
	if id != nil {
		databaseCommerce, err := r.CommercesService.GetById(*id)

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

	databaseCommerce, err := r.CommercesService.GetForUser(user.ID.Hex())

	if err != nil {
		return nil, err
	}

	if databaseCommerce == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	return databaseCommerce.ToModel(), nil
}

// Product is the resolver for the product field.
func (r *queryResolver) Product(ctx context.Context, id string) (*model.Product, error) {
	databaseProduct, err := r.ProductsService.GetById(id)

	if err != nil {
		return nil, err
	}

	return databaseProduct.ToModel(), nil
}

// Commands is the resolver for the commands field.
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

	databaseCommands, err := r.CommandsService.GetPaginated(decodedCursor, *first, filter)

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

	hasNextPage := !databaseCommands[itemCount-1].IsLast(r.CommandsService)

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

// CommerceCommands is the resolver for the commerceCommands field.
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

		databaseCommerce, err := r.CommercesService.GetForUser(userIDValue)

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

	databaseCommands, err := r.CommerceCommandsService.GetPaginated(decodedCursor, *first, filter)

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

	hasNextPage := !databaseCommands[itemCount-1].IsLast(r.CommerceCommandsService)

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

// Command is the resolver for the command field.
func (r *queryResolver) Command(ctx context.Context, id string) (*model.Command, error) {
	databaseCommand, err := r.CommandsService.GetById(id)

	if err != nil {
		return nil, err
	}

	if databaseCommand == nil {
		return nil, &commands.CommandNotFoundError{}
	}

	return databaseCommand.ToModel(), nil
}

// AllServicesInfo is the resolver for the allServicesInfo field.
func (r *queryResolver) AllServicesInfo(ctx context.Context) ([]*model.ServiceInfo, error) {
	clickandcollect := servicesinfo.ClickAndCollect()
	paniers := servicesinfo.Paniers()

	return []*model.ServiceInfo{
		&clickandcollect,
		&paniers,
	}, nil
}

// ServiceInfo is the resolver for the serviceInfo field.
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

// Panier is the resolver for the panier field.
func (r *queryResolver) Panier(ctx context.Context, id string) (*model.Panier, error) {
	databasePanier, err := r.PaniersService.GetById(id)

	if err != nil {
		return nil, err
	}

	if databasePanier == nil {
		return nil, &paniers.PanierNotFoundError{}
	}

	return databasePanier.ToModel(), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
