package cstream

import (
	"context"
	"fmt"
	"strings"

	"github.com/opencharly/plugin-cstream/candy/plugin-cstream/params"
	"github.com/opencharly/sdk"
	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// gatewayPort is the gateway's in-venue HTTP port. Resolved to a host-reachable
// address through the generic endpoint reverse-leg, never assumed to be published.
const gatewayPort = 8080

// framePath is where the in-venue probe writes the frame it pulls off the track,
// before it is fetched back over the reverse channel to the authored artifact path.
const framePath = "/tmp/cstream-consumed.jpg"

func dispatch(ctx context.Context, req *pb.InvokeRequest, op *spec.Op, in *params.CstreamInput, env *cstreamEnv) (string, error) {
	switch in.Method {
	case "status":
		return statusMethod(ctx, req, env)
	case "frame":
		return frameMethod(ctx, req, in)
	case "login":
		return loginMethod(ctx, req, in)
	default:
		return "", fmt.Errorf("cstream: unknown method %q", in.Method)
	}
}

// statusMethod reports the GATEWAY's own readiness view.
//
// Deliberately not a bare port probe: the gateway reports ready only when a DRM
// render node exists AND the compositor has published a Wayland socket, which is
// strictly more than "something answered on 8080". A port check would pass on a
// gateway that is up and has nothing to stream.
func statusMethod(ctx context.Context, req *pb.InvokeRequest, env *cstreamEnv) (string, error) {
	cc, err := sdk.NewCheckContext(req.GetExecutorBrokerId(), req.GetEnvJson())
	if err != nil {
		return "", fmt.Errorf("cstream: status: %w", err)
	}
	addr, err := cc.ResolveEndpoint(ctx, gatewayPort)
	if err != nil {
		return "", fmt.Errorf("cstream: status: %w", err)
	}
	if addr == "" {
		return "", fmt.Errorf("cstream: status has no resolved gateway endpoint (box=%q)", env.Box)
	}
	body, err := httpGet(ctx, "http://"+addr+"/healthz")
	if err != nil {
		return "", fmt.Errorf("cstream: status: %w", err)
	}
	// Surfaced verbatim so a step can assert on `"ready":true` and on WHICH render
	// node and display the gateway actually observed -- the fields that say why it
	// is not ready when it is not.
	return body, nil
}

// frameMethod pulls a frame back out THROUGH the negotiated WebRTC track.
//
// This is the method the verb exists for. wl: screenshot reads the compositor, so
// it passes while the encoder is misconfigured or the transport is dead; this
// negotiates as a real consumer, so encoding, negotiation and transport are all
// in its failure path.
func frameMethod(ctx context.Context, req *pb.InvokeRequest, in *params.CstreamInput) (string, error) {
	ex, err := sdk.ExecutorFromInvoke(req.GetExecutorBrokerId())
	if err != nil {
		return "", fmt.Errorf("cstream: frame has no host executor attached — it must run inside the venue (%w)", err)
	}
	out, err := ex.VenueCapture(ctx, fmt.Sprintf("/usr/local/bin/cstream-frame-probe %s", framePath))
	if err != nil {
		return out, fmt.Errorf("cstream: frame: %w", err)
	}
	if in.Artifact != "" {
		data, err := ex.GetFile(ctx, framePath, false)
		if err != nil {
			return out, fmt.Errorf("cstream: frame: fetching %s off the venue: %w", framePath, err)
		}
		if err := writeArtifact(in.Artifact, data); err != nil {
			return out, fmt.Errorf("cstream: frame: %w", err)
		}
	}
	return out, nil
}

// loginMethod authenticates one credential through the REAL session leader.
//
// `expect` is not optional decoration. A login probe that only ever checks the
// correct password passes just as happily against a stack that accepts anything,
// so the rejecting case has to be assertable — and the two legs are only
// meaningful together, against the same stack and the same account.
//
// The password goes over stdin, never argv: argv is world-readable through
// /proc/<pid>/cmdline for the whole life of the process.
func loginMethod(ctx context.Context, req *pb.InvokeRequest, in *params.CstreamInput) (string, error) {
	if in.User == "" {
		return "", fmt.Errorf("cstream: login needs a user")
	}
	expect := in.Expect
	if expect == "" {
		expect = "accept"
	}
	ex, err := sdk.ExecutorFromInvoke(req.GetExecutorBrokerId())
	if err != nil {
		return "", fmt.Errorf("cstream: login has no host executor attached (%w)", err)
	}
	cmd := fmt.Sprintf("/usr/local/bin/cstream-login-probe %s %s %s",
		shellQuote(in.User), shellQuote(expect), shellQuote(in.Password))
	out, err := ex.VenueCapture(ctx, cmd)
	if err != nil {
		return out, fmt.Errorf("cstream: login (%s) for %q: %w", expect, in.User, err)
	}
	return out, nil
}

// shellQuote renders a value as a single POSIX shell word. A password may contain
// anything at all, so it can never be interpolated bare.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
