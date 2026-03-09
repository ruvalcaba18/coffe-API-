package admin

import (
	"coffeebase-api/api/dto"
	userstore "coffeebase-api/internal/store/user"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	UserStore *userstore.Store
}

func (userHandler *UserHandler) GetAll(responseWriter http.ResponseWriter, request *http.Request) {
	userList, fetchError := userHandler.UserStore.GetAll()
	if fetchError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(responseWriter).Encode(dto.MapUsersToResponse(userList))
}

func (userHandler *UserHandler) UpdateRole(responseWriter http.ResponseWriter, request *http.Request) {
	userID, _ := strconv.Atoi(chi.URLParam(request, "id"))
	var roleUpdateRequest struct {
		Role string `json:"role"`
	}
	if decodeError := json.NewDecoder(request.Body).Decode(&roleUpdateRequest); decodeError != nil {
		http.Error(responseWriter, "Invalid request body", http.StatusBadRequest)
		return
	}

	if updateError := userHandler.UserStore.UpdateRole(userID, roleUpdateRequest.Role); updateError != nil {
		http.Error(responseWriter, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}
