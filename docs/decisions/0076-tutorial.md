# ADR 0076 — The tutorial: a scripted practice game that teaches the client, not the rules

**Status:** Accepted · 2026-09-19 · unscheduled (no onboarding sprint exists; see
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

The canvas is *CMD CTRL Tutorial*: the table mid-tutorial, the coach card in all
six states, the spotlight and anchoring mechanics, the entry points, and the
eleven steps with anchors and completion predicates.

## 2. Decision

### 2.1 Eleven steps, ordered by when a first turn needs them

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
| 8 | Move the turn along | phase widget | `turn.step` changed | snapshot |
| 9 | Watch the bot | attention strip | `turn.active_seat` back to you | snapshot |
| 10 | Attack | creature row + opponent medallion | a controlled creature has `attacking_target` | snapshot |
| 11 | That is the whole interface | — | button | button |

Step 7 is the one that earns the feature.

### 2.2 The practice table

A private game with one human seat and one `random`-tier bot, both on a fixed
prebuilt deck, created by a dedicated route rather than the ordinary lobby flow
and **not listed in the lobby**.

- **The bot's deck is deliberately slow** — lands and small bodies, no removal,
  no evasion. A practice bot that kills the player during the tutorial is a
  bug, and the cheapest place to fix it is the decklist rather than the policy.
- **Settings are forced for the duration and restored on exit**: `strictMana`
  off, `autoPassPriority` off (so the player actually sees priority reach
  them), `tableLayout` quadrant, `cardSize` medium. The player's own values are
  captured on entry and written back on exit, including an exit by navigation.
- **There is no resume.** Leaving abandons the table and releases the bot seat.
  A half-finished practice game is worth less than a clean restart, and keeping
  one alive means holding a game id and a bot runner for an account that may
  never come back.

### 2.3 The coach card and the scrim

- The coach card docks **bottom-left**. The phase widget owns bottom-right and
  the attention strip owns the top, so that is the only corner that is free at
  every step.
- The scrim is **one element**: a transparent rect with a 9999px spread shadow,
  so the hole is the rect and everything else darkens.
- The scrim is **`pointer-events: none`**. Dimming is a suggestion, never a
  lock — the board stays fully playable at every step.
- **Every step is skippable, and skipping is silent.** No step blocks a click,
  warns, or asks twice. A player who skips six steps and plays the game has
  learned more than one who quit at step three.

### 2.4 Anchors are the e2e selectors, deliberately

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

### 2.6 Entry

An **offer, not a gate**. A card in the lobby for an account with no finished
games, with a permanent dismiss on "Not now" — the second time someone sees a
tutorial they did not ask for, they stop reading the lobby. Replay lives in
Settings → Advanced.

"Has this account finished a game" properly belongs in a user row (ADR 0051,
S34). Before that lands, it is a field in the settings store, which means a
`SETTINGS_VERSION` bump and a migration, and it means the offer reappears on a
new browser. That is acceptable for a first release and should be moved to the
user row when S34 arrives rather than left behind — see §6.

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

1. **The event bus.** `tutorialBus.ts` plus the three emits in `Card.svelte`,
   `Hand.svelte` and `BattlefieldRow.svelte`, with unit tests. Independent of
   everything else and mergeable on its own.
2. **The practice table.** The create route, the fixed tutorial decks, and
   forced-settings capture and restore across every exit path.
3. **Walking skeleton.** Coach card, scrim, the anchoring resolver and its
   missing-anchor fallback, wired with steps 1 and 11 only — a tutorial you can
   start and finish that teaches nothing yet.
4. **The nine middle steps**, their copy, and the predicates.
5. **Entry.** Lobby offer, permanent dismiss, Settings → Advanced replay, the
   settings field and migration.
6. **e2e.** A Playwright spec that walks all eleven steps, which is also the
   regression test for the anchor contract in §2.4.

1 is independent. 3 depends on 1. 4 depends on 2 and 3. 5 and 6 come last.

## 6. Open questions

- **The "finished a game" signal before S34.** Settings-store field now and
  migrate to the user row later, or hold the entry point until ADR 0051's rows
  exist and ship 1–4 without a lobby offer? The second is tidier and delays the
  feature reaching anyone.
- **Does the practice bot need its own tier?** A `tutorial` tier that never
  attacks is more predictable than `random` plus a slow deck, but it is a new
  policy to maintain for one consumer.
- **Should step 7 also teach `adminOverrides`?** Right-click means two
  different things depending on that setting, and the tutorial currently
  teaches only the default. Mentioning both in one step is the kind of
  completeness that makes a tutorial worse.
