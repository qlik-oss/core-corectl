package cmd

import (
	"fmt"

	"github.com/qlik-oss/corectl/internal"
	"github.com/qlik-oss/corectl/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setConnectionsCmd = &cobra.Command{
	Use:     "set <path-to-connections-file.yml>",
	Args:    cobra.ExactArgs(1),
	Short:   "Set or update the connections in the current app",
	Long:    "Set or update the connections in the current app",
	Example: "corectl connection set ./my-connections.yml",

	Run: func(ccmd *cobra.Command, args []string) {
		state := internal.PrepareEngineState(rootCtx, headers, certificates, true)
		separateConnectionsFile := args[0]
		if separateConnectionsFile == "" {
			internal.FatalError("no connections config file specified")
		}
		internal.SetupConnections(rootCtx, state.Doc, separateConnectionsFile)
		if !viper.GetBool("no-save") {
			internal.Save(rootCtx, state.Doc)
		}
	},
}

var removeConnectionCmd = &cobra.Command{
	Use:   "rm <connection-id>...",
	Args:  cobra.MinimumNArgs(1),
	Short: "Remove the specified connection(s)",
	Long:  "Remove one or many connections from the app",
	Example: `corectl connection rm
corectl connection rm ID-1
corectl connection rm ID-1 ID-2`,

	Run: func(ccmd *cobra.Command, args []string) {
		state := internal.PrepareEngineState(rootCtx, headers, certificates, false)
		for _, connection := range args {
			err := state.Doc.DeleteConnection(rootCtx, connection)
			if err != nil {
				internal.FatalErrorf("could not remove connection '%s': %s", connection, err)
			}
		}
		if !viper.GetBool("no-save") {
			internal.Save(rootCtx, state.Doc)
		}
	},
}

var listConnectionsCmd = &cobra.Command{
	Use:     "ls",
	Args:    cobra.ExactArgs(0),
	Short:   "Print a list of all connections in the current app",
	Long:    "Print a list of all connections in the current app",
	Example: "corectl connection ls",

	Run: func(ccmd *cobra.Command, args []string) {
		state := internal.PrepareEngineState(rootCtx, headers, certificates, false)
		connections, err := state.Doc.GetConnections(rootCtx)
		if err != nil {
			internal.FatalErrorf("could not retrieve list of connections: %s", err)
		}
		printer.PrintConnections(connections, viper.GetBool("bash"))
	},
}

var getConnectionCmd = &cobra.Command{
	Use:     "get <connection-id>",
	Args:    cobra.ExactArgs(1),
	Short:   "Show the properties for a specific connection",
	Long:    "Show the properties for a specific connection",
	Example: "corectl connection get CONNECTION-ID",

	Run: func(ccmd *cobra.Command, args []string) {
		state := internal.PrepareEngineState(rootCtx, headers, certificates, false)
		connection, err := state.Doc.GetConnection(rootCtx, args[0])
		if err != nil {
			fmt.Printf("%T\n", err)
			internal.FatalErrorf("could not retrieve connection '%s': %s", args[0], err)
		}
		printer.PrintConnection(connection)
	},
}

var connectionCmd = &cobra.Command{
	Use:   "connection",
	Short: "Explore and manage connections",
	Long:  "Explore and manage connections",
	Annotations: map[string]string{
		"command_category": "sub",
	},
}

func init() {
	connectionCmd.AddCommand(setConnectionsCmd, getConnectionCmd, listConnectionsCmd, removeConnectionCmd)
}
