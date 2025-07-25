package post

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

func (s Service) GetAll(ctx context.Context) ([]Post, error) {
	posts, err := s.queries.GetAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get all posts", zap.Error(err))
		return nil, err
	}

	return posts, nil
}

func (s Service) GetByID(ctx context.Context, id uuid.UUID) (Post, error) {
	post, err := s.queries.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get post by ID", zap.Any("id", id), zap.Error(err))
		return Post{}, err
	}
	return post, nil
}

func (s Service) Create(ctx context.Context, title, content string) (Post, error) {
	post, err := s.queries.Create(ctx, CreateParams{
		Title:   pgtype.Text{String: title, Valid: true},
		Content: pgtype.Text{String: content, Valid: true},
	})
	if err != nil {
		s.logger.Error("Failed to create post", zap.Error(err))
		return Post{}, err
	}
	return post, nil
}

func (s Service) Update(ctx context.Context, id uuid.UUID, title, content string) (Post, error) {
	post, err := s.queries.Update(ctx, UpdateParams{
		ID:      id,
		Title:   pgtype.Text{String: title, Valid: true},
		Content: pgtype.Text{String: content, Valid: true},
	})
	if err != nil {
		s.logger.Error("Failed to update post", zap.String("id", id.String()), zap.Error(err))
		return Post{}, err
	}
	return post, nil
}

func (s Service) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.queries.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete post", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	return nil
}
