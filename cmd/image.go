package cmd

import (
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/Sherlock-Holo/lightc/libstorage/images"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var imagesCmd = &cobra.Command{
	Use:   "images",
	Short: "lightc images",
}

var imagesListCmd = &cobra.Command{
	Use:   "ls",
	Short: "list images",
	Args:  cobra.ExactArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		const tmplStr = "NAME\tSIZE\n{{range .}}{{.Name}}\t{{.Size}}\n{{end}}"

		imgs, err := images.List()
		if err != nil {
			logrus.Fatal(err)
		}

		tmpl := template.Must(template.New("images list").Parse(tmplStr))

		writer := tabwriter.NewWriter(os.Stdout, 0, 2, 4, ' ', tabwriter.TabIndent)
		defer func() {
			_ = writer.Flush()
		}()

		if err := tmpl.Execute(writer, imgs); err != nil {
			logrus.Fatal(err)
		}
	},
}

var imageImportCmd = &cobra.Command{
	Use:   "import [image name] [image path]",
	Short: "import image",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		imageName, imagePath := args[0], args[1]

		if err := images.ImportImage(imagePath, imageName); err != nil {
			logrus.Fatal(err)
		}
	},
}

var imageRemoveCmd = &cobra.Command{
	Use:   "rm [image name]",
	Short: "remove image by name",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		name := args[0]

		if err := images.Delete(name); err != nil {
			logrus.Fatal(err)
		}
	},
}
