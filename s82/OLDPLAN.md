# Plan the pre-AFR sets for S82, newest first, skipping Luke's

Ian decided on 2026-11-23: after the since-AFR gap (#5948), S82 continues with the sets released before Adventures in the Forgotten Realms (2021-07-23), **newest first**.
- **Skip every set Luke has in progress.** Those are set issues labelled `in-progress` (currently ODY #5613, TOR #5612, ONS #5610, LGN #5609, SCG #5608, DST #5605, CHK #5603), plus any set with an open `claude/*` branch or an open PR from Luke's sessions. Also skip the sets whose batches Luke already landed (DIS, MOR, TSP, FUT, PLC, 5DN). Their set issues stay his.
- Roll-ups will drop any duplicates later, but don't plan cards that another branch is building.

Read AGENTS.md and s82/BRIEF.md (the Full card data section) on this branch. You plan; you build nothing.

1. **Data.**
   - `all-cards.json.gz` and `sets-index.json.gz` are on this branch. The index maps set code → name, released_at, set_type, digital and oracle_ids.
   - The S82 set issues are #5459–#5642 (umbrella #5458). Match each to its set code by title.
   - Leave out digital-only sets and the token, memorabilia and art-series sets.
2. **Gaps.** For each S82 set issue released before 2021-07-23, and not skipped:
   - list the set's oracle_ids that have no registration on origin/develop (grep the oracle_id under server/internal/cards);
   - also remove those that an open `feat/*` or `claude/*` branch adds.
   - Order the sets newest first.
3. **Batches.** Group each set's missing cards into batches of about 10, by mechanic.
   - Write rows `{set:"OLD", issue:<set issue>, batch, theme, setcode, missing:[names], oracle_ids:[...], after:[], status:"ready"}` into `plan/s82-batches.json` on `work/s82-batches`.
   - Batch numbers run continuously in newest-first order (b001…).
   - Where a card clearly needs a seam the engine lacks (grep develop for the mechanic), put the seam issue in `after`. Reuse an open `seam` issue, or open one (labels `seam`, `enhancement`, `tier:3-design`, milestone S82).
   - Commit and push.
4. **Brief.** Add an "Older sets (set OLD)" section to s82/BRIEF.md on this branch:
   - Branch `feat/s82-old-b<NNN>`, subject `feat(server): <set name> cards, batch <NNN>`, trailer `Issue: #<set issue>`.
   - Card data from `all-cards.json.gz`, plus the SPLITNOTE.md split-layout rule.
   - Skip anything develop already has. Push even if failing.
   - Commit and push.
5. **Report.** Comment on umbrella #5458 with:
   - a per-set table (set, issue, missing count, batches, seams waited on);
   - the skipped sets and why;
   - the totals.

Final message: the totals (sets, cards, batches), the first 30 batch rows (batch, set, count, after), and the seams opened. Nobody will answer you mid-run.
