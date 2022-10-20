package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

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
	"chemin-du-local.bzh/graphql/pkg/utils"
	"github.com/99designs/gqlgen/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Products is the resolver for the products field.
func (r *cCCommandResolver) Products(ctx context.Context, obj *model.CCCommand) ([]*model.CCProduct, error) {
	return clickandcollect.GetProducts(obj.ID)
}

// User is the resolver for the user field.
func (r *commandResolver) User(ctx context.Context, obj *model.Command) (*model.User, error) {
	user, err := commands.GetUser(obj.ID)

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
	storekeeper, err := users.GetUserById(obj.StorekeeperID)

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
	user, err := commands.CommerceGetUser(obj.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateUser is the resolver for the createUser field.
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

// Login is the resolver for the login field.
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

// UpdateUser is the resolver for the updateUser field.
func (r *mutationResolver) UpdateUser(ctx context.Context, id *string, input map[string]interface{}) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

// CreateCommerce is the resolver for the createCommerce field.
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

// UpdateCommerce is the resolver for the updateCommerce field.
func (r *mutationResolver) UpdateCommerce(ctx context.Context, id string, changes map[string]interface{}) (*model.Commerce, error) {
	databaseCommerce, err := commerces.GetById(id)

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
				productsOfCommerce, err := products.GetForCommerce(databaseCommerce.ID.Hex())

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

	err = commerces.Update(databaseCommerce, image, profilePicture)

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

	// Si le commerçant a souscrit au Click&Collect, le produit doit automatiquement y être ajouté
	databaseCommerce, err := commerces.GetForUser(user.ID.Hex())

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

	commerces.Update(databaseCommerce, nil, nil)

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

// UpdateProduct is the resolver for the updateProduct field.
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

// UpdateProducts is the resolver for the updateProducts field.
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

// UpdateCommerceCommand is the resolver for the updateCommerceCommand field.
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

// CreatePanier is the resolver for the createPanier field.
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

// UpdatePanier is the resolver for the updatePanier field.
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

// Users is the resolver for the users field.
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

// User is the resolver for the user field.
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

	databaseCommerces, totalCount, err := commerces.GetPaginated(decodedCursor, *first, filter)

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

	hasNextPage := !databaseCommerces[itemCount-1].IsLast()

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

// Product is the resolver for the product field.
func (r *queryResolver) Product(ctx context.Context, id string) (*model.Product, error) {
	databaseProduct, err := products.GetById(id)

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

// Command is the resolver for the command field.
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
	databasePanier, err := paniers.GetById(id)

	if err != nil {
		return nil, err
	}

	if databasePanier == nil {
		return nil, &paniers.PanierNotFoundError{}
	}

	return databasePanier.ToModel(), nil
}

// Commerce is the resolver for the commerce field.
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

// Basket is the resolver for the basket field.
func (r *userResolver) Basket(ctx context.Context, obj *model.User) (*model.Basket, error) {
	panic(fmt.Errorf("not implemented"))
}

// RegisteredPaymentMethods is the resolver for the registeredPaymentMethods field.
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

// DefaultPaymentMethod is the resolver for the defaultPaymentMethod field.
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
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *commerceResolver) AddressDetailed(ctx context.Context, obj *model.Commerce) (*model.Address, error) {
	panic(fmt.Errorf("not implemented"))
}
