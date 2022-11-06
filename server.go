package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/graph/resolvers"
	"chemin-du-local.bzh/graphql/internal/auth"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/config"
	"chemin-du-local.bzh/graphql/internal/database"
	"chemin-du-local.bzh/graphql/internal/products"
	"chemin-du-local.bzh/graphql/internal/services/commands"
	"chemin-du-local.bzh/graphql/internal/services/paniers"
	"chemin-du-local.bzh/graphql/internal/users"
	"chemin-du-local.bzh/graphql/pkg/banking"
	"chemin-du-local.bzh/graphql/pkg/mapshandler"
	"chemin-du-local.bzh/graphql/pkg/stripehandler"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/robfig/cron"
	"github.com/rs/cors"
)

const defaultPort = "8082"

func main() {
	// Initialisation du rooter
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Initialisation des services
	commercesService := commerces.NewCommercesService()
	usersService := users.NewUsersService(commercesService)
	productsService := products.NewProductsService()
	paniersService := paniers.NewPaniersService(productsService)
	commandsService := commands.NewCommandsService(usersService)
	commerceCommandsService := commands.NewCommerceCommandsService(usersService, commercesService, commandsService)
	ccCommandsService := commands.NewCCCommandsService(productsService)
	panierCommandsService := commands.NewPanierCommandsService()

	router := chi.NewRouter()
	router.Use(auth.Middleware(usersService))
	// Add CORS middleware around every request
	// See https://github.com/rs/cors for full option listing
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler)

	// Initialisation des config
	configPath := "config.yml"

	if os.Getenv("APP_ENV") == "production" {
		config.InitFromEnv()
	} else {
		config.Init(configPath)
	}

	// Initialisation de la base de données

	shouldDropDb := false
	database.Init(&shouldDropDb)

	// Directives GraphQL
	c := generated.Config{Resolvers: &resolvers.Resolver{
		UsersService:            usersService,
		CommercesService:        commercesService,
		ProductsService:         productsService,
		PaniersService:          paniersService,
		CommandsService:         commandsService,
		CommerceCommandsService: commerceCommandsService,
		CCCommandsService:       ccCommandsService,
		PanierCommandsService:   panierCommandsService,
	}}
	c.Directives.NeedAuthentication = func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		if auth.ForContext(ctx) == nil {
			return nil, &users.UserAccessDenied{}
		}

		return next(ctx)
	}
	c.Directives.HasRole = func(ctx context.Context, obj interface{}, next graphql.Resolver, role model.Role) (interface{}, error) {
		if !auth.ForContext(ctx).HasRole(role) {
			return nil, &users.UserAccessDenied{}
		}

		return next(ctx)
	}

	// Creation des taches régulières
	cron := cron.New()

	cron.AddFunc("0 0 1 * * *", banking.SendBalance)
	cron.Start()

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(c))
	fs := http.FileServer(http.Dir("static"))

	router.Handle("/static/*", http.StripPrefix("/static/", fs))
	router.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)
	router.HandleFunc("/create-setup-intent", func(w http.ResponseWriter, r *http.Request) {
		stripehandler.HanldeCreateSetupIntent(
			w,
			r,
			usersService,
			commercesService,
		)
	})
	router.HandleFunc("/create-order", func(w http.ResponseWriter, r *http.Request) {
		stripehandler.HandleCreateOrder(
			w,
			r,
			usersService,
			commercesService,
			productsService,
			paniersService,
			commandsService,
			commerceCommandsService,
			ccCommandsService,
			panierCommandsService,
		)
	})
	router.HandleFunc("/complete-order", func(w http.ResponseWriter, r *http.Request) {
		stripehandler.HandleCompleteOrder(
			w,
			r,
			usersService,
			commercesService,
			commerceCommandsService,
		)
	})
	router.HandleFunc("/maps/autocomplete", mapshandler.HandleAutocomplete)
	router.HandleFunc("/maps/details", mapshandler.HandlePlaceDetails)

	log.Printf("connect to http://localhost:%s/playground for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
