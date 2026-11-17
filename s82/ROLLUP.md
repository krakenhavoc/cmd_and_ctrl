# S82 roll-up: brief for a cloud worker

You combine several finished S82 card-batch branches into ONE branch that the owner's session turns into one PR (one CI run). Your prompt names the roll-up number N and the branches. Read AGENTS.md first (binding). Go from server/ as `GOTOOLCHAIN=go1.27.0 GOGC=30 go ... -p 1`, one package at a time; no docker; never `go test ./...`.

## Setup
- Data: `git fetch origin work/s82-cards`, then `mkdir -p data/scryfall data/cr && git show FETCH_HEAD:s82-cards.json.gz | gunzip > data/scryfall/default-cards.json && git show FETCH_HEAD:cr.txt > data/cr/cr.txt` (gitignored; never commit).
- `git fetch origin develop` and every named branch; `git switch -c rollup/s82-<N> origin/develop`.

## Merge
- `git merge --no-ff origin/<branch>` for each branch in the order given. Conflicts are almost always in generated or shared registry files (effect-key ledger, token tables, roadmap / engine-seams tables, census, closure fields, label or purpose tables). Resolve by keeping BOTH sides' entries; for generated files take develop's side and regenerate after the last merge (below). Two branches defining the same helper or token row: keep one, point both cards at it, and say so.
- A branch that conflicts in real engine code (not generated files) and that you cannot resolve with confidence: abort that one merge, leave it out, and list it as "left out" with the reason. Do not drop cards from a merged branch.

## Regenerate after the last merge
- Effect keys: `go test ./internal/cards/effects -run TestEveryPersistedEffectKeyResolves -args -update-effect-keys`.
- Roadmap / seams: `go test ./internal/roadmap/ -args -update -update-closed`.
- Snapshot shape only if a test asks: `go test ./internal/game/ -run TestSnapshotShapeIsRecorded -update-shape` (additive only; never edit fixtures).
- Any other generated file a test names with an `-update…` flag: use that flag, never hand-edit.

## Verify (each its own command, from server/)
`go vet ./internal/cards/... ./internal/game/`; `go test -timeout 60m ./internal/cards/effects/`; `CMDCTRL_SCRYFALL_DUMP=$PWD/../data/scryfall/default-cards.json go test -timeout 40m ./internal/cards/coverage/` (allowed failures: only the dump-wide audits the cut-down dump cannot satisfy, e.g. TestPlanarCardListMatchesTheDump, the "catalogued oracle IDs are not in the dump" check, TestRealDumpEveryTokenTemplateMatchesAPrintedToken or TestCounterCostCardsMatchTheirOracleText failing on OTHER sets' cards; anything naming a card or token from these branches is yours to fix); `go test ./internal/game/` if any branch touched internal/game; `go test ./internal/roadmap/ ./internal/docsguard/`; `gofmt -l .` prints nothing. Fix real failures on the roll-up branch with small commits.

## Finish
- Commit trailers: `Sprint: S82 — Card coverage: sets outside Standard, newest first` and `Issue: #5458`. No AI attribution.
- Push `rollup/s82-<N>` (plain push) and stop; do not open a PR.
- Final message, written as PR text: one Summary bullet per batch (set, batch, card count, any new seam / token / effect key), the branches left out (with reason), each check with its result, and every `oracle-ids=` line from the merged branches' final commits, combined into one line.
