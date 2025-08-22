package comment

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

func (s Service) GetAllByPostID(ctx context.Context, postID uuid.UUID) ([]Comment, error) {
	comments, err := s.queries.GetAllByPostID(ctx, postID)
	if err != nil {
		s.logger.Error("Failed to get all comments", zap.Error(err))
		return nil, err
	}

	return comments, nil
}

func (s Service) GetByID(ctx context.Context, id uuid.UUID) (Comment, error) {
	comment, err := s.queries.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get comment by ID", zap.Any("id", id), zap.Error(err))
		return Comment{}, err
	}
	return comment, nil
}

func (s Service) Create(ctx context.Context, content string, postID, userID uuid.UUID, username string) (Comment, error) {
	comment, err := s.queries.Create(ctx, CreateParams{
		Content:    pgtype.Text{String: content, Valid: true},
		AuthorID:   pgtype.UUID{Bytes: userID, Valid: true},
		AuthorName: pgtype.Text{String: username, Valid: true},
		PostID:     postID,
	})
	if err != nil {
		s.logger.Error("Failed to create comment", zap.Error(err))
		return Comment{}, err
	}

	return comment, nil
}

func (s Service) Update(ctx context.Context, id uuid.UUID, content string) (Comment, error) {
	comment, err := s.queries.Update(ctx, UpdateParams{
		ID:      id,
		Content: pgtype.Text{String: content, Valid: true},
	})
	if err != nil {
		s.logger.Error("Failed to update comment", zap.String("id", id.String()), zap.Error(err))
		return Comment{}, err
	}
	return comment, nil
}

func (s Service) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.queries.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete comment", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	return nil
}
