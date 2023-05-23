package handlers

import (
	"errors"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func (h *BaseHandler) GetMessages(c echo.Context) error {
	user := c.Get("user").(*models.User)
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
	recipientIDStr := c.QueryParam("recipientId")
	if recipientIDStr == "" {
		err := errors.New("invalid recipientId, recipientId is required")
		return ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
	}
	recipientID, err := strconv.Atoi(recipientIDStr)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid recipientId, recipientId must be integer", err)
	}

	from := 0
	fromStr := c.QueryParam("from")
	if fromStr != "" {
		var err error
		from, err = strconv.Atoi(fromStr)
		if err != nil {
			return ErrorResponse(c, http.StatusBadRequest, "invalid limit, limit must be integer", err)
		}
	}

	messages, err := h.messageRepo.GetMessages(user.ID, recipientID, from, limit)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to get messages", err)
	}
	return SuccessResponse(c, http.StatusOK, "messages found successfully", messages)
}
