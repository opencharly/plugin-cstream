# plugin-cstream

The out-of-tree charly plugin serving the **`cstream:` check verb** — the probes that
can only be answered from the **stream** side of a cstream deployment.

## Why a verb of its own

`screenshot` already exists on four verbs (`cdp`, `wl`, `vnc`, `spice`), each
asserting its own protocol's view of the pixels. That duplication is deliberate: a
probe is worth having only if it can **fail on the defect it exists to catch**.
`wl: screenshot` reads the **compositor**, so it passes cleanly while the encoder is
misconfigured or the transport is dead. `cstream: frame` pulls a frame back out
**through the negotiated WebRTC track**, so encoding, negotiation and transport are
all in its failure path.

## Methods

| Method | What it asserts |
|---|---|
| `status` | the gateway's **own readiness view** — ready only when a DRM render node exists **and** the compositor has published a socket. Strictly more than "something answered on 8080". |
| `login` | one PAM authentication through the **real session leader**. `expect: accept\|reject` makes the refusing case assertable, so a stack that accepts anything cannot pass. |
| `frame` | a frame pulled back **through the negotiated WebRTC track** and written to a host `artifact:`, gated by `artifact_min_bytes` / `artifact_not_uniform`. |

Every method needs a **running deployment**; under `charly check box` (no running
deployment) each reports `skip` rather than fail.

## How to use it

It is an **external (out-of-tree) plugin**: projects compose it via the
`@github.com/opencharly/plugin-cstream/candy/plugin-cstream:<ref>` candy ref and
charly connects it out-of-process by word at runtime.

```yaml
- '@github.com/opencharly/plugin-cstream/candy/plugin-cstream:<tag>'
```

Then author the verb in a plan:

```yaml
- check: the gateway reports its own readiness
  id: cstream-status
  context: [runtime]
  cstream: status
  stdout:
      - contains: '"ready":true'

- check: a frame pulled back through the negotiated WebRTC track has real content
  id: cstream-frame
  context: [runtime]
  cstream:
      method: frame
      artifact: /tmp/cstream-frame.jpg
      artifact_min_bytes: 5000
      artifact_not_uniform: true
```

The scalar form (`cstream: status`) works because the candy manifest declares
`primary: {cstream: method}` — charly's pre-parse reads it off the raw YAML, so a
project that does not have this candy in scan range fails with
`plugin verb "cstream" takes a MAP input`.

## ⚠️ The `login` password is not private

The venue executor exposes only `VenueCapture(ctx, cmd string)` — there is no stdin
channel to an out-of-process verb — so a password reaches the venue inside a command
string and lands in `sh -c` argv, world-readable through `/proc/<pid>/cmdline`.
`cstream: login` is therefore for **fixture** credentials: an account a disposable
bed created in order to be authenticated. Do not point it at a real one. The
production path (the broker execs the leader over a socketpair; the leader reads the
credential from STDIN) does not have this property and must not grow it.

## Layout

- `candy/plugin-cstream/charly.yml` — the `plugin-cstream:` candy entity (`plugin:`
  block, `primary:`, the `plan:` check).
- `candy/plugin-cstream/plugin.go` — `NewProvider` / `NewMeta` and the embedded schema.
- `candy/plugin-cstream/provider.go` — the `Invoke` dispatch and the box-mode skip.
- `candy/plugin-cstream/methods.go` — the `status` / `login` / `frame` methods.
- `candy/plugin-cstream/support.go` — the bounded `httpGet` helper.
- `candy/plugin-cstream/schema/cstream.cue` — the self-contained `#CstreamInput`.
- `candy/plugin-cstream/params/cue_types_gen.go` — the generated params struct.
- `candy/plugin-cstream/cmd/serve/main.go` — the out-of-process entrypoint.
- `candy/plugin-cstream/schema_contract_test.go` — the CUE↔Go drift tests.
- `CHANGELOG/` — per-CalVer release history.
- `.github/workflows/tag-on-merge.yml` — CalVer tag + `CHANGELOG/` on merge.

## Related

- Owning skill: none yet — `/charly-internals:plugin` is the authoring reference for
  the plugin surface, and `/charly-check:check` the check orchestrator. The missing
  `skill:` entity is recorded against
  [opencharly/opencharly#291](https://github.com/opencharly/opencharly/issues/291).
- [`pod-cstream`](https://github.com/opencharly/pod-cstream) — the cstream transport
  spine (the gateway, session leader and streamer) this verb probes.
- [`layer-cstream-desktop`](https://github.com/opencharly/layer-cstream-desktop) —
  the nested desktop the streamer carries.
- [`distro-cachyos`](https://github.com/opencharly/distro-cachyos) — hosts the
  `check-cstream-pod` bed that composes this plugin.
- [`opencharly/opencharly`](https://github.com/opencharly/opencharly) — the umbrella.
