package cmd

import (
	"log"

	"github.com/Sherlock-Holo/lightc/libexec"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach [container id]",
	Short: "attach a container",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		id := args[0]

		if err := libexec.Attach(id); err != nil {
			log.Fatal(err)
		}
	},
}
