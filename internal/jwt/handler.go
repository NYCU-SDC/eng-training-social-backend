package jwt

import (
	"context"
	"github.com/NYCU-SDC/eng-training-social-backend/internal"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
)

type JWTIssuer interface {
	New(ctx context.Context, user User) (string, error)
	GetUserByRefreshToken(ctx context.Context, refreshToken uuid.UUID) (User, error)
	GenerateRefreshToken(ctx context.Context, user User) (RefreshToken, error)
	InactivateRefreshToken(ctx context.Context, refreshToken uuid.UUID) error
}

type Response struct {
	AccessToken    string `json:"accessToken"`
	ExpirationTime int64  `json:"expirationTime"`
	RefreshToken   string `json:"refreshToken"`
}

type Handler struct {
	logger *zap.Logger
	tracer trace.Tracer

	validator *validator.Validate

	jwtIssuer JWTIssuer
}

func NewHandler(
	logger *zap.Logger,
	validator *validator.Validate,
	jwtIssuer JWTIssuer) *Handler {
	return &Handler{
		validator: validator,
		logger:    logger,
		tracer:    otel.Tracer("jwt/handler"),
		jwtIssuer: jwtIssuer,
	}
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Validate the request and extract the refresh token
	pathRefreshToken := r.PathValue("refreshToken")
	if pathRefreshToken == "" {
		h.logger.Warn("Missing refresh token in request path")
		http.Error(w, "Missing refresh token", http.StatusBadRequest)
		return
	}
	refreshTokenID, err := internal.ParseUUID(pathRefreshToken)
	if err != nil {
		h.logger.Warn("Invalid refresh token format", zap.Error(err), zap.String("refreshToken", pathRefreshToken))
		http.Error(w, "Invalid refresh token format", http.StatusBadRequest)
		return
	}

	// Get the user associated with the refresh token
	jwtUser, err := h.jwtIssuer.GetUserByRefreshToken(r.Context(), refreshTokenID)
	if err != nil {
		h.logger.Warn("Refresh token invalid", zap.Error(err))
		http.Error(w, "Failed to get user by refresh token", http.StatusInternalServerError)
		return
	}

	// Generate a new JWT and refresh token
	jwtToken, err := h.jwtIssuer.New(r.Context(), jwtUser)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err), zap.String("userID", jwtUser.ID.String()))
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
	}

	newRefreshToken, err := h.jwtIssuer.GenerateRefreshToken(r.Context(), jwtUser)
	if err != nil {
		h.logger.Error("Failed to generate new refresh token", zap.Error(err), zap.String("userID", jwtUser.ID.String()))
		http.Error(w, "Failed to generate new refresh token", http.StatusInternalServerError)
		return
	}

	// Inactivate the old refresh token
	err = h.jwtIssuer.InactivateRefreshToken(r.Context(), refreshTokenID)
	if err != nil {
		h.logger.Error("Failed to inactivate old refresh token", zap.Error(err), zap.String("refreshTokenID", refreshTokenID.String()))
		http.Error(w, "Failed to inactivate old refresh token", http.StatusInternalServerError)
		return
	}

	internal.WriteJSONResponse(w, http.StatusOK, Response{
		AccessToken:    jwtToken,
		ExpirationTime: newRefreshToken.ExpirationDate.Time.Unix(),
		RefreshToken:   newRefreshToken.ID.String(),
	})
}
