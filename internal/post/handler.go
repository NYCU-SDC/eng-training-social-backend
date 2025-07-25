package post

import (
	"context"
	"github.com/NYCU-SDC/eng-training-social-backend/internal"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Request struct {
	Title   string `json:"title" validate:"required"`
	Content string `json:"content" validate:"required"`
}
type Response struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

type Store interface {
	GetAll(ctx context.Context) ([]Post, error)
	GetByID(ctx context.Context, id uuid.UUID) (Post, error)
	Create(ctx context.Context, title, content string) (Post, error)
	Update(ctx context.Context, id uuid.UUID, title, content string) (Post, error)
	Delete(ctx context.Context, id uuid.UUID) error
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

func (h *Handler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := h.store.GetAll(r.Context())
	if err != nil {
		h.logger.Error("Failed to get all posts", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get posts")
		return
	}

	response := make([]Response, len(posts))
	for i, post := range posts {
		response[i] = Response{
			ID:        post.ID,
			Title:     post.Title.String,
			Content:   post.Content.String,
			CreatedAt: post.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: post.UpdatedAt.Time.Format(time.RFC3339),
		}
	}

	internal.WriteJSONResponse(w, http.StatusOK, response)
	return
}

func (h *Handler) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	id, err := internal.ParseUUID(pathID)
	if err != nil {
		h.logger.Error("Invalid UUID", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	post, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get post by ID", zap.Error(err), zap.String("id", id.String()))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get post")
		return
	}

	response := Response{
		ID:        post.ID,
		Title:     post.Title.String,
		Content:   post.Content.String,
		CreatedAt: post.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: post.UpdatedAt.Time.Format(time.RFC3339),
	}

	internal.WriteJSONResponse(w, http.StatusOK, response)
	return
}

func (h *Handler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	var request Request
	err := internal.ParseAndValidateRequestBody(r.Context(), h.validator, r, &request)
	if err != nil {
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	post, err := h.store.Create(r.Context(), request.Title, request.Content)
	if err != nil {
		h.logger.Error("Failed to create post", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	response := Response{
		ID:        post.ID,
		Title:     post.Title.String,
		Content:   post.Content.String,
		CreatedAt: post.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: post.UpdatedAt.Time.Format(time.RFC3339),
	}

	internal.WriteJSONResponse(w, http.StatusCreated, response)
	return
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	id, err := internal.ParseUUID(pathID)
	if err != nil {
		h.logger.Error("Invalid UUID", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	var request Request
	err = internal.ParseAndValidateRequestBody(r.Context(), h.validator, r, &request)
	if err != nil {
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	post, err := h.store.Update(r.Context(), id, request.Title, request.Content)
	if err != nil {
		h.logger.Error("Failed to update post", zap.Error(err), zap.String("id", id.String()))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to update post")
		return
	}

	response := Response{
		ID:        post.ID,
		Title:     post.Title.String,
		Content:   post.Content.String,
		CreatedAt: post.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: post.UpdatedAt.Time.Format(time.RFC3339),
	}

	internal.WriteJSONResponse(w, http.StatusOK, response)
	return
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	id, err := internal.ParseUUID(pathID)
	if err != nil {
		h.logger.Error("Invalid UUID", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	err = h.store.Delete(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete post", zap.Error(err), zap.String("id", id.String()))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to delete post")
		return
	}

	return
}
