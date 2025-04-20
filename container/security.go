package container

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SecurityProfile mendefinisikan profil keamanan untuk container
type SecurityProfile struct {
	Name          string   `json:"name"`
	SeccompProfile string  `json:"seccomp_profile"`
	Capabilities   []string `json:"capabilities"`
	NoNewPrivs     bool     `json:"no_new_privs"`
	ReadOnlyRootfs bool     `json:"read_only_rootfs"`
	AppArmorProfile string `json:"apparmor_profile"`
}

// DefaultSecurityProfile memberikan profil keamanan default
func DefaultSecurityProfile() SecurityProfile {
	return SecurityProfile{
		Name:          "default",
		SeccompProfile: "default",
		Capabilities:   []string{"CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER", "MKNOD", "NET_RAW", "SETGID", "SETUID", "SETFCAP", "SETPCAP", "NET_BIND_SERVICE", "SYS_CHROOT", "KILL", "AUDIT_WRITE"},
		NoNewPrivs:     true,
		ReadOnlyRootfs: false,
		AppArmorProfile: "minidocker-default",
	}
}

// RestrictedSecurityProfile memberikan profil keamanan yang lebih ketat
func RestrictedSecurityProfile() SecurityProfile {
	return SecurityProfile{
		Name:          "restricted",
		SeccompProfile: "restricted",
		Capabilities:   []string{"CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER", "NET_BIND_SERVICE", "SETGID", "SETUID"},
		NoNewPrivs:     true,
		ReadOnlyRootfs: true,
		AppArmorProfile: "minidocker-restricted",
	}
}

// PrivilegedSecurityProfile memberikan profil dengan semua capabilities
func PrivilegedSecurityProfile() SecurityProfile {
	return SecurityProfile{
		Name:          "privileged",
		SeccompProfile: "unconfined",
		Capabilities:   []string{"ALL"},
		NoNewPrivs:     false,
		ReadOnlyRootfs: false,
		AppArmorProfile: "unconfined",
	}
}

// GetSecurityProfile mendapatkan profil berdasarkan nama
func GetSecurityProfile(name string) (SecurityProfile, error) {
	switch strings.ToLower(name) {
	case "default":
		return DefaultSecurityProfile(), nil
	case "restricted":
		return RestrictedSecurityProfile(), nil
	case "privileged":
		return PrivilegedSecurityProfile(), nil
	default:
		return SecurityProfile{}, fmt.Errorf("profil keamanan '%s' tidak dikenal", name)
	}
}

// GetSeccompProfile mendapatkan path ke file profil seccomp
func GetSeccompProfile(name string) (string, error) {
	// Lokasi default untuk profil seccomp
	profilesDir := "/etc/minidocker/seccomp"

	// Cek apakah direktori ada
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		// Buat direktori jika belum ada
		if err := os.MkdirAll(profilesDir, 0755); err != nil {
			return "", fmt.Errorf("gagal membuat direktori profil seccomp: %v", err)
		}
		
		// Untuk implementasi nyata, kita akan mengisi dengan profil default
		// Ini hanya simulasi
		createDefaultSeccompProfiles(profilesDir)
	}

	// Cek profil seccomp
	if name == "unconfined" {
		return "", nil // Tidak perlu profil untuk unconfined
	}

	profilePath := filepath.Join(profilesDir, fmt.Sprintf("%s.json", name))
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("profil seccomp '%s' tidak ditemukan", name)
	}

	return profilePath, nil
}

// ApplySecurityProfile menerapkan profil keamanan ke container
func ApplySecurityProfile(profile SecurityProfile, containerID string) error {
	fmt.Printf("Menerapkan profil keamanan '%s' untuk container %s\n", profile.Name, containerID)
	fmt.Printf("  - Seccomp: %s\n", profile.SeccompProfile)
	fmt.Printf("  - AppArmor: %s\n", profile.AppArmorProfile)
	fmt.Printf("  - Capabilities: %s\n", strings.Join(profile.Capabilities, ", "))
	fmt.Printf("  - NoNewPrivs: %t\n", profile.NoNewPrivs)
	fmt.Printf("  - ReadOnlyRootfs: %t\n", profile.ReadOnlyRootfs)

	// Ini hanya simulasi, pada implementasi sebenarnya
	// kita akan menggunakan syscall dan library seperti libseccomp, libcap, dll

	return nil
}

