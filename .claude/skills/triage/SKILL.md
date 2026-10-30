---
name: triage
description: Full-backlog triage of this repo's (cmd_and_ctrl's) open GitHub issues, done the same way every time. It verifies each issue against origin/develop, closes what is already done with evidence, and gives every open issue a type, one tier, one priority and a milestone (needs-triage to 0). It also opens themed sprints with tracker issues and a docs/sprints.md PR, parks owner questions on the issues, and reports P0 and P1 first. Use it whenever the user says "triage", "/triage", "triage the issues", "groom the backlog", "sort the open tickets", "clear needs-triage" or "plan sprints from the open issues", even without the word triage. Not for reviewing or closing pull requests, and not for fixing issues.
---

# Triage the issue backlog

A triage reads every open issue in this repository and leaves the tracker
telling the truth:

- everything already done on `develop` is closed, with evidence;
- every open issue has a type, one tier, one priority and a milestone, so
  `needs-triage` is empty;
- themes with no home get a new sprint;
- decisions only the owner can make are parked on their issues;
- the owner gets a report that leads with anything urgent.

A triage fixes nothing and doesn't touch pull requests.

The scripts are in `.claude/skills/triage/scripts/`; the commands below call
that `$S` (`S=.claude/skills/triage/scripts`). Run them from the repo root or a
worktree root. They work out the repository from `gh
repo view` and keep the snapshot in `$TRIAGE_DIR` (default
`$CLAUDE_JOB_DIR/tmp/triage`; never `/tmp`). The writing scripts space writes
about 2 seconds apart and back off 60 seconds on a 403 or 429. Every script is
safe to rerun, and `DRY_RUN=1` makes the writing ones print instead of write.

## The owner's rules

These are the repo owner's standing preferences. Where a habit of yours
disagrees, they win.

1. **Verify against code, not docs.** Fetch, then read `origin/develop` and
   its history before you call anything "already fixed" or "still blocked".
   `docs/engine-seams.md` runs stale: six of ten past "blocked" verdicts drawn
   from it were wrong. Epic checklists and "Sprint: SNN" lines in issue bodies
   go stale the same way.
2. **Close done work now, with evidence.** A fix merged into `develop` means
   close the issue now, with a comment naming the PR and the merge commit.
   Nothing auto-closes on `develop`. A claim doesn't stop the close once the
   claimed PR has merged.
3. **Respect claims.** An issue claimed in the last 7 days with no merged PR
   is hands-off: fill its gaps (milestone, missing labels) and nothing else.
   A triage claims nothing itself and assigns no one. Whoever starts work on
   an issue claims it first: assign krakenhavoc, comment and add
   `in-progress`.
4. **New themed sprints are welcome.** When five or more issues share a theme
   that no open sprint covers, open `S<n> — <theme>`, with a tracker issue and
   a `docs/sprints.md` section in a `docs:` PR, rather than force a fit.
5. **Park owner questions and keep going.** Comment "**Owner decision
   needed:** …", with the options numbered and the recommended one first, add
   `question`, and move on. List them all in the report, then ask them with
   AskUserQuestion.
6. **Recommend the most rules-faithful option**, even when it is the larger
   change. The owner always picks it.
7. **Labels:** a type, one `tier:*` and one of P0 to P3 on every issue.
   - P0: prod down, data loss or unplayable.
   - P1: next up; a crash or broken game in normal play.
   - P2: soon.
   - P3: when convenient.

   `needs-triage` clears by itself when a milestone is set. `player-report`
   issues come from the in-game button: triage them first and put any P0 or P1
   at the top of the report.
8. **GitHub limits:**
   - At most about one write every 2 seconds; back off 60 seconds on a 403 or
     429.
   - Never use bare `gh run watch`. Poll about every 90 seconds, and read a
     403 as a rate limit, not a red run.
   - `gh pr checks` has no `--json`.
9. **Git:**
   - Make worktrees off `origin/develop`, under `.claude/worktrees/`.
   - Squash-merge into `develop` when CI is green; that has standing
     approval. Never merge into `main`: promotions are the owner's.
   - Docs-only commits and PR titles are `docs: <subject>`, with no scope.
   - Follow Conventional Commits, and add no Claude attribution.
