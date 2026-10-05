//go:build windows

package mcpseat

import (
	"errors"
	"syscall"
)

// lockFile holds path open with no sharing, which makes a second open
// fail with a sharing violation until the first handle closes. Windows
// closes the handle when the process dies.
func lockFile(path string) (func(), error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, err := syscall.CreateFile(p, syscall.GENERIC_READ|syscall.GENERIC_WRITE, 0, nil,
		syscall.OPEN_ALWAYS, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		if errors.Is(err, syscall.Errno(32)) { // ERROR_SHARING_VIOLATION
			return nil, errTableHeld
		}
		return nil, err
	}
	return func() { _ = syscall.CloseHandle(h) }, nil
}
