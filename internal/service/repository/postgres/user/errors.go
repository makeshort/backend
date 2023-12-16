package user

import "errors"

var (
	ErrUserNotExists     = errors.New("repo.user: user doesn't exists")
	ErrUserAlreadyExists = errors.New("repo.user: user already exists")
)

func IsErrUserNotExists(err error) bool {
	return errors.Is(err, ErrUserNotExists)
}

func IsErrUserAlreadyExists(err error) bool {
	return errors.Is(err, ErrUserAlreadyExists)
}
