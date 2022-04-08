package model

import "time"

type PanierCommand struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	PickupDate time.Time `json:"pickupDate"`
}
