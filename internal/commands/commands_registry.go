package commands

import (
	"context"
	"fmt"

	"github.com/danielmiranda22/gator/internal/cli"
	"github.com/danielmiranda22/gator/internal/database"
	"github.com/danielmiranda22/gator/internal/handlers"
)

func RegisterCommands(cmds *cli.Commands) {
	cmds.Register("login", handlers.Login)
	cmds.Register("register", handlers.Register)
	cmds.Register("reset", handlers.Reset)

	cmds.Register("users", handlers.ListUsers)

	cmds.Register("feeds", handlers.ListFeeds)
	cmds.Register("addfeed", middlewareLoggedIn(handlers.AddFeed))

	cmds.Register("agg", handlers.Agg)

	cmds.Register("follow", middlewareLoggedIn(handlers.Follow))
	cmds.Register("following", middlewareLoggedIn(handlers.Following))
	cmds.Register("unfollow", middlewareLoggedIn(handlers.Unfollow))

	cmds.Register("browse", middlewareLoggedIn(handlers.Browse))
	cmds.Register("search", middlewareLoggedIn(handlers.Search))

	cmds.Register("like", middlewareLoggedIn(handlers.Like))
	cmds.Register("unlike", middlewareLoggedIn(handlers.Unlike))
}

func middlewareLoggedIn(
	handler func(*cli.State, cli.Command, database.User) error,
) func(*cli.State, cli.Command) error {

	return func(s *cli.State, cmd cli.Command) error {
		user, err := s.DB.GetUser(
			context.Background(),
			s.Cfg.CurrentUserName,
		)

		if err != nil {
			return fmt.Errorf("not logged in: %w", err)
		}

		return handler(s, cmd, user)
	}
}