10. **Stay in scope.** Pull requests, including bounty-bot fork PRs, have
    their own flow. Don't implement fixes during a triage.

## The run

1. **Read the ground.**
   - AGENTS.md: "Model tiers" and the sprint paragraph in section 4.
   - The open sprints' goals and the Backlog prose, read from `git show
     origin/develop:docs/sprints.md`. The checkout is often an old branch.
   - `$S/query.py milestones` for the open milestones.
2. **Snapshot:** `$S/snapshot.sh`. The first run takes about two minutes, so
   give it a long timeout. It fetches origin and prints the counts and any
   P0 or P1.
3. **Urgent first.** Run `$S/query.py urgent` and read each issue it lists
   with `$S/query.py show <n>`. Tell the owner about a P0 straight away.
4. **Close what is done.**
   - `$S/query.py done` lists open issues whose closing PR has merged into
     `develop`. For any other issue you suspect is done, use
     `$S/query.py evidence <n>`.
   - Confirm in the code, then post the evidence template with
     `$S/comment.sh <n> <body> completed`.
   - Do this before relabelling, so no time goes on dead issues.
5. **Claims.** `$S/query.py claims` sorts the claimed issues:
   - ACTIVE: fill gaps only.
   - MERGED: close, as in step 4.
   - STALE: if krakenhavoc claimed it, unassign and comment "free to pick
     up"; if someone else did, report it.
6. **Decide every issue, offline.**
   - Get an overview with `$S/query.py lines`, read each issue with
     `query.py show <n>`, and see what is missing with `query.py gaps`.
   - Write one tab-separated row per change to `$TRIAGE_DIR/decisions.tsv`:
     `<n>  <milestone prefix or ->  <labels to add>  <labels to remove>`.
   - Note duplicates, owner questions and homeless themes as you go.
   - For a backlog much bigger than 100 issues, see
     [references/fan-out.md](references/fan-out.md).
7. **Open new sprints** before writing decisions that point into them:
   [references/sprints-doc.md](references/sprints-doc.md).
8. **Write the decisions.** Do a `DRY_RUN=1` pass and read it, then run it
   for real:

   ```sh
   while IFS=$'\t' read -r n ms add rm; do $S/edit.sh "$n" "$ms" "$add" "$rm"; done < $TRIAGE_DIR/decisions.tsv
   ```

   About 100 edits take about 4 minutes, so run the loop in the background or
   with a long timeout. If it stops partway, run it again.
9. **Park the questions:** `$S/comment.sh <n> <body>`, then
   `$S/edit.sh <n> - question`.
10. **Trackers and the docs PR.**
    - Append "Added at the <date> triage:" to the tracker of each sprint
      that gained issues.
    - Close each tracker whose checklist issues are all closed.
    - Put every `docs/sprints.md` change in one `docs:` PR, as in
      [references/sprints-doc.md](references/sprints-doc.md). Wait with
      `$S/ci-status.sh <pr>`, squash-merge when it is green, and remove the
      worktree and branch.
11. **Re-audit.** Issues arrive during a run (two did on 2026-10-30).
    - Run `$S/snapshot.sh` again; it refetches only the changed issues.
    - Run `$S/query.py gaps` and `$S/query.py done`, and triage what is new.
    - Stop when `gaps` is empty, or lists only claimed issues you are
      reporting.
12. **Report**, then ask the parked questions.

## Decision rules in brief

The full rules, with the reasons and examples, are in
[references/decision-rules.md](references/decision-rules.md). Read it before
the first decision of a run.

- **Type:**
  - `fix(...)` titles: `bug`.
  - `seam:` titles: `seam` and `enhancement`.
  - Refactors: `tech-debt`.
  - Flakes: `bug` and `tech-debt`.
  - Player reports keep `player-report`, plus their kind.
- **Tier:** the weakest model that can take the issue to a merged PR alone,
  judged by its hardest part (AGENTS.md table). These are tier 3: a schema or
  wire change, a new seam, a CR edge case, anything waiting on the owner or an
  ADR, an epic. Keep the filer's tier unless it is clearly wrong.
- **Priority:**
  - Keep the filer's priority if it is plausible.
  - A seam behind a set mechanic or three or more cards is P2; a one-card
    seam is P3.
  - An unreproduced player-facing bug is P2.
- **Milestone:**
  - Match the sprint's Goal, not its name.
  - Single-card seams and work waiting on a decision or an outside dependency
    go to Backlog.
  - `Cards batch NN:` issues stay milestone-less by design.
  - Ignore "Sprint:" lines in issue bodies.
- **Duplicates:** keep the older issue, or the one with the claim, the PR or
  the discussion. Close the other with `comment.sh … duplicate <kept>`,
  carrying over anything new. A second report of a bug is evidence for its
  priority.
- **Ask the owner** when the answer changes scope or behaviour and no ADR
  settles it: competing rules-valid designs, a product or UX direction,
  privacy, or closing something as not planned. Decide labels and milestones
  yourself.

Comment bodies for closing, parking, recording a decision, duplicates and
trackers are in [references/templates.md](references/templates.md).

## The report

Lead with what is urgent, and keep it to one line per issue:

```
## Triage <date>: <N> open issues

