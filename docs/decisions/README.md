# Architecture decision records

Each ADR is `NNNN-slug.md`, and its number appears in both the filename and the H1.
`server/internal/docsguard` fails CI on a duplicate number or a mismatch, and skips this README.

## Index

0001 WS library … 0108 turn-scoped effects, object history and damage shields; 0109 rule gates, 
land types, mana and cost components; 0110 remember me: durable sign-in, account settings, admins 
and saved setups; 0111 the action dock; 0112 signed-in home, player mode and one decks page; 0113 
small seams for the S58 deck requests; 0114 the Ring tempts you; 0115 commanders die (CR 903.9a); 
0116 a card filter on the revealed-hand pick; 0117 click to act and a per-colour mana stepper; 0118 
strict payment by default, Cast anyway, and alternative costs; 0119 a stack you can follow; 0120 
expand a player's board; 0121 animated dice: an opening roll at the table, dice you can watch, and 
a Roll a die action; 0122 an agent at the table: a local MCP seat; 0123 monitoring: metrics, logs, 
dashboards and alerts; 0124 admin views: accounts, games and who is on now; 0125 a walkthrough that 
keeps up: first-use hints and a refreshed tutorial; 0126 bots that play their decks: the heuristic 
prices what a card does; 0127 answering repeated prompts for you; 0128 playmats: an image on your 
account that shows behind your battlefield; … 0136 planning the turn's mana: rock first, draw 
first, two spells over one; … 0141 bestow: a creature cast as an Aura; 0144 choose a Background and 
the partner pairings; 0145 meld: one permanent, two cards; 0146 voting: will of the council and 
council's dilemma; 0147 auto-assign combat damage, and assigning it on the blockers; 0148 capacity 
limits for a public launch; … 0165 game statistics: a record derived from every finished game; 0167 
controlling another player's turn: Emrakul, the Promised End; 0171 one mana and life ward payment; 
0172 companion: a card outside the game; 0173 shield counters; 0176 a movable stack panel; 0178 life 
instructions after a replaced move; 0179 shop links for the real cards; 0180 secret council; 0181 deploy credentials: deploy runners, environment keys, and failing closed; 0182 your own model key for AI seats; 0183 compleated: entry loyalty from Phyrexian life payments; 0184 weighted model budget and a share per table; 0186 crewing and saddling: who paid, and what they count for; 0187 unless any player pays: one offer per land, graveyard exile, and an either payment; 0188 turn machinery, the rest: skipped turns, effects bound to one turn or one combat, and showing the turn; 0189 the rest of face-down cards and manifest dread; 0193 demonstrate a loop: record one iteration, propose N, replay it; 0199 a hand played revealed; 0200 activating a mana ability of each land; 0201 restarting the game

## Picking a number

**Check every branch, not just the one you are on.** ADR files live in `docs/decisions/` and the
number is in the filename *and* the H1, so two branches that both grab "the next number" collide
silently. They do **not** conflict at merge time: the filenames differ, so git merges both cleanly
and neither author learns. The collision is invisible in the diff and invisible in the merge, and
only shows up when somebody lists the directory — by which point the number is in commit messages,
issue bodies and cross-links in other ADRs.

`TestADRNumbersAreUniqueAndMatchTheirHeading` (`server/internal/docsguard`) fails CI on a duplicate
number and on an H1 that disagrees with its filename. It exists because asking authors to check was
tried first and three collisions reached `main` anyway. If it fires, renumber the ADR with fewer
inbound links (`git grep` its filename), fix its H1, and update the ADR range line in §3.

```bash
git fetch --all --prune
for b in $(git branch -r --format='%(refname:short)' | grep -v HEAD); do
  git ls-tree --name-only "$b" docs/decisions/
done | sed 's|.*/||' | cut -d- -f1 | sort -u
```

Take the first number that does not appear, and say in the PR body which branches you checked.
Numbers are **not** reused when an ADR is renumbered or abandoned — `0005`, `0024`, `0029` and
`0030` are permanently unused for exactly that reason. If a number you already used turns out to be
taken, renumber **your** file (title, filename and every inbound link) rather than asking the other
branch to move; the one that merges first keeps the number.
