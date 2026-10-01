# ADR 0096 — The monarch from a card effect

**Status:** Accepted · 2026-09-28 · S40 — S30's tail: protection, regeneration, emblems, winning by effect
**Issue:** [#1722](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1722) (relates to #1720, tracker [#883](https://github.com/krakenhavoc/cmd_and_ctrl/issues/883))
**Numbering:** swept with the AGENTS.md §4 check on 2026-09-28: `git fetch --all --prune`, then
every `docs/decisions/` file name in the history of every remote branch (37 heads). The highest
number present anywhere is **0095** (`0095-deck-coverage-and-deck-requests.md`), and no open
issue's claim comment reserves 0096. Reserved on #1722 before any code was written.
**Why a new ADR:** no ADR owns the monarch. The mechanic arrived with #375 / #899 as a bug fix
(`game/monarch.go`, the two CR 725.2 triggered abilities driven by a listener); ADR 0041 mentions
it only as two keyed bodies (`monarch/crown`, `monarch/draw`).
**Builds on:** [ADR 0026](0026-delayed-triggers.md) and its 2026-09-18 amendment (event-conditioned
delayed triggers, #663), [ADR 0041](0041-game-persistence.md) (keyed bodies and conditions, the
effect-key ledger, the snapshot shape record), [ADR 0063](0063-durations-and-control.md) (the
`Duration` model), [ADR 0074](0074-triggered-mana-abilities.md) (triggered mana abilities),
[ADR 0045](0045-combat-restrictions.md)'s 2026-09-28 amendment (block capacity, #1706).

---

## Context

The monarch (CR 725) is a designation, two inherent triggered abilities (CR 725.2) and a rule for
when its holder leaves (CR 725.4). All three were in the engine. What was missing was the one
thing every monarch card prints: **"you become the monarch."**

`Game.SetMonarch` is the sandbox action behind `set_monarch`, and it takes `g.mu` itself. A
resolving spell's `OnResolve` and a trigger's `Effect` already hold that lock, so calling it from a
card deadlocks. There was no locked-context twin, so no catalog card could use the monarch —
Palace Jailer, Queen Marchesa, the six Courts and the rest were all out of reach, and #1720 skipped
Entourage of Trest for exactly this reason.

Two more things a card needs were also missing:

- **An event.** `SetMonarch` wrote `Game.Monarch` and emitted nothing, so "whenever you become the
  monarch" (Custodi Lich) and "whenever an opponent becomes the monarch" (Knights of the Black
  Rose) had nothing to watch, and a static that reads the crown (Entourage of Trest's extra block)
  had nothing to invalidate on.
- **The crown as the turn began.** Knights of the Black Rose's "if you were the monarch as the
  turn began" is a fact about a moment that has passed.

## Decision

### 1. One write, and it announces itself

`becomeMonarchLocked` (`game/monarch.go`) is the only write to `Game.Monarch`. Every route goes
through it: the card effect (`SetMonarchForEffect`), the CR 725.2 combat-damage steal
(`monarch/crown`), the CR 725.4 hand-on (`monarchLeftTheGameLocked`) and the sandbox's manual
`SetMonarch`. It emits **`EventMonarchChanged`**, with `Actor` = the new monarch (`uuid.Nil` when
the designation is cleared) and `Target` = the previous one.

It emits **only on a real change**. CR 725.3 has one monarch, and "as a player becomes the
monarch" is a change of holder: a player who is already the monarch and is told to become it again
does not become it. So a second Court entering while you wear the crown does not trigger your
Custodi Lich, and nothing on the card side has to guard for it.

The manual set sharing the write is deliberate. A table that corrects the crown by hand gets the
same triggers and the same static re-read as the card would have produced; the sandbox action is a
way to reach a game state, not a way around what that state means. `SetMonarch` now also refuses a
seat that has left the game (`ErrPlayerNotFound`), which `becomeMonarchLocked` would otherwise have
ignored silently.

### 2. `SetMonarchForEffect`

`(*Game).SetMonarchForEffect(player)` is the `effect_api.go`-convention entry point: the caller
holds `g.mu` in write mode. It returns `ErrPlayerNotFound` for an ID that names no seat (a caller
bug) and does nothing for a seat that has left (CR 800.4a — the effect does as much as it can).

### 3. The layer pass invalidates on it

`layerVersionBump` bumps on `EventMonarchChanged`, unconditionally. "As long as you're the monarch"
is a layer input that no permanent moving stands in for — the crown changes hands with the board
untouched — and it changes a handful of times a game, so a gate would save nothing.

### 4. The crown as the turn began is a stamp

`TurnTally.MonarchAtStart`, set by `resetTurnTallyLocked` from `Game.Monarch` as each turn begins
and read through `Game.MonarchAsTurnBegan`. It rides the turn tally, so `Clone`, the snapshot and
undo carry it with no new plumbing. The snapshot shape grew by one additive field
(`turnTally.monarchAtStart`, recorded in `testdata/snapshot_shape/v7.txt`, no version bump): a file
written before it restores with nobody stamped, which only makes Knights of the Black Rose's
intervening "if" false for the rest of that turn — weaker, never stronger.

### 5. The card vocabulary

In `cards/effects/monarch.go`:

| Shape | Printed text |
|---|---|
| `BecomeTheMonarch{}` | "you become the monarch" (Player defaults to the controller) |
| `WhenThisEntersYouBecomeTheMonarch(name)` | the ETB every monarch permanent prints |
| `YouBecameTheMonarch` / `WheneverYouBecomeTheMonarch` | "Whenever you become the monarch" (Custodi Lich) |
| `AnOpponentBecameTheMonarch` | "Whenever an opponent becomes the monarch" (Knights of the Black Rose) |
| `YoureTheMonarch(g, you)` | "if / as long as / while you're the monarch" |
| `AnOpponentIsTheMonarch(g, you)` | "if an opponent is the monarch" (Queen Marchesa) |

Where the condition is read follows the card. The Courts' "If you're the monarch, … instead" is a
clause of the effect, read at resolution (CR 608.2), not an intervening "if" — a Court whose
controller lost the crown in response still pays the small half. Queen Marchesa's and Knights of
the Black Rose's are intervening "if"s (CR 603.4), checked at the trigger and again at resolution.
Entourage of Trest's is a layer-6 static (§3). Regal Behemoth's is the condition of a triggered mana
ability (ADR 0074), read as the land is tapped.

### 6. "Until an opponent becomes the monarch" is an event-conditioned delayed trigger

Palace Jailer exiles a creature "until an opponent becomes the monarch". The issue asked whether
the ADR 0063 duration model can express that. It cannot, and should not: a `Duration` says when a
*continuous effect* stops applying, and this is a one-shot exile with an "until" (CR 610.3), whose
end is a second one-shot that returns the card.

The shape that fits is the one ADR 0026's 2026-09-18 amendment built for "when you next cast": a
delayed trigger with an **event** condition. The exile schedules one with `On:
[EventMonarchChanged]`, the keyed condition `monarch/an-opponent-became` (new monarch ≠ the
trigger's controller, who is the player that controlled the Jailer's ability, CR 603.7d), the
exiled card as its payload, flicker's `flicker/return-exiled-to-owners` body, and an **Indefinite**
`Duration` — the "this turn" default every other event-conditioned trigger takes would silently
release nothing and drop the trigger at cleanup. Keyed, so a table with a creature jailed is a
restore point.

The return lives on the game, not on the Jailer, because the event is about the crown: a Jailer
that dies keeps its creature exiled, and one bounced and recast exiles a second creature without
releasing the first. An opponent who already wears the crown when the exile resolves releases
nothing — they did not *become* the monarch after it.

### 7. The log stays silent about it

`EventMonarchChanged` is on the log gate's silent list
(`protocol/log_event_kind_gate_test.go`): the line that moved the crown already says so — the
`LogResolve` of "Palace Jailer — you become the monarch" or "the monarch — Bob becomes the
monarch", or the `LogEliminated` of the monarch who left — and `GameView.monarch` carries the
holder. No wire change.

## Consequences

- Twelve catalog cards: Court of Grace, Court of Ambition, Court of Cunning, Court of Bounty,
  Court of Ire, Queen Marchesa, Palace Jailer, Entourage of Trest, Custodi Lich, Knights of the
  Black Rose, Regal Behemoth and Throne of the High City. Eleven are `full`; Palace Jailer carries
  one caveat (below).
- A test for CR 725.4's second clause (the active player is the monarch and leaves: the NEXT
  player takes it) now exists beside the first.
