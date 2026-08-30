// Package cstream is the charly plugin serving the `cstream` check verb — the
// probes that can only be answered from the STREAM side of a cstream deployment.
//
// Why a verb of its own rather than methods on `wl:`. `screenshot` already exists
// on four verbs (cdp, wl, vnc, spice), each asserting its own protocol's view of
// the pixels, and that duplication is deliberate: a probe is only worth having if
// it can fail on the defect it exists to catch. `wl: screenshot` reads the
// COMPOSITOR, so it passes cleanly while the encoder is misconfigured or the
// transport is dead. `cstream: frame` pulls a frame back out THROUGH the
// negotiated WebRTC track, so encoding, negotiation and transport are all in its
// failure path. Folding it into `wl:` would produce a probe that cannot fail on
// the bug it was written for.
//
// Scope is deliberately narrow. Only methods whose result the deployment can
// actually produce today are served; a method that could only ever return a
// duplicate of the `http:` verb, or that has no implementation behind it, is
// worse than absent because it reads as coverage.
//
// Dual-placement by construction: the SAME NewProvider()/NewMeta() compile INTO
// charly in-process when listed in compiled_plugins, or cmd/serve serves them
// OUT-OF-PROCESS over go-plugin gRPC when they are not.
package cstream

import (
	"embed"

	"github.com/opencharly/sdk"
	pb "github.com/opencharly/spec/proto"
)

//go:embed schema/*.cue
var schemaFS embed.FS

// NewProvider returns the cstream verb provider (the Invoke dispatch surface).
func NewProvider() pb.ProviderServer { return &provider{} }

// NewMeta advertises verb:cstream plus the plugin's self-contained CUE schema.
func NewMeta() pb.PluginMetaServer {
	return sdk.NewMeta("2026.242.0800",
		[]sdk.ProvidedCapability{{Class: "verb", Word: "cstream", InputDef: "#CstreamInput"}},
		schemaFS)
}
