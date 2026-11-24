# #5948 class (c): owner-approved work (Ian, 2026-11-23)

Read the #5948 comments first: the GAP2 planner's comment, which has the class (c) list with clauses under Q1–Q4, and Ian's decisions comment after it.
Then read AGENTS.md, s82/BRIEF.md (GAP and GAP2 sections), s82/SPLITNOTE.md and docs/adding-cards.md.
Card data: s82-cards.json.gz on this branch (unpack as BRIEF.md says). Rules: cr.txt on this branch. Grep it before citing any rule.

Your prompt names ONE task below. Do only that task.

## Task CONJURE: conjure, perpetual and boons (Q1, 13 cards)
- Write the ADR amendments to ADR 0099 and ADR 0103. Ian approved: a card may conjure a real catalog card mid-game, apply perpetual changes, and give boons.
- Design the engine seam:
  - A conjured card is a new card object built from the catalog row by name. It has an owner and goes into the named zone (hand, library or battlefield; random picks among names included).
  - Perpetual modifications follow the card across zones.
  - A boon is a player-scoped one-shot emblem-like trigger.
  - Snapshots stay additive; follow the closure ratchet.
- The client must show a conjured card like any other card. Check that nothing assumes every card came from a decklist (deck counts, the reveal log, the bot seat).
- Build all 13 Q1 cards Full, with no caveats, and with engine-driven tests.
- Branch `feat/5948-conjure`.

## Task COMMANDERCOLOUR: choose a colour before the game (Q2, 3 cards)
- Build "If <this> is your commander, choose a colour before the game begins. <It> is the chosen colour." for The Prismatic Piper, Faceless One and Clara Oswald.
- Add a pre-game prompt for the commander's controller: engine and protocol (docs/protocol.md) and client.
- The chosen colour sets the card's colour and its colour identity for deck validation, per the CR (grep 903 in cr.txt).
- Handle bots: the bot seat answers the prompt sensibly.
- Clara Oswald's other abilities must also be Full; check the `gapDoctorsCompanionCards` helper.
- Leave the other Q2 cards out.
- Branch `feat/5948-commander-colour`.

## Task SECRET: hidden or simultaneous choices (Q3, 5 cards)
- Build a generic secret-choice-then-reveal prompt in the engine, protocol and client:
  - a number (0 or more, or within a range, with an "unused" rule);
  - a creature pick (each player picks among their own and others');
  - a guess.
- The answers stay hidden from other seats in game state and on the wire until the reveal. Check the redaction path the same way hidden hands are checked.
- Bots answer too.
- Build Itazura, Lingering Wick; Duelists' Convocation International; The Toymaker's Trap; Gollum, Scheming Guide; and Call to the Void. All Full.
- Branch `feat/5948-secret-choices`.

## Task TRIAGE: the remaining one-offs (Q4)
For every class (c) card not covered by Q1–Q3, decide which group it belongs to:
- (1) **a real printed card** (not Mystery Booster 2 playtest or Un-style text) whose clause fits a contained engine seam;
- (2) **real**, but it needs protocol or client work;
- (3) **playtest or Un-style text** that needs a design call.

Then:
- For (1) and (2): open seam issues, grouped by mechanic. Use labels `seam`, `enhancement`, `tier:3-design`, milestone S82. Reuse an existing issue where there is one.
- Add GAP3 plan rows (set "GAP3", about 10 cards each, `after` naming the seam issues) to `plan/s82-batches.json` on `work/s82-batches`.
- For (3): comment on #5948 with a short list. One line per card, with the clause and your recommendation (build or leave out). Ian answers it.
- Do not build cards in this task.

## Every build task
- Run the checks BRIEF.md lists, plus:
  - `go test ./internal/game/` for engine changes;
  - the protocol, ws and lobby tests for wire changes;
  - client tests (`npm ci && npx vitest run` in client/) for client changes;
  - labels per AGENTS.md §5.
- Open one PR, `--base develop`, in the AGENTS.md §4 format:
  - "Relates to #5948", and Closes for any seam issue you finish;
  - trailers `Sprint: S82 — Card coverage: sets outside Standard, newest first`, `Issue: #5948`.
- `gh pr merge --squash --auto`. Wait for CI.
  - If develop moves and the PR conflicts: merge develop in. Plain push, never force.
  - Fix real failures. Rerun failures that are only infrastructure. The flake `TestNetheresePuzzleWardIgnoredRollDoesNotTrigger` gets one rerun.
  - Stop after 8 hours.
- If a card can't be built truthfully, leave it out and say why.

Final message: the PR number and state, what was built, and what was left out and why.

Nobody will answer you mid-run. Never end your turn while a build, test or CI run is still going.
