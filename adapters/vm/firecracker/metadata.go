package firecracker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mikrolite/mikrolite/cloudinit"
	"github.com/mikrolite/mikrolite/core/domain"
	"github.com/mikrolite/mikrolite/defaults"
)

type metadata struct {
	Latest map[string]string `json:"latest"`
}

func (f *Provider) saveMetadata(vm *domain.VM) (string, error) {
	metadataFile := filepath.Join(f.ss.Root(), "metadata.json")

	meta := &metadata{
		Latest: map[string]string{},
	}

	for key, value := range vm.Status.Metadata {
		if key == cloudinit.NetworkConfigDataKey {
			continue
		}

		decodedValue, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return "", fmt.Errorf("decoding metadata %s: %w", key, err)
		}

		meta.Latest[key] = string(decodedValue)
	}

	data, err := json.MarshalIndent(meta, "", " ")
	if err != nil {
		return "", fmt.Errorf("marshalling metdata: %w", err)
	}

	file, err := f.fs.OpenFile(metadataFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaults.DataFilePerm)
	if err != nil {
		return "", fmt.Errorf("opening metadata file %s: %w", metadataFile, err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return "", fmt.Errorf("writing metdata to file: %w", err)
	}

	return metadataFile, nil
}
