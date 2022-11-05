package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
)

// User is the resolver for the user field.
func (r *commandResolver) User(ctx context.Context, obj *model.Command) (*model.User, error) {
	user, err := r.CommandsService.GetUser(obj.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Commerces is the resolver for the commerces field.
func (r *commandResolver) Commerces(ctx context.Context, obj *model.Command) ([]*model.CommerceCommand, error) {
	databaseCommerceCommands, err := r.CommerceCommandsService.GetForCommand(obj.ID)

	if err != nil {
		return nil, err
	}

	commerceCommands := []*model.CommerceCommand{}

	for _, databaseCommerceCommand := range databaseCommerceCommands {
		commerceCommands = append(commerceCommands, databaseCommerceCommand.ToModel())
	}

	return commerceCommands, nil
}

// Commerce is the resolver for the commerce field.
func (r *commerceCommandResolver) Commerce(ctx context.Context, obj *model.CommerceCommand) (*model.Commerce, error) {
	commerce, err := r.CommerceCommandsService.GetCommerce(obj.ID)

	if err != nil {
		return nil, err
	}

	return commerce, nil
}

// Cccommands is the resolver for the cccommands field.
func (r *commerceCommandResolver) Cccommands(ctx context.Context, obj *model.CommerceCommand) ([]*model.CCCommand, error) {
	databaseCCCommands, err := r.CCCommandsService.GetForCommmerceCommand(obj.ID)

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
	databasePanierCommands, err := r.PanierCommandsService.GetForCommerceCommand(obj.ID)

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
	user, err := r.CommerceCommandsService.GetUser(obj.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Command returns generated.CommandResolver implementation.
func (r *Resolver) Command() generated.CommandResolver { return &commandResolver{r} }

// CommerceCommand returns generated.CommerceCommandResolver implementation.
func (r *Resolver) CommerceCommand() generated.CommerceCommandResolver {
	return &commerceCommandResolver{r}
}

type commandResolver struct{ *Resolver }
type commerceCommandResolver struct{ *Resolver }
