# Land the open roll-ups one at a time

Several S82 roll-up PRs are open. Each one that merges makes the others conflict with develop again. You land them **one at a time**, in the order your prompt gives.

Read AGENTS.md and s82/PLACE.md on this branch first. PLACE.md is the recipe for bringing one branch up to date with develop on the split `effects` layout: merge develop, resolve, place any flat files, run the checks.

For each PR in order:

1. Fetch origin/develop and the PR's branch.
2. If the PR is already MERGED, go to the next one.
3. Bring the branch up to date with develop by following PLACE.md. Rules for conflicts:
   - Table files (token tables, `effect_keys.txt`, purpose test maps and similar): keep both sides.
   - `docs/engine-seams.md` is generated. Resolve `server/internal/roadmap/registry.go` by keeping both sides' intent (a card built on either side leaves its seam's Waiting list), then regenerate the doc with the repo's generator. Never hand-merge the doc.
   - Duplicate oracle IDs, or identical bodies across the branch and develop: keep develop's copy and drop or share the branch's.
4. Run the checks PLACE.md lists. Push (plain push, never force). Commit with the trailers `Sprint: S82 — Card coverage: sets outside Standard, newest first` and `Issue: #5948`.
5. Make sure auto-merge is on (`gh pr merge <n> --squash --auto`). Wait for CI, polling `gh pr view <n> --json state,mergeable` and `gh pr checks <n>` every few minutes.
   - If it merges: go to the next PR.
   - If develop moved and it conflicts again: go back to step 3.
   - If a check fails: read the job log. Fix a real failure on the branch and push.
   - The effects test `TestNetheresePuzzleWardIgnoredRollDoesNotTrigger` (tests/cr, `dice_5681_test.go`) is known to flake. If that is the only failure, rerun the failed jobs (`gh run rerun <run-id> --failed`) once. If it fails twice in a row, note it in your final message.
   - A runner shutdown or OOM (exit 137/143) is infrastructure: rerun the failed jobs.
6. Stop after the last PR has merged, or after 4 hours of wall clock, whichever comes first.

Final message: for each PR, whether it merged, what you resolved, and any check that needed a rerun.

Nobody will answer you mid-run. Do not end your turn while waiting; poll in the foreground or in an until-loop.
