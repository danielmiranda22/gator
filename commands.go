package main

import "fmt"

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
