package image

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	ImageDir = "/var/run/minidocker/images"
)

// ImageConfig merepresentasikan konfigurasi image
type ImageConfig struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
	Cmd     []string `json:"cmd"`
	Env     []string `json:"env"`
}

// InitImageDir membuat direktori untuk menyimpan image
func InitImageDir() error {
	if err := os.MkdirAll(ImageDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori image: %v", err)
	}
	return nil
}

// ExtractImage mengekstrak image tar.gz ke directory rootfs
func ExtractImage(imageName, targetDir string) error {
	// Inisialisasi direktori image jika belum ada
	if err := InitImageDir(); err != nil {
		return err
	}

	// Cek apakah image sudah tersedia dalam cache
	imagePath := filepath.Join(ImageDir, imageName+".tar.gz")
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		// Download image jika belum tersedia
		if err := downloadImage(imageName, imagePath); err != nil {
			return fmt.Errorf("gagal download image %s: %v", imageName, err)
		}
	}

	// Buka file tar.gz
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("gagal membuka image file: %v", err)
	}
	defer file.Close()

	// Dekompresi gzip
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("gagal membaca gzip: %v", err)
	}
	defer gzr.Close()

	// Ekstrak tar
	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("gagal membaca tar: %v", err)
		}

		// Normalisasi path
		path := filepath.Join(targetDir, header.Name)
		
		// Cegah path traversal
		if !strings.HasPrefix(path, targetDir) {
			return fmt.Errorf("path traversal terdeteksi: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Buat direktori
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("gagal membuat direktori %s: %v", path, err)
			}
		case tar.TypeReg:
			// Buat direktori parent
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Errorf("gagal membuat parent dir untuk %s: %v", path, err)
			}
			// Buat file
			outFile, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("gagal membuat file %s: %v", path, err)
			}
			defer outFile.Close()
			
			// Set permissions
			if err := os.Chmod(path, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("gagal set permission %s: %v", path, err)
			}
			
			// Salin konten
			if _, err := io.Copy(outFile, tr); err != nil {
				return fmt.Errorf("gagal menyalin konten ke %s: %v", path, err)
			}
		case tar.TypeSymlink:
			// Buat simlink
			if err := os.Symlink(header.Linkname, path); err != nil {
				return fmt.Errorf("gagal membuat symlink %s -> %s: %v", path, header.Linkname, err)
			}
		default:
			fmt.Printf("Mengabaikan %s (tipe=%d)\n", header.Name, header.Typeflag)
		}
	}

	// Ekstrak dan tulis konfigurasi image
	if err := extractImageConfig(imageName, targetDir); err != nil {
		fmt.Printf("Warning: gagal ekstrak konfigurasi image: %v\n", err)
	}

	return nil
}

// downloadImage mendownload image dari registry
func downloadImage(imageName, targetPath string) error {
	// Jika image berupa busybox, alpine, atau ubuntu - gunakan image dasar yang sudah disediakan
	switch imageName {
	case "alpine":
		// Untuk demo, kita buat image Alpine dasar
		if err := createBasicAlpineImage(targetPath); err != nil {
			return err
		}
	case "busybox":
		// Untuk demo, kita buat image Busybox dasar
		if err := createBasicBusyboxImage(targetPath); err != nil {
			return err
		}
	default:
		return fmt.Errorf("image %s tidak didukung. Hanya alpine dan busybox yang didukung untuk demo", imageName)
	}

	return nil
}

// createBasicAlpineImage membuat image Alpine sederhana
func createBasicAlpineImage(targetPath string) error {
	// Untuk DEMO saja - pada implementasi nyata perlu download dari registry
	fmt.Println("Image Alpine tidak ditemukan. Dalam implementasi lengkap, ini akan diunduh dari registry")
	fmt.Println("Untuk demo, kita akan membuat placeholder image...")
	
	// Demo: Buat file tar.gz kosong dengan beberapa file demo
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}
	
	// Buat file tar.gz
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Setup gzip + tar writer
	gzw := gzip.NewWriter(file)
	defer gzw.Close()
	tw := tar.NewWriter(gzw)
	defer tw.Close()
	
	// Tambah file bin/sh
	binDir := "bin"
	if err := addTarDir(tw, binDir); err != nil {
		return err
	}
	
	// Tambah file shell
	shellContent := "#!/bin/sh\necho 'MiniDocker Alpine Demo Shell'\n/bin/sh\n"
	if err := addTarFile(tw, "bin/sh", shellContent, 0755); err != nil {
		return err
	}
	
	// Tambah file /etc/os-release
	etcDir := "etc"
	if err := addTarDir(tw, etcDir); err != nil {
		return err
	}
	
	osReleaseContent := "NAME=\"Alpine Linux\"\nID=alpine\nVERSION_ID=3.16.0\nPRETTY_NAME=\"MiniDocker Alpine Demo\"\n"
	if err := addTarFile(tw, "etc/os-release", osReleaseContent, 0644); err != nil {
		return err
	}
	
	// Tambah direktori yang diperlukan
	for _, dir := range []string{"proc", "sys", "dev", "tmp", "usr", "var"} {
		if err := addTarDir(tw, dir); err != nil {
			return err
		}
	}
	
	// Tulis konfigurasi image
	config := ImageConfig{
		Name:    "alpine",
		Version: "3.16.0",
		Cmd:     []string{"/bin/sh"},
		Env:     []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
	}
	
	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}
	
	if err := addTarFile(tw, "image-config.json", string(configJSON), 0644); err != nil {
		return err
	}
	
	return nil
}

