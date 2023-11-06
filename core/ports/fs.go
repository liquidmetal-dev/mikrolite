package ports

import "github.com/spf13/afero"

// FileSystem represents a driven port for the filesystem.
type FileSystem interface {
	afero.Fs
}
