package model

import "time"

type CommerceCommand struct {
	ID         string    `json:"id"`
	PickupDate time.Time `json:"pickupDate"`
	Status     string    `json:"status"`
}
