package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/services/paniers"
)

// Products is the resolver for the products field.
func (r *panierResolver) Products(ctx context.Context, obj *model.Panier) ([]*model.PanierProduct, error) {
	return r.PaniersService.GetProducts(obj.ID)
}

// Panier is the resolver for the panier field.
func (r *panierCommandResolver) Panier(ctx context.Context, obj *model.PanierCommand) (*model.Panier, error) {
	databasePanierCommand, err := r.PanierCommandsService.GetById(obj.ID)

	if err != nil {
		return nil, err
	}

	if databasePanierCommand == nil {
		return nil, &paniers.PanierNotFoundError{}
	}

	databasePanier, err := r.PaniersService.GetById(databasePanierCommand.PanierID.Hex())

	if err != nil {
		return nil, err
	}

	if databasePanier == nil {
		return nil, &paniers.PanierNotFoundError{}
	}

	return databasePanier.ToModel(), nil
}

// Panier returns generated.PanierResolver implementation.
func (r *Resolver) Panier() generated.PanierResolver { return &panierResolver{r} }

// PanierCommand returns generated.PanierCommandResolver implementation.
func (r *Resolver) PanierCommand() generated.PanierCommandResolver { return &panierCommandResolver{r} }

type panierResolver struct{ *Resolver }
type panierCommandResolver struct{ *Resolver }
