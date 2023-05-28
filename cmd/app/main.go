package main

import (
	"fmt"
	TokenService "github.com/aerosystems/nix-junior-chat-back/internal/services/token_service"
	"github.com/aerosystems/nix-junior-chat-back/pkg/gormclient"
	"github.com/labstack/gommon/log"

	"github.com/aerosystems/nix-junior-chat-back/internal/handlers"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/aerosystems/nix-junior-chat-back/internal/storage"
	"github.com/aerosystems/nix-junior-chat-back/pkg/redisclient"
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
	clientGORM := gormclient.NewClient()
	clientGORM.AutoMigrate(models.User{}, models.Message{})

	clientREDIS := redisclient.NewClient()

	userRepo := storage.NewUserRepo(clientGORM, clientREDIS)
	messageRepo := storage.NewMessageRepo(clientGORM)
	tokenService := TokenService.NewService(clientREDIS)

	app := Config{
		BaseHandler: handlers.NewBaseHandler(
			userRepo,
			messageRepo,
			*tokenService,
		),
		UserRepo:     userRepo,
		TokenService: tokenService,
	}

	e := app.NewRouter()
	app.AddMiddleware(e)
	e.Logger.SetLevel(log.INFO)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", webPort)))
}
