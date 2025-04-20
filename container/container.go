package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/user/minidocker/image"
	"github.com/user/minidocker/pkg/utils"
)

// Container merepresentasikan informasi container
type Container struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Image     string    `json:"image"`
	Status    string    `json:"status"`
	Pid       int       `json:"pid"`
	CreatedAt time.Time `json:"created_at"`
	Volumes   []string  `json:"volumes"`
	Ports     []string  `json:"ports"`
	Memory    string    `json:"memory"`
	CPU       string    `json:"cpu"`
	LogFile   string    `json:"log_file"`
}

const (
	ContainerDir = "/var/run/minidocker/containers"
	StateStopped = "stopped"
	StateRunning = "running"
)

// initContainerDir membuat direktori untuk menyimpan data container
func initContainerDir() error {
	if err := os.MkdirAll(ContainerDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori container: %v", err)
	}
	return nil
}

// RunContainer menjalankan container baru
func RunContainer(imageName, containerName string, volumes []string, ports []string, memory string, cpu string) error {
	if err := initContainerDir(); err != nil {
		return err
	}

	// Buat ID container unik jika nama tidak diberikan
	containerID := containerName
	if containerID == "" {
		containerID = utils.GenerateID(8)
	}

	// Buat direktori root container
	containerRootDir := filepath.Join(ContainerDir, containerID)
	if err := os.MkdirAll(containerRootDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori container: %v", err)
	}

	// Buat file log
	logFile := filepath.Join(containerRootDir, "container.log")
	logFd, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("gagal membuat file log: %v", err)
	}
	logFd.Close()

	// Ekstrak image
	rootfs := filepath.Join(containerRootDir, "rootfs")
	if err := os.MkdirAll(rootfs, 0755); err != nil {
		return fmt.Errorf("gagal membuat rootfs: %v", err)
	}

	// Ekstrak image ke rootfs
	if err := image.ExtractImage(imageName, rootfs); err != nil {
		return fmt.Errorf("gagal ekstrak image: %v", err)
	}

	// Setup port mapping jika diberikan
	if len(ports) > 0 {
		if err := setupPortMapping(ports); err != nil {
			fmt.Printf("Warning: gagal setup port mapping: %v\n", err)
		}
	}

	// Fork child process dengan namespace baru
	cmd := exec.Command("/proc/self/exe", "internal-start", rootfs)

	// Setup namespaces
	// Catatan: Ini hanya bekerja di Linux
	// Di Windows dan OS lain, kode ini diabaikan
	if utils.IsLinux() {
		cmd.SysProcAttr = createLinuxSysProcAttr()
	}
	
	// Redirect output ke file log
	logOutput, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("gagal membuka file log: %v", err)
	}
	defer logOutput.Close()

	cmd.Stdin = os.Stdin
	cmd.Stdout = logOutput
	cmd.Stderr = logOutput

	// Set environment variables
	cmd.Env = append(os.Environ(), 
		fmt.Sprintf("MINIDOCKER_MEMORY=%s", memory),
		fmt.Sprintf("MINIDOCKER_CPU=%s", cpu),
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("gagal menjalankan container: %v", err)
	}

	// Tulis metadata container
	container := Container{
		ID:        containerID,
		Name:      containerName,
		Image:     imageName,
		Status:    StateRunning,
		Pid:       cmd.Process.Pid,
		CreatedAt: time.Now(),
		Volumes:   volumes,
		Ports:     ports,
		Memory:    memory,
		CPU:       cpu,
		LogFile:   logFile,
	}

	containerJSON, err := json.Marshal(container)
	if err != nil {
		return fmt.Errorf("gagal menyimpan metadata container: %v", err)
	}

	if err := ioutil.WriteFile(
		filepath.Join(containerRootDir, "config.json"),
		containerJSON,
		0644,
	); err != nil {
		return fmt.Errorf("gagal menulis config.json: %v", err)
	}

	fmt.Printf("Container %s berhasil dibuat dan dijalankan dengan PID %d\n", containerID, cmd.Process.Pid)
	return nil
}

