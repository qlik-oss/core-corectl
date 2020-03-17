package printer

import (
	"fmt"
	"github.com/qlik-oss/corectl/pkg/dynconf"
	"github.com/qlik-oss/corectl/pkg/log"
	"os"
	"sort"

	"github.com/olekukonko/tablewriter"
)

// PrintContext prints all information in a context
func PrintContext(name string, handler *dynconf.ContextHandler) {
	if name == "" {
		name = handler.Current
		if name == "" {
			fmt.Println("No current context")
			return
		}
	}
	context := handler.Get(name)
	if context == nil {
		fmt.Printf("No context with name: '%s'\n", name)
		return
	}
	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Comment: %s\n", context.GetString("comment"))
	fmt.Printf("Server: %s\n", context.GetString("server"))
	fmt.Printf("Certificates: %s\n", context.GetString("certificates"))
	fmt.Println("Headers:")
	for k, v := range context.Headers() {
		fmt.Printf("    %s: %s\n", k, v)
	}
}

// PrintCurrentContext prints the current context
func PrintCurrentContext(name string) {
	if name == "" {
		fmt.Println("Context: <NONE>")
	} else {
		fmt.Printf("Context: %s\n", name)
	}
}

// PrintContexts prints a list of contexts to standard out
func PrintContexts(handler *dynconf.ContextHandler, mode log.PrintMode) {
	var sortedContextKeys []string
	for k := range handler.Contexts {
		sortedContextKeys = append(sortedContextKeys, k)
	}

	sort.Strings(sortedContextKeys)

	if mode.BashMode() {
		for _, v := range sortedContextKeys {
			PrintToBashComp(v)
		}
	} else {
		writer := tablewriter.NewWriter(os.Stdout)
		writer.SetAutoFormatHeaders(false)
		writer.SetRowLine(true)
		header := []string{"Name", "Server", "Current", "Comment"}
		writer.SetHeader(header)

		for _, k := range sortedContextKeys {
			context := handler.Get(k)
			row := []string{k, context.GetString("server"), "", context.GetString("comment")}
			if k == handler.Current {
				// In case we change header order
				for i, h := range header {
					if h == "Current" {
						row[i] = "*"
						break
					}
				}
			}
			writer.Append(row)
		}
		writer.Render()
	}
}
