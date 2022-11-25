package database

import (
	"context"
	"fmt"
	"log"

	"chemin-du-local.bzh/graphql/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoContext = context.TODO()

var CollectionUsers *mongo.Collection
var CollectionCommerces *mongo.Collection
var CollectionProducts *mongo.Collection
var CollectionCommand *mongo.Collection
var CollectionCommerceCommand *mongo.Collection
var CollectionCCCommand *mongo.Collection
var CollectionPanierCommands *mongo.Collection
var CollectionPaniers *mongo.Collection

// Initialise la base de données à partir des informations données
// dans la configuration
func Init(shouldDrop *bool) {
	fmt.Println("Initialising Database connection...")
	// Toutes les informations sont récupérés dans la configuration
	// pour éviter de se retrouver avec des identifiants secrets présent
	// dans le code
	connectionString := config.Cfg.Database.ConnectionString

	clientOptions := options.Client().ApplyURI(connectionString)
	client, err := mongo.Connect(MongoContext, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Il est préférable de ping la base de données pour vérifier
	// qu'on peut bien s'y connecter
	err = client.Ping(MongoContext, nil)

	if err != nil {
		log.Fatal(err)
	}

	// On affecte les collections pour pouvoir y accéder plus facilement
	// dans l'API
	databaseName := config.Cfg.Database.Name
	usersCollectionName := config.Cfg.Database.Collections.Users
	commercesCollectionName := config.Cfg.Database.Collections.Commerces
	produtsCollectionName := config.Cfg.Database.Collections.Products
	commandeCollectionName := config.Cfg.Database.Collections.Commands
	commerceCommandeCollectionName := config.Cfg.Database.Collections.CommerceCommands
	cccommandeCollectionName := config.Cfg.Database.Collections.CCCommands
	panierCommandsCollectionName := config.Cfg.Database.Collections.PanierCommands
	panierCollectionName := config.Cfg.Database.Collections.Paniers

	CollectionUsers = client.Database(databaseName).Collection(usersCollectionName)
	CollectionCommerces = client.Database(databaseName).Collection(commercesCollectionName)
	CollectionProducts = client.Database(databaseName).Collection(produtsCollectionName)
	CollectionCommand = client.Database(databaseName).Collection(commandeCollectionName)
	CollectionCommerceCommand = client.Database(databaseName).Collection(commerceCommandeCollectionName)
	CollectionCCCommand = client.Database(databaseName).Collection(cccommandeCollectionName)
	CollectionPanierCommands = client.Database(databaseName).Collection(panierCommandsCollectionName)
	CollectionPaniers = client.Database(databaseName).Collection(panierCollectionName)

	// Si on veut vider la bdd à l'initialisation, on le fait
	if shouldDrop != nil && *shouldDrop {
		client.Database(databaseName).Drop(MongoContext)
	}
}
