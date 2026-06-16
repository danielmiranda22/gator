package ui

import (
	"fmt"
	"time"

	"github.com/danielmiranda22/gator/internal/database"
	"github.com/google/uuid"
)

func PrintPosts[T any](posts []T, userName string, page int, limit int, sort string, filter string) {
	if len(posts) == 0 {
		fmt.Printf("%sNo posts found for %s on page %d.%s\n", ColorYellow, userName, page, ColorReset)
		if filter != "" {
			fmt.Printf("%sTry a different filter or a lower page number.%s\n", ColorGray, ColorReset)
		}
		return
	}

	fmt.Printf("\n%s%sPosts for %s%s", ColorBold, ColorCyan, userName, ColorReset)
	fmt.Printf("%s — page %d — sort: %s — limit: %d%s\n", ColorGray, page, sort, limit, ColorReset)
	if filter != "" {
		fmt.Printf("%sFilter:%s %s\n", ColorYellow, ColorReset, filter)
	}
	fmt.Printf("%s════════════════════════════════════════════════════════════%s\n", ColorGray, ColorReset)

	for _, post := range posts {
		// Use type switch if your sqlc-generated rows have different struct types, or
		// just copy your previous print loop for each branch if needed.
		switch p := any(post).(type) {
		case database.GetPostsForUserWithPaginationRow:
			RenderPost(p.ID, p.Title, p.FeedName, p.Url, p.Description.String, p.PublishedAt.Time)
		case database.GetPostsForUserOldestWithPaginationRow:
			RenderPost(p.ID, p.Title, p.FeedName, p.Url, p.Description.String, p.PublishedAt.Time)
		case database.GetPostsForUserFilterByTitleRow:
			RenderPost(p.ID, p.Title, p.FeedName, p.Url, p.Description.String, p.PublishedAt.Time)
		case database.SearchPostsForUserRow:
			RenderPost(p.ID, p.Title, p.FeedName, p.Url, p.Description.String, p.PublishedAt.Time)
		case database.GetLikedPostsForUserRow:
			RenderPost(p.ID, p.Title, p.FeedName, p.Url, p.Description.String, p.PublishedAt.Time)
		default:
			fmt.Printf("%sUnknown post type: %T%s\n", ColorRed, post, ColorReset)
		}
	}
}

func RenderPost(id uuid.UUID, title, feedName, url, description string, publishedAt time.Time) {
	fmt.Printf("%s📅 %s%s  %s%s\n",
		ColorBlue,
		publishedAt.Format("02 Jan 2006"),
		ColorGray,
		feedName,
		ColorReset,
	)

	fmt.Printf("%s%s%s\n", ColorBold, title, ColorReset)

	fmt.Printf("    %sID: %s%s\n", ColorMagenta, id, ColorReset)

	if description != "" {
		fmt.Printf("    %s%s%s\n", ColorGray, truncate(description, 140), ColorReset)
	}

	fmt.Printf("    %s🔗 %s%s\n", ColorGreen, url, ColorReset)
	fmt.Printf("%s────────────────────────────────────────%s\n", ColorGray, ColorReset)
}

func truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max-3] + "..."
}
