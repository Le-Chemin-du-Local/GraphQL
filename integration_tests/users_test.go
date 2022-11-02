package integrationtests_tests

import (
	"fmt"
	"testing"

	"chemin-du-local.bzh/graphql/graph"
	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/config"
	"chemin-du-local.bzh/graphql/internal/database"
	"chemin-du-local.bzh/graphql/internal/users"
	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	// Les modèles
	userEmail := "user@me.com"
	userPhone := "0652809335"
	userPassword := "mySuperPassword"

	storekeeperEmail := "commercant@me.com"
	storekeeperPhone := "0876539003"
	storekeeperPassword := "mySecretPassword"
	commerceSiret := "0000000000"
	commerceName := "Mon Super Commerce"
	commerceAddressNumber := "54"
	commerceAddressRoute := "Rue Nationale"
	commerceAddressPostalCode := "35650"
	commerceAddressCity := "Le Rheu"
	commerceLatitude := 48.09312057495117
	commerceLongitude := -1.7779691219329834

	// Initialisation des services
	usersService := users.NewUsersService()

	// Initialisation des config
	configPath := "config_tests.yml"
	config.Init(configPath)

	// Initialisation de la base de données
	shouldDropDb := true
	database.Init(&shouldDropDb)

	// Le client
	resolvers := graph.Resolver{
		UsersService: usersService,
	}
	c := client.New(
		handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})),
	)

	t.Run("register basic user", func(t *testing.T) {
		var resp struct {
			CreateUser struct {
				ID    string `json:"id"`
				Email string `json:"email"`
				Phone string `json:"phone"`
				Role  string `json:"role"`
			}
		}

		q := fmt.Sprintf(`
			mutation {
				createUser(input: {
					email: "%s"
					phone: "%s"
					password: "%s"
				}) {
					id 
					email
					phone 
					role
				}
			}
		`, userEmail, userPhone, userPassword)

		c.MustPost(q, &resp)

		// On vérifie la bonne création en base de données
		dbUser, err := usersService.GetUserByEmail(userEmail)

		require.NoError(t, err)
		require.Equal(t, userEmail, dbUser.Email)
		require.Equal(t, userPhone, dbUser.Phone)
		require.Equal(t, users.USERROLE_USER, dbUser.Role)

		// On vérifie le mot de passe
		require.True(t, users.CheckPasswordHash(userPassword, dbUser.PasswordHash))

		// On vérifie que ça match avec le résultat GraphQL
		require.Equal(t, userEmail, resp.CreateUser.Email)
		require.Equal(t, userPhone, resp.CreateUser.Phone)
	})

	t.Run("register basic user a second time", func(t *testing.T) {
		var resp struct {
			CreateUser struct {
				ID    string `json:"id"`
				Email string `json:"email"`
				Phone string `json:"phone"`
				Role  string `json:"role"`
			}
		}

		q := fmt.Sprintf(`
			mutation {
				createUser(input: {
					email: "%s"
					phone: "%s"
					password: "%s"
					}
				}) {
					id u
					email
					phone 
					role
				}
			}
		`, userEmail, userPhone, userPassword)

		err := c.Post(q, &resp)

		require.Error(t, err)
	})

	t.Run("register a storekeeper", func(t *testing.T) {
		var resp struct {
			CreateUser struct {
				ID       string          `json:"id"`
				Email    string          `json:"email"`
				Phone    string          `json:"phone"`
				Role     string          `json:"role"`
				Commerce *model.Commerce `json:"commerce"`
			}
		}

		q := fmt.Sprintf(`
			mutation {
				createUser(input: {
					email: "%s"
					phone: "%s"
					password: "%s"
					commerce: {
						siret: "%s"
						name: "%s"
						address: {
							number: "%s"
							route: "%s"
							postalCode: "%s"
							city: "%s"
						}
						latitude: %.5f
						longitude: %.5f
						phone: "%s"
						email: "%s"
					}
				}) {
					id 
					email
					phone 
					role
					commerce {
						id
						siret
						name
					}
				}
			}
		`,
			storekeeperEmail,
			storekeeperPhone,
			storekeeperPassword,
			commerceSiret,
			commerceName,
			commerceAddressNumber,
			commerceAddressRoute,
			commerceAddressPostalCode,
			commerceAddressCity,
			commerceLatitude,
			commerceLongitude,
			userPhone,
			userEmail,
		)

		c.MustPost(q, &resp)

		// On vérifie la bonne création en base de données
		dbUser, err := usersService.GetUserByEmail(storekeeperEmail)

		require.NoError(t, err)
		require.Equal(t, storekeeperEmail, dbUser.Email)
		require.Equal(t, storekeeperPhone, dbUser.Phone)
		require.Equal(t, users.USERROLE_USER, dbUser.Role)

		// On vérifie le mot de passe
		require.True(t, users.CheckPasswordHash(storekeeperPassword, dbUser.PasswordHash))

		// On vérifie que ça match avec le résultat GraphQL
		require.Equal(t, storekeeperEmail, resp.CreateUser.Email)
		require.Equal(t, storekeeperPhone, resp.CreateUser.Phone)
		require.Equal(t, commerceName, resp.CreateUser.Commerce.Name)
		require.Equal(t, commerceSiret, resp.CreateUser.Commerce.Siret)
	})
}
