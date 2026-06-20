package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/danielmiranda22/gator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type PostService struct {
	db *database.Queries
}

func NewPostService(db *database.Queries) *PostService {
	return &PostService{
		db: db,
	}
}

func (s *PostService) GetPostsForUserWithPagination(
	ctx context.Context,
	userId uuid.UUID,
	limit int32,
	offset int32,
) ([]database.GetPostsForUserWithPaginationRow, error) {

	posts, err := s.db.GetPostsForUserWithPagination(ctx, database.GetPostsForUserWithPaginationParams{
		UserID: userId,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't get posts: %w", err)
	}

	return posts, nil
}

func (s *PostService) GetLikedPostsForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]database.GetLikedPostsForUserRow, error) {

	posts, err := s.db.GetLikedPostsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get liked posts: %w", err)
	}

	return posts, nil
}

func (s *PostService) GetPostsForUserFilterByTitle(
	ctx context.Context,
	userId uuid.UUID,
	titleFilter string,
	limit int32,
) ([]database.GetPostsForUserFilterByTitleRow, error) {

	posts, err := s.db.GetPostsForUserFilterByTitle(ctx, database.GetPostsForUserFilterByTitleParams{
		UserID: userId,
		Title:  titleFilter,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't get filtered posts for user: %w", err)
	}

	return posts, nil
}

func (s *PostService) GetPostsForUserOldestWithPagination(
	ctx context.Context,
	userId uuid.UUID,
	limit int32,
	offset int32,
) ([]database.GetPostsForUserOldestWithPaginationRow, error) {

	posts, err := s.db.GetPostsForUserOldestWithPagination(ctx, database.GetPostsForUserOldestWithPaginationParams{
		UserID: userId,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't get oldest posts for user: %w", err)
	}

	return posts, nil
}

func (s *PostService) GetPostLikeByPostAndUser(
	ctx context.Context,
	postID uuid.UUID,
	userID uuid.UUID,
) (database.PostLike, error) {

	postLike, err := s.db.GetPostLikeByPostAndUser(ctx, database.GetPostLikeByPostAndUserParams{
		PostID: postID,
		UserID: userID,
	})
	if err != nil {
		return database.PostLike{}, fmt.Errorf("couldn't get post like by post and user: %w", err)
	}

	return postLike, nil
}

func (s *PostService) CreatePost(
	ctx context.Context,
	feedId uuid.UUID,
	rssTitle string,
	rssDescription string,
	rssLink string,
	publishedAt sql.NullTime,
) (database.Post, error) {

	post, err := s.db.CreatePost(ctx, database.CreatePostParams{
		ID:          uuid.New(),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		FeedID:      feedId,
		Title:       rssTitle,
		Description: sql.NullString{String: rssDescription, Valid: true},
		Url:         rssLink,
		PublishedAt: publishedAt,
	})

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // unique_violation error code in PostgreSQL meaning the post already exists
			return database.Post{}, nil
		}
		return database.Post{}, err
	}

	return post, nil
}

func (s *PostService) CreatePostLike(
	ctx context.Context,
	postID uuid.UUID,
	userID uuid.UUID,
) (database.PostLike, error) {

	postLike, err := s.db.CreatePostLike(ctx, database.CreatePostLikeParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PostID:    postID,
		UserID:    userID,
	})
	if err != nil {
		return database.PostLike{}, fmt.Errorf("couldn't create post like: %w", err)
	}

	return postLike, nil
}

func (s *PostService) GetPostByID(
	ctx context.Context,
	postID uuid.UUID,
) (database.GetPostByIDRow, error) {

	post, err := s.db.GetPostByID(ctx, postID)
	if err != nil {
		return database.GetPostByIDRow{}, fmt.Errorf("couldn't get post by ID: %w", err)
	}

	return post, nil
}

func (s *PostService) DeletePostLike(
	ctx context.Context,
	postID uuid.UUID,
	userID uuid.UUID,
) error {

	err := s.db.DeletePostLike(ctx, database.DeletePostLikeParams{
		PostID: postID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("couldn't delete post like: %w", err)
	}

	return nil
}

func (s *PostService) SearchPostsForUser(
	ctx context.Context,
	userID uuid.UUID,
	searchTerm string,
) ([]database.SearchPostsForUserRow, error) {

	posts, err := s.db.SearchPostsForUser(ctx, database.SearchPostsForUserParams{
		UserID:      userID,
		Term:        sql.NullString{String: searchTerm, Valid: true},
		LimitCount:  20,
		OffsetCount: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't search posts for user: %w", err)
	}

	return posts, nil
}
