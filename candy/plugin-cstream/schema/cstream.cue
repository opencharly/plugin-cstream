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
	//   login   one PAM authentication through the real session leader
	//
	// `frame` -- a frame pulled back THROUGH the negotiated WebRTC track -- is the
	// method this verb ultimately exists for, because wl: screenshot reads the
	// compositor and so cannot fail on an encoding or transport defect. It is NOT
	// offered yet: signalling negotiates a session end-to-end and no media follows
	// (WP3d, cause not yet understood). Offering a method whose gate cannot pass
	// would read as coverage while providing none.
	//
	// Methods the architecture also lists (session-list, stats, volume, clipboard,
	// idle, input, logout) are NOT served yet: the control surface they read does
	// not exist in the streamer, and a method that cannot be exercised reads as
	// coverage while providing none.
	method: "status" | "login"

	// user / password — `login` only.
	//
	// ⚠️ THIS PASSWORD IS NOT PRIVATE. The venue executor exposes only
	// VenueCapture(ctx, cmd string) — there is no stdin channel to an
	// out-of-process verb — so the value reaches the venue inside a command
	// string and lands in `sh -c` argv, which is world-readable through
	// /proc/<pid>/cmdline for the life of that exec.
	//
	// So `cstream: login` is for FIXTURE credentials: an account a disposable bed
	// created in order to be authenticated. Do not point it at a real one.
	//
	// The production path does not have this property and must not grow it: the
	// broker (root) execs the leader over a socketpair and the leader reads the
	// credential from STDIN, never argv. That discipline is deliberate and is
	// tested; this verb simply cannot reach it through the executor it is given.
	user?:     string
	password?: string

	// expect — `login` only: whether this credential SHOULD be accepted.
	// A login probe that only ever checks the correct password passes just as
	// happily against a stack that accepts anything, so the rejecting case has to
	// be expressible.
	expect?: "accept" | "reject"

}
