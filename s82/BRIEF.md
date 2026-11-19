# S82 card batches: brief for a cloud worker (umbrella #5458)

You build ONE batch of cards from the ten sets that most recently left Standard (MAT, MOM, ONE, BRO, DMU, SNC, NEO, VOW, MID, AFR, with their Commander sets). Your prompt names the set and the batch number. You push a branch and stop; you do not open a PR.

## Reprint pass (set "RPR", issue #5908)
If your prompt says set RPR: the cards are reprints and bonus-sheet cards that the dev catalog counts in the S82 sets, Edge of Eternities, Star Trek and TMNT (most are older cards). Same rules as below. Branch `feat/s82-rpr-b<NN>`, commit subject `feat(server): RPR cards, batch <NN>`, trailer `Issue: #5908`. There are no per-batch planner notes in #5908; read each card from the dump. Many of these cards share mechanics with cards already on develop: grep for the mechanic and reuse.

## Cards missing from sets since AFR (set "GAP", issue #5948)
If your prompt says set GAP: the cards are ones the dev catalog counts in a set released since AFR (Commander Legends: Battle for Baldur's Gate, Double Masters 2022, Warhammer 40,000, Jumpstart 2022, Dominaria Remastered, Lord of the Rings and its Commander decks, Commander Masters, Doctor Who, Ravnica Remastered, Fallout, Modern Horizons 3, Assassin's Creed, Mystery Booster 2, Innistrad Remastered, Mystery Booster Commander and a few small sets) but does not have yet. Many are reprints of older cards. Same rules as below. Branch `feat/s82-gap-b<NN>`, commit subject `feat(server): GAP cards, batch <NN>`, trailer `Issue: #5948`. The plan row carries `oracle_ids` in the same order as `missing`: use them, since a few names have several printings. Read each card from the dump. Many share mechanics with cards already on develop (initiative, the Ring tempts you, squad, villainous choice, time travel, rad counters, energy, eldrazi spawn, freerunning, ...): grep for the mechanic and reuse. If a card is already on origin/develop, skip it and say so. Mystery Booster 2 playtest cards are real cards for this purpose; if one has a rule the engine can't express, skip it with the exact clause under `## Skipped`. Push your branch even if a check is failing or a card needs an owner decision: say so in the final commit message; a run that ends without a push loses all its work.

## Bonus sheets and child sets (set "BNS", issue #5929)
If your prompt says set BNS: the cards are bonus-sheet cards that real draft and play boosters open (Enchanting Tales, Breaking News, Special Guests, The List) and the cards of Ravnica: Clue Edition and Foundations Jumpstart, which the dev catalog counts under MKM and FDN. Most are reprints of older cards. Same rules as below. Branch `feat/s82-bns-b<NN>`, commit subject `feat(server): BNS cards, batch <NN>`, trailer `Issue: #5929`. The plan row also carries `oracle_ids` in the same order as `missing`: use them, since a few names have several printings. Read each card from the dump. Many share mechanics with cards already on develop: grep for the mechanic and reuse. If a card is already on origin/develop, skip it and say so.

## Do not stop early
Nobody answers you mid-run. Never end your turn while a test, build or lint is still running: wait in the foreground (raise the Bash timeout) or with an until-loop, and keep going until the final commit is pushed.

## Where things are (this session cannot reach Scryfall)
- The batch's card list: `git fetch origin work/s82-batches` then `git show FETCH_HEAD:plan/s82-batches.json`; pick the row with your `set` and `batch`. Build exactly its `missing` cards.
- The set issue (`issue` in that row) has the planner's notes for each batch: read your batch's section (`gh issue view <n>` or `curl -s https://api.github.com/repos/krakenhavoc/cmd_and_ctrl/issues/<n>`).
- Card data and the rules: `git fetch origin work/s82-cards`, then
  `mkdir -p data/scryfall data/cr && git show FETCH_HEAD:s82-cards.json.gz | gunzip > data/scryfall/default-cards.json && git show FETCH_HEAD:cr.txt > data/cr/cr.txt`.
  The dump has every printing of your cards and their tokens (oracle text, faces, all_parts). Read each card's full object from it (python `json.load`, then filter by name). `data/` is gitignored: never commit those files.
- Comprehensive Rules (pinned 2026-09-25 edition): `data/cr/cr.txt`. Grep and quote it before citing any rule number; never cite from memory.

## Setup
- `git fetch origin && git switch -c feat/s82-<set>-b<NN> origin/develop` (NN two digits, e.g. `feat/s82-mom-b03`).
- Read AGENTS.md (binding: commit format, closure ratchet, snapshots additive) and the docs/adding-cards.md subsections you need.
- Go: from server/, `GOTOOLCHAIN=go1.27.0 GOGC=30 go ... -p 1`, one package at a time. No docker. Never `go test ./...`.
- Before each card, check it isn't already on origin/develop (grep its oracle_id under server/internal/cards) or on another `feat/s82-*` branch (`git ls-remote origin 'refs/heads/feat/s82-*'`, then grep that branch); skip it if so.

## Merge-friendly (many batches land in parallel)
- A new shared helper goes in a new file for your batch, `server/internal/cards/effects/s82_<set>_b<NN>_helpers.go`, never appended to a shared file. Grep origin/develop and the other `feat/s82-*` branches first and reuse.
- Don't reformat or reorder code you didn't change.