**Urgent.** P0, then P1, then player reports: `#n P1 <title>`: why, its milestone, what is next. "None." if none.

**Closed as done (k).** `#n`: PR #p (<sha>), and what the develop check showed.

**Owner decisions parked (k).** `#n`: the question; the recommended option.

**New sprints.** S81 — <theme>: tracker #t, k issues; docs PR #p (merged as <sha>).

**Filed.** About k issues relabelled and milestoned; needs-triage now 0 (or each leftover and why). Sprints that gained issues: S75 +3, ….

**Left alone.** Claimed and active: `#n` (who, PR). Others' stale claims: `#n` (who, since when).

**Notes.** Anything surprising: stale docs, issues that arrived mid-run, epics whose checklists are wrong.
```

Then ask the parked questions with AskUserQuestion, not as prose: up to four
per call, recommended option first, marked "(Recommended)". Record each
answer on its issue ([templates.md](references/templates.md)). If you are a
subagent without that tool, end with the list so your caller can ask.

## What a run looks like

The 2026-10-30 run, on launch day:

- 107 open issues reviewed in one pass.
- 3 closed as done, each naming its PR and merge commit. The owner's rule
  then closed 13 more whose claimed PRs had merged; under rule 2 those now
  close in the same pass.
- About 95 relabelled and milestoned, taking `needs-triage` to 0.
- 3 new themed sprints (S78 Launch week, S79 Card coverage, S80 Table
  effects), each with a milestone and a tracker (#4648 to #4650), all added
  to `docs/sprints.md` by docs PR #4653.
- 5 owner questions parked with `question`.
- A report that opened with P0 and P1.

## Files

| Path | What it does |
|---|---|
| `scripts/snapshot.sh [since]` | Read-only. Fetches origin and saves the open issues (bodies, comments), labels, milestones and recent PRs. On a rerun it fetches only what changed. |
| `scripts/query.py <view>` | Read-only views of the snapshot: `lines`, `gaps`, `urgent`, `show`, `recent`, `claims`, `questions`, `evidence`, `done`, `milestones`. |
| `scripts/edit.sh` | Sets an issue's milestone (a title prefix is enough) and labels, paced. |
| `scripts/comment.sh` | Comments, or closes with a comment (`completed`, `not planned`, `duplicate`); skips what is already done. |
| `scripts/new-sprint.sh` | Creates a sprint's milestone and tracker issue; reuses them if they exist. |
| `scripts/sprints_md.py` | Renders a sprint spec into a tracker body, a `docs/sprints.md` section and an index row; adds a sprint, marks one done, applies exact-text edits. |
| `scripts/ci-status.sh <pr>` | Reads a PR's checks once. Exits 0 green, 1 red, 2 running, 3 rate-limited. |
| `references/decision-rules.md` | The full decision rules. |
| `references/templates.md` | Comment and PR bodies. |
| `references/sprints-doc.md` | Opening, extending and finishing sprints, and the docs PR. |
| `references/fan-out.md` | Splitting a big triage across subagents. |
