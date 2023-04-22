package handlers

import (
	"errors"
	"github.com/aerosystems/nix-junior-chat-back/internal/models"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"image"
	"io"
	"net/http"
	"os"
)

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
	user.Image = uuid.New().String() + ".png"

	dst, err := os.Create(path + user.Image)
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
