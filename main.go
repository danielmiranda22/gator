package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/danielmiranda22/gator/internal/config"
	"github.com/danielmiranda22/gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	programState := &state{
		cfg: &cfg,
		db:  database.New(db),
	}

	cmds := commands{}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerListUsers)
	cmds.register("feeds", handlerListFeeds)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("agg", handlerAgg)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))
	cmds.register("search", middlewareLoggedIn(handlerSearch))

	args := os.Args
	if len(args) < 2 {
		log.Fatal("usage: gator <command> [args...]")
	}

	cmd := command{
		name: args[1],
		args: args[2:],
	}
	if err := cmds.run(programState, cmd); err != nil {
		log.Fatal(err)
	}
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("not logged in: %v", err)
		}
		return handler(s, cmd, user)
	}
}
