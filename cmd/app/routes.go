package main

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/aerosystems/nix-junior-chat-back/docs" // docs are generated by Swag CLI, you have to import it
)

func (app *Config) NewRouter() *echo.Echo {
	e := echo.New()

	e.GET("/docs/*", echoSwagger.WrapHandler)

	e.POST("/v1/users/registration", app.BaseHandler.Registration)
	e.POST("/v1/users/confirmation", app.BaseHandler.Confirmation)

	e.POST("/v1/users/login", app.BaseHandler.Login)
	e.POST("/v1/users/logout", app.BaseHandler.Logout, app.AuthorizationMiddleware())

	e.POST("/v1/tokens/refresh", app.BaseHandler.RefreshToken, app.AuthorizationMiddleware())

	return e
}
