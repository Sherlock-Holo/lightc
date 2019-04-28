package cmd

import (
	"strings"

	"github.com/Sherlock-Holo/lightc/libexec"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	execTTY bool
)

var execCmd = &cobra.Command{
	Use:   "exec [container id ] [cmd...]",
	Short: "exec into a running container",
	Args:  cobra.MinimumNArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		id := args[0]
		cmd := strings.Join(args[1:], " ")

		if err := libexec.Exec(id, cmd, execTTY); err != nil {
			logrus.Fatal(err)
		}
	},
}
