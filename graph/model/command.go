package model

import "time"

type Command struct {
	ID           string    `json:"id"`
	CreationDate time.Time `json:"creationDate"`
	User         string    `json:"user"`
}
