package cmd

import (
	"fmt"

	"github.com/Sherlock-Holo/lightc/libexec"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill [container id]",
	Short: "kill a container",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		id := args[0]

		if err := libexec.Stop(id, true); err != nil {
			logrus.Fatal(err)
		}

		fmt.Println(id)
	},
}
