# ADR 0066 — Granted cast and play permissions

**Status:** accepted
**Date:** 2026-09-18
**Sprint:** S42 — Casting from non-hand zones: granted permissions and alternative costs
**Issues:** [#652](https://github.com/krakenhavoc/cmd_and_ctrl/issues/652) (granted cast permissions), [#765](https://github.com/krakenhavoc/cmd_and_ctrl/issues/765) (play and cast from the top of your library). Tracker [#885](https://github.com/krakenhavoc/cmd_and_ctrl/issues/885).

**Numbering:** every remote branch was swept with the AGENTS.md §4 loop on
2026-09-18 (`git ls-remote --heads origin`, 266 heads; `git log --all
--diff-filter=A --name-only -- docs/decisions/` over the fetched refs).
`0060` is on `develop`; `0061` (token creation and discard),
`0062` (abilities and special actions from the hand), `0063` (durations and
control) and `0064` (emblems) are held by in-flight branches, and `0065` is
claimed by a fifth parallel agent that has not pushed. `0066` is the first
free number. `0005`, `0024`, `0029` and `0030` stay permanently unused.

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
| `Player.CastPermissions []CastPermission` | `ScopeCards` | `UntilTurn` (Snapcaster, Past in Flames, impulse exile) or `WhileInZone` (airbend, warp) |
| derived from the battlefield through `CardDef.CastPermissions` | `ScopeStanding` | for as long as the source permanent remains — nothing to expire |
| — | — | — |

Standing permissions are **derived, never stored**, for the reason
`land_drops.go` already gives about `CatalogAdditionalLandPlays`: two
Underworld Breaches have to compose, and one of them leaving must not revoke a
permission the other is still granting. Deriving them per query is also the
whole of "for as long as the source remains" — there is no expiry code to get
wrong, and a Breach that is exiled in response to the cast stops granting
before the cast is validated, which is what CR 702.138 says.

**The duration model, and a follow-up this ADR owes.** #755/#756's
`game.Duration` (`UntilEndOfTurn`, `UntilYourNextTurn`, `ForAsLongAs`,
`Indefinite`, ADR 0063) is the project's one duration vocabulary. It landed
on `develop` *while this work was in flight*, and the permission's window is
still the `{UntilTurn, NotBeforeTurn, WhileInZone}` triple `ExilePlayPermission`
carried. That is a second vocabulary, it is not meant to stay, and it is
recorded here rather than left to be discovered.

Two things make it a contained debt. The stored permissions need exactly two
of `Duration`'s four kinds — "until end of turn" (Snapcaster, Past in Flames,
impulse exile) and "while the card stays in the zone" (airbend, warp) — and
the standing ones need none at all, because deriving them from the battlefield
IS `ForAsLongAs`. And `CastPermission.Active` is the single function that reads
the triple, so the swap is that function plus its call sites.

It is not free, which is why it is a follow-up rather than a line in this PR:
`Duration` expiry needs the game (`durationExpiredLocked`), so `Active(player,
turn)` becomes a method on `*Game` and every caller — including
`faceForCastLocked` and about a hundred test literals — moves with it. The
follow-up is tracked; `NotBeforeTurn` survives it either way, because warp's
"on a later turn" is a FLOOR and `Duration` has no concept of one.

### 2. CR 400.7 is an object-identity check, not a sweep

A grant ends when its card leaves the zone **by any route**, not only when it
is cast. Before this ADR that was enforced by zeroing `Card.ExilePlay` at
every exit — in `MoveCard`, in both cast branches, in the entry paths. That
discipline works only as long as every author remembers it, and moving the
store off the card would have multiplied the sites rather than removed them.

So identity is stamped instead: `Card.ObjectEpoch` is an integer `MoveCard`
increments on every zone change, and `PermissionCardRef` is `{ID, Epoch}`. A
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
  mechanism and is not built.
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

- The permission's window is still its own three fields rather than
  `game.Duration`, which is a second duration vocabulary for as long as the
  follow-up is open. It is the one thing in this ADR that is not yet the
  single model the rest of it argues for.

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
