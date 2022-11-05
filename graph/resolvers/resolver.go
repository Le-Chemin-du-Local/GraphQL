package resolvers

import (
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/products"
	"chemin-du-local.bzh/graphql/internal/services/commands"
	"chemin-du-local.bzh/graphql/internal/services/paniers"
	"chemin-du-local.bzh/graphql/internal/users"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UsersService            users.UsersService
	CommercesService        commerces.CommercesService
	ProductsService         products.ProductsService
	PaniersService          paniers.PaniersService
	CommandsService         commands.CommandsService
	CommerceCommandsService commands.CommerceCommandsService
	CCCommandsService       commands.CCCommandsService
	PanierCommandsService   commands.PanierCommandsService
}
