package resolver_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/graph/resolvers"
	"chemin-du-local.bzh/graphql/internal/auth"
	"chemin-du-local.bzh/graphql/internal/mocks"
	"chemin-du-local.bzh/graphql/internal/users"
	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func addContext(user *users.User) client.Option {
	return func(bd *client.Request) {
		ctx := context.WithValue(bd.HTTP.Context(), auth.UserCtxKey, user)
		bd.HTTP = bd.HTTP.WithContext(ctx)
	}
}

// Test sur la création d'un utilisateur
func TestMutationResolver_CreateUser(t *testing.T) {
	// Les modèles
	existingUserID := primitive.NewObjectID()
	existingUserEmail := "test@user.com"
	existingUserPhone := "0652809335"
	existingUserRole := users.USERROLE_USER
	existingUserPassword, _ := users.HashPassword("mySuperPassword")
	existingUserCreationDate := time.Date(2022, 10, 27, 15, 0, 0, 0, time.Local)
	existingUser := users.User{
		ID:           existingUserID,
		CreatedAt:    existingUserCreationDate,
		Email:        existingUserEmail,
		Phone:        existingUserPhone,
		Role:         existingUserRole,
		PasswordHash: existingUserPassword,
	}

	expectedUserID := primitive.NewObjectID()
	expectedUserEmail := "register@user.com"
	expectedUserPhone := "0652809335"
	expectedUserRole := users.USERROLE_USER
	expectedUserPassword, _ := users.HashPassword("mySuperPassword")
	expectedUserCreationDate := time.Now()
	expectedUser := users.User{
		ID:           expectedUserID,
		CreatedAt:    expectedUserCreationDate,
		Email:        expectedUserEmail,
		Phone:        expectedUserPhone,
		Role:         expectedUserRole,
		PasswordHash: expectedUserPassword,
	}

	testUsersService := new(mocks.UsersService)
	resolvers := resolvers.Resolver{UsersService: testUsersService}
	c := client.New(
		handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})),
	)

	testUsersService.On("GetUserByEmail", existingUser.Email).Return(existingUser, nil)
	testUsersService.On("GetUserByEmail", mock.AnythingOfType("string")).Return(nil, nil)
	testUsersService.On("Create", mock.AnythingOfType("model.NewUser")).Return(&expectedUser, nil)

	// Test la création d'un utilisateur
	t.Run("create a new user", func(t *testing.T) {
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
					password: "mySuperPassword"
				}) {
					id 
					email
					phone 
					role
				}
			}
		`, expectedUserEmail, expectedUserPhone)

		c.MustPost(q, &resp)
		require.Equal(t, expectedUserID.Hex(), resp.CreateUser.ID)
		require.Equal(t, expectedUserEmail, resp.CreateUser.Email)
		require.Equal(t, expectedUserPhone, resp.CreateUser.Phone)
		require.Equal(t, expectedUserRole, resp.CreateUser.Role)
	})

	// Test la création d'un utilisateur avec un mail en majuscule
	t.Run("create a new user with capital email", func(t *testing.T) {
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
					password: "mySuperPassword"
				}) {
					id 
					email
					phone 
					role
				}
			}
		`, strings.ToUpper(expectedUserEmail), expectedUserPhone)

		c.MustPost(q, &resp)
		require.Equal(t, expectedUserID.Hex(), resp.CreateUser.ID)
		require.Equal(t, expectedUserEmail, resp.CreateUser.Email)
		require.Equal(t, expectedUserPhone, resp.CreateUser.Phone)
		require.Equal(t, expectedUserRole, resp.CreateUser.Role)
	})

	// Test la création d'un utilisateur
	t.Run("create an existing user", func(t *testing.T) {
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
					password: "mySuperPassword"
				}) {
					id 
					email
					phone 
					role
				}
			}
		`, existingUserEmail, existingUserPhone)

		err := c.Post(q, &resp)

		require.Error(t, err)
	})

	// Test la création d'un utilisateur
	t.Run("create an existing user with capital email", func(t *testing.T) {
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
					password: "mySuperPassword"
				}) {
					id 
					email
					phone 
					role
				}
			}
		`, strings.ToUpper(existingUserEmail), existingUserPhone)

		err := c.Post(q, &resp)

		require.Error(t, err)
	})

	// Test la création d'un utilisateur invalide
	t.Run("create a user with invalid email", func(t *testing.T) {
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
					password: "mySuperPassword"
				}) {
					id 
					email
					phone 
					role
				}
			}
		`, "invalide email", existingUserPhone)

		err := c.Post(q, &resp)

		require.Error(t, err)
	})
}

// Test sur la connexion
func TestMutationResolver_Login(t *testing.T) {
	userID := primitive.NewObjectID()
	userEmail := "test@user.com"
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

	testUsersService := new(mocks.UsersService)
	resolvers := resolvers.Resolver{UsersService: testUsersService}
	c := client.New(
		handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})),
	)

	testUsersService.On("GetUserByEmail", userEmail).Return(&user, nil)
	testUsersService.On("Authenticate", model.Login{
		Email:    userEmail,
		Password: userPassword,
	}).Return(true)
	testUsersService.On("Authenticate", mock.AnythingOfType("model.Login")).Return(false)

	// Tests de l'authentication réussi
	t.Run("login user", func(t *testing.T) {
		var resp struct {
			Login string
		}

		q := fmt.Sprintf(`
			mutation {
				login(input: {
					email: "%s"
					password: "%s"
				})
			}
		`, userEmail, userPassword)

		c.MustPost(q, &resp)
	})

	// Tests de l'authentication mauvais mot de passe
	t.Run("login user wrong password", func(t *testing.T) {
		var resp struct {
			Login string
		}

		q := fmt.Sprintf(`
			mutation {
				login(input: {
					email: "%s"
					password: "%s"
				})
			}
		`, userEmail, "wrong password")

		err := c.Post(q, &resp)
		require.Error(t, err)
	})

	// Tests de l'authentication mauvais email
	t.Run("login user wrong email", func(t *testing.T) {
		var resp struct {
			Login string
		}

		q := fmt.Sprintf(`
			mutation {
				login(input: {
					email: "%s"
					password: "%s"
				})
			}
		`, "wrong email", userPassword)

		err := c.Post(q, &resp)
		require.Error(t, err)
	})
}

// Test sur la récupération de plusieurs utilisateurs
func TestQueryResolver_Users(t *testing.T) {
	// Les modèles
	user1ID := primitive.NewObjectID()
	user1Email := "test1@user.com"
	user1Phone := "0652809335"
	user1Role := users.USERROLE_USER
	user1Password, _ := users.HashPassword("mySuperPassword")
	user1CreationDate := time.Date(2022, 10, 27, 15, 0, 0, 0, time.Local)

	user2ID := primitive.NewObjectID()
	user2Email := "test2@user.com"
	user2Phone := "0652809335"
	user2Role := users.USERROLE_STOREKEEPER
	user2Password, _ := users.HashPassword("mySuperPassword")
	user2CreationDate := time.Date(2022, 10, 27, 15, 0, 0, 0, time.Local)

	user1 := users.User{
		ID:           user1ID,
		CreatedAt:    user1CreationDate,
		Email:        user1Email,
		Phone:        user1Phone,
		Role:         user1Role,
		PasswordHash: user1Password,
	}

	user2 := users.User{
		ID:           user2ID,
		CreatedAt:    user2CreationDate,
		Email:        user2Email,
		Phone:        user2Phone,
		Role:         user2Role,
		PasswordHash: user2Password,
	}

	// Les tests

	t.Run("query all users", func(t *testing.T) {
		testUsersService := new(mocks.UsersService)
		resolvers := resolvers.Resolver{UsersService: testUsersService}
		c := client.New(
			handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})),
		)

		testUsersService.On("GetAllUser").Return([]users.User{
			user1,
			user2,
		}, nil)

		var resp struct {
			Users []struct {
				ID    string `json:"id"`
				Email string `json:"email"`
				Phone string `json:"phone"`
				Role  string `json:"role"`
			}
		}

		q := `
			query {
				users {
					id
					email
					phone
					role
				}
			}
		`

		c.MustPost(
			q,
			&resp,
		)

		testUsersService.AssertCalled(t, "GetAllUser")
		require.Len(t, resp.Users, 2)

		require.Equal(t, user1ID.Hex(), resp.Users[0].ID)
		require.Equal(t, user1Email, resp.Users[0].Email)
		require.Equal(t, user1Phone, resp.Users[0].Phone)
		require.Equal(t, user1Role, resp.Users[0].Role)

		require.Equal(t, user2ID.Hex(), resp.Users[1].ID)
		require.Equal(t, user2Email, resp.Users[1].Email)
		require.Equal(t, user2Phone, resp.Users[1].Phone)
		require.Equal(t, user2Role, resp.Users[1].Role)
	})
}

// Tests sur la récupération d'un utilisateur spécifique
func TestQueryResolver_User(t *testing.T) {
	// Les modèles
	userID := primitive.NewObjectID()
	userEmail := "test@user.com"
	userPhone := "0652809335"
	userRole := users.USERROLE_USER
	userPassword, _ := users.HashPassword("mySuperPassword")
	userCreationDate := time.Date(2022, 10, 27, 15, 0, 0, 0, time.Local)
	user := users.User{
		ID:           userID,
		CreatedAt:    userCreationDate,
		Email:        userEmail,
		Phone:        userPhone,
		Role:         userRole,
		PasswordHash: userPassword,
	}

	// Les tests
	t.Run("query authenticated user", func(t *testing.T) {
		var resp struct {
			User struct {
				ID    string `json:"id"`
				Email string `json:"email"`
				Phone string `json:"phone"`
				Role  string `json:"role"`
			}
		}

		testUsersService := new(mocks.UsersService)
		resolvers := resolvers.Resolver{UsersService: testUsersService}
		c := client.New(
			handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})),
			addContext(&user),
		)

		testUsersService.On("GetUserById", userID.Hex()).Return(&user, nil)

		q := `
			query GetUser($id: ID) {
				user(id: $id) {
					id
					email
					phone
					role
				}
			}
		`

		c.MustPost(
			q,
			&resp,
			client.Var("id", nil),
		)

		testUsersService.AssertNotCalled(t, "GetUserById", userID.Hex())
		require.Equal(t, userID.Hex(), resp.User.ID)
		require.Equal(t, userEmail, resp.User.Email)
		require.Equal(t, userPhone, resp.User.Phone)
		require.Equal(t, userRole, resp.User.Role)
	})

	t.Run("query authenticated user without authentication", func(t *testing.T) {
		var resp struct {
			User struct {
				ID    string `json:"id"`
				Email string `json:"email"`
				Phone string `json:"phone"`
				Role  string `json:"role"`
			}
		}

		testUsersService := new(mocks.UsersService)
		resolvers := resolvers.Resolver{UsersService: testUsersService}
		c := client.New(
			handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})),
		)

		testUsersService.On("GetUserById", userID.Hex()).Return(&user, nil)

		q := `
			query GetUser($id: ID) {
				user(id: $id) {
					id
					email
					phone
					role
				}
			}
		`

		err := c.Post(
			q,
			&resp,
			client.Var("id", nil),
		)

		testUsersService.AssertNotCalled(t, "GetUserById", nil)
		require.Error(t, err)
	})

	t.Run("query a user by its id", func(t *testing.T) {
		var resp struct {
			User struct {
				ID    string `json:"id"`
				Email string `json:"email"`
				Phone string `json:"phone"`
				Role  string `json:"role"`
			}
		}

		testUsersService := new(mocks.UsersService)
		resolvers := resolvers.Resolver{UsersService: testUsersService}
		c := client.New(
			handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})),
		)

		testUsersService.On("GetUserById", userID.Hex()).Return(&user, nil)

		q := `
			query GetUser($id: ID) {
				user(id: $id) {
					id
					email
					phone
					role
				}
			}
		`

		c.MustPost(
			q,
			&resp,
			client.Var("id", userID.Hex()),
		)

		testUsersService.AssertCalled(t, "GetUserById", userID.Hex())
		require.Equal(t, userID.Hex(), resp.User.ID)
		require.Equal(t, userEmail, resp.User.Email)
		require.Equal(t, userPhone, resp.User.Phone)
		require.Equal(t, userRole, resp.User.Role)
	})

	t.Run("query with authentified by its id", func(t *testing.T) {
		var resp struct {
			User struct {
				ID    string `json:"id"`
				Email string `json:"email"`
				Phone string `json:"phone"`
				Role  string `json:"role"`
			}
		}

		testUsersService := new(mocks.UsersService)
		resolvers := resolvers.Resolver{UsersService: testUsersService}
		c := client.New(
			handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})),
			addContext(&user),
		)

		testUsersService.On("GetUserById", userID.Hex()).Return(&user, nil)

		q := `
			query GetUser($id: ID) {
				user(id: $id) {
					id
					email
					phone
					role
				}
			}
		`

		c.MustPost(
			q,
			&resp,
			client.Var("id", user.ID.Hex()),
		)

		testUsersService.AssertCalled(t, "GetUserById", userID.Hex())
		require.Equal(t, userID.Hex(), resp.User.ID)
		require.Equal(t, userEmail, resp.User.Email)
		require.Equal(t, userPhone, resp.User.Phone)
		require.Equal(t, userRole, resp.User.Role)
	})

	t.Run("query non existant user", func(t *testing.T) {
		var resp struct {
			User struct {
				ID    string `json:"id"`
				Email string `json:"email"`
				Phone string `json:"phone"`
				Role  string `json:"role"`
			}
		}

		testUsersService := new(mocks.UsersService)
		resolvers := resolvers.Resolver{UsersService: testUsersService}
		c := client.New(
			handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})),
		)

		testUsersService.On("GetUserById", userID.Hex()).Return(&user, nil)

		q := `
			query GetUser($id: ID) {
				user(id: $id) {
					id
					email
					phone
					role
				}
			}
		`

		err := c.Post(
			q,
			&resp,
			client.Var("id", primitive.NewObjectID().Hex()),
		)

		testUsersService.AssertNotCalled(t, "GetUserById", nil)
		require.Error(t, err)
	})

	t.Run("query with authentified non existant user", func(t *testing.T) {
		var resp struct {
			User struct {
				ID    string `json:"id"`
				Email string `json:"email"`
				Phone string `json:"phone"`
				Role  string `json:"role"`
			}
		}

		testUsersService := new(mocks.UsersService)
		resolvers := resolvers.Resolver{UsersService: testUsersService}
		c := client.New(
			handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers})),
			addContext(&user),
		)

		testUsersService.On("GetUserById", userID.Hex()).Return(&user, nil)

		q := `
			query GetUser($id: ID) {
				user(id: $id) {
					id
					email
					phone
					role
				}
			}
		`

		err := c.Post(
			q,
			&resp,
			client.Var("id", primitive.NewObjectID().Hex()),
		)

		testUsersService.AssertNotCalled(t, "GetUserById", nil)
		require.Error(t, err)
	})
}
