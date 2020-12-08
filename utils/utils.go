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
	kmsProviders, keyVaultNamespace, err := CreateKmsProviders()
	clientEncryptionOpts := options.ClientEncryption().
		SetKmsProviders(kmsProviders).
		SetKeyVaultNamespace(keyVaultNamespace)

	clientEncryption, err := mongo.NewClientEncryption(client, clientEncryptionOpts)
	if err != nil {
		return nil, err
	}

	return clientEncryption, err
}

func CreateKmsProviders() (map[string]map[string]interface{}, string, error) {
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
