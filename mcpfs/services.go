package mcpfs

import (
	"github.com/jlrickert/go-std/pkg"
)

// Services holds external dependencies for components in this package.
// Prefer using Services.Env and Services.Clock when wiring tests or production code.
type Services struct {
	Env   std.Env
	Clock std.Clock
}

func NewDefaultServices() *Services {
	return &Services{
		Env:   &std.OsEnv{},
		Clock: std.OsClock{},
	}
}
