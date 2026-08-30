package cstream

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/opencharly/plugin-cstream/candy/plugin-cstream/params"
	"github.com/opencharly/sdk"
	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// gatewayPort is the gateway's in-venue HTTP port. Resolved to a host-reachable
// address through the generic endpoint reverse-leg, never assumed to be published.
const gatewayPort = 8080

func dispatch(ctx context.Context, req *pb.InvokeRequest, op *spec.Op, in *params.CstreamInput, env *cstreamEnv) (string, error) {
	switch in.Method {
	case "status":
		return statusMethod(ctx, req, env)
	case "login":
		return loginMethod(ctx, req, in)
	case "frame":
		return frameMethod(ctx, req, in)
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

// loginMethod authenticates one credential through the REAL session leader.
//
// `expect` is not optional decoration. A login probe that only ever checks the
// correct password passes just as happily against a stack that accepts anything,
// so the rejecting case has to be assertable — and the two legs are only
// meaningful together, against the same stack and the same account.
//
// ⚠️ The password is NOT private on this path. The executor exposes only
// VenueCapture(ctx, cmd string), so the credential travels inside a command
// string and ends up in `sh -c` argv — world-readable via /proc/<pid>/cmdline.
// That is why the schema restricts this method to fixture credentials.
//
// The leader itself reads its credential from STDIN precisely to avoid this, and
// the production route (broker execs leader over a socketpair) keeps that
// property. This verb cannot, and says so rather than implying otherwise.
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

// frameVenuePath is where the probe writes the decoded frame inside the venue,
// before it is pulled back over the reverse channel. A fixed path is fine: the
// probe overwrites it and the file is removed after each pull.
const frameVenuePath = "/tmp/cstream-frame.jpg"

// frameMethod pulls one frame back THROUGH the negotiated WebRTC track.
//
// This is the method the verb exists for. Everything else that looks at pixels
// in this stack looks at the WRONG END of the pipeline:
//
//   - `wl: screenshot` reads the COMPOSITOR, so it cannot fail on an encoding or
//     transport defect at all;
//   - `stream-probe` reads the tap on the producer's OWN pipeline, so it proves
//     frames are produced and says nothing about any of them leaving the process.
//
// The frame returned here was encoded, negotiated, sent, received and decoded by
// a real consumer, so it fails on a defect in any of those. The defect that made
// this method unshippable until now is on record: gst-plugins-bad was missing, so
// webrtcbin did not exist and webrtcsink failed to build its session pipeline with
// a WARN -- sessions negotiated and no media followed, while every producer-side
// check passed.
//
// The bytes are written to in.Artifact (a HOST path) BEFORE the provider's
// artifact validators run, so artifact_min_bytes / artifact_not_uniform gate what
// a consumer actually received.
func frameMethod(ctx context.Context, req *pb.InvokeRequest, in *params.CstreamInput) (string, error) {
	if in.Artifact == "" {
		// Without a file to validate this method would report success for a stream
		// that sent nothing -- exactly the failure it exists to catch.
		return "", fmt.Errorf("cstream: frame needs an artifact path to write the decoded frame to")
	}
	ex, err := sdk.ExecutorFromInvoke(req.GetExecutorBrokerId())
	if err != nil {
		return "", fmt.Errorf("cstream: frame has no host executor attached (%w)", err)
	}
	out, err := ex.VenueCapture(ctx, "/usr/local/bin/cstream-frame-probe "+shellQuote(frameVenuePath))
	if err != nil {
		// The probe's own transcript names the cause (missing webrtcbin, no producer,
		// nothing received) -- surface it rather than a bare exit status.
		return out, fmt.Errorf("cstream: frame: %w", err)
	}
	data, err := ex.GetFile(ctx, frameVenuePath, false)
	if err != nil {
		return out, fmt.Errorf("cstream: frame: pulling %s: %w", frameVenuePath, err)
	}
	if err := os.WriteFile(in.Artifact, data, 0o644); err != nil {
		return out, fmt.Errorf("cstream: frame: writing %s: %w", in.Artifact, err)
	}
	_ = ex.VenueRunSilent(ctx, "rm -f "+shellQuote(frameVenuePath))
	return fmt.Sprintf("%s\ncstream: frame saved to %s (%d bytes)", out, in.Artifact, len(data)), nil
}
