package session

import "errors"

var (
	ErrSessionNotExists          = errors.New("repo.session: session doesn't exists")
	ErrRefreshTokenAlreadyExists = errors.New("repo.session: this refresh token already exists")
)

func IsErrSessionNotExists(err error) bool {
	return errors.Is(err, ErrSessionNotExists)
}

func IsErrRefreshTokenAlreadyExists(err error) bool {
	return errors.Is(err, ErrRefreshTokenAlreadyExists)
}
