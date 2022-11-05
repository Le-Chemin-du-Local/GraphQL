package resolvers

import (
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/services/commands"
	"chemin-du-local.bzh/graphql/internal/users"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UsersService            users.UsersService
	CommercesService        commerces.CommercesService
	CommandsService         commands.CommandsService
	CommerceCommandsService commands.CommerceCommandsService
}
