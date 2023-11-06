package vm

import (
	"errors"
	"fmt"
)

func NewUnknownProvider(name string) error {
	return unknowProvider{
		name: name,
	}
}

type unknowProvider struct {
	name string
}

func (e unknowProvider) Error() string {
	return fmt.Sprintf("%s in an unknown vm provider type", e.name)
}

func IsUnknownProvider(err error) bool {
	e := &unknowProvider{}

	return errors.As(err, e)
}
