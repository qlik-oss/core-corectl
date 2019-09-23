## corectl app rm

Remove the specified app

### Synopsis

Remove the specified app

```
corectl app rm <app-id> [flags]
```

### Examples

```
corectl app rm APP-ID
```

### Options

```
  -h, --help       help for rm
      --suppress   Suppress confirmation dialogue
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

* [corectl app](corectl_app.md)	 - Explore and manage apps

