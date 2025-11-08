package pages

import (
	"github.com/mbertschler/foundation/auth"
	"github.com/mbertschler/foundation/db"
)

type Handler struct {
	DB   *db.DB
	Auth *auth.Handler
}

func NewHandler(database *db.DB) *Handler {
	return &Handler{
		DB:   database,
		Auth: auth.NewHandler(database),
	}
}
