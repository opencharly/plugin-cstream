// Command serve is the OUT-OF-PROCESS entrypoint for the cstream verb plugin: a
// thin shim serving the importable provider over go-plugin gRPC via sdk.Serve.
package main

import (
	cstream "github.com/opencharly/plugin-cstream/candy/plugin-cstream"
	"github.com/opencharly/sdk"
)

func main() { sdk.Serve(cstream.NewProvider(), cstream.NewMeta()) }
