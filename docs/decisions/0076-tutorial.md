# ADR 0076 — The tutorial: a scripted practice game that teaches the client, not the rules

**Status:** Accepted · 2026-09-19 · its own sprint, number unassigned (see
[#1073](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1073))
**Issues:** [#1073](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1073) (tracker)
**Numbering:** on 2026-09-19 `docs/decisions/` on `develop` was listed and the
highest number present was `0075-table-settings-and-host-controls.md`, so this
ADR takes **0076**. 0005, 0024, 0029 and 0030 stay permanently unused per
AGENTS.md §4. **This is weaker evidence than 0075's**, which checked every
remote and local branch after a `git fetch --all --prune`: this session has no
git credentials (base-station and the cloud container both fail
`git ls-remote`), so it could only list `develop` through the API and confirm
that the one open PR, [#1072](https://github.com/krakenhavoc/cmd_and_ctrl/pull/1072),
touches no ADR. An unpushed branch could hold a 0076. Re-check before merging.
**Builds on:** [ADR 0033](0033-ai-bot-seat.md) (the practice bot, its tiers and
pacing), [ADR 0047](0047-keyboard-shortcuts.md) (the keymap and its overlay,
which this hands off to), [ADR 0028](0028-admin-context-menu.md) (what
right-click does when `adminOverrides` is on)
**Relates to:** [ADR 0051](0051-user-database.md) — the "has this account
finished a game" signal the entry point needs properly belongs in a user row;
§2.6 says what to do before that lands.

---

## 1. Context

There is no onboarding of any kind. A new player lands in the lobby, joins a
table, and meets a client with several affordances that have no visible surface.
The ones that cannot be deduced from looking at the table:

- **Right-click opens a card's abilities.** `Card.svelte`'s `handleContextMenu`
  raises `ManaAbilityMenu` for a permanent's mana and activated abilities.
  Nothing on the card face advertises it, and with
  `gameplay.adminOverrides` on (off by default, #170) the same gesture opens a
  different, larger menu instead. A player who never right-clicks cannot
  activate anything.
- **The hand is clipped to a peek.** It shows the top 62% of each card and
  lifts on hover (#858). Someone who does not hover cannot read their own
  cards' rules text.
- **Lands are piles.** Same-name lands collapse under a count badge and fan out
  on hover (#858), so "click a specific Forest" is a gesture, not a click.
- **The phase widget is three different things.** `next` advances a step,
  `hold` keeps priority so you can answer your own spell (#323's
  `autoPassOwnStack` is on by default, so without `hold` your own stack does
  not stop), and `autopass` is a mode with a safety belt. Per-step stops live
  in Settings (S13).
- **The attention strip is where the stack lives**, and Counter and Pass sit on
  its top item rather than anywhere near the cards.

> **Amended by [ADR 0111](0111-action-dock.md) §1 and §7 (S56 PR 2, #1958):** the
> phase widget is now the header of the action dock, `region "actions"`, in the
> screen's bottom-right corner, with `next`, Pass turn, `hold`, `autopass` and
> `bluff` in it. Counter sits on the stack's top item; Pass is the dock's `next`,
> and the stack card no longer has a Pass or a hold of its own.
- **Mana is implicit.** Clicking a land taps it; casting auto-taps, and
  `strictMana` is off by default, so a new player never learns they were
  supposed to pay for anything.
- **Combat is two clicks in two different places**: your creature, then the
  opponent's portrait.
- **The rail piles do different things.** Library draws; graveyard and exile
  open the zone browser.

What already exists to build on: the bot seat (ADR 0033, `internal/aiseat`, the
`random` tier), `prebuiltDecks.ts`, the settings store and its migration chain,
and the `aria-label` contract `tests-e2e/tests/board-layout.spec.ts` asserts on.

The owner decided the following on 2026-09-19:

| Question | Decision |
|---|---|
| Who is it for | **Players who know Magic and are new to this client.** |
| Does it teach the rules | **No.** No phases, no stack, no priority as a concept, no commander tax, no 40 life or 21 commander damage. |
| Form | **A scripted game against a practice bot.** |
| First artefact | **A design canvas**, before implementation. |
| Scheduling | **Its own sprint, all six sub-PRs together** rather than landing the event bus early into S35. |

The canvas is *CMD CTRL Tutorial*: the table mid-tutorial, the coach card in all
six states, the spotlight and anchoring mechanics, the entry points, and the
eleven steps with anchors and completion predicates.

## 2. Decision

### 2.1 Eleven steps, ordered by when a first turn needs them

> **Amended by [ADR 0125](0125-a-walkthrough-that-keeps-up.md) §5 (S65, #2313):** fourteen steps, not eleven. The opening
> roll is step 2, the stack is step 8 and the commander is step 10, and a step's
> `teaches` marks the hints it covers as seen when it completes. The table below
> is the original eleven; `client/src/lib/tutorialSteps.ts` is the current list.

Each step teaches exactly one thing the UI does not advertise. The ordering is
the turn itself, so every lesson arrives at the moment it is needed rather than
as a list up front.

| # | Step | Anchor | Advances when | Source |
|---|---|---|---|---|
| 1 | A five-minute practice game | — | button | button |
| 2 | Read your hand | `[aria-label="your hand"]` | hover ≥ 600ms | **DOM** |
| 3 | Play a land | `[aria-label="your hand"]` | controlled Land count +1 | snapshot |
| 4 | Lands stack into piles | `[aria-label="lands"]` | hover ≥ 600ms | **DOM** |
| 5 | Tap a land for mana | `[aria-label="lands"]` | `mana_pool.length > 0` | snapshot |
| 6 | Cast a creature | `[aria-label="your hand"]` | controlled creature count +1 | snapshot |
| 7 | **Abilities live on right-click** | card instance id | ability menu opened | **event** |
| 8 | Move the turn along | `region "actions"` (the action dock, ADR 0111) | `turn.step` changed | snapshot |
| 9 | Watch the bot | `region "attention"` (the attention strip, ADR 0111 §4) | `turn.active_seat` back to you | snapshot |
| 10 | Attack | creature row + opponent medallion | a controlled creature has `attacking_target` | snapshot |
| 11 | That is the whole interface | — | button | button |

Step 7 is the one that earns the feature.

**Step 7 teaches the default gesture only.** With `gameplay.adminOverrides` on
(ADR 0028) right-click opens a different and larger menu, but that setting is
off by default, and a player who turned it on went looking in Settings and
already knows what it does. Attaching a caveat about an unmet setting to the
step that already carries the most new information is the kind of completeness
that makes a tutorial worse.

### 2.2 The practice table

> **Amended by [ADR 0125](0125-a-walkthrough-that-keeps-up.md) §5.2 (S65, #2313):** the practice table opens the opening roll
> (`StartWithOpeningRoll`), so the player rolls as at a real table. A practice bot
> that wins it hands the first turn to the player (`aiseat.SeatSpec.FirstTurnTo`),
> so seat 0 still takes turn one. The bot stays on the `random` tier.

A private game with one human seat and one `random`-tier bot, both on a fixed
prebuilt deck, created by a dedicated route rather than the ordinary lobby flow
and **not listed in the lobby**.

- **The bot stays on the `random` tier.** A dedicated `tutorial` tier that never
  attacks would be more predictable, but it is a new policy surface in `aiseat`
  maintained for exactly one caller, and it would diverge from what a real
  opponent does — the player's second game should not contradict their first.
  The decklist carries the difficulty instead. Revisit only if playtesting the
  walking skeleton shows the bot doing something a new player misreads.
- **The bot's deck is deliberately slow** — lands and small bodies, no removal,
  no evasion. A practice bot that kills the player during the tutorial is a
  bug, and the cheapest place to fix it is the decklist rather than the policy.
- **Settings are forced for the duration and restored on exit**: `strictMana`
  off, `autoPassPriority` off (so the player actually sees priority reach
  them), `tableLayout` quadrant, `cardSize` medium. The player's own values are
  captured on entry and written back on exit, including an exit by navigation.

  > **Amended by [ADR 0118](0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md)
  > §1 (S59 PR 4, #2188):** the practice table forces `strictMana` **on**, so
  > the tutorial teaches the table the player will meet, where a click taps the
  > lands. The other three forced values and the restore are unchanged. Step 1's
  > copy is now "You are seated against a practice bot. Click a card your lands
  > can pay for and the game taps them for you. You can undo, so nothing here
  > can go wrong." The Context line "`strictMana` is off by default" is history:
  > strict payment is the default since settings v19.
- **There is no resume.** Leaving abandons the table and releases the bot seat.
  A half-finished practice game is worth less than a clean restart, and keeping
  one alive means holding a game id and a bot runner for an account that may
  never come back.

### 2.3 The coach card and the scrim

- The coach card docks **bottom-left**. The phase widget owns bottom-right and
  the attention strip owns the top, so that is the only corner that is free at
  every step.

  > **Amended by [ADR 0111](0111-action-dock.md) §4 (S56 PR 2, #1958):** it is
  > the action dock (`region "actions"`) that owns the bottom-right corner now.
  > The coach card stays bottom-left for the same reason.

  > **Amended 2026-10-02 (sub-PR 3, #1079):** bottom-left is kept clear by the
  > layout now; the corner is not simply empty. The coach card publishes its
  > live size as `--coach-w` / `--coach-h`, the same way the dock publishes
  > `--dock-w` / `--dock-h`.
  >
  > - **Desktop:** while the card shows, the viewer's own panel keeps an
  >   empty cell that size at the left of its bottom row (`.coach-spacer`,
  >   the twin of `.dock-spacer`), and the hand centres in what is left.
  > - **Phone (≤599px):** the card is a strip stacked on the dock bar. The
  >   play area's bottom padding grows by the strip's height, and the strip
  >   folds to one line, like the dock's sheets.
  > - **No tutorial running:** the layout is unchanged. There is no spacer,
  >   no class and no padding.
  >
  > Why: ADR 0111's dock takes the bottom row's right end, and the hand fan
  > is centred in the rest.
  >
  > - At **1280×800** the free strip at the left was about 150px. The
  >   seven-card fan ran from x 177 to 762, so any readable card covered its
  >   left two cards and the hand-sort button. The card takes clicks, so it
  >   would have blocked them.
  > - At **390×844** there was no free corner at all: the hand sits directly
  >   on the full-width dock bar.
  > - At 1920×1080 the corner was free.
  >
  > The owner chose this over three alternatives:
  >
  > - moving the card to another corner whenever it collides;
  > - putting the coach inside the dock;
  > - shipping desktop-only and hiding the tutorial on phones.
  >
  > The card is sized for 1280: `clamp(280px, 24vw, 340px)`, about 307px
  > there. The fan still fits beside it.

  > **Amended 2026-10-02 (#1081 follow-up):** the desktop cell takes the
  > card's **width** from the bottom row and no height. It stretches to the
  > row the hand and the dock's cell already make.
  >
  > - Taking the card's height as well made the bottom row as tall as the
  >   card, about 194px against the dock's 154px. The battlefield rows paid
  >   for it: the creature row lost 40px at both 1280×800 (128px to 88px)
  >   and 1920×1080 (260px to 220px).
  > - The card is now taller than the row and rises above it on the left.
  >   That is over the battlefield's left margin, which is empty because
  >   both battlefield rows centre their cards.
  > - `tests-e2e/tests/tutorial-layout.spec.ts` measures it on a practice
  >   table at both sizes. Every row is the same height with the coach up as
  >   with it gone, and the hand starts right of the card.
- The scrim is **one element**: a transparent rect with a 9999px spread shadow,
  so the hole is the rect and everything else darkens.
- The scrim is **`pointer-events: none`**. Dimming is a suggestion, never a
  lock — the board stays fully playable at every step.
- **Every step is skippable, and skipping is silent.** No step blocks a click,
  warns, or asks twice. A player who skips six steps and plays the game has
  learned more than one who quit at step three.

### 2.4 Anchors are the e2e selectors, deliberately

> **Amended by [ADR 0125](0125-a-walkthrough-that-keeps-up.md) §2 (S65, #2313):** anchors come from the label registry,
> `client/src/lib/labels.ts` (`L.<key>`, listed in `docs/labels.md`). A unit test,
> `labels.test.ts`, fails any PR whose tutorial step or hint anchors to a label no
> component renders, so a rename is caught on the PR, not only by the nightly.

Every spotlight anchors to an `aria-label` that `board-layout.spec.ts` already
asserts on. A refactor that moves the hand or renames a zone therefore breaks a
Playwright test **before** it breaks a new player's first session. This is the
only mechanism that keeps a tutorial honest across six months of UI churn, and
it means those labels stop being a test detail and become a user-facing
contract — AGENTS.md should say so.

A step whose anchor cannot be found **advances itself and logs**, rather than
spotlighting empty space or waiting on a predicate that can never fire. Losing
one step is recoverable; a wedged tutorial on a first session is not.

### 2.5 Step completion needs a small client event bus

**Only six of the eleven steps can read their own completion out of the game
snapshot.** Steps 2 and 4 are hover. Step 7 is a card-local menu opening —
`Card.svelte`'s `manaMenuOpen` is component state and never reaches the wire.
Steps 1 and 11 are buttons. This is the one piece of plumbing the feature
cannot avoid, and it is why the decision belongs here rather than in step
content.

The bus is deliberately tiny: a Svelte store in `client/src/lib/tutorialBus.ts`
exposing `emit(name)` and a subscription, carrying **three** event names:

- `ability-menu-opened` — emitted by `Card.svelte` where it sets `manaMenuOpen`
- `hand-hovered` — emitted by `Hand.svelte`
- `pile-hovered` — emitted by `BattlefieldRow.svelte` on a `.pile.multi`

Three components gain one line each. It is **not** a general telemetry or
analytics bus, and it must not grow into one: anything that can be read from
the snapshot is read from the snapshot. When the tutorial is not running the
emits are a no-op store write.

The bus is independently mergeable but **ships with its consumer** rather than
early into another sprint (§1, scheduling). Landing a bus with no caller invites
exactly the unrelated uses §2.5 exists to prevent.

### 2.6 Entry

> **Amended by [ADR 0125](0125-a-walkthrough-that-keeps-up.md) §4 and §6 (S65, #2313):** the lobby offer is the
> `lobby.practice` hint, offered once with a permanent "Not now". The settings
> field is the synced `help` settings group (`seen`, `tipsOff`), which discharges
> the obligation below to move it to the user row. Replay lives in the Help menu
> and in Settings → Advanced → Tips and the tutorial.

An **offer, not a gate**. A card in the lobby for an account with no finished
games, with a permanent dismiss on "Not now" — the second time someone sees a
tutorial they did not ask for, they stop reading the lobby. Replay lives in
Settings → Advanced.

"Has this account finished a game" properly belongs in a user row (ADR 0051,
S34). It does not exist yet. **Decided 2026-09-19: ship the settings-store
field now**, with a `SETTINGS_VERSION` bump and migration, and move it to the
user row when S34 lands. The cost is that the offer reappears on a new browser
and that there is a migration to remember; both are cheap next to the
alternative, which was to hold the entry point and ship sub-PRs 1–4 with no way
to discover the tutorial. A tutorial nobody can find does nothing, and S34 is a
full sprint cycle away.

The migration is a **tracked obligation, not a nice-to-have**: when ADR 0051's
rows land, the field moves and the settings entry is retired. Sub-PR 5 records
it.

## 3. Consequences

- The tutorial **inherits the bot runner's reliability**. If `aiseat` stalls,
  the tutorial stalls at step 9 in front of exactly the audience least equipped
  to understand why. The mitigation is step 9's skip plus a timeout that
  advances on its own; the real fix is bot reliability, and a tutorial is a
  good forcing function for it.
- Three client components gain an emit, and the bus is new client state. Small,
  but it is a new category of thing in this codebase and it will attract
  unrelated uses; §2.5's restriction needs to survive review.
- The `aria-label` contract becomes load-bearing for a shipped feature rather
  than only for tests. Worth an AGENTS.md line so nobody renames one casually.
- `prebuiltDecks.ts` gains a fixed tutorial pair. They must stay catalog-only
  cards, or the engine cannot actually play the game the tutorial scripts.
- A settings field and schema bump, retired when ADR 0051's user rows land.
- Forced settings must be restored on **every** exit path, including a closed
  tab. A player who abandons the tutorial and finds `strictMana` silently off
  in their real games will not connect the two.

## 4. Alternatives considered

- **A fully scripted sandbox with no bot.** Immune to engine randomness and to
  `aiseat` stalling, and every step lands identically. Rejected because it is
  much more content to author and because it teaches a table that does not
  behave like a real one — the player's second game would contradict their
  first.
- **Coach marks over the player's first real game.** Cheapest, and needs no
  authored game. Rejected because a real game cannot be scripted: a first real
  game here is usually four players and chaotic, the steps would fire in an
  arbitrary order, and several would not apply at all.
- **A static help page or reference overlay.** Discoverable only if you go
  looking. We probably still want one, and `?` (ADR 0047) already half-covers
  it, but it does not solve "the player does not know right-click exists",
  because you cannot look up a gesture you do not suspect.
- **Teach Magic as well, as a second track.** A different and much larger
  product. Arena does it in twenty minutes with voice, art and a bespoke
  scripted opponent, and we would do it worse. The audience decision in §1
  closes this.
- **Advance every step on a Next button.** Removes the bus entirely. Rejected
  because a tutorial you can click through without doing anything teaches
  nothing; the gestures are the content.

## 5. Implementation plan (sub-PRs)

> **Amended by [ADR 0125](0125-a-walkthrough-that-keeps-up.md) (S65, #2313):** sub-PRs 5 and 6 were delivered there. Sub-PR 5
> (entry) is [ADR 0125](0125-a-walkthrough-that-keeps-up.md) §6 (#2345); sub-PR 6 (e2e) is `tutorial-walk.spec.ts` and
> `hints.spec.ts` (#2393, closing #1085), which walk all fourteen steps nightly and
> fail on any `has no anchor` console line.

1. **The event bus.** `tutorialBus.ts` plus the three emits in `Card.svelte`,
   `Hand.svelte` and `BattlefieldRow.svelte`, with unit tests. Technically
   independent, but it ships in this sprint with the rest (§2.5).
2. **The practice table.** The create route, the fixed tutorial decks, and
   forced-settings capture and restore across every exit path.
3. **Walking skeleton.** Coach card, scrim, the anchoring resolver and its
   missing-anchor fallback, wired with steps 1 and 11 only — a tutorial you can
   start and finish that teaches nothing yet.
4. **The nine middle steps**, their copy, and the predicates.
5. **Entry.** Lobby offer, permanent dismiss, Settings → Advanced replay, the
   settings field and migration, and the tracked obligation to move it to the
   user row at S34.
6. **e2e.** A Playwright spec that walks all eleven steps, which is also the
   regression test for the anchor contract in §2.4.

1 is independent. 3 depends on 1. 4 depends on 2 and 3. 5 and 6 come last.

## 6. Open questions

The three questions this ADR was drafted with were put to the owner on
2026-09-19 and answered the same day. They are folded into §2 above and
recorded here for the trail:

- **The "finished a game" signal before S34** → the settings-store field now,
  migrated to ADR 0051's user row when S34 lands. §2.6.
- **A dedicated `tutorial` bot tier** → no. `random` plus the slow decklist,
  revisited only if the walking skeleton shows the bot doing something a new
  player misreads. §2.2.
- **`adminOverrides` in step 7** → no. The tutorial teaches the default gesture
  only. §2.1.

**Still open:** which sprint. The owner chose a dedicated sprint carrying all
six sub-PRs rather than landing the bus early, but the number and the slot are
not assigned. The tracker is #1073.

## 7. Implementation notes

**Sub-PR 2, the practice table (#1078, S54).** The code moved between this ADR
and the build, so §2.2 and §3 map onto it as follows:

- **The decks are server-side.** `prebuiltDecks.ts` is now only the picker's
  types and wording; the pre-built decks live in `server/internal/decks`. The
  tutorial pair is `decks/tutorial.go` (*First Steps* and *Practice Partner*,
  both mono-green), kept outside the pickers' registry so neither picker offers
  them, and held by the same per-deck tests as the four picker decks plus their
  own: every card `CompletenessFull`, no card that targets, and for the bot no
  evasion keyword, and — against the Scryfall dump — power three or less and no
  removal text.
- **The create route** is `POST /games/practice`, open to any session (the lobby's
  `POST /games` is admin-only, which a tutorial for new players cannot be). It
  seats the caller and the bot and **starts the game with the player at seat 0
  taking turn one**, skipping the opening roll, because §2.1's steps follow the
  player's first turn.
- **"No resume" is enforced server-side** as well as by the client: a practice
  table has no rows in the S34 database and its room writes no restore point,
  so a restart ends it; one per person (a second replaces the first); at most
  eight at once; and a table with no commit for 20 minutes is reaped, which is
  the server's answer to a closed tab that never said so.
- **The session is swapped and restored as well as the settings.** Opening a
  practice table mints a player session for its seat, as any join does, which
  replaces whatever session the player held. Leaving puts that back alongside
  the four forced settings; `POST /games/{id}/practice/leave` also resets the
  session cookie, since the server reads the cookie first.
- **The entry is `#/practice`**, a route that opens a table and moves on to it.
  No link to it is added here; the lobby offer and Settings' replay (sub-PR 5)
  point at it.

**Sub-PR 3, the walking skeleton (#1079, S54).**

- **Where the parts are.**
  - `client/src/lib/tutorial.ts` is the step machine. It is pure and has
    the six coach states.
  - `tutorialAnchor.ts` is the anchor resolver.
  - `tutorialSteps.ts` is the script. It has steps 1 and 11 only, and
    sub-PR 4 inserts the nine middle steps between them.
  - `components/tutorial/` holds `CoachCard.svelte`,
    `TutorialScrim.svelte` and `TutorialCoach.svelte`.
  - `Game.svelte` mounts the coach on this tab's practice table only
    (`practiceTable`).
- **Anchors are scoped.** Every opponent's panel carries the same
  `"lands"` and `"creatures"` lists as the viewer's. So a battlefield
  anchor is `{ label: "lands", within: "your board" }`, the scoping
  `board-layout.spec.ts` already uses. Step 8 anchors to
  `region "actions"` and step 9 to `region "attention"` (ADR 0111 §10).
  Step 7 anchors to `data-instance-id`.
- **"Missing" includes no area.** An anchor counts as missing when no
  element matches, or when the elements that match have no area: an empty
  list is 0px tall.
  - The step waits 1.5s for the element to render, then advances itself
    and logs.
  - An anchor that disappears mid-step is treated the same way.
- **What the buttons do.** Skip tutorial and Finish hide the coach and
  leave the player at the practice table, which is still a game. Replay
  opens a fresh table (`#/practice`).

**Sub-PR 4, the nine middle steps (#1081, S54).**

- **Where they are.** Steps 2–10 are in `tutorialSteps.ts`, each with its
  copy, anchor and predicate. The step machine gained five optional
  fields; steps 1 and 11 use none of them.
  - `first`, a detour. While the board does not allow the step yet, the
    card says what to do first and points at it. It never blocks.
  - `cannot`. The board can no longer produce the step's action, so the
    step advances itself and logs, as for a missing anchor (§2.4).
  - `hover`, for steps 2 and 4 (below).
  - An anchor or a status line computed from the board.
  - A timeout on an action step, not only on a watch step.
- **The copy follows today's client.** Undo is in the dock beside
  autopass (ADR 0111 PR 3), not in the ⋯ menu.
- **What the practice table does that §2.1 did not foresee.**
  - The game opens in the player's upkeep. With `autoPassPriority` forced
    off (§2.2), nothing moves until they press `next`. A land and a
    creature need a main phase, so steps 3 and 6 open on a detour,
    "First, your main phase", that points at the dock.
  - The same forced setting holds the player's own spell on the stack
    until they pass. It also gives them priority at every step of the
    bot's turn, so the bot's turn waits on them.
    - Step 6 detours to "Now let it resolve".
    - Step 9 teaches the dock's autopass toggle (the owner's choice,
      2026-10-02, over a step that asks for `next` at every pause). It
      spotlights the toggle, and completes once autopass is on and the
      bot's turn is running by itself. Step 10 then watches the bot's turn
      ("Watch the bot play", spotlighting the bot's board) until it is
      the player's own.
    - Autopass is session state, not the forced `autoPassPriority`
      setting, which it leaves alone. Steps 3 and 6 come before it.
    - Its safety belt clears the toggle when the player's own main phase
      comes round (`autopassDecision.ts` rule 2). Turned on there, it
      clears at once, so step 9 detours out of the main phase first.
    - Step 10 finds the toggle off, unless the player has set autopass
      to outlive their main phase (`autopassPersistThroughTurns`, not
      forced). Then step 10 asks them to switch it off before combat.
    - Step 9 still advances on its own after 90 seconds (§3), and gives
      up when the bot's turn has been pressed through by hand.
  - On turn one the player has one land, and one Forest is not a pile.
    Step 4 completes on a 600ms rest on the lands row, pile or not. Its
    hint says the next Forest joins this one.
  - `strictMana` off does not mean every spell can be cast. The server's
    move list still leaves out a spell the player cannot pay for, and the
    hand dims it. Step 6's copy points at the lit cards. When no creature
    in hand is castable, the step gives up and logs.
  - The coach's cell squeezed the viewer's creature row. §2.3's second
    amendment fixed the layout, so step 7 points at the newest creature
    with a menu again, then at a permanent with one.
  - A creature card can still be taller than its row on a short panel,
    and the row scrolls. The hole goes round the part of an anchor its
    scrolling ancestors show.
  - Step 1 said "Mana is not enforced". It now says what the table does:
    the hand offers only what your mana could pay for, but a cast spends
    your pool and waives the rest, so lands never need tapping first.
- **Hover is timed on the anchor.** The bus says the pointer arrived, not
  that it stayed, and a pointer crossing the hand on its way to the dock
  is not reading it.
  - The coach checks `:hover` on the step's anchor on the poll that
    already measures it, and completes the step after 600ms of rest.
  - On a device with no hover, the step's bus event is the whole
    gesture. A step with no event to wait for moves on after 12 seconds.
- **Step 10's opponent anchor is the portrait's `data-seat-id`**, the
  attribute CombatArrows already anchors to. The portrait's aria-label
  carries the life total, so it cannot be the anchor.
