# Build one GAP2 batch, with the seams it waits on, and land it

Your prompt names one GAP2 batch (b01–b13). It is the GAP2 row in `plan/s82-batches.json` on `work/s82-batches`. Its `after` lists the seam issues it waits on; b01 has none.

Read on this branch: AGENTS.md, s82/BRIEF.md (the GAP and GAP2 sections, and every rule), and s82/SPLITNOTE.md. Also read docs/adding-cards.md and the seam issues (`gh issue view <n>`).

1. Branch `feat/s82-gap2-b<NN>` off origin/develop.
2. **Seams.** For each issue in `after`, if it is still open and not built on develop:
   - Build the seam as contained, additive engine work. Cite CR rules only after grepping data/cr/cr.txt.
   - Follow the closure ratchet and keep snapshots additive. If an ADR is affected, write the amendment, or a new ADR after checking every remote branch for the next free number.
   - Make the commit for the seam separate from the cards.
   - If a seam needs an owner decision, protocol or client work, or is much bigger than the issue says: leave it, comment on the issue with why, and build only the cards that don't need it.
   - Another GAP2 routine may be building a different seam in parallel. Keep yours in new files where you can, so the merges stay simple.
3. **Cards.** Build the row's `missing` cards (with `oracle_ids`) Full with no caveats, per BRIEF.md, using the split layout (`scripts/split-effects.sh --place`).
   - Skip any card develop already has.
   - Where the planner noted a declared skip in an existing test (e.g. `tests/av/batch29_test.go`), update that test.
4. **Checks.** Run every check BRIEF.md lists, plus `go test ./internal/game/` if you touched the engine, and TestEffectsTreeLayout.
5. **PR.** Open one PR, `--base develop`, title `feat(server): GAP2 cards, batch <NN>` (add "and <seam>" if you built seams).
   - Body in the AGENTS.md §4 format.
   - Issues: `Closes #<seam>` for each seam you finished, and "Relates to #5948".
   - Trailers: `Sprint: S82 — Card coverage: sets outside Standard, newest first` and `Issue: #5948`, or the seam issue on seam commits.
   - Run `gh pr merge --squash --auto`.
6. **Waiting for CI.** Poll every few minutes until it merges.
   - If develop moves and it conflicts: merge develop in (plain push, never force). Keep both sides of table files. Regenerate generated docs (engine-seams.md, effect keys) with their tools; never hand-merge them.
   - Fix real failures.
   - Rerun failures that are only infrastructure (OOM, exit 137/143, runner shutdown). The flake `TestNetheresePuzzleWardIgnoredRollDoesNotTrigger` gets one rerun.
   - Stop after 6 hours of wall clock.

Final message: the PR number and state, the seams built or left (with why), and the cards built or skipped.

Nobody will answer you mid-run. Never end your turn while a build, test or CI run is still going.
