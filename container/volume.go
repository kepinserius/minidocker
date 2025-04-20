package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Volume merepresentasikan informasi volume
type Volume struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Driver    string    `json:"driver"`
	Mountpoint string   `json:"mountpoint"`
	CreatedAt time.Time `json:"created_at"`
	Labels    map[string]string `json:"labels"`
}

const (
	VolumeDir = "/var/run/minidocker/volumes"
)

// InitVolumeDir membuat direktori untuk menyimpan data volume
func InitVolumeDir() error {
	if err := os.MkdirAll(VolumeDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori volume: %v", err)
	}
	return nil
}

// CreateVolume membuat volume baru
func CreateVolume(name string, labels map[string]string) (*Volume, error) {
	if err := InitVolumeDir(); err != nil {
		return nil, err
	}

	// Periksa apakah volume dengan nama ini sudah ada
	volumes, err := ListVolumes()
	if err != nil {
		return nil, err
	}

	for _, v := range volumes {
		if v.Name == name {
			return &v, fmt.Errorf("volume dengan nama '%s' sudah ada", name)
		}
	}

	// Buat ID unik
	volumeID := name
	if volumeID == "" {
		// Gunakan timestamp sebagai ID default
		volumeID = fmt.Sprintf("vol_%d", time.Now().UnixNano())
	}

	// Buat direktori volume
	volumePath := filepath.Join(VolumeDir, volumeID)
	if err := os.MkdirAll(volumePath, 0755); err != nil {
		return nil, fmt.Errorf("gagal membuat direktori volume: %v", err)
	}

	// Buat mountpoint
	mountPoint := filepath.Join(volumePath, "data")
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return nil, fmt.Errorf("gagal membuat mountpoint: %v", err)
	}

	// Buat volume
	volume := Volume{
		ID:         volumeID,
		Name:       name,
		Driver:     "local", // Gunakan driver local sebagai default
		Mountpoint: mountPoint,
		CreatedAt:  time.Now(),
		Labels:     labels,
	}

	// Simpan metadata volume
	volumeJSON, err := json.Marshal(volume)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan metadata volume: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(volumePath, "config.json"),
		volumeJSON,
		0644,
	); err != nil {
		return nil, fmt.Errorf("gagal menulis config.json: %v", err)
	}

	fmt.Printf("Volume %s berhasil dibuat di %s\n", volume.Name, volume.Mountpoint)
	return &volume, nil
}

// ListVolumes menampilkan daftar volume
func ListVolumes() ([]Volume, error) {
	if err := InitVolumeDir(); err != nil {
		return nil, err
	}

	files, err := os.ReadDir(VolumeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Volume{}, nil
		}
		return nil, fmt.Errorf("gagal membaca direktori volume: %v", err)
	}

	var volumes []Volume
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		configPath := filepath.Join(VolumeDir, file.Name(), "config.json")
		data, err := os.ReadFile(configPath)
		if err != nil {
			continue
		}

		var volume Volume
		if err := json.Unmarshal(data, &volume); err != nil {
			continue
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
}

// GetVolume mendapatkan volume berdasarkan nama atau ID
func GetVolume(nameOrID string) (*Volume, error) {
	volumes, err := ListVolumes()
	if err != nil {
		return nil, err
	}

	for _, v := range volumes {
		if v.ID == nameOrID || v.Name == nameOrID {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("volume '%s' tidak ditemukan", nameOrID)
}

// RemoveVolume menghapus volume
func RemoveVolume(nameOrID string, force bool) error {
	volume, err := GetVolume(nameOrID)
	if err != nil {
		return err
	}

	// Periksa apakah volume sedang digunakan
	if !force {
		// Cek apakah volume masih digunakan oleh container
		containers, err := getContainers()
		if err != nil {
			return err
		}

		for _, c := range containers {
			if c.Status == StateRunning {
				for _, v := range c.Volumes {
					if volumeInUse(v, volume.Name) || volumeInUse(v, volume.ID) {
						return fmt.Errorf("volume '%s' masih digunakan oleh container '%s'", volume.Name, c.ID)
					}
				}
			}
		}
	}

	// Hapus direktori volume
	volumePath := filepath.Join(VolumeDir, volume.ID)
	if err := os.RemoveAll(volumePath); err != nil {
		return fmt.Errorf("gagal menghapus volume: %v", err)
	}

	fmt.Printf("Volume %s berhasil dihapus\n", volume.Name)
	return nil
}

// volumeInUse memeriksa apakah volume digunakan dalam spesifikasi volume
func volumeInUse(volumeSpec string, volumeName string) bool {
	// Format volume: nama_volume:target_path atau nama_volume:target_path:mode
	parts := filepath.SplitList(volumeSpec)
	if len(parts) > 0 {
		return parts[0] == volumeName
	}
	return false
}

// MountVolume memasang volume pada container
func MountVolume(volumeSpec string, containerRootfs string) error {
	parts := filepath.SplitList(volumeSpec)
	if len(parts) < 2 {
		return fmt.Errorf("format volume tidak valid, harus: volume_name:container_path[:mode]")
	}

	volumeName := parts[0]
	targetPath := filepath.Join(containerRootfs, parts[1])

	// Dapatkan volume
	volume, err := GetVolume(volumeName)
	if err != nil {
		return err
	}

	// Pastikan direktori target ada
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori target: %v", err)
	}

	// Buat link ke volume
	// Pada sistem nyata, ini akan menggunakan mount point
	fmt.Printf("Memasang volume %s ke %s\n", volume.Name, targetPath)

	return nil
}

// BackupVolumeData melakukan backup data volume ke file tar
func BackupVolumeData(volumeName, backupPath string) error {
	volume, err := GetVolume(volumeName)
	if err != nil {
		return err
	}

	// Pastikan direktori backup ada
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori backup: %v", err)
	}

	// Simulasi proses backup
	fmt.Printf("Membuat backup volume %s ke %s\n", volume.Name, backupPath)
	
	// Di implementasi nyata, ini akan menggunakan tar untuk mengompresi data
	// cmd := exec.Command("tar", "-czf", backupPath, "-C", volume.Mountpoint, ".")
	// return cmd.Run()

	// Simulasi saja
	backupFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("gagal membuat file backup: %v", err)
	}
	defer backupFile.Close()

	fmt.Fprintf(backupFile, "Backup dari volume %s pada %s\n", volume.Name, time.Now().Format(time.RFC3339))
	
	return nil
}

// RestoreVolumeData memulihkan data volume dari file tar
func RestoreVolumeData(volumeName, backupPath string) error {
	volume, err := GetVolume(volumeName)
	if err != nil {
		return err
	}

	// Periksa apakah file backup ada
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("file backup tidak ditemukan: %s", backupPath)
	}

	// Simulasi proses restore
	fmt.Printf("Memulihkan backup ke volume %s dari %s\n", volume.Name, backupPath)
	
	// Di implementasi nyata, ini akan menggunakan tar untuk mengekstrak data
	// cmd := exec.Command("tar", "-xzf", backupPath, "-C", volume.Mountpoint)
	// return cmd.Run()

	return nil
} 