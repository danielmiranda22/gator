package cli

import (
	"context"

	"github.com/danielmiranda22/gator/internal/config"
	"github.com/danielmiranda22/gator/internal/service"
)

type State struct {
	Cfg      *config.Config
	Ctx      context.Context
	Services *service.Services
}
