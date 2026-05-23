package main

import (
	"context"
	"fmt"

	"github.com/danielmiranda22/gator/internal/rss"
)

// Background service, runs in a loop

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

	feed, err := rss.FetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		fmt.Printf("error fetching feed: %v\n", err)
		return
	}

	fmt.Printf("Fetching feed: %s\n", nextFeed.Name)
	for _, item := range feed.Channel.Item {
		fmt.Printf("  * %s\n", item.Title)
	}
}
