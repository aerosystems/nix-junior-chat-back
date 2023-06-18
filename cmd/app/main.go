package main

import (
	"fmt"
	ChatService "github.com/aerosystems/nix-junior-chat-back/internal/services/chat_service"
	TokenService "github.com/aerosystems/nix-junior-chat-back/internal/services/token_service"
	FirebaseClient "github.com/aerosystems/nix-junior-chat-back/pkg/firebase_client"
	"github.com/aerosystems/nix-junior-chat-back/pkg/gorm_client"
	"github.com/labstack/gommon/log"

	"github.com/aerosystems/nix-junior-chat-back/internal/handlers"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/aerosystems/nix-junior-chat-back/internal/storage"
	"github.com/aerosystems/nix-junior-chat-back/pkg/redis_client"
)

const webPort = 80

type Config struct {
	BaseHandler  *handlers.BaseHandler
	UserRepo     models.UserRepository
	TokenService *TokenService.Service
}

// @title NIX Junior: Chat App
// @version 1.0
// @description Backend App for simple social Live Chat

// @contact.name Artem Kostenko
// @contact.url https://github.com/aerosystems

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Should contain Access JWT Token, with the Bearer started

// @host localhost:80
// @BasePath /
func main() {
	clientGORM := GormClient.NewClient()
	clientGORM.AutoMigrate(models.User{}, models.Message{}, models.Chat{}, models.Device{})

	clientREDIS := RedisClient.NewClient()

	firebaseApp, _, _ := FirebaseClient.SetupFirebase()

	userRepo := storage.NewUserRepo(clientGORM, clientREDIS)
	messageRepo := storage.NewMessageRepo(clientGORM)
	chatRepo := storage.NewChatRepo(clientGORM)

	tokenService := TokenService.NewService(clientREDIS)
	chatService := ChatService.NewChatService(firebaseApp, clientREDIS, userRepo, messageRepo, chatRepo)

	baseHandler := handlers.NewBaseHandler(userRepo, messageRepo, chatRepo, tokenService, chatService)

	app := Config{
		BaseHandler:  baseHandler,
		UserRepo:     userRepo,
		TokenService: tokenService,
	}

	e := app.NewRouter()
	app.AddMiddleware(e)
	e.Logger.SetLevel(log.INFO)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", webPort)))
}
