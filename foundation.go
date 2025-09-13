package foundation

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Context struct {
	Context context.Context
	Config  *Config
}

type Request struct {
	*Context
	Request *http.Request
	Params  httprouter.Params
}