// createDefaultSeccompProfiles membuat profil default seccomp (simulasi)
func createDefaultSeccompProfiles(dir string) {
	// Default profile (permisif tapi masih aman)
	defaultProfile := `{
	"defaultAction": "SCMP_ACT_ERRNO",
	"architectures": ["SCMP_ARCH_X86_64", "SCMP_ARCH_X86", "SCMP_ARCH_X32"],
	"syscalls": [
		{
			"names": [
				"accept", "access", "arch_prctl", "bind", "brk", "capget", "capset", "chdir", "chmod",
				"chown", "clone", "close", "connect", "dup", "dup2", "dup3", "epoll_create", "epoll_ctl",
				"epoll_wait", "execve", "exit", "exit_group", "faccessat", "fadvise64", "fchdir", "fchmod",
				"fcntl", "fdatasync", "flock", "fork", "fstat", "fstatfs", "fsync", "futex", "getcwd",
				"getdents", "getdents64", "getegid", "geteuid", "getgid", "getpeername", "getpgrp", 
				"getpid", "getppid", "getpriority", "getrandom", "getresgid", "getresuid", 
				"getrlimit", "getrusage", "getsockname", "getsockopt", "gettid", "gettimeofday", 
				"getuid", "listen", "lseek", "madvise", "mkdir", "mknod", "mmap", "mprotect", 
				"munmap", "nanosleep", "open", "openat", "pipe", "pipe2", "poll", "prctl", 
				"pread64", "prlimit64", "pwrite64", "read", "readlink", "recvfrom", "recvmsg", 
				"rename", "rmdir", "rt_sigaction", "rt_sigprocmask", "rt_sigreturn", "rt_sigsuspend", 
				"select", "sendfile", "sendmsg", "sendto", "set_robust_list", "set_tid_address", 
				"setgid", "setgroups", "setitimer", "setpgid", "setrlimit", "setsid", "setsockopt", 
				"setuid", "sigaltstack", "socket", "socketpair", "stat", "statfs", "symlink", 
				"sysinfo", "umask", "uname", "unlink", "vfork", "wait4", "write", "writev"
			],
			"action": "SCMP_ACT_ALLOW"
		}
	]
}`

	// Restricted profile (lebih ketat)
	restrictedProfile := `{
	"defaultAction": "SCMP_ACT_ERRNO",
	"architectures": ["SCMP_ARCH_X86_64", "SCMP_ARCH_X86", "SCMP_ARCH_X32"],
	"syscalls": [
		{
			"names": [
				"accept", "access", "brk", "close", "dup", "dup2", "dup3", "epoll_create", "epoll_ctl",
				"epoll_wait", "exit", "exit_group", "faccessat", "fchdir", "fcntl", "fdatasync", 
				"flock", "fstat", "fstatfs", "fsync", "futex", "getcwd", "getdents", "getdents64", 
				"getegid", "geteuid", "getgid", "getpeername", "getpgrp", "getpid", "getppid", 
				"getpriority", "getrandom", "getresgid", "getresuid", "getrlimit", "getrusage", 
				"getsockname", "getsockopt", "gettid", "gettimeofday", "getuid", "lseek", "madvise", 
				"mmap", "mprotect", "munmap", "nanosleep", "open", "openat", "pipe", "pipe2", "poll", 
				"prctl", "pread64", "prlimit64", "pwrite64", "read", "readlink", "recvfrom", "recvmsg", 
				"rt_sigaction", "rt_sigprocmask", "rt_sigreturn", "rt_sigsuspend", "select", 
				"sendfile", "sendmsg", "sendto", "set_robust_list", "set_tid_address", "setitimer", 
				"sigaltstack", "socket", "socketpair", "stat", "statfs", "sysinfo", "umask", "uname", 
				"wait4", "write", "writev"
			],
			"action": "SCMP_ACT_ALLOW"
		}
	]
}`

	// Simpan profil ke file
	os.WriteFile(filepath.Join(dir, "default.json"), []byte(defaultProfile), 0644)
	os.WriteFile(filepath.Join(dir, "restricted.json"), []byte(restrictedProfile), 0644)
}

// GetCapabilities menerjemahkan nama capability ke nilai bitmask
func GetCapabilities(caps []string) uint64 {
	// Ini hanya simulasi, pada implementasi sebenarnya
	// kita akan menggunakan library libcap
	
	// Beberapa capability umum dan nilai simulasi
	capMap := map[string]uint64{
		"CHOWN": 0x1,
		"DAC_OVERRIDE": 0x2,
		"DAC_READ_SEARCH": 0x4,
		"FOWNER": 0x8,
		"FSETID": 0x10,
		"KILL": 0x20,
		"SETGID": 0x40,
		"SETUID": 0x80,
		"SETPCAP": 0x100,
		"NET_BIND_SERVICE": 0x200,
		"NET_RAW": 0x400,
		"SYS_CHROOT": 0x800,
		"AUDIT_WRITE": 0x1000,
		"SETFCAP": 0x2000,
		"MAC_OVERRIDE": 0x4000,
		"MAC_ADMIN": 0x8000,
		"ALL": 0xFFFFFFFFFFFFFFFF,
	}

	var result uint64 = 0
	
	for _, cap := range caps {
		if cap == "ALL" {
			return 0xFFFFFFFFFFFFFFFF
		}
		
		if val, ok := capMap[cap]; ok {
			result |= val
		}
	}
	
	return result
} 