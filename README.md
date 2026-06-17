# Gator

A terminal-first RSS aggregator built in Go.

Gator lets you subscribe to RSS feeds, aggregate posts into PostgreSQL, browse articles from the command line, search content, manage subscriptions, and read posts through an interactive TUI.

> Originally built as part of the [Boot.dev's](https://boot.dev) curriculum and now actively extended with additional features, architectural improvements, and experiments in Go CLI development.

---

## Features

* User management and login system
* RSS feed subscription management
* Background feed aggregation
* Concurrent feed fetching
* Follow and unfollow feeds
* Browse posts from followed feeds
* Interactive terminal UI (TUI)
* Search posts by keyword
* Like and unlike posts
* View liked posts
* Pagination, filtering, and sorting support
* PostgreSQL-backed persistence

---

## Requirements

* Go 1.22+
* PostgreSQL 14+

---

## Installation

```bash
go install github.com/danielmiranda22/gator@latest
```

Verify the installation:

```bash
gator
```

---

## Configuration

Create a configuration file at:

```bash
~/.gatorconfig.json
```

Example:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable"
}
```

> For local PostgreSQL development, `sslmode=disable` is typically required.

---

## Database Setup

Create a PostgreSQL database:

```sql
CREATE DATABASE gator;
```

Run the migrations:

```bash
goose postgres "postgres://username:password@localhost:5432/gator?sslmode=disable" up
```

Or apply the SQL migration files manually from:

```text
sql/schema/
```

---

## Quick Start

### Create and log into a user

```bash
gator register john wick
gator login john wick
```

### Add a feed

```bash
gator addfeed "Hacker News" https://hnrss.org/frontpage
```

Feeds added by a user are automatically followed.

### Follow an existing feed

```bash
gator follow https://hnrss.org/frontpage
```

### Start the aggregator

Run in a separate terminal:

```bash
gator agg 1m
```

Fetch feeds every minute using a single worker.

Multiple workers can be used:

```bash
gator agg 1m 5
```

This fetches feeds concurrently using 5 workers.

---

## Browsing Posts

Show the latest posts:

```bash
gator browse
```

Show more posts:

```bash
gator browse --limit 10
```

Browse older posts first:

```bash
gator browse --sort oldest
```

Filter by title:

```bash
gator browse --filter golang
```

Paginate results:

```bash
gator browse --page 2
```

Combine options:

```bash
gator browse --limit 10 --sort oldest --page 3
```

Show help:

```bash
gator browse --help
```

---

## Interactive TUI

Launch the terminal UI:

```bash
gator browse tui
```

View liked posts in the TUI:

```bash
gator browse liked tui
```

---

## Search

Search posts from feeds you follow:

```bash
gator search golang
```

---

## Likes

Like a post:

```bash
gator like <post_id>
```

Remove a like:

```bash
gator unlike <post_id>
```

View liked posts:

```bash
gator browse liked
```

---

## Feed Management

List all feeds:

```bash
gator feeds
```

List followed feeds:

```bash
gator following
```

Unfollow a feed:

```bash
gator unfollow <feed_url>
```

---

## User Management

List users:

```bash
gator users
```

Reset all users:

```bash
gator reset
```

---

## Available Commands

| Command                    | Description             |
| -------------------------- | ----------------------- |
| `register <name>`          | Create a user           |
| `login <name>`             | Log in as a user        |
| `users`                    | List all users          |
| `reset`                    | Delete all users        |
| `addfeed <name> <url>`     | Add a feed              |
| `feeds`                    | List all feeds          |
| `follow <url>`             | Follow a feed           |
| `following`                | List followed feeds     |
| `unfollow <url>`           | Unfollow a feed         |
| `agg <duration> [workers]` | Run feed aggregation    |
| `browse`                   | Browse posts            |
| `browse tui`               | Launch TUI              |
| `browse liked`             | Show liked posts        |
| `browse liked tui`         | Show liked posts in TUI |
| `search <term>`            | Search posts            |
| `like <post_id>`           | Like a post             |
| `unlike <post_id>`         | Remove a like           |

---

## Project Structure

```text
internal/
├── app/
├── commands/
├── config/
├── database/
├── handlers/
├── rss/
├── service/

ui/
├── colors.go
├── render.go
└── tui.go
```

---

## Roadmap

Potential future improvements:

* OPML import/export
* Feed categories and tags
* Read/unread tracking
* Full-text search improvements
* Saved searches
* Multi-feed filtering
* Web UI companion
* Docker deployment

---

## License

MIT
