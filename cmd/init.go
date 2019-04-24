package cmd

import (
	"github.com/Sherlock-Holo/lightc/libexec"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "DO NOT RUN IT DIRECTLY",
	Run: func(_ *cobra.Command, _ []string) {
		if err := libexec.InitPorcess(); err != nil {
			logrus.Fatal(err)
		}
	},
}
