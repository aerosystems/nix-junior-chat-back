package handlers

import (
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	ChatService "github.com/aerosystems/nix-junior-chat-back/internal/services/chat_service"
	"github.com/aerosystems/nix-junior-chat-back/pkg/redisclient"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"reflect"
	"strconv"
)

type MessageResponseBody struct {
	Content     string `json:"content" example:"bla-bla-bla"`
	RecipientID int    `json:"recipientId" example:"1"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // TODO: remove this line in production
}

// Chat godoc
// @Summary Chat [WebSocket]
// @Description Chat with users based on WebSocket
// @Tags chat
// @Param token query string true "Access JWT Token"
// @Param chat body MessageResponseBody true "body should contain content and recipient_id for sending message"
// @Failure 401 {object} Response
// @Router /ws/chat [get]
func (h *BaseHandler) Chat(c echo.Context) error {
	clientREDIS := redisclient.NewClient()
	token := c.QueryParam("token")
	accessTokenClaims, err := h.tokenService.DecodeAccessToken(token)
	if err != nil {
		return ErrorResponse(c, 401, "invalid token", err)
	}

	_, err = h.tokenService.GetCacheValue(accessTokenClaims.AccessUUID)
	if err != nil {
		return ErrorResponse(c, 401, "invalid token", err)
	}

	user, err := h.userRepo.FindByID(accessTokenClaims.UserID)
	if err != nil {
		return ErrorResponse(c, 401, "user not found", err)
	}

	c.Logger().Info(fmt.Sprintf("client %d connected", user.ID))
	user.IsOnline = true
	if err := h.userRepo.Update(user); err != nil {
		c.Logger().Error(err.Error())
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	if err := ChatService.OnConnect(user, ws, clientREDIS); err != nil {
		ChatService.HandleWSError(err, "error sending message", ws)
	}

	closeCh := ChatService.OnDisconnect(user, ws)

	ChatService.OnChannelMessage(ws, h.messageRepo, user)

loop:
	for {
		select {
		case <-closeCh:
			user.IsOnline = false
			if err := h.userRepo.Update(user); err != nil {
				c.Logger().Error(err.Error())
			}
			break loop
		default:
			ChatService.OnClientMessage(ws, clientREDIS, h.messageRepo, h.chatRepo, user)
		}
	}

	return nil
}

// CreateChat godoc
// @Summary Create Chat by User ID
// @Tags chat
// @Accept  json
// @Produce application/json
// @Param	user_id	path	int	true	"User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.Chat}
// @Failure 400 {object} Response
// @Failure 409 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/{user_id}/chat [post]
func (h *BaseHandler) CreateChat(c echo.Context) error {
	user := c.Get("user").(*models.User)
	chatUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid user id", err)
	}

	if user.ID == chatUserID {
		err := fmt.Errorf("user can't create chat with himself")
		return ErrorResponse(c, http.StatusBadRequest, "user can't create chat with himself", err)
	}

	chatUser, err := h.userRepo.FindByID(chatUserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid user", err)
	}
	if chatUser == nil || reflect.DeepEqual(*chatUser, models.User{}) {
		err := fmt.Errorf("user with id %d not found", chatUserID)
		return ErrorResponse(c, http.StatusBadRequest, "user not found", err)
	}

	chat, err := h.chatRepo.FindPrivateChatByUsersArray([]*models.User{user, chatUser})
	if err != nil && err != gorm.ErrRecordNotFound {
		return ErrorResponse(c, http.StatusInternalServerError, "error while creating chat", err)
	}

	if chat != nil {
		err := fmt.Errorf("private chat between %d & %d Users already exists", user.ID, chatUser.ID)
		return ErrorResponse(c, http.StatusConflict, "chat already exists", err)
	}

	var newChat = models.Chat{
		Type: "private",
		Users: []*models.User{
			user,
			chatUser,
		},
	}
	if err := h.chatRepo.Create(&newChat); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error while creating chat", err)
	}
	return SuccessResponse(c, http.StatusCreated, "chat created successfully", newChat)
}

// GetChat godoc
// @Summary Get Chat by User ID
// @Tags chat
// @Accept  json
// @Produce application/json
// @Param	user_id	path	int	true	"User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.Chat}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/{user_id}/chat [get]
func (h *BaseHandler) GetChat(c echo.Context) error {
	user := c.Get("user").(*models.User)
	chatUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid user id", err)
	}
	if user.ID == chatUserID {
		err := fmt.Errorf("user can't create chat with himself")
		return ErrorResponse(c, http.StatusBadRequest, "user can't create chat with himself", err)
	}

	chatUser, err := h.userRepo.FindByID(chatUserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid user", err)
	}
	if chatUser == nil || reflect.DeepEqual(*chatUser, models.User{}) {
		err := fmt.Errorf("user with id %d not found", chatUserID)
		return ErrorResponse(c, http.StatusBadRequest, "user not found", err)
	}

	chat, err := h.chatRepo.FindPrivateChatByUsersArray([]*models.User{user, chatUser})
	fmt.Println(chat)
	if err != nil && err != gorm.ErrRecordNotFound {
		return ErrorResponse(c, http.StatusInternalServerError, "error while searching chat", err)
	}
	if err == gorm.ErrRecordNotFound {
		return ErrorResponse(c, http.StatusNotFound, "chat not found", err)
	}
	if reflect.DeepEqual(*chat, models.Chat{}) {
		err := fmt.Errorf("private chat between %d & %d Users not found", user.ID, chatUser.ID)
		return ErrorResponse(c, http.StatusNotFound, "chat not found", err)
	}
	return SuccessResponse(c, http.StatusOK, "chat successfully found", chat)
}

// DeleteChat godoc
// @Summary Delete Chat by ChatID
// @Tags chat
// @Accept  json
// @Produce application/json
// @Param	chat_id	path	int	true	"Chat ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 403 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/chat/{chat_id} [delete]
func (h *BaseHandler) DeleteChat(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("chat_id")
	chatID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid chat's userId", err)
	}

	chat, err := h.chatRepo.FindByID(chatID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return ErrorResponse(c, http.StatusInternalServerError, "error while deleting chat", err)
	}
	if err == gorm.ErrRecordNotFound {
		return ErrorResponse(c, http.StatusNotFound, "chat not found", err)
	}

	if chat == nil || reflect.DeepEqual(*chat, models.Chat{}) {
		err := fmt.Errorf("private chat between %d not found", chatID)
		return ErrorResponse(c, http.StatusNotFound, "chat not found", err)
	}

	if chat.Type != "private" {
		return ErrorResponse(c, http.StatusForbidden, "does not have access to delete this chat", nil)
	}

	for _, item := range chat.Users {
		if item.ID == user.ID {
			err := h.chatRepo.Delete(chat)
			if err != nil {
				return ErrorResponse(c, http.StatusInternalServerError, "failed to delete chat", err)
			}

			return SuccessResponse(c, http.StatusOK, "successfully deleted chat", nil)
		}
	}

	err = fmt.Errorf("does not have access to delete this chat")
	return ErrorResponse(c, http.StatusForbidden, "does not have access to delete this chat", err)
}
