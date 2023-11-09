package ports

import "context"

// DiskService is a driven port for a service that creates disk images.
type DiskService interface {
	// Create will create a new disk.
	Create(ctx context.Context, input DiskCreateInput) error
}

// DiskType represents the type of disk.
type DiskType int

const (
	// DiskTypeFat32 is a FAT32 compatible filesystem.
	DiskTypeFat32 DiskType = iota
	// DiskTypeISO9660 is an iso filesystem.
	DiskTypeISO9660
)

// DiskCreateInput are the input options for creating a disk.
type DiskCreateInput struct {
	//Path is the filesystem path of where to create the disk.
	Path string
	// Size is how big the disk should be. It uses human readable formats
	// such as 8Mb, 10Kb.
	Size string
	// VolumeName is the name to give to the volume.
	VolumeName string
	// Type is the type of disk to create.
	Type DiskType
	// Files are the files to create in the new disk.
	Files []DiskFile
	// Overwrite specifies if the image file already exists whether
	// we should overwrite it or return an error.
	Overwrite bool
}

// DiskFile represents a file to create in a disk.
type DiskFile struct {
	// Path is the path in the disk image for the file.
	Path string
	// ContentBase64 is the content of the file encoded as base64.
	ContentBase64 string
}
