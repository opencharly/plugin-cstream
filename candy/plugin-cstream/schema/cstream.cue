// The `cstream` verb's OWN CUE schema — the typed plugin_input for a `cstream:`
// check step. Single source: it generates ../params/cue_types_gen.go AND is
// served over Describe so the host validates every authored step against it.
//
// SELF-CONTAINED: every field is a bare primitive referencing NO base def, so it
// compiles standalone and splices onto the base.
//
// FIELD SPLIT: only cstream-EXCLUSIVE fields live here. The matchers
// exit_status/stdout/stderr and the general modifiers timeout/eventually stay on
// #Op and are read off the step by the runner — they are NOT reproduced here.
#CstreamInput: {
	// method — what to ask the stream.
	//
	// The set is deliberately small. Each method is here because it can FAIL on a
	// defect nothing else catches:
	//
	//   status  the gateway's own readiness view -- it reports ready only when a
	//           DRM render node exists AND the compositor has published a socket,
	//           which is strictly more than "the port answers"
	//   frame   a frame pulled back out THROUGH the negotiated WebRTC track, so
	//           encoding, negotiation and transport are all in the failure path.
	//           This is the method the verb exists for: wl: screenshot reads the
	//           compositor and so cannot fail on any of them.
	//   login   one PAM authentication through the real session leader
	//
	// Methods the architecture also lists (session-list, stats, volume, clipboard,
	// idle, input, logout) are NOT served yet: the control surface they read does
	// not exist in the streamer, and a method that cannot be exercised reads as
	// coverage while providing none.
	method: "status" | "frame" | "login"

	// user / password — `login` only. The password reaches the leader over stdin,
	// never argv: argv is world-readable through /proc/<pid>/cmdline for the life
	// of the process.
	user?:     string
	password?: string

	// expect — `login` only: whether this credential SHOULD be accepted.
	// A login probe that only ever checks the correct password passes just as
	// happily against a stack that accepts anything, so the rejecting case has to
	// be expressible.
	expect?: "accept" | "reject"

	// artifact — `frame` only: where to write the pulled frame.
	artifact?: string
	// artifact_min_bytes — a size floor. A uniform field compresses to almost
	// nothing, so this is the cheap content guard against a black or placeholder
	// frame passing as a real one.
	artifact_min_bytes?: int @go(ArtifactMinBytes,type=int)
}
