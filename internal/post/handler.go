package post

import (
	"context"
	"github.com/NYCU-SDC/eng-training-social-backend/internal"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/jwt"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/reaction"
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
	ID         uuid.UUID         `json:"id"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	AuthorID   uuid.UUID         `json:"authorId"`
	CreatedAt  string            `json:"createdAt"`
	UpdatedAt  string            `json:"updatedAt"`
	ReactionMe reaction.Response `json:"reactionMe"`
}

type Store interface {
	GetAll(ctx context.Context) ([]Post, error)
	GetByID(ctx context.Context, id uuid.UUID) (Post, error)
	Create(ctx context.Context, title, content string, userID uuid.UUID) (Post, error)
	Update(ctx context.Context, id uuid.UUID, title, content string) (Post, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type reactionStore interface {
	GetByPostIDAndUserID(ctx context.Context, postID, userID uuid.UUID) (reaction.ReactionType, error)
}

type Handler struct {
	logger        *zap.Logger
	validator     *validator.Validate
	store         Store
	reactionStore reactionStore
}

func NewHandler(logger *zap.Logger, validator *validator.Validate, store Store, reactionStore reactionStore) *Handler {
	return &Handler{
		logger:        logger,
		validator:     validator,
		store:         store,
		reactionStore: reactionStore,
	}
}

func (h *Handler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := h.store.GetAll(r.Context())
	if err != nil {
		h.logger.Error("Failed to get all posts", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get posts")
		return
	}

	jwtUser, err := jwt.GetUserFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user from context", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	response := make([]Response, len(posts))
	for i, post := range posts {
		postReaction, err := h.reactionStore.GetByPostIDAndUserID(r.Context(), post.ID, jwtUser.ID)
		if err != nil {
			h.logger.Error("Failed to get reaction for post", zap.Error(err), zap.String("postID", post.ID.String()), zap.String("userID", jwtUser.ID.String()))
			internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get post reaction")
			return
		}

		response[i] = Response{
			ID:        post.ID,
			Title:     post.Title.String,
			Content:   post.Content.String,
			AuthorID:  post.AuthorID.Bytes,
			CreatedAt: post.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: post.UpdatedAt.Time.Format(time.RFC3339),
			ReactionMe: reaction.Response{
				Reaction: postReaction,
			},
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

	jwtUser, err := jwt.GetUserFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user from context", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	post, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get post by ID", zap.Error(err), zap.String("id", id.String()))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get post")
		return
	}

	postReaction, err := h.reactionStore.GetByPostIDAndUserID(r.Context(), post.ID, jwtUser.ID)
	if err != nil {
		h.logger.Error("Failed to get reaction for post", zap.Error(err), zap.String("postID", post.ID.String()), zap.String("userID", jwtUser.ID.String()))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get post reaction")
		return
	}

	response := Response{
		ID:        post.ID,
		Title:     post.Title.String,
		Content:   post.Content.String,
		AuthorID:  post.AuthorID.Bytes,
		CreatedAt: post.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: post.UpdatedAt.Time.Format(time.RFC3339),
		ReactionMe: reaction.Response{
			Reaction: postReaction,
		},
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

	jwtUser, err := jwt.GetUserFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user from context", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	post, err := h.store.Create(r.Context(), request.Title, request.Content, jwtUser.ID)
	if err != nil {
		h.logger.Error("Failed to create post", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to create post")
		return
	}

	response := Response{
		ID:        post.ID,
		Title:     post.Title.String,
		Content:   post.Content.String,
		AuthorID:  post.AuthorID.Bytes,
		CreatedAt: post.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: post.UpdatedAt.Time.Format(time.RFC3339),
		ReactionMe: reaction.Response{
			Reaction: reaction.ReactionTypeNONE,
		},
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

	jwtUser, err := jwt.GetUserFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user from context", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get user from context")
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

	postReaction, err := h.reactionStore.GetByPostIDAndUserID(r.Context(), post.ID, jwtUser.ID)
	if err != nil {
		h.logger.Error("Failed to get reaction for post", zap.Error(err), zap.String("postID", post.ID.String()), zap.String("userID", jwtUser.ID.String()))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get post reaction")
		return
	}

	response := Response{
		ID:        post.ID,
		Title:     post.Title.String,
		Content:   post.Content.String,
		AuthorID:  post.AuthorID.Bytes,
		CreatedAt: post.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: post.UpdatedAt.Time.Format(time.RFC3339),
		ReactionMe: reaction.Response{
			Reaction: postReaction,
		},
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
