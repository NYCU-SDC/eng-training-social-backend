package user

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service struct {
	logger  *zap.Logger
	queries *Queries
}

func NewService(logger *zap.Logger, db DBTX) Service {
	return Service{
		logger:  logger,
		queries: New(db),
	}
}

func (s Service) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	user, err := s.queries.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user by ID", zap.String("id", id.String()), zap.Error(err))
		return User{}, err
	}
	return user, nil
}

func (s Service) FindOrCreate(ctx context.Context, email, username string) (User, error) {
	exists, err := s.queries.ExistsByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Failed to check if user exists by email", zap.String("email", email), zap.Error(err))
		return User{}, err
	}
	if exists {
		user, err := s.queries.GetByEmail(ctx, email)
		if err == nil {
			s.logger.Info("User found by email", zap.String("email", email), zap.String("username", user.Username))
			return user, nil
		}
	}

	user, err := s.queries.Create(ctx, CreateParams{
		Email:    email,
		Username: username,
	})
	if err != nil {
		s.logger.Error("Failed to find or create user", zap.String("email", email), zap.String("username", username), zap.Error(err))
		return User{}, err
	}
	return user, nil
}
