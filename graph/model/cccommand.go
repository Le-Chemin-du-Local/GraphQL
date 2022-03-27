package model

import "time"

type CCCommand struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	PickupDate time.Time `json:"pickupDate"`
}
