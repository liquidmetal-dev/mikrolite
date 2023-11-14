package filesystem

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mikrolite/mikrolite/core/domain"
	"github.com/mikrolite/mikrolite/core/ports"
	"github.com/spf13/afero"
)

const (
	dataFilePerm = 0o644
)

func NewStateService(vmName string, rootStateDir string, fs afero.Fs) (ports.StateService, error) {
	stateDir := filepath.Join(rootStateDir, vmName)

	if vmName != "" {
		if err := fs.MkdirAll(stateDir, dataFilePerm); err != nil {
			return nil, fmt.Errorf("creating state directory %s: %w", stateDir, err)
		}
	}

	return &stateService{
		fs:           fs,
		stateDir:     stateDir,
		rootStateDir: rootStateDir,
	}, nil
}

type stateService struct {
	fs           afero.Fs
	stateDir     string
	rootStateDir string
}

func (s *stateService) GetVM() (*domain.VM, error) {
	exists, err := afero.Exists(s.fs, s.configFileName())
	if err != nil {
		return nil, fmt.Errorf("checking if vm config exists: %w", err)
	}
	if !exists {
		return nil, nil
	}

	vm := &domain.VM{}

	if err := s.readJSONFile(vm, s.configFileName()); err != nil {
		return nil, fmt.Errorf("reading vm from state: %w", err)
	}

	return vm, nil
}

func (s *stateService) SaveVM(vm *domain.VM) error {
	if err := s.writeToFileAsJSON(vm, s.configFileName()); err != nil {
		return fmt.Errorf("saving vm to state: %w", err)
	}

	return nil
}

func (s *stateService) ListVMs() ([]*domain.VM, error) {
	fileInfos, err := afero.ReadDir(s.fs, s.rootStateDir)
	if err != nil {
		return nil, fmt.Errorf("error reading state dir %w", err)
	}

	vms := []*domain.VM{}
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			continue
		}

		s.stateDir = filepath.Join(s.stateDir, fileInfo.Name())

		vm, err := s.GetVM()
		if err != nil {
			return nil, fmt.Errorf("error getting vm: %w", err)
		}

		if vm != nil {
			vms = append(vms, vm)
		}
	}

	s.stateDir = ""

	return vms, nil
}

func (s *stateService) LogPath() string {
	return fmt.Sprintf("%s/vm.log", s.stateDir)
}

func (s *stateService) StdoutPath() string {
	return fmt.Sprintf("%s/vm.stdout", s.stateDir)
}

func (s *stateService) StderrPath() string {
	return fmt.Sprintf("%s/vm.stderr", s.stateDir)
}

func (s *stateService) GetMetadata() (map[string]string, error) {
	meta := metadata{}

	err := s.readJSONFile(&meta, s.metadaFilename())
	if err != nil {
		return nil, fmt.Errorf("firecracker metadata: %w", err)
	}

	return meta.Latest, nil
}

func (s *stateService) SaveMetadata(meta map[string]string) error {
	decoded := &metadata{
		Latest: map[string]string{},
	}

	// Try to base64 decode values, if we can't decode, use them at they are,
	// in case we got them without base64 encoding.
	for key, value := range meta {
		decodedValue, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			decoded.Latest[key] = value
		} else {
			decoded.Latest[key] = string(decodedValue)
		}
	}

	err := s.writeToFileAsJSON(decoded, s.metadaFilename())
	if err != nil {
		return fmt.Errorf("firecracker metadata: %w", err)
	}

	return nil
}

func (s *stateService) GetPID() (int, error) {
	pidFile := s.pidFileName()

	fileExists, err := afero.Exists(s.fs, pidFile)
	if err != nil {
		return -1, fmt.Errorf("checking if pid file %s exists: %w", pidFile, err)
	}
	if !fileExists {
		return 0, nil
	}

	file, err := s.fs.Open(pidFile)
	if err != nil {
		return -1, fmt.Errorf("opening pid file %s: %w", pidFile, err)
	}

	data, err := os.ReadFile(file.Name())
	if err != nil {
		return -1, fmt.Errorf("reading pid file %s: %w", pidFile, err)
	}

	pid, err := strconv.Atoi(string(bytes.TrimSpace(data)))
	if err != nil {
		return -1, fmt.Errorf("converting data to int: %w", err)
	}

	return pid, nil
}

func (s *stateService) SavePID(pid int) error {
	pidFile := s.pidFileName()
	file, err := s.fs.OpenFile(pidFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, dataFilePerm)
	if err != nil {
		return fmt.Errorf("opening pid file %s: %w", pidFile, err)
	}

	defer file.Close()

	_, err = fmt.Fprintf(file, "%d", pid)
	if err != nil {
		return fmt.Errorf("writing pid %d to file %s: %w", pid, pidFile, err)
	}

	return nil
}

func (s *stateService) Root() string {
	return s.stateDir
}

func (s *stateService) pidFileName() string {
	return fmt.Sprintf("%s/vm.pid", s.stateDir)
}

func (s *stateService) configFileName() string {
	return fmt.Sprintf("%s/vm.json", s.stateDir)
}

func (s *stateService) metadaFilename() string {
	return fmt.Sprintf("%s/metadata.json", s.stateDir)
}

func (s *stateService) readJSONFile(cfg interface{}, inputFile string) error {
	file, err := s.fs.Open(inputFile)
	if err != nil {
		return fmt.Errorf("opening file %s: %w", inputFile, err)
	}

	data, err := os.ReadFile(file.Name())
	if err != nil {
		return fmt.Errorf("reading file %s: %w", inputFile, err)
	}

	err = json.Unmarshal(data, cfg)
	if err != nil {
		return fmt.Errorf("unmarshalling: %w", err)
	}

	return nil
}

func (s *stateService) writeToFileAsJSON(cfg interface{}, outputFilePath string) error {
	data, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		return fmt.Errorf("marshalling: %w", err)
	}

	file, err := s.fs.OpenFile(outputFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, dataFilePerm)
	if err != nil {
		return fmt.Errorf("opening output file %s: %w", outputFilePath, err)
	}

	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("writing output file %s: %w", outputFilePath, err)
	}

	return nil
}

type metadata struct {
	Latest map[string]string `json:"latest"`
}
