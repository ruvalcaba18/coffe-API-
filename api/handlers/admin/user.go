package admin

import (
	"coffeebase-api/api/dto"
	"coffeebase-api/api/response"
	"coffeebase-api/internal/apperrors"
	usermodel "coffeebase-api/internal/models/user"
	userstore "coffeebase-api/internal/store/user"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userStore userstore.Store
}

// --- Public ---

func NewUserHandler(userStore userstore.Store) *UserHandler {
	return &UserHandler{
		userStore: userStore,
	}
}

func (userHandler *UserHandler) GetAll(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userList, error := userHandler.userStore.GetAll(httpRequest.Context())
	if error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	response.SendJSON(responseWriter, http.StatusOK, dto.MapUsersToResponse(userList))
}

func (userHandler *UserHandler) UpdateRole(responseWriter http.ResponseWriter, httpRequest *http.Request) {
	userIDString := chi.URLParam(httpRequest, "id")
	userID, conversionError := strconv.Atoi(userIDString)
	if conversionError != nil {
		response.SendError(responseWriter, apperrors.ErrInvalidID)
		return
	}

	var request struct {
		Role usermodel.UserRole `json:"role"`
	}
	if error := response.DecodeJSON(httpRequest, &request); error != nil {
		response.SendError(responseWriter, error)
		return
	}

	if error := userHandler.userStore.UpdateRole(httpRequest.Context(), userID, request.Role); error != nil {
		response.SendError(responseWriter, apperrors.ErrInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}
