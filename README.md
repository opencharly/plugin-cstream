# plugin-cstream

Out-of-tree charly plugin serving the `cstream` check verb — the probes that can only be answered
from the **stream** side of a cstream deployment.

**Why not methods on `wl:`.** `wl: screenshot` reads the compositor, so it passes cleanly while the
encoder is misconfigured or the transport is dead. `cstream: frame` pulls a frame back out through
the negotiated WebRTC track, so encoding, negotiation and transport are all in its failure path —
it can fail on the bug it exists to catch, which is the whole point of a probe.

**Two methods are served: `status` and `login`.**

`frame` is **withheld**, not implemented-and-broken. Signalling negotiates a session end to end and
no media follows (WP3d, cause not yet understood), so its gate cannot pass — and a method whose
gate cannot pass reads as coverage while providing none. The paragraph above says what it is *for*;
it arrives when it can be proven.

The architecture also lists `session-list`, `stats`, `volume`, `clipboard`, `idle`, `input` and
`logout`. None is served: the streamer has no control surface behind them.
