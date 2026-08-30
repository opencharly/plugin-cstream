package cstream

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/opencharly/plugin-cstream/candy/plugin-cstream/params"
	"github.com/opencharly/sdk"
	"github.com/opencharly/sdk/kit"
	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// cstreamEnv is the plugin-side decode of the CheckEnv the host ships as
// Operation.Env for a `cstream:` step.
type cstreamEnv struct {
	Box  string `json:"box"`
	Mode string `json:"mode"` // "live" | "box"
}

type provider struct{ pb.UnimplementedProviderServer }

// Invoke runs one `cstream:` operation.
//
// It uses BOTH reverse legs, because the two halves of a cstream deployment are
// reachable in different ways and neither substitutes for the other:
//
//   - ResolveEndpoint for the gateway's HTTP surface (the host owns the venue /
//     port-mapping machinery this out-of-process plugin cannot reach)
//   - ExecutorFromInvoke for anything that has to run INSIDE the venue — pulling a
//     frame back through the WebRTC track, and authenticating through the leader
func (provider) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeReply, error) {
	var op spec.Op
	if len(req.GetParamsJson()) > 0 {
		if err := json.Unmarshal(req.GetParamsJson(), &op); err != nil {
			return sdk.ResultJSON("fail", "cstream: decode op: "+err.Error())
		}
	}
	var in params.CstreamInput
	kit.DecodeInput(op.PluginInput, &in)
	var env cstreamEnv
	if len(req.GetEnvJson()) > 0 {
		_ = json.Unmarshal(req.GetEnvJson(), &env)
	}
	method := in.Method

	// Every method here needs a RUNNING deployment: a stream to pull from, a
	// gateway to answer, a PAM stack to authenticate against. Under `charly check
	// box` there is none, so skip rather than fail — the same contract the other
	// live-deployment verbs keep.
	if env.Mode == "box" {
		return sdk.ResultJSON("skip", fmt.Sprintf(
			"cstream: %s requires a running deployment (skip under charly check box)", method))
	}

	out, runErr := dispatch(ctx, req, &op, &in, &env)

	// `frame` is the one artifact-producing method, so it is the only one whose
	// artifact validators (min bytes / not uniform) run.
	return sdk.VerbVerdict("cstream", method, out, runErr, &op, method == "frame")
}
