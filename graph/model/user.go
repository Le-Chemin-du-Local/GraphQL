package model

import "time"

type User struct {
	ID             string     `json:"id"`
	CreatedAt      *time.Time `json:"createdAt"`
	Email          string     `json:"email"`
	Phone          string     `json:"phone"`
	Role           string     `json:"role"`
	Addresses      []*Address `json:"addresses"`
	DefaultAddress *Address   `json:"defaultAddress"`
	Gender         *string    `json:"gender"`
	FirstName      *string    `json:"firstName"`
	LastName       *string    `json:"lastName"`
	Birthdate      *time.Time `json:"birthdate"`
	CommerceID     *string    `json:"commerce"`
}
