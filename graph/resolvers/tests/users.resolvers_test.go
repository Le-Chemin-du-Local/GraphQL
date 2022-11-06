package resolver_test

import (
	"testing"
	"time"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/graph/resolvers"
	"chemin-du-local.bzh/graphql/internal/address"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/mocks"
	"chemin-du-local.bzh/graphql/internal/users"
	"chemin-du-local.bzh/graphql/pkg/geojson"
	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserResolverCommerce(t *testing.T) {
	// Les modèles
	userID := primitive.NewObjectID()
	userEmail := "user@me.com"
	userPhone := "0652809335"
	userRole := users.USERROLE_USER
	userPassword := "mySuperPassword"
	userPasswordHash, _ := users.HashPassword(userPassword)
	userCreationDate := time.Date(2022, 10, 27, 15, 0, 0, 0, time.Local)
	user := users.User{
		ID:           userID,
		CreatedAt:    userCreationDate,
		Email:        userEmail,
		Phone:        userPhone,
		Role:         userRole,
		PasswordHash: userPasswordHash,
	}

	storekeeperID := primitive.NewObjectID()
	storekeeperEmail := "commercant@me.com"
	storekeeperPhone := "0876539003"
	storekeeperRole := users.USERROLE_STOREKEEPER
	storekeeperPassword := "mySecretPassword"
	storekeeperPasswordHash, _ := users.HashPassword(storekeeperPassword)
	storekeeperCreationDate := time.Date(2022, 10, 27, 15, 0, 0, 0, time.Local)
	storekeeper := users.User{
		ID:           storekeeperID,
		CreatedAt:    storekeeperCreationDate,
		Email:        storekeeperEmail,
		Phone:        storekeeperPhone,
		Role:         storekeeperRole,
		PasswordHash: storekeeperPasswordHash,
	}

	commerceID := primitive.NewObjectID()
	commerceSiret := "0000000000"
	commerceName := "Mon Super Commerce"
	commerceAddressNumber := "54"
	commerceAddressRoute := "Rue Nationale"
	commerceAddressPostalCode := "35650"
	commerceAddressCity := "Le Rheu"
	commerceLatitude := 48.09312057495117
	commerceLongitude := -1.7779691219329834
	commerce := commerces.Commerce{
		ID:            commerceID,
		StorekeeperID: storekeeperID,
		Siret:         commerceSiret,
		Name:          commerceName,
		Address: address.Address{
			ID:         primitive.NewObjectID(),
			Number:     &commerceAddressNumber,
			Route:      &commerceAddressRoute,
			PostalCode: &commerceAddressPostalCode,
			City:       &commerceAddressCity,
		},
		AddressGeo: geojson.GeoJSON{
			Type:        "Point",
			Coordinates: []float64{commerceLatitude, commerceLongitude},
		},
	}

	testUserService := new(mocks.UsersService)
	testCommercesService := new(mocks.CommercesService)
	resolvers := resolvers.Resolver{
		UsersService:     testUserService,
		CommercesService: testCommercesService,
	}

	c := client.New(
		handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})),
	)

	userNotFoundErr := &users.UserNotFoundError{}
	testUserService.On("GetUserById", userID.Hex()).Return(&user, nil)
	testUserService.On("GetUserById", storekeeperID.Hex()).Return(&storekeeper, nil)
	testUserService.On("GetUserById", mock.AnythingOfType("string")).Return(nil, userNotFoundErr)

	testCommercesService.On("GetForUser", storekeeper.ID.Hex()).Return(&commerce, nil)
	testCommercesService.On("GetForUser", mock.AnythingOfType("string")).Return(nil, nil)

	q := `
		query GetUser($id: ID) {
			user(id: $id) {
				id
				email
				phone
				role
				commerce {
					id
					siret
					name
					address {
						number
						route
						postalCode
						city
					}
				}
			}
		}
	`

	t.Run("query user with commerce", func(t *testing.T) {
		var resp struct {
			User struct {
				ID       string          `json:"id"`
				Email    string          `json:"email"`
				Phone    string          `json:"phone"`
				Role     string          `json:"role"`
				Commerce *model.Commerce `json:"commerce"`
			}
		}

		c.MustPost(
			q,
			&resp,
			client.Var("id", storekeeperID),
		)

		require.Equal(t, storekeeperID.Hex(), resp.User.ID)
		require.Equal(t, storekeeperEmail, resp.User.Email)
		require.Equal(t, storekeeperPhone, resp.User.Phone)
		require.Equal(t, storekeeperRole, resp.User.Role)
		require.Equal(t, commerceID.Hex(), resp.User.Commerce.ID)
		require.Equal(t, commerceSiret, resp.User.Commerce.Siret)
		require.Equal(t, commerceName, resp.User.Commerce.Name)
		require.Equal(t, &commerceAddressNumber, resp.User.Commerce.Address.Number)
		require.Equal(t, &commerceAddressRoute, resp.User.Commerce.Address.Route)
		require.Equal(t, &commerceAddressPostalCode, resp.User.Commerce.Address.PostalCode)
		require.Equal(t, &commerceAddressCity, resp.User.Commerce.Address.City)
	})

	t.Run("query user without commerce", func(t *testing.T) {
		var resp struct {
			User struct {
				ID       string          `json:"id"`
				Email    string          `json:"email"`
				Phone    string          `json:"phone"`
				Role     string          `json:"role"`
				Commerce *model.Commerce `json:"commerce"`
			}
		}

		c.MustPost(
			q,
			&resp,
			client.Var("id", userID),
		)

		require.Equal(t, userID.Hex(), resp.User.ID)
		require.Equal(t, userEmail, resp.User.Email)
		require.Equal(t, userPhone, resp.User.Phone)
		require.Equal(t, userRole, resp.User.Role)
		require.Nil(t, resp.User.Commerce)
	})

}
