package handlers

import (
	"errors"
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	TokenService "github.com/aerosystems/nix-junior-chat-back/internal/services/token_service"
	"github.com/aerosystems/nix-junior-chat-back/pkg/validators"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type RegistrationRequestBody struct {
	Username string `json:"username" example:"username"`
	Password string `json:"password" example:"P@ssw0rd"`
}

type LoginRequestBody struct {
	Username string `json:"username" example:"username"`
	Password string `json:"password" example:"P@ssw0rd"`
}

// Registration godoc
// @Summary registration user by credentials
// @Description Username should contain:
// @Description - lower, upper case latin letters and digits
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Description Password should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param registration body RegistrationRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/auth/register [post]
func (h *BaseHandler) Registration(c echo.Context) error {
	var requestPayload RegistrationRequestBody

	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if err := validators.ValidateUsername(requestPayload.Username); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid username", err)
	}

	err := validators.ValidatePassword(requestPayload.Password)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid password", err)
	}

	//checking if username is existing
	user, _ := h.userRepo.FindByUsername(requestPayload.Username)
	if user != nil {
		err = fmt.Errorf("user with username %s already exists", requestPayload.Username)
		return ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
	}

	// hashing password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestPayload.Password), 12)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error hashing password", err)
	}

	// creating new user
	newUser := models.User{
		Username: requestPayload.Username,
		Password: string(hashedPassword),
		Image:    os.Getenv("URL_PREFIX_IMAGES") + "default.png",
	}
	err = h.userRepo.Create(&newUser)

	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating user", err)
	}

	// TODO make & run methods for confirm user account

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("user with username: %s successfully created", requestPayload.Username), nil)
}

// Login godoc
// @Summary login user by credentials
// @Description Username should contain:
// @Description - lower, upper case latin letters and digits
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Description Password should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Description Response contain pair JWT token_service, use /v1/token_service/refresh for updating them
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param login body LoginRequestBody true "raw request body"
// @Success 200 {object} Response{data=TokensResponseBody}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/auth/login [post]
func (h *BaseHandler) Login(c echo.Context) error {
	var requestPayload LoginRequestBody

	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if err := validators.ValidateUsername(requestPayload.Username); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid username", err)
	}

	if err := validators.ValidatePassword(requestPayload.Password); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid password", err)
	}

	//checking if user is existing
	user, err := h.userRepo.FindByUsername(requestPayload.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		return ErrorResponse(c, http.StatusBadRequest, "error while finding user", err)
	}
	if user == nil {
		err := fmt.Errorf("user with username %s does not exist", requestPayload.Username)
		return ErrorResponse(c, http.StatusNotFound, "user not found", err)
	}

	valid, err := h.userRepo.PasswordMatches(user, requestPayload.Password)
	if err != nil || !valid {
		err := errors.New("invalid credentials")
		return ErrorResponse(c, http.StatusUnauthorized, err.Error(), err)
	}

	// create pair JWT token_service
	ts, err := h.tokenService.CreateToken(user.ID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating token_service", err)
	}

	// add refresh token UUID to cache
	if err = h.tokenService.CreateCacheKey(user.ID, ts); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating cache token_service", err)
	}

	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("user %s is logged in", requestPayload.Username), tokens)
}

// Logout godoc
// @Summary logout user
// @Tags auth
// @Accept  json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/auth/logout [post]
func (h *BaseHandler) Logout(c echo.Context) error {
	// receive AccessToken Claims from context middleware
	accessTokenClaims, ok := c.Get("user").(*TokenService.AccessTokenClaims)
	if !ok {
		err := errors.New("internal transport token error")
		return ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
	}

	err := h.tokenService.DropCacheTokens(*accessTokenClaims)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error clearing cache token", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("user %s successfully logged out", accessTokenClaims.AccessUUID), accessTokenClaims)
}
