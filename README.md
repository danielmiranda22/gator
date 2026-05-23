# Gator

A CLI tool for aggregating RSS feeds and reading posts from the terminal.

## Requirements

- [Go 1.22+](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)

## Install

```bash
go install github.com/danielmiranda22/gator@latest
```

## Config

Create `~/.gatorconfig.json`:

```json
{
  "db_url": "postgres://username:@localhost:5432/gator?sslmode=disable"
}
```

> `sslmode=disable` is required for local PostgreSQL connections.

## Usage

```bash
# Create a user and log in
gator register <name>
gator login <name>

# Add a feed and follow it
gator addfeed "Feed Name" <url>
gator follow <url>

# Run the aggregator (leave running in a separate terminal)
gator agg 1m

# Browse posts from feeds you follow
gator browse        # default 2 posts
gator browse 10     # show 10 posts
```

## Other commands

| Command                | Description           |
| ---------------------- | --------------------- |
| `gator users`          | List all users        |
| `gator feeds`          | List all feeds        |
| `gator following`      | List feeds you follow |
| `gator unfollow <url>` | Unfollow a feed       |
| `gator reset`          | Delete all users      |
