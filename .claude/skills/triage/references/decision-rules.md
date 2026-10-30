# Decision rules

How to decide each open issue. Most of these come from the owner's standing
preferences; the rest from what earlier triages got wrong. Where a rule names
a sprint, it is the 2026-10-30 mapping, given as an example. Read the current
sprints' goals before you reuse it.

## Contents

- [Is it already done?](#is-it-already-done)
- [Claims](#claims)
- [Type](#type)
- [Tier](#tier)
- [Priority](#priority)
- [Milestone](#milestone)
- [New sprints](#new-sprints)
- [Duplicates](#duplicates)
- [Owner questions](#owner-questions)
- [Leave alone](#leave-alone)

## Is it already done?

Decide this from the code on `origin/develop`, never from a doc. On
2026-09-19, six of ten "blocked on a seam" verdicts read from
`docs/engine-seams.md` were wrong: the seam had shipped, or the blocker was a
different one. Epic checklists and the "Sprint: SNN" line in an issue body go
stale the same way.

- Fetch first (`snapshot.sh` does it), then read `origin/develop`, not the
  checkout, which is often an old feature branch.
- A merged PR with `Closes #n` in its `## Issues` section is a strong lead
  (`query.py done` lists them). Before you close the issue, check two things:
  that the change is on `origin/develop`, and that it covers the issue's whole
  ask. A PR that only mentions the issue ("found during #4455", "Refs #n") is
  not a fix.
- Useful checks:
  - `git grep -n '<pattern>' origin/develop -- <paths>`
  - `git show origin/develop:<file>`
  - `git log origin/develop -E --grep='#<n>([^0-9]|$)'`
- A card or seam: is the card registered (`git grep -l -F '"<Card Name>'
  origin/develop -- server/internal/cards`)? Does the primitive exist under
  `server/internal/cards/effects` or `server/internal/game`?
- A citation fix: grep the old citation and confirm that only valid uses are
  left.
- A client cue or feature: grep the client lib for it (for example
  `client/src/lib/sounds.ts` or `tableEffects.ts`).
- An epic or umbrella: check each child with `gh issue view <n> --json
  state,closedAt`, not the checklist ticks.

Verdicts:

- **Done.** Close it now as completed, with the evidence comment from
  [templates.md](templates.md). Nothing auto-closes on `develop`; the
  promotion closer only acts when `main` moves. If the issue has no milestone,
  give it the milestone of the sprint its PR named, so the milestone counts
  stay honest.
- **Partly done.** Comment on what is left and keep it open. Re-scope the
  title only if it is unclaimed.
- **Done, but its acceptance check can't run yet** (for example a nightly that
  runs only on `main`). Keep it open and say so in a comment.
- **Still blocked.** Name the missing primitive you looked for and where.

## Claims

Someone has claimed an issue if it is assigned, labelled `in-progress`, named
by an open PR, or has a claim comment ("Claimed", "Picking this up").
irobinson010's agents claim with a bare "Claimed" comment and don't add
`in-progress`. `query.py claims` sorts them:

- **ACTIVE** (claimed in the last 7 days, or an open PR): hands off. Fill
  gaps only: a missing milestone, type, tier or priority. Add `in-progress`
  when an open PR shows the work is under way. Don't re-scope, retitle,
  relabel what is already set, reassign or close it. List it in the report.
- **MERGED** (the claimed PR is merged into `develop`): close it with evidence,
  like any other done issue. The claim doesn't block closing once the work has
  merged. The 2026-10-30 run left 13 of these open, and the owner's rule then
  closed them.
- **STALE** (the claim is older than 7 days and has no open PR): if
  krakenhavoc claimed it, unassign, drop `in-progress` and comment "free to
  pick up". If someone else claimed it, list it in the report and leave the
  claim alone.
- **UNDATED** (an assignee or label but no claim comment): read the issue
  timeline before you treat it as either.

Triage itself claims nothing. Don't assign anyone while triaging: on
2026-10-04 the owner had about 50 bulk-assigned issues unassigned again. Claim
an issue (assign krakenhavoc, comment, add `in-progress`) only when work on it
actually starts. That is outside this skill.

## Type

At least one of `bug`, `enhancement`, `documentation`, `tech-debt`, `seam`,
`question`. `query.py gaps` checks the same set.

- A `fix(...)` title, or something broken: `bug`.
- A `seam:` title: `seam` and `enhancement`.
- A refactor, or deferred cleanup: `tech-debt`.
- A flaky test: `bug` and `tech-debt`.
- A player report keeps `player-report` and gets its kind (`bug` or
  `enhancement`).
- `outage` (from the uptime watcher) also needs `bug`. `deck-request` also
  needs `enhancement`.
- `question` marks a parked owner decision. Give the issue its real type as
  well, unless the issue is nothing but the decision.

## Tier

Exactly one of `tier:1-mechanical`, `tier:2-standard` and `tier:3-design`:
the weakest model that can take the issue to a merged PR unsupervised. The
table is in AGENTS.md, "Model tiers". Tier by the hardest part, and pick the
higher tier when unsure. Keep the filer's tier unless it is clearly wrong.

These are always tier 3: a snapshot-schema or wire change, a new engine seam,
CR edge cases (layers, replacement ordering, copies, loops), anything that
waits on an owner decision or an ADR, a diagnosis with no repro, and epics.

## Priority

Exactly one of P0 to P3.

| Label | Meaning |
|---|---|
| P0 | Production down, data loss, or games unplayable. Tell the owner as soon as you see one; don't wait for the report. |
| P1 | Next up: a crash or broken game reachable in normal play. |
| P2 | Soon: a real cost, or a bug with a workaround. |
| P3 | When convenient: cleanup, polish, nice to have. |

- Keep the filer's priority when it is plausible. Don't churn.
- A seam that blocks a set mechanic or three or more cards is P2 (sneak,
  mayhem and either/or costs on 2026-10-30). A one-card seam is P3.
- A player-facing bug reachable in normal play but not yet reproduced ("bug?")
  is P2 until a repro says otherwise.
- `player-report` issues come from the in-game button. Triage them first, and
  put any P0 or P1 at the top of the report.

## Milestone

Every open issue gets one, except `Cards batch NN:` issues: those are the
rolling catalog track and stay milestone-less by design
(`.github/workflows/issue-triage.yml`). Setting a milestone clears
`needs-triage` by itself.

- Sprint themes are narrow. Match the issue to a sprint's **Goal** in
  `docs/sprints.md` (read it from `origin/develop`) and the milestone
  description, not to its name alone.
- Don't trust a "Sprint: SNN" line in an issue body. Batch agents copy it
  forward long after it stops fitting.
- A tracker issue sits in its own sprint's milestone.
- Single-card seams, work that waits on an owner decision or an outside
  dependency (a model host, a paid key) and ADR-blocked engine work go to the
  **Backlog** milestone, and get a line in the Backlog prose of
  `docs/sprints.md`.
- An issue with a parked owner question goes where the recommended option
  would put it, or to Backlog if the decision blocks all of the work.
- Examples from 2026-10-30:
  - heuristic-bot bugs went to S75 ("sharper bots");
  - idle and archived-table launch issues went to S70;
  - set coverage (`Card coverage: <set>`, the umbrella #4452) and every seam
    and fix its batches filed went to S79;
  - table cues and their polish (epic #4399) went to S80;
  - launch-week player reports and flakes went to S78.

## New sprints

Propose and open a new themed sprint when five or more issues share a theme
that no open sprint's goal covers. Don't stuff them into a half-related one:
the owner prefers new milestones to forced fits. With fewer than five, use the
nearest sprint that genuinely fits, or Backlog.

A new sprint is a milestone named `S<n> — <theme>` (em dash), a tracker issue
with the same title, and a section in `docs/sprints.md`, all opened in the same
run. [sprints-doc.md](sprints-doc.md) has the format. Take the next unused
number: check the sprint index and every milestone, open and closed.

When every checklist issue of a sprint is closed, close its tracker with a
comment naming the last closers, and mark the sprint done in the same docs PR.

## Duplicates

- Keep the older issue, or the one with the discussion, the claim or the PR.
  Close the other with `comment.sh <n> <body> duplicate <kept>`, saying what
  it duplicates and carrying over any detail the kept issue lacks.
- A second player report of the same bug is evidence: link it on the kept
  issue and reconsider the priority.
- Never close a claimed issue as a duplicate of an unclaimed one; close the
  unclaimed one.
- Overlap without being the same ask: link both ways with one line each, and
  keep both.

## Owner questions

Ask when the answer changes scope or behaviour and no ADR settles it:

- two or more rules-valid designs;
- a product or UX direction;
- privacy or data kept about players;
- closing something as not planned;
- anything you would otherwise have to guess.

Don't ask about labels, tiers, priorities or milestones. Decide those
yourself.

- Park the question on the issue and move on; don't wait. The format is in
  [templates.md](templates.md): "**Owner decision needed:**", then numbered
  options with the recommended one first, then "Parked at the <date> triage."
  Add the `question` label.
- Recommend the option most faithful to the Comprehensive Rules, even when it
  is the larger change. The owner always picks it, and name the CR rule. For
  non-rules questions, recommend what fits PLAN.md's scope and the
  automation-first direction.
- If the issue body already frames options, restate them in the standard form
  rather than inventing new ones.
- List every parked question in the report, then ask them with
  AskUserQuestion (up to four per call, recommended option first, marked
  "(Recommended)").
- When the owner answers, record it on the issue: "**Owner decision
  (<date>):** <the choice and what it means>". Remove `question`, and re-tier
  and re-milestone the issue if the answer changed its scope.
- On a rerun, `query.py questions` shows parked questions that have later
  comments. Read them: an answer may already be there.

## Leave alone

- Pull requests, including bounty-bot fork PRs. Those have their own flow:
  close with a neutral comment and no thanks, and the owner decides on blocks.
- Fixing anything. Triage records what is true and what is next; a P0 is
  reported straight away, not fixed inside the triage.
- Promotions to `main`, which are the owner's.
