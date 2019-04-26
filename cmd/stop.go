package cmd

import (
	"fmt"

	"github.com/Sherlock-Holo/lightc/libexec"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [container id]",
	Short: "stop a container",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		id := args[0]

		if err := libexec.Stop(id, false); err != nil {
			logrus.Fatal(err)
		}

		fmt.Println(id)
	},
}
