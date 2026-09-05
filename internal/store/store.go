// Package store persists secrets as a DPAPI-encrypted JSON file.
package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"tokenfinder/internal/crypto"
)

const vaultVersion = 1

type vaultFile struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
}

// Store is an in-memory list of entries backed by an encrypted file.
type Store struct {
	mu      sync.Mutex
	path    string
	entries []Entry
}

// Open loads the vault at path, creating an empty one if it does not exist.
func Open(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// Path returns the vault file location.
func (s *Store) Path() string { return s.path }

func (s *Store) load() error {
	raw, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.entries = nil
		return nil
	}
	if err != nil {
		return err
	}
	plain, err := crypto.Unprotect(raw)
	if err != nil {
		return fmt.Errorf("decrypt vault: %w", err)
	}
	var vf vaultFile
	if err := json.Unmarshal(plain, &vf); err != nil {
		return fmt.Errorf("parse vault: %w", err)
	}
	s.entries = vf.Entries
	return nil
}

// save writes the vault atomically: encrypt, write temp file, rename.
func (s *Store) save() error {
	plain, err := json.MarshalIndent(vaultFile{Version: vaultVersion, Entries: s.entries}, "", "  ")
	if err != nil {
		return err
	}
	enc, err := crypto.Protect(plain)
	if err != nil {
		return fmt.Errorf("encrypt vault: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, enc, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// List returns a name-sorted copy of all entries.
func (s *Store) List() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return sortedCopy(s.entries)
}

// Search returns entries whose name, kind or note contains query (case-insensitive).
func (s *Store) Search(query string) []Entry {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return s.List()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Entry
	for _, e := range s.entries {
		if strings.Contains(strings.ToLower(e.Name), q) ||
			strings.Contains(strings.ToLower(e.Kind), q) ||
			strings.Contains(strings.ToLower(e.Note), q) {
			out = append(out, e)
		}
	}
	return sortedCopy(out)
}

// Get looks an entry up by ID.
func (s *Store) Get(id string) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.entries {
		if e.ID == id {
			return e, true
		}
	}
	return Entry{}, false
}

// Upsert inserts a new entry (empty ID) or replaces an existing one.
func (s *Store) Upsert(e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	e.UpdatedAt = now
	if e.ID == "" {
		e.ID = newID()
		e.CreatedAt = now
		s.entries = append(s.entries, e)
		return s.save()
	}
	for i := range s.entries {
		if s.entries[i].ID == e.ID {
			e.CreatedAt = s.entries[i].CreatedAt
			s.entries[i] = e
			return s.save()
		}
	}
	return fmt.Errorf("entry %s not found", e.ID)
}

// Delete removes an entry by ID.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.entries {
		if s.entries[i].ID == id {
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("entry %s not found", id)
}

func sortedCopy(in []Entry) []Entry {
	out := make([]Entry, len(in))
	copy(out, in)
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
