package mongo

import (
	"chrome-extension-back-end/auth"
	"chrome-extension-back-end/models"
	"chrome-extension-back-end/utils"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Email        string             `bson:"email"`
	Password     string             `bson:"password"`
	PersonalData []string           `bson:"personalData"`
}

type UserDTO struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty"`
	Email        string              `bson:"email"`
	Password     *primitive.Binary   `bson:"password"`
	PersonalData []*primitive.Binary `bson:"personalData"`
}

type UserRepository struct {
	dbToStore *mongo.Client
}

func NewUserRepository(databaseToStore *mongo.Client) *UserRepository {
	return &UserRepository{
		dbToStore: databaseToStore,
	}
}

func (r UserRepository) GetUserByEmailAndPassword(ctx context.Context, email, password string) (*models.User, error) {

	var myArray []string

	clientEncryption, err := utils.CreateClientEncryption(r.dbToStore)
	if err != nil {
		return nil, err
	}

	user := new(UserDTO)
	err = r.dbToStore.Database("extensiondb").Collection("user_collection").FindOne(ctx, bson.M{
		"email": email,
	}).Decode(user)

	if err != nil {
		return nil, err
	}

	decrypted, err := clientEncryption.Decrypt(context.TODO(), *user.Password)
	if err != nil {
		return nil, err
	}

	unquotedDecrypt, err := strconv.Unquote(decrypted.String())
	if err != nil {
		return nil, err
	}

	if unquotedDecrypt != password {
		return nil, auth.ErrInvalidPassword
	}

	if user.PersonalData != nil {
		for _, v := range user.PersonalData {
			if v != nil {
				decryptedVal, err := clientEncryption.Decrypt(context.TODO(), *v)
				if err != nil {
					return nil, err
				}
				unquoted, _ := strconv.Unquote(decryptedVal.String())
				myArray = append(myArray, unquoted)
			}
		}
	}

	foundUser := toUser(unquotedDecrypt, myArray, user.Email, user.ID.Hex())

	return foundUser, nil
}

func (r UserRepository) CreateUser(ctx context.Context, user *models.User) (err error) {

	myrray := make([]*primitive.Binary, len(user.PersonalData))

	clientEncryption, err := utils.CreateClientEncryption(r.dbToStore)
	if err != nil {
		return err
	}

	dataKeyID := utils.CreateKey(r.dbToStore, clientEncryption)

	rawValueType, rawValueData, err := bson.MarshalValue(user.Password)
	if err != nil {
		return err
	}

	rawValue := bson.RawValue{Type: rawValueType, Value: rawValueData}

	encryptionOpts := options.Encrypt().
		SetAlgorithm("AEAD_AES_256_CBC_HMAC_SHA_512-Deterministic").
		SetKeyID(dataKeyID)

	encryptedField, err := clientEncryption.Encrypt(context.Background(), rawValue, encryptionOpts)
	if err != nil {
		return err
	}

	if user.PersonalData != nil {
		for _, v := range user.PersonalData {
			rawValueType, rawValueData, err := bson.MarshalValue(v)
			if err != nil {
				return err
			}
			rawValue := bson.RawValue{Type: rawValueType, Value: rawValueData}
			encryptedField, err := clientEncryption.Encrypt(context.Background(), rawValue, encryptionOpts)
			if err != nil {
				return err
			}
			myrray = append(myrray, &encryptedField)
		}
	}

	model := toUserDTO(&encryptedField, user.Email, myrray)

	res, err := r.dbToStore.Database("extensiondb").Collection("user_collection").InsertOne(ctx, model)
	if err != nil {
		return err
	}

	user.Id = res.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r UserRepository) GetUserById(ctx context.Context, id string) (user *models.User, err error) {

	var myArray []string

	clientEncryption, err := utils.CreateClientEncryption(r.dbToStore)
	if err != nil {
		return nil, err
	}

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	myUser := new(UserDTO)
	err = r.dbToStore.Database("extensiondb").Collection("user_collection").FindOne(ctx, bson.M{
		"_id": objectId,
	}).Decode(myUser)

	decrypted, err := clientEncryption.Decrypt(context.TODO(), *myUser.Password)
	if err != nil {
		return nil, err
	}

	unquotedDecrypt, err := strconv.Unquote(decrypted.String())
	if err != nil {
		return nil, err
	}

	if myUser.PersonalData != nil {
		for _, v := range myUser.PersonalData {
			if v != nil {
				decryptedVal, err := clientEncryption.Decrypt(context.TODO(), *v)
				if err != nil {
					return nil, err
				}
				unquoted, _ := strconv.Unquote(decryptedVal.String())
				myArray = append(myArray, unquoted)
			}
		}
	}

	foundUser := toUser(unquotedDecrypt, myArray, myUser.Email, myUser.ID.Hex())

	return foundUser, nil
}

func (r UserRepository) UpdateUser(ctx context.Context, dto *models.PatchDTO) (err error) {
	result, err := r.dbToStore.Database("extensiondb").Collection("user_collection").UpdateOne(
		ctx,
		bson.M{"email": dto.Email},
		bson.D{
			{"$set", bson.D{{"personalData", dto.PersonalData}}},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %v Documents!\n", result)

	return nil
}

func toUser(email string, myPersonalData []string, password string, id string) *models.User {
	return &models.User{
		Id:           id,
		Email:        email,
		PersonalData: myPersonalData,
		Password:     password,
	}
}

func toUserDTO(binary *primitive.Binary, string2 string, binary2 []*primitive.Binary) *UserDTO {
	return &UserDTO{
		Password:     binary,
		Email:        string2,
		PersonalData: binary2,
	}
}
