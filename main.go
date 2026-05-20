package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/danielmiranda22/gator/internal/config"
	"github.com/danielmiranda22/gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands map[string]func(*state, command) error

func (c commands) register(name string, handler func(*state, command) error) {
	c[name] = handler
}

func (c commands) run(s *state, cmd command) error {
	handler, exists := c[cmd.name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return handler(s, cmd)
}

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

	s := &state{
		cfg: &cfg,
		db:  database.New(db),
	}

	cmds := commands{}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)

	args := os.Args
	if len(args) < 2 {
		log.Fatal("usage: gator <command> [args...]")
	}

	cmd := command{name: args[1], args: args[2:]}
	if err := cmds.run(s, cmd); err != nil {
		log.Fatal(err)
	}
}