// RunContainerWithSecurity menjalankan container dengan profil keamanan tertentu
func RunContainerWithSecurity(imageName, containerName string, volumes []string, ports []string, memory string, cpu string, secProfile SecurityProfile) error {
	if err := initContainerDir(); err != nil {
		return err
	}

	// Buat ID container unik jika nama tidak diberikan
	containerID := containerName
	if containerID == "" {
		containerID = utils.GenerateID(8)
	}

	// Buat direktori root container
	containerRootDir := filepath.Join(ContainerDir, containerID)
	if err := os.MkdirAll(containerRootDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori container: %v", err)
	}

	// Buat file log
	logFile := filepath.Join(containerRootDir, "container.log")
	logFd, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("gagal membuat file log: %v", err)
	}
	logFd.Close()

	// Ekstrak image
	rootfs := filepath.Join(containerRootDir, "rootfs")
	if err := os.MkdirAll(rootfs, 0755); err != nil {
		return fmt.Errorf("gagal membuat rootfs: %v", err)
	}

	// Ekstrak image ke rootfs
	if err := image.ExtractImage(imageName, rootfs); err != nil {
		return fmt.Errorf("gagal ekstrak image: %v", err)
	}

	// Setup port mapping jika diberikan
	if len(ports) > 0 {
		if err := setupPortMapping(ports); err != nil {
			fmt.Printf("Warning: gagal setup port mapping: %v\n", err)
		}
	}
	
	// Terapkan profil keamanan
	if err := ApplySecurityProfile(secProfile, containerID); err != nil {
		fmt.Printf("Warning: gagal menerapkan profil keamanan: %v\n", err)
	}

	// Jika read-only filesystem, atur
	if secProfile.ReadOnlyRootfs {
		fmt.Printf("Setting up read-only rootfs untuk container %s\n", containerID)
		// Di implementasi nyata, ini akan menggunakan mount dengan opsi ro
	}

	// Fork child process dengan namespace baru
	cmd := exec.Command("/proc/self/exe", "internal-start", rootfs)

	// Setup namespaces
	// Catatan: Ini hanya bekerja di Linux
	// Di Windows dan OS lain, kode ini diabaikan
	if utils.IsLinux() {
		cmd.SysProcAttr = createLinuxSysProcAttr()
	}
	
	// Redirect output ke file log
	logOutput, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("gagal membuka file log: %v", err)
	}
	defer logOutput.Close()

	cmd.Stdin = os.Stdin
	cmd.Stdout = logOutput
	cmd.Stderr = logOutput

	// Set environment variables
	cmd.Env = append(os.Environ(), 
		fmt.Sprintf("MINIDOCKER_MEMORY=%s", memory),
		fmt.Sprintf("MINIDOCKER_CPU=%s", cpu),
		fmt.Sprintf("MINIDOCKER_SECCOMP=%s", secProfile.SeccompProfile),
		fmt.Sprintf("MINIDOCKER_APPARMOR=%s", secProfile.AppArmorProfile),
		fmt.Sprintf("MINIDOCKER_NO_NEW_PRIVS=%t", secProfile.NoNewPrivs),
		fmt.Sprintf("MINIDOCKER_READONLY=%t", secProfile.ReadOnlyRootfs),
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("gagal menjalankan container: %v", err)
	}

	// Tulis metadata container
	container := Container{
		ID:        containerID,
		Name:      containerName,
		Image:     imageName,
		Status:    StateRunning,
		Pid:       cmd.Process.Pid,
		CreatedAt: time.Now(),
		Volumes:   volumes,
		Ports:     ports,
		Memory:    memory,
		CPU:       cpu,
		LogFile:   logFile,
	}

	containerJSON, err := json.Marshal(container)
	if err != nil {
		return fmt.Errorf("gagal menyimpan metadata container: %v", err)
	}

	if err := ioutil.WriteFile(
		filepath.Join(containerRootDir, "config.json"),
		containerJSON,
		0644,
	); err != nil {
		return fmt.Errorf("gagal menulis config.json: %v", err)
	}

	fmt.Printf("Container %s berhasil dibuat dan dijalankan dengan PID %d\n", containerID, cmd.Process.Pid)
	fmt.Printf("Profil keamanan: %s\n", secProfile.Name)
	return nil
}

