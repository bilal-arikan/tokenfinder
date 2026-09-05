package store

import (
	"os"
	"path/filepath"
)

const (
	appDirName    = "TokenFinder"
	vaultFileName = "vault.dat"
)

// DefaultVaultPath returns %APPDATA%\TokenFinder\vault.dat.
func DefaultVaultPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appDirName, vaultFileName), nil
}
