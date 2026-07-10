package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/OrioXZ/7solutions-backend-challenge/internal/domain"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/repository"
	"github.com/OrioXZ/7solutions-backend-challenge/internal/service"
)

type UserManagementService interface {
	Create(ctx context.Context, input service.CreateUserInput) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Update(ctx context.Context, id string, input service.UpdateUserInput) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}

type UserHandler struct {
	users UserManagementService
}

func NewUserHandler(users UserManagementService) *UserHandler {
	return &UserHandler{users: users}
}

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateUserRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if err := decodeSingleJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	user, err := h.users.Create(r.Context(), service.CreateUserInput{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		handleUserError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dataResponse[userResponse]{Data: toUserResponse(user)})
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	user, err := h.users.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		handleUserError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dataResponse[userResponse]{Data: toUserResponse(user)})
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil {
		handleUserError(w, err)
		return
	}

	response := make([]userResponse, 0, len(users))
	for index := range users {
		response = append(response, toUserResponse(&users[index]))
	}

	writeJSON(w, http.StatusOK, dataResponse[[]userResponse]{Data: response})
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	var request updateUserRequest
	if err := decodeSingleJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	user, err := h.users.Update(r.Context(), r.PathValue("id"), service.UpdateUserInput{
		Name:  request.Name,
		Email: request.Email,
	})
	if err != nil {
		handleUserError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dataResponse[userResponse]{Data: toUserResponse(user)})
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.users.Delete(r.Context(), r.PathValue("id")); err != nil {
		handleUserError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toUserResponse(user *domain.User) userResponse {
	return userResponse{
		ID:        user.ID.Hex(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func handleUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserIDInvalid):
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", err.Error())
	case errors.Is(err, service.ErrNameRequired),
		errors.Is(err, service.ErrEmailInvalid),
		errors.Is(err, service.ErrPasswordTooShort),
		errors.Is(err, service.ErrUpdateRequired):
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, repository.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "user was not found")
	case errors.Is(err, repository.ErrEmailAlreadyExists):
		writeError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", "email is already registered")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
	}
}