// ListContainers menampilkan daftar container
func ListContainers() error {
	if err := initContainerDir(); err != nil {
		return err
	}

	containers, err := getContainers()
	if err != nil {
		return err
	}

	fmt.Printf("%-12s %-15s %-15s %-10s %-10s %-10s %-10s\n", 
		"ID", "NAME", "IMAGE", "STATUS", "PID", "PORTS", "CREATED")
	
	for _, c := range containers {
		// Cek apakah container masih berjalan
		pidRunning := false
		if c.Pid > 0 {
			pidPath := fmt.Sprintf("/proc/%d", c.Pid)
			if _, err := os.Stat(pidPath); err == nil {
				pidRunning = true
			}
		}

		status := c.Status
		if !pidRunning && status == StateRunning {
			status = StateStopped
			updateContainerStatus(c.ID, StateStopped)
		}

		// Format port untuk display
		portDisplay := "none"
		if len(c.Ports) > 0 {
			portDisplay = strings.Join(c.Ports, ", ")
		}

		// Format created time
		createdAgo := time.Since(c.CreatedAt).Round(time.Second)

		fmt.Printf("%-12s %-15s %-15s %-10s %-10d %-10s %s ago\n", 
			c.ID, c.Name, c.Image, status, c.Pid, portDisplay, createdAgo)
	}

	return nil
}

// StopContainer menghentikan container yang sedang berjalan
func StopContainer(containerID string) error {
	if err := initContainerDir(); err != nil {
		return err
	}

	containerDir := filepath.Join(ContainerDir, containerID)
	configPath := filepath.Join(containerDir, "config.json")

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("container dengan ID %s tidak ditemukan", containerID)
	}

	var container Container
	if err := json.Unmarshal(data, &container); err != nil {
		return fmt.Errorf("gagal membaca konfigurasi container: %v", err)
	}

	if container.Status == StateStopped {
		return fmt.Errorf("container %s sudah berhenti", containerID)
	}

	// Kirim sinyal untuk menghentikan proses
	pid := container.Pid
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("proses container tidak ditemukan: %v", err)
	}

	// Kirim sinyal
	// Di Linux akan menggunakan SIGTERM dan SIGKILL
	// Di Windows, hanya Kill yang tersedia
	if err := killProcess(process); err != nil {
		return fmt.Errorf("gagal menghentikan proses: %v", err)
	}

	// Cleanup port mapping jika ada
	if len(container.Ports) > 0 {
		cleanupPortMapping(container.Ports)
	}

	// Update status container
	if err := updateContainerStatus(containerID, StateStopped); err != nil {
		return err
	}

	// Tambahkan log stop
	logFile := container.LogFile
	if logFile != "" {
		logOutput, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			defer logOutput.Close()
			fmt.Fprintf(logOutput, "\n[%s] Container stopped\n", time.Now().Format(time.RFC3339))
		}
	}

	fmt.Printf("Container %s berhasil dihentikan\n", containerID)
	return nil
}

// ContainerLogs menampilkan logs dari container
func ContainerLogs(containerID string, follow bool) error {
	return LogsFromContainer(containerID, follow)
}

