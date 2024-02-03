package model

import (
	"errors"

	"github.com/timshannon/badgerhold/v4"
)

var ErrNotFound = errors.New("not found")

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.Is(err, badgerhold.ErrNotFound)
}
