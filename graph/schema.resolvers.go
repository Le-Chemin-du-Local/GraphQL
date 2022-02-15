package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/auth"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/users"
	"chemin-du-local.bzh/graphql/pkg/jwt"
)

func (r *commerceResolver) Storekeeper(ctx context.Context, obj *model.Commerce) (*model.User, error) {
	storekeeper, err := users.GetUserById(obj.StorekeeperID)

	if err != nil {
		return nil, err
	}

	return storekeeper.ToModel(), nil
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

func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	databaseUser, err := users.GetUserById(id)

	if err != nil {
		return nil, err
	}

	if databaseUser == nil {
		return nil, &users.UserNotFoundError{}
	}

	return databaseUser.ToModel(), nil
}

func (r *queryResolver) Commerces(ctx context.Context) ([]*model.Commerce, error) {
	databaseCommerces, err := commerces.GetAll()

	if err != nil {
		return nil, err
	}

	commerces := []*model.Commerce{}

	for _, datadatabaseCommerce := range databaseCommerces {
		commerce := datadatabaseCommerce.ToModel()

		commerces = append(commerces, commerce)
	}

	return commerces, nil
}

func (r *queryResolver) Commerce(ctx context.Context, id string) (*model.Commerce, error) {
	databaseCommerce, err := commerces.GetById(id)

	if err != nil {
		return nil, err
	}

	if databaseCommerce == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	return databaseCommerce.ToModel(), nil
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
