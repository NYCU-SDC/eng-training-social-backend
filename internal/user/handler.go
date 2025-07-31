package user

import (
	"context"
	"github.com/NYCU-SDC/eng-training-social-backend/internal"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Response struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

type Store interface {
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
}

type Handler struct {
	logger    *zap.Logger
	validator *validator.Validate
	store     Store
}

func NewHandler(logger *zap.Logger, validator *validator.Validate, store Store) *Handler {
	return &Handler{
		logger:    logger,
		validator: validator,
		store:     store,
	}
}

func (h Handler) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	id, err := internal.ParseUUID(pathID)
	if err != nil {
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	user, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user by ID", zap.String("id", id.String()), zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	response := Response{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Time.Format(time.RFC3339),
	}

	internal.WriteJSONResponse(w, http.StatusOK, response)
	return
}
