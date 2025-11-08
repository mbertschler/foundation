package auth

import "github.com/mbertschler/foundation/db"

type Handler struct {
	DB *db.DB
}

func NewHandler(database *db.DB) *Handler {
	return &Handler{
		DB: database,
	}
}
