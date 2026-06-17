package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/danielmiranda22/gator/internal/cli"
	"github.com/danielmiranda22/gator/internal/database"
	"github.com/danielmiranda22/gator/ui"
	"github.com/google/uuid"
)

func Login(s *cli.State, cmd cli.Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: login <username>")
	}
	username := cmd.Args[0]

	// login requires the user to exist in the DB
	_, err := s.DB.GetUser(s.Ctx, username)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%suser %s%q%s%s does not exist%s", ui.ColorRed, ui.ColorBold, username, ui.ColorRed, ui.ColorReset, ui.ColorReset)
	}
	if err != nil {
		return fmt.Errorf("error looking up user: %v", err)
	}

	if err := s.Cfg.SetUser(username); err != nil {
		return err
	}
	fmt.Printf("%s%sLogged in as %s%s\n", ui.ColorGreen, ui.ColorBold, username, ui.ColorReset)
	return nil
}

func Register(s *cli.State, cmd cli.Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: register <username>")
	}
	username := cmd.Args[0]

	// check if user already exists
	_, err := s.DB.GetUser(s.Ctx, username)
	if err == nil {
		// no error = user found = already exists
		return fmt.Errorf("%suser %s%q%s%s already exists%s", ui.ColorRed, ui.ColorBold, username, ui.ColorRed, ui.ColorReset, ui.ColorReset)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		// real DB error
		return fmt.Errorf("error checking for user: %v", err)
	}

	// create the user
	newUser, err := s.DB.CreateUser(s.Ctx, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	// log in as the new user
	if err := s.Cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Printf("User created successfully: %+v\n", newUser)
	return nil
}

func Reset(s *cli.State, cmd cli.Command) error {
	if err := s.DB.DeleteAllUsers(s.Ctx); err != nil {
		return err
	}
	fmt.Println("Users reset successfully")
	return nil
}

func ListUsers(s *cli.State, cmd cli.Command) error {
	users, err := s.DB.GetAllUsers(s.Ctx)
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
