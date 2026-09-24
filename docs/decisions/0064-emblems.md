# ADR 0064 — Emblems (CR 114)

**Status:** Accepted · 2026-09-18 · S40 — S30's tail: protection, regeneration, emblems, winning by effect
**Issue:** [#623](https://github.com/krakenhavoc/cmd_and_ctrl/issues/623)
**Numbering:** on 2026-09-18, after `git fetch origin` (261 remote heads),
every `docs/decisions/` file ever added on any ref was listed with
`git log --all --diff-filter=A --name-only -- 'docs/decisions/*'`. The highest
number claimed anywhere is **0063**; 0061, 0062 and 0063 are on parallel
branches not yet merged to `develop`. 0064 is the first free number.
0052 was reserved for this ADR in the headers of
[0053](0053-combat-damage-beats.md), [0054](0054-dice-rolls-and-coin-flips.md),
[0056](0056-infect-wither-toxic.md), [0057](0057-win-and-lose-by-effect.md),
[0058](0058-doesnt-untap.md) and [0059](0059-turn-machinery.md); the owner
reassigned 0052 to the bot decision harness before this was written, so the
reservation is void and those headers are stale prose, not links. 0005, 0024,
0029 and 0030 stay permanently unused per AGENTS.md §4.
**Related:** [ADR 0032](0032-planeswalkers.md) (omitted ultimates stay
omitted), [ADR 0012](0012-layer-system.md) (the layer pass that finds
statics), [ADR 0018](0018-triggers-on-the-stack.md) and
[ADR 0026](0026-delayed-triggers.md) (the harvester that finds triggers),
[ADR 0028](0028-admin-context-menu.md) (the admin move verb an emblem must
refuse), [ADR 0041](0041-game-persistence.md) (snapshot schema policy),
[ADR 0060](0060-leaving-the-game.md) (CR 800.4a — a departed player's objects
leave with them)

## Context

An emblem is the one object in Magic that exists only in the command zone,
has no characteristics at all, and cannot be interacted with (CR 114). Four
catalogued or wanted cards have waited on it since S27: Elspeth, Sun's
Champion's −7, Wrenn and Six's −7, Teferi, Hero of Dominaria's −8, and the
back face of Invasion of New Phyrexia ([#626](https://github.com/krakenhavoc/cmd_and_ctrl/issues/626)).
Per ADR 0032 each of the three registered planeswalkers omits its ultimate
rather than stubbing it, and says so in `Caveats`; `docs/engine-seams.md`
carries the **Emblems** row.

**The rules** (Comprehensive Rules effective August 7 2026):

- **CR 114.1** — "An emblem is a marker that represents an object that has
  no characteristics other than the abilities defined by the effect that
  created it." No name, no types, no mana cost, no power or toughness.
- **CR 114.2** — "Emblems exist in the command zone."
- **CR 114.3** — "An emblem has no characteristics other than the abilities
  defined by the effect that created it. … Abilities of emblems function in
  the command zone."
- **CR 114.4** — "An emblem is neither a card nor a permanent. Emblem isn't
  a card type."
- **CR 114.5** — "An effect that creates an emblem is written 'You get an
  emblem with [ability]'. … The emblem is put into its owner's command
  zone." Its owner and controller are the player the effect names.
- **CR 800.4a** — when a player leaves the game, every object they own
  leaves with them. An emblem is owned, so it goes.

Two engine facts framed the decision, and both were measured rather than
assumed:

1. **The layer pass gathers statics from exactly two places.**
   `activeStaticAbilitiesLocked` (`server/internal/game/layers.go`) walks
   `g.Battlefield` through `CatalogStaticAbilities(CatalogKey(card))`, and
   prepends the S32 turn-scoped registry. `StaticAbility.AppliesTo` and
   `.Apply` both take a `source *Card`, and every "creatures you control"
   predicate in the catalog reads `source.Controller`. So a static needs a
   `*Card` to hang on — an emblem that is not `*Card`-shaped forces either a
   second `ContinuousEffect` adapter or a nil-source special case in every
   predicate.
2. **The trigger harvester is already zone-parameterised.**
   `triggerHarvester.OnEvent` calls `g.harvestFromZone(&pass, g.Battlefield)`
   and `harvestFromZone(pass *harvestPass, z *Zone)` takes any `*Zone`, reads
   abilities through `CatalogTriggers(CatalogAbilityKey(card))`, and builds
   the item with `NewTriggeredItem(source, …)`. A `*Zone` of emblem `Card`s is
   one more argument to a function that already exists.

Those two shapes point at the same answer, which is why this ADR is short on
alternatives: an emblem that is a `Card` in a `Zone` reuses the layer pass,
the harvester, LKI, the clone, the snapshot and the wire, and needs no second
registry and no second walk.

## Decisions

### 1. An emblem is a `game.Card` in a per-player command-zone slice

An emblem is created as an ordinary `game.Card` value:

```go
Card{
    InstanceID:           uuid.New(),
    Name:                 "Elspeth, Sun's Champion emblem",
    OracleID:             "emblem:05e6b243-48a6-4a42-bc5f-413441de9c33",
    TypeLine:             "",          // CR 114.1: no types
    Owner:                owner,
    Controller:           owner,       // CR 114.5
    EnteredBattlefieldAt: <stamp>,     // CR 613.7c timestamp
}
```

and pushed onto a new per-player zone, `Player.Emblems`, built as
`newZone(ZoneCommand, playerID)`. **It is the command zone** (CR 114.2). The
engine splits the command zone's storage into two slices — `Player.Command`,
the commander pile, and `Player.Emblems` — and that split is the load-bearing
part of this decision.

**Why not one slice.** Every existing reader of `Command.Cards` means
"commander cards", and the audit found six places where an emblem sharing
that slice would be actively wrong and one where it would be deleted:

| site | what it would do to an emblem |
|---|---|
| `legal/cast.go:50` | offers every command-zone card as castable, unfiltered |
| `game/cast_zones.go:183` | `case ZoneHand, ZoneCommand: return nil` — unconditional cast permission |
| `protocol/view.go:1633` | stamps modes / costs / alt-cost offers on every command-zone card |
| `game/mutations.go:1033`, `:1472`, `legal/cast.go:137`, `lobby/http.go:1005` | charge and increment commander tax on a `from_zone=="command"` cast |
| `aiseat/heuristic/heuristic.go:339`, `aiseat/model/prompt.go:189` | fold it into "cards I can play" |
| `actions/actions.go:488` → `game/mutations.go:3670` | admin `move_card` pulls it out, no `IsCommander` gate |
| `game/game.go:598` | `ReplaceDeck` truncates `p.Command.Cards[:0]` |
| **`game/token_existence.go:74`** | the CR 704.5d state-based action **deletes** anything token-shaped in the command zone |

Filtering all eight is eight chances to miss one, and the one that is missed
is a bug that only shows up mid-game. A second slice is exclusion **by
construction**: `ZoneRef` is `{Kind, Owner}` with no discriminator, so
`zoneFromRefLocked`, `routeDestinationLocked`, `castSourceZoneLocked`,
`findCardZoneLocked` and `findCardByIDLocked` all resolve `{command, owner}`
to `p.Command` and can never reach an emblem. That is not an accident we are
tolerating — it is CR 114 enforced by the type of the container: **an emblem
cannot be cast, moved, targeted, sacrificed or destroyed, because no verb in
the engine can name it.**

**Why not player-scoped statics with no object.** The rejected alternative —
a `Player.Emblems []EmblemRecord` with no `Card` — needs a second
`ContinuousEffect` adapter (nothing to point `source *Card` at), a second
trigger walk (`harvestFromZone` takes a `*Zone` of `Card`), a second LKI
story, its own snapshot mirror, and its own clone. It buys nothing: the
`Card` struct is a bag of zero values for an object with no characteristics,
and the zero values are exactly what CR 114.1 asks for.

### 2. Identity is a synthetic catalog key, not a new `Card` field

The emblem's `OracleID` is `"emblem:" + <the creating card's CatalogKey>` —
`game.EmblemKeyPrefix` and `game.EmblemKey(sourceKey)`. Because `CatalogKey`
returns `c.OracleID` for a face-0 card, the emblem's catalog key **is** its
oracle ID, so `CatalogLookup`, `CatalogStaticAbilities`, `CatalogTriggers`
and `CatalogAbilityKey` all work on an emblem with no change at all.
`Card.IsEmblem()` is a prefix test on that key.

No field is added to `game.Card`. Three reasons: a real Scryfall oracle ID is
a UUID and can never collide with the prefix; `TestCardHasNoInteriorPadding`
polices `Card`'s bool block, so a new flag is a layout change; and the key
**is** the identity — CR 114.1 says an emblem has no characteristics other
than the abilities the effect defined, and the key is precisely the handle on
those abilities. Deriving the same fact twice (a flag and a key) is how the
two disagree later.

Using the creating card's key as the suffix means the emblem is looked up
where the card that makes it is written, and a multi-face card gets the right
one for free: Invasion of New Phyrexia's back face registers under
`"<oracle>#1"`, so its emblem is `"emblem:<oracle>#1"`.

`TypeLine` stays **empty**. CR 114.1 says no types, and it also keeps the
emblem clear of `IsToken()` — which is a `TypeLine` substring test, and whose
CR 704.5d state-based action sweeps the command zone.

### 3. The abilities come from the catalog, through one new `Spec` slot

A card that makes an emblem declares it once, next to the ability that
creates it (AGENTS.md §7 "Adding a `Spec` slot"):

```go
Register(Spec{
    OracleID: "05e6b243-…",
    Name:     "Elspeth, Sun's Champion",
    Emblem: &EmblemSpec{
        Label: "Elspeth, Sun's Champion emblem",
        Text:  "Creatures you control get +2/+2 and have flying.",
        Static: []game.StaticAbility{ … },
    },
    Activated: []ActivatedAbility{{
        Label:  "−7: You get an emblem with …",
        Cost:   LoyaltyCost(-7),
        Effect: func(g *game.Game, item *game.StackItem) error {
            return CreateEmblem{}.Apply(NewContext(g, item))
        },
    }},
})
```

`effects.Register` builds a **second** `game.CardDef` from `Spec.Emblem` and
files it in the existing `defs` map under `game.EmblemKey(spec.OracleID)`.
The emblem's def carries the emblem's `Static`, its `Triggered`, and one new
`CardDef.Emblem *EmblemDef{Label, Text}` slot holding the board label and the
oracle text.

It goes into `defs` and **not** into `registry`, which is what
`effects.All()` returns. That keeps the card-coverage census counting cards:
an emblem is not a card (CR 114.4) and must not appear in `Keys`,
`WholeCards` or the completeness split.

**Stored data, never stored closures.** The emblem object on the `Player`
carries a key, not a `StaticAbility`. A game holding a closure cannot be
restored from a snapshot, which is the reason the audit gave for rejecting
the record-with-abilities shape, and the reason `ScopedStatic`'s
`ContinuationCensus` entry exists. Here there is nothing to census: the
abilities are rebuilt by the restoring binary from the catalog, exactly as a
battlefield permanent's are.

### 4. Statics through the layer pass's source list; triggers through the harvester's walk

Two one-line changes, each adding a source list to a pass that already
exists. No new pass, no new registry, no second walk.

**Statics.** `activeStaticAbilitiesLocked` gains
`out = append(out, g.emblemContinuousEffectsLocked()...)`. The helper builds
the same unexported `staticContinuousEffect` adapter the battlefield and the
turn-scoped registry use, bound to a `*Card` pointing into `p.Emblems.Cards`
and to the emblem's creation timestamp. `live` is false: a CR 613.1f
ability-removing effect cannot reach an emblem (nothing can target one), for
the same reason a turn-scoped static cannot be silenced.

The timestamp is CR 613.7c — a continuous effect from a static ability of an
emblem takes its timestamp when the emblem is created — and it is stored in
the `Card.EnteredBattlefieldAt` field, which is the object's CR 613.7
timestamp slot under a battlefield-flavoured name.

**Triggers.** `triggerHarvester.OnEvent` gains
`g.harvestFromEmblemsLocked(&pass)`, which is `harvestFromZone` over each
seat's `Emblems` zone. `harvestFromZone` already reads
`CatalogAbilityKey`, runs `AppliesTo`, and calls `NewTriggeredItem(source, …)`
with the source's controller, so an emblem's "whenever **you** draw a card"
resolves `ByYou` against `source.Controller` with no special case. The
trigger goes on the stack, takes its CR 603.3d target when it is put there,
is dropped when no legal target exists, and can be responded to — like every
other trigger.

Both seams mean a card author writes an emblem's abilities in **exactly** the
vocabulary they write a permanent's: `game.StaticAbility` with a layer and a
sub-layer, `Targeting(WheneverYouDraw(…), spec)` with an `Effect` closure.
There is no emblem dialect.

### 5. One seam creates one: `CreateEmblemForEffect`

`(*Game).CreateEmblemForEffect(owner uuid.UUID, sourceCardID uuid.UUID) error`
is the whole creation surface, wrapped for card files by the
`CreateEmblem{Player, Source}` primitive whose zero fields default to
`ctx.Controller()` and `ctx.Source()`. It resolves the source card in any
zone (`findCardByIDLocked`, so a planeswalker that died in response to its own
ultimate still names its emblem), derives the key, refuses with an
`EventEffectError` when no emblem is registered under it, appends the object,
bumps `layerVersion` and recomputes.

It does **not** emit a new `EventKind`. The resolving ability is already
logged with its printed label ("−7: You get an emblem with …"), the emblem
itself is a persistent chip on the board, and a new kind would want a
`LogEntry` kind and a client renderer for a line the log already carries.
Deferred on purpose, not forgotten.

### 6. An emblem never leaves, except with its owner

Nothing removes an emblem. There is no `MoveCard` path to it (Decision 1), it
is not a permanent so no state-based action sees it, it is not on the
battlefield so no board wipe touches it, and the death of the planeswalker
that made it is irrelevant — the emblem is an independent object, not a
continuous effect sourced from the walker.

The one exit is **CR 800.4a**: a departed player's objects leave the game
with them. `removeObjectsOwnedByLocked` (ADR 0060) gains `sweep(p.Emblems)`,
one line in the function that already sweeps every other zone by ownership.
Their statics stop applying on the recompute that CR 800.4a's step 2 already
runs; their triggers stop firing because the zone is empty.

### 7. Persistence: carried, and the snapshot schema is bumped

`Player.Emblems` is a `*Zone` and is **carried** in
`snapshot_drift_test.go`'s scheme — serialised by `snapshotPlayer`, restored
by `restorePlayer`, deep-copied by `clonePlayer`, and therefore correct
across an undo (`Room.Apply` clones, `Room.Undo` calls `RestoreFrom`) and
across a deploy (`CaptureSnapshot` / `RestoreStrict`). It classifies as
`carried` and not `rebuilt` for the obvious reason: which emblems a player
has is not derivable from anything else on the board.

`SnapshotSchemaVersion` goes **1 → 2**, and `minRestorableSchema` stays 1.
Strictly, an additive field that zero-values correctly needs no bump, and the
const's own doc says so — that rule is about a NEW binary reading an OLD
file, and the new binary reads a pre-emblem file correctly. The bump is for
the other direction: without it, an **older** binary would accept a restore
point written after this lands and silently restore a game with the emblems
missing, which is a game that is wrong in a way nobody can see. `checkSchema`
already refuses a file newer than the reader (`ErrSchemaTooNew`); the bump is
what makes it fire. The cost is that every post-emblem restore point is
refused by a pre-emblem binary, emblems or not, which is the correct trade
for a sandbox whose restore points are a deploy-survival mechanism and not an
archive.

### 8. Wire: `players[i].emblems[]`, public, unredacted

`PlayerView` gains `emblems []EmblemView` with
`{instance_id, label, text}` — the board label and the emblem's oracle text,
so the client can render a chip with hover text and nothing else has to know
what an emblem is. It is **not** a `ZoneView`: a `ZoneView` carries
`CardView`s, and a `CardView` of an object with no characteristics is a row
of empty strings that the hover-zoom, the card-image route and the targeting
layer would all have to learn to skip.

No redaction. Emblems are public information in paper Magic — the emblem is
face up in the command zone and anybody may read it — so `FilterViewFor`
leaves the field alone, in the same posture as `delayed_triggers` and
`life_history`.

The client shows them as chips on the player panel next to the identity
disc (`PlayerIdentity.svelte`), which is where the 2026-09-16 decision on
#623 put them. The command-zone pile stays commander-only.

### 9. Bots see emblems the way they see anthems, and are offered nothing

The evaluator already reads effective characteristics, so Elspeth's emblem
shows up as +2/+2 and flying on the creatures it applies to with no bot
change at all. The legal-move enumerator walks `p.Hand` and `p.Command` and
never sees `p.Emblems`, so it cannot offer an emblem as a cast, an
activation, a target, an attacker or a blocker. A test pins the negative.

### 10. First PR scope: the seam plus three planeswalker ultimates

- **Elspeth, Sun's Champion −7** — static, two `StaticAbility` entries
  (layer 7c +2/+2, layer 6 flying, "creatures you control"). Goes to
  `CompletenessFull`; its caveat and `TestElspethDeclaresItsOmittedUltimate`
  are replaced by the test that pins the emblem.
- **Teferi, Hero of Dominaria −8** — triggered and targeted,
  `Targeting(WheneverYouDraw(…), TargetPermanent("target permanent an
  opponent controls", OpponentControls()))` with `ExileTarget` in the
  closure. Stays `CompletenessCaveats` for its unrelated +1 untap caveat;
  the emblem caveat goes.
- **Wrenn and Six −7** — **still omitted**, and the caveat is rewritten to
  name what is actually left. Its emblem's entire text is "Instant and
  sorcery cards in your graveyard have retrace", and retrace needs two
  things this ADR does not deliver: retrace itself (an alternative cost —
  cast from the graveyard by discarding a land), and a granted cast
  permission a player can hold over a set of cards they did not print
  ([#652](https://github.com/krakenhavoc/cmd_and_ctrl/issues/652)).

  Registering it would offer a −7 that kills the planeswalker, puts a chip
  on the board reading "…have retrace", and then never offers a retrace
  cast. That is the failure ADR 0032 named — an ability whose label promises
  something and delivers a loyalty payment — moved one step later, and a
  player who paid seven loyalty for it gets no signal at all in game.
  Emblems stop being Wrenn's blocker and #652 becomes its only one; the day
  #652 lands, this card gains an `Emblem` slot with one `Static` in it and
  nothing else about it changes.

  `effects.Register` enforces the same judgement for everyone: an
  `EmblemSpec` with no `Static` and no `Triggered` panics at boot. There is
  no way to declare an emblem the engine cannot make do anything.

Invasion of New Phyrexia (#626) is **not** in scope: its back face also needs
a reflexive tap prompt, and #626 decided the whole card stays out until all
three abilities work. Emblems stop being its blocker.

### 11. What is explicitly not in scope

- **A second emblem from the same source.** CR 114 allows a player to get
  two copies of the same emblem, and this implementation permits it —
  `p.Emblems` is a slice and both apply. There is no dedup and no legend
  rule for emblems, which is correct (CR 704.5j is about permanents).
- **An emblem with an activated or mana ability.** None is printed on a card
  in this catalog's reach. `CardDef` would carry them already; the
  activation path reads the battlefield, so the day one shows up it needs
  the same one-line source-list treatment as the other two passes.
- **An emblem-granted replacement effect.** Same shape, same answer.
- **An emblem created by anything but a card's own `Spec`.** There is no
  "make an emblem with this arbitrary ability" API, because there is no
  printed card that needs one and a closure built at resolution time could
  not be snapshotted.
- **A new `EventKind`** (Decision 5) and a **log entry kind**.

## Consequences

### Good

- Emblems close the last S27 seam that is not a card-file follow-up, and
  unblock three registered planeswalkers plus #626's remaining blocker list.
- Card authors write emblem abilities in the vocabulary they already know.
  The three cards here add one `Spec` slot's worth of declaration each and
  no engine code.
- Every rule CR 114 states — can't be cast, can't be moved, isn't a
  permanent, isn't a card, can't be targeted, never leaves — is enforced by
  the container rather than by a check somebody has to remember to write.
- The object rides the existing clone, snapshot, undo and view machinery,
  so an emblem survives an undo, a deploy and a reconnect for free.

### Tradeoffs

- **The command zone is two slices.** Anybody reasoning about "the command
  zone" now has to know that, and a future reader who adds a third
  command-zone tenant has a precedent to follow rather than a rule. The
  alternative was eight filters in eight files and a state-based action that
  eats emblems.
- **An emblem is unreachable by the admin affordances.** The sandbox's
  standing fallback is "a player can fix it by hand", and here they cannot:
  there is no move verb, no context menu and no way to remove an emblem
  short of an undo. That is CR 114 and it is the one place this feature
  deliberately gives up the escape hatch. If a table ever needs it, the fix
  is a named admin verb, not a reachable zone.
- **The schema bump costs every restore point**, not only the ones with
  emblems (Decision 7).
- **One of the four cards the seam row names is still waiting** — Wrenn and
  Six on #652 (Decision 10), and Invasion of New Phyrexia on #626's other
  two blockers. The seam row moves to Closed with two of four converted,
  which is honest but is not the whole row.

## Amendment (2026-09-23, #1315): emblems can widen a turn-based action, not only add Static / Triggered

Decision 4's `EmblemSpec` had two ability slots, `Static` and `Triggered` —
every emblem catalogued so far speaks one of those two vocabularies. Teferi,
Who Slows the Sunset's −7 does not: "Untap all permanents you control
during each opponent's untap step" and "You draw a card during each
opponent's draw step" are both **turn-based-action widenings**
(CR 502.3, CR 504.1), the same shape `effects.Spec.UntapStep` already gives
a battlefield permanent (untap.go, #74) for exactly the reason that file's
header gives: neither clause uses the stack, so writing either as a
`Triggered` entry would land it a step late, on the stack, answerable by a
counter or a tap-in-response with no printed basis.

**What changed.** `EmblemSpec` gained two fields, `UntapStep
[]game.UntapStepPermission` and `DrawStep []game.DrawStepPermission`,
mirroring `Static` / `Triggered` exactly: declared on the card's `Spec.Emblem`,
projected into the emblem's `game.CardDef` by `buildEmblemDef`, and reachable
through the same synthetic `"emblem:<oracle>"` key every other emblem
ability already uses. `checkEmblemSpec`'s "an emblem needs at least one
ability" guard now accepts either new field in place of Static/Triggered,
so an emblem whose whole text is these two clauses (Teferi's) is not an
`EmblemSpec` with "no abilities".

**`DrawStepPermission` is new** (`game/draw_step.go`) — CR 504.1's turn-based
draw had no widening hook at all before this, because no catalogued card
needed one. It is `UntapStepPermission`'s shape one turn-based action over:
an `AppliesTo(g, source, activePlayer) bool` and a `Drawer(g, source)
uuid.UUID` naming who draws, gathered once per draw step
(`activeDrawStepPermissionsLocked`) and applied in `StepDraw`'s case in
`finishStepEntryLocked` — in the SAME turn-based action as the active
player's own CR 504.1 draw, before `EventBeginDrawStep` announces the step
to the trigger harvester. This is deliberately NOT how Howling Mine and
Dictate of Kruphix's "at the beginning of each player's draw step, that
player draws an additional card" work: those print "at the beginning of",
which is a trigger, and both are registered as one (`howling_mine.go`,
`dictate_of_kruphix.go`) — they go on the stack and land strictly after the
turn-based draw. Teferi's emblem prints no such clause, and it does not go
on the stack.

**The untap-step gather now also walks emblems.**
`activeUntapStepPermissionsLocked` (untap.go) previously walked only
`g.Battlefield.Cards` — a `UntapStepPermission` could be declared on
`Spec.UntapStep` (a permanent) but not reached from `EmblemSpec.UntapStep`,
because nothing read the command zone for one. It now walks every seat's
`p.Emblems` too, mirroring the second walk `emblemContinuousEffectsLocked`
and `harvestFromEmblemsLocked` already do for statics and triggers
(Decision 4) — one line of precedent, extended rather than reinvented. The
emblem walk reads `CatalogKey`, not `CatalogAbilityKey`: nothing in the
game can name an emblem to remove its abilities (emblem.go), so there is no
CR 613.1f removal state to consult, exactly as the two existing emblem
walks already read. `activeDrawStepPermissionsLocked` is written with both
walks from the start, for the same reason.

**Consequences.** A future emblem that widens a turn-based action this
engine does not yet model (say, a hypothetical "skip your discard step")
still has no slot — this amendment adds exactly two, for the two clauses a
real card needs today, and follows `Spec.UntapStep`'s own precedent rather
than generalising ahead of a card that asks for it. Nothing about Decision
4's Static/Triggered path changes; the two new fields are additive and the
zero value of both is "declares neither", which is every emblem catalogued
before this card.

Proof card: Teferi, Who Slows the Sunset
([teferi_who_slows_the_sunset.go](../../server/internal/cards/effects/teferi_who_slows_the_sunset.go)).
Tracker: [#884](https://github.com/krakenhavoc/cmd_and_ctrl/issues/884).

## Note (2026-09-24, #1275): a fifth slot — activation timing

`EmblemSpec` gains `ActivationTimings []game.ActivationTiming`, the emblem-side
mirror of `Spec.ActivationTimings` (#1208), and the per-player activation
timing walk (`activationTimingVerdictLocked`, game/activation_timing.go) walks
every seat's `p.Emblems` after the battlefield. That is the fifth emblem walk,
beside the layer pass's, the trigger harvest's and #1315's two turn-based-action
gathers, and it does not change what an emblem IS: the object, its key, its
zone, its persistence and its wire are all as Decisions 1–8 left them. The
design is recorded where the type lives, in
[ADR 0066's 2026-09-24 amendment](0066-granted-cast-and-play-permissions.md).

Proof card: Teferi, Temporal Archmage's −10.
