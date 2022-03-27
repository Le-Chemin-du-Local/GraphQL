package model

type Panier struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"`
}
