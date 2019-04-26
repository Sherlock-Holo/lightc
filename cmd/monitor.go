package cmd

import (
	"github.com/Sherlock-Holo/lightc/libexec/monitor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "DO NOT RUN IT DIRECTLY",
	Args:  cobra.ExactArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		if err := monitor.Monitor(); err != nil {
			logrus.Fatal(err)
		}
	},
}
