# Gator

A CLI feed aggregator built in Go. Fetches RSS feeds and stores posts in PostgreSQL so you can read them from the terminal.

---

## Requirements

- [Go 1.22+](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)

---

## Install

```bash
go install github.com/danielmiranda22/gator@latest
```

This compiles and puts the `gator` binary in your `$GOPATH/bin`. No need for `go run .` after this.

---

## Setup

**1. Create the database**

```bash
# Linux
sudo -u postgres psql
# Mac
psql postgres
```

```sql
CREATE DATABASE gator;
```

**2. Run migrations**

```bash
goose -dir sql/schema postgres "postgres://localhost:5432/gator" up
```

**3. Create the config file**

Gator looks for `~/.gatorconfig.json`. Create it:

```json
{
  "db_url": "postgres://localhost:5432/gator?sslmode=disable"
}
```

---

## Commands

**Register and log in**

```bash
gator register alice    # create a new user + log in as alice
gator login alice       # log in as an existing user
gator users             # list all users (* marks the current user)
```

**Manage feeds**

```bash
gator addfeed "TechCrunch" https://techcrunch.com/feed/   # add a feed + auto-follow it
gator feeds                                                # list all feeds
gator follow https://news.ycombinator.com/rss             # follow an existing feed
gator following                                           # list feeds you follow
gator unfollow https://news.ycombinator.com/rss           # unfollow a feed
```

**Aggregate and browse**

```bash
# run the aggregator — leave this running in a separate terminal
gator agg 1m       # fetch feeds every 1 minute
gator agg 30s      # fetch feeds every 30 seconds

# browse posts from feeds you follow
gator browse       # show 2 most recent posts (default)
gator browse 10    # show 10 most recent posts
```

**Other**

```bash
gator reset        # delete all users (useful for development)
```

---

## Development

```bash
git clone https://github.com/danielmiranda22/gator
cd gator
go build .
./gator register alice
```

---

## Tech stack

- [Go](https://go.dev) — language
- [PostgreSQL](https://www.postgresql.org) — database
- [sqlc](https://sqlc.dev) — type-safe SQL code generation
- [goose](https://github.com/pressly/goose) — database migrations
- [lib/pq](https://github.com/lib/pq) — PostgreSQL driver
