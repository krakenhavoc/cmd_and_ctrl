# ADR 0022 — Impulse exile: playing cards from a zone that isn't yours

**Status:** accepted (S21 sub-PR 6)
**Extends:** [ADR 0021](0021-additional-costs.md) (announce-time gates on `CastSpell`).

## Context

"Impulse exile" is the community name for:

> Exile the top card of your library. Until end of turn, you may
> play it.

It is card advantage with a shot clock, and it is why a red deck can
play a long game. It is also everywhere: the Pirates decklist triage
put it top of the blocked list, and the mechanic generalises far
past that deck.

The engine had no way to express it. `castSourceZoneLocked` accepted
`hand` and `command` and nothing else, so a card in exile was inert
by construction. Worse, exile is a **shared, public zone**: unlike
hand and command, "the card is in this zone" grants nobody anything.
Ragavan exiles off the top of the player he *hit* and lets *you*
cast it — so the permission names a player who is usually not the
card's owner, and every other player must be refused.

## Decisions

### 1. The permission lives on the card, as a value struct

```go
type ExilePlayPermission struct {
    Player    uuid.UUID
    UntilTurn int
    CastOnly  bool
    AnyColor  bool
}
```
on `Card.ExilePlay`.

Not a `map[uuid.UUID][]uuid.UUID` on `Game`, for three reasons that
each rule it out on their own:

- It names a player who is not the owner and not the controller, so
  no existing per-player structure is the right home.
- It expires. A card that leaves exile and comes back later must
  not inherit a grant nobody made.
- It has to survive `Clone` / `RestoreFrom`, because undo has to
  roll it back. A value struct rides `cloneCard`'s `out := c` for
  free; a pointer would alias across the snapshot and silently
  break undo.

### 2. Expiry is a turn number, not a sweep

`UntilTurn` is the last turn on which the grant is live, and
`Active(player, turn)` is the only question anyone asks. A cleanup
sweep alone would be wrong under extra turns and would leave a
window where a re-exiled card inherits a stale grant; the number is
self-describing and correct on its own.

The cleanup-step sweep still exists — alongside the damage wipe and
the turn-scoped replacement clear — but it is hygiene, not
correctness: it stops the exile zone accumulating dead grants that
the wire would otherwise ship to every client forever.

The grant is also zeroed as the card leaves exile, so a spent grant
can't come back with the card.

### 3. `CastOnly` is the card, not a shortcut

Ragavan says "you may **cast** that card". Breeches says "you may
**play** those cards". The difference is a land: playing one is a
special action, not a cast (CR 305.1, 115.2a), so a land off the top
of Ragavan's trigger is stranded in exile forever. That is the
printed card, and modelling it costs one bool — so the client
offers no button on that card rather than a button that fails.

### 4. "Spend mana as though it were mana of any color" folds into generic

Breeches' second clause is a rewrite of the cost, and the pool
solver already handles the shape: a requirement that any token can
satisfy **is** a generic requirement. `asAnyColorCost` moves each
colored slot into `Generic` and the existing `CanPay` / `SpendMana`
work unchanged.

Colorless `{C}` requirements are left alone: "mana of any color"
does not include colorless (CR 106.1b).

Only observable under the strict mana gate — in permissive mode
nothing checks the cost at all — but it's three lines, and getting
it wrong later would be a silent rules bug rather than a visible
one.

### 5. The grant is public, so no per-viewer stamping

`CardView.exile_play` carries the holder's ID and rides every
viewer's snapshot. The trigger that created it resolved in the
open; hiding it would be a fiction, and it saves the per-viewer
stamping pass that `legal_targets` needs. The client offers the
button only when `player` is the viewer.

## Consequences

- Ragavan, Nimble Pilferer and Breeches, Brazen Plunderer ship.
- Every "exile the top card, you may play it" card in Magic is now
  a catalog entry rather than an engine change — the mechanic is
  common enough that this is the point.
- The exile zone browser gains its first *play* affordance. Until
  now it was a read-only pile plus owner move buttons.

## Known gaps

- **Batching (CR 603.1).** Breeches reads "whenever one or more
  Pirates you control deal damage to your opponents… exile the top
  card of each of those opponents' libraries" — one trigger for the
  whole combat. The engine emits one damage event per source, so
  three Pirates hitting three players fires three triggers with one
  card each: same cards exiled. Two Pirates hitting the **same**
  player is where it diverges — paper exiles one card, this exiles
  two. The fix is event batching, which is its own piece of work.
- **Breeches, Eager Pillager** is in the decklist and not here. Its
  attack trigger is modal *with a once-per-turn-per-mode ledger*
  ("choose one that hasn't been chosen this turn"), which is new
  machinery on `TriggeredAbility`, not on impulse exile.
- **Timing still applies.** The grant lets you play the card; it
  does not change when. A stolen sorcery needs your main phase and
  an empty stack, and a stolen land uses your land drop — which the
  sandbox doesn't track, so it's unlimited.
- **No "you may play it until the end of your next turn"** duration
  yet. `UntilTurn` can express it; nothing computes it, because no
  card here needs it.
