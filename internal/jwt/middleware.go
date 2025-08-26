package jwt

import (
	"context"
	"errors"
	"github.com/NYCU-SDC/eng-training-social-backend/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
)

type Verifier interface {
	Parse(ctx context.Context, tokenString string) (User, error)
}

type Middleware struct {
	logger *zap.Logger
	tracer trace.Tracer

	verifier Verifier
}

func NewMiddleware(logger *zap.Logger, verifier Verifier) Middleware {
	name := "middleware/jwt"
	tracer := otel.Tracer(name)

	return Middleware{
		tracer:   tracer,
		logger:   logger,
		verifier: verifier,
	}
}

func (m Middleware) StrictHandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			m.logger.Warn("Authorization header required")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := m.verifier.Parse(r.Context(), token)
		if err != nil {
			m.logger.Warn("Authorization header invalid", zap.Error(err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		m.logger.Debug("Authorization header valid")
		r = r.WithContext(context.WithValue(r.Context(), internal.UserContextKey, user))
		next.ServeHTTP(w, r)
	}
}

func (m Middleware) OptionalHandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			m.logger.Info("Anonymous request")
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.verifier.Parse(r.Context(), token)
		if err != nil {
			m.logger.Warn("Authorization header invalid", zap.Error(err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		m.logger.Debug("Authorization header valid")
		r = r.WithContext(context.WithValue(r.Context(), internal.UserContextKey, user))
		next.ServeHTTP(w, r)
	}
}

func GetUserFromContext(ctx context.Context) (User, error) {
	user, ok := ctx.Value(internal.UserContextKey).(User)
	if !ok {
		return User{}, errors.New("user not found in context")
	}
	return user, nil
}

func GetUserOrNilFromContext(ctx context.Context) *User {
	user, ok := ctx.Value(internal.UserContextKey).(User)
	if !ok {
		return nil
	}
	return &user
}
