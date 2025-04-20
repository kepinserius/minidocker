package cmd

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/user/minidocker/container"
)

// RunCommand - Perintah untuk menjalankan container
func RunCommand() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Jalankan container dengan image tertentu",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "image",
				Aliases: []string{"i"},
				Usage:   "Image yang digunakan (contoh: alpine)",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Nama container",
			},
			&cli.StringSliceFlag{
				Name:    "volume",
				Aliases: []string{"v"},
				Usage:   "Mount volume (format: host-path:container-path)",
			},
			&cli.StringSliceFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Map port (format: host-port:container-port)",
			},
			&cli.StringFlag{
				Name:    "memory",
				Aliases: []string{"m"},
				Usage:   "Batasan memory (format: 64m, 128m, 256m)",
				Value:   "64m",
			},
			&cli.StringFlag{
				Name:    "cpu",
				Aliases: []string{"c"},
				Usage:   "Batasan CPU dalam persentase (0-100)",
				Value:   "10",
			},
			&cli.StringFlag{
				Name:    "security-profile",
				Aliases: []string{"s"},
				Usage:   "Profil keamanan (default, restricted, privileged)",
				Value:   "default",
			},
			&cli.BoolFlag{
				Name:    "read-only",
				Usage:   "Jalankan container dengan filesystem read-only",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "privileged",
				Usage:   "Jalankan container dalam mode privileged (mengesampingkan security-profile)",
				Value:   false,
			},
		},
		Action: func(ctx *cli.Context) error {
			imageName := ctx.String("image")
			containerName := ctx.String("name")
			volumes := ctx.StringSlice("volume")
			ports := ctx.StringSlice("port")
			memory := ctx.String("memory")
			cpu := ctx.String("cpu")
			
			// Security options
			securityProfile := ctx.String("security-profile")
			readOnly := ctx.Bool("read-only")
			privileged := ctx.Bool("privileged")
			
			// Jika privileged, override security profile
			if privileged {
				securityProfile = "privileged"
			}
			
			// Dapatkan profil keamanan berdasarkan nama
			secProfile, err := container.GetSecurityProfile(securityProfile)
			if err != nil {
				return err
			}
			
			// Override read-only flag jika diberikan
			if readOnly {
				secProfile.ReadOnlyRootfs = true
			}
			
			return container.RunContainerWithSecurity(imageName, containerName, volumes, ports, memory, cpu, secProfile)
		},
	}
}

// ListCommand - Perintah untuk melihat daftar container
func ListCommand() *cli.Command {
	return &cli.Command{
		Name:    "ps",
		Aliases: []string{"list"},
		Usage:   "Daftar semua container",
		Action: func(ctx *cli.Context) error {
			return container.ListContainers()
		},
	}
}

// StopCommand - Perintah untuk menghentikan container
func StopCommand() *cli.Command {
	return &cli.Command{
		Name:  "stop",
		Usage: "Hentikan container yang sedang berjalan",
		ArgsUsage: "CONTAINER_ID",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 1 {
				return fmt.Errorf("Diperlukan ID container")
			}
			containerId := ctx.Args().First()
			return container.StopContainer(containerId)
		},
	}
}

// LogsCommand - Perintah untuk melihat logs container
func LogsCommand() *cli.Command {
	return &cli.Command{
		Name:  "logs",
		Usage: "Tampilkan logs dari container",
		ArgsUsage: "CONTAINER_ID",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "follow",
				Aliases: []string{"f"},
				Usage:   "Ikuti logs secara real-time",
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 1 {
				return fmt.Errorf("Diperlukan ID container")
			}
			containerId := ctx.Args().First()
			follow := ctx.Bool("follow")
			return container.ContainerLogs(containerId, follow)
		},
	}
}

// ExecCommand - Perintah untuk menjalankan perintah di dalam container yang berjalan
func ExecCommand() *cli.Command {
	return &cli.Command{
		Name:  "exec",
		Usage: "Jalankan perintah di dalam container yang sedang berjalan",
		ArgsUsage: "CONTAINER_ID COMMAND [ARGS...]",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 2 {
				return fmt.Errorf("Diperlukan ID container dan perintah")
			}
			containerId := ctx.Args().First()
			command := ctx.Args().Slice()[1:]
			return container.ExecInContainer(containerId, command)
		},
	}
}

// VolumeCreateCommand - Perintah untuk membuat volume
func VolumeCreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "volume-create",
		Usage: "Buat volume baru",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Nama volume",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:    "label",
				Aliases: []string{"l"},
				Usage:   "Label volume (format: key=value)",
			},
		},
		Action: func(ctx *cli.Context) error {
			name := ctx.String("name")
			labelSlice := ctx.StringSlice("label")
			
			// Konversi array label menjadi map
			labels := make(map[string]string)
			for _, label := range labelSlice {
				parts := strings.SplitN(label, "=", 2)
				if len(parts) == 2 {
					labels[parts[0]] = parts[1]
				}
			}
			
			_, err := container.CreateVolume(name, labels)
			return err
		},
	}
}

