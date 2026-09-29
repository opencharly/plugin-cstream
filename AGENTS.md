# AGENTS.md — plugin-cstream

Standalone plugin repo for the `cstream:` check verb (`verb:cstream`) — the probes
that can only be answered from the **stream** side of a cstream deployment. The
plugin is a Go module at `candy/plugin-cstream/` (module path
`github.com/opencharly/plugin-cstream/candy/plugin-cstream`). There is **no root
`charly.yml`**, so `charly box validate` at the repo root has no project manifest to
parse; the candy is validated from a staging project that carries a `discover:`
block.

This repo has **no `skill:` entity** in its candy manifest, so there is no dedicated
owning skill projected into the marketplace corpus. The gap is recorded against
`opencharly/opencharly#291` (the batch that authors missing `skill:` entities).

Canonical files:

- `candy/plugin-cstream/charly.yml` — the `plugin-cstream:` candy entity (`plugin:`
  block, `primary:`, the `plan:` check).
- `candy/plugin-cstream/provider.go` — the `Invoke` dispatch and the box-mode skip.
- `candy/plugin-cstream/methods.go` — the `status` / `login` / `frame` methods.
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
- There is **no root `charly.yml`**, so `charly box validate` at the repo root
  parses nothing and the org-wide candy gate
  (`opencharly/.github/.github/workflows/candy-validate.yml`) **skips cleanly**.
  The manifest is gated only from a consumer project that has this candy in scan
  range (the `check-cstream-pod` witness box), which is what catches a malformed
  `plugin:` block.
- The merge gate is the **org-wide** `charly/pr-validator` (required check
  `validate / validate`, defined in `opencharly/.github`); this repo has **no**
  per-repo candy gate.
- R10 consumer: the `check-cstream-pod` bed in `opencharly/distro-cachyos`, whose
  witness box composes this candy.

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
