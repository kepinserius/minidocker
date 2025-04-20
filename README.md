# MiniDocker - Container Runtime Sederhana

![MiniDocker Logo](https://via.placeholder.com/150x150.png?text=MiniDocker)

MiniDocker adalah implementasi sederhana dari container runtime yang terinspirasi oleh Docker. Proyek ini dibangun dengan fokus untuk tujuan pendidikan dan pemahaman dasar teknologi container, menggunakan Golang untuk memanfaatkan fitur-fitur inti Linux seperti namespaces, cgroups, dan mount isolation untuk menyediakan isolasi proses.

## Daftar Isi

- [Pendahuluan](#pendahuluan)
- [Fitur](#fitur)
- [Persyaratan Sistem](#persyaratan-sistem)
- [Cara Menggunakan](#cara-menggunakan)
- [Perintah Tersedia](#perintah-tersedia)
- [Arsitektur](#arsitektur)
- [Teknologi dan Konsep Kunci](#teknologi-dan-konsep-kunci)
- [Perbandingan dengan Docker](#perbandingan-dengan-docker)
- [Keterbatasan dan Area Pengembangan](#keterbatasan-dan-area-pengembangan)
- [Rencana Pengembangan di Masa Depan](#rencana-pengembangan-di-masa-depan)
- [Referensi](#referensi)
- [Lisensi](#lisensi)

## Pendahuluan

MiniDocker dibuat untuk memahami fundamental teknologi container dengan menerapkan konsep-konsep kunci seperti:

1. Isolasi proses menggunakan namespaces Linux
2. Pengelolaan sumber daya melalui cgroups v2
3. Mekanisme mount filesystem dan layering image
4. Pengelolaan container melalui CLI sederhana

Proyek ini mendemonstrasikan bagaimana Docker dan container runtime lainnya bekerja di level rendah, memberikan wawasan tentang teknologi yang membentuk dasar containerization modern.

## Fitur

- **CLI Sederhana**:

  - `run`: Menjalankan container baru dengan opsi keamanan dan resource limits
  - `ps`/`list`: Menampilkan daftar container
  - `stop`: Menghentikan container yang sedang berjalan
  - `logs`: Melihat output logs container dengan opsi real-time follow
  - `exec`: Menjalankan perintah dalam container yang sedang berjalan

- **Isolasi Container**:

  - Namespace isolation (PID, UTS, MNT, NET, IPC)
  - Resource limits dengan cgroups v2 (memory dan CPU)
  - Filesystem isolation dengan chroot dan pivot_root
  - Security profiles (default, restricted, privileged)

- **Manajemen Image**:

  - Format image tar.gz sederhana
  - Dukungan untuk Alpine dan BusyBox
  - Registry lokal sederhana
  - Pull, push, tag, dan images commands

- **Volume Management**:

  - Membuat, menampilkan, dan menghapus volume
  - Backup dan restore volume
  - Mounting volume ke container

- **Networking**:

  - Port mapping sederhana
  - Simulasi NAT untuk jaringan container

- **Keamanan**:

  - Seccomp profiles untuk membatasi syscalls
  - Capabilities management
  - AppArmor profiles
  - Read-only filesystem option

- **Kompatibilitas Lintas Platform**:
  - Implementasi penuh di Linux
  - Mode simulasi untuk Windows dan platform lainnya

## Persyaratan Sistem

### Dependensi

- Go 1.21 atau lebih baru
- Sistem Operasi Linux (untuk fungsionalitas penuh)
- Akses root/sudo (untuk namespace dan cgroups)

### Untuk Pengembangan

```bash
# Clone repositori
git clone https://github.com/user/minidocker.git
cd minidocker

# Menginstall dependensi runtime
go mod tidy

# Build
go build -o minidocker
```

## Cara Menggunakan

### Menjalankan Container

```bash
# Menjalankan container Alpine
sudo ./minidocker run --image alpine --name my-alpine-container

# Menjalankan container dengan batasan resource
sudo ./minidocker run --image busybox --memory 128m --cpu 20

# Menjalankan container dengan port mapping
sudo ./minidocker run --image nginx -p 8080:80

# Menjalankan container dengan volume
sudo ./minidocker run --image ubuntu -v my_vol:/data

# Menjalankan container dengan profil keamanan
sudo ./minidocker run --image alpine --security-profile restricted
```

### Melihat Container yang Berjalan

```bash
sudo ./minidocker ps
# atau
sudo ./minidocker list
```

### Melihat Logs Container

```bash
# Melihat logs
sudo ./minidocker logs <container_id>

# Mengikuti logs secara real-time
sudo ./minidocker logs -f <container_id>
```

### Menjalankan Perintah di Container

```bash
sudo ./minidocker exec <container_id> ls -la
```

### Menghentikan Container

```bash
sudo ./minidocker stop <container_id>
```

### Manajemen Volume

```bash
# Membuat volume baru
sudo ./minidocker volume-create --name my_volume

# Melihat daftar volume
sudo ./minidocker volume-list

# Menghapus volume
sudo ./minidocker volume-rm my_volume

# Backup volume
sudo ./minidocker volume-backup my_volume /path/to/backup.tar

# Restore volume
sudo ./minidocker volume-restore my_volume /path/to/backup.tar
```

### Manajemen Image

```bash
# Melihat daftar image
sudo ./minidocker images

# Pull image dari registry
sudo ./minidocker pull alpine:latest

# Push image ke registry
sudo ./minidocker push myimage:latest

# Membuat tag image baru
sudo ./minidocker tag alpine:latest myalpine:v1

# Menjalankan registry lokal
sudo ./minidocker registry-start -p 5000
```

## Perintah Tersedia

MiniDocker menyediakan berbagai perintah untuk mengelola container, volume, dan image:

### Container Management

- `run`: Menjalankan container baru
- `ps`/`list`: Menampilkan daftar container
- `stop`: Menghentikan container yang sedang berjalan
- `logs`: Melihat output logs container
- `exec`: Menjalankan perintah dalam container yang sedang berjalan

### Volume Management

- `volume-create`: Membuat volume baru
- `volume-list`: Menampilkan daftar volume
- `volume-rm`: Menghapus volume
- `volume-backup`: Backup data volume ke file
- `volume-restore`: Restore data volume dari file backup

### Image Management

- `images`: Menampilkan daftar image
- `pull`: Mengunduh image dari registry
- `push`: Mengunggah image ke registry
- `tag`: Membuat tag baru untuk image
- `registry-start`: Menjalankan registry lokal

### Opsi Keamanan

- `--security-profile`: Menentukan profil keamanan (default, restricted, privileged)
- `--read-only`: Menjalankan container dengan filesystem read-only
- `--privileged`: Menjalankan container dalam mode privileged

### Resource Limits

- `--memory`: Batasan memory (format: 64m, 128m, 256m)
- `--cpu`: Batasan CPU dalam persentase (0-100)

## Arsitektur

MiniDocker terdiri dari beberapa komponen utama:

### 1. Command Line Interface (CLI)

Interface pengguna berbasis command line yang menyediakan perintah dasar untuk mengelola container.

### 2. Container Manager

Bertanggung jawab untuk:

- Mengelola siklus hidup container (create, start, stop)
- Menyimpan metadata container
- Isolasi namespace dan cgroups
- Pengelolaan proses container

### 3. Image Manager

Menyediakan fungsi untuk:

- Ekstraksi dan pengelolaan image
- Simulasi download dari registry
- Manajemen layer image sederhana

### 4. Volume Manager

Bertanggung jawab untuk:

- Membuat dan mengelola persistent volumes
- Menyediakan mekanisme backup dan restore
- Melakukan mounting ke container

### 5. Security Manager

Bertanggung jawab untuk:

- Menerapkan profil keamanan pada container
- Mengelola seccomp profiles
- Membatasi capabilities container
- Mendukung read-only filesystem

### 6. Registry Server

Menyediakan:

- HTTP API sederhana untuk pull/push image
- Manajemen metadata image
- Catalog API

### Diagram Alir Operasi

#### Proses `run`:

```
+----------------+     +---------------+     +---------------------+
| Parse Command  | --> | Create Rootfs | --> | Extract Image       |
+----------------+     +---------------+     +---------------------+
         |                                            |
         v                                            v
+----------------+     +---------------+     +---------------------+
| Setup Namespace| <-- | Create Config | <-- | Prepare Environment |
+----------------+     +---------------+     +---------------------+
         |
         v
+----------------+     +---------------+     +---------------------+
| Setup Mounts   | --> | Pivot Root    | --> | Setup Security      |
+----------------+     +---------------+     +---------------------+
         |
         v
+----------------+     +---------------+
| Setup Cgroups  | --> | Execute Shell |
+----------------+     +---------------+
```

### Struktur Direktori

MiniDocker mengorganisasi data sebagai berikut:

- `/var/run/minidocker/containers/`: Menyimpan metadata dan rootfs container
- `/var/run/minidocker/images/`: Menyimpan image cache
- `/var/run/minidocker/volumes/`: Menyimpan persistent volumes
- `/var/run/minidocker/registry/`: Menyimpan image registry
- `/etc/minidocker/seccomp/`: Menyimpan seccomp profiles

## Siklus Hidup Container

Proses container dalam MiniDocker melalui beberapa tahap:

1. **Create**: Membuat konfigurasi container dan direktori rootfs
2. **Start**: Menjalankan proses container dengan namespace baru
3. **Setup Security**: Menerapkan profil keamanan
4. **Execute**: Menjalankan program/shell dalam container
5. **Monitor**: Memantau status container
6. **Stop**: Mengirim sinyal untuk menghentikan container

## Teknologi dan Konsep Kunci

### Namespaces Linux

MiniDocker menggunakan 5 namespace Linux untuk isolasi:

- **UTS Namespace**: Isolasi hostname dan domain name
- **PID Namespace**: Isolasi proses ID
- **Mount Namespace**: Isolasi filesystem mount
- **Network Namespace**: Isolasi network stack
- **IPC Namespace**: Isolasi Inter-Process Communication

Implementasi menggunakan syscall `clone()` dengan flag namespaces:

```go
syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC
```

### Control Groups (cgroups v2)

Control Groups digunakan untuk membatasi dan mengisolasi penggunaan sumber daya:

- **Memory**: Membatasi penggunaan memori (64 MB default)
- **CPU**: Membatasi penggunaan CPU (10% default)

Contoh kode implementasi cgroups:

```go
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
```

### Filesystem dan Image Layers

MiniDocker menggunakan konsep image layer berbasis filesystem:

1. **Rootfs**: Direktori yang berisi sistem operasi dasar
2. **Image Layer**: Implementasi sederhana menggunakan tar.gz
3. **Mount Isolation**: Menggunakan pivot_root untuk isolasi filesystem

### Volume Management

Sistem volume memberikan data persistence:

1. **Volume Creation**: Membuat volume dengan metadata
2. **Volume Mounting**: Memasang volume ke container
3. **Backup/Restore**: Menyimpan dan memulihkan data volume

### Registry Server

Registry server sederhana menyediakan:

1. **API**: Implementasi sederhana dari Docker Registry v2 API
2. **Image Metadata**: Penyimpanan metadata image
3. **Image Tags**: Dukungan untuk tag dan versioning

### Container Security

Mekanisme keamanan meliputi:

1. **Seccomp Profiles**: Membatasi syscalls yang dapat digunakan
2. **Capabilities**: Membatasi Linux capabilities pada container
3. **AppArmor Profiles**: Menerapkan AppArmor policies
4. **NoNewPrivs**: Mencegah escalation privileges
5. **Read-Only Rootfs**: Mencegah perubahan pada filesystem

### Pivot Root

Pivot root digunakan untuk mengubah root filesystem untuk container:

```go
if err := syscall.PivotRoot(rootfs, pivotDir); err != nil {
    return fmt.Errorf("pivot_root: %v", err)
}
```

## Perbandingan dengan Docker

| Fitur               | Docker                | MiniDocker                      |
| ------------------- | --------------------- | ------------------------------- |
| Virtualisasi        | Container             | Container                       |
| Namespace Isolation | Penuh                 | Dasar (UTS, PID, MNT, NET, IPC) |
| Resource Limits     | cgroups v1/v2         | cgroups v2 (dasar)              |
| Image Format        | OCI Image Format      | tar.gz sederhana                |
| Networking          | Bridge, Host, Overlay | Port mapping sederhana          |
| Storage Drivers     | overlay2, btrfs, dll  | Sederhana (tanpa CoW)           |
| Volume Mounts       | Bind, Volume, tmpfs   | Basic volume management         |
| Security            | seccomp, AppArmor     | seccomp, AppArmor (sederhana)   |
| Registry            | Docker Hub, Private   | Registry lokal sederhana        |
| Orchestration       | Swarm, Kubernetes     | Tidak ada                       |

## Keterbatasan dan Area Pengembangan

MiniDocker sengaja dibuat sederhana dan memiliki beberapa keterbatasan:

1. **Networking**: Implementasi network bridge masih sederhana
2. **Security**: Implementasi seccomp dan AppArmor belum lengkap
3. **Storage Driver**: Tidak ada implementasi copy-on-write
4. **Image Registry**: Implementasi registry masih sangat dasar
5. **Resource Controls**: Implementasi cgroups minimal
6. **Windows Support**: Hanya berjalan dalam mode simulasi di Windows

## Rencana Pengembangan di Masa Depan

Berikut adalah 10 area pengembangan utama yang bisa dilakukan untuk meningkatkan MiniDocker:

### 1. Dukungan Networking yang Lebih Baik

- Implementasi network namespace yang lebih lengkap
- Dukungan bridge network untuk komunikasi antar container
- Mekanisme port mapping untuk mengekspos port container ke host
- Implementasi DNS resolver sederhana untuk container

### 2. Sistem Storage yang Lebih Canggih

- Implementasi overlay filesystem untuk layer images yang lebih efisien
- Dukungan volume yang lebih baik dengan manajemen volume terpisah
- Mekanisme persistent storage yang lebih handal
- Backup dan restore volume container

### 3. Peningkatan Keamanan

- Implementasi seccomp profile untuk membatasi syscall
- Dukungan AppArmor/SELinux untuk keamanan tambahan
- Capabilities management untuk membatasi hak akses container
- User namespace untuk mapping user container ke host

### 4. Perluasan Fitur Image

- Dukungan untuk format OCI (Open Container Initiative)
- Implementasi image registry sederhana
- Sistem caching yang lebih efisien untuk layer image
- Fitur untuk membuat image kustom dari container yang berjalan

### 5. Perbaikan UI dan UX

- Menambahkan fitur logs untuk melihat output container
- Implementasi fitur exec untuk menjalankan perintah di container yang sudah berjalan
- Dashboard web sederhana untuk monitoring
- Notifikasi dan alert untuk container yang bermasalah

### 6. Dukungan Orkestrasi Sederhana

- Kemampuan untuk mengelola beberapa container sebagai satu kesatuan
- Fitur auto-restart untuk container yang crash
- Load balancing sederhana antar container
- Health check dan auto-recovery

### 7. Optimasi Performa

- Memory footprint yang lebih kecil
- Start-up time yang lebih cepat
- Penggunaan CPU yang lebih efisien
- Mekanisme scaling yang lebih baik untuk beban kerja besar

### 8. Dokumentasi dan Testing

- Menambahkan unit test dan integration test
- Contoh penggunaan dan tutorial yang lebih lengkap
- Dokumentasi API untuk pengembangan lebih lanjut
- Panduan kontribusi untuk para pengembang

### 9. Cross-platform yang Lebih Baik

- Dukungan yang lebih baik untuk Windows dengan WSL2
- Mode simulasi yang lebih realistis untuk sistem non-Linux
- Kompatibilitas dengan macOS
- Dukungan untuk arsitektur ARM (Raspberry Pi, dll)

### 10. Fitur CI/CD

- Integrasi dengan pipeline CI/CD populer
- Dukungan untuk automated build dan deployment
- Automatic updates untuk image
- Mekanisme untuk versioning dan rollback container

Pengembangan fitur-fitur di atas akan membantu MiniDocker menjadi lebih mirip dengan Docker asli sekaligus memberikan nilai pendidikan yang lebih tinggi sebagai proyek pembelajaran tentang teknologi container.

## Implementasi Teknis

Beberapa poin penting dalam implementasi MiniDocker:

### Container Runtime

```go
// Menjalankan container baru
func RunContainer(imageName, containerName string, volumes []string, ports []string, memory string, cpu string) error {
    // Setup rootfs dan ekstrak image
    // ...

    // Setup namespaces untuk isolasi
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS |
                    syscall.CLONE_NEWPID |
                    syscall.CLONE_NEWNS |
                    syscall.CLONE_NEWNET |
                    syscall.CLONE_NEWIPC,
    }

    // Mulai proses container
    // ...
}
```

### Security Profiles

```go
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
```

### Platform Support

MiniDocker dirancang dengan dukungan lintas platform:

```go
// Deteksi platform
if runtime.GOOS != "linux" {
    // Mode simulasi/demo
} else {
    // Implementasi penuh Linux
}
```

## Referensi

1. [Linux Namespaces](https://man7.org/linux/man-pages/man7/namespaces.7.html)
2. [Control Groups v2](https://www.kernel.org/doc/Documentation/cgroup-v2.txt)
3. [OCI Runtime Specification](https://github.com/opencontainers/runtime-spec)
4. [Pivot Root](https://man7.org/linux/man-pages/man2/pivot_root.2.html)
5. [Go Documentation](https://golang.org/doc/)
6. [Syscall Package in Go](https://golang.org/pkg/syscall/)
7. [Docker Registry HTTP API V2](https://docs.docker.com/registry/spec/api/)
8. [Seccomp in Containers](https://docs.docker.com/engine/security/seccomp/)
9. [Linux Capabilities](https://man7.org/linux/man-pages/man7/capabilities.7.html)
10. [AppArmor](https://wiki.ubuntu.com/AppArmor)

## Lisensi

Proyek ini dilisensikan di bawah [MIT License](LICENSE).

---

Dibuat untuk tujuan pendidikan dan pemahaman konsep-konsep container. Tidak disarankan untuk digunakan dalam lingkungan produksi.
