package main

import (
	"fmt"
	"log"
	"os"

	"github.com/danielmiranda22/gator/internal/config"
)

// application state — grows later when DB is added
type state struct {
	cfg *config.Config
}

// a command the user typed
type command struct {
	name string
	args []string
}

// registry of all commands
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

// handlerLogin — sets the current user in the config
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("usage: login <username>")
	}

	username := cmd.args[0]
	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Printf("Logged in as %s\n", username)
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	s := &state{cfg: &cfg}

	cmds := commands{}
	cmds.register("login", handlerLogin)

	args := os.Args // ["gator", "login", "dano"]
	if len(args) < 2 {
		log.Fatal("usage: gator <command> [args...]")
	}

	cmd := command{
		name: args[1],  // "login"
		args: args[2:], // ["dano"]
	}

	if err := cmds.run(s, cmd); err != nil {
		log.Fatal(err)
	}
}
