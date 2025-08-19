package follow

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Response struct {
	Follow bool `json:"follow"`
}

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

func (s *Service) ExistsByID(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	following, err := s.queries.ExistByID(ctx, ExistByIDParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
	if err != nil {
		s.logger.Error("Failed to get following by ID", zap.String("follower_id", followerID.String()), zap.String("following_id", followingID.String()), zap.Error(err))
		return following, err
	}

	return following, nil
}

func (s *Service) Create(ctx context.Context, followerID, followingID uuid.UUID) error {
	_, err := s.queries.Create(ctx, CreateParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
	if err != nil {
		s.logger.Error("Failed to create follow relationship", zap.String("follower_id", followerID.String()), zap.String("following_id", followingID.String()), zap.Error(err))
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, followerID, followingID uuid.UUID) error {
	err := s.queries.Delete(ctx, DeleteParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
	if err != nil {
		s.logger.Error("Failed to delete follow relationship", zap.String("follower_id", followerID.String()), zap.String("following_id", followingID.String()), zap.Error(err))
		return err
	}

	return nil
}
