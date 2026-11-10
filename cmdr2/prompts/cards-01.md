You add Commander-deck cards to the catalog of cmd_and_ctrl (GitHub krakenhavoc/cmd_and_ctrl), an original Go Magic: The Gathering Commander engine. You build TWO small batch files of similar cards on ONE branch, push it, and STOP. Do NOT open a pull request. You run in a cloud environment that cannot reach Scryfall or wizards.com.

1. `git fetch origin && git fetch origin work/cmdr-cards`. Read `git show origin/work/cmdr-cards:cmdr2/BRIEF.md` in full: it is binding. Read AGENTS.md.
2. Your batches: `cmdr2/batches/01.json` (Simple lands; 12 cards) and `cmdr2/batches/02.json` (Mana rocks, mana creatures and mana triggers; 10 cards), read with `git show origin/work/cmdr-cards:<path>`. Read each file's `reuse` note before the cards and open the helpers and precedent cards it names.
3. `git switch -c feat/cmdr2-b01 origin/develop`. Build batch 1 first, card by card, committing every two or three cards. When batch 1 passes its tests, `git push -u origin feat/cmdr2-b01`. Then build batch 2 the same way.
4. Run the checks in the brief's "Before the final push", write the final commit message the brief describes (cards built, new helpers and token rows, `## Skipped` if any, the `oracle-ids=` line), push, and stop.

Helpers you write go in `server/internal/cards/effects/cmdr2_b01_helpers.go` with names starting `cmdr2B01`. Build only the cards in your two files. If you are running low on time or context, stop adding cards, make what you have pass, and push: a smaller green branch is worth more than a larger broken one; name the cards you did not reach in the final commit message under `## Not reached`.

FINAL MESSAGE: the branch, cards built, cards skipped or not reached and why, check results.
