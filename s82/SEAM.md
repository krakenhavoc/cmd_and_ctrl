# S82 seam: brief for a cloud worker

Your prompt names a seam issue #N. Read it in full (`gh issue view N` or `curl -s https://api.github.com/repos/krakenhavoc/cmd_and_ctrl/issues/N`): it names the card, the gap and what "done" means. Read AGENTS.md (binding) and the docs/adding-cards.md subsections you need. Everything in s82/BRIEF.md applies (data setup from this branch, Go flags, card rules, oracle fixtures you generate yourself, checks, linter), with these differences:

- Branch: `feat/<N>-<short-slug>` from origin/develop.
- Build the seam the smallest additive way the engine allows: data, not a func field reachable from Game; the closure census may only fall; snapshot changes additive only (re-record with `-update-shape`, never edit fixtures). Cite CR rules only after grepping data/cr/cr.txt. Then build the card(s) the issue names, Full, no caveats, with engine-driven tests including the edge cases the issue lists.
- Grep develop first: the mechanism may already exist under another name; reuse it if so and say so.
- If the seam truly needs protocol + client work or an owner decision, stop, push what you have, and say exactly what decision is needed.
- Before pushing also run: `go test ./internal/game/`, `GOMEMLIMIT=6GiB go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run ./internal/cards/... ./internal/game/` (0 findings), and `CMDCTRL_SCRYFALL_DUMP=$PWD/../data/scryfall/default-cards.json go test ./internal/decks/ -run RealDumpPurpose -v -count=1` (your card must not be named).
- Trailers: `Sprint: S82 — Card coverage: sets outside Standard, newest first` and `Issue: #N`. Push the branch and stop; no PR. Final commit message = PR text with `Closes #N` and an `oracle-ids=` line.