## Goal: every card Full, no caveats
- Every card ships CompletenessFull with an engine-driven test, no caveats. Never stamp Full falsely. Stronger than printed is wrong; weaker is wrong too.
- Each set's mechanics (kicker, domain, enlist, read ahead, prototype, offering, blitz, casualty, ninjutsu, disturb, daybound/nightbound, training, exploit, soulbond, dungeons, dice...) are on develop: grep and reuse. The S82 seams #5644–#5683 are all on develop except #5673 (can't crew Vehicles).
- Contained, additive engine work is fine and expected: build a small seam if a card needs one (data, not a func field reachable from Game; `server/internal/game/testdata/closure_fields.txt` census may only fall; snapshot changes additive, re-record with `go test ./internal/game/ -run TestSnapshotShapeIsRecorded -update-shape`).
- If a clause needs protocol + client work or an owner decision, leave that card out and list it under `## Skipped` with the exact clause.

## Card rules
- One file per card, `server/internal/cards/effects/<snake_name>.go`, one init(). Opt-in by oracle_id (never scryfall_id). Double-faced / adventure / saga / class / other layouts: find a precedent on develop with grep.
- One printed triggered ability = one registered trigger. Last-known information when the source may have left (CR 608.2h). Intervening-"if" rechecks (CR 603.4). Read keywords and types through the layered view.
- Negative tests use CastSpellParams{Strict: true} / strict activation; test refusals by illegal target, timing or a pool that can't pay.
- TestNoNewExactClonesInTheCatalog: if a card's body matches an existing card's, call or extract one shared helper; never reshape it to dodge the gate.
- Token rows: match the printed token in the dump exactly (name, types, P/T, colours, abilities). Before adding a row to `tokens_table.go`, check develop and the other `feat/s82-*` branches for the same token and reuse its key. An X/X token (size set as it is created) whose printed token card is */* can never match the fixed-size check in TestRealDumpEveryTokenTemplateMatchesAPrintedToken: add its row to `knownUnmatchedTokens` (tokens_realdump_manual_test.go) with a comment naming the card, at the end of the map (roll-up 4 hit this with Wurmquake). With the cut-down dump that test also reports other sets' tokens; only rows you added are yours.
- One-sentence triggers use `effects.Do` primitives that declare a footprint (`cards/effects/footprint.go`); never add a card to `testdata/unfootprinted_trigger_lines.txt`.
- Fixed-amount riders (scry N, surveil N, amass N, venture, learn...) are declared with `Purpose` (TestRiderPurposesMatchThePrintedCard); board wipes, tutors, ramp and draw engines declare their Purpose (ADR 0126).
- Effect keys only via `go test ./internal/cards/effects -run TestEveryPersistedEffectKeyResolves -args -update-effect-keys`. Roadmap: `go test ./internal/roadmap/ -args -update -update-closed`.
- Test gotchas: g.MoveCardByID takes the game lock itself (deadlocks inside WithWriteLock); attack triggers need lockInAttacks(t, g) after DeclareAttacker; batched triggers settle after g.RunStateChecksForTest().

## Oracle fixtures (you make them now)
From server/: `CMDCTRL_SCRYFALL_DUMP=$PWD/../data/scryfall/default-cards.json go test -p 1 ./internal/cards/coverage/ -run TestOracleFixtureIsCurrent -update-oracle -oracle-ids=<your ids, comma-separated>`. Only your own ids; never regenerate them all. Commit the new files under testdata/oracle.

## Before you push
From server/, each as its own command: `go vet ./internal/cards/... ./internal/game/`; `go test ./internal/cards/effects/`; `CMDCTRL_SCRYFALL_DUMP=$PWD/../data/scryfall/default-cards.json go test -timeout 40m ./internal/cards/coverage/` (it must pass except the two dump-wide audits that the cut-down dump cannot satisfy, `TestPlanarCardListMatchesTheDump` and the "catalogued oracle IDs are not in the dump" check of `TestOracleFixtureIsCurrent`; name them in the commit message; CI runs them on the full dump); `go test ./internal/game/` only if you touched internal/game; `go test ./internal/roadmap/`; `gofmt -l` on your files prints nothing; the CI linter, `GOMEMLIMIT=6GiB go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run ./internal/cards/effects/` (and `./internal/game/` if you touched it), with 0 findings; and the purpose audit, `CMDCTRL_SCRYFALL_DUMP=$PWD/../data/scryfall/default-cards.json go test ./internal/decks/ -run RealDumpPurpose -v -count=1`: none of your cards may be named under "read as a board wipe and declare no Sweep" (declare the Sweep, or if it is truly not a wipe, add it to `reviewedNotAWipe` with the reason in its own gofmt group at the end of the map).

## Push and stop
- Commit trailers exactly:
  Sprint: S82 — Card coverage: sets outside Standard, newest first
  Issue: #<the set issue>
- No AI attribution lines anywhere.
- Push `feat/s82-<set>-b<NN>` and STOP: do NOT open a PR. Put the PR-body text in the FINAL commit message: a one-line subject `feat(server): <SET> cards, batch <NN>`, then the cards built, any new seam / token rows / effect keys, `## Skipped` with exact clauses (if any), the checks you ran with results, and one line `oracle-ids=<id>,<id>,...`.
- Your final message: cards built, skipped (with clauses), each check with its result.
