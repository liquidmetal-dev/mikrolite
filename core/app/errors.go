package app

import "errors"

var (
	ErrNotImplemented  = errors.New("Not implemented")
	ErrVmSpecRequired  = errors.New("VM spec is required")
	ErrNameRequired    = errors.New("name is required")
	ErrNoKernelSource  = errors.New("no kernel source supplied")
	ErrVMAlreadyExists = errors.New("VM already exists")
)
