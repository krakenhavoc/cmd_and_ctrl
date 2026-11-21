# Move roll-ups 33 and 34 onto the split layout

The #5906 split has merged into develop (PR #5975): `server/internal/cards/effects` is now a core package plus card shards and test packages. Two open roll-up PRs were built on the old flat layout and must be moved:

- PR #6065, branch `rollup/s82-33`
- PR #6064, branch `rollup/s82-34`

Read AGENTS.md and the "This takes effect when #5906's split lands" section of docs/adding-cards.md first.

Do the steps below for each branch, one at a time. Start with rollup/s82-34; #6064 is first in the merge queue.

1. Check out the branch. Merge origin/develop into it with a merge commit, never a rebase.
   - Conflicts on files that moved: take develop's layout, then re-add the branch's own content.
   - Conflicts in table files (token tables, effect_keys.txt and similar): keep both sides.
2. Find every card file and test file the branch added or changed in the old flat directory: `git diff --name-only origin/develop...HEAD -- server/internal/cards/effects/` (files directly in that directory). Run `scripts/split-effects.sh --place <those files>` from the repository root, then `scripts/split-effects.sh --check` if the tool has it.
3. Some edited files were develop's own cards that the roll-up changed. For example, rollup/s82-34 edited `elda_conjurer_of_spectacle.go` to call `gapB45DemonstrateBuild`. Develop's copy has now moved to a shard. Re-apply the edit to the moved file instead of re-adding the flat file. The helper must be exported or placed where the tool's rules say, so both callers can reach it.
4. Run the checks: `go build ./...` and `go vet` on the affected packages. Then `go test ./internal/cards/effects/...` (TestEffectsTreeLayout included), `go test ./internal/cards/coverage/` (clones and the catalog guard; the known cut-down-dump failures are fine) and `./internal/game/`. Then golangci-lint v2.14.0 on `./internal/cards/... ./internal/game/`.
5. Push the branch once (plain push, never force). Commit with the trailers `Sprint: S82 — Card coverage: sets outside Standard, newest first` and `Issue: #5948`. Do not merge the PR; auto-merge is on.
6. Do the other branch. After #6064 merges, #6065 may conflict again; if so, merge develop again and re-check.

Final message: for each branch, what moved, any edits you re-applied, and each check with its result.

Nobody will answer you mid-run. Do not end your turn while a check is still running; wait in the foreground or in an until-loop.
