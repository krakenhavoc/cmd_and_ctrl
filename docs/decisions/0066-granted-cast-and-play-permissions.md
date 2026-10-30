# ADR 0066 — Granted cast and play permissions

**Status:** accepted
**Date:** 2026-09-18
**Sprint:** S42 — Casting from non-hand zones: granted permissions and alternative costs
**Issues:** [#652](https://github.com/krakenhavoc/cmd_and_ctrl/issues/652) (granted cast permissions), [#765](https://github.com/krakenhavoc/cmd_and_ctrl/issues/765) (play and cast from the top of your library). Tracker [#885](https://github.com/krakenhavoc/cmd_and_ctrl/issues/885).
**Accepted amendment:** 2026-10-08, a permission that lasts "until you exile another card with this" ([#2539](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2539)), with owner answers of 2026-10-08; see Amendments at the end.

**Numbering:** every remote branch was swept with the AGENTS.md §4 loop on
2026-09-18 (`git ls-remote --heads origin`, 266 heads; `git log --all
--diff-filter=A --name-only -- docs/decisions/` over the fetched refs).
`0060` was on `develop`; `0061` (token creation and discard),
`0062` (abilities and special actions from the hand), `0063` (durations and
control) and `0064` (emblems) were held by in-flight branches, and `0065`
(modal and multi-target clauses) was claimed by a fifth parallel agent that
had not pushed. All five merged to `develop` while this branch was in flight,
and so have `0067`-`0071`; the sweep was repeated immediately before each
push. `0066` is still the first free number, and is now the one gap in the
sequence on `develop` — this ADR fills it. `0005`, `0024`, `0029` and `0030` stay permanently unused.

## Context

S29 shipped casting from a non-hand zone for cards that **print** the
permission: flashback (#411), escape (#510), warp (#413) and per-card
`CastableZones` (#409). Every one of those answers is keyed by oracle ID, so
the permission belongs to the *card* and to every copy of it:

- `game/cast_zones.go` — `CastableZonesFor` / `CardCastableFromZone` read
  `CatalogCastableZones(oracleID)` and nothing else.
- `game/alternative_cost.go` — `AlternativeCostByKey` reads the catalog, so
  `validateAlternativeCost` rejects a `"flashback"` claim on a card that does
  not print flashback, and `altCostExilesFromStack` is catalog-only too.
- `protocol/view.go` — `castable_here` on a graveyard card is catalog-only.
- `legal/cast.go` — the bot enumerator's sources are hand and command only
  (#673).

The one per-instance grant is `ExilePlayPermission` (`game/exile_play.go`), a
value on `Card`. It is exile-only, carries no alternative-cost key, no timing
bypass and no way to name a set of cards, and there is no library zone in the
cast path at all: `castZoneFromWire` accepts hand, command, exile and
graveyard, and `castSourceZoneLocked` resolves those four.

Eight cards want the same thing from four different directions:

| card | shape |
|---|---|
| Snapcaster Mage | one target card in your graveyard gains flashback until EOT, cost = its mana cost |
| Past in Flames | every instant/sorcery card in your graveyard gains flashback until EOT — the set is **locked on resolution** (CR 611.2c) |
| The Grim Captain's Locker | `{T}`: every creature card in your graveyard gains "Escape—{3}{B}, Exile four other cards" until EOT — same shape as Past in Flames |
| Underworld Breach | **static**: every nonland card in your graveyard has escape for as long as Breach is on the battlefield (CR 702.138) |
| Bolas's Citadel | you may play the top card of your library, paying life equal to its mana value rather than its mana cost |
| Oracle of Mul Daya | play with the top card revealed; you may play **lands** from the top of your library |
| Courser of Kruphix | play with the top card revealed; you may play lands from the top of your library |
| Realmwalker | you may look at the top card; you may cast **creature spells of the chosen type** from the top |

They are one seam. The purpose of this ADR is to make sure the engine grows
**one** answer for all eight and for the consumers queued behind them —
foretell (#658), madness (#657), suspend (#659), Muldrotha, Lier.

## Decisions

### 1. One permission type, one query, three homes

`game.CastPermission` (`server/internal/game/cast_permission.go`) is the only
model of "an effect lets you cast or play a card from somewhere its own text
does not allow". `ExilePlayPermission` is **deleted**, not kept alongside it:
impulse exile, airbend, warp, cascade and the Siege face grant all become
`CastPermission` values. There is no second per-instance grant.

```go
type CastPermission struct {
    Player uuid.UUID     // who may cast or play — often not the owner
    Zone   ZoneKind      // ZoneExile | ZoneGraveyard | ZoneLibrary
    Scope  PermissionScope
    Cards  []PermissionCardRef // Scope == ScopeCards: the named objects
    Filter PermissionFilter    // Scope == ScopeStanding: which cards qualify
    ...
}
```

`Scope` is the whole of #652's "instance grants versus static grants"
question, reduced to two values because the third shape collapses into the
first:

- `ScopeCards` — a list of named card objects. **One** card (Snapcaster, a
  Siege, an impulse-exiled card) and **a set locked at resolution** (Past in
  Flames, the Locker) are the same thing: CR 611.2c says a one-shot
  continuous effect locks the set of objects it affects when it resolves, so
  the resolution walks the graveyard once and writes down the instances. A
  card that reaches the graveyard afterwards is not in the list and has no
  flashback. Nothing has to re-evaluate.
- `ScopeStanding` — a rule over a zone, with a pure-data `PermissionFilter`
  saying which cards in it qualify. Underworld Breach and the whole library-top
  family are this.

The three homes are chosen by who owns the lifetime, and each one makes a
duration free:

| home | scope | lifetime |
|---|---|---|
| `Player.CastPermissions []CastPermission` | `ScopeCards` | a `game.Duration` (ADR 0063): `UntilEndOfTurn` for Snapcaster, Past in Flames and impulse exile, `WhileInZone` for airbend and warp |
| derived from the battlefield through `CardDef.CastPermissions` | `ScopeStanding` | for as long as the source permanent remains — nothing to expire |
| — | — | — |

Standing permissions are **derived, never stored**, for the reason
`land_drops.go` already gives about `CatalogAdditionalLandPlays`: two
Underworld Breaches have to compose, and one of them leaving must not revoke a
permission the other is still granting. Deriving them per query is also the
whole of "for as long as the source remains" — there is no expiry code to get
wrong, and a Breach that is exiled in response to the cast stops granting
before the cast is validated, which is what CR 702.138 says.

**The duration model.** #755/#756's `game.Duration` (ADR 0063) is the
project's one duration vocabulary, and since #945 a `CastPermission` carries
it: `CastPermission.Duration`, swept through the same `durationExpiredLocked`
every `ScopedStatic` is swept through. `NotBeforeTurn` sits beside it and is
not a duration — see the amendment at the foot of this ADR.

### 2. CR 400.7 is an object-identity check, not a sweep

A grant ends when its card leaves the zone **by any route**, not only when it
is cast. Before this ADR that was enforced by zeroing `Card.ExilePlay` at
every exit — in `MoveCard`, in both cast branches, in the entry paths. That
discipline works only as long as every author remembers it, and moving the
store off the card would have multiplied the sites rather than removed them.

So identity is stamped instead: `Card.ObjectEpoch` is an integer `MoveCard`
increments on every zone change, and `PermissionCardRef` is `{ID, Epoch}`.
This branch and #973 (the CR 732 loop breaker's per-object tally key) reached
for the same field independently and named it the same thing; #973 merged
first, so the field and its bump are theirs and this ADR is its second
reader. That two seams converged on it is the argument for it. A
named permission applies to the object it was granted to and to no other. A
Snapcaster target that is exiled and returned to the graveyard is a new object
with a new epoch and has no flashback (CR 400.7); a cascade hit that is cast
has already had its epoch bumped by the move to the stack, so the grant is
spent without anybody writing `= CastPermission{}`. One integer replaces six
clear sites, and it is the rule rather than a proxy for it.

The stored slice is still swept at cleanup — `clearExpiredCastPermissionsLocked`
drops permissions whose turn has passed and permissions every one of whose
cards has moved on — but only to stop the slice growing, never for
correctness.

### 3. Both halves of a cast consult the grants, through one function

`CastableZonesFor(oracleID)` stays pure and catalog-only: it is what the card
says about itself and several call sites want exactly that. The cast path,
the view and the bot enumerator all go through

```go
func (g *Game) CastPermissionForLocked(playerID uuid.UUID, card Card, zone ZoneKind) *CastPermission
```

which is `nil` for hand and the command zone and otherwise the permission
that opens the cast. `validateCastPathLocked` takes it in place of the old
`hasExileGrant bool`, so exile stops being a special case in the switch: a
granted graveyard cast, a granted exile cast and a granted library-top cast
reach the same rules.

It does **not** answer `nil` for a card whose own text already opens the
zone. Gravecrawler under an Underworld Breach is castable both ways and the
caster picks; which *price* wins is a separate question, settled in one other
place (below).

The **price** rides the same object. A grant that names an alternative-cost
key synthesises the `AlternativeCost` the cast is judged under:

```go
func (p *CastPermission) AlternativeCostFor(card Card) *AlternativeCost
```

Snapcaster's grant becomes `{Key: "flashback", ManaCost: card.ManaCost,
FromZone: ZoneGraveyard, ExileOnLeavingStack: true}`; Underworld Breach's
becomes `{Key: "escape", ManaCost: card.ManaCost, FromZone: ZoneGraveyard,
ExileFromGraveyard: <exactly three other cards>}`; Bolas's Citadel's becomes
`{Key: "bolas_citadel", ManaCost: "", LifeEqualToManaValue: true}`. The
synthesised offer is built on demand and never stored, so the stored type
stays free of closures and the snapshot stays a plain mirror.

Sharing the keys with the printed keywords is deliberate and is what #652
asked for: CR 702.34a's "if the flashback cost was paid, exile it" and CR
702.138b's "escaped" read `StackItem.AltCost`, which is set from the claim
however the permission arrived. `altCostExilesFromStack` therefore takes the
resolved offer rather than re-reading the catalog, because by then the catalog
has nothing to say about a Snapcaster'd Brainstorm.

**A card that both prints and is granted the same key keeps its printed
cost.** The catalog is consulted first and a grant never overrides an offer
the card already makes: Deep Analysis flashed back under Past in Flames costs
`{1}{U}` + 3 life, not `{3}{U}`. This is the strictly-weaker direction, and it
is the direction a sandbox must err in (#259).

*Amended 2026-10-16 (#3938, [below](#amendment--2026-10-16-3938-a-printed-and-a-granted-instance-of-one-keyword-are-two-offers)):*
the printed offer still keeps the bare key, and the grant's offer is a second
instance of the keyword (CR 113.2c), offered beside it under
`GrantedInstanceKey`. Deep Analysis under Past in Flames is offered both
`{1}{U}` + 3 life and `{3}{U}`, and its owner picks.

### 4. Playing and casting from the top of your library (CR 401.5)

`castZoneFromWire` gains `library`, and `castSourceZoneLocked` resolves it to
the caller's own library — a per-player zone, so the lookup already scopes the
permission to "your library". The land-play path needs no change beyond that:
lands are played through `cast_spell`, so a land on top of the library reaches
`CastSpell`'s land branch through the same door as a land in hand, spends the
same land drop (CR 305.2, `LandDropsRemainingLocked`) and runs the same CR 614
entry pipeline.

Two restrictions ride the permission rather than the zone:

- **`TopOfLibraryOnly`.** A library grant opens exactly one card: the one
  currently on top. `CastPermissionForLocked` compares the card's instance to
  `Library.Cards[len-1]` at announce. CR 601.2a then moves it to the stack and
  exposes the next card, which per CR 401.5 is a different object the cast in
  progress may not see — nothing has to enforce that separately, because the
  announce check has already happened and the permission is re-derived from
  scratch next time.
- **`Filter`.** `PermissionFilter{LandsOnly}` is Oracle of Mul Daya and
  Courser of Kruphix; `{CreatureOnly, ChosenCreatureType}` is Realmwalker,
  reading the type named as it entered (`Card.NamedTribe`, S26); the zero
  value is Bolas's Citadel, which opens everything. It is a small pure-data
  struct rather than a predicate closure so a permission can be snapshotted;
  it grows a field when a card needs one, and a card that needs a real
  predicate gets a `ScopeStanding` permission derived from a `CardDef` that
  can compute it.

### 5. One visibility rule for the library top, kept on the position

A player may only play a card from the top of their library if they can *see*
it (CR 401.5). The two printed spellings are "you may look at the top card of
your library any time" (private to its owner — Realmwalker, Bolas's Citadel)
and "play with the top card of your library revealed" (public — Oracle of Mul
Daya, Courser of Kruphix).

`Card.KnownBy` is per card and cannot express this: the top card changes
identity on every draw, mill, shuffle, scry and cast, and stamping the new one
means finding every library mutation. So the rule is **per player and kept on
the position**, as `CardDef.LibraryTopVisible` derived from the battlefield
exactly like a standing permission:

```go
func (g *Game) LibraryTopVisibilityLocked(playerID uuid.UUID) LibraryTopVisibility
// LibraryTopHidden | LibraryTopOwner | LibraryTopRevealed
```

It is resolved **lazily, at the two places that need an answer** — the view
projection and `CastPermissionForLocked` — and not eagerly after every
library mutation. Eager stamping would have to hook `PopTop`, `PushTop`,
every shuffle and every search; lazy resolution is correct by construction
because the question is always asked about "whatever is on top *now*".

The view marks the current top card as visible to the players the rule names
**in the projection**, not on the card. It cannot write `Card.KnownBy` — the
view runs under a read lock — and it should not want to: being KNOWN and being
VISIBLE WHERE IT SITS are different facts, and conflating them is how a
one-shot "reveal the top two cards of your library" would turn into a
permanent window into the library. So each frame stamps the projected card's
`knowers` set (which the S13.5 redactor already consults) plus an unexported
`libraryTop` marker, and both are re-derived next frame. CR 401.6 — a top card
that stops being revealed and is revealed again is a new object — falls out,
because nothing remembers the old answer.

The per-viewer filter is then widened by exactly one case: an opponent's
library, which was wholesale-hidden, keeps its **top card only, and only when
that viewer knows it AND the standing rule put it there**. A tucked card in
the middle of a library that still carries a stale knower is not exposed,
because only the last element is considered; and a card merely revealed once
is not exposed either, because it carries no `libraryTop` marker. A spectator
still gets the library wholesale-hidden, since their knower check answers true
for everything.

### 6. Timing and the gates a grant does not open

**A grant never bypasses the sorcery-speed gate unless it says so.**
`CastPermission.Timing` is `TimingNormal` (the card's own timing rules apply —
every card in this ADR), `TimingFlash` ("as though it had flash", CR 702.8 —
the shape madness and the Descendants' Path family will want) or
`TimingSorcery`. `CastSpell` reads it in one place, next to the existing
`HasKeyword(&card, "flash")` check. Bolas's Citadel does not make a sorcery
castable on an opponent's turn, and Underworld Breach does not make a sorcery
in the graveyard an instant.

**The land drop still counts.** CR 305.2 is about playing a land, not about
where the land was, so Oracle of Mul Daya's land from the top consumes the
turn's drop like any other. Raising the allowance is a separate effect
(`GrantAdditionalLandPlayForEffect`, `CatalogAdditionalLandPlays`) that Oracle
does not have.

**"Cast only if" conditions are not built here.** #760 designs the one
announce-time cast gate (CR 101.2 "can't cast" restrictions and "cast only
if" conditions), and a granted permission that carries a condition belongs in
that gate rather than in a second one. `CastPermissionForLocked` carries a
hook comment naming #760 and nothing else; the four graveyard cards and the
four library cards need no condition.

### 7. Wire and client

`castable_here` on a graveyard card, and the alternative-cost offers stamped
beside it, come from the merged view — the catalog's answer for a card that
prints its own permission, plus `CastPermissionForLocked` for a granted one —
so a Snapcaster'd card in the graveyard renders the same cast affordance a
Faithless Looting does, with the synthesised flashback offer in its picker.
The library gains the same two fields on its top card, and the client shows a
"top of library" affordance only when the card is visible to that viewer; a
hidden top card ships as a back, as it always has.

`exile_play` stays on the wire under its own name and shape. It is the
impulse-exile grant the client already renders, and renaming a wire field that
four client modules and a test suite read would be a breaking protocol change
(`docs/protocol.md`, schema evolution rules) bought for nothing. It is now
projected from the permission store rather than from `Card.ExilePlay`.

### 8. Bots

`legal/cast.go` enumerates granted sources through the same function the
engine uses, so the bot can never be offered a cast `CastSpell` will refuse.
That closes the **enumerator half** of #673 for granted permissions and for
the library top; the other half of #673 — enumerating the *printed*
graveyard permissions (flashback, escape, warp) and their alternative costs —
is a separate expansion of the same loop and is not claimed here.

### 9. Undo, clone and persistence

Permissions are **data**, classified `carried`:

- `Player.CastPermissions` — carried. Who may cast what is not derivable from
  the board, and a restore that dropped it would silently revoke a Snapcaster
  or a cascade hit that has not been cast yet. `game.Duration` is pure data
  too, so the follow-up above does not change this classification.
- `Card.ObjectEpoch` — carried. A restore that dropped it would reset every
  card to epoch 0 and revive every permission ever granted against it.
- Standing permissions and library-top visibility are **not stored at all**,
  so they are not classified: they are re-derived from the battlefield on the
  next query, which is the definition of `rebuilt` applied to something that
  was never a field.

The type holds no closures, which is what keeps `CardSnapshot` able to mirror
it and `TestEmbeddedDomainTypesStayPureData` able to prove it.

### 10. Interactions that need nothing new

A card cast from the top of the library whose entry has a replacement pauses
through #919's entry frame like any other: the land branch already sets
`entryResumable` and the resume finishes the push from whatever zone the
window opened over, and the source zone is carried on the replacement event
(`OldZone: ZoneLibrary`). No new pause, no new resume frame.

Cost modifiers already see the real source zone (`CostQuery.FromZone`, #638),
and a granted cast passes `src.Kind` the same way a printed one does, so
Mm'menon-style "not cast from hand" restrictions will read `library` and
`graveyard` correctly the day they are written.

## Out of scope, stated

- **Muldrotha, the Gravetide** — "you may play a land and one permanent card
  of each type from your graveyard *during each of your turns*" is a
  per-type, per-turn allowance. `ScopeStanding` + `PermissionFilter` is the
  right home for the permission half; the per-type-per-turn tally is a second
  mechanism and is not built. (Since built, with Aminatou's Augury beside it —
  see the amendment of 2026-10-09, #2167.)
- **Resolution-time "reveal it and cast it free"** — Descendants' Path,
  Rashmi, Planetarium of Wan Shi Tong. Those cast during a resolution and
  ignore timing; they belong to the `PendingChoiceMayCast` cascade family and
  to `TimingFlash`, not to a standing permission.
- **Casting from another player's library or graveyard** (Bribery-style).
  `castSourceZoneLocked` resolves per-player zones to the caller's own, which
  is what every card in this ADR wants.
- **#760's announce-time cast gate** — hooked, not built.
- **Foretell (#658), madness (#657), suspend (#659)** — each is a
  `CastPermission` with a key and, for madness, `TimingFlash`; none is built
  here. The point of the model is that they are card work plus a key.
  (All three have since shipped on it — see the amendments below.)

## Consequences

### Good

- One type, one query and one price function cover the graveyard, exile and
  the library, and the three consumers queued behind #652 become card work.
- CR 400.7 stops being a discipline and becomes a check. Six "clear the
  grant" sites collapse into one integer.
- "For as long as the source remains" costs nothing, because standing
  permissions are derived rather than stored — the same trick that already
  makes two Explorations compose.
- The bot enumerator and the client read the same function the engine
  validates with, so neither can advertise a cast the engine will refuse.

### Tradeoffs

- `CastPermission` is a wide struct — wider than `ExilePlayPermission` was —
  and most of it is zero on most permissions. That is the price of one model;
  the alternative was the two models #652 forbids.
- `PermissionFilter` is a closed set of pure-data flags, so a standing
  permission whose filter needs a real predicate has to declare it on a
  `CardDef` instead. No card in this ADR does.
- Library-top visibility is resolved lazily, so a *test* that wants to assert
  "who knows the top card" has to ask the view or the permission lookup
  rather than read `KnownBy` after a mutation. That is the cost of not
  hooking every library mutation, and it is the cheaper side.
- The opponent-library projection is no longer "always wholesale-hidden". It
  keeps exactly one card and only when the viewer is a knower of it, which is
  a narrower widening than the hand's, but it is a widening.


## Amendment (2026-09-18, #945): the window is `game.Duration`, and the debt is paid

The follow-up Decision 1 recorded has landed. `CastPermission.UntilTurn` and
`CastPermission.WhileInZone` are **gone**; the permission carries one
`Duration` (ADR 0063), and `NotBeforeTurn` stays.

**What moved.**

| before | after |
|---|---|
| `UntilTurn: g.Turn.Number` — Snapcaster, Past in Flames, the Locker, impulse exile, cascade, a Siege's face grant, Containment Construct, Neyali | `Duration: g.UntilEndOfTurnDuration()`, and the zero `Duration` is stamped to exactly that by `GrantCastPermissionForEffect` |
| `UntilTurn: g.Turn.Number + 2` plus an upkeep delayed trigger — Reckless Impulse, Wrenn's Resolve, Prosper, Cori Mountain Monastery | `Duration: g.UntilEndOfYourNextTurnDuration(controller)` |
| `WhileInZone: true` — airbend, warp, foretell (#987), Aerial Extortionist, every derived standing permission | `Duration: game.WhileInZoneDuration()` |
| `UntilTurn: g.Turn.Number` on suspend's free cast (#987) | `Duration: g.UntilEndOfTurnDuration()`, stamped explicitly because the trigger fires in an UPKEEP and a reader should see which turn it names |
| `NotBeforeTurn` — warp (CR 702.185a), foretell (CR 702.143a) | unchanged |

**One liveness function**, `(*Game).CastPermissionActiveForEffect`, replaces
the value method `CastPermission.Active(player, turn)`. It reads
`durationExpiredLocked`, so a permission and a continuous effect with the same
clause end at the same moment by construction. `GrantsFace` lost its window
check with it: every caller reaches a permission through
`CastPermissionForLocked`, which has already asked, so a second copy of the
rule would only be a copy to drift. `faceForCastLocked` lost its `turn`
argument for the same reason.

**The sweep runs twice**, not once. `sweepCastPermissionsLocked(true)` at the
cleanup step (CR 514.2) and `sweepCastPermissionsLocked(false)` as a turn
begins (CR 500.1) — the two moments `sweepScopedStaticsLocked` runs at for the
same reason. The statics' third moment, the top of every layer recompute, buys
a permission nothing: only `ForAsLongAs` can go false between turns and a
permission never carries one, because "for as long as the source remains" is
already free for a standing permission that is derived from the battlefield on
every query.

**`NotBeforeTurn` is not a duration, and that is why it survived.** CR 611.2
says when a continuous effect ENDS; it has no vocabulary for when one starts,
because a continuous effect starts when it is created. A cast permission is
the one thing in the engine that can be granted now and open later — warp's
"you may cast it from exile on a later turn" (CR 702.185a) and foretell's
identical clause (CR 702.143a) — so the floor is a field of the permission and
applies whatever the `Duration` says. Foretell (#658, shipped in #987) uses exactly this
pair — `Duration: WhileInZoneDuration()` plus `NotBeforeTurn` — and so does warp; neither
needed a second field.

**The card-level prize.** `EndCastPermissionAtTurnForEffect` and
`b19EndImpulseGrantWithThisTurn` are deleted. "Until the end of your next
turn" used to be stamped two rounds out as a backstop with a CR 603.7 delayed
trigger in the holder's next upkeep pulling it back, because no single ROUND
number means that clause for every seat. ADR 0063's seat-turn counter says it
directly, so Reckless Impulse, Wrenn's Resolve, Prosper, Tome-Bound and Cori
Mountain Monastery schedule nothing at all.

**Snapshot schema 2 → 3.** A stored permission's JSON changed shape
(`untilTurn` / `whileInZone` → `duration`). A v2 file would decode into the
zero `Duration`, which reads as "until end of turn, unstamped" — so every
airbend and warp grant in a restored pre-v3 game would lapse at the next
cleanup. That is a semantic shift in an existing field, which is what
`SnapshotSchemaVersion` is for. `Player.CastPermissions` stays classified
`carried` in `snapshot_drift_test.go`: `Duration` is pure data, so the type is
still embedded by value and still marshals.


## Amendment (2026-09-18, #978): the view stamps exile through the engine's castability predicate

Decision 7 gave exile `exile_play` and left it there. Everything else the
cast dialog needs — `legal_targets`, `clauses`, `modes`, `additional_cost`,
`tap_cost`, `target_cost_notes`, `phyrexian_symbols`, `alternative_costs` —
was stamped by `stampLegalTargets`, which walks hand, the command zone, the
graveyard and the library top. All four are per-seat zones. Exile is a shared
top-level one, so an exiled card a permission opened arrived carrying
`target_mode` and `mana_cost` and nothing else, and once #977 routed the
impulse button through the one cast chain that chain had nothing to read: an
exiled modal spell got no mode picker and a targeted one relied on client
heuristics.

**One stamping function, called once per zone.** The per-card body of
`stampLegalTargets` is now `stampCastOffers(g, caster, card, key, zone,
granted)`. `stampLegalTargets` calls it for the four per-seat zones;
`stampGrantedPermissions` calls it for exile, where it already holds the
permission the engine answered with. Not a copy — the graveyard's answer and
exile's come out of the same lines.

**The gate is `CastPermissionForLocked`**, re-asked for the holder
`CastPermissionOnCardForEffect` named. That matters twice: it is the function
`CastSpell` validates with and `legal/cast.go` enumerates with, so the three
cannot disagree; and it refuses a window that has not OPENED yet (warp's
CR 702.185a floor), which is exactly right — the grant is still shown, so the
client can grey the button, but there is no cast to compute a target set for.

**The key is the granted FACE's** (ADR 0034). `game.CatalogKey` of the card
with the permission's face applied, so a defeated Siege's back-face cast
ships the back face's modes and target clause rather than the battle's. The
per-seat zones keep reading the bare oracle ID, because no permission there
names a face.

**The stamps are per viewer, and that is new for a shared zone.**
`CardView.castOffersFor` (unexported, never on the wire) carries the seat the
stamps were computed for, and `FilterViewFor` drops them for everybody else —
spectators and admins included, for the reason `legalMovesFor` gives. A legal
target set is narrowed by hexproof, shroud and "target opponent", so one
seat's answer is not another's to read. Everywhere else this was implied by
the zone belonging to the seat.

`redactCardForViewer` also learned to clear `modes`, `additional_cost`,
`legal_targets`, `clauses` and `cant_cast` for a non-knower. Before this pass
those only ever landed on a card in a zone the filter drops wholesale, so
nothing cleared them; a face-down foretold card in exile is a card in a SHARED
zone that carries them, and "this spell chooses two of three modes" names a
card.

**Two things ADR 0073 gets for free, because it stamps through the same
function.** `optional_costs` and `cant_cast` now reach an exiled or
library-top card as well: a permission opens a ZONE, and a cast restriction
shuts the cast anyway (CR 101.2). And the precedence between them is fixed
here — `stampGrantedPermissions` runs after `stampLegalTargets`, and it used
to set `castable_here` back to true over the gate's refusal, so a granted
flashback under Grafdigger's Cage rendered a button the engine would reject.
It now leaves `castable_here` alone when `cant_cast` is set.

Additive on the wire (`v` unchanged): a client that ignores the new fields on
an exiled card behaves exactly as it did.

---

## Note (2026-09-18, #719): the face restriction is a list

`CastPermission.Face int` is now `CastPermission.Faces []int`, and the
reason is the one the amendment above gives for the window: a sentinel
that has to carry two meanings carries neither.

Face was documented as *"ZERO MEANS THIS PERMISSION DOES NOT SPEAK
ABOUT FACES, not face 0"*, which was true and sufficient while the only
face-naming grant was a defeated Siege's back face (S32). CR 715.4 is
the second one, and it names the **creature** half of an adventure card
— face 0. With one integer, "the grant opens face 0 and nothing else"
and "the grant has no opinion" are the same value, and the second
reading is the one that would let an Adventure half be cast a second
time out of exile.

It is the same defect #945 took out of `UntilTurn` one field up — 0 was
both the zero value and a turn number — and it is fixed the same way,
by giving the field a value that means "nothing to say" rather than by
adding a companion boolean. Empty is "no opinion"; a non-empty list
NARROWS to exactly those faces. It is also the shape
`Card.CastableFaces` already returns, so `faceForCastLocked` composes
the two rather than arbitrating between them, and it stays pure data
like `Cards` — `cloneCastPermissions` reallocates it beside the card
refs.

Three readers, all thin, and none of them checks the window (#945's
rule: the one liveness test is `CastPermissionActiveForEffect`, and
every caller has already been through `CastPermissionForLocked`):

- `GrantsFaces(playerID)` — the full set, read only by
  `faceForCastLocked`.
- `GrantsFace(playerID)` — the single face, for the enumerator and the
  pricer.
- `NamedFace()` — the single face with no PLAYER either, for the view,
  which asks what a grant opens before it knows whose it is (the grant
  is public information) and has to label a warp grant's greyed-out
  button with the half it will open.

A permission naming SEVERAL faces answers false to the last two and
narrows the caller's request in the first — nothing declares one today,
and a choice is the honest thing to do with one if anything ever does.

The behaviour of a single-face grant is unchanged in every respect,
including the deliberate departure ADR 0034's S32 addendum records: the
named face is the ANSWER and the caller's requested face is ignored.

## Amendment (2026-09-19, #696 / #695): one pricer, one payability predicate

This ADR's "one type, one query and one price function" claim was true of
the *engine*. It was not true of everything that has to agree with the
engine, and two consumers had quietly kept partial copies.

### One pricer (#696)

`GET /games/{id}/auto-tap-preview` parsed `card.ManaCost`, found the
commander tax by scanning the caller's command zone, and applied the cost
modifiers itself. That copy knew nothing about the four things this ADR and
ADR 0073 added to the price:

- the alternative cost claimed at announce (`resolveAlternativeCostLocked`),
- a granted permission's flat override (`CastPermission.Cost` — airbend's
  `{2}`, cascade's `{0}`),
- the "spend mana as though any colour" fold (`CastPermission.AnyColor`),
- the mana half of the announced optional additional costs (ADR 0073 §3),

nor about the face being cast (ADR 0034) or the source zone reaching the
cost modifiers. The reported failure is the clean one: an exiled
`{5}{R}{R}` under a `{0}` grant previewed as
`{"ok":false,"missing":["{R}","{R}","{1}","{1}","{1}","{1}","{1}"]}` while
`CastSpell` charged nothing. The reverse — a flashback cost higher than the
printed one previewing as affordable — handed the player a cast that failed.

The rule now: **nothing outside `internal/game` re-derives a cast's price.**

```go
func (g *Game) PriceCast(playerID uuid.UUID, card Card, params CastSpellParams) (CastPrice, error)
func (g *Game) PriceCastForEffect(playerID uuid.UUID, card Card, params CastSpellParams) (CastPrice, error)
```

`params` is the **announcement** — the same `CastSpellParams` the caller
would send to `CastSpell` — and `CastPrice` carries what every consumer
needs out of one walk: `Printed` / `Paid` (the CR 118.9 pair of strings),
`Card` (the face-materialised copy every other announce gate reads), `Base`
and `Total`.

The face is settled by `faceForCastLocked` in the shape the amendment above
leaves it — the permission is read off the UNFACED card, because a grant
belongs to the instance and not to a face, and its `Faces` then narrow the
caller's request. So the pricer and `CastSpell` agree about which half of an
adventure or a modal DFC is being priced without either of them re-deriving
it.

`Base` and `Total` are the two levels this file already had and did not
name. `printedCostLocked` is the chosen cost string plus the commander tax
plus the optional-cost mana plus the any-colour fold; `effectiveCostLocked`
is that with the cost modifiers applied and the convoke / waterbend
subtraction taken off. The split is not new and is not cosmetic: the
announce-time tap budget is measured against the first, and so is the bot
enumerator's {X} search, because convoke settles the `{X}` slot into generic
as part of *paying* (CR 601.2h) and an enumerator reading `Total` would
never offer Chord of Calling at an X above zero.

Four readers, one walk: `applyCastCostLocked` (the payment),
`applyAutoTapLocked` (the tapper), `lobby.autoTapPreview` (the endpoint) and
`legal.castMovesForCard` (the bot). Add a component in `printedCostLocked`
or `costAfterModifiersLocked` and all four get it.

The endpoint takes the announcement off its query string — `from_zone`,
`alternative_cost`, `optional_costs`, `tap_ids`, `face` — and the client
builds those from the cast payload it is about to send. A malformed one is
a 400 rather than a silent default: a preview that quietly priced a
different cast from the one the button will send is worse than no preview.

### One payability predicate (#695)

The S28 comment in `viewOfAlternativeCosts` already stated the rule — "a
greyed-out button the server would reject is worse than no button" — and
implemented one third of it. `AlternativeCost.Available` asks the offer's
`Condition` and nothing else, so Force of Will at 0 life and Snuff Out at 3
were listed, selectable, and refused with `ErrInvalidParam` the moment they
were chosen.

```go
func (g *Game) AlternativeCostPayableLocked(playerID, castID uuid.UUID, alt *AlternativeCost) bool
```

Three questions, in the order announce asks them: the `Condition`; CR
119.4's life (**exactly N is payable** — paying down to zero is legal and
the state-based action that follows is CR 704.5a's business, not the
view's); and CR 601.2b's card component, counted against the caster's own
hand, graveyard or battlefield with the spell itself excluded (CR 601.2a has
already moved it to the stack, which is what escape's printed "other"
means).

**Mana is deliberately not asked.** CR 601.2g lets the caster activate mana
abilities after the cost is chosen, so "you cannot afford it yet" is not a
reason to withhold the offer — that is what the auto-tapper and the strict
gate are for. Every other component is settled by the board at the moment
the offer is read.

Three readers: `protocol.viewOfAlternativeCosts`, `legal.grantedCastMoves`
and `validateAlternativeCostPaymentLocked`, which shares the life predicate
(`AlternativeCost.LifePayableBy`) and the per-card predicate
(`altCostCardOKLocked`) with it rather than keeping its own. The bot's extra
conservatism — never pay life down to *exactly* zero, because the line loses
the game — stays in `legal`, on top of the rule, labelled as the policy it
is.

### Consequences

- `pay_options` is never present-and-empty on the wire again: an offer with
  nothing to pay it is not offered.
- A card whose only cast path out of a zone is an unpayable offer still
  stamps `castable_here` — an escape card in a graveyard too small to pay
  for it, say. Left for a follow-up rather than argued away: the #978
  amendment above already clears `castable_here` when `cant_cast` is set,
  so the stamp is plainly willing to carry a refusal, and "the bit would
  mean two things" is no longer a reason. The reason it is not done here is
  narrower and is a scope one — the refusal is per-OFFER rather than
  per-card, so the answer is "no payable offer remains AND this zone
  requires one", which needs `zoneBoundAlternativeCosts` at the view and is
  a third thing #695 was not asked for. Filed as #1015; the predicate it
  would call already exists. **Closed by the #1012 / #1015 amendment at the
  foot of this ADR**, which derives the bit from the price list rather than
  adding a fourth reader of the predicate.

## Amendment — 2026-09-19 (#657): madness is the third keyword on the model, and `TimingFlash`'s first user

"Out of scope, stated" listed madness as card work plus a key, and
predicted the one field it would need that nothing else did. Both
halves held. Madness (CR 702.35) shipped on this model with no change
to `CastPermission` at all; what it added lives in
`server/internal/game/madness.go`, and this amendment records the four
decisions taken while building it.

**1. The grant is the whole of "cast it by paying its madness cost".**
`CastPermission{Zone: ZoneExile, AltCostKey: "madness", Cost: <the
madness cost>, Timing: TimingFlash, CastOnly: true}`, over the ONE
exiled object the trigger was about, for the rest of the turn. The key
is the ADR's Decision 3 working as designed: `StackItem.AltCost` reads
back `"madness"`, so CR 702.35b's "if its madness cost was paid"
([#653](https://github.com/krakenhavoc/cmd_and_ctrl/issues/653),
Avacyn's Judgment) will work off the same field a printed keyword
writes. The price is a real alternative cost, so CR 107.3b locks X at
0 through the shared `CastCostFor` rule with nothing madness-specific.

**2. `TimingFlash` is now used, and it means CR 608.2g.** The value
was added by this ADR "for the shape madness will want" and had no
user until now (suspend's grant took it second, for the narrower
reason that its trigger resolves in an upkeep). What it expresses here
is the general clause: a card cast during the resolution of a
triggered ability ignores timing restrictions. A madness sorcery
discarded on an opponent's turn is castable on that turn, and the
sorcery-speed gate in `CastSpell` reads the permission rather than
growing a second branch — which is what the field was for.

**3. The window is bounded on BOTH sides, and the far side is a
delayed trigger.** The ADR's windows all end by expiring. Madness is
the first permission whose card must not merely stop being castable
when the window shuts: CR 702.35b says an uncast card goes to its
owner's graveyard. Declining says so immediately, inside the trigger's
resolution. Accepting and then not casting is the case the
grant-instead-of-inline-cast simplification invents, and it is closed
the way cascade closes its twin — a `DelayedTrigger` at the beginning
of the next end step puts a card still sitting in exile under a
madness grant into its owner's graveyard. The narrowing check is the
grant itself (`AltCostKey == "madness"`), so a card that was cast,
re-exiled or re-granted is left alone.

**4. A keyword that installs a REPLACEMENT as well as a permission
belongs in the engine, not on the card.** Foretell and suspend each
needed one declaration (`Spec.SpecialActions`) because a special
action is one verb. Madness needs two abilities that must agree —
the CR 702.35a replacement over `RepEventDiscard` and the exile-zone
trigger that offers this permission — so the card declares the PRICE
and nothing else (`Spec.Madness string`), and `effects.buildDef` grows
`game.MadnessReplacement()` and `game.MadnessTrigger(cost)` from it.
That is the same bargain the suspend declaration makes for its two
triggers, and it is the reason a madness card file is one field: two
cards cannot spell the keyword two ways and a third cannot forget
half of it.

Declared and unchanged from the ADR's posture on cascade: the accepted
cast is a grant rather than an inline cast, so a CR 117.3b response
window exists between the offer and the cast that paper does not have.
One narrow looseness is recorded in `madness.go`: the trigger cannot
tell a madness exile from a Rest in Peace exile of the same discard,
and offers the cast for both — a choice CR 616.1 gives the discarding
player, where the two lines produce the same board state in this
engine.

## Amendment — 2026-09-19 (#1012 / #1015): the view reads the one price list, and the wire says where the printed cost stands

The #978 amendment above gave exile the same announce surface a hand card
has, by making `stampCastOffers` the one body every zone calls. It left one
thing alone, and #673's `game.CastOffersForLocked` is what makes it fixable:
`stampCastOffers` was still ANSWERING "which prices may this cast claim" on
its own, and its answer differed from the engine's in two places and was
silent in a third.

### The view stops answering and starts reading (#1012)

```go
func (g *Game) CastOffersForLocked(playerID uuid.UUID, card Card, zone ZoneKind, grant *CastPermission) []*AlternativeCost
```

is now `protocol.stampCastOffers`'s only source for `alternative_costs`, as
it already was `legal.castMovesFromZone`'s and as the gates inside it are
`CastSpell`'s. What that deletes:

- **The wholesale overwrite.** The grant's offer was stamped first and the
  card's printed set replaced the whole slice, so a Gravecrawler in the
  graveyard under an Underworld Breach showed only what the card prints and
  the Breach's escape offer was missing from the picker — a cast
  `resolveAlternativeCostLocked` would have accepted. Precedence is now
  stated once, in the engine, in the order announce judges a claim: the
  card's own first, the grant's after, a granted key the card also prints
  dropped rather than listed twice.
- **The second payability filter.** `viewOfAlternativeCosts` kept its own
  call to `AlternativeCostPayableLocked` after #1004 put the rule in one
  place; it is a projection now and decides nothing. Two copies of a filter
  is the shape of the drift this amendment is about, even while they agree.
- **The second question.** `grantedCast` used to answer "may this viewer
  cast this, and at what price". It answers the first half only; the price
  half is derived inside `CastOffersForLocked` from the same permission.

`stampCastOffers` takes the `*CastPermission` rather than the offer it
synthesises, which is what let `stampLegalTargets`'s graveyard and library
gate stop reading the OFFER as a proxy for the permission. A grant that
charges a flat price — Bolas's Citadel's life, an impulse grant that charges
the printed cost — synthesises no offer, so that walk used to skip the card
entirely and `stampGrantedPermissions` then painted `castable_here` back on
with no clauses behind it. A library top under a Citadel now carries its
modes, its target clause and its `cant_cast` like every other cast surface.

### `castable_here` is derived, in one place, from that list plus the gate (#1015)

The #695 amendment above recorded the debt: an offer is only stamped when it
is payable, so a card whose ONLY path out of a zone is a filtered offer kept
a cast surface with nothing behind it, and the announce path refused the
click with `ErrCastCostRequired`. Escape is the canonical case — a graveyard
too small to pay "exile N other cards" is a graveyard the card cannot be
cast from at all.

The bit is now one expression, in `stampCastOffers`, and set nowhere else:

```go
c.CastableHere = c.CantCast == "" && len(offers) > 0
```

An EMPTY price list is a real answer rather than a degenerate one, which is
`CastOffersForLocked`'s own documented contract, and the #978 rule ("a card
the gate refuses is not a cast surface") falls out as the other clause of
the same sentence rather than as a separate assignment in a later pass.

The exception #1015's checklist named holds because the engine's rules make
it hold, not because the view special-cases it: a grant that charges the
printed cost puts a nil entry in the list (rule 4 of
`validateCastPathLocked`), so the card stays a cast surface with an
unpayable offer beside it. The OTHER direction is pinned too, and it is the
one a reader will get wrong: a card that DECLARES and PRICES its own zone
owes that price even under such a grant (rule 3), so an unpayable bound
offer leaves no cast at all. (Reversed by the 2026-10-16 amendment, #1987:
the grant lets such a card be cast for its printed cost too, so it stays a
cast surface on that price.)

`stampGrantedPermissions`'s non-exile branch, which set the bit a second
time behind a `cant_cast` guard, is now a `continue`. It also used to set it
for a permission held by somebody ELSE — a cast surface on a card carrying
no offers, no targets and no gate for any viewer — and that half is filed
as #1022 rather than silently kept: the fix wants exile's per-viewer
`castOffersFor` shape, and a public bit with a per-viewer answer is a wire
decision, not a bug fix.

### The wire says when the printed cost is not claimable (#1012)

`castable_here` is one bit and means "you may cast this from here", never
"you may cast this from here for the cost in the corner". The client
inferred the second sentence from the shape of the offer list, which is
right for a Faithless Looting and wrong for a Gravecrawler under a Breach.

`CardView.alternative_cost_required` (`alternative_cost_required`, omitted
when false) is the missing half, and it is exactly "the nil entry is not in
`CastOffersForLocked`'s list". A flag rather than a nil-keyed entry in
`alternative_costs`, for two reasons: the entry would break every client
that walks that list, and the two questions are genuinely different — the
list is what you may pay INSTEAD, the flag is whether the printed cost is
still on the table. Stamped and stripped with `alternative_costs`, because
it is only meaningful beside it.

Client side, the flag is read in three places and inferred in none:
`AlternativeCostModal` DROPS the "Its mana cost" row rather than greying it
(an option that cannot be taken for this cast is not a choice the player
declined) and defaults the picker to the first offer; the zone browser's
one-way-in button label counts the printed cost as a way in; and
`timing.ts`'s tooltip derivation weighs the printed target clause only when
the printed cost is a price this cast may claim.

Additive on the wire (`v` unchanged).

---

## Amendment (2026-09-19, #1022): a permission names an OBJECT, and the graveyard stamp is its holder's

Decision 1 said a permission is one type with one query and three homes,
and Decision 7 stamped it on the wire. Neither said whose ZONE the
object had to be in, and three surfaces answered that question by
accident instead:

- `CastSpell` resolved `from_zone: "graveyard"` to `p.Graveyard` and
  nothing else, so a card in another seat's graveyard was
  `ErrCardNotFound` however live the permission was;
- `legal/cast.go`'s enumerator walked `p.Graveyard` only, so no bot
  could see the cast either;
- `stampLegalTargets` walked each seat's graveyard asking
  `grantedCast(g, owner, …)`, so the VIEW showed nothing at all —
  no offers, no targets, no gate, for anybody. That is the bug #1022
  was filed for, after #1015 removed the bare `castable_here` the
  non-exile branch used to paint on for a permission somebody else
  held.

**The permission is the key, and the only key.** A `ScopeCards`
permission names an INSTANCE, which is a statement about an object
wherever it sits — Wrexial's "you may cast target instant or sorcery
card from that player's graveyard" is the printed shape, and the model
has always been able to express it. So:

- `castSourceZoneLocked` takes the card ID and, when the caster's own
  graveyard does not hold it, returns the graveyard that does —
  **only** when `CastPermissionForLocked` says the caster may cast it
  from there. A card no permission covers is `ErrCardNotFound`, exactly
  as before, so a client naming a card it has no business naming sees
  no change.
- The enumerator walks every seat's graveyard once
  `AnyCastPermissionsForEffect` is true, with a `mine` flag on each
  pile: a card's own declaration opens only its owner's graveyard,
  because flashback (CR 702.34a), escape (CR 702.138a) and
  Gravecrawler all print "your graveyard".
- The view computes `CastOffersForLocked` for the HOLDER and marks the
  card with `castOffersFor`, which is exile's shape from the #978
  amendment above, on the one per-seat zone that can carry another
  seat's answer. `FilterViewFor` strips them for every other viewer.

**`castable_here` became per-viewer, for those cards only.** The bit is
public everywhere else because everywhere else it is the same answer
for every viewer; on a card whose stamps are one seat's it is that
seat's answer, and shipping it publicly is precisely the thing #1015
took off this surface. `stripCastOffersNotFor` clears it beside the
offers it summarises. Exile never sets it (its button reads
`exile_play`), so nothing else is touched.

**One holder per card.** A `CardView` is one struct, so when the zone's
owner may cast the card too, the owner wins — the view only asks the
other seats once the owner's answer is no — and a second holder sees
the public zone with no stamps. Exile has had the same limitation since
#978 (`CastPermissionOnCardForEffect` answers with one permission), and
lifting it is a wire shape rather than a bug fix.

**And a rule that was true by accident is now written down.** A
STANDING permission is derived from a permanent's printed text and its
`PermissionFilter` has no ownership clause, so Underworld Breach's
"each nonland card in YOUR graveyard has escape" was scoped only by
every caller looking at its own pile. The moment any caller could ask
about another seat's, a Breach would have given escape to the whole
table. `CastPermissionForLocked` now refuses a standing GRAVEYARD
permission over a card its holder does not own (a card in a graveyard
is in its owner's, CR 404.3). `ScopeCards` is deliberately exempt:
naming an instance is naming an object wherever it sits, which is the
whole distinction between the two scopes. Exile is shared by
construction and needs no such rule, and CR 401.5's library permissions
were already pinned to the holder's own library by
`permissionPositionOKLocked` — which is also why a cross-seat LIBRARY
grant still opens nothing.

Additive on the wire (`v` unchanged): the new stamps appear on a card
that carried none, and only for the seat entitled to them.

## Amendment (2026-09-19, #1035 / #1037): the position rule follows the card, and the stamps are per holder

Two halves of the same sentence, and the amendment above is where both
of them were written down as limitations: "a cross-seat LIBRARY grant
still opens nothing" and "one holder per card".

### The library a permission opens is the one the CARD is in (#1035)

Decision 4 put CR 401.5's "the top card of your library" on the
permission, and `permissionPositionOKLocked` enforced it against
`p.Library` — where `p` is the permission HOLDER. That scoped every
library permission anybody had written, because every printed library
clause before Xanathar says "your library", and it did it by accident:
for a permission over somebody ELSE's library the holder's own pile
does not contain the card, so the check failed and the permission
opened nothing, silently, before any of the three surfaces could ask.

A position rule is about a position in a pile, and the pile is the card
owner's (CR 401.1). The check reads the library the card is in. What
that exposes is the same two questions #1022 answered for the
graveyard, plus one the graveyard does not have:

- **Whose pile may a STANDING permission reach.**
  `permissionReachesPileLocked` (was `standingPermissionReachesCard`)
  covers the library as well now. No seat named means the holder's own,
  so two Coursers of Kruphix on one table do not play lands off each
  other's revealed top card. `CastPermission.ZoneOwner` is the way out
  and the only way out: a grant that names a seat reaches that seat's
  pile and no other. It is checked for a STORED standing permission as
  well as a derived one, because the cross-seat grant is stored —
  Xanathar, Guild Kingpin's is made by an upkeep trigger over one
  chosen opponent, until end of turn. A derived one can never name a
  seat, and `standingCastPermissionsLocked` zeroes the field to keep
  that true: the catalog is static, and "each opponent's library" would
  be one stored permission per opponent granted at resolution.

- **Which pile the cast path reaches into.** `castSourceZoneLocked`
  resolves `from_zone: "library"` to the caller's own and, when the
  card is not in it, to the pile it IS in — only under a permission the
  caller holds. One `foreignPileForCastLocked` serves the graveyard and
  the library, because it is one rule; splitting it per zone is how the
  library came to be missing the branch. The bot enumerator walks every
  seat's library TOP once anything has granted anything, with the same
  `mine` flag the graveyard walk carries.

- **The visibility, which is the new one.** Decision 5 keeps CR 401.5's
  "you may look" on the POSITION and derives it from the permanents the
  library's OWNER controls. That cannot express a look at somebody
  else's library: the clause is granted by a resolution to one chosen
  player rather than printed as a static ability, and
  `LibraryTopVisibility` is per library where this is per (library,
  viewer). So a cross-seat grant carries its own —
  `CastPermission.SeesLibraryTop`, Xanathar's "you may look at the top
  card of their library any time" — and `LibraryTopVisibleToLocked` is
  the one place both spellings are read.
  `LibraryTopKnowersLocked` is built out of it per seat, so the seat
  that may cast the top card is a seat the projection has already made
  a knower of it, by construction rather than by two functions
  agreeing. A grant that opens another seat's library top WITHOUT the
  look opens nothing, which is the same guard decision 5 put on a card
  file that declares one half.

No catalog card needs any of this yet; Xanathar is the printed shape
and the fixtures name it. The hole is written down so the next card
file does not fall into it, the way #1022 wrote the graveyard's down.

### The stamps are per holder, and the wire does not change (#1037)

The amendment above said a `CardView` is one struct, so the
announce-time stamps were one seat's answer written into it: the
graveyard took the first seat in seat order and exile the first live
permission. The second holder got the public zone, the public
`exile_play` and no picker, for a cast the engine would accept.

The wire was never the problem — a frame is built per viewer already.
The projection is now too:

- `castStamps` is the announce-time surface as a value (the field list
  `stripCastOffersNotFor` used to clear by hand), and
  `CardView.castOffers` is a map of seat → `castStamps`, replacing
  #978's single `castOffersFor`.
- The exported fields carry the answer that is PUBLIC and nothing else.
  For a graveyard or a library that is the pile owner's own cast out of
  their own pile; for exile there is no owner, so nothing is public but
  the grant. `stampLegalTargets` stamps that one;
  `stampGrantedPermissions` files every other holder, for exile, every
  graveyard and every library's top card.
- `applyCastStampsFor`, in `FilterViewFor`, promotes this viewer's
  entry and drops the rest — only for a knower, since the stamps name
  the card as loudly as its mana cost does. Nothing needs stripping any
  more, because the private answers were never in the exported fields.
- **`exile_play` is resolved per viewer.** It stays public and still
  names a seat, but a holder gets THEIR OWN grant rather than whichever
  live permission came first: it carries the cost override, the faces
  and the any-color clause the cast would actually use, and a client
  reading somebody else's would render the wrong button.

**`castable_here` is the pile owner's answer, or yours.** It stays
public on a per-seat pile, because a flashback cost is printed on a
card in a public zone, and #1022's per-viewer exception disappears with
`castOffersFor`: a holder's private bit rides their `castStamps` and
reaches their frame alone. The client reads the pair — the bit, and
whether `exile_play` names this viewer — in both of its readers
(`zoneBrowser.logic.castableFromZone`, `libraryTop.libraryTopPlayable`),
which is what turns the server's per-holder stamp into a button at all.

Additive on the wire (`v` unchanged), and in the direction the last two
amendments already went: a card that carried one seat's stamps now
carries each holder's own, on their own frame.

## Amendment (2026-09-21, #1055): `castable_here` is the viewer's own answer

The amendment above ended with a sentence that was a bug report as much
as a decision: **"`castable_here` is the pile owner's answer, or
yours."** One bit, two meanings, chosen by which card you are looking
at — the pile owner's on most, the viewer's own on a card the viewer
holds a grant over — and both client readers had to pair it with the
(also public) `exile_play` to find out which one they were holding.
That is the [#891](https://github.com/krakenhavoc/cmd_and_ctrl/issues/891)
shape exactly: the field's NAME is a statement about the viewer and its
VALUE was a statement about somebody else.

**The bit is now per viewer, everywhere.** It is stamped for a seat
that may actually make the cast — the pile's owner for a printed
flashback or escape, a `CastPermission` holder for a granted one, both
of them on their own frames when both are true — and is absent on every
other copy of the same card, spectators and admins included.

### The split is "about the card" versus "about a player"

`stampLegalTargets` still computes the pile owner's answer, once per
card per zone, because that is the only seat a per-seat walk knows
about. What changed is where it goes: `castStamps.applyPublicTo` writes
the half every viewer legitimately sees to the exported fields, and
`CardView.stampsFor` files the WHOLE answer under the owner's own seat,
exactly as `stampGrantedPermissions` already did for a foreign holder.
`applyCastStampsFor` then promotes one seat's entry as before — the
owner's is simply one more entry in the map now, rather than a public
default nobody could opt out of.

**Public**, because a card in a graveyard or on top of a revealed
library is a card every player may pick up and read, and because every
one of these is computed over public state: `alternative_costs` and
`alternative_cost_required`, `modes`, `additional_cost`,
`optional_costs`, `tap_cost`, `target_cost_notes`, `phyrexian_symbols`,
`cant_cast`. An escape offer is priced by the size of a graveyard
everybody can count; a Rule of Law is on the battlefield.

**Per viewer**, because each answers "what may YOU announce":

- `castable_here` — the field this amendment is about.
- `legal_targets` and `clauses` — narrowed by hexproof, shroud,
  protection and "target opponent", so seat A's list is not seat B's to
  read. `applyCastStampsFor` has said that about a SPECTATOR since
  #978 ("a legal target set is not a view of the board, it is a view of
  what one specific player may announce"); this is the same sentence
  applied to the bystander a public stamp used to reach.

They travel together, and they have to: a bit whose offers were
stripped is #1015's button with nothing behind it, and offers beside a
bit that is false are a price list for a cast this viewer cannot make.
Keeping the price list public and the bit private is the only split
where neither half lies.

### What it buys, and what it costs

Both client readers lose their second gate and become one field read:
`castableFromZone(card, zoneKind)` and `libraryTopPlayable(zone)` no
longer take the viewer's or the owner's seat, and `docs/protocol.md`
loses the paragraph that explained the pair-read. `cardAsFace` clears
`castable_here` with the rest of the announce surface, because the
server computed it for the face the grant names.

The cost is that a bystander can no longer see that a card in somebody
else's graveyard is castable BY THEM. Nothing in the client rendered
that, and nothing should: it is the other seat's affordance, it is
visible to them, and a spectator who wants to know whether a flashback
is live can read the card's public price list exactly as a player at a
paper table would.

Additive on the wire in the `omitempty` direction that is safe — the
field goes out on fewer frames, never on more — so a reader that
already falls back to `false` for an absent bit needs no change and
`v` does not move.

---

## Amendment (2026-09-21, #1166): the hand and the command zone take the same path

The amendment above split the announce surface into a public half and
three per-viewer fields, and applied it to the two zones that could
carry a foreign holder's answer — the graveyard and the library top.
It left the hand and the command zone on the pre-#1055 path: one
`castStampsFor` call whose whole answer, `legal_targets` included, went
straight into the exported fields, with no per-viewer promotion at all.

That was two live leaks of one seat's answer, of very different sizes.

**The command zone, with nothing in the way.** It is public — every
seated player is a knower of every card in it — and `FilterViewFor` ran
no strip over it whatsoever. A commander with a target clause put its
OWNER's legal target set on all four frames, for a cast CR 903.4 gives
exactly one seat.

**A revealed hand card, with a hand-rolled strip in the way.**
`keepKnownInHandZone` had been clearing `legal_targets` and `clauses`
by hand since S20, so the leak was covered — by a second list of
fields, in a second place, which is precisely the drift `castStamps`
was created to end. Those two lines are gone and the hand is routed
through `applyCastStampsFor` like every other cast surface.

**What `keepKnownInHandZone` still does, and why that is not the same
thing.** It also drops `modes`, `alternative_costs`,
`alternative_cost_required`, `tap_cost`, `phyrexian_symbols` and
`target_cost_notes` from a revealed card on a non-owner's frame. That
is not the per-viewer split — those fields ARE facts about the card —
it is the documented scope of the fields, which say "the viewer's own
hand". A hand is not a public zone the way a graveyard is: a revealed
card is one card the viewer has been shown, not a pile they may read.
The two narrowings are independent and both stay.

**`castable_here` was never involved.** `castStampsFor`'s switch only
ever sets it for `ZoneGraveyard` and `ZoneLibrary` — every card in a
hand or a command zone is a cast candidate, and an always-true flag
would be noise the client had to ignore — so neither zone has ever
carried one.

The "wide blast radius" this was deferred for did not materialise:
no test read its own hand's announce fields off an unfiltered
`ViewOfGame`. `stampCastOffers` was left with one caller, which was
this one, so its body and its history moved onto `castStampsFor` and
every zone now calls the same function.

Nothing moves on the wire in the direction that breaks a reader: the
fields go out on fewer frames, never on more.

---

## Note (2026-09-21, #1171 / #1169): the zone question is asked of every face, and the hand's public half is an allowlist

Two coherence items out of #1170, both of the S48 #891 shape — a
surface that says one thing and means another — and neither changes a
decision above. They change WHERE two rules live.

### "Does this card open this zone" is one predicate (#1171)

The rule was already written down: a cast out of a non-hand zone needs
either a permission or the CARD's own declaration
(`validateCastPathLocked` rule 2, `CastableZonesFor`). It was being
asked twice, of different faces.

The bot's legal-move enumerator asked it of EVERY castable face,
because an MDFC's halves are separate catalog entries (ADR 0034's
`<oracle_id>#N`) and only the back may print flashback. The view asked
it of `CardView.oracleID` — the BARE oracle ID, which resolves to face
0's entry whatever the other halves declare — so a card whose back
face opened the graveyard was enumerated as a legal move for a bot and
stamped with nothing at all: no `castable_here`, no price list, no
target clause, and a zone browser with no button behind a cast
`CastSpell` would have accepted. Silent, and latent: no catalog card
declares it today, and the view's coherence fixture (#1024) seeded
only single-faced cards, so the two answers were never compared for a
card that has more than one.

`game.CardCastableFromAnyFace` is that question, once, read by
`legal/cast.go` and by `protocol.stampLegalTargets`. It is the same
move #992 made one question over — `game.CastableFacesUnder` is the
one answer to "which faces may a cast choose" — and for the same
reason: a face one side offers and the other does not is either a bot
move the announce path refuses or a picker row with nothing behind it.

The fixture has a modal DFC whose back face declares AND prices the
graveyard now, and `viewPrices` reads the union over the card's block
and its `faces[i]` blocks, which is the shape the enumerator's per-face
walk produces on the other side of the comparison.

### A hand's public half is narrower, and it is an allowlist (#1169)

The #1166 amendment above left `keepKnownInHandZone` clearing six
announce fields from a hand-rolled list of field names. That narrowing
is right — a hand is not a public zone the way a graveyard is, and a
revealed card is one card the viewer has been SHOWN, not a pile they
may read — but a list in the per-viewer filter is a second list of
cast-surface fields in a second place, which is the drift `castStamps`
exists to end. Being a list of what to REMOVE, it had already gone
stale twice: it never covered `optional_costs` (added by ADR 0073
after it), and since #992 it never covered the per-face blocks at all,
so a knower of a revealed adventure card read its owner's whole
per-face price list — including an offer's `pay_options`, which for a
pitch cost is a list of instance IDs out of the hand the viewer was
shown exactly one card of.

Shape (1) of the issue, the conservative one: `castStamps.publicIn`
takes the zone kind, and a HAND's public half is narrower than a
graveyard's. `keepKnownInHandZone` goes back to being purely "drop the
cards this viewer is not a knower of" and carries no field list at
all.

The narrowing is written as an ALLOWLIST (`handPublicCastSurface`) —
the four fields that stay, rather than the six that go — because the
question a new announce field has to answer is "may somebody who was
shown this card read it", and the safe default for a field nobody has
thought about is no. What stays is what is not cost-shaped:
`target_mode` (the printed prompt shape, and already public on a
revealed card's faces since #992), `additional_cost` and
`optional_costs` (printed clauses whose pickers read the public
battlefield), and `cant_cast` (a Rule of Law on the battlefield, which
everybody can see).

A reflection guard in `face_down_view_test.go` places every field of
`CastSurfaceView` on one side of that line and fails on one nobody has
placed, in the shape the redaction allowlist beside it already uses.

Nothing moves on the wire in the direction that breaks a reader: the
fields go out on fewer frames, never on more.

### What is NOT in these two

`modes` and `alternative_costs[i]` carry a `legal_targets` of their
own, computed for the seat the stamp was built for, and those ride the
PUBLIC half on a public pile — so a bystander reads the pile owner's
per-mode legal target set off a modal card in a graveyard. That is the
#1055 sentence one level down inside a nested view, it needs a
decision about nested per-viewer data rather than a field move. Filed
as #1172.

The CLIENT half of #1171 is filed as #1173: the zone browser's cast
gate reads the card's `castable_here`, which for a pile is face 0's
answer, so it will not offer the cast a back face opens even though
the frame now carries one on `faces[i]`.

## Note (2026-09-22, #1172): the nested legal sets ride the same per-seat split

The note above ended by filing the one thing it did not fix: `modes`
and `alternative_costs` stay PUBLIC on a public pile and each carries a
`legal_targets` of its own, computed for the seat the stamp was built
for. A modal card in a graveyard shipped the pile owner's per-mode
legal set to every viewer.

### The shape, and why it is not either of the two the issue named

#1172 named two candidate shapes and neither is what landed:

1. *"The nested sets go private"* — described as a deep copy of two
   view structs on every public stamp, and "the public half stops
   being one assignment". The first half is right and is the cost
   paid; the second is not. `applyPublicTo` was already
   `publicIn(kind).applyTo(c)`, and `publicIn` was already the one
   function that knows which fields are which (#1169). Adding two
   lines to it keeps one place, and the promotion on the other side
   stays exactly one assignment of the whole `CastSurfaceView`, which
   is what hands a seat its nested sets back without a second
   mechanism.
2. *"The whole field goes private"* — refused, for the reason the
   issue gives: `modes` is the printed text of a modal card and a
   bystander at a paper table reads it off the graveyard.

So: **the nested per-viewer fields ride the same per-seat `castStamps`
/ `castOffers` split as the top-level ones.** The public projection
carries the mode's and the offer's static facts; each viewer's frame
gets its own nested lists, promoted with the block around them. It is
#1170's move one level further in — an embedded `CastSurfaceView` per
face is how the split travels down to a face, and this is how it
travels down to a field inside one.

### Which fields, and the rule that decides

Four, and they are the four that are computed from the BOARD for one
seat rather than read off the card:

- `modes[i].legal_targets` and `modes[i].clauses` — narrowed by
  hexproof, shroud, protection and "target opponent" exactly as the
  card's own clause is.
- `alternative_costs[i].legal_targets` — the clause the spell has when
  this price is paid, resolved against the board through the same
  `LegalTargetsForEffect`.
- `alternative_costs[i].pay_options` — NOT a target list (a cost does
  not target, CR 601.2h) but a list of instance IDs picked out by "you
  control" / "your hand" / "your graveyard", so it answers "what may
  YOU pay". #1169 already called it out as the live leak inside a
  revealed hand card's offer list, where the whole offer list is
  dropped; this is the same sentence on a public pile, where the offer
  stays and the list does not.

Everything else in both blocks is printed text — the prompt, the
counts, the labels, `target_mode`, the offer's key, mana cost, life,
pay label, CR 107.3b's X lock and CR 107.4's Phyrexian count — and
stays public with its parent. The rule, stated once: **does this field
come off the printed CARD, or off the board for one PLAYER.**

### The copy is load-bearing

`publicIn` takes its stamp by value, but `Modes` is a POINTER and
`AlternativeCosts` a slice header, and the public projection and the
asking seat's own answer come out of ONE `castStampsFor` call
(`stampLegalTargets` stamps publicly and then files the same value for
the seat). Blanking through the pointer would take the nested sets off
the OWNER's frame as well as off the bystander's. That is
`applyFaceCastStampsFor`'s bug — a shared backing array written in
place — arriving one level further in, and `publicModeSpec` /
`publicAlternativeCosts` copy for the same reason. One allocation per
modal card and per priced card per public stamp, and none for a card
that is neither.

### The guard

`face_down_view_test.go`'s `castSurfaceScopes` table places every field
of `CastSurfaceView` on the public / private line and fails on a field
nobody has placed. Two more tables do the same for `ModeOptionView` and
`AlternativeCostView`, so a field added to a nested block is a decision
somebody makes rather than one nobody notices — and the same test
asserts that the public strip did not reach through into the seat's own
copy.

The #1024 coherence fixture gains a modal card in a graveyard and a
modal clause on the modal DFC's back face, so the bystander loop asks
the nested question per card and per face; the owner's half of the same
fixture asserts all four lists survive the promotion.

### Consequences

Nothing on the wire moves in the direction that breaks a reader: four
fields go out on fewer frames and never on more, so `v` does not move
and a client that ignores them behaves as it did. Every client reader
of the nested sets takes a `CardView` out of the viewer's own snapshot,
so all of them were already reading their own frame; the one behaviour
worth pinning — that a targeted-looking bullet arriving with no legal
set opens no picker rather than a free-form prompt — now has a test.

## Amendment — 2026-09-22 (#1195): the timing rule is per PLAYER, and CR 307.1 is asked in one place

Decision 6 above put a timing rule on a permission — `CastPermission.Timing`,
one of `TimingNormal` / `TimingFlash` / `TimingSorcery`, read next to
`HasKeyword(&card, "flash")` in `CastSpell` — and that is the right shape for
a grant that opens ONE cast of ONE object: madness (#657) and suspend's free
cast (#659) are both exactly that. It is the wrong shape for the sentence
Vedalken Orrery prints. "You may cast spells as though they had flash" names
no card, names no zone, and outlives no particular object; it is a statement
about a PLAYER. So is its inverse, which the same seam blocks from the other
side: "each opponent can cast spells only any time they could cast a sorcery".

`docs/engine-seams.md`'s row **Per-player "cast as though it had flash"** is
nine recorded cards; the audit puts 15 of 21 behind it. This amendment builds
it, and the shape it builds is the one this ADR already uses twice.

### 1. `CastTiming` is a per-player statement, and it reuses three vocabularies

```go
type CastTiming struct {
    Player   uuid.UUID
    Timing   GrantTiming      // TimingFlash, TimingSorcery, TimingYourTurnOnly
    Filter   PermissionFilter // the zero filter is "spells"
    FromZone ZoneKind         // zero is "from anywhere"
    Duration Duration
    Affects  CastTimingAffects
    Source, SourceName, Label
}
```

Nothing here is new vocabulary, and that is deliberate:

- **`GrantTiming`** is Decision 6's own enum, gaining one value.
  `TimingYourTurnOnly` is "you can cast spells only during your turn"
  (Dosan the Falling Leaf), which is **not** `TimingSorcery` — Dosan leaves
  you every instant-speed window on your own turn and takes away the rest.
  No `CastPermission` declares it, exactly as nothing declared `TimingFlash`
  when this ADR reserved it.
- **`PermissionFilter`** is Decision 1's, gaining `NoncreatureOnly` and
  `SorceryOnly`. A timing statement narrows by card type the same way a
  standing permission does ("you may cast CREATURE spells as though they had
  flash", Yeva), and a second flag struct saying `CreatureOnly` again would
  be two spellings of one predicate.
- **`Duration`** is ADR 0063's, unchanged. Emergence Zone is
  `UntilEndOfTurn`, Teferi's +1 is `UntilYourNextTurn`, and a derived
  statement is `WhileInZone` — the same three `CastPermission.Duration`
  carries, swept by the same `durationExpiredLocked`.

`Affects` is the one field with no precedent, and it exists because a
catalog entry is static and cannot name a seat: `TimingAffectsYou` (the
source's controller — Orrery, Leyline, Yeva), `TimingAffectsEachOpponent`
(Teferi, Time Raveler; Teferi, Mage of Zhalfir) and `TimingAffectsEachPlayer`
(Dosan). The derivation expands it against the battlefield and stamps
`Player`; a STORED statement is granted to a player by name and carries
`TimingAffectsYou` by construction.

### 2. Two homes, and they are the two this ADR already has

**Derived, never stored** for a statement whose duration is a permanent's
presence: `Spec.CastTimings` → `CardDef.CastTimings` → `CatalogCastTimings`,
walked per query through **`CatalogAbilityKey`** — so a Vedalken Orrery under
a CR 613.1f ability-removing effect stops granting, two Orreries compose, and
one leaving cannot revoke the other's grant. Exactly Decision 1's third home
and exactly `standingCastPermissionsLocked`'s argument.

**Stored on the player** (`Player.CastTimings`) for a statement that outlives
its source: Emergence Zone sacrifices itself and the permission lasts the
turn; Teferi's +1 resolves and the planeswalker may die. Granted by
`GrantCastTimingForEffect`, swept at the same two moments
`sweepCastPermissionsLocked` runs (cleanup, and the beginning of a turn), by
the same expiry function. Cloned and snapshotted beside `CastPermissions`,
for the same reason: who may cast when is not derivable from the board.

### 3. ONE read, and CR 101.2 decides the order inside it

```go
func (g *Game) CastTimingOpenLocked(playerID uuid.UUID, card Card,
    zone ZoneKind, perm *CastPermission) bool
```

"May this player begin to cast this card, out of this zone, right now?"
(CR 307.1). Four steps, in this order, and the order IS the rules:

1. **The card.** An instant, or a card with flash (CR 702.8), is
   instant-speed. This is the `HasKeyword(&card, "flash")` read that used to
   sit inline in `CastSpell`.
2. **The permission** (Decision 6, unchanged). `TimingFlash` opens it,
   `TimingSorcery` shuts it. Madness and suspend reach the window through
   this branch exactly as before.
3. **The per-player GRANTS.** Any live `TimingFlash` statement naming this
   player and covering this card and zone opens the window.
4. **The per-player RESTRICTIONS, last, because CR 101.2 says "can't" beats
   "can".** A `TimingSorcery` statement shuts the window whatever step 3
   said; a `TimingYourTurnOnly` statement refuses the cast outright when it
   is not this player's turn. An opponent's Vedalken Orrery does not get them
   past your Teferi, and that falls out of the placement rather than needing
   a rule of its own.

Then: a shut window means `sorcerySpeedOpenLocked` must be open.

**A land play is not here.** CR 305.1 and CR 116.2a make playing a land a
special action, not a cast, and every card in this row writes about casting
SPELLS. `CastSpell`'s land branch keeps its own `sorcerySpeedOpenLocked`
check beside this call, which is the same split `CastGateLocked` documents
for the same reason: a Dosan that stopped a land play would be a rule nobody
printed.

### 4. Three callers, which is the whole point

The same three ADR 0073 §7 names for the cast gate, and for the same reason
— the read sits BESIDE the gate rather than inside it, because "you cannot
cast this" and "you cannot cast this **yet**" are different answers and the
client greys them differently:

- **`CastSpell`** (`mutations.go`), replacing the inline
  `requiresSorcerySpeed` computation. Refuses with `ErrSorcerySpeedRequired`,
  unchanged.
- **`legal.castMovesPayingOptional`** (`legal/cast.go`), replacing its copy
  of the same three lines, so a bot is never offered a cast the engine will
  refuse for timing — and never denied one an Orrery opens.
- **`protocol.castStampsFor`** (`view.go`), where `castable_here` gains its
  third input. #1015 derived that bit from two things (a claimable price, and
  the cast gate); it is now three, and the third is this predicate. A
  flashback sorcery in a graveyard is a cast surface on your main phase and
  not in an opponent's end step — which is what the announce path has always
  answered and what the wire did not say.

### 5. Out of scope, stated

- **Activated abilities.** Several of these cards print an "activate
  abilities only as a sorcery" clause beside the cast clause (Grand
  Abolisher, Teferi, Mage of Zhalfir's sibling family).
  `ActivatedAbility.SorcerySpeed` (`activated.go:337`) is a per-ABILITY flag
  read by `activated.go:713`, and a per-player statement about activations
  would want the same two homes and the same one read this amendment builds
  for casts. It is a sibling, not a part: nothing in this row is blocked on
  it, and a card that prints both gets the cast half and a caveat.
- **A per-spell COUNT.** Winding Canyons is "until end of turn, you may cast
  **a** creature spell as though it had flash" — one spell, not a window.
  There is no allowance for the read to consult (the same gap
  `CastTally` fills for Rule of Law and does not fill here), and a duration
  that expired on use would be a sixth `DurationKind` with one user.
  Caveated.
- **"As though it had flash" said of ONE named card.** Every card on the row
  says it of a player. A per-object version is `CastPermission.Timing`, which
  already exists and is untouched.
- **The wire does not grow a field.** `castable_here` already means "you may
  cast this from here", and "not at this timing" is a reason it is false, not
  a second bit. A client that wants to distinguish "banned" from "not yet"
  already has `cant_cast` for the first.

---

## Amendment — 2026-09-23 (#1208): the ACTIVATION twin, and why it is a second type

The 2026-09-22 amendment above built CR 307.1's per-player half for CASTS and
named its sibling as out of scope: *"Activated abilities. …
`ActivatedAbility.SorcerySpeed` is a per-ABILITY flag read by `activated.go`,
and a per-player statement about activations would want the same two homes and
the same one read this amendment builds for casts."* This amendment builds it,
and it turns out to want the same SHAPE, one of the two homes, and a type of
its own.

### The cards, which are three and not the three the issue predicted

`docs/engine-seams.md`'s closed row and issue #1208 both name Grand Abolisher
and the "activate abilities only as a sorcery" family. **Both were stale by the
time this was built.** Grand Abolisher's activation half — "During your turn,
your opponents can't activate abilities of artifacts, creatures, or
enchantments" — landed with #1210 as an `ActivationRestriction`, because it is a
BAN and not a timing statement, and the card ships `full`. Every other printed
restriction on somebody's activations says "can't be activated" (Cursed Totem,
Linvala, Collector Ouphe, Pithing Needle, Karn, Damping Matrix) and goes through
the same gate.

What has no home is the GRANT, and the whole printed population is eight cards,
of which three are in scope here:

- **The Wandering Emperor** — "As long as The Wandering Emperor entered this
  turn, you may activate her loyalty abilities any time you could cast an
  instant." `caveats` since S14 for exactly this, and the seam's only caveated
  card.
- **Teferi, Master of Time** — "You may activate loyalty abilities of Teferi on
  any player's turn any time you could cast an instant." The same clause with
  no condition.
- **Leonin Shikari** — "You may activate equip abilities any time you could
  cast an instant."

The other five are Forge Anew (the same equip clause plus a cost gap), Teferi,
Temporal Archmage and Teferi's Talent (an EMBLEM carries the statement), Jace's
Machinations, and an Un-set card. Two of those want a second home and are named
under Scope below.

### Decision 1 — two TYPES, one vocabulary

`game.ActivationTiming` (`server/internal/game/activation_timing.go`) is not
`CastTimingRule` with a "casts / activations / both" scope field, and the reason
is the homes rather than taste:

```go
type ActivationTiming struct {
    Label      string
    Timing     GrantTiming                  // ADR 0066's enum, unchanged
    Covers     func(q ActivationQuery) bool // #1210's query, unchanged
    ActiveWhen Designation                  // ADR 0071's gate, unchanged
}
```

A cast timing statement has a STORED home — Emergence Zone's "this turn",
Teferi, Time Raveler's +1 — so it must be pure data, which is why its narrowing
is a `PermissionFilter`, a `ZoneKind` and a `CastTimingAffects` enum. An
activation timing statement has only a DERIVED home (Decision 5), so `Covers`
can be a PREDICATE — and it has to be, because what the printed cards narrow on
is the ability's SOURCE and the ABILITY, two things a filter over *the object
being cast* says neither of. Folding both into one type would have given every
field two meanings and made every read start by asking which half it was
looking at, which is the argument ADR 0073's #1210 amendment already makes for
`ActivationQuery` not being a widened `CastQuery`.

So the shape is copied and the type is not. `GrantTiming` gains no value,
`ActivationQuery` and `ActivationAbility` are #1210's, the collection walk is
`ActivationRestrictionsForCard`'s line for line, and the fold order is this
amendment's §3. There is no `Affects` enum here: "you" is
`q.Controller == q.Source.Controller`, written once in the two constructors in
`cards/effects/activation_timing.go`, which is the spelling
`OpponentsSourcesCantActivate` next door already uses.

### Decision 2 — the narrowing is asked of the query, not of a filter

Issue #1208 asked what `PermissionFilter` would filter on here, and the answer
is that it would filter on the wrong noun: on the cast side it narrows the
object being cast, and on this side the two things a card narrows are

- **which SOURCES** — "loyalty abilities of Teferi" (`q.Card`), and
- **which ABILITIES** — "loyalty abilities", "equip abilities" (`q.Ability`).

Both are already fields of `ActivationQuery`, so `Covers` asks them. What
`ActivationAbility` gained is two bools and a third:

```go
SorcerySpeed bool // CR 602.5d, the ability's printed clause
Loyalty      bool // CR 606.3, derived from AbilityCost.Loyalty
Equip        bool // CR 702.6, set by EquipAbility
```

`Loyalty` is the field ADR 0073's #1210 scope note predicted — *"a restriction
on LOYALTY abilities as a class … is a bool on `ActivationAbility`, not a second
gate"* — arriving for the timing read first. `ActivationAbilityOf(shape)` is the
one place a shape becomes an identity, so CR 606.3 cannot be spelled differently
in the three callers. `Equip` names a KEYWORD rather than adding a kind: equip
stays an ordinary activated ability, and `EquipOnlyAbility` plus
`TestEveryEquipAbilityIsMarked` exist because three cards wrote their narrowed
equip out by hand and a hand-written equip is one that can forget a field.

### Decision 3 — a mana ability is not in this window at all

Not a carve-out, and not #1210's answer. CR 605.3a gives a mana ability its own
window — whenever its controller has priority, AND whenever a payment is being
made, inside a cost, mid-resolution — and that is not the CR 602.5d window a
timing statement opens or narrows. So `ActivationTimingOpenLocked` returns true
for `Ability.Mana` before it walks anything: a fast POSITIVE, and a rule rather
than a safety valve.

The card the issue worried about is real and is already handled elsewhere: Grand
Abolisher stops mana abilities too, with no "unless they're mana abilities"
clause, and it does so through `ActivationGateLocked`, where #1210 put the mana
decision **on the card** because half the printed cards exempt them and half do
not. The two functions differ here because they answer different questions: the
gate says an activation is BANNED and the card decides whether that reaches mana
abilities; this read says an activation is not open YET, and for a mana ability
the question does not arise.

### Decision 4 — a grant DOES reach loyalty abilities

Issue #1208 proposed the opposite, on the grounds that CR 606.3 is a rule and
not a printed clause any effect overrides. CR 101.1 says otherwise, and so do
the cards: both planeswalkers on this row print exactly "you may activate
loyalty abilities … any time you could cast an instant". A read that refused to
reach them would make the only two printed users of the seam unwritable.

What must not happen is a statement about "abilities" generally silently opening
a loyalty ability, and that falls out of the predicate rather than out of a rule
anyone has to remember: both constructors test `q.Ability.Loyalty`, and a
statement that does not ask is not about them. CR 606.3's OTHER half — one
loyalty activation per turn per planeswalker — is untouched and keeps its own
check, before this one.

### Decision 5 — ONE home, because that is all the cards want

The cast side has two homes because Emergence Zone sacrifices itself and the
permission outlives it. Nothing on this side does: every card that prints an
activation timing statement is a permanent whose static says it, for as long as
it is there. So `Spec.ActivationTimings` → `CardDef.ActivationTimings` →
`CatalogActivationTimings`, walked per query through **`CatalogAbilityKey`**,
and nothing is stored — no `PlayerStatic` payload, no fourth kind of entry on
#1197's slice, no sweep, no clone, no snapshot field. Two Shikari compose
(harmlessly — the verdict is a bit), a source under a CR 613.1f ability-removing
effect stops saying it, and one bounced in response shuts the window before the
activation is validated.

"As long as she entered this turn" rides the derivation rather than a duration:
`game.EnteredThisTurn` (#1009's per-object entry tally), read live. It is
deliberately **not** `Card.SummonedThisTurn`, which survives until its
controller's untap step — and the Emperor has flash, so she lands on somebody
else's turn nearly every time, and the marker would have given her instant-speed
loyalty abilities for a whole turn cycle. `Designation` could not have expressed
it either: `Designation.Active(c Card)` takes a Card and nothing else, and
"entered this turn" lives on the Game.

### Decision 6 — ONE read, and CR 101.2 decides the order inside it

```go
func (g *Game) ActivationTimingOpenLocked(activator uuid.UUID, card Card,
    zone ZoneKind, ability ActivationAbility) bool
```

1. **Mana abilities are not asked about** (Decision 3).
2. **The ability's own timing** — `SorcerySpeed` (CR 602.5d) or `Loyalty`
   (CR 606.3) answer to the sorcery window; everything else is instant-speed
   (CR 117.1b).
3. **The per-player GRANTS.**
4. **The per-player RESTRICTIONS, last, because CR 101.2 says "can't" beats
   "can".** Nothing declares them today; the placement is the rule stated once
   rather than a rule to be discovered the day a card does. `TimingYourTurnOnly`
   refuses outright rather than narrowing, exactly as `CastTimingOpenLocked`
   reads it.

Then a shut window means `sorcerySpeedOpenLocked` must be open. It sits BESIDE
`ActivationGateLocked` rather than inside it, for the reason ADR 0073's #1195
note gives about the cast pair: "banned" and "not yet" are different answers and
`cant_activate` means the first.

### Decision 7 — four callers, and the wire grows one field

`ActivateCatalogAbility`, `legal.abilityMovesForSource` (which dropped the
`speed` parameter it threaded through two functions to keep a copy of the rule),
`protocol.viewOfActivatedAbilities`, and **`ActivateLoyalty`** — the sandbox
manual loyalty verb, which `internal/legal` skips by design but which is still
CR 606.3's window, and which a statement about "loyalty abilities of
planeswalkers you control" reaches on a walker the catalog has never heard of.
The same argument `gatherTapSources` is the activation gate's fourth caller
under. Its sorcery-speed check moved below the battlefield lookup, because the
read needs the object.

**The wire grows `activated_abilities[i].timing_closed`**, where the cast side
needed nothing. `castable_here` already existed and already meant the engine's
answer; the activation rows carried only `sorcery_speed`, the ability's PRINTED
clause, and `client/src/lib/timing.ts` re-derived CR 307.1 to grey them — its
own comment called it *"the LAST rules derivation left in this file"*. A
per-player statement is board state the client cannot see, so the bit has to
come from the server. It is NEGATIVE and `omitempty`, so it is absent on every
instant-speed row; the client still writes the SENTENCE (no priority, split
second, a non-empty stack, somebody else's turn) from the snapshot, because that
is in the snapshot and is what a player wants to read. The sandbox loyalty rows,
which have no ability row at all, keep the client-side predicate.

### Scope, stated

- **The EMBLEM home.** Teferi, Temporal Archmage's −10 and Teferi's Talent both
  make an emblem that says "you may activate loyalty abilities of planeswalkers
  you control on any player's turn any time you could cast an instant". An
  emblem is an object that exists for as long as the statement does, so it is
  the DERIVED home one zone over — the walk would gain `Player.Emblems` beside
  the battlefield, and `EmblemSpec` a slot. Not built: `EmblemSpec` copies only
  `Static` and `Triggered` today, the one card behind it also needs a
  library-look prompt that does not exist for its +1, and a slot with no user is
  a slot that drifts. **Built by #1275** — see the 2026-09-24 amendment at the
  end of this ADR.
- **A RESTRICTION with a duration**, and a restriction at all. No catalogued
  card declares `TimingSorcery` or `TimingYourTurnOnly` on this side, because a
  printed activation restriction says "can't be activated". The fold is written
  and tested; the constructors are not, and a card that needs one should add the
  constructor beside the two that exist.
- **"During your turn" as a narrowing** exists on the constructor
  (`EquipAbilitiesAtInstantSpeed`'s `onlyDuringYourTurn`) and is unused: Forge
  Anew prints it, and also prints a `{0}` equip cost, which is the
  "cost modification for activated abilities" row and not this one.
- **Thousand-Year Elixir's "as though those creatures had haste"** is a
  different seam — CR 302.6's tap-symbol restriction, not CR 602.5's window —
  and nothing here touches it.

## Amendment — 2026-09-23 (#1318): `TimingPlot`, a window no grant widens

Aven Interrupter's "exile target spell. It becomes plotted." needs CR 702.170d:
the owner may cast the card from exile without paying its mana cost "during
their main phase while the stack is empty during any turn after the turn in
which it became plotted". Every clause but one is a field this ADR already
has: the owner as `Player`, exile as `Zone` under `ScopeCards`, `Cost: "{0}"`,
`WhileInZone`, and a `NotBeforeTurn` floor. `game.PlotExiledCardForEffect`
(`game/plot.go`) builds the permission. [ADR 0013 §5ad](0013-replacement-effects.md)
records the rest of #1318.

The one missing clause is the timing, and `TimingSorcery` is the wrong answer.
Decision 3 of the 2026-09-22 amendment puts the per-player GRANTS after the
permission's override, so Vedalken Orrery's "as though they had flash" beats
a permission's `TimingSorcery`. That is right for a madness or suspend cast,
where the timing is the card's. It is wrong for a plotted card. The plot rule is
the permission's own window, so a plotted instant, a plotted card with flash,
and a plotted card under an Orrery are all cast only in their owner's main
phase with the stack empty.

`TimingPlot` is a fourth `GrantTiming` value. `CastTimingOpenLocked` answers it
before step 1 and returns `sorcerySpeedOpenLocked` without reading the card or
the grants. The restrictions (step 4) cannot narrow it, because Dosan's "only
during your turn" and Teferi's "only as a sorcery" are both already true of the
window. Only a `CastPermission` may carry it. A per-player statement has no use
for it, and `effects.Register`'s timing check does not offer it.

The `NotBeforeTurn` floor is `Turn.Number` or `Turn.Number + 1`, depending on
whether the card was plotted on its owner's own turn. `Turn.Number` counts
rounds, and the window only opens on the owner's turn, so this gives exactly
"any later turn". An extra turn the owner takes in the same round waits a round.
That is weaker, never stronger. `plot.go` explains the floor.
`TestAFlashGrantDoesNotWidenThePlotWindow` is the back-out proof: with
`TimingSorcery` it fails.

**Later amendment — 2026-09-23 (#1342): the plot keyword reuses this
permission.** The hand special action (CR 702.170a) is now built as the fourth
CR 116.2 kind. Its performer exiles the card face up and calls
`PlotExiledCardForEffect` on it, so the permission, `TimingPlot` and the floor
above are shared by both routes to the plotted state. A card can only be
plotted from hand on its owner's own turn, so its floor is always
`Turn.Number + 1`. [ADR 0062's 2026-09-23 plot
amendment](0062-abilities-and-special-actions-from-the-hand.md) has the
per-kind rows.

---

## Amendment — 2026-09-23 (#1314): a standing permission gated by a Class
level or a condition

Decision 1 and `standingCastPermissionsLocked` gave a permanent's printed
"you may cast/play …" a home — derived from the battlefield, re-evaluated on
every query — and never asked whether the permanent HAS the ability right
now. Every other gateable slot in the catalog (`StaticAbility`,
`TriggeredAbility`, `ActivatedAbilityShape`, `CostModifier`) carries an
`ActiveWhen Designation` (ADR 0071) precisely so a Class's level-2 line, a
solved Case's line, or a station's threshold line is not offered before it
exists. `CastPermission` was the one slot ADR 0071 missed, and Fortune
Teller's Talent's level 2 — "As long as you've cast a spell this turn, you
may play cards from the top of your library" — needed it and something ADR
0071 does not name at all: a card-specific "as long as …" clause that is not
a Class level, a solved Case, or a charge-counter threshold.

### Decision 1 — the gate is a SEPARATE type, not two more fields on `CastPermission`

The obvious change is `ActiveWhen Designation` and `Condition func(...) bool`
added directly to `CastPermission`. It compiles, and it fails
`snapshot_drift_test.go`'s `TestSnapshotMirrorsHaveNoFuncs`: `CastPermission`
is dual-purpose — a catalog declaration on one path (`CardDef.CastPermissions`,
never snapshotted) and, on the other, the exact type STORED on
`Player.CastPermissions` and mirrored verbatim into `GameSnapshot`. A func
field on the shared type poisons the stored side even though no stored
instance would ever set it: the guard reflects over the TYPE, not over which
instances happen to be nil.

So the gate lives on `CastPermissionGate`, a wrapper reachable only from
`CardDef` and a new hook, never from anything a snapshot touches:

```go
type CastPermissionGate struct {
    Permission CastPermission
    ActiveWhen Designation
    Condition  func(g *Game, controller, source uuid.UUID) bool
}

var CatalogGatedCastPermissions func(oracleID string) []CastPermissionGate
```

The same posture `ActivatedAbility.Condition` and `TriggeredAbility.AppliesTo`
already have — both are catalog-only closures on catalog-only types — one
struct over, because `CastPermission` is the one permission-shaped type that
is also player state.

**`CatalogCastPermissions`'s existing signature is untouched.** Seven test
files across three packages stub it as `func(oracleID string) []CastPermission`;
widening it (or wrapping every entry) would have meant migrating all seven for
a gate exactly one card uses today. The new hook is additive:
`standingCastPermissionsLocked` walks both, in order — the ordinary hook
first (byte-for-byte the pre-#1314 loop), then, if `CatalogGatedCastPermissions`
is non-nil, the gated one, `activeOnly`-filtered by `ActiveWhen` and then by
`Condition` — and both funnel into one shared stamping step
(`stampStandingPermissionLocked`) so a gated and an ungated permission from
the same card cannot disagree about what Scope, ZoneOwner, Source or Duration
a "standing" permission means.

### Decision 2 — `ActiveWhen` and `Condition` are independent tests, in that order

`ActiveWhen` is asked first, through the same `activeOnly` helper
`StaticAbilitiesForCard` and `CostModifiersForCard` already use, for the same
reason: it is the cheap, common-case test (almost every card has no gate at
all) and it decides whether the entry is even a candidate. `Condition` runs
only on what survives — Fortune Teller's Talent's level-2 line asks it
exactly once the Class is level 2 or greater, never at level 1, which is
observable: a test asserting the closure's call count pins it
(`TestGatedStandingPermissionNeedsBothTheLevelAndTheCondition`,
`cast_permission_test.go`).

Two independent booleans rather than one merged predicate, because CR 716.2a
gates the LINE ("as long as this Class is level 2 or greater, it has …") and
"as long as you've cast a spell this turn" gates the CLAUSE printed on that
line — two different rules, from two different parts of the Comprehensive
Rules, and folding them into one closure would have made a future card that
needs `ActiveWhen` alone (a Case's solved line that grants a plain permission)
write a `Condition` that always returns `true` to get there.

`Condition`'s signature mirrors `ActivatedAbility.Condition`
(`func(g *Game, controller, source uuid.UUID) bool`) rather than inventing a
third shape: `controller` is the permission-holder asking ("you" in "you've
cast a spell"), `source` is the permanent contributing it — not necessarily
the same seat if the permanent changes hands, which is why the walk passes
the CURRENT controller rather than a captured one.

### Decision 3 — the fast negative has to see BOTH hooks

`AnyCastPermissionsForEffect` — the check the enumerator and the view take
before walking every graveyard, every library and the whole of exile — used
to answer only from `CatalogCastPermissions`. A card gated ENTIRELY behind
`CatalogGatedCastPermissions` (Fortune Teller's Talent has no ungated
permission at all) made this answer `false` regardless of whether the gate
was open, which skipped the library walk outright — not "offered nothing
because the gate is shut", but "never asked". `TestEnumeratorOffersAGatedLibraryTopPlayOnlyWhenBothHalvesHold`
(`internal/legal`) and `TestCastableHereWaitsOnAGatedPermission`
(`internal/protocol`) both caught this in the writing of this amendment — the
first draft passed the game-package model test and failed both surface tests,
because the model test calls `standingCastPermissionsLocked` directly and
never goes through the fast negative at all.

The fix does not evaluate the gate in the fast path: `AnyCastPermissionsForEffect`
answers "could anything open one of the expensive zones", not "does one
apply right now" — evaluating `ActiveWhen`/`Condition` there would just move
the question the walk exists to ask into the wrong function, for a check
whose entire purpose is being cheaper than the walk it guards.

### Decision 4 — `Spec.GatedCastPermissions`, a slot beside `Spec.CastPermissions`

Card files declare a gated entry through a new `Spec` slot rather than
widening `Spec.CastPermissions`'s element type, for the same reason
`CastPermissionGate` is a separate engine type: every existing card using
`Spec.CastPermissions` (Realmwalker, Bolas's Citadel, Courser of Kruphix,
Oracle of Mul Daya, Underworld Breach) keeps its literal unchanged.
`gatedStandingCastPermissions` mirrors `standingCastPermissions`'s Scope/Duration
normalisation one level down, into `CastPermissionGate.Permission`.

`specDesignations` (the effects package's cross-slot walk that backs
Register's Room-door-gate refusal, ADR 0071 decision 3) grew a fifth source —
`spec.GatedCastPermissions[i].ActiveWhen` — so a card that gated a permission
on a door the engine cannot yet honour fails at boot exactly as one gating a
static or a trigger on it already does.

**Cards shipped:** Fortune Teller's Talent (`caveats` → `full`; level 1's
`LibraryTopVisible: game.LibraryTopOwner` and level 2's `GatedCastPermissions`
entry together close both of the card's remaining gaps, since level 2's
permission needs level 1's visibility to open anything at all).

---

## Amendment — 2026-09-23 (#1316): the CAST-BAN twin

`cast_gate.go`'s `CastGateLocked` answers CR 101.2's "can beats can't" for a
cast from two sources — a static on a permanent (`CastRestriction`) and the
spell's own condition (`CastConditionFor`) — and its own doc comment named
the gap this amendment closes: *"BANS WITH A DURATION. Silence's 'this turn'
and Reflector Mage's 'until your next turn' want the turn-scoped and
permanent-duration registries. A third source slots into `castRestrictionsLocked`
without changing this function's signature; that is the extension point."*
Avatar's Wrath ("Until your next turn, your opponents can't cast spells from
anywhere other than their hands") and Mandate of Peace ("Your opponents
can't cast spells this turn") are exactly that shape: a ban with a CR 611.2
duration, created by a resolving spell that is gone — often exiled by its own
text — a moment after it grants the ban.

### Decision 1 — the fourth payload on `PlayerStatic`, not a new registry

This ADR's own 2026-09-22 amendment built `CastTimingRule` for the sibling
question — a per-player statement about WHEN a cast is legal — and put it on
`PlayerStatic` rather than on a registry of its own, for the argument ADR
0085 Decision 1 makes at length for the life-total lock one payload over: a
statement about a PLAYER for a CR 611.2 duration wants the slice that already
has a duration, a sweep, a clone and a snapshot field, not a fourth of each.
A cast BAN is the same statement pointed the other way — CR 101.2's "can't"
rather than "may" — so it takes the same home: `PlayerStatic.CastBan
CastBanRule`, told apart from `Keyword`, `Timing` and `LifeTotalLocked` by its
own presence bit (`CastBanRule.Kind`, zero value `CastBanNone`) for the reason
those three doc comments already give and `CastBanRule`'s own repeats: its
zero value otherwise ("no exception, no count") IS a real statement — Mandate
of Peace's outright ban — not "nothing to say", so it cannot borrow a
sentinel off an existing field the way `LifeTotalLocked`'s plain bool does.

### Decision 2 — `CastBanRule`'s two shapes, and the third the issue named without a card

```go
type CastBanRule struct {
    Kind           CastBanKind // CastBanOutright | CastBanMaxPerTurn
    Filter         PermissionFilter
    ExceptFromZone ZoneKind
    MaxPerTurn     int
}
```

`CastBanOutright` with `ExceptFromZone` zero is Mandate of Peace. With
`ExceptFromZone: ZoneHand` it is Avatar's Wrath — "from anywhere other than
their hands" is a ban that reaches every zone but one, which is why the field
is an EXCEPTION rather than a target list the way `CastTimingRule.FromZone`
narrows a GRANT to one zone: a grant naming a zone opens only that one, a ban
naming an exception closes every other one, and reusing `FromZone`'s own
meaning here would have inverted it silently on the read side.

`CastBanMaxPerTurn` is the seam issue's third named shape — "each player
can't cast more than one spell each turn," granted rather than printed on a
permanent — with no catalogued card behind it yet. Built and tested
(`TestMaxPerTurnBanReadsTheSameTallyTheStaticRestrictionDoes`,
`internal/game`) against the model rather than a card, because the issue
named it as a shape the storage has to support, not as a card to ship; it
reads the identical `Game.CastTallyFor` tally `EachPlayerMaxSpellsPerTurn`
(the printed, derived version, `cast_restriction.go`) already reads, so a
granted cap and a printed one can never disagree about what "one spell this
turn" counts.

### Decision 3 — one reader, and it IS the third source `CastGateLocked` already reserved

```go
func (g *Game) castBanForbidsLocked(playerID uuid.UUID, card Card, zone ZoneKind) (label string, source uuid.UUID, forbidden bool)
```

Walks `p.Statics`, skips every entry that is not a `CastBan` or whose
duration has expired (the same "test it here too, not only in the sweep"
posture `playerLifeTotalCantChangeLocked` and `castTimingVerdictLocked` take,
for the identical reason: the sweep is hygiene at known moments and the
reader has to be right between them), and returns the first live entry that
forbids — `CantCastError` carries one reason, and CR 101.2 does not ask which
"can't" arrived first among several.

`CastGateLocked` calls it once, between the battlefield's static
`CastRestriction`s and the spell's own `CastConditionFor` — external "can't"s
before the card's own, which is the order the function already documents for
the two sources it had. **No new caller was needed anywhere else**: unlike
`CastTimingRule`, which had to add itself to three separate call sites
(`CastSpell`, `legal.castMovesPayingOptional`, `protocol.castStampsFor`)
because CR 307.1 had no single existing choke point, `CastGateLocked` already
IS the one function ADR 0073 §7 built for exactly this question, and it
already has all three callers. Extending its insides extends all three at
once, and the wire needs no new field: `cant_cast` already means "an effect
prevents this cast", stamped from `CastGateLocked`'s own error, so a card
silenced by Avatar's Wrath greys exactly the way one silenced by Rule of Law
already does (`TestCantCastIsStampedFromAGrantedBan`, `internal/protocol`).

`AnyCastRestrictionsForEffect` — the fast negative beside the gate — grew a
matching check (`anyLiveCastBanForEffect`, one pass over the seats) for the
same reason Decision 3 of the companion amendment above extended
`AnyCastPermissionsForEffect`: a table with nothing on the battlefield but a
live granted ban must not answer `false`.

### Decision 4 — one primitive, `RestrictCasting`, one grant per opponent

`effects.RestrictCasting{Player, Rule, Label, Duration}` is the card-facing
wrapper, calling `GrantCastBanForEffect` — the same shape `LockLifeTotal` and
`GainPlayerKeyword` already take for their own `PlayerStatic` payloads. "Your
opponents can't cast spells" is granted once PER OPPONENT
(`for _, opp := range ctx.Opponents()`), not as one table-wide statement:
`PlayerStatic` is per-seat by construction, and a card that meant the
CASTER too would say "each player", which neither proof card does.

**Cards shipped:** Avatar's Wrath (`full`) and Mandate of Peace (`full` —
CR 724.2's "end the combat phase" (#1317) landed alongside this seam via PR #1343, so
the card ships with neither half caveated).

---

## Amendment (2026-09-23, [#1369](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1369)): a permanent's ability rows keep their hidden-zone lists for the controller

Every amendment since #1055 drew one line through the announce surface of a
card in a PILE: a fact about the card is public, and what one seat may
announce or pay rides `castOffers` to that seat alone. The BATTLEFIELD never
had the line drawn, because nothing there needed it. `stampActivatedAbilities`
computes a permanent's `activated_abilities` and `mana_abilities` once, with
the controller as "you", and a permanent is known to every seat — so every
row reached every viewer intact, and `FilterViewFor`'s per-card redaction
(#95) only ever removes them from a card the viewer cannot identify.

That was true until three fields started reading the controller's HAND:

- `activated_abilities[i].discard_cost_options` — "Discard a creature card"
  (Fauna Shaman, #660);
- `mana_abilities[i].discard_cost_options` — "Discard a card" (Skirge
  Familiar, #1213);
- `mana_abilities[i].exile_cost_options` — "Exile a card from your hand"
  (Cadaverous Bloom, #1283);
- and, since #1297 landed while this was in flight,
  `activated_abilities[i].exile_cost_options` when its `exile_cost_zone` is
  `"hand"` — "Exile a card from your hand" (Holistic Wisdom).

For an unfiltered clause the list's LENGTH is the hand size, which is public
(CR 402.3 hides the cards, not the count). For a filtered clause it is not:
Fauna Shaman's list told every opponent how many creature cards its
controller held, and its instance IDs were a handle on specific cards that
had been revealed, bounced or seen in another zone — the "the ID alone is
the leak" argument PR #513 made about pending-choice options, arriving on a
permanent.

### Decision 1 — the same carrier, one zone over

`CardView.abilityOffers` is `castOffers`' twin for the battlefield: a map
keyed by seat, holding that seat's WHOLE `activated_abilities` /
`mana_abilities` rows. `CardView.fileAbilityOffers` runs last in
`stampActivatedAbilities`, after every mana stamp that writes into the rows
in place; it leaves the public half on the exported fields and files the
controller's full rows under their seat. `FilterViewFor` promotes one seat's
entry through `applyAbilityOffersFor`, which is `applyCastStampsFor`'s rules
unchanged:

- **Only the controller's seat has an entry.** A battlefield ability is
  activated by its controller alone (CR 602.2), so no other seat has an
  answer to file, and an opponent's frame keeps the public rows.
- **The empty viewerID gets nothing private.** A spectator, an admin and a
  replay reader get the public rows — the same posture `legal_moves` and the
  cast surface take, and the one place this departs from the issue's text,
  which proposed leaving spectators on "see everything". A spectator is not a
  seat, and no seat's hand is theirs to count; the unseated viewer who most
  needs this redaction is the one watching over a player's shoulder.
- **Only a knower is promoted**, so a face-down permanent's rows that the
  redaction has just cleared are never handed back to a non-knower. Its
  controller is always a knower (CR 708.5).

Promotion is two slice-header assignments on this viewer's copy of the card,
and the public copies are ALLOCATED rather than blanked in place, because the
controller's rows are the slice already on the card — #1172's "the copy is
load-bearing", one surface over. Nothing is filed or allocated for a
permanent whose rows carry no hidden list, which is nearly all of them.

### Decision 2 — the line is "does this field read a zone the viewer cannot see"

`publicActivatedAbilityRow` and `publicManaAbilityRow` are the one place the
hidden fields are named, and a scope table in
`ability_row_privacy_view_test.go` places EVERY field of both views
(the embedded `CounterCostView` included) as public or hidden-zone, failing
on a field nobody placed. The sweep that table records:

- **Hidden-zone, now per seat:** the discard lists always, and the exile
  lists whenever `exile_cost_zone` is not `"graveyard"`. Nothing on either
  view reads a LIBRARY.
- **Public by zone:** an exile list stamped `exile_cost_zone: "graveyard"`
  (Grim Lavamancer, Moorland Haunt) lists cards in a pile every viewer may
  pick up and read, so it stays on the public row on both views.
- **Public, and why:** the counts (`discard_cost_n`, `exile_cost_n`,
  `counter_cost_n`) are the numbers PRINTED in the clause; the labels are
  the clause's words; `counter_cost_max` and every other option list —
  `sacrifice_options`, `crew_options`, `return_options`,
  `tap_others_options`, `waterbend.options`, `counter_cost_options` — is read
  off the battlefield, which every viewer can count for themselves; the
  verdicts (`condition_unmet`, `timing_closed`, `cant_activate`,
  `exhausted`, `charged_mana_cost`) read public state, as their own field
  comments already say; and an activated ability's `legal_targets` never
  reaches a hand (`game.zonesOfKindLocked`'s hidden-zone caution: nothing in
  the catalog targets a card in hand).

`exile_cost_options` is split BY THE ZONE it names, through one predicate,
`exileListIsHidden`, which both strips call. It is written as "anything but
the graveyard" rather than "the hand", so a row stamped with no zone — or
with a pile added tomorrow — is private until somebody decides otherwise.
That is the direction that does not leak. The graveyard list stays public
because this amendment's line is "does it read a zone the viewer cannot
see", and a graveyard is not one. #1172 moved a graveyard `pay_options` onto
the asking seat's frame for a different reason: a cast surface on a pile
answers for whichever seat asks, and that seat varies. A battlefield ability
has exactly one payer, and its graveyard is on the table for everyone.

The battlefield-read option lists stay public on purpose. They are also only
the controller's to pay with, and by #1172's rule they could move too; they
name nothing an opponent cannot already see, and moving them would make every
permanent with a sacrifice or crew clause file a per-seat copy for no
information gain. The table is where that decision is written down, so the
day one of them starts reading a hidden zone is a table edit somebody has to
make.

### Consequences

- The wire moves only in the safe direction: three fields go out on fewer
  frames and never on more, so `v` does not move. Every client reader of the
  lists opens a picker on the viewer's OWN permanent, which reads the
  viewer's own frame; nothing renders an opponent's cost options.
- The enumerator is untouched — it reads the game, not the view — and
  `TestControllerFrameAndEnumeratorAgreeOnHandCosts` pins that the
  controller's frame still offers every card the enumerator would pay with,
  for all three fields that read the hand on a fixture.
- The catalog proof test covers Fauna Shaman, Skirge Familiar, Cadaverous
  Bloom and Holistic Wisdom (hidden), with Grim Lavamancer as the control
  (graveyard, public on every frame).
- The crash-recovery dump and a pinned replay marshal the UNFILTERED view, so
  they now carry the public rows: a replay reader sees no hand lists, which
  is what an unseated viewer should see anyway.
- #1297 (PR #1373) merged first and added `ExileCostN` / `ExileCostLabel` /
  `ExileCostOptions` / `ExileCostZone` to the activated view and
  `ExileCostZone` to the mana view. They are placed in the scope table:
  the count, label and zone are public, and the options fall under the zone
  rule above.

## Amendment — 2026-09-24 (#1275): the activation twin's derived home, one zone over

The #1208 amendment above built per-player activation timing with ONE home —
a permanent on the battlefield, read through `CatalogAbilityKey` — and named
the emblem as out of scope. Teferi, Temporal Archmage's −10 is the card:
*"You get an emblem with 'You may activate loyalty abilities of planeswalkers
you control on any player's turn any time you could cast an instant.'"*
(Teferi's Talent grants a −12 that makes the same emblem.)

### Decision 1 — it is still Decision 5's ONE home, at a second address

An emblem exists in the command zone (CR 114.2), its abilities function there
(CR 114.3), and nothing removes it but its owner leaving the game (CR 800.4a).
So its presence IS the duration, which is the whole of Decision 5's argument
for deriving rather than storing. Nothing new is stored: no `PlayerStatic`
payload, no sweep, no clone or snapshot field (the emblem itself is already
carried, ADR 0064 Decision 7). What changed is WHERE the derivation looks:
`activationTimingVerdictLocked` folds the battlefield and then every seat's
`Player.Emblems`, with the same fold closure, so the verdict is still three
independent bits and the walk order still does not matter.

The emblem half goes through `ActivationTimingsForCard` like the battlefield
half. For an emblem `CatalogAbilityKey` is its `CatalogKey` — nothing can name
an emblem to remove its abilities — so the ability-removal test costs a nil
check and never fires. ADR 0064's #1315 amendment makes the same observation
for the untap- and draw-step walks.

### Decision 2 — the slot is `EmblemSpec.ActivationTimings`, guarded like the Spec's

`EmblemSpec` gains `ActivationTimings []game.ActivationTiming`, projected by
`buildEmblemDef` onto the emblem's `CardDef` and reachable through the
synthetic `emblem:<oracle>` key by the existing `CatalogActivationTimings`
hook. `Register`'s three refusals (a statement that says nothing, has no
label, or covers nothing) moved into `checkActivationTimings`, which both
`Spec.ActivationTimings` and `EmblemSpec.ActivationTimings` go through, so the
two homes cannot drift in what they refuse. `checkEmblemSpec`'s "an emblem
needs an ability" accepts a timing statement as one.

`Covers` receives the emblem as `q.Source`, so "you" is `q.Source.Controller`
— the emblem's owner, which never changes. The constructor is
`effects.LoyaltyAbilitiesOfYourPlaneswalkersAtInstantSpeed`:
`ThisSourcesLoyaltyAbilitiesAtInstantSpeed`'s three narrowings with the
self-reference (CR 201.5) widened to "a planeswalker whose controller is the
emblem's owner, activated by that player".

### Decision 3 — no reader changes, and CR 606.3's count is not the window

Decision 7's four callers — `ActivateCatalogAbility`,
`legal.abilityMovesForSource`, `protocol.viewOfActivatedAbilities`
(`timing_closed`) and the sandbox `ActivateLoyalty` — already ask
`ActivationTimingOpenLocked`, so all four see the emblem without an edit. The
wire does not change.

The emblem opens the WINDOW. CR 606.3's other half, one loyalty ability per
planeswalker per turn, is `Game.LoyaltyActivatedThisTurn`, read after the
window in both activation paths and in the enumerator. It still holds on every
turn the emblem opens, for each planeswalker separately.

### Out of scope, stated

- **Teferi's Talent's granted −12.** The emblem it makes is this one, but the
  ability lives on the ENCHANTED planeswalker, and a static that grants an
  activated ability to another permanent is the open "Abilities granted to
  other permanents" seam. The card ships with its draw trigger and a caveat.
- **Other emblem timing statements.** None is printed that is not this one;
  the constructors cover what the cards say.

Proof card: Teferi, Temporal Archmage (`full`). Tracker
[#883](https://github.com/krakenhavoc/cmd_and_ctrl/issues/883).

## Amendment — 2026-09-24 (#1573): a face-down exile the holder may look at, and "mana of any TYPE"

Two gaps in the exile-and-cast family, both found building the Edea deck
(#1565, slice E2). Each left a card weaker than printed or showed the table
more than it should see, so three cards shipped with caveats: Hostage Taker
(gap 2), Gonti, Night Minister and Outrageous Robbery (both gaps).

### Decision 1 — `FaceDownPermitted`, an exile kind whose viewer is the permission holder

*"… exiles it face down. They may play that card for as long as it remains
exiled"* (Gonti); *"… exiles the top X cards of their library face down. You
may look at and play those cards"* (Outrageous Robbery). ADR 0069's viewers
table had no row for this. `exiled` is nobody, `foretold` is the owner, and
`hideaway` (ADR 0091) is the controller of another object. The viewer here is
**whoever the cast permission names**. So there is a new exile-side kind,
`FaceDownPermitted` (`"permitted"`). Like the other two exile kinds it has no
CR 708.2 body and keeps its catalog entry, because casting it is the point.

**The grant stamps the look.** The primitive lands the card face down with
no knowers and grants the permission once the card is in exile. The grant has
to come second, because it names the object's post-move epoch (Decision 2).
`GrantCastPermissionForEffect` is the one write path for every permission. It
now adds the holder as a knower of any `permitted` card the new ScopeCards
exile grant names (`stampPermittedViewersLocked`). So the rule sits with the
write, not with one primitive, and a later grant over the same card gets the
look too. It touches **only** this kind. Necropotence's `exiled` card stays
unreadable even if a permission happens to name it. That is pinned by
`TestAGrantDoesNotOpenAPlainFaceDownExile`. `faceDownViewersLocked` answers
the same set, derived from the stored permissions
(`castPermissionHoldersLocked`), so the kind table stays the only place that
defines who may look.

A **knower stamp rather than a live derivation**, for CR 406.3's reason: a
player who may look keeps looking until the card leaves exile, and every
permission that makes this kind lasts exactly that long. Nothing needs a sweep
like hideaway's, because a permission's holder cannot change hands. The card
turns face up when it leaves exile by any route (MoveCard's `ClearFaceDown`).
A cast puts it on the stack face up, known to every seat.

**The primitive** is `Game.ExileTopFaceDownWithPermissionForEffect(from,
grantTo, n, perm)`, reached from the catalog as `effects.ExileTopWithPermission{FaceDown:
true, WhileExiled: true}`. It picks the top n cards **up front** and moves each
one by ID through the shared exit primitive, chained through the route's
continuation. That is `millPlanLocked`'s argument: a commander among the cards
is offered the command zone (CR 903.9), the question leaves it on top of the
library, and re-reading the top would return the same card. Granting from the
continuation also means a commander whose owner declines is exiled and granted
when the answer arrives. It does not land with no grant, which was #1336's
defect. A card that went anywhere but exile gets nothing.

**Nothing downstream changed.** The cast path, the enumerator and the view
already cast a face-down exiled card under a permission (foretell, hideaway).
The view's non-knower redaction already strips `exile_play` and the whole cast
surface off a card the viewer cannot read, and the log already names a card
only to its knowers. So the owner and every other seat get a card back
labelled `permitted` and nothing about casting it. The holder gets the card,
`exile_play`, `castable_here` and `cast_prices`.

**The unseated frame is not a leak.** The spectator frame (`viewerID == ""`)
knows every card by FilterViewFor's documented contract (#191). That holds for
every face-down exile the engine makes, Necropotence's and a foretold card
included, and this kind follows it. What the unseated frame never gets is a
cast surface: `castable_here` and `cast_prices` are per seat (#1055), and the
view test pins that.

### Decision 2 — `CastPermission.AnyType`, folded in the one pricer

`AnyColor` is *"as though it were mana of any color"*, and `asAnyColorCost`
deliberately leaves `{C}` alone because colorless is not a color (CR 106.1b).
Hostage Taker, Gonti and Outrageous Robbery print *"mana of any TYPE"*, and
colorless IS a type. So a stolen Thought-Knot Seer or Reality Smasher needs
no real colorless mana. `CastPermission.AnyType` is a separate bool rather
than a widened `AnyColor`, for two reasons. Breeches and the rest of the
any-colour family must keep their `{C}` exception. And a stored field that
changed meaning would reprice every snapshot already on disk. `AnyType`
implies `AnyColor`, and a permission setting both is read as `AnyType`.

`spendAsThoughAny(grant, cost)` is the one reading. It applies the any-type
fold (every requirement, `{C}` included, into generic), the any-colour fold,
or nothing. It is called in `printedCostLocked`, which is where the any-colour
fold always lived. That makes it ADR 0066's one pricer (the #696 amendment):
the hand payment, the auto-tapper (`applyAutoTapLocked` →
`effectiveCostLocked`), the auto-tap preview, the bot enumerator's
affordability probe and the view's `cast_prices` all read the same folded
cost. None of them needed an edit.

**One inherited trade, stated.** A Phyrexian slot folds to generic under
either clause, so its "or 2 life" half is lost. `asAnyColorCost` has always
done this. It is weaker than printed, never stronger, and no card this
amendment touches has a Phyrexian symbol.

### Decision 3 — the wire grows `exile_play.any_type`

`ExilePlayView.AnyType` (`any_type`) is a label beside `any_color`. When it
is set, `any_color` is set too, so a client that reads only the older key
still says "any colour". Neither key changes what the client may do.
`castable_here` and `cast_prices` already reflect the fold. The zone
browser's tooltip reads the new key. [protocol.md](../protocol.md) documents
both the key and the `permitted` kind.

### Undo, clone and persistence

Nothing new is stored. The kind rides `Card.FaceDownKind`, the look rides
`Card.KnownBy`, and `AnyType` is a field of the stored `CastPermission`. All
three were already cloned and snapshotted verbatim (Decision 9;
`TestEmbeddedDomainTypesStayPureData` still holds). The game tests pin the
round trip and an undo of both the exile and the cast.

The on-disk shape grows one key, `seats[].castPermissions[].anyType`,
recorded in `testdata/snapshot_shape/v7.txt` with no schema bump. Both
directions degrade safely. A file written before the key restores it as
false. A binary from before the key drops it and restores an any-colour
grant, because every card that sets `AnyType` also sets `AnyColor`. So an old
binary makes the card weaker (a `{C}` needs real colorless mana again), never
stronger.

### Out of scope, stated

- **The face-up impulse primitive's raw `MoveCard`.**
  `ExileTopWithPermissionForEffect` still skips the CR 903.9 window, so a
  commander impulse-exiled off the top of its owner's library is not offered
  the command zone. The face-down primitive above goes through the route. The
  face-up one is unchanged because it returns the exiled IDs synchronously to
  callers that read them on the next line.
- **A permission-holder look that outlives a SHORTER grant.** Every card that
  makes this kind grants "for as long as it remains exiled". A future card
  that exiles face down with an until-end-of-turn grant would keep its
  holder's look after the grant lapsed. That is CR 406.3's rule, not a bug,
  but no card needs it yet.

Proof cards: Hostage Taker, Gonti, Night Minister and Outrageous Robbery, all
`full`. Tracker [#885](https://github.com/krakenhavoc/cmd_and_ctrl/issues/885);
deck tracker [#1565](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1565).

## Amendment — 2026-09-28 (#1589): a Phyrexian symbol keeps its life half under a spend grant

The #1573 amendment stated one inherited trade: under either "spend mana as
though" clause, a Phyrexian slot folded to generic and lost its "or 2 life"
half. That was weaker than printed, never stronger, but it was wrong. CR 107.4f
says a Phyrexian symbol, hybrid ones included, is paid with one mana of its colour **or** 2 life. A grant that says
"spend mana as though it were mana of any color" (or "of any type") widens
which mana pays the first half. It says nothing about the second half.

Before this, a Dismember (`{1}{B/P}{B/P}`) that Gonti exiled cost three mana.
Worse, the view still stamped `phyrexian_symbols: 2` from the printed cost, so
the client offered a life payment that `strikePhyrexianLifeLocked` then refused
as an over-claim.

### Decision — widen the slot, don't fold it

`spendAsThoughAny` is still the one reading, called in `printedCostLocked`,
the one pricer. Both folds now keep every Phyrexian requirement and mark it
`ColorRequirement.AnyMana`. The slot stays Phyrexian, so `PhyrexianSymbols`
counts it and the strike may claim it for 2 life. Its mana half admits any
mana. Everything else folds exactly as before: coloured slots always, `{C}`
under any type only (CR 106.1b). A `{C}` under any colour is still not payable
by coloured mana.

`Options` stays as printed, and so does `String()`. So `cast_prices` and the
preview's `cost` render `{1}{B/P}{B/P}` rather than `{3}`, and the
missing-mana breakdown names `{B/P}`. There is no Scryfall symbol for "any
mana or 2 life". The printed symbol plus the grant's `any_color` / `any_type`
label is the honest rendering.

**One predicate for the mana half.** `ColorRequirement.Admits(color)` is
`AnyMana || matchColor(color, Options)`. Every payment path asks it instead of
reading `Options`: the pool solver (`attemptSpend`, `MissingFor`), the
auto-tapper's slot match (`pickMatchingSlot`), the Phyrexian strike's ranking
(`PhyrexianLifePlan`), pay-time colour picks (`mostRestrictiveRequirement`,
which also sorts a widened slot as the widest), and convoke's colour cover
(`payerCoversRequirement`). So the hand payment, the auto-tapper, the auto-tap
preview, the bot enumerator's affordability probe and `cast_prices` still
agree without a line of their own. They all read the one pricer's cost, and
that cost is paid through one predicate.

**A widened slot is paid last.** The pool solver pays requirements greedily,
first match wins. A slot that admits anything, walked before a narrower one,
could take the only token the narrower one needed. Picture a cost modifier
appending `{W}` after the fold: without the ordering, the wildcard would take
the Plains and the `{W}` would go unpaid. So the folds emit widened slots
last, and `widenedLast` reorders them at the pool solver and at the
auto-tapper's `solveColored` too. The auto-tapper backtracks, so there the
ordering saves search budget rather than answers.

### What does not change

Nothing new is stored. `AnyMana` lives on a `ParsedCost` the pricer computes
on every read, and no snapshot or wire struct carries a `ColorRequirement`.
The wire keys are unchanged; only the content of a `cast_prices` string under
a grant changes. [protocol.md](../protocol.md) says so beside
`exile_play.any_type`. A cast with no grant never sets `AnyMana`, so its
requirements are exactly the parse.

### Out of scope, stated

- **The cast enumerator never pays a spell's Phyrexian symbol with life.**
  `legal/cast.go` prices the mana path only, while `affordablePayment`
  (`legal/abilities.go`) offers the life path for activated abilities. A bot
  holding a stolen Dismember and one land gets no move, grant or not. That
  gap was there before this amendment and is not this seam.
- **Convoke under a spend grant.** A convoked creature pays a widened slot
  when it has any colour. A colourless creature still pays it only as
  generic, as it would with no grant. No printed card puts convoke and a
  Phyrexian symbol on a stolen spell.

Tracker [#887](https://github.com/krakenhavoc/cmd_and_ctrl/issues/887).

---

## Amendment — 2026-09-28 (#1665): miracle, a hand permission that opens one printed offer

Miracle (CR 702.94) is madness moved one zone over: a keyword that lets a
player cast a card for an alternative cost during the resolution of a
triggered ability. It is built from the pieces this ADR and its #657
amendment already have, and adds three small things. The code is
`server/internal/game/miracle.go`.

**1. The price is a printed offer, and the permission is the right to
claim it.** A miracle card declares
`AlternativeCosts: []game.AlternativeCost{Miracle("{W}")}`. Like
`Flashback` and `Escape`, the constructor bundles the rewrite with the
price. For miracle the rewrite is one new field,
`AlternativeCost.RequiresGrant`: the offer can be claimed only while a live
`CastPermission` for the card object carries the same key.
`validateCastPathLocked` checks this. `CastSpell` and `CastOffersForLocked`
both call it, so the view, the bot enumerator and the announce path cannot
disagree about when `{W}` is on the menu. A refused claim returns
`ErrAltCostNotGranted`. Madness puts the price on the permission (`Cost`).
Miracle cannot do that: the card sits in a hand, where every cast of it
would read the permission, and `CastCostFor` would then reprice the
printed cast too. So the miracle permission carries no `Cost`, and the
printed offer does the pricing.

**2. The hand is a permission zone for exactly one kind of grant.**
`CastPermissionForLocked` no longer answers nil for a hand. It returns
a live stored `ScopeCards` permission naming the card object, and nothing
else. A standing permission is never read there. The hand still needs no
permission to be cast from (CR 601.2). What the miracle grant adds is
`TimingFlash` (CR 608.2g), which lets a Terminus drawn in an opponent's
upkeep be cast there. The same timing would turn a hard-cast Terminus into
an instant, so a hand permission is scoped to its own claim by
`CastPermission.ForClaim`. That returns nil for a cast that claims any
other offer. `CastSpell` narrows as soon as the claim is resolved, before
the price, the CR 107.3b rule and the timing gate. The enumerator narrows
per offer in `castMovesForCard`. So the printed cost keeps sorcery timing
while the grant is live, and a bot is offered the miracle cast at instant
speed but the printed one only where a sorcery could be cast. Every
non-hand permission passes through `ForClaim` unchanged.

**3. The trigger watches draws from the hand.** `TriggeredAbility.Zones`
gains `ZoneHand` (the #925 dimension; `supportedTriggerZones`). The
per-event index keeps the walk narrow: a hand is walked only for
`EventDrawCard`, and only when some registered card declares a hand
trigger. "The first card you've drawn this turn" is read off
`Game.DrawnThisTurn`, which `actuallyDrawCardLocked` appends to before it
emits the event. It is a fact about the draw, not a count the harvest has
to trust. A draw on another player's turn counts. The CR 702.94a "you may
reveal" is the trigger's `OptionalPrompt` ("Reveal Terminus for its
miracle cost {W}?"), so the client needs nothing new. A "yes" reveals the
card to the table through `RevealForEffect` and puts a keyed item
(`miracle/offer`, ADR 0041 P9) on the stack. The item carries the revealed
object as `Params.Object` (instance and CR 400.7 epoch). `effects.buildDef`
grows the trigger from the `Miracle` cost, the same bargain madness makes:
a card file declares the keyword once.

**The window, and where it is still wider than paper.** CR 702.94b: a
card that has left the hand is not the card that was revealed. That
includes a card that left and came back, which has a new epoch. The
resolution then grants nothing. When it does grant, the permission names
the object, so the cast itself ends it (CR 400.7). As with madness, the
cast is an ordinary `cast_spell` after the trigger resolves, not an inline
cast. The window madness bounds with an end-step cleanup is shut much
tighter here, because an open-ended "{W} at instant speed" in a hand would
be stronger than printed:

- the holder **passing priority** closes it (`closeMiracleWindowLocked`,
  called from `PassPriority` before the pass runs, because that pass can
  resolve the next miracle trigger);
- the permission's zero `Duration` still ends it at the end of the turn
  at the latest.

What is left is CR 117.3b. When the miracle card was drawn on another
player's turn, the active player receives priority first after the trigger
resolves. They may act before the owner casts, and the owner may then cast
the miracle in response. Madness declares the same gap. For a draw-step
miracle, where the owner is the active player, nothing is wider.
`Game.AdvanceStep` (the sandbox step driver) passes through
`passPriorityLocked` and not `PassPriority`, so a window left open when a
player advances the step by hand lasts to the end of the turn.

**Out of scope, stated.** The client's cost picker still lists "Its mana
cost" beside "Miracle {W}" while the grant is live. `alternative_cost_required`
stays false, because the printed cost is claimable, just not at every
moment. Picking it outside a sorcery window gets the ordinary
`ErrSorcerySpeedRequired` toast. Topdeck the Halls ("decorated cards in
your hand have miracle {S}") grants the keyword and is not legal in
Commander. Temporal Mastery waits on extra turns (the engine cannot add a
turn). Revenge of the Hunted was waiting on "all creatures able to block it do
so", which #1597 has since shipped, so it is one card file away and not
built here. Also still open: Bonfire of the Damned on a "target player
or planeswalker" clause whose sweep follows the chosen target's
controller.

Tracker [#885](https://github.com/krakenhavoc/cmd_and_ctrl/issues/885).

---

## Amendment — 2026-10-02 (#1600): a player's "spend mana as though it were mana of any color"

The #1573 and #1589 amendments made "spend mana as though it were mana of
any color" a property of one PERMISSION: Breeches, Hostage Taker and Gonti
grant it to cast one spell, and `spendAsThoughAny` folds it in the one cast
pricer. Three catalogued or catalogue-ready cards print the same words with
no spell attached:

- **Chromatic Orrery** — "You may spend mana as though it were mana of any
  color." Its controller, every cost.
- **Mycosynth Lattice** — "Players may spend mana as though it were mana of
  any color." Every player, every cost.
- **Oath of Nissa** — "You may spend mana as though it were mana of any color
  to cast planeswalker spells." Its controller, one kind of payment.

CR 609.4b: such an effect changes how a player may pay a cost. It does not
change the cost, and it does not change the mana. The Orrery rulings add the
two edges that matter here: any mana pays a coloured symbol, colourless
included, and mana with a "spend this only on …" restriction keeps it.
Colourless is not a colour (CR 106.1b), so coloured mana still cannot pay a
`{C}`.

### Decision 1 — a battlefield static, read live, on `Spec.AnyColorSpend`

`game.AnyColorSpendStatic` is `{Label, Whose, Covers}`. `Whose` is the
controller (`AnyColorSpendYou`) or every player
(`AnyColorSpendEveryPlayer`). `Covers` is an optional predicate over the
payment's `ManaSpendContext`; nil covers every payment, and Oath of Nissa's
reads `Purpose == cast` and the type `Planeswalker`. The card files use three
constructors (`effects/any_color_spend.go`).

It is DERIVED, the way Leyline of Sanctity's hexproof and Platinum Emperion's
life lock are. `spendsManaAsAnyColorLocked` walks the battlefield on each
question, keyed by `CatalogAbilityKey`. Two Orreries compose, one leaving
cannot revoke the other's grant, an Orrery that loses its abilities grants
nothing, and nothing is stored, cloned or snapshotted. No card grants it for
a duration, so there is no stored half.

### Decision 2 — read at the PAYMENT, not in a pricer

A permission belongs to one spell, so the cast pricer is the right home for
its fold. A player grant covers every cost the player pays, and those costs
are priced by five functions (the cast pricer, the CR 602 activation pricer,
the mana-ability pricer, the special-action pricer, the attack-tax pricer)
and by none at all for a pay-unless. So the grant is read where the cost
meets the mana. `costAsPaidByLocked(payer, ctx, cost)` is the one reading,
called once at the top of each payment:

| Payment | Site |
|---|---|
| a cast, from the pool | `applyCastCostLocked` |
| a cast, auto-tapped | `applyAutoTapLocked` |
| a CR 602 activation, a special action, an attack tax | `payAbilityManaCostLocked` |
| a mana ability's own mana cost | `ActivateManaAbility` |
| an attack tax's affordability check | `attackTaxAffordableLocked` |
| a pay-unless, may-pay or ward payment | `payCostLocked` |

The same function answers every affordability question outside the engine:
`internal/legal`'s one probe (`canPayExcluding`, which the cast, activation,
special-action, delve, waterbend and pay-unless moves all ask), its mana-move
and attack-tax checks, and the auto-tap preview (`costAsPaidForPreview`).
The view's castability is the enumerator's digest (`legal_actions`), so it
agrees by construction. Tests pin each pair: the view against the payment,
the enumerator against dispatch, and the preview against `CastSpell`.

Two consequences follow, and both are deliberate:

- **The price shown stays printed.** No pricer is touched, so `cast_prices`,
  an ability row's `charged_mana_cost` and the preview's `cost` read
  `{U}{U}` under an Orrery. That is CR 609.4b's own sentence. The permission
  fold has always rewritten `cast_prices`, and this amendment leaves that as
  it is.
- **The widening comes after convoke, waterbend and delve.** Each payment
  site reads the cost the pricer already reduced by the taps and the delved
  cards. A creature tapped for convoke is not mana, so under an Orrery a
  green creature still cannot convoke a `{U}`. The widening comes BEFORE the
  Phyrexian strike, as the permission fold does, so a widened `{B/P}` keeps
  its "or 2 life" half (#1589).

### Decision 3 — widen, don't fold, and pay the printed colour first

The permission fold moves a coloured symbol into the generic demand. For a
player grant that would make a card worse: the solvers pay generic mana
colourless-first, so a Firespout cast off an Orrery's colourless and a
Mountain would spend the colourless on its `{R/G}` and lose the red mode.
"You MAY spend" never makes a payment worse.

So `widenForAnyColorSpend` marks each coloured requirement
`ColorRequirement.AnyMana`, the bit #1589 introduced for Phyrexian slots. It
keeps `Options` and leaves `{C}` and snow alone. Both solvers changed in the
same way, and they had to change together:

- **The pool solver** (`ManaPool.attemptSpend`, mirrored by `MissingFor`)
  pays every coloured slot with a token of a colour it PRINTS
  (`firstPrintedMatch`). A widened slot that finds none joins the generic
  demand, which any token pays.
- **The auto-tapper** (`solveColored`) places a widened slot on a source of
  a printed colour, and otherwise defers it to `recruitGeneric`.
  `pickMatchingSlot` now matches printed colours only. Deferral is the last
  branch, so the backtracking tries every printed placement first.
- **The executor's colour picks** (`mostRestrictiveRequirement`) give a
  widened slot the option it prints when a multi-colour source has one.

"Payable?" has the same answer as admitting any mana at once: a widened slot
that is not matched costs one generic, and any mana pays that. Only "with
what?" changes. The Phyrexian slots a permission widens follow the same
rule, which is the one visible change outside this seam: a stolen Dismember
paid out of a pool that holds black mana now spends the black on its
`{B/P}` rather than whatever token came first. The #1589 tests pass
unchanged.

### What it does not cover

- **"Spend WHITE mana as though it were mana of any color. You may spend other
  mana only as though it were colorless mana"** (Celestial Dawn, False Dawn,
  and the blue forms on Quicksilver Elemental and Grell Philosopher). That is
  a restriction on which mana widens, and a second restriction on all other
  mana. Neither card is catalogued.
- **A grant over one permanent's own abilities** (Manascape Refractor,
  Agatha's Soul Cauldron, Nathan Drake). `ManaSpendContext` carries the
  source's characteristics and not its identity, so "this artifact's
  abilities" has nothing to match. None is catalogued.
- **Mycosynth Lattice's second line**, "all cards that aren't on the
  battlefield, spells, and permanents are colorless". Layer 5 could make
  permanents colourless, but the layer pass does not reach a spell on the
  stack or a card in another zone, and half of that line is worse than none.
  It stays the card's one caveat.

Proof cards: Chromatic Orrery and Oath of Nissa (full), Mycosynth Lattice
(its spend line). Tracker [#887](https://github.com/krakenhavoc/cmd_and_ctrl/issues/887).

> **Extended by [ADR 0040](0040-mana-pipeline.md)'s 2026-10-03 amendment**
> (#1600): `costAsPaidByLocked` now takes the announced X and first FOLDS a
> cost's own "spend only <colour> mana" clause (Throne of Eldraine, Crypt Rats)
> into coloured symbols, then widens. So the grant reaches those symbols as it
> reaches a printed `{W}` — CR 609.4b, and the Celestial Dawn rulings — and the
> sites listed in Decision 2 read both in one call.

## Amendment — 2026-10-06 (#1928): a permission's spend grant widens, it no longer folds

`spendAsThoughAny` used to fold a cast permission's "spend mana as though it
were mana of any color / any type" into the generic demand inside
`printedCostLocked`. That ran BEFORE convoke, so a stolen convoke spell's
coloured symbols could be paid by tapping any creature: a stolen Stoke the
Flames ({2}{R}{R}) could be convoked entirely with green creatures. CR 609.4b
changes how MANA may pay a cost, and a creature is not mana (CR 702.51a: it
pays {1} or one mana of its own colour).

**Decision.** `asAnyColorCost` and `asAnyTypeCost` now WIDEN instead of fold,
the way the player static (`widenForAnyColorSpend`) and the Phyrexian slots
(#1589) already did: each affected slot keeps its printed options and gains
`ColorRequirement.AnyMana`, and `{C}` stays unwidened under the any-colour
grant. The widened slots are ordered after the ones kept as printed. The
function stays in the one pricer, so the payment, the auto-tapper, the
preview, the bot enumerator and the view still read one answer, and the
convoke subtraction that follows reads `payerCoversRequirement` by the PRINTED
colours (`matchColor(color, req.Options)`, never `Admits`), so a creature pays
a coloured symbol only if it shares the colour, and otherwise only generic.

**Consequences.**

- `ParsedCost.FoldedColored` is gone. Delve (ADR 0100 §1) still cannot pay a
  coloured symbol, because the symbol is no longer in `Generic`; its budget is
  `Generic` plus X, no subtraction. The no-generic mana tag (#2170) reads the
  widened slots through the `deferred` count it already had.
- The price shown is the printed one. `cast_prices` and the auto-tap preview
  render `{2}{C}` rather than `{3}`, which is what CR 609.4b says the cost is.
  `Printed` on the badge now holds for a granted cast from the command zone
  too.
- The same convoke gap closes for Phyrexian slots, the narrower form #1589
  recorded.

Out of scope, unchanged: an any-colour grant on a spell with waterbend (the
extra cost is generic and the grant does not touch it).

## Amendment — 2026-10-07 (#2528): retrace, an additional cost on a priced graveyard offer, and an emblem that grants permissions

Retrace (CR 702.81a) is "you may cast this card from your graveyard by
discarding a land card in addition to paying its other costs". Printed on
ten cards and granted by Six ("During your turn, nonland permanent cards in
your graveyard have retrace") and by Wrenn and Six's emblem ("Instant and
sorcery cards in your graveyard have retrace"). The model already had every
piece but one: a graveyard cast zone (`Spec.CastableZones`), a standing
permission over a filtered graveyard (Underworld Breach's escape), a timing
rule on a permission (`TimingYourTurnOnly`, Tinybones) and a discard that is
paid as a cost (`DiscardCauseCost`). The missing piece was a cost component
that discards a card from hand and is owed only on the graveyard cast.

### Decision 1: the discard rides a priced offer, not `Spec.AdditionalCost`

`Spec.AdditionalCost` is unconditional per spell: it is paid on every cast
of the card, hand or graveyard. Retrace's discard is owed only when the cast
is a retrace cast, and a printed `AdditionalCost` that depended on the cast
zone would need the zone threaded through the additional-cost validator, the
view's cost stamp, the enumerator and the client's prompt chain. The
alternative-cost path already carries a zone binding (`FromZone`), a claim on
the wire (`alternative_cost`), a card component paid from `alt_cost_ids`, a
view stamp (`pay_options`), an enumerator walk (`AltCostCandidatesLocked`)
and a client picker. So retrace is an `AlternativeCost` with a new card
component, `DiscardFromHand`, whose ManaCost is the card's PRINTED cost and
never empty, which is what keeps it from being read as a replacement of the
mana. The discard goes through `discardCardsLocked` with `DiscardCauseCost`,
the same path an additional discard cost uses: it may not pause (CR 601.2h),
it emits `EventDiscardCard` so a discard payoff sees the land, and the land
is gone even if the spell is countered. A card's own `AdditionalCost` is
untouched and still applies.

**What the model gives up, stated.** Retrace is not an alternative cost (CR
702.81a), and here it is claimed as one, so a cast cannot combine it with
another alternative cost (overload, flashback, escape). No catalog card can:
the ten printed retrace cards print no other alternative cost, and a card
that prints a graveyard cast of its own keeps its own price under a granted
retrace, by rule 4 of `validateCastPathLocked` as for every grant. A card
with an ADDITIONAL cost of its own (an optional kicker, a discard) pays it as
well; that composition is the point of using the offer's price rather than
replacing it.

`effects.Retrace(printedCost)` bundles the zone binding, the key, the land
discard and the printed price. It takes the cost as a string, like
`Flashback`, because the offer is catalog data built before any card exists;
`TestPrintedRetraceCardsDeclareBothHalves` pins every card's string and the
fixture check in the PR pins it against the Scryfall dump. Unlike flashback it
carries no `ExileOnLeavingStack`: a retraced spell resolves into the
graveyard and can be retraced again.

### Decision 2: a granted retrace is a permission field, `DiscardLandCard`

`CastPermission.AlternativeCostFor` synthesises the offer a permission prices
the cast under. It gains one boolean, `DiscardLandCard`, that fills
`DiscardFromHand` with the same `game.RetraceDiscardSpec()` the catalog
constructor uses, so a printed and a granted retrace pay one price by
construction. The key is the shared `game.AltCostKeyRetrace`, the way
"escape" is, so a rule that reads "retraced" reads one string however the
cast arrived. `PermissionFilter` gains `NonLandPermanentOnly` for Six; the
instants-and-sorceries filter Past in Flames wrote already existed.

Six is a plain `Spec.CastPermissions` entry with `Timing: TimingYourTurnOnly`
(CR "during your turn", the card's own timing still in force, so a flash
creature is an instant-speed cast on Six's controller's turn and nothing on
anyone else's). It is derived from the battlefield on every query, so it
lasts as long as Six does, covers a card milled by the very attack that
triggers it, and two Sixes compose.

### Decision 3: an emblem may declare standing permissions

Wrenn and Six's emblem is a rule over a player and a zone, not an effect
that resolves, so it is the `ScopeStanding` shape, and an emblem's abilities
function in the command zone (CR 114.3). `EmblemSpec.CastPermissions` is the
`Spec.CastPermissions` slot one zone over, normalised by the same
`standingCastPermissions`, and `standingCastPermissionsLocked` walks the
seat's own `Player.Emblems` after the battlefield, exactly as the
activation-timing read already has a second address there (#1275). An emblem
leaves only with its owner (CR 800.4a), so its presence is the duration and
nothing is stored. Ungated only: no printed emblem gates a permission on a
designation.

### What does not change

The wire. `alternative_costs[].pay_options` already carries a card-shaped
payment, and the client's `AltCostPaymentModal` already renders it from
`pay_label`; the new component sets `pay_label` to "a land card". The
snapshot gains two additive bool fields on `CastPermission` and one on
`PermissionFilter`; both zero-value to "not retrace", so a restore point
written before them restores unchanged.

### Out of scope, stated

- Deeproot Historian ("Merfolk and Druid cards in your graveyard have
  retrace") needs a two-subtype filter; `PermissionFilter.CreatureType` names
  one. Left to a follow-up.
- Cenn's Enlistment, Call the Skybreaker, Worm Harvest and Formless Genesis
  print retrace and need tokens the token table has no printed-token row for
  (Kithkin Soldier, a 5/5 flying Elemental, Worm, a variable Shapeshifter);
  Glamerdye needs text-change effects and Reality Scramble a type-matching
  reveal. Not catalogued here.
## Amendment — 2026-10-07 (#2530): "cards exiled with this permanent", when a replacement did the exiling

Valgavoth, Terror Eater: "If a card you didn't control would be put into an
opponent's graveyard from anywhere, exile it instead. During your turn, you
may play cards exiled with Valgavoth. If you cast a spell this way, pay life
equal to its mana value rather than pay its mana cost." The #1117 triage held
the second sentence back because a replacement effect runs before the card
moves, so it has no post-move moment at which to register a stored permission
over the card that lands, and because `PermissionFilter` had no "exiled by
this permanent" member.

This is a dated amendment, not a new ADR: it adds one filter flag and one
card-carried link to the model decision 1 already defines, and writes no
second grant type.

### Decision 1 — a STANDING permission, not a stored one

The permission is `Spec.CastPermissions` — derived off the battlefield on every
query, never stored — the shape Tinybones, Bauble Burglar (#2179) and Bolas's
Citadel already use. That is what makes the stored-permission problem go away:
nothing is registered when the card is exiled, so nothing has to be registered
over a card that has not landed yet. It also reaches cards exiled before the
permission was asked for, and ends the moment Valgavoth leaves, because a
permanent that is gone grants nothing (decision 1's "for as long as the source
remains", free).

### Decision 2 — `Card.ExiledWith`, a link stamped after the move

CR 607.2a's "exiled with [this permanent]" for an exile a REPLACEMENT made is a
card-carried link, `Card.ExiledWith PermissionCardRef` — the permanent object
`{instance, epoch}`, the shape `Card.HiddenBy` (ADR 0091) already has. It is not
the effects package's event-log record (`b27ExiledWith`), which opens a window
on an `EventResolve` for an ABILITY; a replacement resolves nothing, so there is
no event to key on, and a pure-data filter has a `Card` in hand and no log.

A replacement still cannot write to the card, so it leaves the link on the
event: `ReplacementEvent.ExiledWith`, set by the same `Replace` that rewrites
`NewZone` (the `ShuffleDestinationLibrary` pattern, ADR 0013 §5ah). The two
functions that perform every replaced exit's physical landing —
`executeZoneRouteLocked` and `executeBattlefieldLeaveLocked` — call
`stampExiledWithLocked` once the card has landed, and it writes the link only
if the destination zone really is exile. A move a later replacement redirected
to a command zone, hand or library stamps nothing. `MoveCard` clears the link on
every move, so a card that leaves exile and comes back by another route was
exiled with nothing.

The shared `effects.GraveyardBecomesExile` gets `LinkExiled`, off by default:
Rest in Peace and the Leylines exile without a link, because nothing refers to
the cards they exile.

### Decision 3 — `PermissionFilter.ExiledWithSource`

A catalog entry is static and cannot name an object, so the filter carries a
flag and the derivation fills in the object. `ExiledWithSource` says "the card
must be linked to the permanent granting this"; `stampStandingPermissionLocked`
writes the granting permanent's `{InstanceID, ObjectEpoch}` into
`PermissionFilter.ExiledWith`, the way it already fills `CreatureType` for
`FromChosenType`; `Matches` compares the card's link to it. A flag with no
derived object matches nothing. The epoch is the CR 400.7 check: a Valgavoth
that is bounced and replayed keeps its `InstanceID` and is a new object, so
the cards its earlier self exiled are not exiled with it. Every consumer of
`Matches` (the cast validator, the enumerator, the view) gets the narrowing
with no change.

### Decision 4 — the price is Citadel's, unchanged

"Pay life equal to its mana value rather than pay its mana cost" is
`LifeEqualToManaValue` with an `AltCostKey` (`valgavoth_terror_eater`): a CR 118.9
alternative cost with a CR 119.4 life component, refused when the player cannot
pay, never offered on a land (a land is played, with no cost). The printed mana
cost is not on offer for a card exiled this way: a zone that has a bound offer
can only be cast from by claiming it (decision 3 above). "During your turn" is
`TimingYourTurnOnly`, which leaves the card's own timing in force.

### Snapshot

One additive key, `exiledWith`, on every card (`Card.ExiledWith`), and
`exiledWith` / `exiledWithSource` under every `PermissionFilter`. A file written
before it decodes as "exiled with nothing", which is every game before it; a
binary before it that drops the key leaves an eaten card unlinked, which errs
toward weaker, never stronger. `snapshot_shape/v7.txt` is regenerated; no schema
bump.

### Out of scope, stated

- Cards exiled by an ABILITY's resolution and then played (Court of Locthwain,
  Currency Converter) keep `b27ExiledWith` and a stored grant. Moving them onto
  this link would be a second change; the two coexist because they answer
  different questions (an ability's window versus a replacement's event).
- No wire change: `Card.ExiledWith` is engine state, like `HiddenBy`.

## Amendment — 2026-10-08 (#2550): a grant filter over a union of creature types

Deeproot Historian ("Merfolk and Druid cards in your graveyard have retrace")
is Six's grant with a different filter. `PermissionFilter.CreatureType` names
one subtype, and "Merfolk and Druid cards" is a UNION, not a card that is both:
a Merfolk Wizard and an Elf Druid both qualify.

### Decision: `CreatureTypesAny [2]string`, read through the existing vocabulary

A card qualifies when it has at least one of the named creature types. The
check is `cardHasCreatureType` applied to each non-empty slot, the same
`CreatureTypesOf` read `CreatureType` already uses, so a changeling in the
graveyard qualifies. It composes with `CreatureType` as a further AND, and the
zero array constrains nothing, so a filter written before it matches exactly
what it did.

It is a two-slot ARRAY rather than a slice because `PermissionFilter` is
compared with `==` throughout the engine and its tests (`NextSpell`, the
statics, permission de-duplication); a slice field makes the struct
incomparable and fails the build in several places. Two is the widest printed
union this seam has met; a card that needs a third widens the array.

### Snapshot

One additive key, `creatureTypesAny`, under every `PermissionFilter`. A file
written before it decodes as the zero array, which is "no union constraint". A
binary before it that drops the key could only matter for a STORED permission
carrying it, and none exists: Deeproot Historian's is a catalog standing
permission, derived on every query and never stored. `snapshot_shape/v7.txt` is
regenerated; no schema bump.

### Cards

Deeproot Historian ships Full on it. The four retrace token makers (Cenn's
Enlistment, Call the Skybreaker, Worm Harvest, Formless Genesis) need no engine
change, only token-table rows and, for Formless Genesis, a hand-built X/X
template (its size is the land cards in the graveyard at resolution, which
includes the land discarded to pay for retrace). Cackling Counterpart's caveat,
that a token copy skipped a copied card's as-enters clause, was stale since
#762 and is cleared; `TestATokenCopyRunsTheCopiedCardsAsEnters` pins the
behaviour.

### Out of scope, stated

- Glamerdye (change the text of a spell or permanent: a colour-word rewrite) and
  Reality Scramble (reveal until a card shares a card type with the put-back
  permanent) print retrace but each needs a card-specific effect with no
  primitive yet. Left on their own.
- Jund 'Em Out is not Commander-legal and is skipped.

---

## Amendments

The dated amendments above each have a section of their own. From 2026-10-08 new ones go under this heading.

### 2026-10-08 (#2539): a permission that lasts "until you exile another card with this" (Accepted)

**Status:** Accepted (owner answers 2026-10-08). The owner chose the recommended option on every question; see "Owner answers (2026-10-08)" at the end of this amendment.
**Registry row:** `play-until-you-exile-another` (`server/internal/roadmap/registry.go:2143`).
**Line numbers** below are on `develop` at `62fa97c64`.

#### The cards

A search of the Scryfall default-cards dump for "until you exile another", "exiles another card with" and similar wordings finds exactly three cards. All three are Commander-legal and none is catalogued:

| Card | The exile | The permission |
|---|---|---|
| Unstable Amulet (artifact) | "{T}, Pay {E}{E}: Exile the top card of your library." | "You may play it until you exile another card with this artifact." |
| Furious Rise (enchantment) | "At the beginning of your end step, if you control a creature with power 4 or greater, exile the top card of your library." | "You may play that card until you exile another card with this enchantment." |
| Superior Foes of Spider-Man (creature) | "Whenever you cast a spell with mana value 4 or greater, you may exile the top card of your library." | "If you do, you may play that card until you exile another card with this creature." |

The rulings settle what the clause means:

- **Unstable Amulet (2024-06-07):** "If Unstable Amulet leaves the battlefield before you play the most recently exiled card, you can play that card for as long as it remains exiled." Furious Rise has the same ruling (2020-06-23).
- **Unstable Amulet (2024-06-07):** "You pay all costs and follow all normal timing rules for cards played with Unstable Amulet's last ability. For example, if the exiled card is a land card, you may play it only during your main phase while the stack is empty."
- **Furious Rise (2020-06-23):** "If you exile a land card, you can't play it unless you have land plays available." And: "Playing an exiled card causes it to leave exile. You can't play it multiple times."

#### What the rules say

I checked every rule below against the Comprehensive Rules effective September 25, 2026.

- **CR 611.2a.** "A continuous effect generated by the resolution of a spell or ability lasts as long as stated by the spell or ability creating it (such as 'until end of turn'). If no duration is stated, it lasts until the end of the game." The duration here is stated, and it ends on an EVENT: you exile another card with this permanent.
- **CR 607.2a.** "If an object has an activated or triggered ability printed on it that instructs a player to exile one or more cards and an ability printed on it that refers either to 'the exiled cards' or to cards 'exiled with [this object],' these abilities are linked." Each of the three cards has one exiling ability, and the duration is part of that same ability, so "another card with this" means another card exiled by this ability of this object (CR 607.1c: an ability can be linked to itself).
- **CR 400.7.** "An object that moves from one zone to another becomes a new object with no memory of, or relation to, its previous existence." This rule applies twice:
  - **To the permanent.** An Amulet that leaves and comes back is a new object. What the new object exiles is not "another card with this artifact" for a card the old object exiled. The earlier card stays playable for as long as it remains exiled, which is the ruling above.
  - **To the card.** A card that leaves exile is a new object. ADR 0066 decision 2's object identity already ends the permission.
- **CR 116.2a and CR 305.1.** Playing a land is a special action, takes the turn's land play, and needs a main phase with an empty stack. "Play", not "cast", so a land exiled this way may be played.
- **CR 603.3a.** A triggered ability is controlled by the player who controlled its source when it triggered. That player is the "you" in Furious Rise's and Superior Foes' permission. For the Amulet, the player who activated it is the "you".

#### What the engine has today

- **`CastPermission`** (`game/cast_permission.go:284`) ends on a `Duration` (`:465`). The kinds a permission carries are `UntilEndOfTurn`, `WhileInZone`, `UntilYourNextTurn` and `Indefinite` (the table at `:440`).
- **`ForAsLongAs` is excluded on purpose** (`:451`). Its condition watches a battlefield object, and a standing permission gets "for as long as the source remains" for free, because it is derived from the battlefield on every query (`standingCastPermissionsLocked`).
- **One expiry function.** `durationExpiredLocked` (`game/duration.go:788`) answers "is this over" for every scoped effect and every permission. `CastPermissionActiveForEffect` (`game/cast_permission.go:639`) asks it on every query, and `sweepCastPermissionsLocked` (`:1505`) asks it at cleanup and at each turn's start, as hygiene.
- **Restore is fail-closed for durations.** `Duration.Problem` (`game/duration.go:277`) refuses an unknown kind, and `snapshot_durations.go` asks it of every stored duration, including a permission's. An older binary refuses such a file with `ErrUnknownEffectKey` and keeps it. Other fields on `CastPermission` are not checked: an older binary drops an unknown one without a word.
- **The "exiled with" link.**
  - `Card.ExiledWith` (`game/card.go:940`) and `PermissionFilter.ExiledWithSource` (`game/cast_permission.go:236`) came with #2530 (Valgavoth).
  - `stampExiledWithLocked` (`game/exiled_with.go:30`) writes the link, but only for a REPLACEMENT's exile. Its two callers are `executeBattlefieldLeaveLocked` (`game/mutations.go:5412`) and `game/zone_route.go:789`, both fed by `ReplacementEvent.ExiledWith`.
  - An exile made by an ability's resolution does not stamp it. The effects package keeps a separate record for those, `b27ExiledWith` (`cards/effects/batch27_helpers.go:204`), which walks the event log for a move into exile while a labelled ability of the source was resolving. 18 card files read it.
  - `MoveCard` clears `ExiledWith` on every move (`game/zone.go:406`).
- **Impulse exile.** `ExileTopWithPermissionThenForEffect` (`game/exile_play.go:93`) exiles the top N, hands its continuation the cards that actually LANDED (a commander may go to the command zone instead, #1587), and grants the permission on those cards with `grantImpulseExilePermissionLocked` (`:149`). The permission is `ScopeCards`, naming each card object `{instance, epoch}`.
- **Provenance.** `CastPermission.Source` (`:616`) is an instance ID only. It does not name the source OBJECT.

#### Three designs

##### (a) A new `DurationKind` with a source ref, ended by the next exile

`UntilSourceExilesAnother` joins `DurationKind`. A permission with that kind carries:

- **`Duration.Source` and a new `Duration.SourceEpoch int`.** Together they name the source OBJECT as `{instance, Card.ObjectEpoch}`, the shape of `StackItem.SourceObject` (`game/stack.go:203`, #1418). So CR 400.7 is checked rather than assumed. `SourceEpoch` is omitted when zero.
- **`Duration.Player`,** the "you" (the permission's holder).
- **`Duration.Ended bool`,** omitted when false.

`Duration.Problem` refuses `SourceEpoch` and `Ended` on every other kind, the way it refuses `CounterKind` and the combat stamp.

`durationExpiredLocked` gains one case: the duration is over when `Ended` is set. Nothing else ends it, so a source that has left the battlefield leaves the permission open for as long as the card stays in exile. That is the ruling, and it comes from the `ScopeCards` object check (ADR 0066 decision 2) with no new code.

What sets `Ended` is a game-package helper the three cards share, for example `ExileTopUntilAnotherForEffect(item, n)`. Its continuation runs once the exile has landed:

1. **Nothing landed** (an empty library, or a commander sent to the command zone instead): change nothing. No card was exiled, so the earlier permission stays.
2. **Close the earlier windows.** For each stored permission held by the item's controller whose duration is `UntilSourceExilesAnother`, with the same source object as the item's source and not yet ended, set `Ended`.
3. **Grant the new window.** Grant the new `ScopeCards` permission over the landed card with a fresh `UntilSourceExilesAnother` duration naming the same source object.

**The source object** is the item's `SourceObject`: the object the ability came from, as it was when the ability was put on the stack. Every ability item carries it (`stampTriggerSource`, `game/triggers.go:960`, for a trigger). That is what CR 607.2a needs. An Amulet sacrificed in response to its own activation still exiles a card "with this artifact" when the ability resolves, and that exile ends the earlier window. A NEW Amulet object never matches the old one's permissions.

##### (b) An exile-time sweep that revokes earlier permissions from the same source

The same helper and the same continuation, but no new kind:

- **The data.** The permission carries a new `CastPermission` field, `EndsWhenSourceExiles PermissionCardRef` (the source object), and an `Indefinite` duration.
- **The sweep.** Step 2 deletes the matching permissions instead of marking them.

It is the smallest change. Two costs:

- **A second home for expiry.** Whether a permission is over stops being one function's answer. ADR 0063 and #945 made `durationExpiredLocked` the only one, so that a permission and a continuous effect with the same clause end at the same moment by construction.
- **A rollback stronger than printed.** `CastPermission`'s fields are not checked on restore. An older binary drops `EndsWhenSourceExiles`, reads `Indefinite`, and lets the newest card stay playable after the source exiles another (if the older binary has the card at all). The #2530 amendment's snapshot note holds its new keys to the opposite rule: a rollback "errs toward weaker, never stronger".

##### (c) A standing `ExiledWithSource` permission with a latest-only ordinal

The card declares a `Spec.CastPermissions` entry over exile with `Filter.ExiledWithSource` and a new `LatestOnly` flag. The exile stamps `Card.ExiledWith` plus an ordinal, and the filter matches only the card with the source's highest ordinal. Three problems:

- **Weaker than printed.** A standing permission is derived from the battlefield, so it ends the moment the Amulet leaves. The ruling says the newest card stays playable "for as long as it remains exiled". This is the deciding objection.
- **It needs stamping first.** It depends on ability-resolution exiles stamping `ExiledWith`, which they do not do today (see below).
- **"Latest" must be monotone.** Once the newest card is played and leaves exile, the one before it must not become playable again. So the ordinal has to count cards that are no longer there, which is a counter on the source object that `MoveCard` must not reset.

#### Recommendation: (a)

(a) is the only design that is exact in every case below and fails closed on a rollback.

- **One expiry function.** `durationExpiredLocked` still answers every "is this over", and the duration vocabulary stays the one place a window is spelled (ADR 0063).
- **The ruling for free.** A source that leaves never sets `Ended` again, and the `ScopeCards` object check ends the permission when the card leaves exile.
- **CR 607.2a by construction.** Only the linked ability's own helper sets `Ended`. No other exile, whether by Bojuka Bog, a Rest in Peace replacement or another ability of the same permanent, can close the window.
- **"You".** Step 2 matches the permission's holder as well as the source object. If an opponent gains control of the Amulet and activates it, they exile a card with it, but "you" did not, so your window stays open. This reading follows the text literally. No ruling addresses it; see owner question 4.
- **Fails closed.** An older binary refuses the new kind through `Duration.Problem` and keeps the file (ADR 0041 P4).

#### Lands

Nothing new. The permission is not `CastOnly`, so it opens playing a land (CR 305.1). That is the same special action every impulse "play" grant opens: it needs an available land play and a main phase with an empty stack (CR 116.2a). Playing the land moves it out of exile, which ends the permission by object identity. The rulings that normal timing applies are what the engine does for every granted play.

#### Snapshot

- **Additive within schema v7.** A new `DurationKind` value and two new `Duration` fields (`SourceEpoch` and `Ended`, both `omitempty`) are a growth of the on-disk vocabulary, which ADR 0041 phase 3 allows within a version. `testdata/snapshot_shape/v7.txt` gains the two keys through `-update-shape`. Nothing is renamed or retyped, so there is no bump.
- **Rollback.** A binary from before this change finds an unknown kind, and unknown keys, on a stored permission's duration. `Duration.Problem` names it, `unknownDurations` returns it, and the restore is refused with `ErrUnknownEffectKey`. The file is kept for the roll-forward, the same outcome as an unknown effect key.
- **The corpus.** A new v7 board with an Amulet, an open window and an ended window goes into `TestWriteSnapshotCorpus` beside the others. The writer never touches an existing fixture.
- **No closure.** The helper's continuation is the existing impulse continuation (#1587), and the new duration is pure data. `closure_fields.txt` does not change.

#### Should ability-resolution exiles start stamping `ExiledWith` generally?

Not in this change, and (a) does not need it.

- **The link is not enough for CR 607.2a.** `Card.ExiledWith` names the source OBJECT and not the ABILITY. Some cards have two abilities that exile, and only one of them is linked to the reader. `b27ExiledWith` keys on the ability's label for exactly that reason. A general stamp would need an ability ref beside the object ref, `{instance, epoch, ability}`, with the same identity ADR 0041 P9 gives a catalog row.
- **The other readers.** Moving the 18 `b27ExiledWith` readers onto a card-carried link would make "exiled with this" data rather than an event-log walk. That is worth doing, but it is a refactor across 18 card files and its own snapshot field. It should be its own issue.
- **The helper may stamp it anyway.** The new helper can stamp `ExiledWith` on the card it exiles, at no cost, so that a later reader could find it. That is owner question 3, recommended no, because no card reads it.

#### Bots and the wire

- **No wire change.** The open window is an ordinary `exile_play` surface. An ended window drops out of the cast surface the way an expired one does today.
- **No bot change.** `DurationKind.String` gains "until you exile another card with it" for the log. The bot sees a permission like any impulse grant and plays from it under the existing rules. It cannot see that activating the Amulet again would close the window on an unplayed card, so it may waste one. That is the same blind spot it has with an impulse window that ends this turn.

#### Delivery

One PR holds three parts:

- **The engine:** the kind, the field, the expiry case, `Problem`, the helper, the shape record and the corpus board.
- **Unstable Amulet,** which needs ADR 0129's energy, already built. Its second ability needs a cast-from-anywhere-but-hand trigger, and the batch that builds it checks one exists.
- **The registry row** flips to implemented, with a `docs/engine-seams/closed/2539-play-until-you-exile-another.md` fragment.

Furious Rise and Superior Foes of Spider-Man then need no engine work. They are recipe cards on the same helper, so they can ride the PR or a card batch.

The PR touches stored permissions and the restore check, so it runs the branch E2E and the real-dump audits.

#### Owner questions

1. **Which design?**
   - (a) Recommended: a new `DurationKind`, `UntilSourceExilesAnother`. Its window is marked ended by the linked exile, so expiry stays in `durationExpiredLocked` and an older binary refuses the file.
   - (b) An exile-time sweep that deletes the earlier permissions, with an `Indefinite` duration and a new `CastPermission` field. It is smaller, but a rollback reads the open window as unbounded.
   - (c) A standing `ExiledWithSource` permission with a latest-only ordinal. Not recommended: it ends when the source leaves, which the rulings say it must not do.
2. **A source that leaves the battlefield.** The rulings for Unstable Amulet (2024-06-07) and Furious Rise (2020-06-23) say the most recently exiled card stays playable for as long as it remains exiled. Under (a) that holds with no extra code, and an ability already on the stack when the source left still closes the old window when it resolves (CR 607.2a, source as it last existed). Recommended: accept both behaviours as stated.
3. **Stamp `Card.ExiledWith` in the new helper too?**
   - (a) Recommended: no. Nothing reads it for these cards, and a general move of ability-resolution exiles onto the link (with an ability ref for CR 607.2a) is filed as its own issue covering the 18 `b27ExiledWith` readers.
   - (b) Yes: stamp it now, so a later card can read it.
4. **"You" when control of the source changes.**
   - (a) Recommended: the window closes only when the permission's holder exiles another card with the same source object, which is the literal text.
   - (b) Any exile by that object's linked ability closes it, whoever controls the source.
5. **Delivery.**
   - (a) Recommended: one PR with the engine change and Unstable Amulet, plus Furious Rise and Superior Foes of Spider-Man if their other clauses check out.
   - (b) The engine with Unstable Amulet only, with the other two in a later card batch.

#### Owner answers (2026-10-08)

The owner chose option (a), the recommended one, on every question.

1. **Design (a).** A new `DurationKind`, `UntilSourceExilesAnother`, carrying `Source`, `SourceEpoch` and `Ended`. The linked exile marks the earlier window ended, and `durationExpiredLocked` reads that.
2. **Both behaviours are accepted when the source leaves.** The most recently exiled card stays playable for as long as it remains exiled. An ability already on the stack still closes the old window when it resolves.
3. **Do not stamp `Card.ExiledWith` in the new helper.** The general move of ability-resolution exiles onto the link, with an ability ref, is [#2651](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2651).
4. **Only the holder closes the window.** It closes only when the permission's holder exiles another card with the same source object.
5. **One PR with all three cards.** It holds the engine change plus Unstable Amulet, Furious Rise and Superior Foes of Spider-Man.


## Amendment — 2026-10-09 (#2797): the stored half of an activation-timing statement ("until end of turn")

The ACTIVATION twin of the cast-timing statement (#1208, above) is derived: a permanent or an emblem makes it
for as long as it is there, so the source's presence is the duration and nothing is stored. A resolving
INSTANT has no presence to derive from, and Jace's Machinations says "Until end of turn, you may activate
loyalty abilities of Jace planeswalkers you control on any player's turn any time you could cast an instant".

That statement is stored, the way a granted cast-timing statement is (`GrantCastTimingForEffect`): a
`PlayerStatic` with a CR 611.2 `Duration`, swept and read through `durationExpiredLocked`. It is the eleventh
payload on that slice, `PlayerStatic.ActivationTiming` (`game.ActivationTimingGrant{Timing, Subtype}`,
`game/activation_timing_grant.go`), told apart from the others by `Timing` being non-empty. It is PLAIN DATA,
not the derived statement's `Covers` closure, because it is snapshotted and a closure could not be brought
back. What the printed card narrows on is two facts and both are data: the ability is a LOYALTY ability, and
its permanent is a planeswalker of the named subtype that the holder controls. `Subtype` empty is every
planeswalker (the Teferi emblem's width).

The reader is the same `activationTimingVerdictLocked`, which folds the holder's stored statements first and
then the battlefield's and the emblems' derived ones, so the engine, the bot's list and the wire's
`timing_closed` all agree (`TestEnumeratorAndViewHonourAStoredInstantSpeedStatement`). The three narrowings the
Teferi helpers document hold here too, and CR 606.3's other half is untouched: it opens the WINDOW, and the
once-per-permanent tally is read after it. The card side is `effects.GrantLoyaltyAbilitiesAtInstantSpeed{Subtype}`
(`activation_timing_grant.go`).

Pinned by `TestMachinationsOpensJaceLoyaltyAbilitiesAtInstantSpeed` (a Jace in the declare attackers step, a
Chandra not, a second ability the same turn refused), `TestMachinationsStatementCoversMyJacesOnAnOpponentsTurnAndEndsWithTheTurn`
and `TestMachinationsStatementSurvivesASnapshotRestore`. The snapshot shape is additive
(`seats[].statics[].activationTiming`, recorded in `v7.txt`): a file written before it reads as no statement, and a
binary before it drops the key and loses only a temporary grant, so there is no version bump.


## Amendment — 2026-10-09 (#2167): a permission spent once per card type

Two cards want a permission whose budget is not a count of casts but one use per card TYPE:

- **Muldrotha, the Gravetide**: "During each of your turns, you may play a land and cast a permanent
  spell of each permanent type from your graveyard. (If a card has multiple permanent types, choose one
  as you play it.)" A STANDING graveyard permission, derived off the battlefield, and the card this ADR
  put out of scope in 2026-09.
- **Aminatou's Augury**: "Until end of turn, for each nonland card type, you may cast a spell of that
  type from among the exiled cards without paying its mana cost." A STORED permission over the eight
  cards it exiled, made as it resolves.

`CastsLeft` (#1729) is one flat count, and nothing recorded which type a cast through a permission used.

### Decision 1 — the budget is data on the permission, read by `CoversCard`

`CastPermission` gains `PerType []string`, the card types it opens one use of each (CR 205.2a's names,
lowercase, the spelling `Card.HasCardType` reads), and `PerTypeUsed []string`, the ones spent. Empty
`PerType` is no budget, which is every permission written before the field. `game.PermanentPermissionTypes`
(Muldrotha's six) and `game.NonlandPermissionTypes` (Augury's eight) are the two budgets the cards declare.

`CoversCard` asks the budget as well as the scope (`perTypeOpens`): a permission opens a card only while
some face its holder may cast has a type left to spend. That is the one test the cast path, the bot
enumerator and the view already share, so a spent type closes the card on all three at once. Every face
is asked, because CR 712.11c judges the face being cast: under Augury, a creature with an Adventure is
still open as a sorcery once its creature use is spent.

`PermissionTypeChoices(card)` lists the types a play or cast of `card` could spend. **A land is played,
never cast (CR 305.1), so it can spend only "land"**, whatever else it is: an artifact land under
Muldrotha uses the land play, and under Augury (no "land" in the budget) it opens nothing, which is the
ruling that an exiled land with another type can't be played later. A spell can never spend "land".

### Decision 2 — the type is the caster's, named at announce, judged against the spell

`CastSpellParams.PermissionType` (wire `permission_type`) names the type. It is settled in `CastSpell`
after the face is materialised and after a face-down cast is stamped, because Muldrotha's ruling
(2020-11-10) reads the type "of the card as it's played or cast". One possible type needs no answer;
two or more without one is `ErrPermissionTypeRequired`, because the reminder text makes it the player's
choice and the engine choosing would spend a type they meant to keep. A type the spell lacks, the budget
does not name or has spent, or any type on a cast that spends no budget, is `ErrPermissionTypeNotOffered`;
a card with no type left is `ErrPermissionTypeSpent`.

A cast the card's OWN text allows spends nothing, by `CastsLeft`'s rule (`castUsesGrantLocked`): the
Muldrotha ruling "if multiple effects allow you to play a card from your graveyard, you must announce
which permission you're using".

The type is spent once the play or cast is made: beside `consumeLimitedGrantLocked` for a spell, after
the entry for a land (or when its entry pauses on a prompt, since the play is already made). A land play
through Muldrotha is still a land play: the CR 305.2 drop is checked and counted as for any other.

### Decision 3 — where the spent types live follows where the permission lives

A STORED permission (Augury's) carries `PerTypeUsed` itself, written copy-on-write for the undo reason
`sweepCastPermissionsLocked` gives, and is dropped once its whole budget is spent.

A DERIVED permission (Muldrotha's) is rebuilt on every query and can hold nothing, so its spent types go
to a new turn-tally cell, `TurnTally.PermissionTypes`, keyed (holder, granting object, type) with
`ObjectTallyKey`. `standingCastPermissionsLocked` copies them back onto the stamped permission, which is
what lets `CoversCard` read both kinds the same way. The key carries the rulings:

- the turn tally empties as each turn begins, which is "during each of your turns";
- the granting OBJECT is CR 400.7, so "a new Muldrotha … you may play another land or spell of that type";
- the HOLDER is control, so a Muldrotha that changes hands opens its new controller's own set.

`TimingYourTurnOnly` (#2179) is the "during each of your turns" window for spells and leaves each card's
own timing in force ("you must follow the normal timing permissions").

### Decision 4 — the client asks, the bot is offered one move per type

The view stamps `permission_types` on the cast surface (`CastSurfaceView`, per face, the holder's answer
only — `publicIn` strips it), from `Game.PermissionTypeOptionsLocked`, the same list the enumerator
walks. A face with no type left is not `castable_here`. The list is RANKED
(`RankPermissionTypesLocked`): the type the permission's other cards need least comes first, so an
artifact creature cast beside a plain creature in Muldrotha's graveyard defaults to the artifact.

The client's cast chain asks the type right after the face, when the face carries two or more
(`PermissionTypeModal`), with that first entry selected, and sends it as `permission_type`.

The enumerator offers one move per type when there are two or more, in the ranked order and labelled
"… as an artifact", so the heuristic, which takes the first of equal moves, spends the scarce type.
A land play with "land" spent, and a cast with no type left, are not offered.

### Snapshot

Additive within v7: `seats[].castPermissions[].perType`, `…perTypeUsed` and `turnTally.permissionTypes`,
all omitted when empty. An older binary drops them and reads the permission as unbudgeted, which is
stronger than printed for at most the rest of one turn on a rollback. No version bump, as for every other
`CastPermission` field added inside v7.

### Decision 5 — a permission that charges the printed cost lists the card's other prices

Muldrotha's ruling: "You must pay the costs to cast a spell this way. If it has an alternative cost, you
may cast it for that cost instead", and "you can cast a card with bestow as an enchantment spell"
(ADR 0141). The announce path already accepted an unbound claim (evoke, bestow, overload) out of a zone
a permission opens, since `validateCastPathLocked`'s rule 1 binds only a zone-bound offer, but
`CastOffersForLocked` listed only the zone-bound ones, so neither the client nor the bot could choose
it. It now lists the card's unbound offers too, for a permission that charges the PRINTED cost (no
`AltCostKey`, no `Cost`): Muldrotha, and every impulse exile, where the same rule holds. A permission
with a price of its own ("without paying its mana cost", escape, Bolas's Citadel) lists none, because
CR 118.9a allows one alternative cost per cast — Aminatou's Augury's ruling.

`perTypeOpens` asks a bestow card's Aura spell as well as its printed face, so a bestow creature stays
open as the enchantment once the creature use is spent; `PermissionTypeOptionsLocked` reads a bestow
offer's spell the same way, and the view counts a face as a cast surface if any price leaves it a type.

### Out of scope, stated

- **`budgeted-free-casts`** (#2017), the mana-value sibling, is a different budget and is not built here.

## Amendment — 2026-10-10 (#2719): a cast ban on players who attacked you this turn

Sandswirl Wanderglyph (Unstable Glyphbridge's back face): "Each opponent who attacked you or a planeswalker you control this turn can't cast spells." Its rulings: if the Wanderglyph leaves the battlefield on that turn, the opponent can cast spells; an attack on a battle you protect does not count. The attack half of the card is ADR 0063's amendment of the same date.

### Decision 1 — a battlefield `CastRestriction`, not a granted `CastBanRule`

The #1316 cast-ban twin is for a ban that outlives its source. This one is printed on a permanent and ends when the permanent leaves (CR 604.2), which is exactly what `Spec.CastRestrictions` already gives: `CastGateLocked` walks the battlefield, so the casting verb, the enumerator and the view's `cant_cast` stamp all follow. Writing a `CastBanRule` onto each opponent as they attack would need a sweep when the Wanderglyph leaves. ADR 0109 §5 rejected that shape for an emblem's ban for the same reason. The constructor is `effects.OpponentsWhoAttackedYouCantCast(label)`, beside `OpponentsCantCast`; "you" is the source's controller (`q.Source.Controller`), so a stolen Wanderglyph protects its new controller.

### Decision 2 — the attack record remembers whose planeswalker it was

`Game.AttackedYouOrYourPlaneswalkerThisTurn(player, you)` reads `TurnTally.Attacks` (ADR 0059 Decision 8). Only the active player declares attackers (CR 506.2, 508.1), so `player` must be the active player and every record is theirs. A record whose `Defender` is `you` counts. So does one on a planeswalker `you` controlled, but the question is asked later in the turn, when that planeswalker may have died or changed hands. So `AttackRecord` gains `PlaneswalkerController`, stamped by the tally listener from the attacked permanent at the `EventAttack`, and left nil for a player or a battle. `Player.LastTurnAttacks` copies the record and carries it too, unread.

Rejected: reading the planeswalker's controller at cast time. A planeswalker killed in combat has no controller to read, and that is the common case.

### Snapshot

Additive: `turnTally.attacks[].planeswalkerController` and `seats[].lastTurnAttacks[].planeswalkerController`, recorded in `snapshot_shape/v7.txt`. A tally restored from an older file has no stamp, so an attack declared before the upgrade on a planeswalker does not count for the rest of that turn (weaker than printed). An attack on you is read from `Defender` and is unaffected.

### Tests

`game/cant_attack_player_test.go` (`TestAttackedYouOrYourPlaneswalkerThisTurnReadsTheDeclaration`: another player and a battle you control don't count, a planeswalker that has since left does, a non-active player attacked nobody) and `cards/effects/unstable_glyphbridge_test.go` (the gate after attacking you, your planeswalker, your battle and another player, and the ban lifted when the Wanderglyph is destroyed).

## Amendment — 2026-10-10 (#2559): every player plays the cards they exiled, and "can't play cards from your hand"

**Memory Vessel**: "{T}, Exile this artifact: Each player exiles the top seven cards of their library. Until
your next turn, players may play cards they exiled this way, and they can't play cards from their hand.
Activate only as a sorcery." The issue named two pieces the engine did not have:

1. a play permission over exiled cards held by EVERY player at once, each over their own cards, lasting
   "until your next turn";
2. a player-scoped "can't play cards from your hand" with a duration. The cast bans (#1316, above) are
   per card and per zone EXCEPT one (`CastBanRule.ExceptFromZone`); none bans exactly one zone, and none
   reaches a land play.

Shipping the exile half alone would have let every player keep casting from hand as well, which is
stronger than printed (#259), so the card waited for both.

### Decision 1 — the permission needs no new field: one stored permission per player

The model already had everything the first piece asks for. A stored `ScopeCards` permission names card
OBJECTS and is held by one `Player`; `Duration` already accepts `UntilYourNextTurn` (the field's comment
said "nothing prints it on a cast permission yet; it costs nothing to accept"); and impulse exile already
exiles a library's top N and grants the exiled objects to a holder
(`Game.ExileTopWithPermissionForEffect(from, grantTo, n, perm)`). "Each player … they exiled" is that call
once per seated player with `from == grantTo`:

```go
// effects/impulse_exile.go
func EachPlayerExilesTopAndMayPlay(g *game.Game, n int, d game.Duration) error
```

The window is stamped by the CARD against its own controller (`g.UntilYourNextTurnDuration(controller)`
for the Vessel), so every player's permission ends at the same moment, as the activator's next turn
begins, and CR 800.4m's "until that player's turn would have begun" for a departed activator is the
duration's existing rule. Nothing about the permission is new, so the cast path, the enumerator and the
view read it as they read any impulse grant: a player sees, and may play, only their own; a land among
them is played and spends that player's land drop (§6, CR 305.2); normal timing applies.

Etali, Primal Storm already called the same primitive once per seat with a SINGLE holder; this is the
same loop with the owner as holder. A second card needs exactly this and nothing else:
**Rocco, Street Chef** ("At the beginning of your end step, each player exiles the top card of their
library. Until your next end step, each player may play the card they exiled this way."), stamped with
`UntilYourNextEndStepDuration` (#2373). Its second ability, "whenever a player plays a land from exile or
casts a spell from exile", is Prosper's Pact Boon read for any player (`aCardWasPlayedFromExile`).

### Decision 2 — the hand ban is ONE stored record read by BOTH gates

"Play" is two acts with two gates: a spell is cast (CR 601, `CastGateLocked`, cast_gate.go) and a land
is played as a special action (CR 305.1, CR 116.2a, `LandPlayGateLocked`, ADR 0109 §4). "Can't play
cards from your hand" is one statement about both, so it is stored once and asked by both, never built
as a third gate:

- `CastGateLocked` refuses a cast whose source zone is `ZoneHand` (after the granted cast bans, before
  the spell's own condition);
- `LandPlayGateLocked` refuses a land play whose source zone is `ZoneHand` (after the stored land bans).

Each gate already has every caller it needs — the cast path, `CanPlayLandDuringResolutionForEffect`
(CR 305.2a), `legal`'s cast and land enumeration, and the view's `cant_cast` stamps — so all of them see
the ban without a line changed at any of them. A cast or a land play out of any other zone is untouched,
and so is everything out of a hand that is not a play: cycling and channel (activated abilities),
foretell (a special action, CR 702.143a), ninjutsu (an ability), and "put a land card from your hand onto
the battlefield" (not a land play, CR 305.4).

**Where it is stored.** A `ScopedEffect` of a new rules kind, `cantPlayFromHand` (reader `rule`, scope
game, reads `Player`; zero is every player), registered by `Game.CantPlayFromHandForEffect(source,
player, d, label)` (`game/hand_play_ban.go`). The alternatives were rejected for the reasons ADR 0109
§4 gave `cantPlayLands`:

- a `CastBanRule` kind (`ExceptFromZone`'s inverse) reaches only casts, so the land half would still
  need a second record; and an older binary does not check `CastBanKind` on restore, so a rollback would
  silently ban nothing, which is stronger than printed for the rest of the window;
- an unknown mod kind is refused on restore with `ErrUnknownEffectKey`, the file kept — the rollback
  case the vocabulary is built for.

The one reader, `handPlayBanLocked`, asks `durationExpiredLocked` itself rather than trusting the sweep,
as `castBanForbidsLocked` does: "until your next turn" ends as that turn begins, and the read has to be
right in the moments between sweeps. The refusal is a `*CantCastError` or a `*CantPlayLandError` with the
clause "You can't play cards from your hand — Memory Vessel".

### Decision 3 — the wire says who holds a grant and until when

The client already drew the viewer's OWN exile grants (the exile strip, the zone browser's button).
Under Memory Vessel every opponent's pile holds seven cards its owner may play, and nothing on the board
said so; the button's tooltip also said "playable until end of turn" for every grant.

- `exile_play.until` / `until_player` (`game.PermissionWindowLocked`): `end_of_turn`,
  `end_of_next_turn`, `next_turn`, `next_end_step`, `while_exiled` or `until_another`, and the player the
  window counts against for the three that name one. Public, like the rest of `exile_play`.
- `PlayerView.cant_play_from_hand`: the seat's clause, public, for `cant_play_lands`' reasons. While it
  is set `cant_play_lands` carries the same clause (the ban refuses every land from the hand), and the
  client shows one NO HAND badge for both.

The client reads them in `exileGrants.ts`: the zone browser labels a card another seat may play ("Ana
may play it until Bo's next turn") where the viewer's own gets a button, the button's tooltip names its
real window, and a seat's exile chip (`PileBar`, `SeatSummary`) counts the cards that seat may play
("7 theirs to play"). Each card in a banned hand is already greyed with the clause through its own
`cant_cast`.

### Decision 4 — bots

The enumerator asks both gates, so a bot is never offered a play from a banned hand, and is offered its
own exiled cards through the permission walk it already had. The heuristic needs no change: an exiled
cast already costs no card in hand (#673) and a land from exile is a land drop like any other
(`TestHeuristicSeatPlaysFromExileUnderMemoryVessel`). The model seats' board text (`boardtext`) gains
the two facts the view does: "may play from exile: …" per seat, and "CAN'T PLAY CARDS FROM HAND" on a
banned seat's line.

### Snapshot

Additive within v7. The new kind is one more on-disk identity (ADR 0041 phase 3), fixture
`v7/memory_vessel.json`; the permissions are ordinary stored `CastPermission`s. No version bump.

### Out of scope, stated

- **Uba Mask** ("each player may play lands and cast spells from among cards they exiled with this
  artifact this turn") needs a draw replaced by an exile and a "this turn" mark on the exiled card, not
  only this permission.
- **Knowledge Pool** and **Etali, Primal Conqueror** cast another player's exiled card or a card
  exiled by a different trigger; **Shaman's Trance** bans a graveyard rather than a hand and opens other
  players' graveyards (`play-from-other-players-graveyards`).

## Amendment — 2026-10-13 (#3105): a cast zone a card opens only while a condition holds

Lightwheel Enhancements ("Max speed — You may cast this card from your graveyard"), Gravecrawler ("as long
as you control a Zombie") and Marang River Prowler ("as long as you control a black or green permanent")
print a graveyard permission that depends on the game. `Spec.CastableZones` is unconditional and its hook
(`CatalogCastableZones`) takes an oracle ID with no game, so the zone could not depend on state. No number
is taken: this amends the S29 declaration this ADR already owns.

### Decision 1 — a second declaration, not a wider hook

`Spec.CastableZonesIf []game.ConditionalCastZone{Zone, Cond}`, with `Cond func(g *Game, owner uuid.UUID,
card Card) bool`, reaches the engine through a second catalog hook (`CatalogCastableZonesIf`). The
unconditional hook keeps its signature and every existing card is untouched: nothing in `CastableZones`,
`CastableZonesFor` or `CardCastableFromZone` changed meaning. `CardCastableFromZone` stays the answer to
"does the card open this zone UNCONDITIONALLY"; its three engine-side callers moved to the new predicate.

### Decision 2 — one predicate, asked by everyone

`Game.cardOpensZoneLocked(card, key, zone)` is the unconditional declaration OR a conditional entry whose
condition holds. It is read by `validateCastPathLocked` (the engine), `castUsesGrantLocked` (is a permission
the REASON the cast is legal), `CardCastableFromAnyFaceLocked` (the legal-move enumerator, so a bot, and the
view's `castable_here`, so the client's button) and the delve ordering. That is #1171's lesson applied
before the divergence rather than after it: the condition cannot hold for the engine and fail for the bot.
`CardCastableFromAnyFace` became a method on the game for the same reason (it needs the game to answer); it
is the exported `CardCastableFromAnyFaceLocked`, caller holding `g.mu`, and has two callers.

### Decision 3 — the owner is asked, and the condition is data about the game, not state

A card's own text opens its owner's zone and nobody else's, so `Cond` receives `card.Owner`; "max speed" is
the owner's (ADR 0138 §5). The condition reads the game it is handed, captures nothing and is stored
nowhere: nothing is added to `Game`, `Card` or the snapshot, so Clone, undo and the bot's look-ahead answer
from the copy they hold (`TestConditionalZoneFollowsTheCloneNotTheOriginal`), and the cast button follows a
speed change on the next frame with no invalidation. The closure lives on `CardDef` beside `CastCondition`;
it is not reachable from `Game`, so the closure ratchet is untouched and no ceiling moved.

### Decision 4 — why not `CastCondition`

Oathsworn Vampire and The Indomitable already gate a graveyard cast with `CastCondition` plus a lookup of the
card's own zone. That refuses every cast out of the graveyard while the condition is false, including one a
permission opens: Snapcaster Mage's flashback over a Gravecrawler with no Zombie in play would be refused,
which the printed cards never do (CR 601.2, and rule 4 of `validateCastPathLocked`: a permission must never
take away a path, and must never be blocked by a condition on a path it is not using). The conditional zone
gates only the card's own opening, so a permission still opens the zone
(`TestAPermissionStillOpensAConditionalZoneThatIsClosed`) and a one-spell permission is spent only when it was the
reason (`TestALimitedGrantIsSpentOnlyWhileAConditionalZoneIsClosed`). The two existing cards were left on
`CastCondition`: moving them is a behaviour change to cards this issue does not own, and the stricter
reading is the safe direction.

### Decision 5 — only a graveyard, and a condition is required

`Register` panics on a `CastableZonesIf` zone that is not the graveyard (hand and command are the format's,
exile and the library are per-instance permissions, as for `CastableZones`) and on a nil `Cond`, which would
be an unconditional zone written the long way round. A zone-bound `AlternativeCost` (flashback, escape) may
name a zone listed in either field.

### Snapshot

None. Nothing is stored.

### Out of scope, stated

- **Ebondeath, Dracolich** ("if a creature not named Ebondeath died this turn") and **Eelectrocute** ("as long
  as you've rolled a 6 this turn") need a per-turn record the engine does not keep (deaths by name, die
  results); **Undead Sprinter**, **Glimpse the Cosmos** and **Haakon, Stromgald Scourge** also change the price
  or open other cards. The conditional zone would carry their gates once those records exist.
- **Moving Oathsworn Vampire and The Indomitable** onto `CastableZonesIf` (Decision 4).

## Amendment — 2026-10-13 (#3192): a granted permission that is live only while a condition holds

Neyali ("during any turn you attacked with a token, you may play that card"), Boros Strike-Captain, Goblin
Researcher, Robber of the Rich, Neriv, Pick Up the Pace, Possibility Technician and Summon: Brynhildr print
an exile permission that is neither an until-end-of-turn window nor "for as long as it remains exiled": it
is live on some turns and not others, and the card stays exiled the whole time. Before this a card was
re-granted by a trigger that had to resolve on the right turn, which is weaker than printed (it needs the
source on the battlefield and the trigger to resolve). #3105 solved the same question for the zone a card's
own text opens (`Spec.CastableZonesIf`); this is the per-instance counterpart. No number is taken: it amends
the `CastPermission` model this ADR owns.

### Decision 1 — a condition is a key plus plain parameters, on the permission

`CastPermission.Condition *PermissionCondition` carries a registered key (`game.PermissionConditionKey`,
`RegisterPermissionCondition`) and parameters (`Subtype`, `N`, `ObjectID`/`ObjectEpoch`). A permission is
stored state, mirrored into the snapshot, so a closure is not available (that is why `CastPermissionGate`
exists only for catalog-declared standing permissions); the registry is the `RegisterCastFollowUp` shape, so
the closure-field ratchet does not move. A restore naming a key this binary never registered is refused
(`ErrUnknownEffectKey`), and an unknown key at run time fails closed (dormant), never open.

### Decision 2 — evaluated in the one liveness test

`CastPermissionActiveForEffect` is already the test the cast path, the view's castable flags and `exile_play`
stamp, the legal-move enumerator (and so the bot), the library-top look and the hand branch all read. The
condition is checked there, beside `NotBeforeSeq` and `RequiresNoSpellsCast`, and nowhere else. The
permission is not dropped when the condition is false: the sweep keys on the duration and the named objects,
not on the condition, so the same grant is live again on the next turn the condition holds
(`TestConditionalPermissionFlipsAcrossTurns`).

### Decision 3 — attack history is recorded as the attacker is declared

"Attacked with a token / a commander / a Rogue" is a question about the attacker AS IT WAS DECLARED, asked
later in the turn when a token that traded in combat is gone (CR 704.5d). `PlayerTurnTally.TokensAttacked`
and `CommandersAttacked` and `TurnTally.AttackedSubtypes` (the `EnteredSubtypes` keys and changeling rule)
are written from `EventAttack`; "N or more creatures" and "this specific creature" read the existing
`TurnTally.Attacks` list (distinct objects, so a second combat does not double count; the epoch makes a
blinked creature a new object). "Put a lore counter on this Saga" reads `TurnTally.LoreCounterPut`, written
in `loreCountersPutLocked`, the one choke point for lore counters. All reset at the turn boundary with the
rest of `TurnTally`. "You control a <subtype>" is a pure battlefield read.

### Decision 4 — a spell-type gate is a Filter, not a condition

"If it's a creature spell" (Vivien, Champion of the Wilds) and "if it's an instant or sorcery spell" (Nivix) are
properties of the card, not of the game. `PermissionFilter` already expresses them for a standing permission;
`CoversCard` now applies it to a named-cards permission as well (the zero filter matches everything, so no
existing grant changes).

### Decision 5 — "for as long as you control <source>" needed nothing

`Duration{ForAsLongAs, WhileYouControlSource | WhileSourceOnBattlefield}` is read by the same liveness test and
ends for good when the source leaves or changes hands (CR 611.2b). Lightning, Security Sergeant, Victor Mancha
and Pick Up the Pace use it as is; a source already gone when the trigger resolves gets a window that never
started, so the card is exiled and never playable (`effects.forAsLongAsYouControlWindow`).

### Decision 6 — the wire says the grant is dormant

`exile_play` gains `condition` (the printed clause) and `condition_unmet`. The grant is still public; the
client withholds the play button while the condition is false. Castable flags are per holder and read the
liveness test, so a dormant grant stamps none.

### Snapshot

Additive within v8, omitted when empty: `seats[].castPermissions[].condition`, `turnTally.attackedSubtypes`,
`turnTally.loreCounterPut`, `turnTally.players[*].tokensAttacked` and `commandersAttacked`.

### Out of scope, stated

- **Summon: Brynhildr** needs a "when you next cast a creature spell this turn, it gains haste" delayed cast
  trigger for its chapters II and III as well; the lore-counter condition (`CondLoreCounterOnObject`) is
  ready and pinned in the engine tests.
- **Gandalf, Goblins' Bane, Intet, Gwen Stacy, Taster of Wares, Kotose, Hama, The Day of the Doctor and
  Vivien, Champion of the Wilds** use this plus other pieces (a face-down adventure exile, an optional
  payment, a transforming face, a reveal-and-choose, a graveyard/hand/library search by name, waterbend on a
  granted cast, a legendary-card dig, a loyalty ability's pick-from-three) and were left for later slices.
- **A dormant grant's client caption** (why the card is not playable now) is the printed `condition` text only.

## Amendment (2026-10-13, #3172, cav-g2): a Siege's transformed cast is a window, not a main-phase cast

`SiegeDefeated` now grants the back-face cast with `Timing: TimingFlash` and `LapseOnPass: LapseStaysInExile`, the grant rebound
and cascade use. The cast ignores the card's timing, as the printed "cast it transformed" does, and the window closes the next time
its holder passes priority, leaving the card exiled. A Siege defeated on someone else's turn is no longer lost, and the
end-of-turn duration remains as a backstop. The four Sieges lose `SiegeTransformedCastCaveat`.

Back-out: drop `Timing` and `LapseOnPass` from the grant; `TestSiegeDefeatedOnAnotherPlayersTurnCanStillBeCast` fails.

## Amendment — 2026-10-13: a madness cost can charge life beside its mana (#3212)

Shadowgrange Archfiend prints "Madness—{2}{B}, Pay 8 life" (CR 702.35a: any
cost, and a life payment is a cost, CR 119.4).

- **`Spec.Madness` stays the mana string.** `Spec.MadnessLife` (an int) is the
  life, copied into `CardDef.MadnessLife` and read through
  `game.MadnessLifeFor(oracleID)`, a catalog hook like `XMattersFor`.
- **The grant is unchanged.** `CastPermission` gains no field (so the snapshot
  shape does not change): `AlternativeCostFor` adds `Life` to the synthesised
  offer when `AltCostKey` is `madness`, by asking the catalog, so a restored
  table prices the cast the same way. The offer is an `AlternativeCost.Life`,
  which the existing gates already honour: a player with fewer than N life
  cannot claim it (`lifePayableBy`), and the life is paid with the mana as
  part of the cast.
- The may-cast prompt names it ("madness cost {2}{B} and 8 life").
- **Back-out.** Delete the `AltCostKeyMadness` block in `AlternativeCostFor`:
  `TestShadowgrangeArchfiendMadnessChargesEightLife/enough_life` fails
  (life stays 20, not 12) and `/not_enough_life` succeeds in casting.

## Amendment — 2026-10-13 (#3258): casting a copy of a card

Mizzix's Mastery and Arcane Bombardment print "copy it, and you may cast the copy" over one card or a whole set. Until now the
engine copied a CARD in one place only, the prepare spell (ADR 0090, CR 722.3c), kept alive by a derived permission that reads a
prepared permanent. Every other cast permission opened the card itself. This amendment generalises that copy to one made by an
effect and kept alive by an ordinary stored permission. No new rule, and no new permission field.

### Decision 1 — the copy is a real object marked `Card.CardCopy`

`Game.CopyCardsAndOfferCastsForEffect(CardCopyOffer, cards)` (`game/card_copy.go`) makes a copy of each named card that is in exile
or a graveyard, in the zone the card is in (CR 707.12; a graveyard copy goes to the creating player's graveyard), from the card's
copiable values (CR 707.2). It keeps every one of them, unlike the prepare copy, so its oracle ID is the original's and every
catalog hook resolves it as the original: a copy of a Lightning Bolt is a Lightning Bolt. It carries no counters, status, "exiled
with" link or permission. It is created, not moved, so no zone-move event announces it and nothing "was exiled" (the prepare copy
does the same).

### Decision 2 — the offer is the existing `ScopeCards` permission, one per copy

"Any number of the copies" needs no new shape: the grants are independent, casting one leaves the others, and the order the
player casts them in is the order they go on the stack (the last cast resolves first). The cast is the ordinary cast out of the
zone, so casting triggers fire, targets and modes are chosen, the bot enumerator and the client's cast offers see it, and a `{0}`
price locks X at 0 (CR 107.3b). `CardCopyOffer` carries `Cost` (`{0}` for "without paying its mana cost", empty for "cast the
copy" at the printed cost), `Timing`, `Lapse` and `FollowUp`, all straight onto the permission. The window is the one discover,
cascade and the Siege grants use: `TimingFlash` with `LapseStaysInExile`, the grant model's stand-in for CR 608.2g.

### Decision 3 — the sweep is CR 704.5e, keyed on the offer

`cardCopySweepSBALocked` removes a `CardCopy` in any zone that no stored permission names. Every way a window can close then ends
the copy with no per-site code: the holder passing (`closePassWindowsLocked` also runs the sweep directly, since a pass may reach
no state check before the table looks), the end of the turn (`sweepCastPermissionsLocked`), or the copy leaving its zone. A copy
that is cast is the spell on the stack (`StackItem.IsCopy`, now `card.PrepareCopy || card.CardCopy`) and ceases on resolution; a
permanent spell becomes a token (CR 707.10b), through the path Twincast's copies already use.

### Decision 4 — the effects door is `effects.CopyAndCast`

`CopyAndCast{Cards, Player, Free, FollowUp}` is the card-facing primitive. It always uses `TimingFlash` and `LapseStaysInExile`.
Mizzix's Mastery feeds it the cards its exile really landed (`ExileCardsThenForEffect`'s continuation, so a card a replacement
sent elsewhere is not copied). Arcane Bombardment feeds it every card still exiled with the enchantment (`b27ExiledWith`), so the
copies are made again from the same originals on every trigger.

### Decision 5 — the first-instant-or-sorcery clock

`CastTally.InstantSorcery` counts the instant and sorcery spells a player cast this turn, bumped beside `Artifact` as each spell
becomes cast, so `== 1` on the cast event means "your first".

### Snapshot

Additive within v8, omitted when zero: `cardCopy` on every card path and `spellsCastThisTurn{}.instantSorcery`.

### Out of scope, stated

- **The window closes on the holder's next priority pass** instead of being an inline cast inside the resolution, the posture
  every "you may cast it" a resolution offers takes (ADR 0091, discover.go). A cast on another player's turn can be preceded by
  responses.
- **A total mana value budget** (Uldaros Theorix's "total mana value 6 or less") is the separate `budgeted-free-casts` seam (#2017).
- **Paradigm** (#1866) needs a recurring offer at each first main phase over a card that stays in exile, plus a self-exile and a
  per-name record. **Spellweaver Volute** (#3157) needs its sorcery-cast trigger and the CR 303.4f re-attach (#3158).
  **Kaervek, the Punisher** (#2947) is `CopyAndCast{Free: false, FollowUp: …}` plus the crime fix (#2944).
- **A copy in a hand or a library** is not made; no card asks.
- **Cast from a copy made in another seat's graveyard** is not modelled; the copy goes to the creator's.

Back-out: remove `card.CardCopy` from the `IsCopy` line in `CastSpell`, and `TestCardCopyOfAPermanentCardResolvesAsAToken` fails;
remove the sweep call in `runStateChecksLocked`, and `TestCardCopyWithNoOfferCeasesAtTheNextStateCheck` fails.

## Amendment (2026-10-13, [#3172](https://github.com/krakenhavoc/cmd_and_ctrl/issues/3172), cav-h2): "mana of any type", and a flash window bound to a target

**Any type.** `AnyColorSpendStatic.AnyType` is the grant Vizier of the Menagerie prints, "you can spend mana of any type to
cast creature spells": the colour grant plus colourless being a type, so coloured mana also pays a `{C}` symbol.
`spendGrantLocked` reports whether a covering grant is the any-type kind and `widenForAnyColorSpend(cost, anyType)` then
widens the `{C}` requirement too (Oath of Nissa and the other colour grants still leave it alone). `YouMaySpendManaAsAnyTypeToCast`
is the constructor.

**Flash if it targets.** `FlashIfTargets(cost, what, clause)` is [ADR 0025](0025-alternative-costs.md)'s flash-window
offer (`AsThoughFlash`) carrying the narrower target clause the flash cast must satisfy (`AlternativeCost.Targets`). Timely Ward
offers `{2}{W}` with "enchant creature" narrowed to a commander (`ACommander`); the printed cast is sorcery-speed with
any creature.

Back-out: `AnyType` false makes `TestVizierOfTheMenagerieLetsAnyManaPayACreatureSpellsColorlessSymbol` fail; the flash offer
removed returns Timely Ward to its caveat.

## Amendment — 2026-10-13 (#3201, ADR 0155 §8): a resolving ability's exile stamps "exiled with" too

The 2026-10-07 amendment (#2530) stamped `Card.ExiledWith` only for a replacement's exile. ADR 0155 §8 (accepted 2026-10-13)
extends the stamp to an exile made by a resolving ability, because a "has all activated abilities of all cards exiled with it"
borrow reads its lenders on every layer pass, where the event-log walk (`b27ExiledWith`) is far too slow.

- **The verb is opt-in.** `Game.ExileLinkedForEffect(source, card)` and `ExileLinkedThenForEffect(source, ids, then)` are the
  ordinary exile route with the source OBJECT on it (`zoneRoute.ExiledWith`). The route seeds `ReplacementEvent.ExiledWith`, so
  the stamp lands through the same two finishers as Valgavoth's, and only when the card really lands in exile: a commander its
  owner moves to the command zone, or a card a replacement sends to a graveyard, is linked to nothing. The card side is
  `effects.ExileTarget{Linked: true}`, which names the resolving item's source object (`SourceObjectForEffect`), so a source
  that left and came back is a new object with no claim on the card (CR 400.7).
- **Only the linked ability writes it.** The stamp names the object, not the ability (#2651's ability ref is not added). That is
  enough for CR 607.2a as long as only the linked exile uses the verb, which is how the cards declare it. Two linked pairs on one
  object sharing one stamp is the weakness ADR 0155 Decision 6 already declares.
- **Readers.** `Game.ExiledWithForEffect(ref)` lists the linked cards in exile order. Duplicant and Angel of Serenity moved onto
  it from `b27ExiledWith`; the other event-log readers stay as they are (#2651 tracks them).
- **Invalidation.** A `DependsOnExile` static now also bumps the layer version when a card ARRIVES in exile from a hand, a library
  or the stack (a move from the battlefield or a graveyard already bumped).

No snapshot change: `Card.ExiledWith` is already carried.

## Amendment — 2026-10-14 (#2017): a total mana value budget over the copies of a card

Uldaros Theorix prints "You may cast any number of spells with total mana value 6 or less from among the copies without paying
their mana costs." The 2026-10-13 card-copy amendment (#3258) left this as the separate `budgeted-free-casts` seam: the copies
were one independent permission each, so there was nothing to share a total.

### Decision 1 — `CastPermission.ManaValueBudget` is the REMAINING total one permission may still open

A `*int` on a `ScopeCards` permission (nil, which is every earlier permission, means no budget). It is not a cap on one spell
(`MaxSpellManaValue` is, and does not change); it is what is left of the printed total after the casts already made through the
permission. `spellManaValueWithinCap`, the one test `validateCastPathLocked` makes for the cast path, the view's offers and the
bot enumerator, refuses a spell whose mana value (the face being cast, X at zero) is more than what is left. Any set with a total of
at most N can be cast in some order without going over, so the running total and the printed "total mana value N or less" accept
the same casts, whatever order the caster takes them in.

### Decision 2 — one permission over the whole set, spent by `consumePassClosedGrantLocked`

`CardCopyOffer.ManaValueBudget` (`effects.CopyAndCast.TotalManaValueAtMost`) makes ONE permission naming every copy, per zone
the copies sit in, instead of one per copy. A cast through it keeps the permission with the cast card struck from its named
objects and the budget reduced (`CastPermission.spendManaValueBudget`): a fresh `Cards` slice and a fresh budget pointer, never
written in place, because the originals are shared with the undo snapshots. A permission that names only the card just cast is
removed whole, as every pass-closed grant always was. The window is unchanged: it closes when the holder next passes priority, and
a copy not cast ceases to exist (CR 704.5e).

### Decision 3 — what the budget does not do

It is not a per-type budget (`PerType`, Decision 2 of the 2026-10-09 amendment): a card is cast through it once, for its whole mana
value. Primeval Spawn's "from among them" over cards an ordinary impulse exile put in exile is the same budget over cards that are
not copies; `ExileTopWithPermissionForEffect` still grants one permission per card, so it needs a helper that grants one budgeted
permission over the landed cards (the seam stays `partial` for that card).

### Snapshot

Additive within v8, omitted when nil: `seats[].castPermissions[].manaValueBudget`. A binary from before it drops the key on restore
and reads the permission as having no total, as it reads a `CastsLeft` or `PerType` it does not know.

Back-out: remove the budget test in `spellManaValueWithinCap` and `TestManaValueBudgetIsOnePermissionSpentByEachCast` fails; write
the budget in place in `spendManaValueBudget` and `TestManaValueBudgetIsNeverSpentInPlace` fails; make the offer per copy and both of
those fail.

## Amendment — 2026-10-15 (#3783): a card copy cast for a named alternative cost

Blue Mage's Cane (granted to the equipped creature): "exile up to one target instant or sorcery card from defending player's
graveyard. If you do, copy it. You may cast the copy by paying {3} rather than paying its mana cost." That is the card-copy
offer of the 2026-10-13 amendment with a price other than "free" or "the printed cost" (CR 118.9).

`game.CardCopyOffer.Cost` already carried any price: it is the cost paid instead of the printed mana cost, and a cost
with no `{X}` locks an `{X}` at 0 (CR 107.3b), as the free cast does. The card-facing door, `effects.CopyAndCast`, could only reach
`"{0}"` (`Free`) or the printed cost (empty). It gains `Cost string`, written as a mana cost and passed through to the offer;
`Free` wins when both are set, since `Free` is `Cost: "{0}"`. The copy is cast through the same grant, so its other costs and
cost increases still apply, the copy lives in exile and ceases to exist if it is not cast by the controller's next priority pass,
and the original stays exiled (CR 707.12, CR 704.5e). No wire or snapshot change.

Back-out: pass `Cost` through to the offer no longer and `TestBlueMagesCaneExilesTheDefendersSpellAndOffersACopyForThree` fails
on the permission's cost and on the cast for `{3}`.

## Amendment — 2026-10-15 (#3769): a card remembers how and when it reached a graveyard

Banon, the Returners' Leader: "Once during each of your turns, you may cast a creature spell from among cards in your graveyard that
were put there from anywhere other than the battlefield this turn." The once-per-turn budget is Gisa and Geralf's shape
(`PerType` holding `creature`, `TimingYourTurnOnly`). What was missing is a fact about the card: nothing recorded **where a card was
put into a graveyard from, or when**, so no permission filter could ask.

### Decision 1: the arrival record is Card data, stamped by `MoveCard`

`Card.PutInGraveyardFrom` (the zone the card came from) and `Card.PutInGraveyardThisTurn` (that arrival happened during the current
turn). `MoveCard` is the one zone move, so a mill, a discard, a counter, a destroy, a spell resolving and a sandbox drag are all
recorded without a caller having to remember. It rewrites the record on every move: a card that lands in a graveyard names its
source, and a card that goes anywhere else names nothing. That is CR 400.7 for free: a card that left the graveyard and came back is
a new object judged by its latest arrival only. A melded pair leaving to a graveyard stamps both cards.

A flag rather than a turn number, because `MoveCard` is a zone function and does not hold the game's clock. `onTurnBeganLocked`
clears `PutInGraveyardThisTurn` in every seat's graveyard as each turn begins, the per-turn-flag shape `SummonedThisTurn` already
has. `PutInGraveyardFrom` is a fact that does not expire and stays.

### Decision 2: the permission clause rides the filter

`PermissionFilter.PutInGraveyardThisTurn` is true for a card that arrived this turn from a zone that is not the battlefield
(`Card.PutInGraveyardNotFromBattlefieldThisTurn`). It is a property of the card being cast, not a game-state condition, so it is a
filter flag and not a registered `PermissionCondition` (see the 2026-10-13 second amendment, which draws the same line for a
spell-type gate). The cast path, the enumerator and the view all reach it through `CoversCard`, so they agree by construction.

### Snapshot

Additive within v9: `…cards[].putInGraveyardFrom`, `…cards[].putInGraveyardThisTurn` and
`seats[].castPermissions[].filter.putInGraveyardThisTurn`, all omitted at their zero values. A binary from before it drops the keys
on restore, which reads every card as having arrived nowhere and so shuts a permission that reads the record for the rest of the turn
the game was saved in. `-update-shape` recorded the paths; no version bump.

### What it does not do

Cards that count arrivals ("if three or more cards were put into your graveyard from anywhere other than the battlefield this turn",
Dimir Strandcatcher) cannot read this record: a card that has since left the graveyard no longer carries it. That wants a per-turn
tally, not a per-card fact.

Back-out: stop stamping in `MoveCard` and `TestMoveCardStampsGraveyardArrival` and six of Banon's tests fail; drop the stamp on a melded
pair's parts and `TestMeldedPairStampsBothCards` fails; drop the clear in `onTurnBeganLocked` and `TestGraveyardArrivalEndsWithTheTurn`
and `TestBanonForgetsLastTurnsArrivals` fail; drop the filter clause in `Matches` and `TestPermissionFilterReadsTheArrivalRecord` fails
with Banon opening a creature that died; drop the battlefield exclusion and the same filter test and
`TestBanonOpensDiscardedAndMilledCreaturesButNotOnesThatDied` fail.

## Amendment — 2026-10-15 (#3768): what a permission puts on the permanent its cast produces

`GrantsHaste` (suspend, CR 702.62a) was the only rider a permission could hang on the permanent its cast becomes. Two printed
sentences in deck request #3655 needed more, and both say "this way", which makes them the permission's and not the card's: the
same artifact hard-cast from hand enters untapped, and the same creature hard-cast has no haste.

- Edgar, Master Machinist: "Once during each of your turns, you may cast an artifact spell from your graveyard. If you cast a
  spell this way, that artifact enters tapped."
- Strago and Relm: "You may cast that card without paying its mana cost. If you cast a creature spell this way, it gains haste
  and 'At the beginning of the end step, sacrifice this creature.'"

### Decision 1 — `CastPermission.EntersTapped`, carried by the spell to its entry

Read once in `CastSpell`, where the permission is consumed, and copied onto the spell as `StackItem.EntersTapped`, because the
permission may be spent, swept or gone before the spell resolves. Resolution seeds it onto the entry event with `FaceDown`, before
the CR 614 pipeline runs (CR 614.1c), so the permanent ENTERS tapped: no `EventTapCard`, and every other entry replacement sees it.
A copy of the spell does not carry it (CR 707.10: a copy was not cast this way); `createSpellCopyLocked` lists its fields and leaves it
out. Snapshot: additive `stackMeta[].entersTapped` and `castPermissions[].entersTapped`.

### Decision 2 — `GainsKeywords` / `GainsAbilities`, one scoped effect with no duration

Keywords and ADR 0093 grant keys the permanent gains. Registered at cast, like suspend's haste and for the same reason, as one
layer-6 record (`AddKeywordsMod`, `GrantAbilitiesMod`) pinned to the spell's instance with no controller. The printed text states
no duration (CR 611.2a), so the record lasts for the new condition `WhileSourceOnceItLands`
(`ForThePermanentThisSpellBecomesDuration`): suspend's `WhileYouControlSourceOnceItLands` without the control test. It holds while
the spell is on the stack, then while the permanent is on the battlefield whoever controls it, and is over for good once the object
is anywhere else (a countered spell, or a permanent that left; CR 400.7). An instant or sorcery cast this way registers nothing:
it never becomes an object that could gain anything, which is the printed "if you cast a creature spell" for a permission that
opens only instants, sorceries and creatures. The granted "At the beginning of the end step, sacrifice this creature" is an ability
of the creature, so it fires at every end step it sees, starting with this one, and not once like a delayed trigger. Snapshot:
additive `castPermissions[].gainsKeywords` / `gainsAbilities`; the condition is appended to the enum.

Not covered: a rider with a duration of its own ("it gains haste until end of turn", Thundermane Dragon, which also needs a
power filter on a top-of-library permission).

Back-out: drop `EntersTapped: item.EntersTapped` from the resolving entry event and
`TestEdgarCastsOneArtifactFromTheGraveyardThatEntersTapped` fails; drop the `grantGainsForCastLocked` call and
`TestStragoAndRelmCastCreatureGainsHasteAndIsSacrificedAtEndStep` fails; give `WhileSourceOnceItLands` the control test and the
same test fails, as does `TestPermanentThisSpellBecomesDurationFollowsTheObject`.

## Amendment — 2026-10-15 (#3785): a graveyard cast at its mana cost that counts toward a once-each-turn limit only when made

Emet-Selch of the Third Seat: "Whenever one or more opponents lose life, you may cast target instant or sorcery card from
your graveyard. If that spell would be put into your graveyard, exile it instead. Do this only once each turn." Three of its
pieces already existed: the batch trigger (`OncePerBatch` over the two life-loss events, read through `b04OpponentLostLife`,
CR 603.2c), the exile replacement (`ExileOnResolution`, Torrential Gearhulk's) and the cast follow-up of #2173
(`CastPermission.FollowUp`, Conduit of Worlds). What was missing is small and is the whole of this amendment.

### Decision 1 — the offer is claimed under `AltCostKeyGrantedCost` with the price left empty

Snapcaster's grant is keyed `"flashback"`, which is wrong here for two reasons. A card that prints flashback offers its own
`"flashback"` at its own price, so the claim resolves to the card's offer and Emet's "cast" would be charged the flashback
cost. And a cast made under the card's own text is, by `castUsesGrantLocked`, not the permission's, so the follow-up would
not run. `effects.OfferGraveyardCast` (`granted_permissions.go`) keys the permission `AltCostKeyGrantedCost` with an empty
`Cost`, which `AlternativeCostFor` reads as "the card's mana cost", flash-timed and pass-closed like the free-cast grants
(CR 608.2g), with `ExileOnResolution` and a `FollowUp` key.

### Decision 2 — a claim of the permission's own synthesised offer is the permission's cast

`AlternativeCost.FromPermission` is set only on the offer `CastPermission.AlternativeCostFor` builds, and
`castUsesGrantLocked` answers true for it before asking whether the card opens its own zone. A card the graveyard already
lets you cast (Lingering Souls, Faithless Looting) cast through the permission therefore runs the follow-up and spends
`CastsLeft` and `PerType` budgets, as a Gravecrawler-style card under Underworld Breach's escape does. A claim of the card's
own key stays the card's own cast. The field is never serialised.

### Decision 3 — "do this only once each turn" is written by the follow-up, not by the grant

The limit counts a cast, so it is `Game.NoteDoneThisTurn` (per object, #3247's shape) called from the registered follow-up,
which runs only once the spell is on the stack through the permission. It is read when the ability triggers (no trigger once
spent) and again as it resolves. A declined or lapsed offer writes nothing and leaves the turn's use for a later batch.

### Decision 4 — the discount is a constructor, not machinery

`CostQuery.FromZone` was already on the modifier input; `effects.CastFromGraveyard()` is the predicate, so "spells you cast
from your graveyard cost {2} less" is `CostsLess(2, label, YourSpell(), CastFromGraveyard())`.

### Declared

"If that spell would be put into your graveyard, exile it instead" is implemented with flashback's wider replacement
(CR 702.34a: instead of anywhere else it would go when it leaves the stack). The only difference is an instant or sorcery
that would be shuffled into a library or returned to a hand; it is exiled here, never stronger than printed.

No wire or snapshot change: the follow-up key already rode `castPermissions[].followUp`.

Back-out: drop `FromPermission` from `castUsesGrantLocked` and `TestEmetSelchCastOfAFlashbackCardCountsAndPaysManaCost`
fails (the second loss offers again); drop the `FollowUp` key and `TestEmetSelchOffersOnlyOnceEachTurn` fails.

## Amendment — 2026-10-15 (#3787): a temporary play grant over "cards exiled with" a source, with a cost reduction

Urianger Augurelt's Play Arcanum: "Until end of turn, you may play cards exiled with Urianger Augurelt. Spells you cast
this way cost {2} less to cast." Kayla's Music Box prints the same grant without the discount and over "cards you own".
Two pieces were missing, both small.

**1. A STORED standing permission may carry the "exiled with" filter.** `PermissionFilter.ExiledWithSource` was built for
a STANDING permission derived off the battlefield (Valgavoth), where `stampStandingPermissionLocked` writes the granting
object into `ExiledWith`. A resolution that grants "until end of turn, cards exiled with this" has the object in hand
(`effects.HiddenRefOfActivation`), so it stores a `ScopeStanding` permission over exile with `ExiledWithSource` and
`ExiledWith` both set. That is a rule over the zone, not a set locked at resolution (CR 611.2c is the wrong reading
of "cards exiled with this": a card exiled with it after the activation is playable too), and the epoch in the
reference still ends the claim when the source leaves (CR 400.7). `effects.PlayCardsExiledWith` is the one door.
It is play, not cast, so a land is covered (CR 305.1) and `CastOnly` stays unset.

**2. `CastPermission.CostLess` — "spells you cast this way cost {N} less".** Generic mana taken off the price of a spell
cast through the permission. A reduction, not a flat price: `Cost` replaces the cost the cast would pay, while this
composes with it (a printed cost, a claimed alternative cost) and with the board's cost modifiers. It reaches the pricer
as `CostQuery.GrantReduction`, taken off in `applyCostModifiersLocked` with the board's reductions and emerge's, so every
increase has been added first and a `CostFloor` still applies after (CR 601.2f), and it clamps at {0} like any generic
reduction (`reduceGeneric`, including the mana announced for {X}). The one pricer (`costAfterModifiersLocked`) reads it
from the permission the cast claims (`grantReductionLocked`, the same `CastPermissionForClaimLocked` lookup the price and
the announce path use); the enumerator reads it from the permission it is announcing under, so the payment, the preview,
the view's price and the bot agree (#544). A hand or command-zone cast reads none.

**3. `PermissionFilter.OwnedByHolder`** — Kayla's Music Box's "cards YOU OWN exiled with it". `NotOwnedByHolder`'s mirror,
read in `permissionReachesPileLocked` where the holder and the owner are both known.

### Snapshot

Additive within v9, omitted when zero: `seats[].castPermissions[].costLess` and `...filter.ownedByHolder` (the filter
is also mirrored under the statics that carry one, hence the six lines in `testdata/snapshot_shape/v9.txt`). A file
written before them restores with no reduction and no ownership clause. A binary from before them drops the key and reads
the permission without the discount (weaker than printed) or, for the Box, without the ownership clause, which matters
only to a table that restores mid-turn after the Box changed hands. No wire change: the view prices through
`PriceCastForEffect`, which already reads the grant.

Back-out: pass `GrantReduction: 0` in `costAfterModifiersLocked` and `TestUriangerPlayArcanumCastsAnExiledSpellForTwoLess`
fails; read `0` for it in the enumerator and `TestUriangerEnumerationPricesTheDiscount` fails; take it off before the
increases are added (drop it from the full-modifier path) and `TestUriangerReductionComesOffAfterTheIncreases` fails; drop
the `OwnedByHolder` test and `TestKaylasMusicBoxPlaysOnlyCardsTheHolderOwns` fails.

## Amendment -- 2026-10-15 (#3846): one budget across permissions in different zones (CR 701.17c)

Locke, Treasure Hunter's Mug: "Until end of turn, you may cast a spell from among those cards", where "those cards" are
the cards each player milled. CR 701.17c: a milled card a replacement sent elsewhere (Rest in Peace, Leyline of the
Void: exile) is still one of them. A permission opens a cast out of ONE zone, so a set that spans a graveyard and
exile is two permissions, and a limited count (`CastsLeft`, #1729) on each would be one spell per zone.

**`CastPermission.BudgetGroup`** ties permissions into one budget: a cast that spends any member
(`consumeLimitedGrantLocked`) spends one use of every member (`consumeBudgetGroupLocked`), and a member that reaches zero
is dropped. `uuid.Nil` -- every permission written before the field -- is a budget of its own and takes the old path.
**`Game.GrantCastPermissionAcrossZonesForEffect`** groups a set of cards by where each is now (graveyard or exile, the
only zones a stored permission opens; any other zone is skipped), stores one `ScopeCards` permission per zone, and, when
there is more than one and the count is limited, gives them one `BudgetGroup`.

**The mill that reports where the cards went.** `Game.MillToZoneThenMovedForEffect` is `MillToZoneThenForEffect` with a
different continuation argument: every card the mill took out of the library, wherever it ended up, rather than only
those that landed in the destination. It is a separate entry point (`millTail.allMoved`, plain data), so no existing
caller's semantics change; a card a CR 614 window left in the library, and a cancelled mill, report nothing.

### Snapshot

Additive within v9, omitted when zero: `seats[].castPermissions[].budgetGroup`. A file written before it restores with
every permission a budget of its own. A binary from before it drops the key and reads a split Mug grant as one spell per
zone (stronger than printed, only for a table restored mid-turn across the upgrade). No wire change.

Back-out: stop assigning the group and `TestLockeMugSpansTwoZonesWithOneSpell` fails; mill with
`MillToZoneThenForEffect` in Locke and `TestLockeMugCoversCardsRestInPeaceExiled` fails.

## Amendment — 2026-10-15 (#2017): the total mana value budget over cards an impulse exile landed

Primeval Spawn: "When this creature leaves the battlefield, exile the top ten cards of your library. You may cast any number of
spells with total mana value 10 or less from among them without paying their mana costs." The 2026-10-14 amendment built the
running total (`CastPermission.ManaValueBudget`) for the copies of a card and said Spawn's cards, exiled by
`ExileTopWithPermissionForEffect`, would need "a helper that grants one budgeted permission". The helper turned out to be one
field, because the premise was off by a step: the impulse exile already stamps ONE `ScopeCards` permission over every card that
landed (`grantImpulseExilePermissionLocked`, CR 611.2c), not one per card. Nothing in the engine had to change.

### Decision 1: `effects.ExileTopWithPermission.TotalManaValueAtMost`

Setting it exiles the top `N` cards and grants the holder a free (`Cost: "{0}"`, so X is locked at 0), flash-speed
(`TimingFlash`, the cast being the resolving ability's, CR 608.2g), cast-only permission whose `ManaValueBudget` is the printed
total and whose `LapseOnPass` is `LapseStaysInExile`. It is the same shape the copy offer builds, so the cast gate, the view's
`castable_here` and the bot enumerator all read it through `spellManaValueWithinCap`, and a cast spends it through
`consumePassClosedGrantLocked`. Only the cards that really reached exile are named: a commander the owner sent to the command zone
(CR 903.9) and a library shorter than `N` are both handled by the exile path, not here.

### Decision 2: it is its own window, so it refuses the options that name another

`Free`, `CastOnly`, `WhileExiled`, `Window`, `FaceDown`, `Condition` and `Filter` mean a different grant (a turn-long impulse
window, a hidden pile, a conditional play), and a budgeted free cast with one of them set would be a different, unreviewed card. The
form returns an error rather than silently overriding any of them. `AnyColor` / `AnyType` stay legal (they are about mana
spent on additional costs, which a free cast can still owe).

### Decision 3: what it leaves alone

A land among the cards is exiled and stranded: the grant is cast-only (CR 305.1), so no land play is offered. An uncast card stays
in exile (`LapseStaysInExile`): the printed text says nothing happens to it. A permanent spell cast this way is an ordinary spell
cast from exile, not a copy, so it is not a token, and the original is the card cast.

No wire or snapshot change: `manaValueBudget` was added by the 2026-10-14 amendment and nothing else is new on a permission.

Back-out: drop the `ManaValueBudget` line in `applyBudgeted` and `TestPrimevalSpawnExilesTenAndSharesOneBudgetOfTen` fails (every
card is castable); drop `LapseOnPass` and the permission outlives the pass, so
`TestPrimevalSpawnWindowClosesOnPassAndLeavesTheRestInExile` fails; drop `CastOnly` and
`TestPrimevalSpawnLandIsExiledAndNotCastable` fails on the land play; take the self-replacement off the card and
`TestPrimevalSpawnEntry` fails on the free cast and the reanimation.

## Amendment — 2026-10-16 (#3938): a printed and a granted instance of one keyword are two offers

**The gap.** Decision 3 said "a card that both prints and is granted the same key keeps its printed cost". It consulted the
catalog first, and `CastOffersForLocked` dropped the grant's offer because it repeated a key already listed. That was the
strictly-weaker direction, and it was not the rule. CR 113.2c: "An object may also be granted additional abilities by a spell or
ability. If an object has multiple instances of the same ability, each instance functions independently." So a Faithless Looting
under Lier, Disciple of the Drowned has two flashback abilities. Each says "you may cast this card from your graveyard by paying
[cost] rather than paying its mana cost" (CR 702.34a), and its owner casts it with either one (CR 601.2b: the player announces
which alternative cost they intend to pay). The engine offered only the printed {2}{R}, never Lier's {R}. Past in Flames under a
second Past in Flames was offered its printed {4}{R} and never its {3}{R} mana cost, so its declared `Full` was not true.

**Options.**

1. *Keep the printed offer only* (decision 3 as written). It is weaker than the rules, and it costs the player real mana on the
   cards these grants exist for.
2. *Let the grant replace the printed offer* when it is cheaper. That is not the rule either: the player chooses, and a grant
   whose offer costs more in some other component (life, an escape's exile) would take away the cheaper printed one.
3. **Offer both, each under its own claim** (chosen; the most CR-faithful option). Each instance of the keyword is its own offer
   with its own cost, and the player picks one.

**Decision.**

- **The printed instance keeps the bare key.** `"flashback"` on a card that prints flashback is the printed offer, as before, so
  every claim, saved decision log and bot suite position written before this still means what it meant.
- **The granted instance is claimed as `game.GrantedInstanceKey(key)`**, which is `"flashback/granted"`. `CastOffersForLocked`
  lists it after the card's own offers whenever a permission's offer (`AlternativeCost.FromPermission`) repeats a key the card
  prints (`secondInstanceOf`). It is not listed when it repeats the printed offer's whole price (`sameOfferPrice`: the same mana,
  life and energy, and the same card component), because two abilities at one price are one choice: a printed retrace under
  Wrenn and Six's emblem is still one retrace offer. Nor is it listed when the printed offer is claimable only under a grant of its
  key (`RequiresGrant`): miracle's grant is the right to pay the printed miracle cost (CR 702.94a), not a second miracle. On a card
  that prints no such key, the granted offer is claimed under the bare key, exactly as before. `resolveAlternativeCostLocked`
  refuses the instance spelling wherever `CastOffersForLocked` would not list it (`grantedInstanceOffer`), so each offer has
  exactly one spelling.
- **The spell records the keyword, not the claim.** `StackItem.AltCost` and the cast event's `AltCost` take
  `game.AltCostKeyword(claim)`, which is `"flashback"` whichever instance paid. CR 702.34a's "if the flashback cost was paid" does
  not ask which flashback ability was used, and neither does a card that reads the key back (`PaidAltCost`). The exile on leaving
  the stack comes from the resolved offer (`AltCostExiles`), as it has since decision 3. Nothing new reaches a snapshot, and the
  wire carries one more offer in `alternative_costs`, with a key a client already treats as opaque.
- **The permission behind the claim.** `CastPermissionForClaimLocked` and `ForClaim` match a permission's `AltCostKey` against
  the keyword, so an instance claim is judged under the permission that made it. `castUsesGrantLocked` already counts a claim of a
  permission's own offer as that permission's cast (#3785), so a limited permission is spent by the instance claim and not by the
  printed one.

The same holds for every keyword a permission grants under a key the card may also print: escape (an Uro under Underworld
Breach), madness, and the rest. Flashback is the case the two cards need.

**Cards.** Lier, Disciple of the Drowned loses its caveat and is `Full`. Past in Flames keeps `Full`, which is now true.

**Tests.** `granted_instance_test.go` (game) covers the two offers, the printed claim still at its printed price, the instance
claim at the card's mana cost recording `"flashback"` with the flashback exile, the single offer when the prices match (a flashback,
and a retrace with its land discard), and the refusal of the instance spelling on a card that prints nothing.
`TestMiracleEnumeratorOffersTheMiracleCastOnly` (effects) still holds: a revealed miracle card is offered its miracle cast once. `two_flashbacks_test.go` (effects) covers Lier on a Faithless
Looting (both offers, both enumerated as moves, cast for {R}, exiled) and Past in Flames under a second Past in Flames (both
offers, cast for {3}{R}, exiled).

Back-out: restore the `seen` early return in `CastOffersForLocked`, and both effects tests and
`TestGrantedFlashbackBesideAPrintedOneIsASecondOffer` fail on the offer list. Record the claim instead of the keyword on the stack,
and the stack-item assertions fail.


## Amendment — 2026-10-16 (#3989): a granted cast paid by waterbending its mana value

Hama, the Bloodbender: "For as long as you control Hama, you may cast the exiled card during your turn by waterbending {X}
rather than paying its mana cost, where X is its mana value." The permission, the window and the timing already existed
(`forAsLongAsYouControlThisWindow`, `TimingYourTurnOnly`). The price did not: `CastPermission.Cost` is plain mana, and
`TapPermanentsCostForCastLocked` found a tap cost only on the card, on a next-spell promise or on a permanent's grant.

**The rules.** CR 701.67a: "Waterbend [cost]" means "Pay [cost]. For each generic mana in that cost, you may tap an
untapped artifact or creature you control rather than pay that mana." CR 701.67b: when the waterbend cost is part of a
spell's total cost, the taps may pay only the generic mana in the waterbend cost. Hama's waterbend is not an additional
cost but the ALTERNATIVE cost (CR 118.9: "rather than paying its mana cost"), so CR 118.9d applies: additional costs, cost
increases and cost reductions are applied to it. And the card's own {X}: the cast pays neither its mana cost nor an
alternative cost that includes that X, so 0 is the only legal choice (CR 107.3b).

**Options.**

1. *Cost "{0}" plus a waterbend {N} added on top*, as Waterbender's Restoration's printed waterbend is. Simple, but the
   {N} would be added after the cost modifiers (`costAfterTapsLocked` runs last), so a reduction would find {0} to reduce
   and be lost, and the enumerator, which prices `CastPrice.Base`, would see a free spell.
2. *Stamp `Cost: "{N}"` at grant and grant convoke-like taps.* The mana value is the face being cast (CR 202.3d for a
   split card), unknown at grant, and taps that pay any generic would pay a tax or a kicker too (stronger than printed).
3. **The waterbend {N} is the cost being paid, and the taps cover only it** (chosen; the most CR-faithful). The same split
   an activated ability's "Waterbend {N}:" already makes (`waterbend_cost.go`): the mana lives in the cost, and the tap
   clause names which part of it the taps may pay.

**Decision.**

- **`CastPermission.WaterbendManaValue`** (bool). `Cost` stays empty. `NamesPrice()` reports a permission that prices the
  cast itself (a `Cost`, or this), and the readers that asked `Cost != ""` for that question now ask it: CR 118.6's
  printed-cost check, the printed alternative costs a permission at the printed cost lets a card use, the claimability of a
  granted offer, and the enumerator's no-mana-cost gate.
- **`CastCostFor` charges `WaterbendManaValueCost(card)`**, "{N}" for the face being cast with {X} at 0 (the same read as
  Bolas's Citadel's life), so the commander tax and every cost modifier apply to it (CR 118.9d), the X lock holds
  (CR 107.3b), and `CastPrice.Base` is the waterbend's mana, which is what the enumerator and the view price.
- **`TapPermanentsCostForGrantLocked`** is `TapPermanentsCostForCastLocked` for the permission the cast claims. Under a
  waterbend permission it hands back `effects.Waterbend("{N}")` (installed as `game.WaterbendTapCost`) marked
  **`TapPermanentsCost.InCost`**: its Extra is already in the cost, so `tapPermanentsAdjusted` takes the taps off it rather
  than adding it, and the cast validator measures the budget against the total the modifiers leave (`WaterbendBudget`, as
  for an ability), so a tap never pays a Thalia tax, a kicker or a {1} a reduction already removed (CR 701.67b). The
  validator, the pricer, the view's tap picker (`castStampsBaseFor`) and the cast path all ask it.
- **One tap cost per cast.** Under the permission the waterbend replaces any convoke or improvise the card or the board
  would give the spell. A spell with both is rare and the loss is the weaker direction.

### Snapshot

Additive within v9, omitted when false: `seats[].castPermissions[].waterbendManaValue`. A file written before it has no such
permission. A binary from before it drops the key and reads Hama's permission as a cast at the printed mana cost, which is
only for a table restored across the upgrade while Hama's card is still in exile.

**Card.** Hama, the Bloodbender, declared with a caveat: "A spell cast this way can't also use convoke or improvise." (the one-tap-cost limit above).

**Tests.** `hama_the_bloodbender_test.go` (effects): the mill and the noncreature, nonland pick; the permission ends when Hama
leaves; the cast paid by three taps, by {3} of colourless mana, by two taps and {1}; four taps refused; not on an opponent's
turn; Thalia's {1} paid with mana and never a tap; Goblin Electromancer's reduction coming off the waterbend; an {X}{U}{U}
spell cast as a waterbend {2} with X locked at 0. `waterbend_grant_view_test.go` (protocol): the view's tap picker reads
"Waterbend {3}" with a budget of 3, and the enumerator offers the cast at three mana and not at two.

Back-out: drop the `CastCostFor` clause and the cast is priced at its printed cost, so the colourless and enumerator tests
fail; drop `InCost` and the {N} is charged twice, so every tap test fails; measure the budget against the base cost and
`TestHamaReductionComesOffTheWaterbend` fails.

## Amendment — 2026-10-16 (#2651): "exiled with" names the ability, and a play from exile carries the link

The 2026-10-13 amendment stamped a resolving ability's linked exile with its source OBJECT only, and said the ability ref was
not added. CR 607.2a links an ability to an ability: "the second ability refers only to cards in the exile zone that were put
there as a result of an instruction to exile them in the first ability". An object with two exiling abilities, or a reader that
has to see the link after the card has left exile, needs more than the object. Prowl, Stoic Strategist needs both: "Whenever
Prowl attacks, exile up to one other target tapped creature or Vehicle. For as long as that card remains exiled, its owner may
play it. Whenever a player plays a card exiled with Prowl, you draw a card and convert Prowl."

- **The stamp names the ability too.** `Card.ExiledWithAbility` is the catalog row (`AbilityRef`, ADR 0041 P9) of the item
  that was resolving when the card was exiled, beside `Card.ExiledWith`. That is the identity a stack item already carries
  (`StackItem.Params.Ability`) and the one #3933's entry provenance reads (`Event.CauseAbility`). It is cleared by `MoveCard`
  with the object half.
- **No caller threads it.** `routeCardToZoneLocked` fills `zoneRoute.ExiledWithAbility` from `Game.resolving`
  (`linkedExileAbilityLocked`) when the route names a link, and only while a resolution is open, the guard
  `stampEntryProvenanceLocked` uses. It is the zero ref when nothing is resolving, when a spell is resolving, when the item has
  no catalog row (a legacy closure trigger), or when the item's source object is not the one the route links to. So every linked
  verb (`ExileLinkedForEffect`, `ExileLinkedThenForEffect`, the hideaway exile, the revealed-hand pick) stamps it without
  change.
- **A replacement's link names no ability.** The two landing functions take the ability from the event's route only while
  `ReplacementEvent.ExiledWith` is still the route's own link (`ReplacementEvent.exiledWithAbility`). Valgavoth's replacement
  writes its own link and resolves nothing, so CR 607.2b's link stays an object link.
- **Readers.** `Game.ExiledWithRowForEffect(source, slot, name)` lists the cards in exile linked to that object by that row,
  matched by `AbilityRef.IsRow`: the slot and the row's name (a triggered row's Key, an activated row's Label), as restore's
  Q2 rule matches a row. `ExiledWithForEffect(source)` is unchanged, so the borrowers and Duplicant read every linked exile of
  the object as before. This also closes the weakness ADR 0155 Decision 6 declared, for any reader that wants it.
- **A play from exile carries the link.** Playing the card moves it out of exile and clears its link (CR 400.7), so a trigger on
  the play has nothing left to read on the card. `Event.ExiledWith` and `Event.ExiledWithAbility` copy the link the card had onto
  the spell's `EventCast` from exile (read off the copy `castSpellLocked` takes before the move) and onto a land's Played
  `EventZoneMove` from exile (`entryLanding.playedFromExile`). `game.PlayedCardExiledWithRow(ev, source, slot, name)` is the
  read. A card put onto the battlefield from exile is not played and carries nothing.
- **The event-log readers stay.** The cards that read `b27ExiledWith` keep their walk. It is already keyed on the
  ability (its label), so it is CR 607.2a-exact, and it runs at resolution, not on every layer pass. Moving them would make a
  table restored from a file written before this amendment forget the cards it had exiled, because those cards carry no stamp.
  Skyclave Apparition's `b27EverExiledWith` reads cards that have left exile, which the stamp cannot answer.
- **Card side.** `effects.ExileWithPermission{Linked: true}` is the linked exile with the play grant.

### Snapshot

Additive within v9, omitted when empty: `…cards[].exiledWithAbility` (`{key, slot, ref, name}`) and `events[].exiled_with` /
`events[].exiled_with_ability`. A file written before it restores the object link alone, which is every linked exile before it,
and a reader by row finds nothing there, as it found nothing before. A binary before it drops the keys; it has no reader by row
and no Prowl.

**Card.** Prowl, Stoic Strategist // Prowl, Pursuit Vehicle, Full.

**Tests.** `exiled_with_test.go` (game): the stamp carries the ability and `MoveCard` clears it; a route's ability rides only its
own link; `PlayedCardExiledWithRow` matches a cast and a played land from exile, and not a put, another row, another object or a
link with no ability; the ability survives a snapshot. `prowl_stoic_strategist_test.go` (effects): the attack trigger's exile is
linked to the object and the row, and its owner may play it; only another tapped creature or Vehicle is a target; a cast from
exile and a land played from exile draw and convert; a card linked by a replacement, by another row, to another Prowl object
or to another permanent does not; the back face's counters and its convert on the second resolution.

Back-out: stop filling `zoneRoute.ExiledWithAbility` and the Prowl tests that cast its card fail; drop the cast's
`stampPlayedExiledWith` and the cast test fails while the land test passes.

## Amendment — 2026-10-16 (#1987): playing from other players' graveyards, and a ban on playing from your own

Shaman's Trance: "Other players can't play lands or cast spells from their graveyards this turn. You may play lands and cast
spells from other players' graveyards this turn as though those cards were in your graveyard." Its two 2004-10-04 rulings:
"This card does not allow you to activate the abilities of cards in other people's graveyards", and "You can use Flashback on
cards to cast them."

**The rules.** A spell may be cast only if a rule or effect allows it and none prohibits it (CR 601.3), and "can't" beats
"may" (CR 101.2). The caster controls a spell cast from another player's graveyard (CR 601.2a), and it goes to its owner's
graveyard afterwards (CR 400.3); a land played from there enters under the player who played it (CR 110.2) and is that
turn's land play (CR 305.1, 305.2). "This turn" ends at cleanup (CR 514.2).

### Decision 1 — the permission is a stored standing grant per other seat, and needs no new field

#1022 and #1035 already gave a stored `ScopeStanding` permission a seat (`ZoneOwner`), and taught the cast path
(`foreignPileForCastLocked`), the enumerator (`castZones` walks other seats' piles under any grant) and the view (the
per-holder stamps of `stampGrantedPermissions`) to follow it. Xanathar's library grant was the only user. Shaman's Trance is
`effects.GraveyardPlayThisTurn`'s grant (Yawgmoth's Will) pointed at another pile: one `CastPermission{Zone: graveyard,
Scope: standing, ZoneOwner: <seat>}` per other seat still in the game, at the printed cost, until end of turn.

Rejected: a new "every other graveyard" flag on one permission. It would be read in `permissionReachesPileLocked` only, so it
is smaller, but it is a second spelling of a scope `ZoneOwner` already has, and every surface that reads `PileOwnerFor`
would need to learn it.

### Decision 2 — "as though those cards were in your graveyard" is the card's own text, and rule 3 gives way

The card's own way out of a graveyard (flashback, escape, a Gravecrawler-style permission, an aftermath half) is opened by
`cardOpensZoneLocked`, which asks the card and not whose pile it is in; the enumerator and the view open another seat's pile
only under a permission. So under the grant those offers already reach the caster, which is the flashback ruling.

What did not hold was the other price. `validateCastPathLocked`'s rule 3 ("a zone the card prices must be paid for") refused
a cast at the printed cost of any card that also prints a graveyard price, even when a permission at the printed cost was the
reason the cast was legal. That is wrong for Shaman's Trance, and it was already wrong for Yawgmoth's Will and Muldrotha over
the caster's own graveyard. Rule 3 now gives way to a permission that names no price of its own (`AltCostKey` empty and
`NamesPrice()` false): the caster may cast the card for its mana cost or claim its printed offer. A permission with a price
(Snapcaster's flashback, Underworld Breach's escape, a flat `Cost`) is rule 4's and unchanged.

This reverses a direction the view's #1015 section of this ADR pinned ("a card that DECLARES and PRICES its own zone owes
that price even under such a grant"). That pin was the safe direction, weaker than printed, but not the rules: CR 601.3 makes
the permission the reason the cast is legal, and nothing makes the card's own offer mandatory once something else allows the
cast (Muldrotha's 2020-11-10 ruling, "If it has an alternative cost, you may cast it for that cost instead", reads the same
way). `TestCastableHereUnderAPrintedCostGrantWhenTheCardsOwnPriceIsUnpayable` (protocol) replaces the test that pinned it.
With no permission, rule 3 is unchanged: a flashback card in its owner's graveyard still owes flashback, and ADR 0118's
granted offers still skip it. Under a printed-cost permission a granted offer (Jodah) now reaches such a card, which is
CR 118.9a's "where the printed mana cost could be paid".

### Decision 3 — the public stamp asks the pile rule too

`CastPermissionOnCardForEffect`, the public `exile_play` read, asked a stored permission only whether it covered the card.
A stored standing permission over a graveyard covered every graveyard card at the table, so the holder's grant over one
seat's pile was stamped on every other pile, the holder's own included. It now asks `permissionReachesPileLocked` as the
cast path does. A `ScopeCards` permission is unaffected (the pile rule passes it), and so is exile (shared).

### Decision 4 — the ban is a rules-gate ScopedEffect read by both gates

"Can't play lands or cast spells from their graveyards" is Memory Vessel's hand ban (the 2026-10-10 amendment, Decision 2)
one zone over, stored for the same reasons: a new kind, `cantPlayFromOwnGraveyard` (reader `rule`, scope game, reads
`Player`), registered by `Game.CantPlayFromOwnGraveyardThisTurnForEffect(source, players, label)`
(`game/graveyard_play_ban.go`), one mod per banned player in one record, until end of turn. `uuid.Nil` names nobody, never
everybody, because the printed form is "other players". The reader `graveyardPlayBanLocked` refuses a play only out of the
banned player's OWN graveyard, told by the card's owner (a card in a graveyard is in its owner's, CR 404.1), and asks the
duration itself. `CastGateLocked` and `LandPlayGateLocked` both ask it, so the cast path, a land played during a resolution,
the enumerator and the view's `cant_cast` stamp all see it. Activated abilities of cards in a graveyard are not plays and
stay open.

### Snapshot

Additive within v9: the mod kind `cantPlayFromOwnGraveyard`, and a stored permission with `zoneOwner` over a graveyard (the
field existed). A binary from before this refuses a restore point naming the kind with `ErrUnknownEffectKey`; one that knows
the field reads the permission as it is. Fixture `v9/shamans_trance.json`.

**Card.** Shaman's Trance, with a caveat: the caster's own permissions over "your graveyard" (Underworld Breach's escape,
Lurrus, Muldrotha) do not reach the other graveyards. The cards there are castable at their printed cost anyway; what is lost
is the extra price such a permanent offers, which is the weaker direction. Its prompt picks among one permission per card,
and teaching it to offer a standing permission's price beside a stored one's is a composition question for every graveyard
grant, not this card's.

**Tests.** `graveyard_play_ban_test.go` (game): the ban refuses a spell and a land out of the banned player's own graveyard
and nothing else, and ends with the turn; an empty list bans nobody; a flashback card under a printed-cost permission is
offered and cast at its printed cost; a stored grant over another graveyard is found on that pile only.
`shamans_trance_test.go` (effects): every other graveyard opens and the caster's own does not; a land from another graveyard
is the turn's land play under the caster with its owner kept; a Bolt cast from there is the caster's and returns to its
owner's graveyard; a flashback card there offers both prices and flashes back; the ban and its `cant_cast`; both halves end
with the turn; the enumerator and the view's `castable_here` offer the other graveyards to the caster only; a restore keeps
both halves.

Back-out: drop rule 3's exception and the flashback-offer tests fail; drop the pile check in `CastPermissionOnCardForEffect`
and `TestAStoredGrantOverAnotherGraveyardIsOnlyOnThatPile` fails; drop either gate's read and the ban tests fail.

## Amendment (2026-10-19, [#4128](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4128)): a cast ban on a name, for a duration (Reflector Mage)

The 2026-09-23 amendment (#1316) built the granted cast ban and named Reflector Mage as the card it was for; what it could not
say was WHICH spells. `PermissionFilter` had type and zone words and no name. Reflector Mage bans "spells with the same name as
that creature" until its controller's next turn.

**Decision.**

1. `PermissionFilter.Named string` (empty constrains nothing) is read in `Matches`. It is the name the ability read off the
   creature before it left the battlefield, so a Clone copying a Bear bans "Grizzly Bears", and a token's name counts.
2. The ban is a **kind of its own**, `CastBanNamed`, not `CastBanOutright` with a name filter. This is the rollback half of the
   snapshot rule: a binary from before this change reads a restore point it did not write, drops `filter.named`, and would
   restore an outright ban on every spell. Under its own kind it reads an unknown kind, which `CastBanRule.forbids` answers
   `false` (its default arm), so a rollback can only end the ban early, which is weaker than printed, never stronger. An empty
   `Named` bans nothing.
3. Nothing else changes: the ban is a `PlayerStatic` on the creature's OWNER with a `DurationUntilYourNextTurn`, read by the one
   gate every cast already passes (`CastGateLocked`), so the hand, the command zone and the enumerator are covered, and the
   wire's `cant_cast` stamp needs no new field.

**Snapshot.** Additive within v9: `filter.named` on every `PermissionFilter`. Recorded in `testdata/snapshot_shape/v9.txt`.

**Card.** Reflector Mage, `full`.

**Tests.** `cards/effects/roonb_cards2_test.go`: the bounced name is banned (recast from hand included) and another name is
not; the ban ends as the Mage controller's next turn begins; the OWNER is banned and the thief is not; a Clone's copied name
is the banned one; a target that left in response bans nothing.

Back-out: make `CastBanNamed` forbid nothing and the ban test fails; drop the `Named` clause from `Matches` and the
"another name is not banned" assertion fails; ban `target.Controller` and the owner test fails.

## Amendment (2026-10-19, [#4021](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4021)): any number of exiled spells, each with mana value N or less

Hazoret's Undying Fury: "Shuffle your library, then exile the top four cards. You may cast any number of spells with mana value
5 or less from among them without paying their mana costs." The engine had the two neighbours and not this one. A plain
`ExileTopWithPermission{Free: true}` with `LapseOnPass` is one pass-closed permission that the first cast spends whole ("a spell
from among them", weaker than printed), and `TotalManaValueAtMost` (the 2026-10-15 amendment) keeps the permission for the
rest of the set but caps a running TOTAL, not each spell.

**Decision.**

1. `CastPermission.AnyNumber bool` (`json:"anyNumber,omitempty"`) marks a pass-closed `ScopeCards` permission that a cast does
   not use up. `spendManaValueBudget`, the one place `consumePassClosedGrantLocked` decides between "remove the permission" and
   "keep it for the cards not yet cast", now also keeps an `AnyNumber` permission, with the cast card struck from `Cards` (a
   fresh slice, for the undo snapshots' sake) and no budget to spend. The per-spell ceiling is the field that already existed,
   `MaxSpellManaValue`, judged by `spellManaValueWithinCap` against the face as it is cast with X at zero, so the cast path, the
   view's `castable_here` and the bot enumerator refuse the same casts.
2. `effects.ExileTopWithPermission.EachManaValueAtMost` builds it: the same free (`Cost: "{0}"`), flash-speed, cast-only,
   `LapseStaysInExile` permission `TotalManaValueAtMost` builds, with `MaxSpellManaValue` and `AnyNumber` in place of the budget.
   Both forms are one window, so both refuse the options that name another (`Free`, `CastOnly`, `WhileExiled`, `Window`,
   `FaceDown`, `Condition`, `Filter`), and the two cannot be set together.
3. The rest is the 2026-10-15 amendment's: a land among the cards is stranded, an uncast card stays in exile, the casts are
   grants taken with ordinary `cast_spell` actions after the ability resolves (CR 608.2g), and the window closes when the holder
   passes.

**Rejected: one permission per exiled card** (the way `CopyAndCast` offers one per copy). It gives each card its own window
that closes on its own pass, and a permission per card is a list the bot enumerator and the view would walk twice for no gain;
the one permission already names every card.

**Rejected: widening `TotalManaValueAtMost` to carry a per-spell cap.** A running total spends down and a per-spell cap does not;
one field meaning both is how a total of 5 and a cap of 5 get confused.

**Snapshot.** Additive within v9: `seats[].castPermissions[].anyNumber`, recorded in `testdata/snapshot_shape/v9.txt`. A binary
from before this change reads the field as absent and offers one cast, which is weaker than printed, never stronger.

**Card.** Hazoret's Undying Fury, `full` (its lands clause shipped with #3995).

**Tests.** `cards/effects/hazorets_undying_fury_test.go`: any number are cast and the mana value 6 card is refused; exactly four
leave the library after the shuffle; the pass closes the window and the rest stay in exile; the enumerator keeps offering the
other cards after a cast; the permission survives a restore; the lands are held through the next untap step.

Back-out: drop `perm.AnyNumber = true` in `applyBudgeted` and the second cast, the enumerator and the restore tests fail; drop
`perm.MaxSpellManaValue` and the mana value 6 card is castable, failing the first.

## Amendment 2026-10-19: Shaman's Trance's "as though those cards were in your graveyard" reaches the holder's own permissions (#4177, #4173)

Shaman's Trance (#1987) stored one permission per other seat and left the caster's permissions over "your graveyard"
(Underworld Breach's escape, Yawgmoth's Will, Lurrus, Muldrotha, a flashback grant) reading the caster's own pile only.

`CastPermission.GraveyardsAreYours` is set on the Trance's own stored grants. While one is live
(`graveyardPermissionsReachAllLocked`), `permissionReachesPileLocked` (now a `Game` method) lets every permission its holder has
over THEIR OWN graveyard (a `ZoneOwner` of none) reach a card in any graveyard, so the cast path, the enumerator and the view's
per-holder stamps all agree. The flag rides the stored grants rather than a new scoped record so a resolved Trance stores no new
record: `v9/shamans_trance.json` stays a subset of a fresh render.

Two more changes make an offer from a derived permission reachable beside a stored plain one, which Trance + Breach needs:
`CastPermissionForClaimLocked` also looks at the permissions the holder's permanents give (`standingCastPermissionsLocked`) for
a claimed alternative cost, and `otherGrantedOffersLocked` lists their offers beside the stored grant's. The per-type ranking
(`perTypeOthersLocked`) counts the cards of every graveyard under the flag.

A card cast that way is still its owner's and returns to its owner's graveyard (CR 400.3); what such a permission asks of "your
graveyard" (escape's three exiled cards) is paid from the caster's own.

**Snapshot.** Additive within v9: `seats[].castPermissions[].graveyardsAreYours`. A binary from before this change drops the key
and reads the Trance as it did, which is weaker than printed, never stronger.

Tests: `cards/effects/shamans_trance_reach_test.go`. Back-out: make `permissionReachesPileLocked` ignore the flag and the Breach
escape, the end-of-turn and the restore tests fail.

## Amendment — 2026-10-22 (#2947, #2910): a copy of a card is cast during the resolution that made it

The 2026-10-13 card-copy amendment used a permission that lapses on the holder's next priority pass as a stand-in for CR
608.2g. Kaervek, the Punisher was then declared Full on that composition (#3478), but resolving a real crime trigger left no
decision inside the resolution: the trigger finished, the active player received priority, and the copy was cast afterwards,
after anyone who held priority first had been able to act. CR 707.12 says the copy is created "and then cast while another
spell or ability is resolving", and CR 608.2g says a spell cast that way is cast "except no player receives priority after
it's cast. That spell becomes the topmost object on the stack, and the currently resolving spell or ability continues to
resolve". Every `CopyAndCast` card prints the same instruction, so the stand-in was wrong for all of them in the same way.

A draft of this amendment (2026-10-14, branch `fix/2947-resolution-copy-casts`) proposed a separate door used by Kaervek
alone. That is replaced here by one flag on the existing offer, set by every `CopyAndCast` card.

### Decision

1. **The door.** `CardCopyOffer.DuringResolution` asks for the cast inside the resolution. `CopyCardsAndOfferCastsForEffect`
   still makes the copies and their exact permissions (one per copy, or one over the set for a mana value budget), then queues
   a `resolution_cast` prompt over the copies (`game/resolution_copy_cast.go`). `effects.CopyAndCast` always sets the flag, so
   Mizzix's Mastery, Psionic Ritual, Arcane Bombardment, Isochron Scepter, Blue Mage's Cane, Spellweaver Volute, Uldaros
   Theorix and Kaervek all move to it. Cipher's and paradigm's offers call the engine door without the flag and keep the
   pass-closed window. Outside an open resolution the flag does nothing.
2. **The prompt is data.** `PendingChoiceResolutionCast` carries the chooser, the source, its name (`Reason`) and the copies
   still on offer (`DiscardOptions`). It is queued while the resolution is open, so it is stamped `midResolution`: it blocks the
   table and holds the CR 704.3 boundary (`resolution_pause.go`). No Go continuation is stored. Answering it lets the existing
   resolution bookend carry on (step walks, mode walks, the resolution tail), and the permissions, the copies and
   `ResolutionOpen` already persist. A table paused on it is a restore point.
3. **The cast answer.** A `cast_spell` with `resolution_cast` set to the prompt's id goes to `Game.CastForResolutionChoice`,
   which checks the chooser and the candidate, then calls `castSpellLocked` through the copy's own permission. The actions
   layer asks for no priority for it, as it does for Word of Command's `controlled_play`. An ordinary `CastSpell` of a
   candidate by its chooser is routed to the same path, so the prompt can never be left offering a copy that has gone. The
   price, X, targets, modes, additional costs, cast prohibitions, split second and the "if you do" follow-up are all the
   existing ones. Payment is strict, `force_cast` and a paper payment are ignored, and the auto-tapper may use every ordinary
   mana source the caster does not lock out (CR 608.2g lets them activate mana abilities). Word of Command's land-only rule is
   not used: nobody else is making the choices. The cast notes no action, since nobody receives priority after it.
4. **Settling.** A successful cast takes its copy, and any copy that can no longer be cast, off the prompt. The prompt stays
   open while a copy is left (CR 707.12a: each is its own "may") and is answered once none is. A refused cast (unpaid, an
   illegal target) leaves it open for another try or a decline. Each cast copy becomes the topmost object on the stack, so the
   order of the casts is the order the player chose.
5. **Declining.** The empty `resolve_choice` is "cast no more of them". It is always legal, and it withdraws every permission
   naming the copies left, so the CR 704.5e sweep at the boundary removes them. The originals stay where the earlier
   instruction put them.
6. **Passing priority is not an answer.** `passPriorityAsHolderLocked` now asks the choice gate before it closes the holder's
   pass-closed windows. Before, a refused pass still lapsed those windows, which would have ended the copy's offer under its
   own open prompt. A refused pass now changes nothing, for every blocking prompt.
7. **What waits for the resolution.** Spell-cast triggers (a copy aimed at an opponent is a crime, and Kaervek triggers again)
   and state-based actions are put on the stack and performed at the boundary after the last answer, so they land above the
   copies (CR 603.3). Kaervek's 2 life is the permission's follow-up: it is lost as the copy is cast, inside the resolution,
   and a life replacement prompt it raises is mid-resolution too.
8. **Legal moves and the client.** The enumerator answers the prompt with every cast of each copy that the engine would take
   (`castMovesFromZone` over the copy's zone, each move naming the prompt) and the always-legal decline
   (`legal/resolution_cast.go`). Bots and MCP clients use the same moves. The view lists the copies still on offer as the
   prompt's `options` and `eligible`, face up (they are printed on a face-up card in a public zone), and the wire cap does not
   fold moves that carry `resolution_cast`. The client's `ResolutionCastSheet` shows the copies, the casts and "Don't cast".
9. **Departure.** A chooser who leaves takes the prompt with them (CR 800.4a; the departure table drops it) and the resolution
   finishes without the casts.

### Limits, stated

- An instruction the card's Go body runs after `CopyAndCast` (Mizzix's Mastery's and Psionic Ritual's "Exile this") runs
  before the casts rather than after them, because it is the same body. No card on this door has a later instruction whose
  outcome a cast could change. A later instruction that needs the casts first would be a step (`step_walk.go`).
- Casting a card from hand during a resolution (#2110, `offerFreeHandCast`), discover, cascade, cipher and paradigm keep the
  grant stand-in. They are separate doors and are not changed here.

### Snapshot

No new field and no schema bump. `resolution_cast` is a new value of the pending-choice kind, and a binary that does not know it
refuses the restore point (`KnownChoiceKind`). The frozen corpus is untouched.

Tests: `game/resolution_copy_cast_test.go`, `legal/resolution_cast_test.go`, `protocol/resolution_cast_view_test.go`,
`cards/effects/kaervek_the_punisher_test.go` and the other `CopyAndCast` cards' tests, and
`client/src/lib/resolutionCast.render.test.ts`. Back-outs: drop the `queueResolutionCastLocked` call and
`TestKaervekResolutionQueuesOptionalCastDecision` and `TestResolutionCastHoldsTheResolutionAndCastsInAnyOrder` fail; drop the
early gate in `passPriorityAsHolderLocked` and `TestKaervekDecliningTheCopyLosesNoLife` fails; drop the withdrawal in
`resolveResolutionCastLocked` and `TestResolutionCastRefusalsAndDecline` fails.

## Amendment — 2026-10-23 (#4243): cipher, cascade, discover and paradigm cast during their resolution

The 2026-10-22 amendment moved the `CopyAndCast` cards onto a `resolution_cast` prompt and left four doors on the grant
stand-in: cipher's copy, paradigm's copy, cascade's hit and discover's card. Under the stand-in the cast came after the
trigger had finished, once the holder passed priority, so on another player's turn the active player received priority first
and the table could act before the cast; the "cast" and the "resolving" the rules describe were two moments. The rules name one
moment for each:

- CR 702.85a: cascade's "You may cast that card without paying its mana cost ... Then put all cards exiled this way that
  weren't cast on the bottom of your library in a random order."
- CR 701.57a: discover's "You may cast that card without paying its mana cost ... If you don't cast it, put that card into your
  hand. Put the remaining exiled cards on the bottom of your library in a random order."
- CR 702.99a: cipher's "you may copy the encoded card and you may cast the copy without paying its mana cost".
- CR 702.192a: paradigm's "create a copy of this object in exile. You may cast the copy without paying its mana cost".
- CR 707.12 and CR 608.2g: each is cast "while another spell or ability is resolving", with no priority after it and the
  resolving ability continuing to resolve.

### Decision

1. **Cipher and paradigm** set `CardCopyOffer.DuringResolution`. That is the whole change: the offer, the copies, the prompt and
   the answers are the 2026-10-22 ones. After this, no caller of `CopyCardsAndOfferCastsForEffect` leaves the flag false; it
   stays on the offer for the tests that build one by hand, and outside an open resolution it still does nothing.
2. **Cascade and discover** offer the hit through the same prompt. `CascadeForEffect` and `DiscoverForEffect` walk the library as
   before and stamp the same pass-closed permission on the hit (`{0}`, flash timing, cast only, the mana value cap, and
   `LapseToLibraryBottom` or `LapseToHand` with the `DiscoverGrant`), then queue a `resolution_cast` over the one hit
   (`queueResolutionCastOfferLocked`). The candidate is a real card in exile, not a copy, so `ResolutionCastCandidateLocked` no
   longer asks for `Card.CardCopy`; a card is on offer when one of the chooser's `ScopeCards` permissions names it. The old yes/no
   (`may_cast`) is gone for these two: the prompt's casts are "yes" and its decline is "no", so the cast's targets, modes and X
   are chosen in the one answer.
3. **The cards passed over** ride the prompt as `PendingChoice.CastPile`, plain data. When the prompt is answered, by a cast or a
   decline, `finishResolutionCastLocked` puts the pile on the bottom of the library in a random order together with the hit when it
   was not cast, so the hit's place in the order is as random as the rest (before, an accepted cascade bottomed the pile at the
   answer and the hit later on its own, which left an uncast hit always below the pile).
4. **Declining** is the empty `resolve_choice`. It now lapses the permission exactly as the pass it replaces would have
   (`lapseCastPermissionCollectingLocked`): a discovered card goes to the hand and the discover completes (`EventDiscover`, a
   commander's CR 903.9b question included), a cascade hit joins the pile for the random bottom, and a copy is only withdrawn, as
   before. A cast consumes the permission through `castSpellLocked` as before, and completes the discover right after
   `EventCast`.
5. **What a later instruction reads.** `Discover.Then` was a Go closure that ran on the answer. It is gone. `DiscoverForEffect`
   answers the walk's result at once (CR 701.57c: the discovered card and its mana value do not depend on the answer), and
   `effects.Discover` files it as a step record under `DiscoverStepKey` ("discover": `IDs[0]` the card, `Ns[0]` its mana value). The
   one card that read it, Hit the Mother Lode, is now `ResolveSteps(Discover{N: 10}, Instruction(treasures))`: the step walk parks
   the Treasures behind the prompt and a restored table still makes them (`step_walk.go`, `step_record.go`).
6. **Restore.** Nothing but `CastPile` is new, and the closures are gone from these four doors: the cascade trigger is a keyed
   catalog item, the paradigm and cipher triggers are keyed bodies, and the prompt, the permission and the pile are data. A table
   paused on any of them is a restore point and resumes (tests below). `closure_fields.txt` is unchanged. A table written by a
   binary from before this change that holds a cascade or discover permission keeps the pass-closed window, which this change
   leaves in place as the fallback for an offer made with no open resolution (nothing in play makes one).
7. **Wire and client.** The prompt carries `may_cast_keyword` ("cascade", "discover"), `accept_label` and `decline_label`, the
   fields the `may_cast` it replaces carried. `ResolutionCastSheet` names the exiled card and what the decline does ("Put it on the
   bottom", "Put it into your hand"). `PendingChoice.CastPile` is not on the wire. The enumerator needed no change: it offers each
   cast of the card in the prompt's `DiscardOptions` through `castMovesFromZone`, and the decline. The bot's heuristic answers it
   as any cast: it casts the free spell (`aiseat/resolution_keyword_cast_bot_test.go`).

### Limits, stated

- Casting a card from hand during a resolution (#2110, `offerFreeHandCast` and the other free-cast-from-hand grants) and the
  other impulse-style grants (Chandra, Torch of Defiance's paid cast, the free-or-hand offers) keep the grant stand-in. They are
  separate doors with their own prompts and are not changed here (#4250).
- The 2026-10-22 "Limits, stated" bullet that listed discover, cascade, cipher and paradigm among the doors that keep the
  stand-in no longer holds for those four.
- `mayCast.ts`'s cascade and discover wording is now unreachable from a live table (no `may_cast` is asked under either keyword);
  it stays for a restore of an old table.

### Snapshot

One additive field, `pendingChoices[].castPile` (`omitempty`), recorded with `-update-shape`. A binary from before it reads a
cascade or discover prompt without the pile and so leaves the passed-over cards in exile when the prompt is answered: weaker than
printed, never stronger, and only for a table rolled back across this change while one of those prompts is open. The frozen corpus
is untouched.

Tests: `game/cascade_test.go`, `game/discover_test.go` and `game/paradigm_test.go` (the cast before the resolution finishes,
the decline, the pile and the hit bottomed together, the fallback window), `cards/effects/resolution_keyword_cast_test.go` (a
real cascade card, Hit the Mother Lode, and a cipher card: the cast before the trigger finishes, the decline, a restore
mid-prompt for each, and the Treasures after a restored answer), `legal/resolution_cast_test.go`,
`protocol/resolution_cast_view_test.go`, `aiseat/resolution_keyword_cast_bot_test.go` and
`client/src/lib/resolutionCast.render.test.ts`. Back-outs: drop the `queueResolutionCastOfferLocked` call in
`CascadeForEffect` and `TestCascadeCastsTheHitBeforeTheTriggerFinishes` fails; drop `DuringResolution` in `castCipherCopyLocked`
and `TestCipherCopyIsCastBeforeTheTriggerFinishes` fails, and in `paradigmCopy` `TestParadigmCopyIsCastBeforeTheTriggerFinishes`
fails; drop the `lapse` in `withdrawCardCopyOffersLocked` and `TestDiscoverDeclinePutsTheCardIntoHand` fails; drop `CastPile` from
the snapshot capture and `TestCascadeOfferSurvivesARestoreMidPrompt` fails.


## Amendment — 2026-10-23 (#4250): rebound, suspend, madness and the impulse-style offers cast during their resolution

The 2026-10-23 #4243 amendment moved cipher, cascade, discover and paradigm onto the `resolution_cast` prompt and left the
other "you may cast it" doors that a resolution offers on the grant stand-in: a yes/no `may_cast`, then a permission the holder
used with an ordinary `cast_spell` after the resolution had finished, closing on their next priority pass. On another player's
turn the active player received priority first (CR 117.3b) and the table could act before the cast. The rules put each of these
casts inside the resolution:

- CR 702.88a: rebound's delayed trigger, "you may cast this card from exile without paying its mana cost".
- CR 702.62a: suspend's third ability, "you may play it without paying its mana cost if able. If you don't, it remains exiled."
- CR 702.35a: madness's trigger, "its owner may cast it by paying [cost] rather than paying its mana cost. If that player
  doesn't, they put this card into their graveyard."
- CR 608.2g for the effects that print the cast: Tibalt's Trickery's and Guff Rewrites History's exile-until walk, Chandra,
  Torch of Defiance's paid "you may cast that card. If you don't, ...", and the free-or-hand offers (Ziatora's Envoy and
  Rashmi, Eternities Crafter from the top of the library, Beseech the Mirror from exile). CR 608.2g: "If an effect specifically
  instructs or allows a player to cast a spell during resolution, they do so by following the steps in rules 601.2a–i, except
  no player receives priority after it's cast."

### Decision

1. **Seven doors ask the `resolution_cast` prompt**, each the #4243 shape: stamp the permission the `may_cast`'s "yes" used to
   stamp, then `queueResolutionCastOfferLocked` over the one card, naming the keyword and the two answers. The doors are
   `reboundCast` (`rebound`), `offerSuspendedCastLocked` (`suspend`), `offerMadnessCastLocked` (`madness`),
   `exileUntilCastFreeLocked` (`exile_until`, both the bottoming and the keeping form), `OfferPaidCastFromExileForEffect`
   (`paid_or_else`), the spell branch of `OfferFreePlayFromTopOrHandForEffect` and `CastExiledFreeOrToHandForEffect` (both
   `free_or_hand`). The `may_cast` yes/no is gone for all of them; a cast through the prompt is "yes" and the empty
   `resolve_choice` is "no". The prompt's cast is the existing one: strict payment with the auto-tapper (a madness cost and a
   paid-or-else card's own cost are paid there, CR 601.2g and CR 608.2g's "option to pay mana"), no priority after it.
2. **Each decline is the door's own**, and it is the permission's lapse, as #4243's is (`withdrawCardCopyOffersLocked` lapses a
   pass-closed permission it withdraws):
   - rebound, the keeping exile-until form and paid-or-else: `LapseStaysInExile`, so the card stays in exile; paid-or-else's
     `IfNotCast` follow-up ("if you don't") runs from the lapse, inside the resolution.
   - suspend: its permission was never pass-closed (it held for the turn). Withdrawn with no lapse, so the card stays exiled with
     no counters, CR 702.62a's "it remains exiled".
   - madness: a new lapse, `LapseToGraveyard` (on-disk value `"graveyard"`, additive), puts the card into its owner's graveyard
     through the exit primitive (`putIntoOwnersGraveyardLocked`); it is not a discard. The madness permission now carries it, so
     the fallback below is pass-closed too. The end-step reclaim (`scheduleMadnessGraveyardLocked`) is scheduled only for that
     fallback.
   - the bottoming exile-until form: `LapseToLibraryBottom`, and the cards the walk passed over ride the prompt as
     `PendingChoice.CastPile` (#4243's field), so the hit joins them in one random order, as cascade's does.
   - free-or-hand: `LapseToHand`. For Ziatora's Envoy the card is still on top of the library; the lapse reads the permission's
     zone, as the pass-closed window already did.
3. **A library card is a candidate.** `ResolutionCastCandidateLocked` now accepts a card in a library as well as in exile or a
   graveyard, when one of the chooser's stored `ScopeCards` permissions names it; only Ziatora's Envoy's grant opens a library
   card that way. The enumerator needed no change (`castMovesFromZone` already walks a library card under a permission). The
   view needed none either: a resolution_cast's `options` go through the same knower filter as every prompt's cards, so a
   looked-at top card is the chooser's alone, and the prompt's `reason` names the source, not the card.
4. **Restore.** The prompt, the permission and the pile are data, and the closures are gone from these doors (each `may_cast`
   carried an `OnAccept` and `OnDecline` frame, a census blocker). A table paused on any of them is a restore point and resumes.
   `closure_fields.txt` is unchanged: `mayCastFrame` is still used by the land play and the card-text offers.
5. **The fallback.** Outside an open resolution (no catalog card reaches one; every one of these runs as a trigger, a spell or
   an ability resolves) there is nothing to hold open, and the permission alone is the window, closing on the holder's next pass
   as before. The tests drive it directly.
6. **Wire and client.** The prompt carries `may_cast_keyword`, `accept_label` and `decline_label` as #4243's does.
   `ResolutionCastSheet` words every keyword offer through `keywordCastWords` (`mayCast.ts`): the rule tag, the price ("without
   paying its mana cost", "for its madness cost", "paying its mana cost") and what the decline does. The `may_cast` wording for
   these keywords, and `mayCastKeywordsThatOpenACast` with the modal's "open the cast chain on yes" branch, are gone: no
   `may_cast` is asked under any of them, and a `may_cast` was never a restore point, so no restored table can hold one.

### Cards that changed behaviour

Every rebound card (Staggershock, Ephemerate, Quantum Misalignment, Distortion Strike, World at War, ...), every suspend card
(Rift Bolt, Ancestral Vision, Lotus Bloom, Nihilith, Judoon Enforcers, ...), every madness card (Fiery Temper, Big Game Hunter,
Avacyn's Judgment, Shadowgrange Archfiend, ...), Tibalt's Trickery, Guff Rewrites History, Chandra, Torch of Defiance, Ziatora's
Envoy, Rashmi, Eternities Crafter and Beseech the Mirror. No card's `Completeness` or `Caveats` changed: the change removes a
response window paper does not have.

### Limits, stated

- Casting from hand during a resolution (`offerFreeHandCast`, `lookAtTopAndOfferFreeCast`: Buster Sword, Kari Zev's Expertise,
  Rishkar's Expertise, Surtland Elementalist, Great Intelligence's Plan, Cecily, Aetherworks Marvel, Aetherflux Conduit) keeps
  the keyed pick (#4246) and the pass-closed grant. A hand card on a `resolution_cast` needs the cast held to the permission's
  claim, since a hand card also has its ordinary cast. It is the next PR of #4250.
- The one-off pass-closed grants in `cards/effects/` (Torrential Gearhulk, Arcane Heist, Memory Plunder, Strago and Relm,
  Crabomination, Jin-Gitaxias's chapter III, Fevered Suspicion, Kefka, the Sieges, ...) keep the stand-in. #4250 lists each with
  whether its Oracle text casts during the resolution or opens a real window.
- Ziatora's Envoy's land play stays a `may_cast` (`play_or_hand`): it is a land played during the resolution, already made
  inline when the answer comes in, not a cast.
- The #4243 "Limits, stated" bullet that named Chandra's paid cast and the free-or-hand offers among the doors that keep the
  stand-in no longer holds for them.

### Snapshot

Nothing new in the shape. `PermissionLapse` gains the value `"graveyard"`, written only by a madness permission made after this
change; a binary from before it reads an unknown lapse as no lapse, so on a table rolled back across this change while a madness
prompt is open, a declined card stays in exile rather than going to the graveyard: weaker than printed, never stronger. Two
corpus boards build a different history now (the cast is logged before its trigger's `resolution_finished`), so their files
stay as written and are still restored, and the boards are registered under new names with new files
(`rebound_free_cast_offered_during_resolution.json`, `suspend_haste_cast_during_resolution.json`), as #2494 did.

Tests: `game/rebound_test.go`, `game/suspend_test.go`, `game/madness_test.go`, `game/exile_until_cast_test.go` and
`game/cast_paid_or_else_test.go` (the cast before the resolution finishes, each decline, the fallback window, a restore
mid-prompt for rebound, madness and paid-or-else); `cards/effects` tests for Tibalt's Trickery, Chandra, Torch of Defiance,
Beseech the Mirror and Ziatora's Envoy (each with a restore mid-prompt), and the rebound, suspend, madness, Guff and Rashmi
card tests moved onto the prompt; `legal/rebound_cast_moves_test.go` and `legal/free_play_from_top_test.go`;
`protocol/free_play_from_top_view_test.go`; `aiseat/resolution_door_cast_bot_test.go` (rebound, madness, the top-of-library
cast) and `aiseat/paid_cast_or_else_bot_test.go`; `client/src/lib/mayCast.test.ts` and `resolutionCast.render.test.ts`.
Back-outs: drop the `queueResolutionCastOfferLocked` call in `reboundCast` and `TestReboundOffersAFreeCastAtYourNextUpkeep`
fails, in `offerSuspendedCastLocked` and `TestTheLastTickOffersTheFreeCast` fails, in `OfferPaidCastFromExileForEffect` and
`TestPaidCastOrElseAsksOverTheExiledCardAndDeclineRunsIfNotCast` fails, in `CastExiledFreeOrToHandForEffect` and
`TestBeseechTheMirrorAcceptedGrantLapsesToHand` fails, in `OfferFreePlayFromTopOrHandForEffect` and
`TestZiatorasEnvoyCastsTheTopSpellFromTheLibrary` fails; drop the madness permission's `LapseToGraveyard` and
`TestDecliningMadnessPutsTheCardInTheGraveyard` fails; drop the exile-until offer's `Pile` and
`TestExileUntilCastCastsTheHitBeforeTheResolutionFinishes` fails; drop `ZoneLibrary` from `ResolutionCastCandidateLocked` and
`TestEnumeratorOffersAFreeCastFromTheLookedAtTop` fails.


## Amendment — 2026-10-24 (#4250, part 2): the hand-cast offers and the one-off grants cast during their resolution

The 2026-10-23 #4250 amendment left two groups on the pass-closed stand-in: the "cast a spell from your hand without paying its
mana cost" offers (a keyed pick, #4246, then a pass-closed grant) and the one-off pass-closed grants that card code in
`cards/effects/` stamps from a resolution. Each prints a cast made while its spell or ability resolves (CR 608.2g: "If an effect
specifically instructs or allows a player to cast a spell during resolution, they do so by following the steps in rules
601.2a–i, except no player receives priority after it's cast"). Each now asks the `resolution_cast` prompt inside that resolution.

### Decision

1. **One door for card code.** `Game.OfferResolutionCastForEffect(ResolutionCastOffer)` raises the prompt over cards a stored
   `ScopeCards` permission of the chooser's names. With `Permission` set it first stamps that permission over all the cards as
   ONE permission; a pass-closed one is spent whole by the first cast (`consumePassClosedGrantLocked`), which is "cast a spell
   from among them". Per-card permissions, or an `AnyNumber` or budget permission, leave the prompt open after each cast (CR
   707.12a's reading, as #2947 does for copies), until the chooser declines or none is left. A land is never a candidate (CR
   116.2a and CR 305.9: a land is played, never cast). The door answers false, and stamps nothing, outside an open resolution; card code then
   keeps its old offer, which no catalog card reaches. `effects.offerResolutionCastOver` is the card-side wrapper.
2. **A hand card is a candidate, held to its claim.** `ResolutionCastCandidateLocked` accepts a card in the chooser's hand that a
   stored permission names. A hand card also has its ordinary cast (CR 601.2), and a hand permission prices only the cast that
   claims its offer (`CastPermission.ForClaim`, #1665), so `castForResolutionChoiceLocked` names the permission's
   `AltCostKey` (`ResolutionCastClaimLocked`; `granted-cost` for the free offers, `great-synthesis-free` for chapter III)
   whatever the cast sent, and `legal/resolution_cast.go` keeps only the casts that claim it. The view needed no change: a hand
   card on the prompt goes through the knower filter, so the table sees that a seat owes an answer and not which cards.
3. **"And then", after the casts.** The prompt carries an optional continuation, a registered `ChooseCardsThen` key in
   `PendingChoice.PickThen` with the caller's cards in `OptionCarry` (both existing, persisted fields). `finishResolutionCastLocked`
   runs it once the prompt is answered, by the last cast or the decline, with the cards of the carry now on the stack as its
   picks (nobody received priority since, so nothing else moved them). Three continuations: The Great Synthesis's "Exile this
   Saga, then return it to the battlefield" (`resolution-cast/great-synthesis-flicker`), Kefka's "Then each player who owns a
   spell you cast this way loses life equal to its mana value" (`resolution-cast/kefka-life-loss`), and "put the rest on the
   bottom of your library in a random order" for Aetherworks Marvel and Sunbird's Invocation
   (`resolution-cast/looked-at-rest-to-bottom`). A restore refuses a file whose resolution_cast names a key this binary lacks.
4. **The doors moved.** The hand offers: `offerFreeHandCastOver` (Buster Sword, Kari Zev's Expertise, Rishkar's Expertise,
   Surtland Elementalist, Great Intelligence's Plan, Cecily, Haunted Mage, the Mad Wizard's Lair room) and the "any number" form
   (Aetherflux Conduit, The Great Synthesis's chapter III), keyword `hand_cast`, decline "Keep it in your hand"; the library look
   (Aetherworks Marvel, Sunbird's Invocation), keyword `free_cast`, one permission over the eligible cards where they sit, decline
   "Put them on the bottom". The one-off grants: `grantFreeCasts` when it closes on a pass (Emergent Ultimatum, Ensnared by the
   Mara, Etali, Primal Conqueror, The Key to the Vault, Cosmium Catalyst), `OfferGraveyardCast` (Emet-Selch of the Third Seat,
   paid), Torrential Gearhulk, Arcane Heist, Memory Plunder, Strago and Relm, Crabomination, `exileUntilNonlandGrantingCast` (Amped
   Raptor's energy and The Infamous Cruelclaw's discard, `paid_cast`), Fevered Suspicion, Kefka, Dancing Mad, the budgeted
   `ExileTopWithPermission` (Hazoret's Undying Fury, Primeval Spawn), Conduit of Worlds (paid), Goliath Daydreamer, Power Pack's
   delayed upkeep trigger, Summon: Esper Valigarmanda's chapters (paid, with the mana the chapter adds) and the Sieges
   (`SiegeDefeated`, the back face cast from exile). Uldaros Theorix's casts were already on the prompt (#2947's copies).
5. **What stays.** Nicol Bolas, God-Pharaoh's +2 ("Until end of turn, you may cast that card") opens a real window and keeps it:
   `grantFreeCasts` with no lapse asks no prompt. The keyed picks #4246 registered (`choose-cards/<card>-cast`) stay registered,
   so a table saved while waiting on one still restores and answers; they are also the hand and library offers outside an open
   resolution. Kefka's per-cast `kefka-dancing-mad-life-loss` follow-up stays registered and is still the life loss outside one.
6. **Wire and client.** Three new `may_cast_keyword` values, `hand_cast`, `free_cast` and `paid_cast` (on-disk, additive), worded
   by `keywordCastWords` (`mayCast.ts`): free, or "paying its cost", and where a declined card stays, read off the decline
   label. With more than one card on offer the sheet's heading says "a spell from your hand" or "a spell from among them".

### Cards that changed behaviour

Every card in decision 4. The cast is made inside the resolution: nobody gets priority between the offer and the cast, and a
"then" waits for the casts. Two fixes ride on that order. The Great Synthesis is exiled and returned AFTER its free casts, so the
returned Jin-Gitaxias no longer draws for spells cast through chapter III (it used to be on the battlefield first, which was
stronger than printed). Kefka's life loss now happens after every cast rather than with each. No card's `Completeness` or
`Caveats` changed.

### Snapshot

Nothing new in the shape: the prompt's continuation and carry are the existing `pickThen` and `optionCarry`, and the keywords are
new values of the existing `mayCastKeyword`. A binary from before this change reads a resolution_cast's `pickThen` as a
revealed-hand pick's continuation and refuses the file (unknown key), so a rollback across this change with one of the three
continuations open fails loudly rather than skipping the rest of a resolution. Tables saved on the old keyed picks restore and
answer as before.

Tests: `cards/effects/resolution_cast_one_offs_4250_test.go` (each one-off card and door: the cast logged before the resolution
finishes, the decline leaving the card where it was, a strict restore mid-prompt and the cast on the restored table);
`free_hand_cast_3270_test.go`, `free_cast_keyed_restore_test.go` (the hand and library offers mid-prompt and restored, the old
keyed pick restored and answered, an unknown continuation refused), `aether_free_casts_test.go`, `raffine_a_praetors_test.go`
(chapter III casts before the flicker, and Jin-Gitaxias draws nothing for them), `kuja_a_kefka_test.go` (the life loss waits for
every cast; a restore mid-prompt) and the moved card tests; `protocol/hand_cast_offer_view_test.go`;
`aiseat/free_hand_cast_bot_test.go` (only the claimed casts are offered for a hand card) and
`aiseat/one_off_resolution_cast_bot_test.go` (graveyard, exile with a continuation, library); `client/src/lib/mayCast.test.ts`
and `resolutionCast.render.test.ts`. Back-outs are listed in the PR.

### Note — 2026-10-24 (#4286): Guile's play is made during the counterspell's resolution

Decision 4 left one pass-closed grant out: Guile's "If a spell or ability you control would counter a spell, instead exile that
spell and you may play that card without paying its mana cost" (ADR 0113's 2026-10-19 amendment, `finishCounterSpellLocked` in
`game/counter_spell.go`). It stamped `FreeCastFromExilePermission` and returned, so the play was an ordinary `cast_spell` taken
after the counterspell had finished, and it closed on the player's next pass. The play is made while the countering spell or
ability resolves (CR 608.2g), so it now takes the door every other resolution-time cast takes.

- **The change.** Once the spell has landed in exile, the finisher calls `OfferResolutionCastForEffect` with that permission as
  `Permission`, keyword `free_cast`, accept "Cast it free" and decline "Leave it in exile", over the one exiled card, for the
  countering player (never the spell's owner). The cast is made before the counterspell finishes; a decline leaves the card in
  exile and the counterspell finishes. Nothing changed in the rules the permission carries: free, cast-only, X is 0 (CR 107.3b),
  and offered only if the card really landed in exile. No wire change (`free_cast` is decision 6's value) and no snapshot change
  (the prompt is the existing `resolution_cast`).
- **Every counter runs inside an open resolution.** `counterSpellByEffectLocked` opens the window only when a resolving spell or
  ability is the cause, and every counter in the catalog reaches it there: a hard counter (Counterspell); a declined
  counter-unless-pays (Spell Pierce, and the any-player offers: the prompt guards the countering spell, so it holds the
  resolution open until the answer); a ward (the warded permanent's controller's triggered ability, whose counter runs on the
  payer's decline); and the destination counters (Hinder's top or bottom, Devious Cover-Up's exile, and Remand's hand and
  Memory Lapse's library, which are the same two verbs). A CR 616.1 ordering prompt put to the countered spell's controller holds
  the resolution open too, so the resume that finishes the counter offers the cast inside it.
- **What stays.** If the door answers false, which is a counter whose cause was carried past the end of its resolution (a
  non-blocking pay-unless answered after the table moved on, `carriedCause`, #4279), the replaced counter still grants the
  pass-closed permission alone, as before. No catalog counter reaches it today (checked: every counter-unless-pays prompt
  guards its stack item and so holds the resolution open); it stays so that no future counter can lose the free play.
- **A commander** (ADR 0115) lands in exile at once and is offered like any card; if its owner then takes the command zone
  (CR 903.9a), the countering player's permission no longer opens anything.

Tests: `game/counter_spell_resolution_cast_4286_test.go` (each counter verb, including the hand, exile and both library
destinations, asks the prompt for the countering player alone; a decline leaves the card exiled with no permission; X is locked at
0; no open resolution keeps the pass-closed grant; a commander); `cards/effects/guile_resolution_cast_4286_test.go` (Counterspell,
a declined Spell Pierce, a ward, Hinder to the top and to the bottom, Devious Cover-Up: the cast is logged before the countering
resolution finishes, a decline leaves the card in exile, a strict restore mid-prompt keeps the offer and the cast is made on the
restored table) and the updated `guile_test.go`; `aiseat/guile_resolution_cast_bot_test.go` (the bot answers for a hard counter,
a declined Pierce, a ward and Hinder). Back-outs: asking no prompt in `finishCounterSpellLocked` fails every one of those;
dropping the pass-closed fallback fails `TestAReplacedCounterWithNoResolutionOpenKeepsThePassClosedGrant`.

## Amendment — 2026-10-30 (#4615): `PermissionFilter.ColorsAny`, an any-of colour clause

World War Hulk's chapter I ("The next red or green creature spell you cast this turn can be cast without paying its mana
cost") is Apex Observatory's free-cast promise with a colour on the spell. The filter had a card-type clause and no colour one.

- **The clause.** `PermissionFilter.ColorsAny` is a string of colour letters in the card's own vocabulary (`"RG"`): a card
  qualifies when it has at least one of them (CR 105.2b: a multicoloured object is each of its colours). It is read through
  `Card.HasColor`, the card's effective colours, so a spell is judged by its characteristics as it is cast; a colourless
  card has no colour and never matches (CR 105.2c). It is a further AND beside every other clause, so `CreatureOnly` plus `"RG"`
  is "red or green creature spell", and empty constrains nothing. It is a string rather than the two-slot array
  `CreatureTypesAny` uses because the filter must stay comparable and a one-field string writes nothing when empty (`omitempty`),
  where an array is always written; it also covers a third colour with no widening.
- **Who sees it.** The clause lives in `PermissionFilter.Matches`, the one predicate every consumer already calls: the promise's
  flash, price and free-cast reads (`NextSpellPromise.Filter`), the granted-cost offers the view and `CastSpell` read, and the
  enumerator the bot plays from. No consumer was changed, and `legal/coloured_next_spell_promise_test.go` pins the move list.
- **Snapshot.** Additive: `colorsAny` (omitempty) under every path that reaches a `PermissionFilter`, recorded in
  `testdata/snapshot_shape/v9.txt` with `-update-shape`. An older binary drops the key and reads the filter without the colour clause, which
  widens a restored promise to every creature spell for the rest of that turn; promises end at end of turn.
- **Cards.** World War Hulk (Full). Tests: `game/permission_filter_colors_test.go`, `cards/effects/world_war_hulk_test.go`,
  `legal/coloured_next_spell_promise_test.go`. Back-out: dropping the `ColorsAny` line in `Matches` fails all three.
