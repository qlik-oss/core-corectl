## corectl connection set

Set or update the connections in the current app

### Synopsis

Set or update the connections in the current app

```
corectl connection set <path-to-connections-file.yml> [flags]
```

### Examples

```
corectl connection set ./my-connections.yml
```

### Options

```
  -h, --help   help for set
```

### Options inherited from parent commands

```
  -a, --app string               Name or identifier of the app
      --certificates string      path/to/folder containing client.pem, client_key.pem and root.pem certificates
  -c, --config string            path/to/config.yml where parameters can be set instead of on the command line
      --context string           Name of the context used when connecting to Qlik Associative Engine
  -e, --engine string            URL to the Qlik Associative Engine (default "localhost:9076")
      --headers stringToString   Http headers to use when connecting to Qlik Associative Engine (default [])
      --json                     Returns output in JSON format if possible, disables verbose and traffic output
      --no-data                  Open app without data
  -q, --quiet                    Restrict output to consist of IDs only. This can be useful for scripting.
  -t, --traffic                  Log JSON websocket traffic to stdout
      --ttl string               Qlik Associative Engine session time to live in seconds (default "0")
  -v, --verbose                  Log extra information
```

### SEE ALSO

* [corectl connection](corectl_connection.md)	 - Explore and manage connections

