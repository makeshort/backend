package session

import "errors"

var (
	errSessionNotExists          = errors.New("repo.session: session doesn't exists")
	errRefreshTokenAlreadyExists = errors.New("repo.session: this refresh token already exists")
)

func IsErrSessionNotExists(err error) bool {
	return errors.Is(err, errSessionNotExists)
}

func IsErrRefreshTokenAlreadyExists(err error) bool {
	return errors.Is(err, errRefreshTokenAlreadyExists)
}
