package handlers

import (
	"errors"
	"fmt"
	"github.com/aerosystems/nix-junior-chat-back/internal/helpers"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"image"
	"io"
	"net/http"
	"os"
	"strconv"
)

type UpdateUsernameRequestBody struct {
	Username string `json:"username" example:"username"`
}

type UpdatePasswordRequestBody struct {
	NewPassword string `json:"newPassword" example:"NewP@ssw0rd"`
	OldPassword string `json:"oldPassword" example:"OldP@ssw0rd"`
}

// User godoc
// @Summary Get user data
// @Description Get user data
// @Tags user
// @Accept  json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} Response{data=UserResponse}
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user [get]
func (h *BaseHandler) User(c echo.Context) error {
	user := c.Get("user").(*models.User)

	return SuccessResponse(c, http.StatusOK, "successfully found user data", user)
}

// Block godoc
// @Summary Block user
// @Tags relationship
// @Accept  json
// @Produce application/json
// @Param	id	path	int	true	"Blocked User ID"
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/block/{id} [post]
func (h *BaseHandler) Block(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("id")
	blockedUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid blocked userId", err)
	}
	blockedUser, err := h.userRepo.FindByID(blockedUserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if blockedUser == nil {
		err := fmt.Errorf("user with id %d not found", blockedUserID)
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if blockedUser.ID == user.ID {
		err := fmt.Errorf("user with id %d is the same as current user", blockedUserID)
		return ErrorResponse(c, http.StatusBadRequest, "you can't block yourself", err)
	}
	for _, item := range user.BlockedUsers {
		if item.ID == blockedUser.ID {
			err := fmt.Errorf("user with id %d is already blocked", blockedUserID)
			return ErrorResponse(c, http.StatusBadRequest, "user is already blocked", err)
		}
	}

	user.BlockedUsers = append(user.BlockedUsers, blockedUser)

	if err := h.userRepo.Update(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating user", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("successfully blacklisted user %s (id: %d)", blockedUser.Username, blockedUser.ID), user)
}

// DeleteChat godoc
// @Summary DeleteChat user
// @Tags relationship
// @Accept  json
// @Produce application/json
// @Param	id	path	int	true	"Chat User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/chat/{id} [delete]
func (h *BaseHandler) DeleteChat(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("id")
	chatUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid chat's userId", err)
	}

	for i, item := range user.Chats {
		if item.ID == chatUserID {
			chatUsers := append(user.Chats[:i], user.Chats[i+1:]...)
			err := h.userRepo.ReplaceChatUsers(user, chatUsers)
			if err != nil {
				return ErrorResponse(c, http.StatusInternalServerError, "failed to delete chat with user", err)
			}

			return SuccessResponse(c, http.StatusOK, "successfully deleted chat with user", user)
		}
	}
	err = fmt.Errorf("chat with userId %d does not exist", chatUserID)
	return ErrorResponse(c, http.StatusNotFound, "chat with user does not exist", err)
}

// UploadImage godoc
// @Summary Upload user image
// @Description Uploading user image as file by form-data "image"
// @Tags user
// @Param image formData file true "User image file. The preferred size is 315x315px because the image will resize to 315x315px. Max size: 2MB, Allowed types: 'jpg', 'jpeg', 'png', 'gif'"
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/upload-image [post]
func (h *BaseHandler) UploadImage(c echo.Context) error {
	user := c.Get("user").(*models.User)
	// Source
	file, err := c.FormFile("image")
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "image is required", err)
	}

	if file.Size > 2*1024*1024 {
		err := errors.New("image size is too big")
		return ErrorResponse(c, http.StatusBadRequest, "image size is too big", err)
	}

	src, err := file.Open()
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error opening image", err)
	}
	defer src.Close()

	inputImage, _, err := image.Decode(src)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error decoding image", err)
	}

	width := inputImage.Bounds().Size().X
	height := inputImage.Bounds().Size().Y
	// Resize the cropped image to width = 315px preserving the aspect ratio.
	if width >= height {
		inputImage = imaging.Resize(inputImage, 0, 315, imaging.Lanczos)
	} else {
		inputImage = imaging.Resize(inputImage, 315, 0, imaging.Lanczos)
	}

	// Crop the original image to 315x315px size using the center anchor.
	outputImage := imaging.CropAnchor(inputImage, 315, 315, imaging.Center)

	// Destination
	path := os.Getenv("IMAGES_DIRECTORY_PATH")
	if path == "" {
		err := errors.New("images directory path is not set")
		return ErrorResponse(c, http.StatusInternalServerError, "error searching image", err)
	}
	oldImageName := user.Image
	image := uuid.New().String() + ".png"
	user.Image = os.Getenv("URL_PREFIX_IMAGES") + image

	dst, err := os.Create(path + image)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error handling image", err)
	}
	defer dst.Close()

	if err = imaging.Encode(dst, outputImage, imaging.PNG); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error encoding image", err)
	}

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error moving image", err)
	}

	if err := h.userRepo.Update(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error saving image", err)
	}

	if oldImageName != "" {
		err := os.Remove(path + oldImageName)
		if err != nil {
			c.Logger().Error(err)
		}
	}

	return SuccessResponse(c, http.StatusOK, "Image uploaded successfully", nil)
}

