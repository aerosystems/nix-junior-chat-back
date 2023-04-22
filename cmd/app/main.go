package main

import (
	"fmt"
	"github.com/labstack/gommon/log"

	"github.com/aerosystems/nix-junior-chat-back/internal/handlers"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/aerosystems/nix-junior-chat-back/internal/storage"
	"github.com/aerosystems/nix-junior-chat-back/pkg/myredis"
	"github.com/aerosystems/nix-junior-chat-back/pkg/mysql/mygorm"
)

const webPort = 80

type Config struct {
	BaseHandler *handlers.BaseHandler
	UserRepo    models.UserRepository
	TokensRepo  models.TokensRepository
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
	clientGORM := mygorm.NewClient()
	clientGORM.AutoMigrate(models.User{})
	clientREDIS := myredis.NewClient()
	userRepo := storage.NewUserRepo(clientGORM, clientREDIS)
	tokensRepo := storage.NewTokensRepo(clientREDIS)

	app := Config{
		BaseHandler: handlers.NewBaseHandler(
			userRepo,
			tokensRepo,
		),
		UserRepo:   userRepo,
		TokensRepo: tokensRepo,
	}

	e := app.NewRouter()
	app.AddMiddleware(e)
	e.Logger.SetLevel(log.INFO)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", webPort)))
}
