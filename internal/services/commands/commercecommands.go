package commands

import (
	"time"

	"chemin-du-local.bzh/graphql/graph/model"
	"chemin-du-local.bzh/graphql/internal/commerces"
	"chemin-du-local.bzh/graphql/internal/database"
	"chemin-du-local.bzh/graphql/internal/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const COMMERCE_COMMAND_STATUS_IN_PROGRESS = "INPROGRESS"
const COMMERCE_COMMAND_STATUS_READY = "READY"
const COMMERCE_COMMAND_STATUS_DONE = "DONE"
const COMMERCE_COMMAND_STATUS_CANCELED = "CANCELED"

type CommerceCommand struct {
	ID                   primitive.ObjectID `bson:"_id"`
	CommandID            primitive.ObjectID `bson:"commandID"`
	CommerceID           primitive.ObjectID `bson:"commerceID"`
	PickupDate           time.Time          `bson:"pickupDate"`
	Price                int                `bson:"price"`
	PricePaniers         float64            `bson:"pricePaniers"`
	PriceClickAndCollect float64            `bson:"priceClickAndCollect"`
	PaymentMethod        string             `bson:"paymentMethod"`
	Status               string             `bson:"status"`
}

func (command *CommerceCommand) ToModel() *model.CommerceCommand {
	return &model.CommerceCommand{
		ID:         command.ID.Hex(),
		PickupDate: command.PickupDate,
		Status:     command.Status,
		Price:      float64(command.Price) / 100,
	}
}

func (commerceCommand CommerceCommand) IsLast(commerceCommandsService CommerceCommandsService) bool {
	filter := bson.D{{}}

	opts := options.FindOptions{}
	opts.SetLimit(1)
	opts.SetSort(bson.D{
		primitive.E{
			Key: "_id", Value: -1,
		},
	})

	lastCommerceCommand, err := commerceCommandsService.GetFiltered(filter, &opts)

	if err != nil || len(lastCommerceCommand) <= 0 {
		return false
	}

	return lastCommerceCommand[0].ID == commerceCommand.ID
}

// Le service
type commerceCommandsService struct {
	UsersService     users.UsersService
	CommercesService commerces.CommercesService
	CommandsService  CommandsService
}

type CommerceCommandsService interface {
	Create(input model.NewCommerceCommand, commandID primitive.ObjectID) (*CommerceCommand, error)
	Update(changes *CommerceCommand) error
	GetAll() ([]CommerceCommand, error)
	GetById(id string) (*CommerceCommand, error)
	GetForCommand(commandID string) ([]CommerceCommand, error)
	GetPaginated(startValue *string, first int, filter *model.CommerceCommandsFilter) ([]CommerceCommand, error)
	GetFiltered(filter interface{}, opts *options.FindOptions) ([]CommerceCommand, error)
	GetCommerce(commerceCommandID string) (*model.Commerce, error)
	GetUser(commerceCommandID string) (*model.User, error)
}

func NewCommerceCommandsService(
	usersService users.UsersService,
	commercesService commerces.CommercesService,
	commandsService CommandsService,
) *commerceCommandsService {
	return &commerceCommandsService{
		UsersService:     usersService,
		CommercesService: commercesService,
		CommandsService:  commandsService,
	}
}

func (c *commerceCommandsService) Create(input model.NewCommerceCommand, commandID primitive.ObjectID) (*CommerceCommand, error) {
	commerceObjectID, err := primitive.ObjectIDFromHex(input.CommerceID)

	if err != nil {
		return nil, err
	}

	databaseCommerceCommand := CommerceCommand{
		ID:                   primitive.NewObjectID(),
		CommandID:            commandID,
		CommerceID:           commerceObjectID,
		PickupDate:           input.PickupDate,
		Price:                input.Price,
		PriceClickAndCollect: input.PriceClickAndCollect,
		PricePaniers:         input.PricePaniers,
		PaymentMethod:        input.PaymentMethod,
		Status:               COMMERCE_COMMAND_STATUS_IN_PROGRESS,
	}

	_, err = database.CollectionCommerceCommand.InsertOne(database.MongoContext, databaseCommerceCommand)

	if err != nil {
		return nil, err
	}

	return &databaseCommerceCommand, nil
}

// Mise à jour de la base de données

func (c *commerceCommandsService) Update(changes *CommerceCommand) error {
	databaseCommand, err := c.CommandsService.GetById(changes.CommandID.Hex())

	if err != nil {
		return err
	}

	if databaseCommand == nil {
		return &CommandNotFoundError{}
	}

	filter := bson.D{
		primitive.E{
			Key:   "_id",
			Value: changes.ID,
		},
	}

	_, err = database.CollectionCommerceCommand.ReplaceOne(database.MongoContext, filter, changes)

	if err != nil {
		return err
	}

	commandStatus, err := c.CommandsService.GetTheoricalStatus(databaseCommand.ID.Hex(), c)

	if err != nil {
		return err
	}

	databaseCommand.Status = *commandStatus

	err = c.CommandsService.Update(databaseCommand)

	return err
}

// Getters

func (c *commerceCommandsService) GetAll() ([]CommerceCommand, error) {
	filter := bson.D{{}}

	return c.GetFiltered(filter, nil)
}

func (c *commerceCommandsService) GetById(id string) (*CommerceCommand, error) {
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

	commerceCommands, err := c.GetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	if len(commerceCommands) == 0 {
		return nil, nil
	}

	return &commerceCommands[0], nil
}

func (c *commerceCommandsService) GetForCommand(commandID string) ([]CommerceCommand, error) {
	commandObjectId, err := primitive.ObjectIDFromHex(commandID)

	if err != nil {
		return nil, err
	}

	filter := bson.D{
		primitive.E{
			Key:   "commandID",
			Value: commandObjectId,
		},
	}

	commerceCommands, err := c.GetFiltered(filter, nil)

	if err != nil {
		return nil, err
	}

	return commerceCommands, nil
}

func (c *commerceCommandsService) GetPaginated(startValue *string, first int, filter *model.CommerceCommandsFilter) ([]CommerceCommand, error) {
	var finalFilter bson.M

	if startValue != nil {
		objectID, err := primitive.ObjectIDFromHex(*startValue)

		if err != nil {
			return nil, err
		}

		finalFilter = bson.M{
			"_id": bson.M{
				"$gt": objectID,
			},
		}
	} else {
		finalFilter = bson.M{}
	}

	if filter != nil && filter.CommerceID != nil {
		commerceObjectID, err := primitive.ObjectIDFromHex(*filter.CommerceID)

		if err != nil {
			return nil, err
		}

		finalFilter = bson.M{
			"$and": []bson.M{
				finalFilter,
				{
					"commerceID": bson.M{
						"$eq": commerceObjectID,
					},
				},
			},
		}
	}

	if filter != nil && filter.Status != nil {
		finalFilter = bson.M{
			"$and": []bson.M{
				finalFilter,
				{
					"status": bson.M{
						"$in": filter.Status,
					},
				},
			},
		}
	}

	if filter != nil && filter.Year != nil {
		// Premier cas, on ne précise que l'année
		if filter.Month == nil {
			finalFilter = bson.M{
				"$and": []bson.M{
					finalFilter,
					{
						"pickupDate": bson.M{
							"$gte": time.Date(
								*filter.Year,
								1, 1, 0, 0, 1, 0, time.Local,
							),
						},
					},
					{
						"pickupDate": bson.M{
							"$lte": time.Date(
								*filter.Year+1,
								1, 1, 0, 0, 1, 0, time.Local,
							),
						},
					},
				},
			}
		} else {
			finalFilter = bson.M{
				"$and": []bson.M{
					finalFilter,
					{
						"pickupDate": bson.M{
							"$gte": time.Date(
								*filter.Year,
								time.Month(*filter.Month),
								1, 0, 0, 1, 0, time.Local,
							),
						},
					},
					{
						"pickupDate": bson.M{
							"$lte": time.Date(
								*filter.Year,
								time.Month(*filter.Month),
								31, 0, 0, 1, 0, time.Local,
							),
						},
					},
				},
			}
		}
	}

	opts := options.Find()
	opts.SetLimit(int64(first))

	return c.GetFiltered(finalFilter, opts)
}

func (c *commerceCommandsService) GetFiltered(filter interface{}, opts *options.FindOptions) ([]CommerceCommand, error) {
	commerceCommands := []CommerceCommand{}

	cursor, err := database.CollectionCommerceCommand.Find(database.MongoContext, filter, opts)

	if err != nil {
		return commerceCommands, err
	}

	for cursor.Next(database.MongoContext) {
		var commerceCommand CommerceCommand

		err := cursor.Decode(&commerceCommand)

		if err != nil {
			return commerceCommands, err
		}

		commerceCommands = append(commerceCommands, commerceCommand)
	}

	if err := cursor.Err(); err != nil {
		return commerceCommands, err
	}

	return commerceCommands, nil
}

func (c *commerceCommandsService) GetCommerce(commerceCommandID string) (*model.Commerce, error) {
	databaseCommerceCommand, err := c.GetById(commerceCommandID)

	if err != nil {
		return nil, err
	}

	if databaseCommerceCommand == nil {
		return nil, &CommerceCommandNotFoundError{}
	}

	databaseCommerce, err := c.CommercesService.GetById(databaseCommerceCommand.CommerceID.Hex())

	if err != nil {
		return nil, err
	}

	if databaseCommerce == nil {
		return nil, &commerces.CommerceErrorNotFound{}
	}

	return databaseCommerce.ToModel(), nil
}

func (c *commerceCommandsService) GetUser(commerceCommandID string) (*model.User, error) {
	databaseCommerceCommand, err := c.GetById(commerceCommandID)

	if err != nil {
		return nil, err
	}

	if databaseCommerceCommand == nil {
		return nil, &CommerceCommandNotFoundError{}
	}

	databaseCommand, err := c.CommandsService.GetById(databaseCommerceCommand.CommandID.Hex())

	if err != nil {
		return nil, err
	}

	if databaseCommand == nil {
		return nil, &CommandNotFoundError{}
	}

	databaseUser, err := c.UsersService.GetUserById(databaseCommand.UserID.Hex())

	if err != nil {
		return nil, err
	}

	if databaseUser == nil {
		return nil, &users.UserNotFoundError{}
	}

	return databaseUser.ToModel(), nil
}
