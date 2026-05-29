package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/danielmiranda22/gator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("usage: login <username>")
	}
	username := cmd.args[0]

	// login requires the user to exist in the DB
	_, err := s.db.GetUser(context.Background(), username)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("user %s does not exist", username)
	}
	if err != nil {
		return fmt.Errorf("error looking up user: %v", err)
	}

	if err := s.cfg.SetUser(username); err != nil {
		return err
	}
	fmt.Printf("Logged in as %s\n", username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("usage: register <username>")
	}
	username := cmd.args[0]

	// check if user already exists
	_, err := s.db.GetUser(context.Background(), username)
	if err == nil {
		// no error = user found = already exists
		return fmt.Errorf("user %s already exists", username)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		// real DB error
		return fmt.Errorf("error checking for user: %v", err)
	}

	// create the user
	newUser, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	// log in as the new user
	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Printf("User created successfully: %+v\n", newUser)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if err := s.db.DeleteAllUsers(context.Background()); err != nil {
		return err
	}
	fmt.Println("Users reset successfully")
	return nil
}

func handlerListUsers(s *state, cmd command) error {
	users, err := s.db.GetAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all users: %v", err)
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %v (current)\n", user.Name)
			continue
		}
		fmt.Printf("* %v\n", user.Name)
	}
	return nil
}

// add to handlers.go — agg handler
func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 || len(cmd.args) > 2 {
		return fmt.Errorf("usage: %v <time_between_reqs>", cmd.name)
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %v", err)
	}

	fmt.Printf("Collecting feeds every %v\n", timeBetweenReqs)

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}
	newFeed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID, // from parameter, not state ✅
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}

	fmt.Printf("* Name: %s\n* URL: %s\n", newFeed.Name, newFeed.Url)

	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
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

func handlerListFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetAllFeeds(context.Background())
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
func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("usage: follow <feed_url>")
	}
	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("feed not found: %v", err)
	}
	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
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

func handlerFollowing(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID) // from parameter ✅
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

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("usage: unfollow <feed_url>")
	}
	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("feed not found: %v", err)
	}
	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("error unfollowing feed: %v", err)
	}
	fmt.Printf("Unfollowed: %s\n", feed.Name)
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	// --- Defaults ---
	limit := 2
	sort := "newest"
	titleFilter := ""
	page := 1

	// instructions for flags:
	// --limit <number>: limits the number of posts returned (default: 2)
	// --sort <newest|oldest>: sorts posts by published date (default: newest)
	// --title <title>: filters posts by title (exact match)
	// --page <number>: for pagination, returns the next set of results based on the limit (default: 1)
	if len(cmd.args) == 0 {
		fmt.Println("Usage: browse [--limit <number>] [--sort <newest|oldest>] [--title <title>] [--page <number>]")
		fmt.Println("Default limit is 2, default sort is newest, default page is 1")
		fmt.Println("Example: browse --limit 5 --sort oldest --page 2")
		return nil
	}

	// --- Parse CLI flags ---
	for i := 0; i < len(cmd.args); i++ {
		arg := cmd.args[i]
		switch arg {
		case "--limit":
			if i+1 < len(cmd.args) {
				if val, err := strconv.Atoi(cmd.args[i+1]); err == nil {
					limit = val
				}
				i++
			}
		case "--sort":
			if i+1 < len(cmd.args) {
				if cmd.args[i+1] == "oldest" || cmd.args[i+1] == "newest" {
					sort = cmd.args[i+1]
				}
				i++
			}
		case "--filter":
			if i+1 < len(cmd.args) {
				titleFilter = cmd.args[i+1]
				i++
			}
		case "--page":
			if i+1 < len(cmd.args) {
				if val, err := strconv.Atoi(cmd.args[i+1]); err == nil && val >= 1 {
					page = val
				}
				i++
			}
		}
	}

	ctx := context.Background()
	offset := (page - 1) * limit

	// --- Branching logic ---
	switch {
	// --- Filter by title/description, paginated ---
	case titleFilter != "":
		posts, err := s.db.GetPostsForUserFilterByTitle(ctx, database.GetPostsForUserFilterByTitleParams{
			UserID: user.ID,
			Title:  titleFilter,
			Limit:  int32(limit),
			// If you want pagination here, you'll need to add OFFSET to this query as well!
		})
		if err != nil {
			return fmt.Errorf("couldn't get posts for user: %w", err)
		}
		printPosts(posts, user.Name, page)
		return nil

	// --- Sort oldest, paginated ---
	case sort == "oldest":
		posts, err := s.db.GetPostsForUserOldestWithPagination(ctx, database.GetPostsForUserOldestWithPaginationParams{
			UserID: user.ID,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("couldn't get posts for user: %w", err)
		}
		printPosts(posts, user.Name, page)
		return nil

	// --- Default (sort newest, paginated) ---
	default:
		posts, err := s.db.GetPostsForUserWithPagination(ctx, database.GetPostsForUserWithPaginationParams{
			UserID: user.ID,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("couldn't get posts for user: %w", err)
		}
		printPosts(posts, user.Name, page)
		return nil
	}
}
func printPosts[T any](posts []T, userName string, page int) {
	if len(posts) == 0 {
		fmt.Printf("No posts found for user %s (page %d)\n", userName, page)
		return
	}

	fmt.Printf("Found %d posts for user %s (page %d):\n", len(posts), userName, page)
	for _, post := range posts {
		// Use type switch if your sqlc-generated rows have different struct types, or
		// just copy your previous print loop for each branch if needed.
		switch p := any(post).(type) {
		case database.GetPostsForUserWithPaginationRow:
			fmt.Printf("%s from %s\n", p.PublishedAt.Time.Format("Mon Jan 2"), p.FeedName)
			fmt.Printf("--- %s ---\n", p.Title)
			fmt.Printf("    %v\n", p.Description.String)
			fmt.Printf("Link: %s\n", p.Url)
			fmt.Println("=====================================")
		case database.GetPostsForUserOldestWithPaginationRow:
			fmt.Printf("%s from %s\n", p.PublishedAt.Time.Format("Mon Jan 2"), p.FeedName)
			fmt.Printf("--- %s ---\n", p.Title)
			fmt.Printf("    %v\n", p.Description.String)
			fmt.Printf("Link: %s\n", p.Url)
			fmt.Println("=====================================")
		case database.GetPostsForUserFilterByTitleRow:
			fmt.Printf("%s from %s\n", p.PublishedAt.Time.Format("Mon Jan 2"), p.FeedName)
			fmt.Printf("--- %s ---\n", p.Title)
			fmt.Printf("    %v\n", p.Description.String)
			fmt.Printf("Link: %s\n", p.Url)
			fmt.Println("=====================================")
		}
	}
}
