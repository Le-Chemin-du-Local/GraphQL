package model

import "time"

type Panier struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Quantity    int        `json:"quantity"`
	EndingDate  *time.Time `json:"endingDate"`
	Price       float64    `json:"price"`
}