// ExecInContainer menjalankan perintah dalam container yang sedang berjalan
func ExecInContainer(containerID string, command []string) error {
	if err := initContainerDir(); err != nil {
		return err
	}

	container, err := getContainer(containerID)
	if err != nil {
		return err
	}

	if container.Status != StateRunning {
		return fmt.Errorf("container %s tidak berjalan", containerID)
	}

	// Pastikan container berjalan dengan melakukan nsenter
	// nsenter memungkinkan untuk masuk ke namespace container
	if utils.IsLinux() {
		// Karena nsenter perlu root privilege, kita perlu sudo
		args := []string{
			"-m", "-u", "-i", "-n", "-p",
			"-t", strconv.Itoa(container.Pid),
			"-r",
		}
		args = append(args, command...)

		cmd := exec.Command("nsenter", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	} else {
		// Di Windows, exec tidak bisa dilakukan dengan benar
		// Kita akan simulasikan dengan pesan
		fmt.Println("Simulasi exec di Windows:")
		fmt.Printf("Menjalankan %v di container %s\n", command, containerID)
		return nil
	}
}

// getContainers membaca semua metadata container
func getContainers() ([]Container, error) {
	files, err := ioutil.ReadDir(ContainerDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Container{}, nil
		}
		return nil, fmt.Errorf("gagal membaca direktori container: %v", err)
	}

	var containers []Container
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		configPath := filepath.Join(ContainerDir, file.Name(), "config.json")
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			continue
		}

		var container Container
		if err := json.Unmarshal(data, &container); err != nil {
			continue
		}

		containers = append(containers, container)
	}

	return containers, nil
}

// getContainer mendapatkan info container by ID
func getContainer(containerID string) (Container, error) {
	containerDir := filepath.Join(ContainerDir, containerID)
	configPath := filepath.Join(containerDir, "config.json")

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Container{}, fmt.Errorf("container dengan ID %s tidak ditemukan", containerID)
	}

	var container Container
	if err := json.Unmarshal(data, &container); err != nil {
		return Container{}, fmt.Errorf("gagal membaca konfigurasi container: %v", err)
	}

	return container, nil
}

// updateContainerStatus mengupdate status container
func updateContainerStatus(containerID, status string) error {
	configPath := filepath.Join(ContainerDir, containerID, "config.json")
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("container dengan ID %s tidak ditemukan", containerID)
	}

	var container Container
	if err := json.Unmarshal(data, &container); err != nil {
		return fmt.Errorf("gagal membaca konfigurasi container: %v", err)
	}

	container.Status = status

	updatedData, err := json.Marshal(container)
	if err != nil {
		return fmt.Errorf("gagal mengupdate metadata container: %v", err)
	}

	if err := ioutil.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("gagal menulis config.json: %v", err)
	}

	return nil
}

// isContainerRunning memeriksa apakah container masih berjalan
func isContainerRunning(containerID string) bool {
	container, err := getContainer(containerID)
	if err != nil {
		return false
	}

	if container.Status != StateRunning {
		return false
	}

	// Verifikasi proses masih berjalan
	pidPath := fmt.Sprintf("/proc/%d", container.Pid)
	if _, err := os.Stat(pidPath); err != nil {
		return false
	}

	return true
}

// setupPortMapping mengatur port mapping menggunakan iptables (Linux only)
func setupPortMapping(ports []string) error {
	// Di non-Linux, kita hanya simulasikan
	if !utils.IsLinux() {
		fmt.Println("Simulasi port mapping:", ports)
		return nil
	}

	// Di Linux, kita akan menggunakan iptables untuk port forwarding
	// Ini memerlukan root privileges
	for _, port := range ports {
		parts := strings.Split(port, ":")
		if len(parts) != 2 {
			return fmt.Errorf("format port tidak valid: %s", port)
		}

		hostPort := parts[0]
		containerPort := parts[1]

		// Contoh menambahkan iptables rule (perlu implementasi tambahan)
		fmt.Printf("Setting up port mapping %s->%s\n", hostPort, containerPort)
	}

	return nil
}

// cleanupPortMapping membersihkan port mapping
func cleanupPortMapping(ports []string) {
	// Di non-Linux, kita hanya simulasikan
	if !utils.IsLinux() {
		fmt.Println("Simulasi cleanup port mapping:", ports)
		return
	}

	// Di Linux, kita akan menghapus iptables rules
	for _, port := range ports {
		parts := strings.Split(port, ":")
		if len(parts) != 2 {
			continue
		}

		hostPort := parts[0]
		containerPort := parts[1]

		// Contoh menghapus iptables rule (perlu implementasi tambahan)
		fmt.Printf("Cleaning up port mapping %s->%s\n", hostPort, containerPort)
	}
} 