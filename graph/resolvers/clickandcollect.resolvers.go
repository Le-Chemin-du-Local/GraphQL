package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
)

// Products is the resolver for the products field.
func (r *cCCommandResolver) Products(ctx context.Context, obj *model.CCCommand) ([]*model.CCProduct, error) {
	return r.CCCommandsService.GetProducts(obj.ID)
}

// CCCommand returns generated.CCCommandResolver implementation.
func (r *Resolver) CCCommand() generated.CCCommandResolver { return &cCCommandResolver{r} }

type cCCommandResolver struct{ *Resolver }
