//go:build windows

package system

import (
	"errors"
	"syscall"
)

const processQueryLimitedInformation = 0x1000

func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	handle, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err == nil {
		_ = syscall.CloseHandle(handle)
		return true
	}
	return errors.Is(err, syscall.ERROR_ACCESS_DENIED)
}
