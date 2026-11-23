# Check #5948's done-when on cmd-dev

#5948 (the since-AFR gap run, 2,581 cards across 21 sets, 234 batches) says it is done when every set is at N of N on cmd-dev. The last roll-up, #6076, has merged. Check whether that is true. Read AGENTS.md and the #5948 issue body (`gh issue view 5948`) first.

1. Confirm that cmd-dev (https://cmd-dev.labxp.io) runs current develop. Use the version or health endpoint the server exposes (see docs/lobby.md), or the latest `cmdctrl-cd-dev` run on develop. If the deploy is behind or failed, say so.
2. For each of #5948's sets, get the coverage cmd-dev reports: the public coverage or set-coverage route in docs/lobby.md, or whatever the issue says it measures. If cmd-dev is unreachable from your sandbox, compute the same numbers from develop with the repo's coverage tooling (`internal/cards/coverage`, `cmd/` tools, and the Scryfall data the tooling fetches). Say which source you used.
3. For each set, record:
   - N of N;
   - how many cards are Full and how many have caveats;
   - every card that is missing or not Full, by name.
4. Comment once on #5948 with a table of the sets and the gap list.
   - If every set is complete with no caveats, say the done-when is met. The issue closes on promotion; do not close it.
   - If not, list exactly what is left and why (missing seam, left out on purpose, failed batch).
5. Do not build cards or open PRs. Only measure and report.

Final message: the table and the gap list.
