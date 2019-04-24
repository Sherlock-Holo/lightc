package cmd

import (
	"fmt"
	"net"
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/Sherlock-Holo/lightc/libnetwork"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "container network",
}

var createNetworkCmd = &cobra.Command{
	Use:   "create [network name] [subnet]",
	Short: "create a container network",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		name, subnetStr := args[0], args[1]

		_, subnet, err := net.ParseCIDR(subnetStr)
		if err != nil {
			logrus.Fatal(err)
		}

		if nw, err := libnetwork.NewNetwork(name, *subnet); err != nil {
			logrus.Fatal(err)
		} else {
			fmt.Println(nw.Name, nw.Gateway, nw.Subnet.String())
		}
	},
}

var listNetworkCmd = &cobra.Command{
	Use:   "ls",
	Short: "list container networks",
	Args:  cobra.ExactArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		const listTemplate = "NAME\tgateway\tsubnet\n{{range .}}{{.Name}}\t{{.Gateway}}\t{{.Subnet}}\n{{end}}"

		nws, err := libnetwork.ListNetwork()
		if err != nil {
			logrus.Fatal(err)
		}

		tmpl := template.Must(template.New("network list").Parse(listTemplate))

		writer := tabwriter.NewWriter(os.Stdout, 0, 2, 4, ' ', tabwriter.TabIndent)
		defer func() {
			_ = writer.Flush()
		}()

		if err := tmpl.Execute(writer, nws); err != nil {
			logrus.Fatal(err)
		}
	},
}

var removeNetworkCmd = &cobra.Command{
	Use:   "rm [network]",
	Short: "remove container network",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		name := args[0]

		if err := libnetwork.RemoveNetwork(name); err != nil {
			logrus.Fatal(err)
		}

		fmt.Println(name)
	},
}
