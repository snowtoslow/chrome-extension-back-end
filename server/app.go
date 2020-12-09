package server

import (
	"chrome-extension-back-end/auth"
	authhttp "chrome-extension-back-end/auth/delivery/http"
	authusecase "chrome-extension-back-end/auth/usecase"
	"chrome-extension-back-end/user"
	ushttp "chrome-extension-back-end/user/delivery/http"
	userrepo "chrome-extension-back-end/user/repository/mongo"
	userusecase "chrome-extension-back-end/user/usecase"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	/*"github.com/itsjamie/gin-cors"*/
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"os/signal"
	"time"
)

type App struct {
	httpServer *http.Server
	userUc     user.UseCase
	authUC     auth.UseCase
}

func NewApp() *App {

	dbWithData, err := initClient()
	if err != nil {
		log.Println(err)
	}

	fmt.Println("DATABASE SUCESSFULY CONECTED!", dbWithData)

	userRepo := userrepo.NewUserRepository(dbWithData)

	return &App{
		userUc: userusecase.NewUserUseCase(userRepo),
		authUC: authusecase.NewAuthUseCase(
			*userRepo,
			viper.GetString("auth.hash_salt"),
			[]byte(viper.GetString("auth.signing_key")),
			viper.GetDuration("auth.token_ttl"),
		),
	}
}

func (a *App) Run(port string) error {
	/*router := gin.Default()
	router.Use(
		gin.Recovery(),
		gin.Logger(),
	)*/

	// Initialize a new Gin router
	router := gin.New()

	/*router.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type", //		RequestHeaders: "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,

	}), gin.Recovery(),
		gin.Logger())*/

	router.Use(CORS())

	authhttp.RegisterHTTPEndpoints(router, a.authUC)

	// API endpoints
	authMiddleware := authhttp.NewAuthMiddleware(a.authUC)
	api := router.Group("/api", authMiddleware)

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

func initClient() (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(viper.GetString("mongo.uri")))
	if err != nil {
		log.Fatalf("Error occured while establishing connection to mongoDB")
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return client, nil
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

func CORS() gin.HandlerFunc {
	// TO allow CORS
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		//c.Writer.Header().Set("ValidateHeaders","false")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