// UpdateUsername godoc
// @Summary update username
// @Description Username should contain:
// @Description - lower, upper case latin letters and digits
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Tags user
// @Accept  json
// @Produce application/json
// @Param username body UpdateUsernameRequestBody true "raw request body"
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/update-username [put]
func (h *BaseHandler) UpdateUsername(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	fmt.Println(user)
	if !ok {
		err := errors.New("internal transport token error")
		return ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
	}

	var requestPayload UpdateUsernameRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if err := helpers.ValidateUsername(requestPayload.Username); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "username is incorrect", err)
	}

	if tmpUser, _ := h.userRepo.FindByUsername(requestPayload.Username); tmpUser != nil {
		err := errors.New("username is already taken")
		return ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
	}

	user.Username = requestPayload.Username

	if err := h.userRepo.Update(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating user", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("username successfuly updated to %s", user.Username), nil)
}

// UpdatePassword godoc
// @Summary update password
// @Description OldPassword/NewPassword should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Tags user
// @Accept  json
// @Produce application/json
// @Param password body UpdatePasswordRequestBody true "raw request body"
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/update-password [put]
func (h *BaseHandler) UpdatePassword(c echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		err := errors.New("internal transport token error")
		return ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
	}

	var requestPayload UpdatePasswordRequestBody

	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if err := helpers.ValidatePassword(requestPayload.OldPassword); err != nil {
		prefixErr := errors.New("old password is incorrect. ")
		err = fmt.Errorf("%w%v", prefixErr, err)
		return ErrorResponse(c, http.StatusBadRequest, prefixErr.Error(), err)
	}

	if err := helpers.ValidatePassword(requestPayload.NewPassword); err != nil {
		prefixErr := errors.New("new password is incorrect. ")
		err = fmt.Errorf("%w%v", prefixErr, err)
		return ErrorResponse(c, http.StatusBadRequest, prefixErr.Error(), err)
	}

	ok, err := h.userRepo.PasswordMatches(user, requestPayload.OldPassword)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error matching passwords", err)
	}
	if !ok {
		err := errors.New("old password is incorrect")
		return ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
	}

	if err := h.userRepo.ResetPassword(user, requestPayload.NewPassword); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating password", err)
	}

	return SuccessResponse(c, http.StatusOK, "password successfully updated", nil)
}

// Unfollow godoc
// @Summary Unfollow user
// @Tags relationship
// @Accept  json
// @Produce application/json
// @Param	id	path	int	true	"Followed User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/follow/{id} [delete]
func (h *BaseHandler) Unfollow(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("id")
	followedUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid unfollowed userId", err)
	}

	for i, item := range user.FollowedUsers {
		if item.ID == followedUserID {
			followedUsers := append(user.FollowedUsers[:i], user.FollowedUsers[i+1:]...)
			err := h.userRepo.ReplaceFollowedUsers(user, followedUsers)
			if err != nil {
				return ErrorResponse(c, http.StatusInternalServerError, "failed to unfollow user", err)
			}

			return SuccessResponse(c, http.StatusOK, "successfully unfollowed user", user)
		}
	}
	err = fmt.Errorf("user with id %d is not followed", followedUserID)
	return ErrorResponse(c, http.StatusNotFound, "user is not followed", err)
}

// Unblock godoc
// @Summary Unblock user
// @Tags relationship
// @Accept  json
// @Produce application/json
// @Param	id	path	int	true	"Unblocked User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/block/{id} [delete]
func (h *BaseHandler) Unblock(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("id")
	blockedUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid unfollowed userId", err)
	}

	for i, item := range user.BlockedUsers {
		if item.ID == blockedUserID {
			followedUsers := append(user.BlockedUsers[:i], user.BlockedUsers[i+1:]...)
			err := h.userRepo.ReplaceBlockedUsers(user, followedUsers)
			if err != nil {
				return ErrorResponse(c, http.StatusInternalServerError, "failed to unblock user", err)
			}

			return SuccessResponse(c, http.StatusOK, "successfully unblocked user", user)
		}
	}
	err = fmt.Errorf("user with id %d is not blocked", blockedUserID)
	return ErrorResponse(c, http.StatusNotFound, "user is not blocked", err)
}

// Follow godoc
// @Summary Follow user
// @Tags relationship
// @Accept  json
// @Produce application/json
// @Param	id	path	int	true	"Followed User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/follow/{id} [post]
func (h *BaseHandler) Follow(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("id")
	followedUserID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}

	followedUser, err := h.userRepo.FindByID(followedUserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if followedUser == nil {
		err := fmt.Errorf("user with id %d not found", followedUserID)
		return ErrorResponse(c, http.StatusBadRequest, "invalid followed userId", err)
	}
	if followedUser.ID == user.ID {
		err := fmt.Errorf("user with id %d is the same as current user", followedUserID)
		return ErrorResponse(c, http.StatusBadRequest, "you can't follow yourself", err)
	}
	for _, item := range user.FollowedUsers {
		if item.ID == followedUser.ID {
			err := fmt.Errorf("user with id %d is already followed", followedUserID)
			return ErrorResponse(c, http.StatusBadRequest, "user is already followed", err)
		}
	}

	user.FollowedUsers = append(user.FollowedUsers, followedUser)

	if err := h.userRepo.Update(user); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error updating user", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("successfully followed user %s (id: %d)", followedUser.Username, followedUser.ID), user)
}
