package container

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ImageInfo merepresentasikan informasi image di registry
type ImageInfo struct {
	Name      string            `json:"name"`
	Tag       string            `json:"tag"`
	Size      int64             `json:"size"`
	Digest    string            `json:"digest"`
	CreatedAt time.Time         `json:"created_at"`
	Labels    map[string]string `json:"labels"`
}

const (
	// RegistryDir direktori untuk menyimpan data registry
	RegistryDir = "/var/run/minidocker/registry"
)

// InitRegistryDir membuat direktori untuk menyimpan data registry
func InitRegistryDir() error {
	if err := os.MkdirAll(RegistryDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori registry: %v", err)
	}
	return nil
}

// PullImage mengunduh image dari registry (simulasi)
func PullImage(imageName string) error {
	// Inisialisasi direktori registry
	if err := InitRegistryDir(); err != nil {
		return err
	}

	// Parsing nama image dan tag
	name, tag := parseImageNameTag(imageName)
	if tag == "" {
		tag = "latest"
	}

	fmt.Printf("Mengunduh image %s:%s...\n", name, tag)

	// Simulasi penunduhan dari registry
	// Pada implementasi nyata, ini akan mengunduh dari Docker Hub atau registry lain
	imageURL := fmt.Sprintf("https://example.com/v2/%s/manifests/%s", name, tag)
	fmt.Printf("Simulasi GET %s\n", imageURL)

	// Buat direktori untuk image
	imageDir := filepath.Join(RegistryDir, name)
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori image: %v", err)
	}

	// Buat metadata image (simulasi)
	imageInfo := ImageInfo{
		Name:      name,
		Tag:       tag,
		Size:      12345678, // ukuran simulasi
		Digest:    fmt.Sprintf("sha256:%d", time.Now().UnixNano()), // digest simulasi
		CreatedAt: time.Now(),
		Labels:    map[string]string{"maintainer": "minidocker"},
	}

	// Simpan metadata
	imageJSON, err := json.MarshalIndent(imageInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("gagal menyimpan metadata image: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(imageDir, fmt.Sprintf("%s.json", tag)),
		imageJSON,
		0644,
	); err != nil {
		return fmt.Errorf("gagal menulis metadata: %v", err)
	}

	// Simulasi unduhan layer
	fmt.Println("Menyiapkan layer...")
	time.Sleep(1 * time.Second)
	fmt.Println("Layer 1/3: [====================] 100%")
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Layer 2/3: [====================] 100%")
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Layer 3/3: [====================] 100%")

	fmt.Printf("Image %s:%s berhasil diunduh\n", name, tag)
	return nil
}

// PushImage mengunggah image ke registry (simulasi)
func PushImage(imageName string) error {
	// Parsing nama image dan tag
	name, tag := parseImageNameTag(imageName)
	if tag == "" {
		tag = "latest"
	}

	fmt.Printf("Mengunggah image %s:%s...\n", name, tag)

	// Simulasi pengunggahan ke registry
	fmt.Println("Menyiapkan layer...")
	time.Sleep(1 * time.Second)
	fmt.Println("Layer 1/3: [====================] 100%")
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Layer 2/3: [====================] 100%")
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Layer 3/3: [====================] 100%")

	fmt.Printf("Image %s:%s berhasil diunggah\n", name, tag)
	return nil
}

// ListImages mendapatkan daftar image di registry lokal
func ListImages() ([]ImageInfo, error) {
	// Inisialisasi direktori registry
	if err := InitRegistryDir(); err != nil {
		return nil, err
	}

	dirs, err := os.ReadDir(RegistryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ImageInfo{}, nil
		}
		return nil, fmt.Errorf("gagal membaca direktori registry: %v", err)
	}

	var images []ImageInfo
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		// Baca semua file metadata tag
		imageDir := filepath.Join(RegistryDir, dir.Name())
		files, err := os.ReadDir(imageDir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
				// Baca file metadata
				data, err := os.ReadFile(filepath.Join(imageDir, file.Name()))
				if err != nil {
					continue
				}

				var imageInfo ImageInfo
				if err := json.Unmarshal(data, &imageInfo); err != nil {
					continue
				}

				images = append(images, imageInfo)
			}
		}
	}

	return images, nil
}

