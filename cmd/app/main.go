package main

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/handlers"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/aerosystems/nix-junior-chat-back/internal/storage"
	"github.com/aerosystems/nix-junior-chat-back/pkg/myredis"
	"github.com/aerosystems/nix-junior-chat-back/pkg/mysql/mygorm"
)

const webPort = 8080

type Config struct {
	BaseHandler       *handlers.BaseHandler
	TokensRepo        models.TokensRepository
}

// @title NIX Junior: Chat App
// @version 1.0
// @description Backend App for simple social Live Chat

// @contact.name Artem Kostenko
// @contact.url https://github.com/aerosystems

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:80
// @BasePath /v1
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
		TokensRepo: tokensRepo,
	}

	e := app.NewRouter()
	app.AddMiddleware(e)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", webPort)))
}
