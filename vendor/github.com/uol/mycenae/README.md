Mycenae
===============

Mycenae is a timeseries database that uses cassandra and elasticsearch as data persistence.

To make its usage easier and faster, Mycenae implements parts of OpenTSDB API, for this reason it is compatible with Scolletor and Grafana.

### Development

```
$ cd "${GOPATH}/src"
$ git clone "http://github.com/uol/mycenae.git" "github.com/uol/mycenae"
$ cd "github.com/uol/mycenae"
$ make build test run
```
