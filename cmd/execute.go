package cmd

import "github.com/sirupsen/logrus"

func Execute() {
	rootCmd.AddCommand(
		imagesCmd,
		networkCmd,
		runCmd,
		attachCmd,

		initCmd,
		monitorCmd,
	)

	imagesCmd.AddCommand(
		imagesListCmd,
		imageImportCmd,
		imageRemoveCmd,
	)

	networkCmd.AddCommand(
		createNetworkCmd,
		listNetworkCmd,
		removeNetworkCmd,
	)

	runCmd.Flags().BoolVarP(&tty, "tty", "t", false, "allocate pseudo tty")
	runCmd.Flags().StringVarP(&runNetwork, "net", "n", "", "container join network")
	runCmd.Flags().BoolVar(&removeAfterAttach, "rm", false, "remove container when container stopped")
	runCmd.Flags().BoolVarP(&detach, "detach", "d", false, "detach container")
	runCmd.Flags().StringSliceVarP(&volumes, "volume", "v", nil, "container data volume")
	runCmd.Flags().StringSliceVarP(&envs, "env", "e", nil, "add env into container")

	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
