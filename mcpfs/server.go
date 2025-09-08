package mcpfs

import (
	"context"
	"log/slog"

	"github.com/jlrickert/go-std/pkg"
)

var (
	Version = "dev"
)

// App is a minimal application type used by the serve command.
// It logs when Run is called and uses cfg.Foo to demonstrate config usage.
type App struct {
	Cfg      *Config
	Logger   *slog.Logger
	Services *Services
}

// NewApp constructs an App. It accepts the minimal fields the serve command uses.
func NewApp(cfg *Config, logger *slog.Logger, services *Services) *App {
	if logger == nil {
		logger = std.NewDiscardLogger()
	}
	if services == nil {
		services = NewDefaultServices()
	}
	return &App{Cfg: cfg, Logger: logger, Services: services}
}

// Run executes the application. It logs that it's running and the cfg.Foo
// value. It returns nil on success.
func (a *App) Run(ctx context.Context) error {
	a.Logger.LogAttrs(ctx, slog.LevelInfo, "application running", slog.Any("config", a.Cfg))
	return nil
}
