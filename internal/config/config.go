package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var Cfg Config

type Config struct {
	Settings struct {
		AuthSecret string `yaml:"authSecret"`
	} `yaml:"settings"`
	Database struct {
		Name             string `yaml:"name"`
		ConnectionString string `yaml:"connecionString"`
		Collections      struct {
			Users     string `yaml:"users"`
			Commerces string `yaml:"commerces"`
		} `yaml:"collections"`
	} `yaml:"database"`
}

// Initialise les configuration nécessaire dans le code via
// un fichier yaml dont le chemin est fourni par configPath
func Init(configPath string) {
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
