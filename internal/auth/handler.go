package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/NYCU-SDC/eng-training-social-backend/internal"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/auth/oauthprovider"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/config"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/jwt"
	"github.com/NYCU-SDC/eng-training-social-backend/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
)

type JWTIssuer interface {
	New(ctx context.Context, user jwt.User) (string, error)
	Parse(ctx context.Context, tokenString string) (jwt.User, error)
	GenerateRefreshToken(ctx context.Context, user jwt.User) (jwt.RefreshToken, error)
}

type Response struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

type UserStore interface {
	FindOrCreate(ctx context.Context, email, username string) (user.User, error)
}

type OAuthProvider interface {
	Name() string
	Config() *oauth2.Config
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*oauthprovider.UserInfo, error)
}

type callBackInfo struct {
	code       string
	oauthError string
	callback   url.URL
	redirectTo string
}

type Handler struct {
	logger    *zap.Logger
	config    config.Config
	validator *validator.Validate
	userStore UserStore
	jwtIssuer JWTIssuer
	provider  map[string]OAuthProvider
}

func NewHandler(logger *zap.Logger, config config.Config, validator *validator.Validate, userStore UserStore, jwtIssuer JWTIssuer) *Handler {
	googleProvider := oauthprovider.NewGoogleConfig(
		config.GoogleClientID,
		config.GoogleClientSecret,
		fmt.Sprintf("%s/api/oauth/google/callback", config.BaseURL))

	return &Handler{
		logger:    logger,
		config:    config,
		validator: validator,
		userStore: userStore,
		jwtIssuer: jwtIssuer,
		provider: map[string]OAuthProvider{
			"google": googleProvider,
		},
	}
}

// Oauth2Start initiates the OAuth2 flow by redirecting the user to the provider's authorization URL
func (h *Handler) Oauth2Start(w http.ResponseWriter, r *http.Request) {
	providerName := r.PathValue("provider")
	provider := h.provider[providerName]
	if provider == nil {
		h.logger.Error("OAuth2 provider not found", zap.String("provider", providerName))
		internal.WriteJSONResponse(w, http.StatusNotFound, fmt.Sprintf("OAuth2 provider '%s' not found", providerName))
		return
	}

	callback := r.URL.Query().Get("c")
	redirectTo := r.URL.Query().Get("r")
	if callback == "" {
		callback = fmt.Sprintf("%s/api/oauth/debug/token", h.config.BaseURL)
	}
	if redirectTo != "" {
		callback = fmt.Sprintf("%s?r=%s", callback, redirectTo)
	}
	state := base64.StdEncoding.EncodeToString([]byte(callback))

	authURL := provider.Config().AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)

	h.logger.Info("Redirecting to Google OAuth2", zap.String("url", authURL))
}

// Callback handles the OAuth2 callback from the provider
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	providerName := r.PathValue("provider")
	provider := h.provider[providerName]
	if provider == nil {
		h.logger.Error("OAuth2 provider not found", zap.String("provider", providerName))
		internal.WriteJSONResponse(w, http.StatusNotFound, fmt.Sprintf("OAuth2 provider '%s' not found", providerName))
		return
	}

	// Get the OAuth2 code and state from the request
	callbackInfo, err := h.getCallBackInfo(r.URL)
	if err != nil {
		h.logger.Error("Failed to get callback info", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get callback info: %s", err))
		return
	}

	callback := callbackInfo.callback.String()
	code := callbackInfo.code
	redirectTo := callbackInfo.redirectTo
	oauthError := callbackInfo.oauthError

	if oauthError != "" {
		http.Redirect(w, r, fmt.Sprintf("%s?error=%s", callback, oauthError), http.StatusTemporaryRedirect)
		return
	}

	token, err := provider.Exchange(r.Context(), code)
	if err != nil {
		h.logger.Error("Failed to exchange OAuth2 code for token", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to exchange OAuth2 code for token: %s", err))
		return
	}

	userInfo, err := provider.GetUserInfo(r.Context(), token)
	if err != nil {
		h.logger.Error("Failed to get user info from OAuth2 provider", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get user info from OAuth2 provider: %s", err))
		return
	}

	// Check if the user exists in the database, if not, create a new user
	newUser, err := h.userStore.FindOrCreate(r.Context(), userInfo.Email, userInfo.Username)
	if err != nil {
		h.logger.Error("Failed to find or create user", zap.Error(err), zap.String("email", userInfo.Email), zap.String("username", userInfo.Username))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to find or create user: %s", err))
		return
	}

	jwtToken, refreshTokenID, err := h.generateJWT(r.Context(), newUser)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate JWT token: %s", err))
		return
	}

	var redirectWithToken string
	if redirectTo != "" {
		redirectWithToken = fmt.Sprintf("%s?token=%s&refreshToken=%s&r=%s", callback, jwtToken, refreshTokenID, redirectTo)
	} else {
		redirectWithToken = fmt.Sprintf("%s?token=%s&refreshToken=%s", callback, jwtToken, refreshTokenID)
	}

	http.Redirect(w, r, redirectWithToken, http.StatusTemporaryRedirect)
}

func (h *Handler) getCallBackInfo(url *url.URL) (callBackInfo, error) {

	code := url.Query().Get("code")
	state := url.Query().Get("state")
	oauthError := url.Query().Get("error") // Check if there was an error during the OAuth2 process

	callbackURL, err := base64.StdEncoding.DecodeString(state)
	if err != nil {
		return callBackInfo{}, err
	}

	callback, err := url.Parse(string(callbackURL))
	if err != nil {
		return callBackInfo{}, err
	}

	// Clear the query parameters from the callback URL, due to "?" symbol in original URL
	redirectTo := callback.Query().Get("r")
	callback.RawQuery = ""

	return callBackInfo{
		code:       code,
		oauthError: oauthError,
		callback:   *callback,
		redirectTo: redirectTo,
	}, nil
}

func (h *Handler) DebugToken(w http.ResponseWriter, r *http.Request) {
	e := r.URL.Query().Get("error")
	if e != "" {
		h.logger.Error("Debug endpoint error", zap.String("error", e))
		internal.WriteJSONResponse(w, http.StatusBadRequest, fmt.Sprintf("Error: %s", e))
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		h.logger.Error("Token is required for debug endpoint")
		internal.WriteJSONResponse(w, http.StatusBadRequest, "Token is required for debug endpoint")
		return
	}

	jwtUser, err := h.jwtIssuer.Parse(r.Context(), token)
	if err != nil {
		h.logger.Error("Failed to parse JWT token", zap.Error(err), zap.String("token", token))
		internal.WriteJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse JWT token: %s", err))
		return
	}

	internal.WriteJSONResponse(w, http.StatusOK, jwtUser)
}

func (h *Handler) generateJWT(ctx context.Context, user user.User) (string, string, error) {
	jwtToken, err := h.jwtIssuer.New(ctx, jwt.User(user))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := h.jwtIssuer.GenerateRefreshToken(ctx, jwt.User(user))
	if err != nil {
		return "", "", err
	}

	return jwtToken, refreshToken.ID.String(), nil
}
