// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

type Basket struct {
	Commerces []*BasketCommerce `json:"commerces"`
}

type BasketCommerce struct {
	Commerce *Commerce        `json:"commerce"`
	Products []*BasketProduct `json:"products"`
	Paniers  []*Panier        `json:"paniers"`
}

type BasketProduct struct {
	Quantity float64  `json:"quantity"`
	Product  *Product `json:"product"`
}

type BulkChangesProduct struct {
	ID      string                 `json:"id"`
	Changes map[string]interface{} `json:"changes"`
}

type BusinessHours struct {
	Monday    []*Schedule `json:"monday"`
	Tuesday   []*Schedule `json:"tuesday"`
	Wednesday []*Schedule `json:"wednesday"`
	Thursday  []*Schedule `json:"thursday"`
	Friday    []*Schedule `json:"friday"`
	Saturday  []*Schedule `json:"saturday"`
	Sunday    []*Schedule `json:"sunday"`
}

type CCCommandFilter struct {
	Status *string `json:"status"`
}

type CCProduct struct {
	Quantity int      `json:"quantity"`
	Product  *Product `json:"product"`
}

type CommandConnection struct {
	Edges    []*CommandEdge   `json:"edges"`
	PageInfo *CommandPageInfo `json:"pageInfo"`
}

type CommandEdge struct {
	Cursor string   `json:"cursor"`
	Node   *Command `json:"node"`
}

