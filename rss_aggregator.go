package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/danielmiranda22/gator/internal/database"
	"github.com/danielmiranda22/gator/internal/rss"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func scrapeFeeds(s *state) {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("error getting next feed to fetch: %v\n", err)
		return
	}

	if err := s.db.MarkFeedFetched(context.Background(), nextFeed.ID); err != nil {
		fmt.Printf("error marking feed as fetched: %v\n", err)
		return
	}

	feedData, err := rss.FetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		fmt.Printf("error fetching feed: %v\n", err)
		return
	}

	for _, item := range feedData.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			publishedAt = sql.NullTime{Time: t, Valid: true}
		}

		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			FeedID:      nextFeed.ID,
			Title:       item.Title,
			Description: sql.NullString{String: item.Description, Valid: true},
			Url:         item.Link,
			PublishedAt: publishedAt,
		})
		if err != nil {
			// ✅ proper PostgreSQL error code check — not string matching
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				continue // duplicate URL — expected, skip silently
			}
			log.Printf("couldn't create post: %v", err)
		}
	}
	log.Printf("feed %s collected, %v posts found", nextFeed.Name, len(feedData.Channel.Item))
}
