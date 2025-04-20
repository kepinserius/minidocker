package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/user/minidocker/cmd"
	"github.com/user/minidocker/container"
)

func main() {
	app := &cli.App{
		Name:  "minidocker",
		Usage: "Container runtime sederhana seperti Docker",
		Commands: []*cli.Command{
			cmd.RunCommand(),
			cmd.ListCommand(),
			cmd.StopCommand(),
			cmd.LogsCommand(),
			cmd.ExecCommand(),
			cmd.VolumeCreateCommand(),
			cmd.VolumeListCommand(),
			cmd.VolumeRemoveCommand(),
			cmd.VolumeBackupCommand(),
			cmd.VolumeRestoreCommand(),
			cmd.RegistryStartCommand(),
			cmd.PullCommand(),
			cmd.PushCommand(),
			cmd.ImagesCommand(),
			cmd.TagCommand(),
			{
				Name:     "internal-start",
				Usage:    "Perintah internal untuk memulai container",
				HideHelp: true,
				Hidden:   true,
				Action: func(ctx *cli.Context) error {
					if ctx.NArg() < 1 {
						return fmt.Errorf("rootfs path diperlukan")
					}
					rootfs := ctx.Args().First()
					return container.InternalStartContainer(rootfs)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(fmt.Sprintf("Error menjalankan aplikasi: %v", err))
	}
} 