// VolumeListCommand - Perintah untuk melihat daftar volume
func VolumeListCommand() *cli.Command {
	return &cli.Command{
		Name:  "volume-list",
		Usage: "Daftar semua volume",
		Action: func(ctx *cli.Context) error {
			volumes, err := container.ListVolumes()
			if err != nil {
				return err
			}
			
			fmt.Printf("%-20s %-20s %-40s %-20s\n", "VOLUME NAME", "DRIVER", "MOUNTPOINT", "CREATED")
			for _, v := range volumes {
				createdAgo := ""
				if !v.CreatedAt.IsZero() {
					createdAgo = fmt.Sprintf("%s ago", strings.TrimSpace(
						strings.Replace(
							strings.Replace(
								v.CreatedAt.String(), 
								v.CreatedAt.Format("15:04:05"), 
								"", 
								1,
							), 
							v.CreatedAt.Format("2006-01-02"), 
							"", 
							1,
						),
					))
				}
				
				fmt.Printf("%-20s %-20s %-40s %-20s\n", 
					v.Name, v.Driver, v.Mountpoint, createdAgo)
			}
			
			return nil
		},
	}
}

// VolumeRemoveCommand - Perintah untuk menghapus volume
func VolumeRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:  "volume-rm",
		Usage: "Hapus volume",
		ArgsUsage: "VOLUME_NAME",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Paksa penghapusan volume yang sedang digunakan",
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 1 {
				return fmt.Errorf("Diperlukan nama volume")
			}
			volumeName := ctx.Args().First()
			force := ctx.Bool("force")
			return container.RemoveVolume(volumeName, force)
		},
	}
}

// VolumeBackupCommand - Perintah untuk backup volume
func VolumeBackupCommand() *cli.Command {
	return &cli.Command{
		Name:  "volume-backup",
		Usage: "Backup data volume ke file",
		ArgsUsage: "VOLUME_NAME BACKUP_PATH",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 2 {
				return fmt.Errorf("Diperlukan nama volume dan path backup")
			}
			volumeName := ctx.Args().First()
			backupPath := ctx.Args().Get(1)
			return container.BackupVolumeData(volumeName, backupPath)
		},
	}
}

// VolumeRestoreCommand - Perintah untuk restore volume
func VolumeRestoreCommand() *cli.Command {
	return &cli.Command{
		Name:  "volume-restore",
		Usage: "Restore data volume dari file backup",
		ArgsUsage: "VOLUME_NAME BACKUP_PATH",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 2 {
				return fmt.Errorf("Diperlukan nama volume dan path backup")
			}
			volumeName := ctx.Args().First()
			backupPath := ctx.Args().Get(1)
			return container.RestoreVolumeData(volumeName, backupPath)
		},
	}
}

// Registry commands

// RegistryStartCommand - Perintah untuk memulai registry lokal
func RegistryStartCommand() *cli.Command {
	return &cli.Command{
		Name:  "registry-start",
		Usage: "Jalankan registry container lokal",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Port untuk registry",
				Value:   5000,
			},
		},
		Action: func(ctx *cli.Context) error {
			port := ctx.Int("port")
			return container.StartLocalRegistry(port)
		},
	}
}

// PullCommand - Perintah untuk pull image
func PullCommand() *cli.Command {
	return &cli.Command{
		Name:  "pull",
		Usage: "Unduh image dari registry",
		ArgsUsage: "IMAGE_NAME[:TAG]",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 1 {
				return fmt.Errorf("Diperlukan nama image")
			}
			imageName := ctx.Args().First()
			return container.PullImage(imageName)
		},
	}
}

// PushCommand - Perintah untuk push image
func PushCommand() *cli.Command {
	return &cli.Command{
		Name:  "push",
		Usage: "Unggah image ke registry",
		ArgsUsage: "IMAGE_NAME[:TAG]",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 1 {
				return fmt.Errorf("Diperlukan nama image")
			}
			imageName := ctx.Args().First()
			return container.PushImage(imageName)
		},
	}
}

// ImagesCommand - Perintah untuk menampilkan daftar image
func ImagesCommand() *cli.Command {
	return &cli.Command{
		Name:  "images",
		Usage: "Daftar semua image",
		Action: func(ctx *cli.Context) error {
			images, err := container.ListImages()
			if err != nil {
				return err
			}
			
			fmt.Printf("%-30s %-15s %-15s %-25s\n", "REPOSITORY", "TAG", "SIZE", "CREATED")
			for _, img := range images {
				// Format created time
				createdAgo := ""
				if !img.CreatedAt.IsZero() {
					createdAgo = fmt.Sprintf("%s ago", strings.TrimSpace(
						strings.Replace(
							strings.Replace(
								img.CreatedAt.String(), 
								img.CreatedAt.Format("15:04:05"), 
								"", 
								1,
							), 
							img.CreatedAt.Format("2006-01-02"), 
							"", 
							1,
						),
					))
				}
				
				// Format size
				sizeStr := fmt.Sprintf("%.2f MB", float64(img.Size)/(1024*1024))
				
				fmt.Printf("%-30s %-15s %-15s %-25s\n", 
					img.Name, img.Tag, sizeStr, createdAgo)
			}
			
			return nil
		},
	}
}

// TagCommand - Perintah untuk membuat tag image
func TagCommand() *cli.Command {
	return &cli.Command{
		Name:  "tag",
		Usage: "Buat tag baru untuk image",
		ArgsUsage: "SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 2 {
				return fmt.Errorf("Diperlukan source dan target image")
			}
			sourceImage := ctx.Args().First()
			targetImage := ctx.Args().Get(1)
			return container.TagImage(sourceImage, targetImage)
		},
	}
} 