package commands

import (
	"context"
	"fmt"

	"github.com/danielmiranda22/gator/internal/cli"
	"github.com/danielmiranda22/gator/internal/database"
)

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
