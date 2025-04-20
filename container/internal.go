package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Variabel platform-agnostic untuk implementasi fungsi syscall
var internalSyscallChroot func(path string) error
var internalSetupMounts func(rootfs string) error
var internalSetupCgroups func() error

func init() {
	// Default implementation untuk non-Linux platform
	if runtime.GOOS != "linux" {
		internalSyscallChroot = func(path string) error {
			fmt.Printf("Demo: Chroot ke %s (simulasi)\n", path)
			return nil
		}
		
		internalSetupMounts = func(rootfs string) error {
			setupMountsDemo(rootfs)
			return nil
		}
		
		internalSetupCgroups = func() error {
			fmt.Println("Demo: Setup cgroups (simulasi)")
			return nil
		}
	}
}

// InternalStartContainer memulai proses container dari dalam namespace yang terisolasi
func InternalStartContainer(rootfs string) error {
	// Cek apakah berjalan di Linux
	if runtime.GOOS != "linux" {
		return fmt.Errorf("kontainer hanya bisa berjalan di Linux, bukan di %s", runtime.GOOS)
	}

	fmt.Printf("Memulai container dengan rootfs: %s\n", rootfs)
	
	// Mendapatkan batasan resource
	memLimit := os.Getenv("MINIDOCKER_MEMORY")
	cpuLimit := os.Getenv("MINIDOCKER_CPU")

	// Setup mounts untuk Linux
	if err := internalSetupMounts(rootfs); err != nil {
		return fmt.Errorf("gagal setup mounts: %v", err)
	}
	
	// Setup cgroups hanya di Linux dengan batas yang ditentukan user
	if err := setupCgroupsWithLimits(memLimit, cpuLimit); err != nil {
		fmt.Printf("Warning: gagal setup cgroups: %v\n", err)
	}
	
	// Chroot ke rootfs (hanya di Linux)
	if err := internalSyscallChroot(rootfs); err != nil {
		return fmt.Errorf("gagal chroot: %v", err)
	}
	
	// Change directory ke root
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("gagal chdir ke /: %v", err)
	}
	
	// Untuk demonstrasi, jalankan shell sederhana
	fmt.Println("Menjalankan shell di dalam container...")
	
	// Di implementasi nyata dengan Linux, ini akan menggunakan chroot
	cmdPath := "/bin/sh"
	if runtime.GOOS != "linux" {
		// Di platform lain gunakan perintah yang tersedia
		cmdPath = "sh"
	}
	
	cmd := exec.Command(cmdPath, "-c", "echo 'MiniDocker Container Demo'; /bin/sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gagal menjalankan command di container: %v", err)
	}

	return nil
}

// setupCgroupsWithLimits mengatur cgroups dengan batasan yang ditentukan user
func setupCgroupsWithLimits(memLimit, cpuLimit string) error {
	// Di non-Linux, kita hanya simulasikan
	if runtime.GOOS != "linux" {
		fmt.Printf("Simulasi setup cgroups dengan memory=%s, cpu=%s\n", memLimit, cpuLimit)
		return nil
	}

	// Tentukan pid saat ini
	pid := os.Getpid()

	// Setup cgroup v2
	cgroupRoot := "/sys/fs/cgroup"
	containerCgroup := filepath.Join(cgroupRoot, fmt.Sprintf("minidocker_%d", pid))

	// Membuat cgroup baru
	if err := os.Mkdir(containerCgroup, 0755); err != nil {
		return fmt.Errorf("gagal membuat cgroup: %v", err)
	}

	// Set memory limit
	memBytes, err := parseMemoryLimit(memLimit)
	if err != nil {
		fmt.Printf("Warning: format memory limit tidak valid, menggunakan default: %v\n", err)
		memBytes = 67108864 // 64MB default
	}
	
	memFile := filepath.Join(containerCgroup, "memory.max")
	if err := os.WriteFile(memFile, []byte(strconv.FormatUint(memBytes, 10)), 0644); err != nil {
		return fmt.Errorf("gagal set memory limit: %v", err)
	}

	// Set CPU limit
	cpuValue, err := parseCPULimit(cpuLimit)
	if err != nil {
		fmt.Printf("Warning: format CPU limit tidak valid, menggunakan default: %v\n", err)
		cpuValue = 10000 // 10% dari 100000
	}
	
	cpuFile := filepath.Join(containerCgroup, "cpu.max")
	if err := os.WriteFile(cpuFile, []byte(fmt.Sprintf("%d 100000", cpuValue)), 0644); err != nil {
		return fmt.Errorf("gagal set cpu limit: %v", err)
	}

	// Tambahkan pid ke cgroup
	procsFile := filepath.Join(containerCgroup, "cgroup.procs")
	if err := os.WriteFile(procsFile, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		return fmt.Errorf("gagal tambahkan pid ke cgroup: %v", err)
	}

	fmt.Printf("Berhasil set resource limits: memory=%s (%d bytes), cpu=%s (%d)\n", 
		memLimit, memBytes, cpuLimit, cpuValue)

	return nil
}

