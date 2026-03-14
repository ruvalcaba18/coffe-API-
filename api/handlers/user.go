package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/internal/middleware"
	userStorePackage "coffeebase-api/internal/store/user"
	"encoding/json"
	outputFormatting "fmt"
	"io"
	webServer "net/http"
	"os"
	"path/filepath"
	stringManipulation "strings"
	"time"
)

type UserHandler struct {
	UserStore *userStorePackage.Store
}

func (userHandler *UserHandler) UpdateProfile(responseWriter webServer.ResponseWriter, httpRequest *webServer.Request) {
	currentUserID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	var updateRequest dto.UpdateProfileRequest
	if err := json.NewDecoder(httpRequest.Body).Decode(&updateRequest); err != nil {
		webServer.Error(responseWriter, "Invalid request body", webServer.StatusBadRequest)
		return
	}

	user, err := userHandler.UserStore.GetByID(currentUserID)
	if err != nil {
		webServer.Error(responseWriter, "User not found", webServer.StatusNotFound)
		return
	}

	if updateRequest.FirstName != "" {
		user.FirstName = updateRequest.FirstName
	}
	if updateRequest.LastName != "" {
		user.LastName = updateRequest.LastName
	}
	if updateRequest.Language != "" {
		user.Language = updateRequest.Language
	}
	if updateRequest.Birthday != "" {
		birthday, err := time.Parse("2006-01-02", updateRequest.Birthday)
		if err == nil {
			user.Birthday = birthday
		}
	}

	if err := userHandler.UserStore.Update(&user); err != nil {
		webServer.Error(responseWriter, "Failed to update profile", webServer.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dto.MapUserToResponse(user))
}

func (userHandler *UserHandler) UpdateLanguage(responseWriter webServer.ResponseWriter, httpRequest *webServer.Request) {
	currentUserID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	var languageUpdateRequest dto.UpdateLanguageRequest
	decodingError := json.NewDecoder(httpRequest.Body).Decode(&languageUpdateRequest)
	if decodingError != nil {
		webServer.Error(responseWriter, "Invalid request body", webServer.StatusBadRequest)
		return
	}

	validLanguages := map[string]bool{
		"es":  true,
		"en":  true,
		"fr":  true,
		"de":  true,
		"gsw": true,
	}

	if !validLanguages[languageUpdateRequest.Language] {
		webServer.Error(responseWriter, "Invalid language. Supported: es, en, fr, de, gsw", webServer.StatusBadRequest)
		return
	}

	updateError := userHandler.UserStore.UpdateLanguage(currentUserID, languageUpdateRequest.Language)
	if updateError != nil {
		webServer.Error(responseWriter, "Internal server error", webServer.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(webServer.StatusOK)
	json.NewEncoder(responseWriter).Encode(map[string]string{"message": "Language updated successfully"})
}

func (userHandler *UserHandler) UploadAvatar(responseWriter webServer.ResponseWriter, httpRequest *webServer.Request) {
	currentUserID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	const maximumAllowedFileSize = 2 << 20
	httpRequest.ParseMultipartForm(maximumAllowedFileSize)

	uploadedFile, fileHeader, retrievalError := httpRequest.FormFile("avatar")
	if retrievalError != nil {
		webServer.Error(responseWriter, "Error retrieving the file", webServer.StatusBadRequest)
		return
	}
	defer uploadedFile.Close()

	if fileHeader.Size > maximumAllowedFileSize {
		webServer.Error(responseWriter, "File is too large. Maximum size allowed is 2MB.", webServer.StatusRequestEntityTooLarge)
		return
	}
	uploadDirectoryPath := "./uploads/avatars"
	os.MkdirAll(uploadDirectoryPath, os.ModePerm)
	fileExtension := filepath.Ext(fileHeader.Filename)
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExtensions[fileExtension] {
		webServer.Error(responseWriter, "Invalid file type. Only JPG, PNG, GIF and WEBP are allowed.", webServer.StatusBadRequest)
		return
	}

	typeDetectionBuffer := make([]byte, 512)
	if _, readingError := uploadedFile.Read(typeDetectionBuffer); readingError != nil {
		webServer.Error(responseWriter, "Error reading file", webServer.StatusInternalServerError)
		return
	}
	detectedContentType := webServer.DetectContentType(typeDetectionBuffer)
	if !stringManipulation.HasPrefix(detectedContentType, "image/") {
		webServer.Error(responseWriter, "File is not a valid image", webServer.StatusBadRequest)
		return
	}
	uploadedFile.Seek(0, io.SeekStart)

	uniqueFilename := outputFormatting.Sprintf("%d_%d%s", currentUserID, time.Now().Unix(), fileExtension)
	targetFilePath := filepath.Join(uploadDirectoryPath, uniqueFilename)

	destinationFile, creationError := os.Create(targetFilePath)
	if creationError != nil {
		webServer.Error(responseWriter, "Error saving the file", webServer.StatusInternalServerError)
		return
	}
	defer destinationFile.Close()

	if _, copyingError := io.Copy(destinationFile, uploadedFile); copyingError != nil {
		webServer.Error(responseWriter, "Error saving the file", webServer.StatusInternalServerError)
		return
	}

	avatarPublicURL := outputFormatting.Sprintf("/uploads/avatars/%s", uniqueFilename)
	savingError := userHandler.UserStore.UpdateAvatar(currentUserID, avatarPublicURL)
	if savingError != nil {
		webServer.Error(responseWriter, "Error updating user profile", webServer.StatusInternalServerError)
		return
	}

	json.NewEncoder(responseWriter).Encode(map[string]string{"avatar_url": avatarPublicURL})
}

func (userHandler *UserHandler) GetProfile(responseWriter webServer.ResponseWriter, httpRequest *webServer.Request) {
	currentUserID := httpRequest.Context().Value(middleware.UserIDKey).(int)

	userInstance, fetchError := userHandler.UserStore.GetByID(currentUserID)
	if fetchError != nil {
		webServer.Error(responseWriter, "User not found", webServer.StatusNotFound)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode(dto.MapUserToResponse(userInstance))
}
