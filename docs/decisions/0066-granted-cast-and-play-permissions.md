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
offer leaves no cast at all.

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