// createBasicBusyboxImage membuat image Busybox sederhana
func createBasicBusyboxImage(targetPath string) error {
	// Sama seperti Alpine, untuk demo saja
	fmt.Println("Image Busybox tidak ditemukan. Dalam implementasi lengkap, ini akan diunduh dari registry")
	fmt.Println("Untuk demo, kita akan membuat placeholder image...")
	
	// Demo: Buat file tar.gz kosong dengan beberapa file demo
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}
	
	// Buat file tar.gz
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Setup gzip + tar writer
	gzw := gzip.NewWriter(file)
	defer gzw.Close()
	tw := tar.NewWriter(gzw)
	defer tw.Close()
	
	// Tambah file bin/sh
	binDir := "bin"
	if err := addTarDir(tw, binDir); err != nil {
		return err
	}
	
	// Tambah file shell
	shellContent := "#!/bin/sh\necho 'MiniDocker Busybox Demo Shell'\n/bin/sh\n"
	if err := addTarFile(tw, "bin/sh", shellContent, 0755); err != nil {
		return err
	}
	
	// Tambah file /etc/os-release
	etcDir := "etc"
	if err := addTarDir(tw, etcDir); err != nil {
		return err
	}
	
	osReleaseContent := "NAME=\"Busybox\"\nID=busybox\nVERSION_ID=1.35.0\nPRETTY_NAME=\"MiniDocker Busybox Demo\"\n"
	if err := addTarFile(tw, "etc/os-release", osReleaseContent, 0644); err != nil {
		return err
	}
	
	// Tambah direktori yang diperlukan
	for _, dir := range []string{"proc", "sys", "dev", "tmp", "usr", "var"} {
		if err := addTarDir(tw, dir); err != nil {
			return err
		}
	}
	
	// Tulis konfigurasi image
	config := ImageConfig{
		Name:    "busybox",
		Version: "1.35.0",
		Cmd:     []string{"/bin/sh"},
		Env:     []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
	}
	
	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}
	
	if err := addTarFile(tw, "image-config.json", string(configJSON), 0644); err != nil {
		return err
	}
	
	return nil
}

// addTarDir menambahkan direktori ke archive tar
func addTarDir(tw *tar.Writer, dirName string) error {
	header := &tar.Header{
		Name:     dirName,
		Mode:     0755,
		Typeflag: tar.TypeDir,
	}
	
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("gagal menulis header direktori: %v", err)
	}
	
	return nil
}

// addTarFile menambahkan file ke archive tar
func addTarFile(tw *tar.Writer, name, content string, mode int64) error {
	header := &tar.Header{
		Name:     name,
		Mode:     mode,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	}
	
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("gagal menulis header file: %v", err)
	}
	
	if _, err := tw.Write([]byte(content)); err != nil {
		return fmt.Errorf("gagal menulis konten file: %v", err)
	}
	
	return nil
}

// extractImageConfig mengekstrak konfigurasi image dari rootfs
func extractImageConfig(imageName, rootfs string) error {
	// Baca konfigurasi dari image-config.json
	configPath := filepath.Join(rootfs, "image-config.json")
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		// Untuk demo, jika tidak ada file config, buat konfigurasi default
		config := ImageConfig{
			Name:    imageName,
			Version: "latest",
			Cmd:     []string{"/bin/sh"},
			Env:     []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		}
		
		data, err = json.Marshal(config)
		if err != nil {
			return err
		}
	}
	
	// Tulis konfigurasi ke file
	configTargetPath := filepath.Join(rootfs, "../config.json")
	if err := ioutil.WriteFile(configTargetPath, data, 0644); err != nil {
		return fmt.Errorf("gagal menulis konfigurasi image: %v", err)
	}
	
	return nil
} 