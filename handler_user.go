package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/danielmiranda22/gator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("usage: login <username>")
	}
	username := cmd.args[0]

	// login requires the user to exist in the DB
	_, err := s.db.GetUser(context.Background(), username)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("user %s does not exist", username)
	}
	if err != nil {
		return fmt.Errorf("error looking up user: %v", err)
	}

	if err := s.cfg.SetUser(username); err != nil {
		return err
	}
	fmt.Printf("Logged in as %s\n", username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("usage: register <username>")
	}
	username := cmd.args[0]

	// check if user already exists
	_, err := s.db.GetUser(context.Background(), username)
	if err == nil {
		// no error = user found = already exists
		return fmt.Errorf("user %s already exists", username)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		// real DB error
		return fmt.Errorf("error checking for user: %v", err)
	}

	// create the user
	newUser, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	// log in as the new user
	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Printf("User created successfully: %+v\n", newUser)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if err := s.db.DeleteAllUsers(context.Background()); err != nil {
		return err
	}
	fmt.Println("Users reset successfully")
	return nil
}

func handlerListUsers(s *state, cmd command) error {
	users, err := s.db.GetAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all users: %v", err)
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %v (current)\n", user.Name)
			continue
		}
		fmt.Printf("* %v\n", user.Name)
	}
	return nil
}
