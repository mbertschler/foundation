package foundation

import (
	"context"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/uptrace/bun"
)

type Context struct {
	Context context.Context
	Config  *Config
	DB      *DB
}

type Request struct {
	*Context
	Request *http.Request
	Params  httprouter.Params
}

type DB struct {
	Users UserDB
}

type UserDB interface {
	ByUsername(ctx context.Context, username string) (*User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Insert(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, userID int64) error
	All(ctx context.Context) ([]*User, error)
}

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID             int64     `bun:"id,pk,autoincrement"`
	DisplayName    string    `bun:"display_name,notnull"`
	UserName       string    `bun:"user_name,unique"`
	HashedPassword string    `bun:"hashed_password,notnull"`
	CreatedAt      time.Time `bun:"created_at,nullzero,notnull"`
	UpdatedAt      time.Time `bun:"updated_at,nullzero,notnull"`
}
