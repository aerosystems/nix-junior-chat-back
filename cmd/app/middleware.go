package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"github.com/aerosystems/nix-junior-chat-back/internal/handlers"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (app *Config) AddMiddleware(e *echo.Echo) {
	DefaultCORSConfig := middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
	}
	e.Use(middleware.CORSWithConfig(DefaultCORSConfig))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
}

func (app *Config) AuthTokenMiddleware() echo.MiddlewareFunc {
	AuthorizationConfig := echojwt.Config{
		SigningKey:     []byte(os.Getenv("ACCESS_SECRET")),
		ParseTokenFunc: app.ParseToken,
		ErrorHandler: func(c echo.Context, err error) error {
			return handlers.ErrorResponse(c, http.StatusUnauthorized, "invalid token", err)
		},
	}

	return echojwt.WithConfig(AuthorizationConfig)
}

func (app *Config) AuthUserMiddleware() echo.MiddlewareFunc {
	AuthorizationConfig := echojwt.Config{
		SigningKey:     []byte(os.Getenv("ACCESS_SECRET")),
		ParseTokenFunc: app.GetUser,
		ErrorHandler: func(c echo.Context, err error) error {
			return handlers.ErrorResponse(c, http.StatusUnauthorized, "invalid token", err)
		},
	}

	return echojwt.WithConfig(AuthorizationConfig)
}

func (app *Config) ParseToken(c echo.Context, auth string) (interface{}, error) {
	_ = c
	accessTokenClaims, err := app.TokensRepo.DecodeAccessToken(auth)
	if err != nil {
		return nil, err
	}

	_, err = app.TokensRepo.GetCacheValue(accessTokenClaims.AccessUUID)
	if err != nil {
		return nil, err
	}

	return accessTokenClaims, nil
}

func (app *Config) GetUser(c echo.Context, auth string) (interface{}, error) {
	_ = c
	accessTokenClaims, err := app.TokensRepo.DecodeAccessToken(auth)
	if err != nil {
		return nil, err
	}

	_, err = app.TokensRepo.GetCacheValue(accessTokenClaims.AccessUUID)
	if err != nil {
		return nil, err
	}

	user, err := app.UserRepo.FindByID(accessTokenClaims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
