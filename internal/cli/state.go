package cli

import (
	"context"

	"github.com/danielmiranda22/gator/internal/config"
	"github.com/danielmiranda22/gator/internal/database"
	"github.com/danielmiranda22/gator/internal/service"
)

type State struct {
	DB       *database.Queries
	Cfg      *config.Config
	Ctx      context.Context
	Services *service.Services
}
