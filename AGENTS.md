# AGENTS.md — plugin-cstream

Standalone plugin repo for the `cstream:` check verb (`verb:cstream`) — the probes
that can only be answered from the **stream** side of a cstream deployment. The
plugin is a Go module at `candy/plugin-cstream/` (module path
`github.com/opencharly/plugin-cstream/candy/plugin-cstream`). A root
`charly.yml` carries `discover: - path: candy`, so `charly box validate` at the repo
root scans every candy in `candy/` and gates all of them (opencharly/charly#459).

This repo has **no `skill:` entity** in its candy manifest, so there is no dedicated
owning skill projected into the marketplace corpus. The gap is recorded against
`opencharly/opencharly#291` (the batch that authors missing `skill:` entities).

Canonical files:

- `candy/plugin-cstream/charly.yml` — the `plugin-cstream:` candy entity (`plugin:`
  block, `primary:`, the `plan:` check).
- `candy/plugin-cstream/provider.go` — the `Invoke` dispatch and the box-mode skip.
- `candy/plugin-cstream/methods.go` holds the methods:
  - `status`, `login` and `frame`;
  - the input methods `type`, `key`, `move`, `click` and `scroll`, which drive the
    venue's `cstream-input`.
- `candy/cstream-probes/` — the verb's VENUE half: `cstream-login-probe`,
  `cstream-frame-probe` and `stream-probe`, installed by the `cstream-probes` candy so a
  method and the script it runs live in one repo (R2). `stream-probe` maps a window on
  the streamer's own compositor, so a bed may run it only as a `run:` step; it never
  belongs in a baked plan that `charly check live` runs against real deployments.
- `candy/cstream-probe-account/` — the BED-ONLY fixture giving the runtime account a
  known password for `cstream: login`.
- `candy/plugin-cstream/schema/cstream.cue` — the self-contained `#CstreamInput`.
- `candy/plugin-cstream/schema_contract_test.go` — the CUE↔Go drift tests.
- `.github/workflows/tag-on-merge.yml` — CalVer tag + `CHANGELOG/` on merge.
- `README.md` — user overview only; never agent guidance.

## Load these skills first (R0)

- `/charly-internals:plugin` — the plugin authoring reference: the `plugin:` block,
  the unified Provider model, the external (out-of-process) shape, the per-plugin
  CUE-schema contract, placement. Load before touching the provider or schema.
- `/charly-check:check` — the check orchestrator, the `check:` plan step, the
  runtime context, and the R10 sequence.
- `/charly-internals:git-workflow` — before any git/PR action.

## Build / validate / test

- `go build ./...`, `go vet ./...`, `go test ./...` in `candy/plugin-cstream/` —
  compile, vet, and run the `schema_contract_test.go` drift tests.
- `charly box validate` at the repo root, with the root `charly.yml`'s
  `discover: - path: candy`, gates the plugin, `cstream-probes` and
  `cstream-probe-account` together (3 candies; the CUE schema gate applies to
  `plugin:` blocks). Before that file existed the repo was project-less and the
  stock gate asserted nothing (opencharly/charly#459).
- The merge gate is the **org-wide** `charly/pr-validator` (required check
  `validate / validate`, defined in `opencharly/.github`); this repo has **no**
  per-repo candy gate.
- R10 consumer: the `check-kde-cstream-pod` and `check-kde-cstream-browser` beds in
  `opencharly/pod-kde-cstream`, whose box composes the plugin, `cstream-probes` and
  `cstream-probe-account`.

## Modify this repo

- Edit the `plugin-cstream:` candy entity, the Go source, and `schema/cstream.cue`
  **together** — the schema is the single source for the `params/` struct, so a
  field change not mirrored in the schema desyncs the generated types. Regenerate
  `params/cue_types_gen.go` through the documented pipeline; never hand-edit it.
- A method added to the enum must also be added to the dispatch switch, and the
  drift tests in `schema_contract_test.go` guard that (a method in one but missing
  from another fails at bed time).
- The plugin is **out-of-tree external** (connected out-of-process by word); do not
  describe it as compiled-in.
- Keep `primary:` in the manifest — the scalar form (`cstream: status`) only works
  because charly reads it off the raw YAML.

## Landing

- PR-only. Every change lands through a pull request; the org-required
  `charly/pr-validator` validates the diff and body and arms native auto-merge on
  PASS. Direct pushes to `main` are blocked.
- History lives in `CHANGELOG/` (written by `tag-on-merge` at merge time); the PR
  body IS the changelog.
- The authoritative rulebook is the umbrella `AGENTS.md` in
  `opencharly/opencharly` and `charly/AGENTS.md` in the charly repo. Do not
  restate its rules here.
