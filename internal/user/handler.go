package user

import (
	"context"
	"errors"
	"github.com/NYCU-SDC/eng-training-social-backend/internal"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/follow"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Response struct {
	ID        uuid.UUID       `json:"id"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
	FollowMe  follow.Response `json:"followMe"`
}

type Store interface {
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
}

type followStore interface {
	ExistsByID(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	Create(ctx context.Context, followerID, followingID uuid.UUID) error
	Delete(ctx context.Context, followerID, followingID uuid.UUID) error
}

type Handler struct {
	logger      *zap.Logger
	validator   *validator.Validate
	store       Store
	followStore followStore
}

func NewHandler(logger *zap.Logger, validator *validator.Validate, store Store, followStore followStore) *Handler {
	return &Handler{
		logger:      logger,
		validator:   validator,
		store:       store,
		followStore: followStore,
	}
}

func (h Handler) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	id, err := internal.ParseUUID(pathID)
	if err != nil {
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	jwtUser, err := jwt.GetUserFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user from context", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user by ID", zap.String("id", id.String()), zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	following, err := h.followStore.ExistsByID(r.Context(), jwtUser.ID, user.ID)
	if err != nil {
		h.logger.Error("Failed to check if user is following", zap.String("followerID", jwtUser.ID.String()), zap.String("followingID", user.ID.String()), zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to check following status")
		return
	}

	response := Response{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Time.Format(time.RFC3339),
		FollowMe: follow.Response{
			Follow: following,
		},
	}

	internal.WriteJSONResponse(w, http.StatusOK, response)
	return
}

func (h Handler) FollowHandler(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	id, err := internal.ParseUUID(pathID)
	if err != nil {
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	jwtUser, err := jwt.GetUserFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user from context", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err = h.followStore.Create(r.Context(), jwtUser.ID, id)
	if err != nil {
		h.logger.Error("Failed to follow user", zap.String("followerID", jwtUser.ID.String()), zap.String("followingID", id.String()), zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to follow user")
		return
	}

	response := Response{
		ID:        id,
		Username:  jwtUser.Username,
		Email:     jwtUser.Email,
		CreatedAt: jwtUser.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: jwtUser.UpdatedAt.Time.Format(time.RFC3339),
		FollowMe: follow.Response{
			Follow: true,
		},
	}

	internal.WriteJSONResponse(w, http.StatusOK, response)
}

func (h Handler) UnfollowHandler(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	id, err := internal.ParseUUID(pathID)
	if err != nil {
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	jwtUser, err := jwt.GetUserFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user from context", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err = h.followStore.Delete(r.Context(), jwtUser.ID, id)
	if err != nil {
		h.logger.Error("Failed to unfollow user", zap.String("followerID", jwtUser.ID.String()), zap.String("followingID", id.String()), zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to unfollow user")
		return
	}

	response := Response{
		ID:        id,
		Username:  jwtUser.Username,
		Email:     jwtUser.Email,
		CreatedAt: jwtUser.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: jwtUser.UpdatedAt.Time.Format(time.RFC3339),
		FollowMe: follow.Response{
			Follow: false,
		},
	}

	internal.WriteJSONResponse(w, http.StatusOK, response)
}
