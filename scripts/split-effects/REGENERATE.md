# Regenerating the split of `internal/cards/effects` (#5906, ADR 0225)

The split PR (PR 2 of #5906) is generated, not written. It is regenerated on
the newest `develop` right before it merges, so no card that landed in the
meantime is left behind. This is the runbook for the cloud agent (or person)
who does that, normally right after a roll-up merges into `develop`. Budget:
**under an hour**, most of it CI. Card work does not stop for it: branches cut
before the split land afterwards with `scripts/split-effects.sh --place`.

Everything below runs from the repository root. With a local Go, use it with
`GOTOOLCHAIN=auto` (it fetches the go1.27 that `server/go.mod` names); without
one, prefix Go commands with `scripts/go-docker.sh go` (from `server/`) as
AGENTS.md §5 says. `scripts/split-effects.sh` picks whichever exists.

## 0. Before you start (2 minutes)

- PR 1 of #5906 (the tool, the guard, `shards.txt`, the fix-it check) is
  merged into `develop`. `scripts/split-effects/` and
  `server/internal/cards/effects/shards.txt` exist on `develop`.
- No other branch is mid-way through a split: `core_cards.txt` does not exist
  on `develop` (if it does, the split already landed; stop).
- The machine has about 4 GB free for the tool and 8 GB for the test run.

## 1. Regenerate (about 2 minutes)

```bash
git fetch origin develop feat/5906-split
git checkout -B feat/5906-split origin/develop
GOTOOLCHAIN=auto scripts/split-effects.sh 2>&1 | tee /tmp/split.log
```

That one command is the whole PR. In about 80 seconds of type-checking and
writing, plus a few minutes for the clone baseline, it:

1. places every file (core, card shards, `effectstest`, test packages) and
   moves it;
2. exports and renames every name that crosses a package (reviewed
   collisions come from `scripts/split-effects/overrides.txt`), keys the
   literals that cross, and qualifies every cross-package reference;
3. rewrites moved tests: relative paths, the source scanners' flat view,
   scanner literals that name a renamed identifier, `file.go:N` allowlist keys;
4. applies the fixups in `scripts/split-effects/fixups.go` (each fails loudly
   if its text moved);
5. generates the aggregator (`effects/cards/cards.go`), a `main_test.go` per
   test package, `core_cards.txt`, `scripts/split-effects/table.txt`, and a
   catalog link file in every main package and every package whose tests
   reach the core;
6. turns the completeness guard on (`catalogSharded = true`);
7. edits CI: `scripts/ci-test-shard.sh` pins `internal/cards/effects/...`,
   `ci-cd.yml` drops the effects `-cover` exception and points the token-dump
   job at the token test's package, and `e2e-nightly.yml` likewise;
8. regenerates the clone baseline (`go test ./internal/cards/coverage/ -run
   TestNoNewExactClonesInTheCatalog -update`): the renamed helpers change the
   hash of the bodies they appear in, not the groups.

**If it stops**, the message says where:

- `N collisions took an X name` (a warning, the run goes on): a new helper's
  upper-cased name is taken. Read both declarations, add a reviewed line to
  `overrides.txt`, commit it, and start step 1 again.
- `rename conflicts` / `unresolved placements`: a new shape the rules do not
  cover (for example a test function another test package calls). Fix the
  rule in the tool, or the code on `develop`, then start again.
- `fixup (…): expected the text below exactly once`: a test the fixup edits
  changed on `develop`. Update the entry in `fixups.go`.
- `replaceOnce` on a workflow file: CI changed around the edited lines.
  Update the strings in `run.go` (`afterSplit`).

A change to the tool is a commit on the PR branch's first commit's side: put
it in PR 2 (it is small next to the generated diff) and say so in the body.

## 2. Check (about 15 minutes on 4 cores)

```bash
cd server
export GOTOOLCHAIN=auto
go build ./... && go vet ./...                                       # about 4 minutes
/usr/bin/time -v go test -race -cover -count=1 ./internal/cards/effects/... 2>&1 | tee /tmp/effects-test.log   # 9-11 minutes
cd .. && scripts/split-effects.sh --check                            # prints nothing, exits 0
git diff origin/develop --stat -- server/internal/game/testdata | tail -1   # must be empty
```

CI runs the rest (every other package under `-race -cover`, golangci-lint) in
parallel shards, so do not wait for them here. With time to spare, the same
checks run locally: `go test -race -count=1 ./...` takes about 45 minutes on
4 cores, and `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
run ./...` about 9 minutes cold (build it with `GOTOOLCHAIN=go1.27.0`: the
linter's own module asks for an older Go, and a linter built with it refuses
the go1.27 module).

- Read the peak from `/tmp/effects-test.log` (`Maximum resident set size`,
  the largest single process) and update the table in the PR body.
- A failing test is never skipped or deleted, a snapshot fixture is never
  edited or regenerated, and the closure ratchet's counts
  (`internal/game/testdata/closure_fields.txt`) must not rise: the split
  touches none of them. If one fails, find which rule moved it (a new
  source-scanning test is the usual cause: add the flat-view call it makes
  to `flatCalls` in `split.go`, or a fixup).
- `go test ./...` on `develop` itself cannot finish on a 13 GB machine (the
  effects test build is what this PR fixes); compare with the last green CI
  run instead.

## 3. Commit, push, update the PR (5 minutes, then CI)

```bash
git add -A
git commit -F- <<'EOF'
refactor(server): split internal/cards/effects into core, card shards and test packages (#5906)

Generated by scripts/split-effects on develop at <sha> (REGENERATE.md).
No hand edits beyond scripts/split-effects itself.

Sprint: S82 — Card coverage: sets outside Standard, newest first
Issue: #5906
EOF
git push --force-with-lease -u origin feat/5906-split
```

Then update the PR body: the `develop` commit it was generated on, the peak
memory, the test counts. CI runs in about 30–40 minutes; the roll-up queue
waits for it, not longer. Merge as soon as it is green (the owner merges).

## 4. Right after it merges

- Tell the card workers (the Discord thread, and any agent brief): "after you
  rebase past the split, run the command `TestEffectsTreeLayout` prints:
  `scripts/split-effects.sh --place <your new files>`". The test's message
  says the same, so a worker that only reads CI is covered.
- A rebase that edits an existing card file may need
  `git -c merge.renameLimit=100000 rebase origin/develop` for git to follow
  the move.

## How long it takes

Measured on a 4-core, 15 GB cloud machine on develop at 2af192538
(2026-11-19): the tool 78 s (about 100 s with the clone baseline), `go build
./...` 1m12s, `go vet` 3 minutes, the effects tree's `-race -cover` tests
9-11 minutes, largest process 2.3-2.4 GiB. Steps 1 to 3 take about 20
minutes; with CI, a roll-up merge to a green PR is under an hour.

## Reviewing a regenerated PR

Do not read 38,000 files. Re-run step 1 on the same `develop` commit and diff
against the PR branch: it must be empty (the tool is deterministic: the same
input gives the same bytes). Then read the parts a person wrote:
`scripts/split-effects/` (the tool, `overrides.txt`, `fixups.go`) and the
handful of hand edits listed in the PR body, if any.
