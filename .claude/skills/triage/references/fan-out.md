# Splitting a big triage across subagents

One agent can triage about 100 issues in a session; the 2026-10-30 run did 107
alone. Split the work when the backlog is much bigger than that, or when many
issues need a code check. Otherwise a single pass is simpler and more
coherent.

## Who does what

- **The coordinator** (you) takes the snapshot, hands out slices, and decides
  every milestone and new sprint, so the sprints come out coherent. It also
  makes every GitHub write. GitHub's write limit is per account, not per
  agent, so four agents writing at once would trip it.
- **Subagents are read-only.** Each returns proposals, and the coordinator
  reviews them before writing anything.
- Run 3 or 4 at a time, never more: the workstation's CPU is shared. Start
  the next one as one finishes.
- Choose the model by the work:
  - **Sonnet** for label hygiene: a type, tier or priority missing on issues
    whose text makes it plain, and title-based typing.
  - **Opus** for verdicts and design calls: "already done?" and "still
    blocked?" checked against the code, duplicates, tier 3 calls, and
    drafting owner questions with options.
- A subagent that writes docs (rare: normally the coordinator writes the one
  docs PR) works in its own worktree off `origin/develop`, under
  `.claude/worktrees/`.

## Slicing

Cut slices by theme, not by number range, so one agent sees all of a theme's
duplicates and seams together. Use labels and title prefixes from
`query.py lines`: `seam:`, `Card coverage:`, `fix(client)`, bot, ops and
player reports. Give each subagent 20 to 30 issues.

## The brief

Each subagent knows only what the brief says. Include:

- The goal: propose labels, a milestone and a verdict for these issues, and
  write nothing to GitHub.
- The skill's path, and that `references/decision-rules.md` is the rulebook.
- The snapshot: `TRIAGE_DIR=<path>`. Tell it to use `scripts/query.py show
  <n>` and `query.py evidence <n>`, and not to fetch the issues again.
- Its slice of issue numbers.
- The open milestones with their goals, pasted in, so its milestone
  suggestions use real titles.
- That it must verify against `origin/develop` before saying "done" or
  "blocked", and why: the engine-seams doc and epic checklists run stale.
- Scratch goes in `$CLAUDE_JOB_DIR/tmp`, never `/tmp`.
- The return format: one TSV row per issue, with nothing else in the TSV:

  ```
  <n>	<milestone prefix or ->	<labels to add>	<labels to remove>	<verdict>	<one-line reason or evidence>
  ```

  `<verdict>` is one of: `keep`, `close-done <pr> <sha>`, `partly-done`,
  `duplicate-of <n>`, `question` or `new-sprint <theme>`. For a `question`
  verdict, a draft of the owner question follows the TSV.

## Merging the proposals

- Read every `close-done` and `duplicate-of` yourself before acting on it.
  Those are the writes that are hard to undo.
- Group the `new-sprint` themes. A theme with five or more issues becomes a
  sprint; fold smaller ones into the nearest fitting sprint or Backlog.
- Feed the accepted rows to `edit.sh` and `comment.sh` as in the main
  procedure, with a `DRY_RUN=1` pass first.
