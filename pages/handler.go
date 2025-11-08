package pages

import (
	"github.com/mbertschler/foundation/auth"
	"github.com/mbertschler/foundation/db"
	"github.com/mbertschler/foundation/server/broadcast"
)

type Handler struct {
	DB        *db.DB
	Auth      *auth.Handler
	Broadcast *broadcast.Broadcaster
}

func NewHandler(database *db.DB, broadcaster *broadcast.Broadcaster) *Handler {
	return &Handler{
		DB:        database,
		Auth:      auth.NewHandler(database),
		Broadcast: broadcaster,
	}
}
