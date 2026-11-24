# Older-sets roll-up (set OLD): brief for a cloud worker

Follow s82/ROLLUP.md on this branch exactly, with these differences. Only one roll-up runs at a time, so you are the only one merging.

- **Card data:** `git show FETCH_HEAD:all-cards.json.gz | gunzip > data/scryfall/default-cards.json`, not s82-cards.json.gz.
- **Split layout:** follow s82/SPLITNOTE.md. After the last merge, run `scripts/split-effects.sh --check`. If any card or test file is flat, run `--place <files>`. Then run `go test ./internal/cards/effects/ -run TestEffectsTreeLayout`.
- **Branch:** `rollup/s82-old-<N>`, fresh from origin/develop. N is in your prompt.
- **Which branches:** your prompt lists them.
  - Skip a branch whose cards develop already has (an earlier roll-up merged it).
  - Before merging, run `git log -1 --format=%B origin/<branch>`. If the commit says the worker gave up on all its cards, leave the branch out.
- **Duplicates:** two branches may register the same oracle_id. Keep the first, drop the other file, and say so.
- **Finish: open and land the PR yourself.**
  - `gh pr create --base develop`. Title: `feat(server): older-set cards, roll-up <N>`. The body is your ROLLUP.md final message in AGENTS.md PR format: Sprint `S82 — Card coverage: sets outside Standard, newest first`, and in Issues `Refs #5458` (no closing lines).
  - Then `gh pr merge --auto --squash`.
  - Wait for CI, polling `gh pr checks` in an until-loop. Fix failures with small commits on the roll-up branch until it merges.
  - If develop moves and the PR conflicts, merge origin/develop in. Keep both sides of generated tables and regenerate them.
  - Once it has merged, delete the merged `feat/s82-old-b*` branches you included (`git push origin --delete ...`).
- **Final message:** PR number and state, the cards added, the branches left out (with reasons), and any card that needs an owner decision.
