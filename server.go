package main

import (
	"log"
	"net/http"
	"os"

	"chemin-du-local.bzh/graphql/graph"
	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/internal/auth"
	"chemin-du-local.bzh/graphql/internal/config"
	"chemin-du-local.bzh/graphql/internal/database"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
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

	// Initialisation des config
	configPath := "config.yml"
	if os.Getenv("APP_ENV") == "production" {
		configPath = "config.production.yml"
	}

	config.Init(configPath)

	// Initialisation de la base de données
	database.Init()

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
