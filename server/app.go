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
	"crypto/tls"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	csrf "github.com/utrack/gin-csrf"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"net/http"
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
	// Initialize a new Gin router
	router := gin.New()
	gin.ForceConsoleColor()

	store := cookie.NewStore([]byte("secret"))

	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})) // add a required custom format of logging

	if err := writeLogs(); err != nil {
		log.Println("ERROR WHILE CREATE LOGGING FILE!")
	}

	router.Use(sessions.Sessions("store", store), cors(), csrf.Middleware(csrf.Options{
		Secret: "token123",
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))

	authhttp.RegisterHTTPEndpoints(router, a.authUC)

	// API endpoints
	authMiddleware := authhttp.NewAuthMiddleware(a.authUC)
	api := router.Group("/api", authMiddleware)

	ushttp.RegisterHTTPEndpoints(api, a.userUc)

	cer, err := tls.LoadX509KeyPair("../../cert/server.crt", "../../cert/server.key")
	if err != nil {
		log.Println(err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	a.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      config,
	}

	go func() {
		if err := a.httpServer.ListenAndServeTLS("../../cert/server.crt", "../../cert/server.key"); err != nil {
			log.Fatalf("Failed to listen and serve:", err)
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

func cors() gin.HandlerFunc {
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

// create a log file;
func writeLogs() (err error) {
	// Logging to a file.
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	f, err := os.Create(home + "/Desktop/" + "gin.log")
	if err != nil {
		return err
	}
	gin.DefaultWriter = io.MultiWriter(f)
	return nil
}
