package standard

import (
	"context"
	"fmt"
	"github.com/qlik-oss/corectl/pkg/boot"
	"github.com/qlik-oss/corectl/printer"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/google/go-github/v27/github"
	ver "github.com/hashicorp/go-version"
	"github.com/qlik-oss/corectl/pkg/log"
	"github.com/spf13/cobra"
)

func CreateVersionCommand(version, branch, commit string) *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Args:    cobra.ExactArgs(0),
		Short:   "Print the version of corectl",
		Example: "corectl version",
		Annotations: map[string]string{
			"command_category": "other",
		},

		Run: func(_ *cobra.Command, args []string) {

			if !strings.Contains(version, "dev") {
				checkLatestVersion(version)
			} else {
				fmt.Printf("version: %s\tbranch: %s\tcommit: %s\n", version, branch, commit)
			}
		},
	}
}

func CreateStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Args:  cobra.ExactArgs(0),
		Short: "Print status info about the connection to the engine and current app",
		Long:  "Print status info about the connection to the engine and current app, and also the status of the data model",
		Example: `corectl status
corectl status --app=my-app.qvf`,
		Annotations: map[string]string{
			"command_category": "other",
		},

		Run: func(ccmd *cobra.Command, args []string) {
			comm := boot.NewCommunicator(ccmd)

			if comm.IsSenseForKubernetes() {
				err := comm.RestCaller().CallStd("GET", "v1/tenants/me", "", nil, nil, nil)
				if err != nil {
					log.Fatal(err)
				}
				if comm.AppId() != "" {
					ctx, _, doc, params := comm.OpenAppSocket(false)
					printer.PrintStatus(ctx, doc, params.WebSocketEngineURL(), params.AppId())
				} else {
					printer.PrintStatusRest(comm.RestBaseUrl().String())
				}
			} else {
				if comm.AppId() != "" {
					ctx, _, doc, params := comm.OpenAppSocket(false)
					printer.PrintStatus(ctx, doc, params.WebSocketEngineURL(), params.AppId())
				} else {
					ctx, _, params := comm.OpenGlobalSocket()
					printer.PrintStatus(ctx, nil, params.WebSocketEngineURL(), "")
				}
			}
		},
	}
}

// Function for checking current version against latest released version on github
func checkLatestVersion(version string) {
	client := github.NewClient(nil)
	rel, _, err := client.Repositories.GetLatestRelease(context.Background(), "qlik-oss", "corectl")
	if err != nil {
		// If we cannot connect to github just print the version
		fmt.Printf("corectl version: %s\n", version)
		return
	}
	latestVersion, outdated := isLatestVersion(version, *rel.TagName)
	if outdated {
		// Find absolute path of executable
		executable, _ := os.Executable()
		fmt.Println("--------------------------------------------------")
		fmt.Printf("corectl version: %s, latest version is %s\n", version, latestVersion)
		switch runtime.GOOS {
		case "darwin":
			fmt.Println("To update to the latest version using brew just run:")
			fmt.Print("\n  brew upgrade qlik-corectl\n\n")
			fmt.Println("If you don't already have the qlik-oss tap you have to first add the tap with:")
			fmt.Print("\n  brew tap qlik-oss/taps\n\n")
			fmt.Println("And after that, you have to install it with brew by running:")
			fmt.Print("\n  brew install qlik-corectl\n\n")
			fmt.Println("If you prefer curl, you can run:")
		case "linux":
			fmt.Println("To update to the latest version using snap just run:")
			fmt.Print("\n  snap refresh qlik-corectl\n\n")
			fmt.Println("If you prefer curl, you can run:")
		default:
			fmt.Println("To download the latest version using curl you can run:")
		}
		// Format a download string depending on OS
		var dwnl string
		if runtime.GOOS == "windows" {
			dwnl = fmt.Sprintf(`curl --silent --location "https://github.com/qlik-oss/corectl/releases/download/v%s/corectl-windows-x86_64.zip" > corectl.zip && unzip ./corectl.zip -d "%s" && rm ./corectl.zip`, latestVersion, path.Dir(executable))
		} else {
			dwnl = fmt.Sprintf(`curl --silent --location "https://github.com/qlik-oss/corectl/releases/download/v%s/corectl-%s-x86_64.tar.gz" | tar xz -C /tmp && mv /tmp/corectl %s`, latestVersion, runtime.GOOS, path.Dir(executable))
		}
		fmt.Printf("\n  %s\n\n", dwnl)
		fmt.Println("If you have any problems, questions or feedback you can find us on:")
		fmt.Println("  slack:\thttps://qlikbranch-slack-invite.herokuapp.com/")
		fmt.Println("  github:\thttps://github.com/qlik-oss/corectl")
		fmt.Println("---------------------------------------------------")
	} else {
		fmt.Printf("corectl version: %s\n", version)
	}
}

func isLatestVersion(currentTag string, latestTag string) (string, bool) {
	currentVersion, err := ver.NewVersion(currentTag)
	if err != nil {
		log.Fatalf("Current version is not semantically versioned: %s\n", currentTag)
	}

	latestVersion, err := ver.NewVersion(latestTag[1:]) // Remove 'v' from the tag
	if err != nil {
		log.Fatalf("Latest version is not semantically versioned: %s\n", latestVersion)
	}

	if currentVersion.LessThan(latestVersion) {
		return latestVersion.String(), true
	}
	return "", false
}
