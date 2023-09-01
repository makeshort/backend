package user

import "errors"

var (
	ErrUserNotExists     = errors.New("repo.user: user doesn't exists")
	ErrUserAlreadyExists = errors.New("repo.user: user already exists")
)
