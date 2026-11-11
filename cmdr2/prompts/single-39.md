You add Commander-deck cards to the catalog of cmd_and_ctrl (GitHub krakenhavoc/cmd_and_ctrl), an original Go Magic: The Gathering Commander engine. You build ONE small batch of similar cards on ONE branch, push it, and STOP. Do NOT open a pull request. You run in a cloud environment that cannot reach Scryfall or wizards.com.

1. `git fetch origin && git fetch origin work/cmdr-cards`. Read `git show origin/work/cmdr-cards:cmdr2/BRIEF.md` in full: it is binding. Read AGENTS.md.
2. Your batch: `cmdr2/batches/39.json` (Split, aftermath, flashback, jump-start and board wipes; 13 cards), read with `git show origin/work/cmdr-cards:cmdr2/batches/39.json`. Read its `reuse` note before the cards and open the helpers and precedent cards it names.
3. Some of this batch may already be built by an earlier worker that ran out of time. For EACH card, before building it, grep its oracle id on origin/develop and on every pushed wave branch: `for b in $(git branch -r | grep -E 'origin/(feat/cmdr2-|rollup/)'); do git grep -l <oracle_id> $b -- server/internal/cards/effects; done`. Skip a card that is found, and say so in the final commit message. If a token or helper your cards need is on one of those branches but not on develop, write the IDENTICAL name and body.
4. `git switch -c feat/cmdr2-s39 origin/develop`. Build card by card, committing every two or three cards, and push after the first three cards pass: `git push -u origin feat/cmdr2-s39`.
5. Run the checks in the brief's "Before the final push", write the final commit message the brief describes (cards built, new helpers and token rows, `## Skipped` if any, the `oracle-ids=` line), push, and stop.

Helpers you write go in `server/internal/cards/effects/cmdr2_s39_helpers.go` with names starting `cmdr2S39`. Build only the cards in your file. If you are running low on time or context, stop adding cards, make what you have pass, and push: a smaller green branch is worth more than a larger broken one; name the cards you did not reach under `## Not reached` in the final commit message.

FINAL MESSAGE: the branch, cards built, cards skipped or not reached and why, check results.
