// Package crypto wraps the Windows Data Protection API (DPAPI).
// Data protected here can only be decrypted by the same Windows user account
// on the same machine, which is what we want for a local secret vault.
package crypto

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// entropy is mixed into the DPAPI key so other DPAPI-aware programs running
// as the same user cannot trivially decrypt the vault without knowing it.
var entropy = []byte("TokenFinder.vault.v1")

const description = "TokenFinder vault"

func toBlob(b []byte) windows.DataBlob {
	if len(b) == 0 {
		return windows.DataBlob{}
	}
	return windows.DataBlob{Size: uint32(len(b)), Data: &b[0]}
}

// copyAndFree copies the DPAPI output buffer into Go memory and releases it.
func copyAndFree(out *windows.DataBlob) []byte {
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	res := make([]byte, out.Size)
	copy(res, unsafe.Slice(out.Data, out.Size))
	return res
}

// Protect encrypts data for the current Windows user.
func Protect(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("dpapi: nothing to protect")
	}
	in := toBlob(data)
	ent := toBlob(entropy)
	desc, err := windows.UTF16PtrFromString(description)
	if err != nil {
		return nil, err
	}
	var out windows.DataBlob
	if err := windows.CryptProtectData(&in, desc, &ent, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return nil, err
	}
	return copyAndFree(&out), nil
}

// Unprotect decrypts data previously produced by Protect.
func Unprotect(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("dpapi: nothing to unprotect")
	}
	in := toBlob(data)
	ent := toBlob(entropy)
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(&in, nil, &ent, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return nil, err
	}
	return copyAndFree(&out), nil
}
