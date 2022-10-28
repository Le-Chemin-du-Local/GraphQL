package users

import (
	"log"
	"strings"
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/address"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/database"
	"chemin-du-local.bzh/graphql/internal/registeredpaymentmethod"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

const USERROLE_ADMIN = "ADMIN"
const USERROLE_STOREKEEPER = "STOREKEEPER"
const USERROLE_USER = "USER"

// Type

type User struct {
	ID                       primitive.ObjectID                                `bson:"_id"`
	CreatedAt                time.Time                                         `bson:"createAt"`
	Email                    string                                            `bson:"email"`
	Phone                    string                                            `bson:"phone"`
	Role                     string                                            `bson:"role"`
	Gender                   *string                                           `bson:"gender"`
	FirstName                *string                                           `bson:"firstName"`
	LastName                 *string                                           `bson:"lastName"`
	Birthdate                *time.Time                                        `bson:"birthdate"`
	Addresses                []*address.Address                                `bson:"addresses"`
	DefaultAddressID         *primitive.ObjectID                               `bson:"defaultAddressID"`
	StripID                  *string                                           `bson:"stripeID"`
	RegisteredPaymentMethods []registeredpaymentmethod.RegisteredPaymentMethod `bson:"registeredPaymentMethods"`
	DefaultPaymentMethod     *string                                           `bson:"defaultPaymentMethod"`
	PasswordHash             string                                            `bson:"password_hash"`
}

func (user *User) ToModel() *model.User {
	addresses := []*model.Address{}

	for _, address := range user.Addresses {
		addresses = append(addresses, address.ToModel())
	}

	return &model.User{
		ID:        user.ID.Hex(),
		CreatedAt: &user.CreatedAt,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      user.Role,
		Gender:    user.Gender,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Addresses: addresses,
		Birthdate: user.Birthdate,
	}
}

func (user *User) HasRole(role model.Role) bool {
	if user == nil {
		return false
	}
	if user.Role == USERROLE_USER && role == model.RoleUser {
		return true
	}
	if user.Role == USERROLE_STOREKEEPER && (role == model.RoleUser || role == model.RoleStorekeeper) {
		return true
	}
	if user.Role == USERROLE_ADMIN {
		return true
	}

	return false
}

// Service

type UsersService interface {
	Create(input model.NewUser) (*User, error)
	Update(changes *User) error
	GetAllUser() ([]User, error)
	GetUserById(id string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetFiltered(filter interface{}) ([]User, error)
	Authenticate(login model.Login) bool
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}

// Createur de base de données

func Create(input model.NewUser) (*User, error) {
	hashedPassword, err := HashPassword(input.Password)

	if err != nil {
		log.Fatal(err)
	}

	// On doit créer la première address
	addressID := primitive.NewObjectID()
	addresses := []*address.Address{}

	if input.Address != nil {
		addresses = append(addresses, &address.Address{
			ID:            addressID,
			Number:        input.Address.Number,
			Route:         input.Address.Route,
			OptionalRoute: input.Address.OptionalRoute,
			PostalCode:    input.Address.PostalCode,
			City:          input.Address.City,
		})
	}

	// On a besoin de faire la conversion
	userID := primitive.NewObjectID()
	databaseUser := User{
		ID:               userID,
		CreatedAt:        time.Now(),
		Email:            strings.ToLower(input.Email),
		Phone:            input.Phone,
		Role:             USERROLE_USER,
		Gender:           input.Gender,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		Birthdate:        input.Birthdate,
		Addresses:        addresses,
		DefaultAddressID: &addressID,
		PasswordHash:     hashedPassword,
	}

	_, err = database.CollectionUsers.InsertOne(database.MongoContext, databaseUser)

	if err != nil {
		return nil, err
	}

	// Si le commerce n'est pas nul, il faut le créer
	if input.Commerce != nil {
		_, err = commerces.Create(*input.Commerce, userID)

		if err != nil {
			return nil, err
		}
	}

	return &databaseUser, nil
}

// Mise à jour de la base de données

func Update(changes *User) error {
	filter := bson.D{
		primitive.E{
			Key:   "_id",
			Value: changes.ID,
		},
	}

	_, err := database.CollectionUsers.ReplaceOne(database.MongoContext, filter, changes)

	return err
}

// Getter de base de données

func GetAllUser() ([]User, error) {
	filter := bson.D{{}}

	return GetFiltered(filter)
}

func GetUserById(id string) (*User, error) {
	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	filter := bson.D{
		primitive.E{
			Key:   "_id",
			Value: objectId,
		},
	}

	users, err := GetFiltered(filter)

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}

	return &users[0], nil
}

func GetUserByEmail(email string) (*User, error) {
	filter := bson.D{
		primitive.E{
			Key:   "email",
			Value: email,
		},
	}

	users, err := GetFiltered(filter)

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}

	return &users[0], nil
}

func GetFiltered(filter interface{}) ([]User, error) {
	users := []User{}

	cursor, err := database.CollectionUsers.Find(database.MongoContext, filter)

	if err != nil {
		return users, err
	}

	for cursor.Next(database.MongoContext) {
		var user User

		err := cursor.Decode(&user)

		if err != nil {
			return users, err
		}

		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		return users, err
	}

	return users, nil
}

// Authentification

func Authenticate(login model.Login) bool {
	user, err := GetUserByEmail(strings.ToLower(login.Email))

	if user == nil || err != nil {
		return false
	}

	return CheckPasswordHash(login.Password, user.PasswordHash)
}

// Gestion des mots de passes

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)

	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}
