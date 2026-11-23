# Bring the docs, roadmap, ADRs and sprints up to date with the code

Ian wants the repository's written record to match the code on `origin/develop` as it is today. In the last few weeks thousands of cards and many seams landed through batch runs: S82 rotated sets, the since-AFR GAP run (#5948), the `internal/cards/effects` split (#5906 / PR #5975), commander precons, formats (ADR 0222, epic #5431) and more. The docs have probably drifted.

Work on a fresh branch off `origin/develop` named `docs/sync-2026-11`. Read AGENTS.md, PLAN.md and docs/decisions/README.md first.

## What to check

Compare each item against the code and against GitHub (`gh issue`, `gh pr list --state merged`, milestones) and correct what is stale.

1. **docs/sprints.md.** For each sprint, check:
   - its status;
   - which tracking issues are closed or open;
   - which work landed (check merged PRs);
   - the active sprint and what comes next.
   Tick checklist items that merged PRs completed, including on the sprint tracking issue bodies when an item is clearly done. Don't invent new sprints.
2. **ADRs (docs/decisions/).**
   - The README index must list every ADR file, with its correct title and status.
   - The AGENTS.md "Current range" line must match the highest number on develop. Also check other remote branches, following the README's command.
   - Where the code plainly implements a Proposed ADR, note it. Where the code contradicts an Accepted ADR, note that too. Don't change ADR decisions or statuses yourself: list them for Ian.
3. **Roadmap.**
   - `server/internal/roadmap/registry.go` must match reality: a seam that has landed is not still Waiting, and a card that is built is not still listed as waiting.
   - Regenerate `docs/engine-seams.md` with the repo's generator, never by hand.
   - Update the card coverage numbers wherever docs quote them.
4. **docs/adding-cards.md.**
   - It must describe the split layout (`cards/<shard>/`, `tests/<id>/`, `scripts/split-effects.sh --place/--check`) as the current state, not as something future.
   - Remove "when #5906's split lands" wording.
5. **docs/repo-layout.md, docs/commands.md, docs/env.md, docs/lobby.md endpoint index, docs/protocol.md, docs/labels.md, AGENTS.md §3.**
   - Every package, directory, make target, script, env var, route, wire message and label that exists in the code is listed.
   - Nothing that no longer exists is listed.
   - Grep the code to confirm each one.
6. **PLAN.md.** Only factual drift, such as a "current state" line that the code contradicts. Leave the charter's decisions alone.
7. Broken relative links in docs/ and in the top-level .md files.

## Rules

- Change only docs, plus `roadmap/registry.go` and its generated doc. Change no Go behaviour, tests (other than a docs guard that needs a new entry), cards or fixtures.
- Cite CR rules only against data/cr or the pinned text, as AGENTS.md §6 says.
- Before opening the PR, run the doc guard tests: `go test ./internal/docsguard/ ./internal/roadmap/`, and any test that reads the files you changed.
- Run the client tests too if you touched labels docs that a test reads.
- Commit and PR format follow AGENTS.md §4.
  - Use the active sprint from docs/sprints.md and its tracking issue.
  - If no issue fits, open one titled "Docs sync with the code, November 2026", labelled `tier:2-standard`, and reference it.
- Open one PR, `--base develop`. The body has `## Summary`, `## Sprint`, `## Issues` and `## Test plan`, plus a section "## For Ian" that lists each item needing an owner decision, one line each:
  - ADR status changes;
  - code/ADR contradictions;
  - sprints to close;
  - anything ambiguous.
- Turn on auto-merge (`gh pr merge --squash --auto`) and wait for CI.
  - If develop moves and the PR conflicts, merge develop in. Plain push, never force.
  - In `registry.go`, keep both sides' intent, then regenerate the doc.
  - Another routine is landing Luke's card PRs at the same time, so expect develop to move.
  - Rerun failures that are only infrastructure (OOM, exit 137/143, runner shutdown). The known flake `TestNetheresePuzzleWardIgnoredRollDoesNotTrigger` gets one rerun.
- Stop when it merges, or after 6 hours.

Final message:
- what you changed, grouped by file;
- the PR number and whether it merged;
- the "For Ian" list.

Nobody will answer you mid-run. Do not end your turn while CI is running; poll in the foreground or in an until-loop.
