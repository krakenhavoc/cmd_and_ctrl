# Move one roll-up onto the split layout

The #5906 split merged into develop (PR #5975): `server/internal/cards/effects` is now a core package plus card shards (`cards/<shard>/`) and test packages (`tests/<id>/`). The roll-up branch named in your prompt was built on the old flat layout and must be moved. Read AGENTS.md and the "This takes effect when #5906's split lands" section of docs/adding-cards.md first.

1. Fetch origin develop and the branch, check out the branch, and merge origin/develop into it with a merge commit, never a rebase.
   - Conflicts on files that moved: take develop's layout, then re-add the branch's own content.
   - Conflicts in table files (token tables, effect_keys.txt and similar): keep both sides.
2. List every card and test file the branch added or changed directly in the old flat directory: `git diff --name-only origin/develop...HEAD -- server/internal/cards/effects/` (only files directly in that directory). Run `scripts/split-effects.sh --place <those files>` from the repository root, then `scripts/split-effects.sh --check`.
3. If the branch edited one of develop's own cards, re-apply that edit to the moved file instead of re-adding the flat file. A helper that two packages need must be exported or placed where the tool's rules put it.
4. Run the checks: `go build ./...`, `go vet` on the affected packages, `go test ./internal/cards/effects/...` (TestEffectsTreeLayout included), `go test ./internal/cards/coverage/` (clones and the catalog guard; the known cut-down-dump failures are fine) and `./internal/game/`. Then golangci-lint v2.14.0 on `./internal/cards/... ./internal/game/`.
5. Push the branch once (plain push, never force). Commit with the trailers `Sprint: S82 — Card coverage: sets outside Standard, newest first` and `Issue: #5948`. Do not open or merge a PR.

Final message: what moved, any edits you re-applied, and each check with its result.

Nobody will answer you mid-run. Do not end your turn while a check is still running; wait in the foreground or in an until-loop.
