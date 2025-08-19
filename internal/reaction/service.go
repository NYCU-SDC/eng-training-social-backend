package reaction

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

func NewService(logger *zap.Logger, db DBTX) *Service {
	return &Service{
		logger:  logger,
		queries: New(db),
	}
}

func (s *Service) ReactToPost(ctx context.Context, postID, userID uuid.UUID, reactionType ReactionType) (Reaction, error) {
	reaction, err := s.queries.CreateByPostID(ctx, CreateByPostIDParams{
		PostID:       pgtype.UUID{Valid: true, Bytes: postID},
		UserID:       userID,
		ReactionType: reactionType,
	})
	if err != nil {
		s.logger.Error("Failed to add reaction", zap.Error(err))
		return Reaction{}, err
	}
	return reaction, nil
}

func (s *Service) RemoveToPost(ctx context.Context, postID, userID uuid.UUID) error {
	err := s.queries.DeleteByPostIDAndUserID(ctx, DeleteByPostIDAndUserIDParams{
		PostID: pgtype.UUID{Valid: true, Bytes: postID},
		UserID: userID,
	})
	if err != nil {
		s.logger.Error("Failed to remove reaction", zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) GetByPostIDAndUserID(ctx context.Context, postID, userID uuid.UUID) (Reaction, error) {
	reaction, err := s.queries.GetByPostIDAndUserID(ctx, GetByPostIDAndUserIDParams{
		PostID: pgtype.UUID{Valid: true, Bytes: postID},
		UserID: userID,
	})
	if err != nil {
		s.logger.Error("Failed to get reactions by post ID", zap.Error(err))
		return Reaction{}, err
	}
	return reaction, nil
}

func (s *Service) ReactToComment(ctx context.Context, commentID, userID uuid.UUID, reactionType ReactionType) (Reaction, error) {
	reaction, err := s.queries.CreateByCommentID(ctx, CreateByCommentIDParams{
		CommentID:    pgtype.UUID{Valid: true, Bytes: commentID},
		UserID:       userID,
		ReactionType: reactionType,
	})
	if err != nil {
		s.logger.Error("Failed to add reaction", zap.Error(err))
		return Reaction{}, err
	}
	return reaction, nil
}

func (s *Service) RemoveToComment(ctx context.Context, commentID, userID uuid.UUID) error {
	err := s.queries.DeleteByCommentIDAndUserID(ctx, DeleteByCommentIDAndUserIDParams{
		CommentID: pgtype.UUID{Valid: true, Bytes: commentID},
		UserID:    userID,
	})
	if err != nil {
		s.logger.Error("Failed to remove reaction", zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) GetByCommentIDAndUserID(ctx context.Context, commentID, userID uuid.UUID) (Reaction, error) {
	reaction, err := s.queries.GetByCommentIDAndUserID(ctx, GetByCommentIDAndUserIDParams{
		CommentID: pgtype.UUID{Valid: true, Bytes: commentID},
		UserID:    userID,
	})
	if err != nil {
		s.logger.Error("Failed to get reactions by comment ID", zap.Error(err))
		return Reaction{}, err
	}
	return reaction, nil
}
