package foundation

import (
	"context"
)

type Context struct {
	Context context.Context
	Config  *Config
}
