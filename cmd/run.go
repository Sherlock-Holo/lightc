package cmd

import (
	"fmt"
	"strings"

	"github.com/Sherlock-Holo/lightc/libexec"
	"github.com/Sherlock-Holo/lightc/libstorage/rootfs"
	"github.com/Sherlock-Holo/lightc/libstorage/volume"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	tty               bool
	runNetwork        string
	removeAfterAttach bool
	detach            bool
	envs              []string
	volumes           []string
)

var runCmd = &cobra.Command{
	Use:   "run [image name] [cmd...]",
	Short: "run a container",
	Args:  cobra.MinimumNArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		imageName := args[0]
		cmd := strings.Join(args[1:], " ")

		rootFS, err := rootfs.Create(imageName)
		if err != nil {
			logrus.Fatal(err)
		}

		if err := rootfs.Mount(rootFS); err != nil {
			logrus.Fatal(err)
		}

		info, err := libexec.Run(
			cmd,
			rootFS,
			envs,
			runNetwork,
			volume.Parse(volumes, rootFS.MergedDir),
			tty,
			detach,
			removeAfterAttach,
		)
		if err != nil {
			logrus.Fatal(err)
		}

		if detach {
			fmt.Println(info.ID)
		}
	},
}
