package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/danielmiranda22/gator/internal/cli"
	"github.com/danielmiranda22/gator/internal/database"
	"github.com/danielmiranda22/gator/internal/rss"
	"github.com/google/uuid"
)

// add to handlers.go — agg handler
func Agg(s *cli.State, cmd cli.Command) error {
	if len(cmd.Args) < 1 || len(cmd.Args) > 2 {
		return fmt.Errorf("usage: %v <time_between_reqs> [worker_count]", cmd.Name)
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %v", err)
	}

	workerCount := 1
	if len(cmd.Args) == 2 {
		parsedWorkers, err := strconv.Atoi(cmd.Args[1])
		if err != nil || parsedWorkers < 1 {
			return fmt.Errorf("invalid worker count: %s", cmd.Args[1])
		}
		workerCount = parsedWorkers
	}

	fmt.Printf("Collecting feeds every %v with %d worker(s)\n", timeBetweenReqs, workerCount)

	ticker := time.NewTicker(timeBetweenReqs)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		feeds, err := s.DB.GetNextFeedsToFetch(context.Background(), int32(workerCount))
		if err != nil {
			log.Printf("error getting next feeds to fetch: %v", err)
			continue
		}

		if len(feeds) == 0 {
			log.Println("no feeds to fetch")
			continue
		}

		rss.ScrapeFeedsConcurrently(s, feeds, workerCount)
	}
}

func AddFeed(s *cli.State, cmd cli.Command, user database.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}
	newFeed, err := s.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
		Url:       cmd.Args[1],
		UserID:    user.ID, // from parameter, not state ✅
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}

	fmt.Printf("* Name: %s\n* URL: %s\n", newFeed.Name, newFeed.Url)

	follow, err := s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    newFeed.ID,
		UserID:    user.ID, // from parameter ✅
	})
	if err != nil {
		return fmt.Errorf("error auto-following feed: %v", err)
	}
	fmt.Printf("Now following: %s\n", follow.FeedName)
	return nil
}

func ListFeeds(s *cli.State, cmd cli.Command) error {
	feeds, err := s.DB.GetAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting feeds for user: %v", err)
	}

	for _, feed := range feeds {
		fmt.Printf("Feed owned by %s:\n", feed.UserName)
		fmt.Printf("* ID:            %s\n", feed.ID)
		fmt.Printf("* Created:       %v\n", feed.CreatedAt)
		fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
		fmt.Printf("* Name:          %s\n", feed.Name)
		fmt.Printf("* URL:           %s\n", feed.Url)
		fmt.Printf("* UserID:        %s\n", feed.UserID)
		fmt.Println("-------------------------------------")
	}

	return nil
}

// add these three handlers
func Follow(s *cli.State, cmd cli.Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: follow <feed_url>")
	}
	feed, err := s.DB.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("feed not found: %v", err)
	}
	follow, err := s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    feed.ID,
		UserID:    user.ID, // from parameter ✅
	})
	if err != nil {
		return fmt.Errorf("error following feed: %v", err)
	}
	fmt.Printf("Following feed: %s\n", follow.FeedName)
	fmt.Printf("User: %s\n", follow.UserName)
	return nil
}

func Following(s *cli.State, cmd cli.Command, user database.User) error {
	follows, err := s.DB.GetFeedFollowsForUser(context.Background(), user.ID) // from parameter ✅
	if err != nil {
		return fmt.Errorf("error getting feed follows: %v", err)
	}
	if len(follows) == 0 {
		fmt.Println("You are not following any feeds")
		return nil
	}
	fmt.Println("Feeds you follow:")
	for _, follow := range follows {
		fmt.Printf("  * %s\n", follow.FeedName)
	}
	return nil
}

func Unfollow(s *cli.State, cmd cli.Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: unfollow <feed_url>")
	}
	feed, err := s.DB.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("feed not found: %v", err)
	}
	err = s.DB.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("error unfollowing feed: %v", err)
	}
	fmt.Printf("Unfollowed: %s\n", feed.Name)
	return nil
}
