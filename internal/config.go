package internal

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// ConnectionConfigEntry defines the content of a connection in either the project config yml file or a connections yml file.
type ConnectionConfigEntry struct {
	Type     string
	Username string
	Password string
	Path     string
	Settings map[string]string
}

// ConnectionsConfigFile defines the content of a connections yml file.
type ConnectionsConfigFile struct {
	Connections map[string]ConnectionConfigEntry
}

// ReadConnectionsFile reads the connections config file from the supplied path.
func ReadConnectionsFile(path string) ConnectionsConfigFile {

	var config ConnectionsConfigFile
	source, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Could not find connections file:", path)
		os.Exit(1)
	}

	err = yaml.Unmarshal(source, &config)
	if err != nil {
		FatalError(err)
	}
	return config
}
