package container

import (
	"os"
	"runtime"
	"syscall"
	"time"
)

// Deklarasi fungsi yang diimplementasikan di linux.go dan non_linux.go
// dengan build tags
var createLinuxSysProcAttrInternal func() *syscall.SysProcAttr

// createLinuxSysProcAttr membuat SysProcAttr dengan namespaces untuk Linux
// Fungsi ini hanya dipanggil di Linux
func createLinuxSysProcAttr() *syscall.SysProcAttr {
	// Untuk Windows atau platform lain, berikan nilai default kosong
	if runtime.GOOS != "linux" || createLinuxSysProcAttrInternal == nil {
		return &syscall.SysProcAttr{}
	}
	
	// Panggil implementasi platform spesifik
	return createLinuxSysProcAttrInternal()
}

// killProcess membunuh proses dengan cara yang sesuai platform
func killProcess(process *os.Process) error {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// Di Unix-like OS, kita kirim SIGTERM lalu SIGKILL
		if err := process.Signal(syscall.SIGTERM); err != nil {
			return err
		}

		// Tunggu beberapa saat lalu kirim SIGKILL jika masih berjalan
		time.Sleep(2 * time.Second)
		if err := process.Signal(syscall.SIGKILL); err != nil {
			// Abaikan error ini karena proses mungkin sudah berhenti
			return nil
		}
		return nil
	} else {
		// Di Windows, kita hanya bisa menggunakan Kill()
		return process.Kill()
	}
} 