package reaction

import (
	"context"
	"github.com/NYCU-SDC/eng-training-social-backend/internal"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
)

type Request struct {
	Reaction ReactionType `json:"reaction"`
}

type Response struct {
	Reaction ReactionType `json:"reaction"`
}

type Store interface {
	ReactToPost(ctx context.Context, postID, userID uuid.UUID, reactionType ReactionType) (Reaction, error)
	RemoveToPost(ctx context.Context, postID, userID uuid.UUID) error
	ReactToComment(ctx context.Context, commentID, userID uuid.UUID, reactionType ReactionType) (Reaction, error)
	RemoveToComment(ctx context.Context, commentID, userID uuid.UUID) error
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

func (h *Handler) ReactToPost(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	postID, err := internal.ParseUUID(pathID)
	if err != nil {
		h.logger.Error("Invalid UUID", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	u, err := jwt.GetUserFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user from context", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	var request Request
	err = internal.ParseAndValidateRequestBody(r.Context(), h.validator, r, &request)
	if err != nil {
		h.logger.Error("Failed to parse and validate request body", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.Reaction == ReactionTypeNONE {
		err = h.store.RemoveToPost(r.Context(), postID, u.ID)
		if err != nil {
			h.logger.Error("Failed to remove reaction from post", zap.Error(err))
			internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to remove reaction")
			return
		}

		internal.WriteJSONResponse(w, http.StatusOK, Response{
			Reaction: ReactionTypeNONE,
		})
	} else {
		reaction, err := h.store.ReactToPost(r.Context(), postID, u.ID, request.Reaction)
		if err != nil {
			h.logger.Error("Failed to react to post", zap.Error(err))
			internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to react to post")
			return
		}

		internal.WriteJSONResponse(w, http.StatusOK, Response{
			Reaction: reaction.ReactionType,
		})
	}
}

func (h *Handler) ReactToComment(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	commentID, err := internal.ParseUUID(pathID)
	if err != nil {
		h.logger.Error("Invalid UUID", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	u, err := jwt.GetUserFromContext(r.Context())
	if err != nil {
		h.logger.Error("Failed to get user from context", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	var request Request
	err = internal.ParseAndValidateRequestBody(r.Context(), h.validator, r, &request)
	if err != nil {
		h.logger.Error("Failed to parse and validate request body", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.Reaction == ReactionTypeNONE {
		err = h.store.RemoveToComment(r.Context(), commentID, u.ID)
		if err != nil {
			h.logger.Error("Failed to remove reaction from comment", zap.Error(err))
			internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to remove reaction")
			return
		}

		internal.WriteJSONResponse(w, http.StatusOK, Response{
			Reaction: ReactionTypeNONE,
		})
	} else {
		reaction, err := h.store.ReactToComment(r.Context(), commentID, u.ID, request.Reaction)
		if err != nil {
			h.logger.Error("Failed to react to comment", zap.Error(err))
			internal.WriteJSONResponse(w, http.StatusInternalServerError, "Failed to react to comment")
			return
		}

		internal.WriteJSONResponse(w, http.StatusOK, Response{
			Reaction: reaction.ReactionType,
		})
	}
}
