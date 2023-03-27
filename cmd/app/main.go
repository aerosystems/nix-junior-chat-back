package main

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/handlers"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/aerosystems/nix-junior-chat-back/internal/storage"
	"github.com/aerosystems/nix-junior-chat-back/pkg/myredis"
	"github.com/aerosystems/nix-junior-chat-back/pkg/mysql/mygorm"

	"golang.org/x/oauth2"
)

const webPort = 8080

type Config struct {
	BaseHandler       *handlers.BaseHandler
	GoogleOauthConfig *oauth2.Config
	TokensRepo        models.TokensRepository
}

// @title NIX Trainee 5-6-7-8 tasks
// @version 1.0
// @description Simple REST API for CRUD operations with Comments & Posts entities.

// @contact.name Artem Kostenko
// @contact.url https://github.com/aerosystems

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /v1
func main() {
	clientGORM := mygorm.NewClient()
	clientREDIS := myredis.NewClient()
	userRepo := storage.NewUserRepo(clientGORM, clientREDIS)
	codeRepo := storage.NewCodeRepo(clientGORM)
	tokensRepo := storage.NewTokensRepo(clientREDIS)

	app := Config{
		BaseHandler: handlers.NewBaseHandler(
			userRepo,
			codeRepo,
			tokensRepo,
		),
		TokensRepo: tokensRepo,
	}

	e := app.NewRouter()
	app.AddMiddleware(e)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", webPort)))
}
