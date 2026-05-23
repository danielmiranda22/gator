package main

import (
	"fmt"
)

type command struct {
	Name string
	Args []string
}

type commands map[string]func(*state, command) error

func (c commands) register(name string, handler func(*state, command) error) {
	c[name] = handler
}

func (c commands) run(s *state, cmd command) error {
	handler, exists := c[cmd.Name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
}
