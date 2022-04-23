package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var Cfg Config

type Config struct {
	Settings struct {
		AuthSecret string `yaml:"authSecret"`
	} `yaml:"settings"`
	Stripe struct {
		Key string `yaml:"key"`
	} `yaml:"stripe"`
	Paths struct {
		Static string `yaml:"static"`
	} `yaml:"paths"`
	Database struct {
		Name             string `yaml:"name"`
		ConnectionString string `yaml:"connectionString"`
		Collections      struct {
			Users          string `yaml:"users"`
			Commerces      string `yaml:"commerces"`
			Products       string `yaml:"products"`
			CCCommands     string `yaml:"cccommands"`
			Paniers        string `yaml:"paniers"`
			PanierCommands string `yaml:"paniercommands"`
		} `yaml:"collections"`
	} `yaml:"database"`
}

// Initialise les configuration nécessaire dans le code via
// un fichier yaml dont le chemin est fourni par configPath
func Init(configPath string) {
	fmt.Println("initializing config from path")

	file, err := os.Open(configPath)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	// On doit décoder le fichier de config qui est en YAML
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&Cfg)

	if err != nil {
		log.Fatal(err)
	}
}

func InitFromEnv() {
	fmt.Println("initializing config from env")

	Cfg.Settings.AuthSecret = os.Getenv("AUTH_SECRET")

	Cfg.Stripe.Key = os.Getenv("STRIPE_KEY")

	Cfg.Paths.Static = os.Getenv("PATH_STATIC")

	Cfg.Database.Name = os.Getenv("DATABASE_NAME")
	Cfg.Database.ConnectionString = os.Getenv("CONNECTION_STRING")
	Cfg.Database.Collections.Users = os.Getenv("COLLECTION_USERS")
	Cfg.Database.Collections.Commerces = os.Getenv("COLLECTION_COMMERCES")
	Cfg.Database.Collections.Products = os.Getenv("COLLECTION_PRODUCTS")
	Cfg.Database.Collections.CCCommands = os.Getenv("COLLECTION_CCCOMMANDS")
	Cfg.Database.Collections.Paniers = os.Getenv("COLLECTION_PANIERS")
	Cfg.Database.Collections.PanierCommands = os.Getenv("COLLECTION_PANIERCOMMANDS")
}
