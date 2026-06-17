package handlers

import (
	"fmt"

	"github.com/danielmiranda22/gator/internal/cli"
	"github.com/danielmiranda22/gator/ui"
)

func Login(s *cli.State, cmd cli.Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: login <username>")
	}

	username := cmd.Args[0]

	_, err := s.Services.Users.Login(
		s.Ctx,
		username,
	)

	if err != nil {
		return err
	}

	if err := s.Cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Printf(
		"%s%sLogged in as %s%s\n",
		ui.ColorGreen,
		ui.ColorBold,
		username,
		ui.ColorReset,
	)

	return nil
}

func Register(s *cli.State, cmd cli.Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: register <username>")
	}

	username := cmd.Args[0]

	user, err := s.Services.Users.Register(
		s.Ctx,
		username,
	)

	if err != nil {
		return err
	}

	if err := s.Cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Printf(
		"User created successfully: %+v\n",
		user,
	)

	return nil
}

func Reset(s *cli.State, cmd cli.Command) error {
	if err := s.Services.Users.Reset(s.Ctx); err != nil {
		return err
	}
	fmt.Println("Users reset successfully")
	return nil
}

func ListUsers(s *cli.State, cmd cli.Command) error {
	users, err := s.Services.Users.ListUsers(
		s.Ctx,
	)
	if err != nil {
		return fmt.Errorf("error getting all users: %v", err)
	}

	for _, user := range users {
		if user.Name == s.Cfg.CurrentUserName {
			fmt.Printf("* %v  %s(current)%s\n", user.Name, ui.ColorCyan, ui.ColorReset)
			continue
		}
		fmt.Printf("* %v\n", user.Name)
	}
	return nil
}
