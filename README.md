# plugin-cstream

Out-of-tree charly plugin serving the `cstream` check verb — the probes that can only be answered
from the **stream** side of a cstream deployment.

**Why not methods on `wl:`.** `wl: screenshot` reads the compositor, so it passes cleanly while the
encoder is misconfigured or the transport is dead. `cstream: frame` pulls a frame back out through
the negotiated WebRTC track, so encoding, negotiation and transport are all in its failure path —
it can fail on the bug it exists to catch, which is the whole point of a probe.

Three methods (`status`, `frame`, `login`), deliberately not the eleven the architecture eventually
wants: the rest need a streamer control surface that does not exist yet, and a method that cannot
be exercised reads as coverage while providing none.