// StartLocalRegistry memulai registry HTTP lokal sederhana
func StartLocalRegistry(port int) error {
	// Inisialisasi direktori registry
	if err := InitRegistryDir(); err != nil {
		return err
	}

	// Handler untuk menyajikan file dari direktori registry
	http.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Registry request: %s %s\n", r.Method, r.URL.Path)

		// Implementasi dasar v2 API untuk simulasi
		if r.URL.Path == "/v2/" {
			// API version check
			w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Parse path untuk mendapatkan nama image dan tag
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v2/"), "/")
		if len(parts) < 2 {
			http.Error(w, "Invalid repository name", http.StatusBadRequest)
			return
		}

		// Simple catalog API
		if parts[0] == "_catalog" {
			images, err := ListImages()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var repos []string
			for _, img := range images {
				found := false
				for _, r := range repos {
					if r == img.Name {
						found = true
						break
					}
				}
				if !found {
					repos = append(repos, img.Name)
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"repositories": repos,
			})
			return
		}

		// Handle tags/list
		if len(parts) >= 3 && parts[len(parts)-2] == "tags" && parts[len(parts)-1] == "list" {
			repoName := strings.Join(parts[:len(parts)-2], "/")
			
			// Baca semua tag untuk repositori ini
			images, err := ListImages()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var tags []string
			for _, img := range images {
				if img.Name == repoName {
					tags = append(tags, img.Tag)
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name": repoName,
				"tags": tags,
			})
			return
		}

		// Tidak ditemukan
		http.NotFound(w, r)
	})

	// Mulai server HTTP
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Memulai registry lokal di http://localhost%s\n", addr)
	fmt.Println("Tekan Ctrl+C untuk menghentikan")

	return http.ListenAndServe(addr, nil)
}

// TagImage membuat tag baru untuk image
func TagImage(sourceImage, targetImage string) error {
	// Parsing nama image dan tag
	sourceName, sourceTag := parseImageNameTag(sourceImage)
	if sourceTag == "" {
		sourceTag = "latest"
	}

	targetName, targetTag := parseImageNameTag(targetImage)
	if targetTag == "" {
		targetTag = "latest"
	}

	fmt.Printf("Membuat tag %s:%s dari %s:%s\n", targetName, targetTag, sourceName, sourceTag)

	// Periksa apakah source image ada
	sourceMetadataPath := filepath.Join(RegistryDir, sourceName, fmt.Sprintf("%s.json", sourceTag))
	if _, err := os.Stat(sourceMetadataPath); os.IsNotExist(err) {
		return fmt.Errorf("source image tidak ditemukan: %s:%s", sourceName, sourceTag)
	}

	// Baca metadata source
	data, err := os.ReadFile(sourceMetadataPath)
	if err != nil {
		return fmt.Errorf("gagal membaca metadata: %v", err)
	}

	var imageInfo ImageInfo
	if err := json.Unmarshal(data, &imageInfo); err != nil {
		return fmt.Errorf("gagal membaca metadata: %v", err)
	}

	// Update metadata untuk target
	imageInfo.Name = targetName
	imageInfo.Tag = targetTag

	// Buat direktori untuk target image jika belum ada
	targetDir := filepath.Join(RegistryDir, targetName)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori target: %v", err)
	}

	// Simpan metadata baru
	targetMetadata, err := json.MarshalIndent(imageInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("gagal menyimpan metadata: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(targetDir, fmt.Sprintf("%s.json", targetTag)),
		targetMetadata,
		0644,
	); err != nil {
		return fmt.Errorf("gagal menulis metadata: %v", err)
	}

	fmt.Printf("Tag %s:%s berhasil dibuat\n", targetName, targetTag)
	return nil
}

// DownloadImageFromURL mengunduh image dari URL (simulasi)
func DownloadImageFromURL(url, savePath string) error {
	fmt.Printf("Mengunduh dari %s ke %s...\n", url, savePath)

	// Simulasi unduhan
	time.Sleep(2 * time.Second)

	// Buat direktori jika belum ada
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori: %v", err)
	}

	// Buat file kosong untuk simulasi
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("gagal membuat file: %v", err)
	}
	defer file.Close()

	// Tulis beberapa data simulasi
	_, err = file.WriteString(fmt.Sprintf("Simulasi file image dari %s\n", url))
	if err != nil {
		return fmt.Errorf("gagal menulis file: %v", err)
	}

	fmt.Println("Unduhan selesai")
	return nil
}

// parseImageNameTag memecah string image menjadi nama dan tag
func parseImageNameTag(image string) (name, tag string) {
	parts := strings.SplitN(image, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return image, ""
} 