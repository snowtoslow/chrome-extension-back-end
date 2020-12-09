package utils

import (
	"context"
	"encoding/base64"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func CreateKey(client *mongo.Client, clientEncryption *mongo.ClientEncryption) primitive.Binary {

	keyVaultColl := client.Database(viper.GetString("mongo.name")).Collection(viper.GetString("mongo-secured-keys.key_vault_collection_name"))
	_ = keyVaultColl.Drop(context.TODO())
	// Ensure that two data keys cannot share the same keyAltName.
	keyVaultIndex := mongo.IndexModel{
		Keys: bson.D{{"keyAltNames", 1}},
		Options: options.Index().
			SetUnique(true).
			SetPartialFilterExpression(bson.D{
				{"keyAltNames", bson.D{
					{"$exists", true},
				}},
			}),
	}
	if _, err := keyVaultColl.Indexes().CreateOne(context.TODO(), keyVaultIndex); err != nil {
		log.Println(err)
	}

	dataKeyOpts := options.DataKey().SetKeyAltNames([]string{"go_encryption_example"})
	dataKeyId, err := clientEncryption.CreateDataKey(context.TODO(), "local", dataKeyOpts)
	if err != nil {
		log.Println(err)
	}

	return dataKeyId
}

func CreateClientEncryption(client *mongo.Client) (*mongo.ClientEncryption, error) {
	kmsProviders, keyVaultNamespace, err := createKmsProviders()
	clientEncryptionOpts := options.ClientEncryption().
		SetKmsProviders(kmsProviders).
		SetKeyVaultNamespace(keyVaultNamespace)

	clientEncryption, err := mongo.NewClientEncryption(client, clientEncryptionOpts)
	if err != nil {
		return nil, err
	}

	return clientEncryption, err
}

func createKmsProviders() (map[string]map[string]interface{}, string, error) {
	keyVaultNamespace := viper.GetString("mongo.name") + "." + viper.GetString("mongo-secured-keys.key_vault_collection_name")
	decodeKey, err := base64.StdEncoding.DecodeString(viper.GetString("mongo-secured-keys.local_master_key"))
	if err != nil {
		return nil, keyVaultNamespace, err
	}

	kmsProviders := map[string]map[string]interface{}{
		"local": {
			"key": decodeKey,
		},
	}

	return kmsProviders, keyVaultNamespace, nil
}

func EncryptArrayFields(dataToEncrypt []string, client *mongo.Client) (encryptedArray []*primitive.Binary, err error) {
	if dataToEncrypt != nil {
		for _, v := range dataToEncrypt {
			encryptedString, err := EncryptField(v, client)
			if err != nil {
				return nil, err
			}
			encryptedArray = append(encryptedArray, encryptedString)
		}
	}
	return encryptedArray, nil
}

func EncryptField(field string, client *mongo.Client) (encryptedString *primitive.Binary, err error) {
	clientEncryption, err := CreateClientEncryption(client)
	if err != nil {
		return nil, err
	}

	dataKeyID := CreateKey(client, clientEncryption)

	rawValueType, rawValueData, err := bson.MarshalValue(field)
	if err != nil {
		return nil, err
	}

	rawValue := bson.RawValue{Type: rawValueType, Value: rawValueData}

	encryptionOpts := options.Encrypt().
		SetAlgorithm("AEAD_AES_256_CBC_HMAC_SHA_512-Deterministic").
		SetKeyID(dataKeyID)

	encryptedField, err := clientEncryption.Encrypt(context.Background(), rawValue, encryptionOpts)
	if err != nil {
		return nil, err
	}

	return &encryptedField, nil
}

func DecryptArray(values []*primitive.Binary, client *mongo.Client) (myarray []string, err error) {
	if values != nil {
		for _, v := range values {
			decrypted, err := DecryptField(v, client)
			if err != nil {
				return nil, err
			}
			myarray = append(myarray, decrypted)
		}
	}

	return myarray, nil
}

func DecryptField(value *primitive.Binary, client *mongo.Client) (string, error) {
	clientEncryption, err := CreateClientEncryption(client)
	if err != nil {
		return "", err
	}

	decrypted, err := clientEncryption.Decrypt(context.TODO(), *value)
	if err != nil {
		return "", err
	}

	return decrypted.String(), err
}
