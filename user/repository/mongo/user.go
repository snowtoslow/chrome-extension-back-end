package mongo

import (
	"chrome-extension-back-end/models"
	"chrome-extension-back-end/utils"
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
	dbToStore *mongo.Client
}

func NewUserRepository(databaseToStore *mongo.Client) *UserRepository {
	return &UserRepository{
		dbToStore: databaseToStore,
	}
}

func (r UserRepository) CreateUser(ctx context.Context, user *models.User) (err error) {

	encryptedString, err := utils.EncryptField(user.Email, r.dbToStore)
	if err != nil {
		return err
	}

	encryptedPersonalValues, err := utils.EncryptArrayFields(user.PersonalData, r.dbToStore)
	if err != nil {
		return err
	}

	user.Email = encryptedString
	user.PersonalData = encryptedPersonalValues

	model := toMongoUser(user)
	res, err := r.dbToStore.Database("extensiondb").Collection("user_collection").InsertOne(ctx, model)
	if err != nil {
		return err
	}

	user.Id = res.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r UserRepository) GetUserById(ctx context.Context, id string) (user *models.User, err error) {

	clientEncryption, err := utils.CreateClientEncryption(r.dbToStore)
	if err != nil {
		return nil, err
	}

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var foundDoc bson.M
	err = r.dbToStore.Database("extensiondb").Collection("user_collection").FindOne(ctx, bson.M{
		"_id": objectId,
	}).Decode(&foundDoc)
	if err != nil {
		return nil, err
	}

	// Decrypt the encrypted field in the found document.
	decrypted, err := clientEncryption.Decrypt(ctx, foundDoc["encryptedField"].(primitive.Binary))
	if err != nil {
		return nil, err
	}

	myStr := string(decrypted.Value)

	log.Println(myStr[1:])

	/*myUser := new(User)*/

	/*myUser := new(User)
	err = r.dbToStore.Database("").Collection("").FindOne(ctx, bson.M{
		"_id": objectId,
	}).Decode(myUser)

	if err != nil {
		return nil, err
	}

	return toModel(myUser), nil*/
	return nil, nil
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
