package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/danielmiranda22/gator/internal/database"
	"github.com/danielmiranda22/gator/ui"
	"github.com/google/uuid"
)

type UserService struct {
	db *database.Queries
}

func NewUserService(db *database.Queries) *UserService {
	return &UserService{
		db: db,
	}
}

func (s *UserService) Register(
	ctx context.Context,
	username string,
) (database.User, error) {

	_, err := s.db.GetUser(ctx, username)

	if err == nil {
		// no error = user found = already exists
		return database.User{}, fmt.Errorf("%suser %s%q%s%s already exists%s", ui.ColorRed, ui.ColorBold, username, ui.ColorRed, ui.ColorReset, ui.ColorReset)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return database.User{}, err
	}

	user, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})

	if err != nil {
		return database.User{}, err
	}

	return user, nil
}

func (s *UserService) Login(
	ctx context.Context,
	username string,
) (database.User, error) {

	user, err := s.db.GetUser(ctx, username)

	if errors.Is(err, sql.ErrNoRows) {
		return database.User{}, fmt.Errorf("%suser %s%q%s%s does not exist%s", ui.ColorRed, ui.ColorBold, username, ui.ColorRed, ui.ColorReset, ui.ColorReset)
	}

	if err != nil {
		return database.User{}, err
	}

	return user, nil
}

func (s *UserService) ListUsers(
	ctx context.Context,
) ([]database.User, error) {

	return s.db.GetAllUsers(ctx)
}

func (s *UserService) Reset(
	ctx context.Context,
) error {

	return s.db.DeleteAllUsers(ctx)
}
