package mongo

import (
	"chrome-extension-back-end/models"
	"chrome-extension-back-end/utils"
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	clientEncryption, err := utils.CreateClientEncryption(r.dbToStore)
	if err != nil {
		return err
	}

	dataKeyID := utils.CreateKey(r.dbToStore, clientEncryption)

	model := toMongoUser(user)

	bytesFromUser, err := json.Marshal(model)
	if err != nil {
		return err
	}

	rawValueType, rawValueData, err := bson.MarshalValue(string(bytesFromUser))

	rawValue := bson.RawValue{Type: rawValueType, Value: rawValueData}

	if err != nil {
		return err
	}

	encryptionOpts := options.Encrypt().
		SetAlgorithm("AEAD_AES_256_CBC_HMAC_SHA_512-Deterministic").
		SetKeyID(dataKeyID)

	encryptedField, err := clientEncryption.Encrypt(ctx, rawValue, encryptionOpts)

	res, err := r.dbToStore.Database("extensiondb").Collection("user_collection").InsertOne(ctx, bson.D{{"encryptedField", encryptedField}})
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
	var test interface{}
	if err := decrypted.Unmarshal(&test); err != nil {
		log.Println(err)
	}

	log.Println(test)

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
