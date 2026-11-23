# Plan the last 257 cards of #5948 (the GAP2 pass)

The #5948 check (the latest comment on #5948, 2026-11-23) found 2,324 of 2,581 since-AFR cards Full on develop. **257 are left**: 238 have no registration at all, and 19 (marked †) have a file but no Completeness the scan could read. You plan the work that finishes them. You do not build cards.

Read AGENTS.md, the #5948 comment's gap list (`gh issue view 5948 --comments`), and s82/BRIEF.md on this branch (the worker brief, GAP section).
- Plan rows live in `plan/s82-batches.json` on `work/s82-batches`. Read it there.
- The card dump is `s82-cards.json.gz` on this branch. Unpack it as BRIEF.md says.

1. **The 19 † cards.** Check each one on origin/develop. Is it registered through a shared helper, with Full and no caveats? List the ones that are fine. Any that are not go into the plan.
2. **The 238 unregistered cards.** Find out why each one was left out, using all of these:
   - the `## Skipped` sections in the commit messages of the `feat/s82-gap-b*` branches and the `rollup/s82-*` branches (`git log`, `git ls-remote`);
   - the merged roll-up PR bodies, #5962 and #6064–#6076;
   - grep of origin/develop, in case a card is there under another oracle_id or name.
   Put each card in one of these classes:
   - (a) **buildable now** on existing seams; it was dropped or missed;
   - (b) **needs a contained seam**: name the seam;
   - (c) **needs an owner decision or protocol/client work**: quote the exact clause. Mystery Booster 2 playtest and Un-style text often lands here;
   - (d) **already on develop** under another id: a false gap. Explain why.
3. **Write the plan.** Add rows with `set: "GAP2"`, batches b01…bNN, about 10 cards each, grouped by mechanic, on `work/s82-batches` in `plan/s82-batches.json`.
   - Use the same row shape as the GAP rows: set, issue 5948, batch, theme, missing, oracle_ids in the same order, after, status "ready".
   - Class (a) cards go in first.
   - Class (b) cards go in rows whose `after` names the seam issue.
   - Commit and push `work/s82-batches` (plain push).
4. **Add a GAP2 section to s82/BRIEF.md** on this branch, beside the GAP section.
   - Branch `feat/s82-gap2-b<NN>`, commit subject `feat(server): GAP2 cards, batch <NN>`, trailer `Issue: #5948`.
   - Rules otherwise as GAP: skip a card develop already has, and push even if failing.
   - Also add the split-layout note from s82/SPLITNOTE.md, so new card files are placed with `scripts/split-effects.sh --place`.
   - Commit and push.
5. **Class (b) seams.** Open one issue per seam: labels `seam`, `enhancement`, `tier:3-design`, milestone S82, with the cards, the clauses and the CR rules from data/cr/cr.txt. Check for an existing open seam issue first (`gh issue list --label seam`) and reuse it.
6. **Report.** Comment once on #5948 with:
   - the classification counts;
   - the class (c) list with clauses, as owner questions;
   - the class (d) list;
   - the GAP2 batch count.

Final message: a summary line, then the GAP2 rows (batch, theme, count), the seam issues you opened, and the class (c) list. Nobody will answer you mid-run.
