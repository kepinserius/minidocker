package container

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LogsFromContainer membaca dan menampilkan log dari container dengan ID tertentu
// Parameter follow menentukan apakah logs akan di-follow secara real-time
// Fungsi ini berbeda dengan ContainerLogs di container.go untuk menghindari redeclaration
func LogsFromContainer(id string, follow bool) error {
	// Verifikasi container ada
	containerPath := filepath.Join(ContainerDir, id)
	if _, err := os.Stat(containerPath); os.IsNotExist(err) {
		return fmt.Errorf("container dengan ID '%s' tidak ditemukan", id)
	}

	// Baca config untuk memastikan container valid
	configPath := filepath.Join(containerPath, "config.json")
	_, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("gagal membaca config container: %v", err)
	}

	// Path ke file log container
	logPath := filepath.Join(containerPath, "container.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return fmt.Errorf("file log untuk container '%s' tidak ditemukan", id)
	}

	// Buka file log
	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("gagal membuka file log: %v", err)
	}
	defer file.Close()

	// Jika tidak follow, cukup tampilkan log yang ada
	if !follow {
		_, err = io.Copy(os.Stdout, file)
		return err
	}

	// Jika follow, gunakan scanner untuk membaca log line-by-line
	// dan tetap pantau file untuk perubahan baru
	fmt.Printf("Menampilkan logs untuk container %s (CTRL+C untuk keluar):\n", id)
	
	scanner := bufio.NewScanner(file)
	
	// Pertama, tampilkan log yang sudah ada
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	// Kemudian, pantau file untuk perubahan baru
	for {
		// Dapatkan posisi saat ini di file
		currentPos, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			return fmt.Errorf("error saat membaca posisi file: %v", err)
		}

		// Tunggu sebentar untuk perubahan file
		time.Sleep(100 * time.Millisecond)

		// Dapatkan ukuran file terbaru
		fileInfo, err := os.Stat(logPath)
		if err != nil {
			return fmt.Errorf("error saat membaca info file: %v", err)
		}

		// Jika file bertambah ukurannya
		if fileInfo.Size() > currentPos {
			// Reset scanner dengan file yang sama
			scanner = bufio.NewScanner(file)
			
			// Lanjutkan membaca dari posisi terakhir
			_, err = file.Seek(currentPos, io.SeekStart)
			if err != nil {
				return fmt.Errorf("error saat menyetel posisi file: %v", err)
			}

			// Baca dan tampilkan baris baru
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
		}

		// Verifikasi container masih ada
		if _, err := os.Stat(containerPath); os.IsNotExist(err) {
			return fmt.Errorf("container '%s' dihentikan", id)
		}
	}
} 