package main

import (
	"log"
	"net/http"
	"os"

	"chemin-du-local.bzh/graphql/graph"
	"chemin-du-local.bzh/graphql/graph/generated"
	"chemin-du-local.bzh/graphql/internal/config"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	configPath := "config.yml"
	if os.Getenv("APP_ENV") == "production" {
		configPath = "config.production.yml"
	}

	config.Init(configPath)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
