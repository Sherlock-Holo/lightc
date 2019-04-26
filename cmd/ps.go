package cmd

import (
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/Sherlock-Holo/lightc/libexec"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var allContainers bool

var listCmd = &cobra.Command{
	Use:   "ps",
	Short: "list containers",
	Args:  cobra.ExactArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		const listTemplate = "ID\tIMAGE NAME\tSTATUS\tCOMMAND\tCREATED\n{{range .}}{{.ID}}\t{{.ImageName}}\t{{.Status}}\t{{.CmdStr}}\t{{.CreateTime}}\n{{end}}"

		cInfos, err := libexec.List(allContainers)
		if err != nil {
			logrus.Fatal(err)
		}

		tmpl := template.Must(template.New("list").Parse(listTemplate))
		writer := tabwriter.NewWriter(os.Stdout, 0, 2, 4, ' ', tabwriter.TabIndent)
		defer func() {
			_ = writer.Flush()
		}()

		if err := tmpl.Execute(writer, cInfos); err != nil {
			logrus.Fatal(err)
		}
	},
}
