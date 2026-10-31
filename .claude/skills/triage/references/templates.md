# Comment and PR templates

Write each body to a file in `$TRIAGE_DIR/bodies/` and post it with
`comment.sh`, which skips a body that is already on the issue. Keep comments
short and factual. Name PRs and commits so a reader can check them.

## Closing as done

Name the PR, the merge commit's short SHA and what the check on `develop`
showed:

```
Fixed in #4618, merged into develop as afc49f5f2. Closing now, since nothing
auto-closes on develop. It reaches prod with the next promotion.
```

```
Done on develop: PR #4605 (be5336217, merged 2026-10-30) corrected the search
citations to CR 701.23 across code, tests and docs. A grep of origin/develop
for `701.19` now finds only regenerate citations. Closing as completed
(2026-10-30 triage).
```

When the SHA is already on `main` (`query.py evidence <n>` says `main=yes`),
say "It is live since the <date> promotion" instead of "reaches prod with the
next promotion".

## Partly done

```
Most of this is done on develop: #4581 (3f2a1c0d7) added <what>. Still open:
<what is left, concretely>. Keeping this open for that (2026-11-13 triage).
```

## Owner decision needed

Post the comment, then add the `question` label
(`edit.sh <n> - question`):

```
**Owner decision needed:** how far should the fix reach? Today type-filtered
cast permissions judge a card's front face, so an Adventure or an MDFC
back-face instant or sorcery is never offered.

Options:
1. (Recommended) Judge the face that would be cast (CR 601.3e) for every
   type-filtered cast permission, and recheck each Full stamp that depends on
   today's behaviour. This is the most rules-faithful option.
2. Fix only the permissions a registered card needs today, and leave the rest
   for later.
3. Keep front-face judging, and add a player-facing caveat to the affected
   cards.

Parked at the 2026-10-30 triage. Milestone S79 (#4649).
```

## Owner decision recorded

Post this after the owner answers, then remove `question`
(`edit.sh <n> - "" question`):

```
**Owner decision (2026-10-30):** option 1. <One or two sentences on what that
means for this issue's scope.>
```

## Duplicate

```
Duplicate of #4601, which has the claim and the open PR. The one detail here
it lacked (<detail>) is now noted there. Closing (2026-11-13 triage).
```

## Triage note on an issue

Use this when the triage changed something the filer should know, such as a
privacy limit or a re-scope:

```
Triage note (2026-10-30): in-app reports become public GitHub issues, so an
attached settings copy must hold nothing personal (ADR 0149): no names, IDs or
tokens, only the display and auto-pass preferences. Milestone S78 (#4648).
```

## Free to pick up

Post this on krakenhavoc's own stale claims, after unassigning and removing
`in-progress`:

```
Unassigning: no PR has come of this claim since 2026-10-14. Free to pick up
(2026-11-13 triage).
```

## Closing a finished sprint's tracker

```
Closing the S76 tracker as completed (2026-10-30 triage). Every checklist issue
is closed: the last two were #3525 (closed 2026-10-24 by #4230 and #4252) and
#4126 (closed 2026-10-20 by #4129). The S76 milestone has no other open issue.
```

## Adding issues to an existing sprint's tracker

Append a block to the tracker body: read it with `gh issue view <t> --json body
--jq .body`, edit it, and write it back with `gh issue edit <t> --body-file
<f>`:

```
Added at the 2026-10-30 triage (heuristic bot):
- [ ] #4320 untargeted until-end-of-turn rows (Purphoros, Shadowspear) fire with nothing to affect
- [ ] #4559 Lion Sash's {W} exile ability eats the bot's mana
```

Add a matching sentence to that sprint's Status in `docs/sprints.md`.

## The docs PR body

The title is `docs: <subject>`, with no scope. The body follows AGENTS.md's PR
format. Trackers get `Refs`, never `Closes`.

```
## Summary
- Adds sprint sections and index rows for S78 (<theme>, #4648) and S79 (<theme>, #4649), opened at the 2026-10-30 triage of every open issue.
- Marks S76 done (its tracker #3616 closed with every checklist issue closed), records #4564 joining S70, and parks #4407 in the Backlog.

## Sprint
S78 — <name>

## Issues
Refs #4648
Refs #4649

## Test plan
- [x] Each new section follows the existing shape: goal, tracker, checklist, exit criteria, status.
- [x] Every issue linked in the new sections has the matching milestone on GitHub.
- [ ] CI is green.
```