type CommandPageInfo struct {
	StartCursor string `json:"startCursor"`
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type CommerceCommandConnection struct {
	Edges    []*CommerceCommandEdge   `json:"edges"`
	PageInfo *CommerceCommandPageInfo `json:"pageInfo"`
}

type CommerceCommandEdge struct {
	Cursor string           `json:"cursor"`
	Node   *CommerceCommand `json:"node"`
}

type CommerceCommandPageInfo struct {
	StartCursor string `json:"startCursor"`
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type CommerceConnection struct {
	Edges    []*CommerceEdge   `json:"edges"`
	PageInfo *CommercePageInfo `json:"pageInfo"`
}

type CommerceEdge struct {
	Cursor string    `json:"cursor"`
	Node   *Commerce `json:"node"`
}

type CommerceFilter struct {
	NearLatitude  *float64 `json:"nearLatitude"`
	NearLongitude *float64 `json:"nearLongitude"`
	Radius        *float64 `json:"radius"`
}

type CommercePageInfo struct {
	StartCursor string `json:"startCursor"`
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type Filter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type NewBasket struct {
	Commerces []*NewBasketCommerce `json:"commerces"`
}

type NewBasketCommerce struct {
	CommerceID string              `json:"commerceID"`
	Products   []*NewBasketProduct `json:"products"`
	Paniers    []string            `json:"paniers"`
	PickupDate *time.Time          `json:"pickupDate"`
}

type NewBasketProduct struct {
	Quantity  float64 `json:"quantity"`
	ProductID string  `json:"productID"`
}

type NewBusinessHours struct {
	Monday    []*ScheduleInput `json:"monday"`
	Tuesday   []*ScheduleInput `json:"tuesday"`
	Wednesday []*ScheduleInput `json:"wednesday"`
	Thursday  []*ScheduleInput `json:"thursday"`
	Friday    []*ScheduleInput `json:"friday"`
	Saturday  []*ScheduleInput `json:"saturday"`
	Sunday    []*ScheduleInput `json:"sunday"`
}

type NewCCCommand struct {
	ProductsID []*NewCCProcuct `json:"productsID"`
	PickupDate time.Time       `json:"pickupDate"`
}

type NewCCProcuct struct {
	Quantity  int    `json:"quantity"`
	ProductID string `json:"productID"`
}

type NewCommand struct {
	CreationDate time.Time `json:"creationDate"`
	User         string    `json:"user"`
}

type NewCommerce struct {
	Name            string            `json:"name"`
	Description     *string           `json:"description"`
	StorekeeperWord *string           `json:"storekeeperWord"`
	Address         string            `json:"address"`
	Latitude        float64           `json:"latitude"`
	Longitude       float64           `json:"longitude"`
	Phone           string            `json:"phone"`
	Email           string            `json:"email"`
	Facebook        *string           `json:"facebook"`
	Twitter         *string           `json:"twitter"`
	Instagram       *string           `json:"instagram"`
	BusinessHours   *NewBusinessHours `json:"businessHours"`
	ProfilePicture  *graphql.Upload   `json:"profilePicture"`
	Image           *graphql.Upload   `json:"image"`
}

type NewCommerceCommand struct {
	CommerceID string    `json:"commerceID"`
	PickupDate time.Time `json:"pickupDate"`
}

type NewPanier struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Type        string              `json:"type"`
	Category    string              `json:"category"`
	Quantity    int                 `json:"quantity"`
	Price       float64             `json:"price"`
	Reduction   float64             `json:"reduction"`
	Image       *graphql.Upload     `json:"image"`
	EndingDate  *time.Time          `json:"endingDate"`
	Products    []*NewPanierProduct `json:"products"`
}

type NewPanierCommand struct {
	PanierID   string    `json:"panierID"`
	PickupDate time.Time `json:"pickupDate"`
}

type NewPanierProduct struct {
	Quantity  int    `json:"quantity"`
	ProductID string `json:"productID"`
}

type NewProduct struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       float64         `json:"price"`
	Unit        string          `json:"unit"`
	Tva         float64         `json:"tva"`
	IsBreton    bool            `json:"isBreton"`
	Tags        []string        `json:"tags"`
	Categories  []string        `json:"categories"`
	Image       *graphql.Upload `json:"image"`
}

type NewUser struct {
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

type PanierCommandFilter struct {
	Status *string `json:"status"`
}

type PanierConnection struct {
	Edges    []*PanierEdge   `json:"edges"`
	PageInfo *PanierPageInfo `json:"pageInfo"`
}

type PanierEdge struct {
	Cursor string  `json:"cursor"`
	Node   *Panier `json:"node"`
}

type PanierFilter struct {
	Type *string `json:"type"`
}

type PanierPageInfo struct {
	StartCursor string `json:"startCursor"`
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type PanierProduct struct {
	Quantity int      `json:"quantity"`
	Product  *Product `json:"product"`
}

type Product struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Unit        string   `json:"unit"`
	Tva         float64  `json:"tva"`
	IsBreton    bool     `json:"isBreton"`
	Tags        []string `json:"tags"`
	Categories  []string `json:"categories"`
}

type ProductConnection struct {
	Edges    []*ProductEdge   `json:"edges"`
	PageInfo *ProductPageInfo `json:"pageInfo"`
}

type ProductEdge struct {
	Cursor string   `json:"cursor"`
	Node   *Product `json:"node"`
}

type ProductFilter struct {
	Category *string `json:"category"`
}

type ProductPageInfo struct {
	StartCursor string `json:"startCursor"`
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type Schedule struct {
	Opening string `json:"opening"`
	Closing string `json:"closing"`
}

type ScheduleInput struct {
	Opening string `json:"opening"`
	Closing string `json:"closing"`
}

type Role string

const (
	RoleAdmin       Role = "ADMIN"
	RoleStorekeeper Role = "STOREKEEPER"
	RoleUser        Role = "USER"
)

var AllRole = []Role{
	RoleAdmin,
	RoleStorekeeper,
	RoleUser,
}

func (e Role) IsValid() bool {
	switch e {
	case RoleAdmin, RoleStorekeeper, RoleUser:
		return true
	}
	return false
}

func (e Role) String() string {
	return string(e)
}

func (e *Role) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Role(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Role", str)
	}
	return nil
}

func (e Role) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
