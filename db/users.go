package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mbertschler/foundation"
	"github.com/uptrace/bun"
)

var (
	nilUser *foundation.User

	_ foundation.UserDB = (*usersDB)(nil)
)

type usersDB struct {
	db *bun.DB
}

func (u *usersDB) ByID(ctx context.Context, id int64) (*foundation.User, error) {
	var user foundation.User
	err := u.db.NewSelect().Model(&user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *usersDB) ByUsername(ctx context.Context, username string) (*foundation.User, error) {
	var user foundation.User
	err := u.db.NewSelect().Model(&user).Where("user_name = ?", username).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *usersDB) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	user, err := u.ByUsername(ctx, username)
	if user != nil && err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if user == nil {
		return false, errors.New("user is nil")
	}
	return false, err
}

func (u *usersDB) Insert(ctx context.Context, user *foundation.User) error {
	_, err := u.db.NewInsert().Model(user).Exec(ctx)
	return err
}

func (u *usersDB) Update(ctx context.Context, user *foundation.User) error {
	_, err := u.db.NewUpdate().Model(user).WherePK().Exec(ctx)
	return err
}

func (u *usersDB) Delete(ctx context.Context, userID int64) error {
	_, err := u.db.NewDelete().Model(nilUser).Where("id = ?", userID).Exec(ctx)
	return err
}

func (u *usersDB) All(ctx context.Context) ([]*foundation.User, error) {
	var users []*foundation.User
	err := u.db.NewSelect().Model(&users).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
