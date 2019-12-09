## corectl context rm

Remove one or more contexts

### Synopsis

Remove one or more contexts

```
corectl context rm <context name>... [flags]
```

### Examples

```
corectl context rm local-engine
corectl context rm ctx1 ctx2
```

### Options

```
  -h, --help   help for rm
```

### Options inherited from parent commands

```
  -i, --allow-insecure           Enabling allow-insecure will make it possible to connect using self signed certs
  -a, --app string               Name or identifier of the app
      --certificates string      path/to/folder containing client.pem, client_key.pem and root.pem certificates
  -c, --config string            path/to/config.yml where parameters can be set instead of on the command line
      --context string           Name of the context used when connecting to Qlik Associative Engine
  -e, --engine string            URL to the Qlik Associative Engine (default "localhost:9076")
      --headers stringToString   Http headers to use when connecting to Qlik Associative Engine (default [])
      --insecure                 Enabling insecure will make it possible to connect using self signed certificates
      --json                     Returns output in JSON format if possible, disables verbose and traffic output
      --no-data                  Open app without data
  -t, --traffic                  Log JSON websocket traffic to stdout
      --ttl string               Qlik Associative Engine session time to live in seconds (default "0")
  -v, --verbose                  Log extra information
```

### SEE ALSO

* [corectl context](corectl_context.md)	 - Create, update and use contexts