// parseMemoryLimit mengubah string memori seperti "64m" menjadi bytes
func parseMemoryLimit(limit string) (uint64, error) {
	if limit == "" {
		return 67108864, nil // 64MB default
	}

	limit = strings.ToLower(limit)
	var multiplier uint64 = 1

	if strings.HasSuffix(limit, "k") {
		multiplier = 1024
		limit = limit[:len(limit)-1]
	} else if strings.HasSuffix(limit, "m") {
		multiplier = 1024 * 1024
		limit = limit[:len(limit)-1]
	} else if strings.HasSuffix(limit, "g") {
		multiplier = 1024 * 1024 * 1024
		limit = limit[:len(limit)-1]
	}

	value, err := strconv.ParseUint(limit, 10, 64)
	if err != nil {
		return 0, err
	}

	return value * multiplier, nil
}

// parseCPULimit mengubah persentase CPU menjadi nilai untuk cgroups
func parseCPULimit(limit string) (uint64, error) {
	if limit == "" {
		return 10000, nil // 10% default
	}

	// Hapus "%" jika ada
	limit = strings.TrimSuffix(limit, "%")
	
	value, err := strconv.ParseUint(limit, 10, 64)
	if err != nil {
		return 0, err
	}

	// Batasi nilai antara 0-100%
	if value > 100 {
		value = 100
	}

	// Konversi persentase ke jumlah dari 100000
	// 100% = 100000
	cpuValue := (value * 100000) / 100

	return cpuValue, nil
}

// syscallChroot adalah fungsi pembungkus untuk chroot agar kode tetap berjalan di Windows
func syscallChroot(path string) error {
	if runtime.GOOS != "linux" {
		// Di Windows ini tidak tersedia, jadi kita simulasikan
		fmt.Printf("Demo: Chroot ke %s\n", path)
		return nil
	}
	// Di Linux, panggil syscall sebenarnya
	// Ini akan error di Windows saat build, jadi dikemas dalam fungsi terpisah
	return nil // Diganti dengan syscall.Chroot di platform Linux
}

// setupMountsDemo (hanya contoh untuk dokumentasi)
func setupMountsDemo(rootfs string) {
	fmt.Printf("Demo: Mount /proc ke %s/proc\n", rootfs)
	fmt.Printf("Demo: Mount /sys ke %s/sys\n", rootfs)
	fmt.Printf("Demo: Mount tmpfs ke %s/tmp\n", rootfs)
}

// pivotRootDemo (hanya contoh untuk dokumentasi)
func pivotRootDemo(rootfs string) {
	pivotDir := filepath.Join(rootfs, ".pivot_root")
	fmt.Printf("Demo: Membuat pivot directory di %s\n", pivotDir)
	fmt.Printf("Demo: Pindahkan root filesystem ke %s\n", rootfs)
	fmt.Printf("Demo: Unmount old root dari %s\n", pivotDir)
} 