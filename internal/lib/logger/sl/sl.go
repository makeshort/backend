package sl

import (
	"log/slog"
)

// Err adds on error field to log.
func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
