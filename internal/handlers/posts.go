package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/danielmiranda22/gator/internal/cli"
	"github.com/danielmiranda22/gator/internal/database"
	"github.com/danielmiranda22/gator/ui"
	"github.com/google/uuid"
)

func Browse(s *cli.State, cmd cli.Command, user database.User) error {
	if len(cmd.Args) > 0 && cmd.Args[0] == "tui" {
		posts, err := s.Services.Posts.GetPostsForUserWithPagination(s.Ctx, user.ID, 20, 0)
		if err != nil {
			return err
		}

		tuiPosts := make([]ui.TUIPost, 0, len(posts))
		for _, p := range posts {
			tuiPosts = append(tuiPosts, ui.TUIPost{
				ID:          p.ID,
				Title:       p.Title,
				FeedName:    p.FeedName,
				URL:         p.Url,
				Description: p.Description.String,
				PublishedAt: p.PublishedAt.Time,
			})
		}

		if len(tuiPosts) == 0 {
			fmt.Println("No posts available for TUI.")
			return nil
		}

		return ui.RunTUI(tuiPosts)
	}

	if len(cmd.Args) > 0 && cmd.Args[0] == "liked" {
		if len(cmd.Args) > 1 && cmd.Args[1] == "tui" {
			return LikedTUI(s, user)
		}
		return Liked(s, user)
	}

	limit := 2
	sort := "newest"
	titleFilter := ""
	page := 1

	if len(cmd.Args) == 1 && cmd.Args[0] == "--help" {
		fmt.Printf("%sUsage:%s %s [--limit <number>] [--sort <newest|oldest>] [--filter <text>] [--page <number>]\n",
			ui.ColorCyan, ui.ColorReset, cmd.Name)
		fmt.Printf("%sDefaults:%s limit=2, sort=newest, page=1\n", ui.ColorGray, ui.ColorReset)
		fmt.Printf("%sExample:%s %s --limit 5 --sort oldest --page 2\n",
			ui.ColorGray, ui.ColorReset, cmd.Name)
		return nil
	}

	for i := 0; i < len(cmd.Args); i++ {
		arg := cmd.Args[i]

		switch arg {
		case "--limit":
			if i+1 >= len(cmd.Args) {
				return fmt.Errorf("missing value for --limit")
			}
			val, err := strconv.Atoi(cmd.Args[i+1])
			if err != nil || val < 1 {
				return fmt.Errorf("--limit must be a positive integer")
			}
			limit = val
			i++

		case "--sort":
			if i+1 >= len(cmd.Args) {
				return fmt.Errorf("missing value for --sort")
			}
			val := cmd.Args[i+1]
			if val != "newest" && val != "oldest" {
				return fmt.Errorf("--sort must be 'newest' or 'oldest'")
			}
			sort = val
			i++

		case "--filter":
			if i+1 >= len(cmd.Args) {
				return fmt.Errorf("missing value for --filter")
			}
			titleFilter = cmd.Args[i+1]
			i++

		case "--page":
			if i+1 >= len(cmd.Args) {
				return fmt.Errorf("missing value for --page")
			}
			val, err := strconv.Atoi(cmd.Args[i+1])
			if err != nil || val < 1 {
				return fmt.Errorf("--page must be a positive integer")
			}
			page = val
			i++

		default:
			return fmt.Errorf("unknown option: %s", arg)
		}
	}

	offset := (page - 1) * limit

	switch {
	case titleFilter != "":
		posts, err := s.Services.Posts.GetPostsForUserFilterByTitle(s.Ctx, user.ID, titleFilter, int32(limit))
		if err != nil {
			return err
		}
		ui.PrintPosts(posts, user.Name, page, limit, sort, titleFilter)
		return nil

	case sort == "oldest":
		posts, err := s.Services.Posts.GetPostsForUserOldestWithPagination(s.Ctx, user.ID, int32(limit), int32(offset))
		if err != nil {
			return err
		}
		ui.PrintPosts(posts, user.Name, page, limit, sort, titleFilter)
		return nil

	default:
		posts, err := s.Services.Posts.GetPostsForUserWithPagination(s.Ctx, user.ID, int32(limit), int32(offset))
		if err != nil {
			return fmt.Errorf("couldn't get posts for user: %w", err)
		}
		ui.PrintPosts(posts, user.Name, page, limit, sort, titleFilter)
		return nil
	}
}

func Like(s *cli.State, cmd cli.Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: like <post_id>")
	}

	postUUID, err := uuid.Parse(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid post ID: %v", err)
	}

	post, err := s.Services.Posts.GetPostByID(s.Ctx, postUUID)
	if err != nil {
		return fmt.Errorf("couldn't get post by ID: %w", err)
	}

	_, err = s.Services.Posts.GetPostLikeByPostAndUser(s.Ctx, post.ID, user.ID)
	if err == nil {
		return fmt.Errorf("you already liked this post")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("couldn't check existing like: %w", err)
	}

	_, err = s.Services.Posts.CreatePostLike(s.Ctx, post.ID, user.ID)
	if err != nil {
		return fmt.Errorf("couldn't like post: %w", err)
	}

	fmt.Printf("Liked post: %s\n", post.Title)
	return nil
}

func Unlike(s *cli.State, cmd cli.Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: unlike <post_id>")
	}

	postUUID, err := uuid.Parse(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid post ID: %v", err)
	}

	post, err := s.Services.Posts.GetPostByID(s.Ctx, postUUID)
	if err != nil {
		return fmt.Errorf("couldn't get post by ID: %w", err)
	}

	_, err = s.Services.Posts.GetPostLikeByPostAndUser(s.Ctx, post.ID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("you have not liked this post")
		}
		return fmt.Errorf("couldn't check existing like: %w", err)
	}

	err = s.Services.Posts.DeletePostLike(s.Ctx, post.ID, user.ID)
	if err != nil {
		return fmt.Errorf("couldn't unlike post: %w", err)
	}

	fmt.Printf("Unliked post: %s\n", post.Title)
	return nil
}

func Liked(s *cli.State, user database.User) error {
	posts, err := s.Services.Posts.GetLikedPostsForUser(s.Ctx, user.ID)
	if err != nil {
		return fmt.Errorf("couldn't get liked posts: %w", err)
	}

	ui.PrintPosts(posts, user.Name, 1, len(posts), "liked", "liked posts")
	return nil
}

func LikedTUI(s *cli.State, user database.User) error {
	posts, err := s.Services.Posts.GetLikedPostsForUser(s.Ctx, user.ID)
	if err != nil {
		return err
	}

	tuiPosts := make([]ui.TUIPost, 0, len(posts))
	for _, p := range posts {
		tuiPosts = append(tuiPosts, ui.TUIPost{
			ID:          p.ID,
			Title:       p.Title,
			FeedName:    p.FeedName,
			URL:         p.Url,
			Description: p.Description.String,
			PublishedAt: p.PublishedAt.Time,
		})
	}

	if len(tuiPosts) == 0 {
		fmt.Println("No posts available for TUI.")
		return nil
	}

	return ui.RunTUI(tuiPosts)
}

func Search(s *cli.State, cmd cli.Command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: search <search_term>")
	}

	searchTerm := cmd.Args[0]

	posts, err := s.Services.Posts.SearchPostsForUser(s.Ctx, user.ID, searchTerm)
	if err != nil {
		return fmt.Errorf("couldn't search posts for user: %w", err)
	}

	ui.PrintPosts(posts, user.Name, 1, 20, "search results", searchTerm)
	return nil
}
