package engine

import (
	"context"
	"github.com/qlik-oss/corectl/pkg/boot"
	"github.com/qlik-oss/enigma-go"

	"github.com/qlik-oss/corectl/internal"
	"github.com/spf13/cobra"
)

func CreateUnbuildCmd() *cobra.Command {
	return withLocalFlags(&cobra.Command{
		Use:   "unbuild",
		Args:  cobra.ExactArgs(0),
		Short: "Split up an existing app into separate json and yaml files",
		Long: `Extracts generic objects, dimensions, measures, variables, reload script and connections from an app in an engine into separate json and yaml files.
In addition to the resources from the app a corectl.yml configuration file is generated that binds them all together.
Passwords in the connection definitions can not be exported from the app and hence need to be handled manually.
Generic Object trees (e.g. Qlik Sense sheets) are exported as a full property tree which means that child objects are found inside the parent´s json (the qChildren array).
`,
		Example: `corectl unbuild
corectl unbuild --app APP-ID`,
		Annotations: map[string]string{
			"command_category": "build",
			"x-qlik-stability": "experimental",
		},

		Run: func(ccmd *cobra.Command, args []string) {
			comm := boot.NewCommunicator(ccmd)
			comm.OverrideSetting("no-data", true)
			ctx, global, doc, params := comm.OpenAppSocket(false)
			outdir := params.GetString("dir")

			if outdir == DefaultUnbuildFolder {
				outdir = getDefaultOutDir(ctx, doc, params.App())
			}
			internal.Unbuild(ctx, doc, global, outdir)
		},
	}, "dir")
}
func getDefaultOutDir(ctx context.Context, doc *enigma.Doc, appName string) string {
	appLayout, _ := doc.GetAppLayout(ctx)
	var defaultFolder string
	if appLayout.Title != "" {
		defaultFolder = appLayout.Title
	} else if appName != "" {
		defaultFolder = appName
	}
	return internal.BuildRootFolderFromTitle(defaultFolder)
}
