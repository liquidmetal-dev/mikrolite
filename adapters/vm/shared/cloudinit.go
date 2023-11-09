package shared

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mikrolite/mikrolite/cloudinit"
	"github.com/mikrolite/mikrolite/core/domain"
	"github.com/mikrolite/mikrolite/core/ports"
)

func CreateCloudInitImage(ctx context.Context, includeNetworkConfig bool, vm *domain.VM, ss ports.StateService, ds ports.DiskService) (string, error) {
	cloudInitFile := filepath.Join(ss.Root(), "cloud-init.img")

	files := []ports.DiskFile{}
	for k, v := range vm.Status.Metadata {
		if !cloudinit.IsCloudInitKey(k) {
			continue
		}

		if !includeNetworkConfig && k == cloudinit.NetworkConfigDataKey {
			continue
		}

		dest := fmt.Sprintf("/%s", k)
		files = append(files, ports.DiskFile{
			Path:          dest,
			ContentBase64: v,
		})
	}

	input := ports.DiskCreateInput{
		Path:       cloudInitFile,
		Size:       "8Mb",
		VolumeName: cloudinit.VolumeName,
		Type:       ports.DiskTypeFat32,
		Overwrite:  true,
		Files:      files,
	}
	if err := ds.Create(ctx, input); err != nil {
		return "", fmt.Errorf("creating cloud-init volume %s: %w", cloudInitFile, err)
	}

	return cloudInitFile, nil
}
