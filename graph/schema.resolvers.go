package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/users"
	"chemin-du-local.bzh/graphql/pkg/jwt"
)

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

	token, err := jwt.GenerateToken(user.ID.Hex(), user.Role)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (r *mutationResolver) CreateCommerce(ctx context.Context, input model.NewCommerce) (*model.Commerce, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Commerces(ctx context.Context) ([]*model.Commerce, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Commerce(ctx context.Context, id string) (*model.Commerce, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
