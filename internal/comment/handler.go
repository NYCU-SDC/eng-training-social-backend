package comment

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
	Content string `json:"content" validate:"required"`
}
type Response struct {
	ID         uuid.UUID         `json:"id"`
	Content    string            `json:"content"`
	AuthorID   uuid.UUID         `json:"authorId"`
	CreatedAt  string            `json:"createdAt"`
	UpdatedAt  string            `json:"updatedAt"`
	ReactionMe reaction.Response `json:"reactionMe"`
}

type Store interface {
	GetAllByPostID(ctx context.Context, postID uuid.UUID) ([]Comment, error)
	GetByID(ctx context.Context, id uuid.UUID) (Comment, error)
	Create(ctx context.Context, content string, postID uuid.UUID, userID uuid.UUID) (Comment, error)
	Update(ctx context.Context, id uuid.UUID, content string) (Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type reactionStore interface {
	GetByCommentIDAndUserID(ctx context.Context, commentID, userID uuid.UUID) (reaction.ReactionType, error)
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

func (h *Handler) GetAllByPostIDHandler(w http.ResponseWriter, r *http.Request) {
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

	comments, err := h.store.GetAllByPostID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get all posts", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get posts")
		return
	}

	response := make([]Response, len(comments))
	for i, comment := range comments {
		commentReaction, err := h.reactionStore.GetByCommentIDAndUserID(r.Context(), comment.ID, jwtUser.ID)
		if err != nil {
			h.logger.Error("Failed to get reaction by comment ID and user ID", zap.Error(err))
			internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get reaction")
			return
		}

		response[i] = Response{
			ID:        comment.ID,
			Content:   comment.Content.String,
			AuthorID:  comment.AuthorID.Bytes,
			CreatedAt: comment.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: comment.UpdatedAt.Time.Format(time.RFC3339),
			ReactionMe: reaction.Response{
				Reaction: commentReaction,
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

	comment, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get comment by ID", zap.Error(err), zap.String("id", id.String()))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get comment")
		return
	}

	commentReaction, err := h.reactionStore.GetByCommentIDAndUserID(r.Context(), comment.ID, jwtUser.ID)
	if err != nil {
		h.logger.Error("Failed to get reaction by comment ID and user ID", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get reaction")
		return
	}

	response := Response{
		ID:        comment.ID,
		Content:   comment.Content.String,
		AuthorID:  comment.AuthorID.Bytes,
		CreatedAt: comment.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: comment.UpdatedAt.Time.Format(time.RFC3339),
		ReactionMe: reaction.Response{
			Reaction: commentReaction,
		},
	}

	internal.WriteJSONResponse(w, http.StatusOK, response)
	return
}

func (h *Handler) CreateHandler(w http.ResponseWriter, r *http.Request) {
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

	jwtUser, err := jwt.GetUserFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user from context", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	comment, err := h.store.Create(r.Context(), request.Content, id, jwtUser.ID)
	if err != nil {
		h.logger.Error("Failed to create comment", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	response := Response{
		ID:        comment.ID,
		Content:   comment.Content.String,
		AuthorID:  comment.AuthorID.Bytes,
		CreatedAt: comment.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: comment.UpdatedAt.Time.Format(time.RFC3339),
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

	comment, err := h.store.Update(r.Context(), id, request.Content)
	if err != nil {
		h.logger.Error("Failed to update comment", zap.Error(err), zap.String("id", id.String()))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to update comment")
		return
	}

	commentReaction, err := h.reactionStore.GetByCommentIDAndUserID(r.Context(), comment.ID, jwtUser.ID)
	if err != nil {
		h.logger.Error("Failed to get reaction by comment ID and user ID", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get reaction")
		return
	}

	response := Response{
		ID:        comment.ID,
		Content:   comment.Content.String,
		AuthorID:  comment.AuthorID.Bytes,
		CreatedAt: comment.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: comment.UpdatedAt.Time.Format(time.RFC3339),
		ReactionMe: reaction.Response{
			Reaction: commentReaction,
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
		h.logger.Error("Failed to delete comment", zap.Error(err), zap.String("id", id.String()))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	return
}
