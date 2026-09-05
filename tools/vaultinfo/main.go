// Command vaultinfo prints a summary of a TokenFinder vault without
// revealing secret values: entry count, names, kinds and timestamps.
// Useful when checking whether a vault file can be decrypted on this account.
//
//	go run ./tools/vaultinfo            # default vault
//	go run ./tools/vaultinfo C:\x.dat   # specific file
package main

import (
	"fmt"
	"os"

	"tokenfinder/internal/store"
)

func main() {
	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
	} else {
		p, err := store.DefaultVaultPath()
		if err != nil {
			fail(err)
		}
		path = p
	}

	info, err := os.Stat(path)
	if err != nil {
		fail(err)
	}
	fmt.Printf("vault: %s (%d bytes, modified %s)\n", path, info.Size(), info.ModTime().Format("2006-01-02 15:04:05"))

	st, err := store.Open(path)
	if err != nil {
		fail(err)
	}
	entries := st.List()
	fmt.Printf("entries: %d\n", len(entries))
	for _, e := range entries {
		fmt.Printf("  %-32s %-9s updated %s\n", e.Name, e.Kind, e.UpdatedAt.Format("2006-01-02 15:04"))
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "vaultinfo:", err)
	os.Exit(1)
}
