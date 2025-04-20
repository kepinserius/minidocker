package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// GenerateID menghasilkan ID acak dengan panjang tertentu
func GenerateID(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", os.Getpid())
	}
	return hex.EncodeToString(bytes)
}

// Exists memeriksa apakah path ada
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// ExecuteCommand menjalankan perintah shell
func ExecuteCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gagal menjalankan perintah: %v, output: %s", err, output)
	}
	return string(output), nil
}

// CreateDirectoryIfNotExist membuat direktori jika belum ada
func CreateDirectoryIfNotExist(path string) error {
	if Exists(path) {
		return nil
	}
	return os.MkdirAll(path, 0755)
}

// IsLinux memeriksa apakah berjalan di sistem Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
} 