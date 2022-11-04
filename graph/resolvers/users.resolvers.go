package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/registeredpaymentmethod"
	"chemin-du-local.bzh/graphql/internal/users"
	"chemin-du-local.bzh/graphql/pkg/stripehandler"
)

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

	databaseUser, err := r.UsersService.GetUserById(obj.ID)

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
	databaseUser, err := r.UsersService.GetUserById(obj.ID)

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

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
