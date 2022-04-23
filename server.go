package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"chemin-du-local.bzh/graphql/graph"
	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/auth"
	"chemin-du-local.bzh/graphql/internal/config"
	"chemin-du-local.bzh/graphql/internal/database"
	"chemin-du-local.bzh/graphql/internal/users"
	"chemin-du-local.bzh/graphql/pkg/stripehandler"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/rs/cors"
)

const defaultPort = "8082"

func main() {
	// Initialisation du rooter
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	router := chi.NewRouter()
	router.Use(auth.Middleware())
	// Add CORS middleware around every request
	// See https://github.com/rs/cors for full option listing
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            true,
	}).Handler)

	// Initialisation des config
	configPath := "config.yml"

	if os.Getenv("APP_ENV") == "production" {
		config.InitFromEnv()
	} else {
		config.Init(configPath)
	}

	// Initialisation de la base de données
	database.Init()

	// Directives GraphQL
	c := generated.Config{Resolvers: &graph.Resolver{}}
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

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(c))
	fs := http.FileServer(http.Dir("static"))

	router.Handle("/static/*", http.StripPrefix("/static/", fs))
	router.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)
	router.HandleFunc("/create-payment-intent", stripehandler.HandleCreatePaymentIntent)

	log.Printf("connect to http://localhost:%s/playground for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
