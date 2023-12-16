package url

import "errors"

var (
	ErrShortUrlAlreadyExists = errors.New("repo.url: short url already exists")
	ErrUrlNotFound           = errors.New("repo.url: url not found")
)

func IsErrShortUrlAlreadyExists(err error) bool {
	return errors.Is(err, ErrShortUrlAlreadyExists)
}

func IsErrUrlNotFound(err error) bool {
	return errors.Is(err, ErrUrlNotFound)
}
