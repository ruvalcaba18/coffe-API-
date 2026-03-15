package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	"coffeebase-api/internal/middleware"
	usermodel "coffeebase-api/internal/models/user"
	"coffeebase-api/internal/store/user"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UserHandler struct {
	userStore user.Store
}

// --- Public ---

func NewUserHandler(userStore user.Store) *UserHandler {
	return &UserHandler{
		userStore: userStore,
	}
}

func (userHandler *UserHandler) UpdateProfile(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	currentUserID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	var request dto.UpdateProfileRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	userInstance, error := userHandler.userStore.GetByID(httpRequest.Context(), currentUserID)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrUserNotFound)
		return
	}

	userHandler.applyProfileUpdates(&userInstance, request)

	if error := userHandler.userStore.Update(httpRequest.Context(), &userInstance); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, dto.MapUserToResponse(userInstance))
}

func (userHandler *UserHandler) UpdateLanguage(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	currentUserID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	var request dto.UpdateLanguageRequest
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	if !isValidLanguage(request.Language) {
		response.SendError(responseWriter, apperrors.ErrInvalidRequest)
		return
	}

	if error := userHandler.userStore.UpdateLanguage(httpRequest.Context(), currentUserID, request.Language); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, map[string]string{"message": "Language updated successfully"})
}

func (userHandler *UserHandler) UploadAvatar(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	currentUserID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	uploadedFile, fileHeader, error := userHandler.retrieveAndValidateFile(httpRequest)
	if error != nil {
		response.SendError(responseWriter, error)
		return
	}
	defer uploadedFile.Close()

	avatarURL, error := userHandler.processAndSaveAvatar(currentUserID, uploadedFile, fileHeader)
	if error != nil {
		response.SendError(responseWriter, error)
		return
	}

	if error := userHandler.userStore.UpdateAvatar(httpRequest.Context(), currentUserID, avatarURL); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, map[string]string{"avatar_url": avatarURL})
}

func (userHandler *UserHandler) GetProfile(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	currentUserID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	userInstance, error := userHandler.userStore.GetByID(httpRequest.Context(), currentUserID)
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrUserNotFound)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, dto.MapUserToResponse(userInstance))
}

// --- Private ---

func (userHandler *UserHandler) applyProfileUpdates(userInstance *usermodel.User, request dto.UpdateProfileRequest) {
	if request.FirstName != "" {
		userInstance.FirstName = request.FirstName
	}
	if request.LastName != "" {
		userInstance.LastName = request.LastName
	}
	if request.Language != "" {
		userInstance.Language = request.Language
	}
	if request.Birthday != "" {
		if birthday, error := time.Parse("2006-01-02", request.Birthday); error == nil {
			userInstance.Birthday = birthday
		}
	}
}

func (userHandler *UserHandler) retrieveAndValidateFile(httpRequest *http.Request) (multipart.File, *multipart.FileHeader, error) {
	const maximumAllowedFileSize = 2 << 20 
	httpRequest.ParseMultipartForm(maximumAllowedFileSize)

	file, fileHeader, error := httpRequest.FormFile("avatar")
	if error != nil {
		return nil, nil, apperrors.ErrInvalidRequest
	}

	if fileHeader.Size > maximumAllowedFileSize {
		file.Close()
		return nil, nil, fmt.Errorf("file is too large, max 2MB")
	}

	return file, fileHeader, nil
}

func (userHandler *UserHandler) processAndSaveAvatar(userID int, file io.Reader, header *multipart.FileHeader) (string, error) {
	extension := filepath.Ext(header.Filename)
	if !isAllowedImageExtension(extension) {
		return "", fmt.Errorf("invalid file type")
	}

	uploadDir := "./uploads/avatars"
	os.MkdirAll(uploadDir, os.ModePerm)

	uniqueName := fmt.Sprintf("%d_%d%s", userID, time.Now().Unix(), extension)
	targetPath := filepath.Join(uploadDir, uniqueName)

	dest, error := os.Create(targetPath)
	if error != nil {
		return "", apperrors.ErrInternalServerError
	}
	defer dest.Close()

	if _, error := io.Copy(dest, file); error != nil {
		return "", apperrors.ErrInternalServerError
	}

	return fmt.Sprintf("/uploads/avatars/%s", uniqueName), nil
}

func isAllowedImageExtension(ext string) bool {
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	return allowed[strings.ToLower(ext)]
}
