package printer

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/qlik-oss/corectl/pkg/dynconf"
	"github.com/qlik-oss/corectl/pkg/log"
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
	keys := make([]string, len(context))
	i := 0
	for k := range context {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		if k == "headers" {
			fmt.Println("Headers:")
			for k, v := range context.Headers() {
				fmt.Printf("    %s: %s\n", k, v)
			}
		} else {
			fmt.Printf("%s: %v\n", strings.Title(k), context[k])
		}
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
