# Star Trek card batches: brief for a cloud session

You add Star Trek cards to the cmd_and_ctrl catalog. Main set `trk` is issue #4973 (batches 1-30); Star Trek Commander `trc` is issue #4974 (batches 31-49). One batch = one PR, then stop.

## Where things are (this cloud session cannot reach Scryfall or wizards.com)
- This brief, the card data and the rules live on the branch `work/trek-cards` (never merge it, never PR it):
  `git fetch origin work/trek-cards` then `git show origin/work/trek-cards:trek/batches/NN.json` (NN = two-digit batch number). Each card has its full oracle data, any `related` tokens/parts, and its official `rulings` (follow them).
- Token cards from the sets: `trek/tokens.json`. Batch index: `trek/batches.txt`.
- Comprehensive Rules (effective 2026-09-25, the pinned edition): `trek/cr.txt`. Grep and quote it before citing any rule number; never cite from memory.

## Setup
- `git fetch origin && git switch -c feat/trek-batch-<N> origin/develop`.
- Read AGENTS.md (it is binding: commit and PR format, closure ratchet, snapshots additive) and the docs/adding-cards.md subsections you need.
- Go: run as `GOTOOLCHAIN=go1.27.0 go ...` from server/. Never `go test ./...` on the whole repo at once; run packages one at a time.
- Before each card, check it isn't already on origin/develop (grep the oracle_id under server/internal/cards/effects) or in an open PR (`gh pr list --state open --search "NAME"`); skip it if so.

## Merge-friendly (many batch PRs land in parallel)
- Any new shared helper goes in a new file for your batch, `server/internal/cards/effects/trek_b<N>_helpers.go`, never appended to a shared file. Grep origin/develop for an existing helper first and reuse it.
- Don't reformat or reorder code you didn't change.

## Goal: every card Full, no caveats
- Every card ships CompletenessFull with an engine-driven test, no caveats. Never stamp Full falsely. Stronger than printed is wrong; weaker is wrong too.
- Contained, additive engine work is fine and expected: if a card needs a small new seam, build it (data, not a func field reachable from Game; `server/internal/game/testdata/closure_fields.txt` census may only fall; snapshot changes additive, re-record with `go test ./internal/game/ -run TestSnapshotShapeIsRecorded -update-shape`).
- If a clause needs a large change (protocol + client, or an ADR), leave that card out, and list it in the PR body under `## Skipped` with the exact clause it cannot express. The owner gets a seam agent on it at once.

## Card rules
- One file per card, `server/internal/cards/effects/<snake_name>.go`, one init(). Opt-in by oracle_id (never scryfall_id). Double-faced / adventure / saga / class / station / other layouts: find a precedent on develop with grep.
- One printed triggered ability = one registered trigger. Last-known information when the source may have left (CR 608.2h). Intervening-"if" rechecks (CR 603.4). Read keywords and types through the layered view.
- "Whenever [a player] attacks [another player]" doesn't trigger on an attack at a planeswalker or battle (CR 508.3e).
- Negative tests use CastSpellParams{Strict: true} / strict activation; test refusals by illegal target, timing or a pool that can't pay.
- TestNoNewExactClonesInTheCatalog: if your card's body matches an existing card's, call or extract one shared helper. Never rewrite the same body in a different shape to dodge the gate.
- Token rows: match the printed token in `trek/tokens.json` exactly (name, types, P/T, colours, abilities). An unprinted token from card text is fine: add the row and list it in knownUnmatchedTokens (tokens_realdump_manual_test.go) with a comment naming the card. If an identical token row is already in another open PR, add the identical line.
- Effect keys only via `go test ./internal/cards/effects -run TestEveryPersistedEffectKeyResolves -args -update-effect-keys`. Roadmap: `go test ./internal/roadmap/ -args -update -update-closed`.
- Cards that read as a board wipe, tutor, ramp or draw engine declare their Purpose (ADR 0126).
- Test gotchas: g.MoveCardByID takes the game lock itself (deadlocks inside WithWriteLock); attack triggers need lockInAttacks(t, g) after DeclareAttacker; batched triggers settle after g.RunStateChecksForTest().

## Oracle fixtures (you cannot make them)
Do NOT try to generate oracle fixtures; the dump is not reachable. The owner's machine generates them before landing. So in `./internal/cards/coverage/` the only failures you may leave are the "missing/stale oracle fixture" ones for your own new oracle ids; everything else must pass. Put every oracle id your PR adds on one line in the PR body: `oracle-ids=<id>,<id>,...` (this matters when an id is set through a constant).

## Before the PR
Run, from server/, each as its own command: `go vet ./internal/cards/... ./internal/game/`, `go test ./internal/cards/effects/`, `go test ./internal/cards/coverage/` (see above), `go test ./internal/game/` only if you touched internal/game, `go test ./internal/roadmap/`, and gofmt -l on your files (must print nothing).

## PR
- Commit trailers exactly:
  Sprint: S79 — Card coverage: every Standard set, newest first
  Issue: #4973 (batches 1-30) or #4974 (batches 31-49)
- No AI attribution lines anywhere, and no "Generated by" footer in the PR body.
- `gh pr create --base develop`, title `feat(server): Star Trek cards, batch <N>` (31-49: `feat(server): Star Trek Commander cards, batch <N>`). Body: ## Summary, ## Sprint (S79 — Card coverage: every Standard set, newest first), ## Issues (`Refs #4973` or `Refs #4974` on its own line), ## Skipped (only if any), ## Test plan (checklist), then the `oracle-ids=` line. Do not merge.
- In the PR body's Summary also note any new engine seam, new token rows and new effect keys.
