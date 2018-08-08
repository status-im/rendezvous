Rendezvous server
=================

In order to build a docker image, run:

```bash
make image
```

Server usage:

```
  -a, --address string     listener ip address (default "0.0.0.0")
  -d, --data string        path where ENR infos will be stored. (default "/tmp/rendevouz")
  -g, --generate           dump private key and exit.
  -h, --keyhex string      private key hex
  -k, --keypath string     path to load private key
  -p, --port int           listener port (default 9090)
  -v, --verbosity string   verbosity level, options: crit, error, warning, info, debug (default "info")
```

Option `-g` can be used to generate hex of the private key for convenience.
Option `-h` should be used only in tests.

The only mandatory parameter is keypath `-k`, and not mandatory but i suggest to change data path `-d` not to a temporary
directory.
