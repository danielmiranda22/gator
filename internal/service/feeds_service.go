package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/danielmiranda22/gator/internal/database"
	"github.com/danielmiranda22/gator/ui"
	"github.com/google/uuid"
)

type FeedService struct {
	db *database.Queries
}

func NewFeedService(db *database.Queries) *FeedService {
	return &FeedService{
		db: db,
	}
}

func (s *FeedService) Add(
	ctx context.Context,
	name string,
	url string,
	userID uuid.UUID,
) (database.Feed, error) {

	newFeed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    userID, // from parameter, not state ✅
	})
	if err != nil {
		return database.Feed{}, fmt.Errorf("couldn't create feed: %w", err)
	}
	return newFeed, nil
}

func (s *FeedService) Follow(
	ctx context.Context,
	feedID uuid.UUID,
	userID uuid.UUID,
) (database.CreateFeedFollowRow, error) {

	follow, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    feedID,
		UserID:    userID, // from parameter ✅
	})
	if err != nil {
		return database.CreateFeedFollowRow{}, fmt.Errorf("error auto-following feed: %v", err)
	}
	return follow, nil
}

func (s *FeedService) List(
	ctx context.Context,
) ([]database.GetAllFeedsRow, error) {

	feeds, err := s.db.GetAllFeeds(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting feeds for user: %v", err)
	}

	if len(feeds) == 0 {
		log.Printf("%sNo feeds found.%s", ui.ColorRed, ui.ColorReset)
		return nil, nil
	}

	return feeds, nil
}

func (s *FeedService) GetFeedByURL(
	ctx context.Context,
	url string,
) (database.Feed, error) {

	feed, err := s.db.GetFeedByURL(ctx, url)
	if err != nil {
		return database.Feed{}, fmt.Errorf("feed not found: %v", err)
	}

	return feed, nil
}

func (s *FeedService) GetNextFeedsToFetch(
	ctx context.Context,
	limit int32,
) ([]database.Feed, error) {

	feeds, err := s.db.GetNextFeedsToFetch(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting next feeds to fetch: %v", err)
	}

	return feeds, nil
}

func (s *FeedService) GetFeedFollowsForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]database.GetFeedFollowsForUserRow, error) {

	follows, err := s.db.GetFeedFollowsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting feed follows for user: %v", err)
	}

	return follows, nil
}

func (s *FeedService) Unfollow(
	ctx context.Context,
	feedID uuid.UUID,
	userID uuid.UUID,
) error {

	log.Printf("Unfollowing feed with ID: %s for user with ID: %s", feedID, userID)

	err := s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		FeedID: feedID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("error unfollowing feed: %v", err)
	}
	return nil
}

func (s *FeedService) MarkFeedFetched(
	ctx context.Context,
	feedID uuid.UUID,
) error {

	err := s.db.MarkFeedFetched(ctx, feedID)
	if err != nil {
		return fmt.Errorf("error marking feed as fetched: %v", err)
	}
	return nil
}
