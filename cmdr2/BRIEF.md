# Commander deck cards, wave 2 (issue #5368): brief for a cloud session

You add cards from the Commander decks of Standard sets to the cmd_and_ctrl catalog: Murders at Karlov Manor Commander `mkc`, Outlaws of Thunder Junction Commander `otc`, Lost Caverns of Ixalan Commander `lcc`, Duskmourn Commander `dsc`, Bloomburrow Commander `blc`, Lorwyn Eclipsed Commander `ecc`, Wilds of Eldraine Commander `woc` (issue #5368). Most are reprints of older cards. Your prompt names the batch files (or one seam file) you build. You push ONE branch and STOP. You do NOT open a pull request: the owner rolls every branch of this wave into one PR with one CI run.

## Why the batches look the way they do
A planner grouped the cards by the code they share, so that no two workers write the same helper, token or engine change. Each batch file starts with `theme` and `reuse`: the helpers, tokens and precedent cards already on develop. Read `reuse` first and use what it names. Other workers are building other batches at the same moment, from the same base:
- Build ONLY the cards in your files. Never build a card from another batch, even if it would be convenient.
- If a card of yours needs a helper, token or engine change that `reuse` does not name and that is not on origin/develop, it is yours to write: the plan gave that need to your batch alone. Put card-side helpers in a NEW file `server/internal/cards/effects/cmdr2_b<N>_helpers.go` (N = your first batch), with names that start `cmdr2B<N>`. Never append to a shared file (helpers.go, triggers_common.go, primitives.go, activated.go) unless the change cannot live anywhere else.
- If `reuse` says "grep first" for a sentence another batch shares, run `git fetch origin` and `git grep` the phrase on `origin/feat/cmdr2-*` before writing it; if it is there, write the IDENTICAL helper name and body so the roll-up can keep one.
- If `reuse` names a watch item ("stop and report if it does not fit") and the precedent truly does not fit, leave that one card out and list it under `## Skipped`. Do not build a large engine change in a card batch.

## Where things are (this cloud session cannot reach Scryfall or wizards.com)
- This brief, the card data and the rules live on the branch `work/cmdr-cards` (never merge it, never PR it): `git fetch origin work/cmdr-cards`, then `git show origin/work/cmdr-cards:cmdr2/batches/NN.json` (two-digit batch number) or `cmdr2/seams/sN.json`. Each card has its oracle data and its `related` tokens and parts.
- Token cards of these sets: `cmdr2/tokens.json`. Index of every batch and seam: `cmdr2/batches.txt`. The whole plan: `cmdr2/PLAN.md`.
- Comprehensive Rules (effective 2026-09-25, the pinned edition): `cmdr/cr.txt` on the same branch. Grep and quote it before citing any rule number; never cite from memory.

## Setup
- `git fetch origin && git switch -c <the branch your prompt names> origin/develop`.
- Read AGENTS.md (binding: commit format, closure ratchet, snapshots additive) and only the docs/adding-cards.md subsections you need.
- Go: run as `GOTOOLCHAIN=go1.27.0 go ...` from server/. Never `go test ./...` on the whole repo; run packages one at a time. `go test ./internal/cards/effects/` takes long: pass `-timeout 40m`, and while iterating run only your tests with `-run`.
- Before each card, check it isn't already on origin/develop (grep the oracle_id under server/internal/cards/effects); skip it if so and say so in the final commit.
- Work one card at a time and commit after every two or three cards, so nothing is lost if the session ends early. Push the branch after your first batch file is done and again at the end.
- Don't reformat or reorder code you didn't change.

## Goal: every card Full, no caveats
- Every card ships CompletenessFull with an engine-driven test, no caveats. Never stamp Full falsely. Stronger than printed is wrong; weaker is wrong too.
- Contained, additive engine work is fine and expected: if a card needs a small new seam, build it (data, not a func field reachable from Game; `server/internal/game/testdata/closure_fields.txt` census may only fall; snapshot changes additive, re-record with `go test ./internal/game/ -run TestSnapshotShapeIsRecorded -update-shape`).
- If a clause needs a large change (protocol + client, or an ADR), leave that card out, and list it in your final commit message under `## Skipped` with the exact clause it cannot express. The owner gets a seam agent on it at once.

## Card rules
- One file per card, `server/internal/cards/effects/<snake_name>.go`, one init(). Opt-in by oracle_id (never scryfall_id). Double-faced / adventure / saga / class / station / other layouts: find a precedent on develop with grep.
- One printed triggered ability = one registered trigger. Last-known information when the source may have left (CR 608.2h). Intervening-"if" rechecks (CR 603.4). Read keywords and types through the layered view.
- "Whenever [a player] attacks [another player]" doesn't trigger on an attack at a planeswalker or battle (CR 508.3e).
- Negative tests use CastSpellParams{Strict: true} / strict activation; test refusals by illegal target, timing or a pool that can't pay.
- TestNoNewExactClonesInTheCatalog: if your card's body matches an existing card's, call or extract one shared helper. Never rewrite the same body in a different shape to dodge the gate.
- Token rows: match the printed token in `cmdr2/tokens.json` exactly (name, types, P/T, colours, abilities). An unprinted token from card text is fine: add the row and list it in knownUnmatchedTokens (tokens_realdump_manual_test.go) with a comment naming the card. 
- Effect keys only via `go test ./internal/cards/effects -run TestEveryPersistedEffectKeyResolves -args -update-effect-keys`. Roadmap: `go test ./internal/roadmap/ -args -update -update-closed`.
- Cards that read as a board wipe, tutor, ramp or draw engine declare their Purpose (ADR 0126).
- Test gotchas: g.MoveCardByID takes the game lock itself (deadlocks inside WithWriteLock); attack triggers need lockInAttacks(t, g) after DeclareAttacker; batched triggers settle after g.RunStateChecksForTest().

## Oracle fixtures (you cannot make them)
Do NOT try to generate oracle fixtures; the dump is not reachable. The owner's machine generates them before landing. So in `./internal/cards/coverage/` the only failures you may leave are the "missing/stale oracle fixture" ones for your own new oracle ids; everything else must pass. Put every oracle id your PR adds on one line in the PR body: `oracle-ids=<id>,<id>,...` (this matters when an id is set through a constant).

## Checks that the cloud cannot see

- Run `scripts/go-docker.sh test -timeout 40m ./internal/cards/coverage/` (or `go test` there) and make it pass, apart from missing oracle fixtures, which are generated after you push. In particular `TestOneSentenceTriggersDeclareFootprints`: a one-sentence trigger (draw, drain, damage, ...) must be written with `effects.Do` primitives that declare a footprint (`cards/effects/footprint.go`); never add the card to `testdata/unfootprinted_trigger_lines.txt`.
- Before adding a token row to `tokens_table.go`, check origin/develop and the other open `feat/cmdr2-*` branches for the same token and reuse its key.
- The footprint test skips a card whose oracle fixture is missing, so in the cloud it will pass even when your trigger is wrong. Check the rule yourself: every one-sentence trigger uses `effects.Do` primitives with a footprint (see `CreateTokenAdvanced`, `CounterOnThis` and the `FootprintCond*` list in `game/trigger_independence.go` for precedent).
- Before writing any shared helper (a keyword, a token, a bending action), grep origin/develop and the open `feat/cmdr2-*` branches for an existing one and reuse it. Two batches defining the same helper name breaks the build when the second lands.

## Added 2026-11-05: declare fixed-amount riders

`TestRiderPurposesMatchThePrintedCard` (server/internal/cards/coverage/rider_purpose_test.go, #4978) fails a card that prints a rider with a fixed amount (scry N, surveil N, discover N, amass N, learn, the Ring tempts you, and the others it lists) without declaring it. Declare it with `Purpose: game.Purpose{Scry: 1}` (or the matching field in server/internal/game/purpose.go) on the slot that does it: the Spec for a spell or an enters effect, otherwise the row. Grep develop for a card with the same rider (e.g. `Purpose: game.Purpose{Scry:`) and copy it. The test only sees a card once its oracle fixture exists, so check your cards against the printed text yourself.

## Before the final push
Run, from server/, each as its own command: `go vet ./internal/cards/... ./internal/game/`, `go test -timeout 40m ./internal/cards/effects/`, `go test -timeout 40m ./internal/cards/coverage/` (the only failures you may leave are "missing/stale oracle fixture" for your own new oracle ids: the owner's machine generates fixtures), `go test ./internal/game/` only if you touched internal/game, `go test ./internal/roadmap/`, `golangci-lint run` if it is installed, and `gofmt -l` on your files (must print nothing).

## Finish: push the branch, no PR
- Every commit ends with exactly these trailers (a seam worker uses its own seam issue number in place of 5368, as its prompt says):
  Sprint: S79 — Card coverage: every Standard set, newest first
  Issue: #5368
- No AI attribution lines and no "Generated by" footer anywhere.
- Your FINAL commit message body lists: the cards built; any new engine seam, token rows and effect keys; a `## Skipped` section (card, exact clause, why) only if any; and one line `oracle-ids=<id>,<id>,...` with every oracle id the branch adds.
- `git push -u origin <branch>`. Do NOT run `gh pr create`. Then stop.