- The effect-key ledger gains `condition monarch/an-opponent-became`.

### Open

- **Palace Jailer when its controller leaves.** `eliminatePlayerLocked` drops the departed
  player's delayed triggers (CR 800.4a) *before* the CR 725.4 hand-on crowns someone else, so the
  jailed creature stays exiled for good even though the player who takes the crown is an opponent
  of the one who left. Declared as the card's caveat and pinned by
  `TestPalaceJailerControllerLeavingKeepsTheCreatureExiled`. Closing it means deciding what an
  "until" return owes when the player whose effect created it is gone — a question for every
  event-keyed "until", not for this card.
- **The return uses the stack.** CR 610.3c makes an "until" return an immediate one-shot; here it
  is a delayed trigger, so there is a response window with the creature still in exile. This is the
  posture the Oblivion Ring family (Ossification, Hostage Taker) already takes, uncaveated.
- **Court of Locthwain is not built.** Its first half is `ExileTopWithPermission` (any type,
  while exiled); its monarch half — "until end of turn, you may cast **a** spell from among cards
  exiled with this enchantment without paying its mana cost" — is a *single-use* free cast over a
  set of cards exiled by one source. `CastPermission.Cost: "{0}"` over that set would let every one
  of them be cast free, which is stronger than printed. It waits on a use-count on a granted
  permission.
