package handlers

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/internal/middleware"
	"coffeebase-api/internal/store/user"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UserHandler struct {
	Store *user.Store
}

func (h *UserHandler) UpdateLanguage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req dto.UpdateLanguageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate language
	validLanguages := map[string]bool{
		"es":  true,
		"en":  true,
		"fr":  true,
		"de":  true,
		"gsw": true,
	}

	if !validLanguages[req.Language] {
		http.Error(w, "Invalid language. Supported: es, en, fr, de, gsw", http.StatusBadRequest)
		return
	}

	if err := h.Store.UpdateLanguage(userID, req.Language); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Language updated successfully"})
}

func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	// Limit upload size (2MB - considered normal for profile pictures)
	const MaxSize = 2 << 20
	r.ParseMultipartForm(MaxSize)

	file, header, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Explicit size check
	if header.Size > MaxSize {
		http.Error(w, "File is too large. Maximum size allowed is 2MB.", http.StatusRequestEntityTooLarge)
		return
	}

	// Create uploads directory if not exists
	uploadDir := "./uploads/avatars"
	os.MkdirAll(uploadDir, os.ModePerm)

	// Generate unique filename
	ext := filepath.Ext(header.Filename)

	// 1. Validate extension
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[filepath.Ext(header.Filename)] {
		http.Error(w, "Invalid file type. Only JPG, PNG, GIF and WEBP are allowed.", http.StatusBadRequest)
		return
	}

	// 2. Validate MIME type for security
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	fileType := http.DetectContentType(buff)
	if !strings.HasPrefix(fileType, "image/") {
		http.Error(w, "File is not a valid image", http.StatusBadRequest)
		return
	}
	// Reset file pointer after reading buff
	file.Seek(0, io.SeekStart)

	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Save file to disk
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}

	avatarURL := fmt.Sprintf("/uploads/avatars/%s", filename)
	if err := h.Store.UpdateAvatar(userID, avatarURL); err != nil {
		http.Error(w, "Error updating user profile", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"avatar_url": avatarURL})
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	u, err := h.Store.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.MapUserToResponse(u))
}
