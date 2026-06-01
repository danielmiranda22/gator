package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
		return fmt.Errorf("usage: %v <time_between_reqs> [worker_count]", cmd.name)
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %v", err)
	}

	workerCount := 1
	if len(cmd.args) == 2 {
		parsedWorkers, err := strconv.Atoi(cmd.args[1])
		if err != nil || parsedWorkers < 1 {
			return fmt.Errorf("invalid worker count: %s", cmd.args[1])
		}
		workerCount = parsedWorkers
	}

	fmt.Printf("Collecting feeds every %v with %d worker(s)\n", timeBetweenReqs, workerCount)

	ticker := time.NewTicker(timeBetweenReqs)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		feeds, err := s.db.GetNextFeedsToFetch(context.Background(), int32(workerCount))
		if err != nil {
			log.Printf("error getting next feeds to fetch: %v", err)
			continue
		}

		if len(feeds) == 0 {
			log.Println("no feeds to fetch")
			continue
		}

		scrapeFeedsConcurrently(s, feeds, workerCount)
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
	limit := 2
	sort := "newest"
	titleFilter := ""
	page := 1

	if len(cmd.args) == 1 && cmd.args[0] == "--help" {
		fmt.Printf("%sUsage:%s %s [--limit <number>] [--sort <newest|oldest>] [--filter <text>] [--page <number>]\n",
			colorCyan, colorReset, cmd.name)
		fmt.Printf("%sDefaults:%s limit=2, sort=newest, page=1\n", colorGray, colorReset)
		fmt.Printf("%sExample:%s %s --limit 5 --sort oldest --page 2\n",
			colorGray, colorReset, cmd.name)
		return nil
	}

	for i := 0; i < len(cmd.args); i++ {
		arg := cmd.args[i]

		switch arg {
		case "--limit":
			if i+1 >= len(cmd.args) {
				return fmt.Errorf("missing value for --limit")
			}
			val, err := strconv.Atoi(cmd.args[i+1])
			if err != nil || val < 1 {
				return fmt.Errorf("--limit must be a positive integer")
			}
			limit = val
			i++

		case "--sort":
			if i+1 >= len(cmd.args) {
				return fmt.Errorf("missing value for --sort")
			}
			val := cmd.args[i+1]
			if val != "newest" && val != "oldest" {
				return fmt.Errorf("--sort must be 'newest' or 'oldest'")
			}
			sort = val
			i++

		case "--filter":
			if i+1 >= len(cmd.args) {
				return fmt.Errorf("missing value for --filter")
			}
			titleFilter = cmd.args[i+1]
			i++

		case "--page":
			if i+1 >= len(cmd.args) {
				return fmt.Errorf("missing value for --page")
			}
			val, err := strconv.Atoi(cmd.args[i+1])
			if err != nil || val < 1 {
				return fmt.Errorf("--page must be a positive integer")
			}
			page = val
			i++

		default:
			return fmt.Errorf("unknown option: %s", arg)
		}
	}

	ctx := context.Background()
	offset := (page - 1) * limit

	switch {
	case titleFilter != "":
		posts, err := s.db.GetPostsForUserFilterByTitle(ctx, database.GetPostsForUserFilterByTitleParams{
			UserID: user.ID,
			Title:  titleFilter,
			Limit:  int32(limit),
		})
		if err != nil {
			return fmt.Errorf("couldn't get filtered posts for user: %w", err)
		}
		printPosts(posts, user.Name, page, limit, sort, titleFilter)
		return nil

	case sort == "oldest":
		posts, err := s.db.GetPostsForUserOldestWithPagination(ctx, database.GetPostsForUserOldestWithPaginationParams{
			UserID: user.ID,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("couldn't get oldest posts for user: %w", err)
		}
		printPosts(posts, user.Name, page, limit, sort, titleFilter)
		return nil

	default:
		posts, err := s.db.GetPostsForUserWithPagination(ctx, database.GetPostsForUserWithPaginationParams{
			UserID: user.ID,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			return fmt.Errorf("couldn't get posts for user: %w", err)
		}
		printPosts(posts, user.Name, page, limit, sort, titleFilter)
		return nil
	}
}

var (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorCyan    = "\033[36m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorGray    = "\033[90m"
)

func printPosts[T any](posts []T, userName string, page int, limit int, sort string, filter string) {
	if len(posts) == 0 {
		fmt.Printf("%sNo posts found for %s on page %d.%s\n", colorYellow, userName, page, colorReset)
		if filter != "" {
			fmt.Printf("%sTry a different filter or a lower page number.%s\n", colorGray, colorReset)
		}
		return
	}

	fmt.Printf("\n%s%sPosts for %s%s", colorBold, colorCyan, userName, colorReset)
	fmt.Printf("%s — page %d — sort: %s — limit: %d%s\n", colorGray, page, sort, limit, colorReset)
	if filter != "" {
		fmt.Printf("%sFilter:%s %s\n", colorYellow, colorReset, filter)
	}
	fmt.Printf("%s════════════════════════════════════════════════════════════%s\n", colorGray, colorReset)

	for _, post := range posts {
		// Use type switch if your sqlc-generated rows have different struct types, or
		// just copy your previous print loop for each branch if needed.
		switch p := any(post).(type) {
		case database.GetPostsForUserWithPaginationRow:
			renderPost(p.Title, p.FeedName, p.Url, p.Description.String, p.PublishedAt.Time)
		case database.GetPostsForUserOldestWithPaginationRow:
			renderPost(p.Title, p.FeedName, p.Url, p.Description.String, p.PublishedAt.Time)
		case database.GetPostsForUserFilterByTitleRow:
			renderPost(p.Title, p.FeedName, p.Url, p.Description.String, p.PublishedAt.Time)
		}
	}
}

func truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max-3] + "..."
}

func renderPost(title, feedName, url, description string, publishedAt time.Time) {
	fmt.Printf("%s📅 %s%s  %s%s\n",
		colorBlue,
		publishedAt.Format("02 Jan 2006"),
		colorGray,
		feedName,
		colorReset,
	)

	fmt.Printf("%s%s%s\n", colorBold, title, colorReset)

	if description != "" {
		fmt.Printf("    %s%s%s\n", colorGray, truncate(description, 140), colorReset)
	}

	fmt.Printf("    %s🔗 %s%s\n", colorGreen, url, colorReset)
	fmt.Printf("%s────────────────────────────────────────%s\n", colorGray, colorReset)
}
