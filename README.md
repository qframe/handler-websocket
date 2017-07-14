# handler-websocket
Handler that publishes messages using a websocket

### Create bindata

To compile the static assets into the binary I use `go-bindata-assetfs`.

```bash
$ go get github.com/jteeuwen/go-bindata/...
$ go get github.com/elazarl/go-bindata-assetfs/...
$ go-bindata-assetfs -pkg qhandler_websocket webroot/...
```
