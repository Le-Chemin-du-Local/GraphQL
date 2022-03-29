package model

// Ici on utilise un modèle différent de celui généré car on
// ne veut pas générer une requette pour certaines choses comme
// le commerçant, les produits ou les commentaires temps
// qu'ils ne sont pas demandés.

type Commerce struct {
	ID              string   `json:"id"`
	StorekeeperID   string   `json:"storekeeper"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	StorekeeperWord string   `json:"storekeeperWord"`
	Address         string   `json:"address"`
	Phone           string   `json:"phone"`
	Email           string   `json:"email"`
	Facebook        *string  `json:"facebook"`
	Twitter         *string  `json:"twitter"`
	Instagram       *string  `json:"instagram"`
	Services        []string `json:"services"`
}
