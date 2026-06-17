package cli

import (
	"fmt"
)

type Command struct {
	Name string
	Args []string
}

type Commands map[string]func(*State, Command) error

func (c Commands) Register(name string, handler func(*State, Command) error) {
	c[name] = handler
}

func (c Commands) Run(s *State, cmd Command) error {
	handler, exists := c[cmd.Name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
}
