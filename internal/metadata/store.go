package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const StateFileName = "sqill.json"
const ManifestFileName = "sqill.json"

type Manifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

type InstalledEntry struct {
	Version     string `json:"version"`
	Source      string `json:"source"`
	InstalledAt string `json:"installed_at"`
	Description string `json:"description,omitempty"`
}

type State struct {
	Installed  map[string]InstalledEntry `json:"installed"`
	Registries []string                  `json:"registries"`
}

func NewState() State {
	return State{
		Installed:  map[string]InstalledEntry{},
		Registries: []string{},
	}
}

type Store interface {
	Load() (State, error)
	Save(State) error
	IsInstalled(name string) bool
	Get(name string) (InstalledEntry, error)
	Add(name string, entry InstalledEntry) error
	Remove(name string) error
	Path() string
}

type FileStore struct {
	dir  string
	mu   sync.Mutex
	path string
}

func NewFileStore(dir string) (*FileStore, error) {
	if dir == "" {
		return nil, errors.New("metadata: directory is empty")
	}
	return &FileStore{
		dir:  dir,
		path: filepath.Join(dir, StateFileName),
	}, nil
}

func (s *FileStore) Path() string { return s.path }

func (s *FileStore) Load() (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewState(), nil
		}
		return State{}, fmt.Errorf("read state: %w", err)
	}

	if len(data) == 0 {
		return NewState(), nil
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("parse state: %w", err)
	}
	if state.Installed == nil {
		state.Installed = map[string]InstalledEntry{}
	}
	if state.Registries == nil {
		state.Registries = []string{}
	}
	return state, nil
}

func (s *FileStore) Save(state State) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if state.Installed == nil {
		state.Installed = map[string]InstalledEntry{}
	}
	if state.Registries == nil {
		state.Registries = []string{}
	}

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	tmp, err := os.CreateTemp(s.dir, ".sqill-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename temp: %w", err)
	}
	return nil
}

func (s *FileStore) IsInstalled(name string) bool {
	state, err := s.Load()
	if err != nil {
		return false
	}
	_, ok := state.Installed[name]
	return ok
}

func (s *FileStore) Get(name string) (InstalledEntry, error) {
	state, err := s.Load()
	if err != nil {
		return InstalledEntry{}, err
	}
	entry, ok := state.Installed[name]
	if !ok {
		return InstalledEntry{}, fmt.Errorf("skill %q not installed", name)
	}
	return entry, nil
}

func (s *FileStore) Add(name string, entry InstalledEntry) error {
	state, err := s.Load()
	if err != nil {
		return err
	}
	if state.Installed == nil {
		state.Installed = map[string]InstalledEntry{}
	}
	state.Installed[name] = entry
	return s.Save(state)
}

func (s *FileStore) Remove(name string) error {
	state, err := s.Load()
	if err != nil {
		return err
	}
	if _, ok := state.Installed[name]; !ok {
		return fmt.Errorf("skill %q not installed", name)
	}
	delete(state.Installed, name)
	return s.Save(state)
}

func LoadManifest(dir string) (Manifest, error) {
	path := filepath.Join(dir, ManifestFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}
	return m, nil
}

func WriteManifest(dir string, m Manifest) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	path := filepath.Join(dir, ManifestFileName)
	tmp, err := os.CreateTemp(dir, ".manifest-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename temp: %w", err)
	}
	return nil
}

func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func SortedNames(state State) []string {
	names := make([]string, 0, len(state.Installed))
	for n := range state.Installed {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
