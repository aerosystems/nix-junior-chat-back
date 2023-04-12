package handlers

import (
	"errors"
	"github.com/aerosystems/nix-junior-chat-back/internal/helpers"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

// Search godoc
// @Summary search users
// @Description Search users by username with autocomplete
// @Tags search
// @Accept  json
// @Produce application/json
// @Param q query string true "query string for search by username, minimum 3 characters, maximum 40 characters"
// @Param type query string false "type of search, default: 'user', available: 'user', 'friend', 'blacklist'"
// @Param order query string false "order of search, default: 'asc', available: 'asc', 'desc'"
// @Param limit query int false "limit of search, default: '10', available: '1-1000'"
// @Success 200 {object} Response{data=[]models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/search [get]
func (h *BaseHandler) Search(c echo.Context) error {
	query := c.QueryParam("q")
	if err := helpers.ValidateUsername(query); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid query. "+err.Error(), err)
	}

	tp := c.QueryParam("type")
	if tp == "" {
		tp = "user"
	} else if tp != "user" && tp != "friend" && tp != "blacklist" {
		err := errors.New("invalid type, available types: 'user', 'friend', 'blacklist'")
		return ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
	}

	order := c.QueryParam("order")
	if order == "" {
		order = "asc"
	} else if order != "asc" && order != "desc" {
		err := errors.New("invalid order, available orders: 'asc', 'desc'")
		return ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
	}

	limit := 10
	limitStr := c.QueryParam("limit")
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return ErrorResponse(c, http.StatusBadRequest, "invalid limit, limit must be integer", err)
		}
		if limit < 1 || limit > 1000 {
			err := errors.New("invalid limit, available limits: '1-1000'")
			return ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		}
	}

	users, err := h.userRepo.FindArrayByPartUsername(query, order, limit)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "internal server error finding users", err)
	}
	if users == nil {
		err := errors.New("users not found")
		return ErrorResponse(c, http.StatusNotFound, err.Error(), err)
	}

	return SuccessResponse(c, http.StatusOK, "users found successfully", users)
}
