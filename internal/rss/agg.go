package rss

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/danielmiranda22/gator/internal/cli"
	"github.com/danielmiranda22/gator/internal/database"
	"github.com/danielmiranda22/gator/internal/service"
)

func scrapeFeed(s *cli.State, feed database.Feed) error {
	if err := s.Services.Feeds.MarkFeedFetched(s.Ctx, feed.ID); err != nil {
		return fmt.Errorf("error marking feed as fetched: %w", err)
	}
	log.Printf("fetching feed %s (%s)", feed.Name, feed.Url)

	feedData, err := service.FetchFeed(s.Ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
	}

	for _, item := range feedData.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			publishedAt = sql.NullTime{Time: t, Valid: true}
		}

		_, err = s.Services.Posts.CreatePost(s.Ctx, feed.ID, item.Title, item.Description, item.Link, publishedAt)
		if err != nil {
			log.Printf("couldn't create post for feed %s: %v", feed.Name, err)
		}
	}

	log.Printf("feed %s collected, %d posts found", feed.Name, len(feedData.Channel.Item))
	return nil
}

func ScrapeFeedsConcurrently(s *cli.State, feeds []database.Feed, workerCount int) {
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
