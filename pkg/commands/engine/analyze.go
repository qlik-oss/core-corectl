package engine

import (
	"github.com/qlik-oss/corectl/pkg/boot"
	"strings"

	"github.com/pkg/browser"
	"github.com/qlik-oss/corectl/internal"
	"github.com/qlik-oss/corectl/pkg/log"
	"github.com/qlik-oss/corectl/printer"
	"github.com/spf13/cobra"
)

func CreategetAssociationsCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "assoc",
		Args:    cobra.ExactArgs(0),
		Aliases: []string{"associations"},
		Short:   "Print table associations",
		Long:    "Print table associations",
		Example: `corectl assoc
corectl associations`,

		Run: func(ccmd *cobra.Command, args []string) {
			comm := boot.NewCommunicator(ccmd)
			ctx, _, doc, _ := comm.OpenAppSocket(false)
			data := internal.GetModelMetadata(ctx, doc, comm.RestCaller(), false)
			printer.PrintAssociations(data)
		},
	}
}

func CreategetTablesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tables",
		Args:  cobra.ExactArgs(0),
		Short: "Print tables",
		Long:  "Print tables for the data model in an app",
		Example: `corectl tables
corectl tables --app=my-app.qvf`,

		Run: func(ccmd *cobra.Command, args []string) {
			comm := boot.NewCommunicator(ccmd)
			ctx, _, doc, _ := comm.OpenAppSocket(false)
			data := internal.GetModelMetadata(ctx, doc, comm.RestCaller(), false)
			printer.PrintTables(data)
		},
	}
}

func CreateGetMetaCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "meta",
		Args:  cobra.ExactArgs(0),
		Short: "Print tables, fields and associations",
		Long:  "Print tables, fields, associations along with metadata like memory consumption, field cardinality etc",
		Example: `corectl meta
corectl meta --app my-app.qvf`,

		Run: func(ccmd *cobra.Command, args []string) {
			comm := boot.NewCommunicator(ccmd)
			ctx, _, doc, params := comm.OpenAppSocket(false)
			data := internal.GetModelMetadata(ctx, doc, comm.RestCaller(), false)
			printer.PrintMetadata(data, params.PrintMode())
		},
	}
}

func CreateGetValuesCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "values <field name>",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: listFieldsForCompletion,
		Short:             "Print the top values of a field",
		Long:              "Print the top values for a specific field in your data model",
		Example:           "corectl values FIELD",

		Run: func(ccmd *cobra.Command, args []string) {
			ctx, _, doc, _ := boot.NewCommunicator(ccmd).OpenAppSocket(false)
			printer.PrintFieldValues(ctx, doc, args[0])
		},
	}
}

func CreateGetFieldsCommand() *cobra.Command {
	return WithLocalFlags(&cobra.Command{
		Use:     "fields",
		Args:    cobra.ExactArgs(0),
		Short:   "Print field list",
		Long:    "Print all the fields in an app, and for each field also some sample content, tags and and number of values",
		Example: "corectl fields",

		Run: func(ccmd *cobra.Command, args []string) {
			comm := boot.NewCommunicator(ccmd)
			ctx, _, doc, params := comm.OpenAppSocket(false)
			data := internal.GetModelMetadata(ctx, doc, comm.RestCaller(), false)
			printer.PrintFields(data, false, params.PrintMode())
		},
	}, "quiet")
}

func CreateGetKeysCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "keys",
		Args:    cobra.ExactArgs(0),
		Short:   "Print key-only field list",
		Long:    "Print a fields list containing key-only fields",
		Example: "corectl keys",

		Run: func(ccmd *cobra.Command, args []string) {
			comm := boot.NewCommunicator(ccmd)
			ctx, _, doc, params := comm.OpenAppSocket(false)
			data := internal.GetModelMetadata(ctx, doc, comm.RestCaller(), false)
			printer.PrintFields(data, true, params.PrintMode())
		},
	}
}

func CreateEvalCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "eval <measure 1> [<measure 2...>] by <dimension 1> [<dimension 2...]",
		Args:  cobra.MinimumNArgs(1),
		Short: "Evaluate a list of measures and dimensions",
		Long:  `Evaluate a list of measures and dimensions. To evaluate a measure for a specific dimension use the <measure> by <dimension> notation. If dimensions are omitted then the eval will be evaluated over all dimensions.`,
		Example: `corectl eval "Count(a)" // returns the number of values in field "a"
corectl eval "1+1" // returns the calculated value for 1+1
corectl eval "Avg(Sales)" by "Region" // returns the average of measure "Sales" for dimension "Region"
corectl eval by "Region" // Returns the values for dimension "Region"`,

		Run: func(ccmd *cobra.Command, args []string) {
			ctx, _, doc, _ := boot.NewCommunicator(ccmd).OpenAppSocket(false)
			internal.Eval(ctx, doc, args)
		},
	}
}
func CreateCatwalkCommand() *cobra.Command {
	return WithLocalFlags(&cobra.Command{
		Use:   "catwalk",
		Args:  cobra.ExactArgs(0),
		Short: "Open the specified app in catwalk",
		Long:  `Open the specified app in catwalk. If no app is specified the catwalk hub will be opened.`,
		Example: `corectl catwalk --app my-app.qvf
corectl catwalk --app my-app.qvf --catwalk-url http://localhost:8080`,

		Run: func(ccmd *cobra.Command, args []string) {
			comm := boot.NewCommunicator(ccmd)

			var appSpecified bool
			catwalkURL := comm.GetString("catwalk-url")
			engineURL := comm.WebSocketEngineURL()
			catwalkURL += "?engine_url=" + engineURL
			if appSpecified {
				if ok, err := boot.NewCommunicator(ccmd).AppExists(); !ok {
					log.Fatalln(err)
				}
			}

			if !strings.HasPrefix(catwalkURL, "www") && !strings.HasPrefix(catwalkURL, "https://") && !strings.HasPrefix(catwalkURL, "http://") {
				log.Fatalf("%s is not a valid url\nPlease provide a valid URL starting with 'https://', 'http://' or 'www'\n", catwalkURL)
			}

			err := browser.OpenURL(catwalkURL)
			if err != nil {
				log.Fatalf("could not open URL: %s\n", err)
			}
		},
	}, "catwalk-url")
}

func listFieldsForCompletion(ccmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	comm := boot.NewCommunicator(ccmd)
	ctx, _, doc, _ := comm.OpenAppSocket(false)
	data := internal.GetModelMetadata(ctx, doc, comm.RestCaller(), false)
	result := make([]string, 0)
	for _, item := range data.Fields {
		result = append(result, item.Name)
	}
	return result, cobra.ShellCompDirectiveNoFileComp
}
