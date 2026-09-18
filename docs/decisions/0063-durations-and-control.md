# ADR 0063 — Durations of continuous effects, and control from spells and abilities

**Status:** Accepted · 2026-09-18 · Sprint S38 (Layers) · tracked on
[#755](https://github.com/krakenhavoc/cmd_and_ctrl/issues/755) and
[#756](https://github.com/krakenhavoc/cmd_and_ctrl/issues/756), sprint
tracker [#881](https://github.com/krakenhavoc/cmd_and_ctrl/issues/881).

**Numbering:** on 2026-09-18 every remote head was checked with the
AGENTS.md §4 loop (`git ls-remote --heads origin`, then the add-history
of `docs/decisions/*.md` across all refs). The highest number anywhere
is 0060 (`docs/decisions/0060-leaving-the-game.md`). 0061 is being
drafted in parallel for replaceable token creation and discard, and
0062 for hand abilities and cycling, so this ADR takes 0063.

**Amends:**
- [ADR 0035](0035-until-end-of-turn-effects.md) §3 ("Expiry is stamped,
  not implied") — the stamp stays, but it becomes one case of a
  `Duration` value rather than the only duration the registry can hold.
  §3's "known caveat" about `Turn.Number` counting rounds is resolved
  here, not deferred.
- [ADR 0036](0036-attachments.md) decision 17 — layer 2 is no longer
  reachable only by an Aura.

**Builds on:**
- [ADR 0012](0012-layer-system.md) and
  [ADR 0046](0046-layer-6-authoritative.md) — the CR 613 layer engine,
  the per-bucket timestamp sort, and the fixed-point pass.
- [ADR 0059](0059-turn-machinery.md) Decision 1 — turn identity. That
  ADR defines `Player.TurnsBegun` and explicitly hands the "until your
  next turn" sweep to #755. This ADR implements the `TurnsBegun` half
  of Decision 1 and the sweep; `Turn.Seq`, `Round`, extra turns and
  extra phases stay with #753.
- [ADR 0026](0026-delayed-triggers.md) — a duration that ends at a turn
  boundary is watched the same way a delayed trigger is: at the hook
  that runs when a turn begins, not by a per-card timer.

Line references are to `origin/develop` at `2a0a46bd`.

---

## Context

### One duration, hard-coded

`ScopedStatic` (`server/internal/game/turn_scoped_statics.go:63`) is the
S32 home for a continuous effect that outlives its source. It has
exactly one duration field, `ExpiresAfterTurn int` (`:100`), stamped
unconditionally with `g.Turn.Number` by the only registration function
(`RegisterTurnScopedStaticForEffect`, `:132`) and swept at the first
cleanup step that sees it (`ClearExpiredTurnScopedStaticsLocked`,
`:164`, reached from `sweepTurnEndLocked`, `rotation.go:84`).

So the engine can express "until end of turn" and nothing else. The
file's own "What is NOT here" note (`:28-35`) lists the missing
durations; CR 611.2a ("no stated duration — the effect lasts until the
game ends") and CR 611.2b ("for as long as…") are both unreachable, and
so is "until your next turn", because `Turn.Number` counts *rounds*:
all four seats in a Commander game share one number (`turn.go:188`
bumps it only when the cursor wraps past seat 0).

### Layer 2 works, but only an Aura can reach it

`Layer2Control` is a real bucket (`layers.go:55`, `:320`). The
machinery around it is complete:

- `layerPassLocked` captures `Card.BaseController` on the first
  recompute after a permanent enters (`layers.go:592`), `MoveCard`
  clears it on the way out (`zone.go:208`), and the printed
  characteristic reseeds from it every pass (`characteristic.go:178`).
  **Control therefore reverts by itself** when an effect stops applying
  — no "remember who had it" bookkeeping anywhere.
- `materialiseControlLocked` (`layers.go:660`) writes layer 2's answer
  back onto `Card.Controller` and carries the two CR consequences with
  it: summoning sickness under the new controller (CR 302.6) and
  removal from combat (CR 506.4).
- Two control effects on one permanent sort by timestamp and the later
  one wins (CR 613.7), through the same sort every other layer uses.

The only producer is `ControlAttachedBySource()`
(`cards/effects/attachments.go:294`), used by Mind Control and Control
Magic. No spell, no activated ability and no triggered ability can gain
control of anything, and nothing can exchange control (CR 701.12).

### What the rules ask

- **CR 611.2a** — a continuous effect from a resolved spell or ability
  with no stated duration lasts until the game ends (Agent of
  Treachery).
- **CR 611.2b** — an effect can last "for as long as" a condition
  holds; if the condition is false as the effect would start, it never
  starts, and once it stops being true the effect ends (Sower of
  Temptation: "for as long as this creature remains on the
  battlefield").
- **CR 611.2c** — the set of objects a one-shot continuous effect
  affects is locked in when it starts. Already honoured by the S32
  card-side primitives, which snapshot on `(InstanceID,
  EnteredBattlefieldAt)` (`cards/effects/until_end_of_turn.go:40-51`).
- **CR 400.7** — an object that leaves a zone and returns is a new
  object; effects that were affecting the old one stop.
- **CR 514.2** — "until end of turn" ends during the cleanup step.
- **CR 500.1 / 502.1** — a turn begins with its untap step, and the
  untap step's turn-based actions happen after the turn has begun.
- **CR 613.1b** — control-changing effects are layer 2.
- **CR 701.12a-b** — an exchange of control is one effect: if either
  object can't be exchanged (it has left the battlefield), no control
  is exchanged at all.
- **CR 302.6** — a creature has summoning sickness unless it has been
  under its controller's control since their most recent turn began.
- **CR 506.4** — a permanent that changes control is removed from
  combat.
- **CR 800.4m** — an effect that lasts until a departed player's next
  turn lasts until that turn *would have* begun.

---

## Decision 1 — `ScopedStatic` carries a `Duration` value

`ExpiresAfterTurn int` is replaced by `Duration Duration` — a plain
data struct, no closures, no pointers:

```go
type DurationKind int

const (
    UntilEndOfTurn    DurationKind = iota // CR 514.2
    UntilYourNextTurn                     // CR 611.2b, turn boundary
    ForAsLongAs                           // CR 611.2b, condition
    Indefinite                            // CR 611.2a
)

type Duration struct {
    Kind DurationKind

    // UntilEndOfTurn / UntilYourNextTurn
    Player               uuid.UUID // whose next turn ends it
    ExpiresAtTurnsBegun  int       // that player's TurnsBegun value
    ExpiresAfterTurnsBegun int     // the creating player's, for UEOT

    // ForAsLongAs
    Condition       DurationCondition
    Source          uuid.UUID
    SourceEnteredAt int64

    // any kind
    Pinned          uuid.UUID
    PinnedEnteredAt int64
}
```

**Why a struct and not four registries.** The four durations differ
only in *when the entry stops being in the active set*. Everything else
— the ability, the LKI source, the CR 613.7 timestamp, the immutability
contract, the fresh-slice sweep, the `Clone` block — is identical. Four
registries would be four copies of ADR 0035 §4's trap waiting to be got
wrong three more times.

**Why data and not a predicate closure.** `ScopedStatic`'s closure
contract (`turn_scoped_statics.go:56-62`) exists because the ability's
`AppliesTo` / `Apply` must survive a `Clone` and must not capture a
`*Card`. A duration expressed as `func(*Game) bool` would add a third
closure to a struct the snapshot census already has to refuse
(`snapshot.go:747`). As data, the duration is the part of a scoped
static that a future persistence pass (#515, ADR 0044) can write to
disk and read back verbatim; only the ability stays un-persistable.
That is why the field is classified `carried` in the drift plan even
though today's snapshot drops the whole entry and censuses it.

**The registry is renamed.** `Game.TurnScopedStatics` becomes
`Game.ScopedStatics`, `RegisterTurnScopedStaticForEffect` becomes
`RegisterScopedStaticForEffect(ability, sourceID, label, duration)`
(the name #755 proposed), and `turn_scoped_statics.go` becomes
`scoped_statics.go`. A registry that holds an Agent of Treachery for
the rest of the game must not be called "turn-scoped". The persisted
census counter keeps its wire key `turnScopedStatics` so snapshot files
written by older binaries still decode.

---

## Decision 2 — One expiry function, evaluated at three known moments

```go
// durationExpiredLocked reports whether d has run out as of now.
// endOfTurn is true only in the CR 514.2 cleanup sweep.
func (g *Game) durationExpiredLocked(d Duration, endOfTurn bool) bool
```

This is the only code in the engine that decides whether a continuous
effect is over. It is called from exactly one sweep,
`sweepScopedStaticsLocked`, which is called from exactly three places:

| moment | why |
|---|---|
| `sweepTurnEndLocked` (cleanup step, `rotation.go`) | CR 514.2 — `UntilEndOfTurn` ends here, including for an effect created during the end step |
| `onTurnBeganLocked` (a turn begins, `rotation.go`) | `UntilYourNextTurn` ends here |
| the top of `recomputeLayersLocked` (`layers.go`) | `ForAsLongAs` is a condition the board can falsify at any moment, and every board change already bumps the layer version |

No card ever schedules its own expiry, and no duration has a bespoke
sweep site. A new duration kind is a new case in one `switch`.

**Per kind:**

- **`UntilEndOfTurn`** — expires in the cleanup sweep of the turn it
  was stamped in, exactly as ADR 0035 §3 describes, and at the latest
  when its creating player's next turn begins. The second clause is a
  backstop, not the mechanism: `sweepTurnEndLocked` runs on every path
  a turn can end by, early ones included (ADR 0059 Decision 6), so the
  cleanup sweep is what actually ends it.
- **`UntilYourNextTurn`** — expires when `Player.TurnsBegun` reaches
  the value stamped at registration (`TurnsBegun + 1`). See Decision 3.
- **`ForAsLongAs`** — re-evaluated on every layer pass. Two conditions,
  both data:
  - `WhileSourceOnBattlefield` — the source must still be on the
    battlefield *as the same object*: `(InstanceID,
    EnteredBattlefieldAt)` must both match. This is CR 400.7 and it is
    why the stamp is half the key (Decision 5).
  - `WhileYouControlSource` — the above, and `Card.Controller` must
    still be `Duration.Player`.
- **`Indefinite`** — never expires on time. It is dropped only by the
  pin (Decision 6).

**Timing of "until your next turn": as the turn begins, before untap.**
`onTurnBeganLocked` runs from `beginNextTurnLocked` after the cursor is
stamped on the new seat's untap step and before
`runStepEntryHooksLocked` performs the untap turn-based action. CR
500.1 makes the untap step the first step of the turn and CR 502.1-502.3
put untapping *inside* it, so the turn has begun before anything
untaps. Choosing the untap *step's actions* instead would be observable
and wrong: "creatures your opponents control don't untap during their
next untap step" and "until your next turn" would fight over who goes
first. This ordering also matches the delayed-trigger machinery, which
fires "at the beginning of your next upkeep" from the same hook family
(ADR 0026).

---

## Decision 3 — "Your next turn" counts seat-turns, not rounds

ADR 0059 Decision 1 already specified the counter this needs:

```go
type Player struct {
    // ...
    TurnsBegun int // turns this seat has begun, or would have begun
                   // after leaving (CR 800.4m)
}
```

This ADR implements that field and nothing else from Decision 1.
`beginNextTurnLocked` (the single rotation seam, ADR 0059 Decision 6)
increments it for the seat whose turn begins **and for every eliminated
seat the rotation steps over**, which is what makes CR 800.4m fall out
for free: an effect that lasts until a conceded player's next turn ends
at the moment the cursor walks past their seat, which is exactly when
that turn would have begun. `Start` stamps the starting seat's first
turn, the one turn that does not come through the seam.

A `UntilYourNextTurn` duration therefore stamps
`ExpiresAtTurnsBegun = Player.TurnsBegun + 1` at registration and the
expiry is one integer compare. Every case falls out:

| created during | stamped | ends |
|---|---|---|
| your own turn | your current count + 1 | your next turn |
| an opponent's turn | your last turn's count + 1 | your next turn, this round or the next |
| before you have ever had a turn | 1 | your first turn |
| your turn, and you then concede | as above | the turn the rotation skips for you |

`Turn.Number` is not read by any duration. ADR 0035 §3's "cannot be
expressed by bumping this field alone" caveat is retired.

---

## Decision 4 — Gain control is a layer-2 `ScopedStatic` with a `Duration`

```go
func (g *Game) GainControlForEffect(
    sourceID, target, controller uuid.UUID, d Duration, label string) bool
```

It registers **one** `ScopedStatic` in the same bucket an Aura uses:

```go
StaticAbility{
    Layer:     Layer2Control,
    AppliesTo: pinned(target, enteredAt),          // values only
    Apply:     func(c *Characteristic, …) { c.Controller = controller },
}
```

and returns false without registering anything if the target is not on
the battlefield.

**Why this and not a `Card.Controller` write.** A one-shot mutation
would have to remember the previous controller, hand it back on a
duration it also has to track, and lose every CR 613.7 interaction. In
the layer-2 bucket all of that already exists: `BaseController` is the
reversion, the per-bucket timestamp sort is the CR 613.7 ordering
against Mind Control and against a second theft, and
`materialiseControlLocked` is the single place the consequences fire.
An Agent of Treachery resolved after a Mind Control wins because it has
the later timestamp, not because anything special-cased it.

**The target is pinned on `(InstanceID, EnteredBattlefieldAt)`**, the
CR 611.2c / CR 400.7 key the S32 primitives already use. A stolen
creature flickered in response comes back as a new object and is not
stolen.

**The consequences ride with the materialisation, and one was
missing.** `materialiseControlLocked` cleared `AttackingTarget` and
`BlockingTarget` but left the announcement bookkeeping
(`announcedAttacks`, `announcedBlocks`, `announcedBecameBlocked`,
added by #830/#857/#859) in place, so a creature stolen mid-combat and
re-declared later in the same combat would be silently denied its
"whenever ~ attacks" trigger — the announcement said it had already
been declared. Noted by #871's agent, fixed here: one
`removeFromCombatLocked(*Card)` helper clears the declaration and its
announcements together, and both `materialiseControlLocked` and any
future single-permanent CR 506.4 site call it.

**Summoning sickness is per controller, not per permanent.**
`materialiseControlLocked` already sets `SummonedThisTurn` on a control
delta (CR 302.6), which is why Act of Treason has to grant haste. It is
set, never cleared, so a creature that was already hasty stays hasty.

---

## Decision 5 — Exchange of control is two statics with one timestamp

```go
func (g *Game) ExchangeControlForEffect(sourceID, a, b uuid.UUID, label string) bool
```

CR 701.12b makes an exchange all-or-nothing, so the primitive checks
both objects on the battlefield *before* registering either, and
returns false having changed nothing if either is gone. When both are
present it registers two layer-2 `ScopedStatic`s — A to B's controller,
B to A's controller — that **share one timestamp**. One timestamp is
what makes them one effect for CR 613.7: a later control-changer beats
both halves or neither, never one of them.

Both halves are `Indefinite` (CR 611.2a — Switcheroo states no
duration) and each is pinned to its own object. If one exchanged
permanent later leaves the battlefield, that half stops applying and
the other half stands. That is the printed behaviour: the exchange
already happened.

A registration helper that takes an explicit timestamp is unexported
and used only here; every other caller gets the clock.

---

## Decision 6 — Undo, clone, persistence, and the pin

`Duration` is data, so `Clone` keeps working unchanged (the slice is
copied into a fresh backing array; entries are immutable after
registration, ADR 0035 §4) and the drift plan classifies every field
`carried`. `ScopedStatic` is added to the drift test's classified types
— it was not there before, so a field added to it used to vanish
silently.

The snapshot still refuses a game with a live scoped static as a full
restore point, because `StaticAbility` is two closures. **This is now a
bigger deal than it was**: an `Indefinite` entry lasts the whole game,
where the old registry's worst case was one turn. Two mitigations, both
here:

1. **A pin.** `Duration.Pinned` / `PinnedEnteredAt` optionally name the
   one object the effect exists to affect. When that object is no
   longer on the battlefield as the same object, the entry is dropped
   by the same expiry function. Every gain-control and exchange
   registration sets it, so a stolen permanent that dies takes its
   theft with it and the census goes back to zero. An indefinite entry
   can only accumulate while the permanent it moved is still on the
   battlefield, which bounds it by the size of the board.
2. **The census names it.** `ContinuationCensus` already labels each
   entry; the label now carries the duration, so an operator reading a
   refused restore point can tell a Giant Growth from an Agent of
   Treachery.

Making an entry re-derivable from data (catalog key + ability index +
affected set + duration) is the real fix and stays with #515. This ADR
does not pretend to close it; it makes the duration half of that record
exist.

**A flickered `ForAsLongAs` source ends the effect.** CR 400.7: the
permanent that comes back is a new object, so "for as long as *this
creature* remains on the battlefield" is about an object that no longer
exists. `SourceEnteredAt` is what makes the engine able to tell, and it
is checked, not assumed — the ID alone would keep a blinked Sower's
theft alive.

---

## Decision 7 — Card-side vocabulary

One escape hatch, not three wrappers. `cards/effects/control.go` adds:

```go
GainControl{Target, Controller, Duration, Label}   // CR 613.1b
ExchangeControl{A, B, Label}                       // CR 701.12
StaticForDuration{Ability, Duration, Label}        // any layer, any duration
```

plus four duration builders that read the game for their stamps —
`DurationUntilEndOfTurn`, `DurationUntilYourNextTurn`,
`DurationWhileSourceRemains`, `DurationWhileYouControlSource` — and
`game.IndefiniteDuration()`.

#755 proposed `StaticUntilYourNextTurn` / `StaticIndefinite` /
`StaticWhileSourceRemains` as three named wrappers. They would be three
copies of the same four lines differing in one argument, which is the
shape the clone-baseline guard exists to prevent. `StaticUntilEOT`
stays as it is: it is the one duration with enough callers to earn a
name, and it now forwards to `StaticForDuration`.

---

## Decision 8 — Out of scope, stated

- **Control of a spell** (Commandeer) and **control of a player's turn**
  (Mindslaver). Neither is a permanent's controller; neither is layer 2.
- **"An opponent gains control"** (Humble Defector, Wishclaw Talisman)
  needs the choose-a-player prompt, which has no issue yet. The
  primitive already takes an arbitrary `controller`, so those cards
  need a prompt and nothing else from here.
- **"When you lose control of ~"** triggers (Khârn the Betrayer). There
  is no event for a layer-2 delta; `materialiseControlLocked` would be
  the place to emit one, and it is a separate change with its own
  ordering questions.
- **`TurnScopedReplacements`, `TurnScopedGameEndGates` and
  `TurnScopedBlockRules`** are not folded into the `Duration` model
  here, although ADR 0059 expects #755 to do it eventually. They are
  cleared wholesale at cleanup today and no card needs a longer
  duration from them yet; folding them in without a card to prove it
  would be speculative machinery in three registries at once.
- **Extra turns.** `Player.TurnsBegun` is incremented once per turn
  begun, so an extra turn counts as a turn — which is right — but no
  catalog card grants one (#753).

---

## Consequences

- The engine can express every CR 611.2 duration, and the only place
  that knows what a duration means is one function.
- Layer 2 is reachable by spells, activated abilities and triggered
  abilities, with the Aura path unchanged.
- `Turn.Number`'s round-vs-seat-turn ambiguity stops being a hazard for
  durations; `Player.TurnsBegun` is the counter, and ADR 0059's
  Decision 1 has one of its three fields implemented.
- A mid-combat control change no longer eats the stolen creature's next
  attack announcement.
- An indefinite entry keeps a game off the full-restore path for as
  long as the permanent it moved is on the battlefield. Accepted, with
  the pin bounding it, and #515 owns the real fix.

## Alternatives considered

**A second registry for durable effects.** Rejected: the two would
share every field but one, and both would need the same immutability,
clone and sweep discipline. ADR 0035 §1 made the same call about
`TurnScopedReplacements` for the opposite reason — those genuinely
share nothing.

**A `func(*Game) bool` expiry predicate on each entry.** Rejected: a
third closure on a struct the snapshot already refuses, and it would
make the expiry rule unreadable — every card would carry its own copy
of "for as long as", and they would drift.

**Writing `Card.Controller` directly with a scheduled revert.**
Rejected: it throws away `BaseController`, CR 613.7 ordering and the
existing materialisation, and it needs bookkeeping that is wrong the
first time two effects overlap.

**Expiring "until your next turn" from the untap step's turn-based
actions.** Rejected: the effect ends because the turn began, not
because something untapped, and putting it in the untap action would
order it against "doesn't untap" effects for no reason (ADR 0058).

**Computing the seat-turn index from `(Turn.Number, ActiveSeat)`
arithmetic** instead of adding `Player.TurnsBegun`. Rejected: it works
today and breaks the moment extra turns land, and ADR 0059 already
decided the field.

## Test plan

Engine (`server/internal/game`):

- each `Duration` expires at exactly its boundary: end of turn; the
  player's next turn *beginning* and not an intervening opponent's
  turn; `ForAsLongAs` when the source leaves and when it is flickered
  back (CR 400.7); `Indefinite` across many turns.
- `UntilYourNextTurn` created on the holder's own turn survives the
  whole round.
- gain control removes the object from combat *and* its announcements,
  stamps summoning sickness per controller, and reverts to
  `BaseController` on expiry; two control effects sort by timestamp.
- exchange both ways; exchange with one object gone registers nothing.
- a stolen commander that returns to its owner keeps commander state.
- undo across a control change and across an expiry; snapshot
  round-trip; the census counts and labels.

Catalog (`server/internal/cards/effects`): Act of Treason, Agent of
Treachery, Sower of Temptation, Mass Diminish, Switcheroo.
