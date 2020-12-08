package server

import (
	"chrome-extension-back-end/user"
	ushttp "chrome-extension-back-end/user/delivery/http"
	userrepo "chrome-extension-back-end/user/repository/mongo"
	userusecase "chrome-extension-back-end/user/usecase"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type App struct {
	httpServer *http.Server
	userUc     user.UseCase
}

func NewApp() *App {

	//dbWithData := initDB()

	dbWithKeys, err := initKeyDb()
	if err != nil {
		log.Println("ERROR")
	}

	fmt.Println("DATABASE SUCESSFULY CONECTED!", dbWithKeys)
	//fmt.Println("DATABASE FOR KEYS SUCESSFULY CONECTED!", dbWithKeys)

	benchRepo := userrepo.NewUserRepository(dbWithKeys, viper.GetString("mongo.user_collection"))

	return &App{
		userUc: userusecase.NewUserUseCase(benchRepo),
	}
}

func (a *App) Run(port string) error {
	/*router := gin.Default()
	router.Use(
		gin.Recovery(),
		gin.Logger(),
	)*/

	router := gin.New()

	api := router.Group("/api")

	router.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type", //		RequestHeaders: "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}), gin.Recovery(),
		gin.Logger())

	ushttp.RegisterHTTPEndpoints(api, a.userUc)

	a.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil {
			log.Fatalf("Failed to listen and serve")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return a.httpServer.Shutdown(ctx)

}

func initDB() *mongo.Database {
	client, err := mongo.NewClient(options.Client().ApplyURI(viper.GetString("mongo.uri")))
	if err != nil {
		log.Fatalf("Error occured while establishing connection to mongoDB")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	return client.Database(viper.GetString("mongo.name"))
}

func initKeyDb() (*mongo.Database, error) {
	decodeKey, err := base64.StdEncoding.DecodeString(viper.GetString("mongo-secured-keys.local_master_key"))

	kmsProviders := map[string]map[string]interface{}{
		"local": {
			"key": decodeKey,
		},
	}

	// The MongoDB namespace (db.collection) used to store the encryption data keys.

	keyVaultNamespace := viper.GetString("mongo.name") + "." + viper.GetString("mongo-secured-keys.key_vault_collection_name")

	// The Client used to read/write application data.
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(viper.GetString("mongo.uri")))
	if err != nil {
		panic(err)
	}
	//defer func() { _ = client.Disconnect(context.TODO()) }()

	/*	// Get a handle to the application collection and clear existing data.
		coll := client.Database("test").Collection("coll")
		_ = coll.Drop(context.TODO())
	*/
	// Set up the key vault for this example.
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
	if _, err = keyVaultColl.Indexes().CreateOne(context.TODO(), keyVaultIndex); err != nil {
		return nil, err
	}

	// Create the ClientEncryption object to use for explicit encryption/decryption. The Client passed to
	// NewClientEncryption is used to read/write to the key vault. This can be the same Client used by the main
	// application.
	clientEncryptionOpts := options.ClientEncryption().
		SetKmsProviders(kmsProviders).
		SetKeyVaultNamespace(keyVaultNamespace)
	clientEncryption, err := mongo.NewClientEncryption(client, clientEncryptionOpts)
	if err != nil {
		return nil, err
	}
	//defer func() { _ = clientEncryption.Close(context.TODO()) }()

	// Create a new data key for the encrypted field.
	dataKeyOpts := options.DataKey().SetKeyAltNames([]string{"go_encryption_example"})
	_, err = clientEncryption.CreateDataKey(context.TODO(), "local", dataKeyOpts)
	if err != nil {
		return nil, err
	}
	//dataKeyId

	return client.Database(viper.GetString("mongo.name")), nil
}
