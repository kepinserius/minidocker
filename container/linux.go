//go:build linux
// +build linux

package container

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func init() {
	// Mengaitkan fungsi internal ke variabel global
	createLinuxSysProcAttrInternal = createLinuxSysProcAttrImpl
	
	// Inisialisasi fungsi-fungsi syscall Linux
	internalSyscallChroot = syscall.Chroot
	internalSetupMounts = setupMountsLinux
	internalSetupCgroups = setupCgroupsLinux
}

// Implementasi khusus Linux dari setupMounts
func setupMountsLinux(rootfs string) error {
	// Mount procfs
	procPath := filepath.Join(rootfs, "proc")
	if err := os.MkdirAll(procPath, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori /proc: %v", err)
	}

	if err := syscall.Mount("proc", procPath, "proc", 0, ""); err != nil {
		return fmt.Errorf("gagal mount /proc: %v", err)
	}

	// Mount sysfs
	sysPath := filepath.Join(rootfs, "sys")
	if err := os.MkdirAll(sysPath, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori /sys: %v", err)
	}

	if err := syscall.Mount("sysfs", sysPath, "sysfs", 0, ""); err != nil {
		return fmt.Errorf("gagal mount /sys: %v", err)
	}

	// Mount tmpfs
	tmpPath := filepath.Join(rootfs, "tmp")
	if err := os.MkdirAll(tmpPath, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori /tmp: %v", err)
	}

	if err := syscall.Mount("tmpfs", tmpPath, "tmpfs", 0, ""); err != nil {
		return fmt.Errorf("gagal mount /tmp: %v", err)
	}

	// Pivot root
	if err := pivotRoot(rootfs); err != nil {
		return fmt.Errorf("gagal pivot root: %v", err)
	}

	return nil
}

// pivotRoot memindahkan root filesystem container (khusus Linux)
func pivotRoot(rootfs string) error {
	// Membuat direktori oldroot
	pivotDir := filepath.Join(rootfs, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0700); err != nil {
		return err
	}

	// Pastikan rootfs sudah di-mount
	if err := syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs: %v", err)
	}

	// Pindahkan old_root ke pivot_dir
	if err := syscall.PivotRoot(rootfs, pivotDir); err != nil {
		return fmt.Errorf("pivot_root: %v", err)
	}

	// Chdir ke root baru
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("chdir /: %v", err)
	}

	// Unmount old_root
	pivotDir = "/.pivot_root"
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root: %v", err)
	}

	// Hapus pivot_dir
	if err := os.RemoveAll(pivotDir); err != nil {
		return fmt.Errorf("remove pivot_root: %v", err)
	}

	return nil
}

// setupCgroupsLinux mengatur cgroups untuk batasi resource container
func setupCgroupsLinux() error {
	// Tentukan pid saat ini
	pid := os.Getpid()

	// Setup cgroup v2
	cgroupRoot := "/sys/fs/cgroup"
	containerCgroup := filepath.Join(cgroupRoot, fmt.Sprintf("minidocker_%d", pid))

	// Membuat cgroup baru
	if err := os.Mkdir(containerCgroup, 0755); err != nil {
		return fmt.Errorf("gagal membuat cgroup: %v", err)
	}

	// Batasi memory (64 MB)
	memLimit := "67108864" // 64MB
	memFile := filepath.Join(containerCgroup, "memory.max")
	if err := os.WriteFile(memFile, []byte(memLimit), 0644); err != nil {
		return fmt.Errorf("gagal set memory limit: %v", err)
	}

	// Batasi CPU (10%)
	cpuLimit := "10000" // 10% dari 100000
	cpuFile := filepath.Join(containerCgroup, "cpu.max")
	if err := os.WriteFile(cpuFile, []byte(cpuLimit+" 100000"), 0644); err != nil {
		return fmt.Errorf("gagal set cpu limit: %v", err)
	}

	// Tambahkan pid ke cgroup
	procsFile := filepath.Join(containerCgroup, "cgroup.procs")
	if err := os.WriteFile(procsFile, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		return fmt.Errorf("gagal tambahkan pid ke cgroup: %v", err)
	}

	return nil
}

// Implementasi khusus Linux untuk createLinuxSysProcAttr
func createLinuxSysProcAttrImpl() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | // Hostname & domain
			syscall.CLONE_NEWPID | // Process ID
			syscall.CLONE_NEWNS | // Mount
			syscall.CLONE_NEWNET | // Network
			syscall.CLONE_NEWIPC, // Inter-process communication
	}
} 