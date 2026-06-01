package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/danielmiranda22/gator/internal/database"
	"github.com/danielmiranda22/gator/internal/rss"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func scrapeFeed(s *state, feed database.Feed) error {
	if err := s.db.MarkFeedFetched(context.Background(), feed.ID); err != nil {
		return fmt.Errorf("error marking feed as fetched: %w", err)
	}

	feedData, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
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
			FeedID:      feed.ID,
			Title:       item.Title,
			Description: sql.NullString{String: item.Description, Valid: true},
			Url:         item.Link,
			PublishedAt: publishedAt,
		})
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				continue
			}
			log.Printf("couldn't create post for feed %s: %v", feed.Name, err)
		}
	}

	log.Printf("feed %s collected, %d posts found", feed.Name, len(feedData.Channel.Item))
	return nil
}

func scrapeFeedsConcurrently(s *state, feeds []database.Feed, workerCount int) {
	jobs := make(chan database.Feed)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for feed := range jobs {
				log.Printf("[worker %d] fetching %s", workerID, feed.Name)

				if err := scrapeFeed(s, feed); err != nil {
					log.Printf("[worker %d] error scraping %s: %v", workerID, feed.Name, err)
				}
			}
		}(i + 1)
	}

	for _, feed := range feeds {
		jobs <- feed
	}

	close(jobs)
	wg.Wait()
}
