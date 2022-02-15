package users

import (
	"log"
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

const USERROLE_ADMIN = "ADMIN"
const USERROLE_STOREKEEPER = "STOREKEEPER"
const USERROLE_USER = "USER"

// Type

type User struct {
	ID           primitive.ObjectID `bson:"_id"`
	CreatedAt    time.Time          `bson:"createAt"`
	Email        string             `bson:"email"`
	Role         string             `bson:"role"`
	FirstName    *string            `bson:"firstName"`
	LastName     *string            `bson:"lastName"`
	PasswordHash string             `bson:"password_hash"`
}

func (user *User) ToModel() *model.User {
	return &model.User{
		ID:        user.ID.Hex(),
		CreatedAt: &user.CreatedAt,
		Email:     user.Email,
		Role:      user.Role,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}

func (user *User) HasRole(role model.Role) bool {
	if user.Role == USERROLE_USER && role == model.RoleUser {
		return true
	} else if user.Role == USERROLE_STOREKEEPER && (role == model.RoleUser || role == model.RoleStorekeeper) {
		return true
	} else if user.Role == USERROLE_ADMIN {
		return true
	}

	return false
}

// Createur de base de données

func Create(input model.NewUser) *User {
	hashedPassword, err := HashPassword(input.Password)

	if err != nil {
		log.Fatal(err)
	}

	// On a besoin de faire la conversion
	databaseUser := User{
		ID:           primitive.NewObjectID(),
		CreatedAt:    time.Now(),
		Email:        input.Email,
		Role:         USERROLE_USER,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		PasswordHash: hashedPassword,
	}

	_, err = database.CollectionUsers.InsertOne(database.MongoContext, databaseUser)

	if err != nil {
		log.Fatal(err)
	}

	return &databaseUser
}

// Getter de base de données

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
	user, err := GetUserByEmail(login.Email)

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
