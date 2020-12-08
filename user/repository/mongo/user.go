package mongo

import (
	"chrome-extension-back-end/models"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Email        string             `bson:"email"`
	Password     string             `bson:"password"`
	PersonalData []string           `bson:"personalData"`
}

type UserRepository struct {
	dbToStore *mongo.Collection
	dbForKeys *mongo.Collection
}

func NewUserRepository(databaseToStore *mongo.Database, collectionToStore string, databaseWithKeys *mongo.Database, collectionWithKeys string) *UserRepository {
	return &UserRepository{
		dbToStore: databaseToStore.Collection(collectionToStore),
		dbForKeys: databaseWithKeys.Collection(collectionWithKeys),
	}
}

func (r UserRepository) CreateUser(ctx context.Context, user *models.User) (err error) {
	model := toMongoUser(user)
	res, err := r.dbToStore.InsertOne(ctx, model)
	if err != nil {
		log.Println("EERRR", err)
		return err
	}

	user.Id = res.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r UserRepository) GetUserById(ctx context.Context, id string) (user *models.User, err error) {

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	myUser := new(User)
	err = r.dbToStore.FindOne(ctx, bson.M{
		"_id": objectId,
	}).Decode(myUser)

	if err != nil {
		return nil, err
	}

	return toModel(myUser), nil
}

func toMongoUser(u *models.User) *User {
	return &User{
		Email:        u.Email,
		Password:     u.Password,
		PersonalData: u.PersonalData,
	}
}

func toModel(u *User) *models.User {
	return &models.User{
		Id:           u.ID.Hex(),
		Email:        u.Email,
		Password:     u.Password,
		PersonalData: u.PersonalData,
	}
}
