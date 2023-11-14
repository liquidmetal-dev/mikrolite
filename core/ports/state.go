package ports

import "github.com/mikrolite/mikrolite/core/domain"

type StateService interface {
	Root() string

	GetVM() (*domain.VM, error)
	SaveVM(vm *domain.VM) error
	ListVMs() ([]*domain.VM, error)

	LogPath() string
	StdoutPath() string
	StderrPath() string

	GetMetadata() (map[string]string, error)
	SaveMetadata(metadata map[string]string) error

	GetPID() (int, error)
	SavePID(pid int) error
}
