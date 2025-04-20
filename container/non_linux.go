//go:build !linux
// +build !linux

package container

import (
	"syscall"
)

func init() {
	// Mengaitkan fungsi internal ke variabel global
	createLinuxSysProcAttrInternal = createLinuxSysProcAttrImpl
}

// Dummy implementation untuk platform non-Linux
func createLinuxSysProcAttrImpl() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
} 