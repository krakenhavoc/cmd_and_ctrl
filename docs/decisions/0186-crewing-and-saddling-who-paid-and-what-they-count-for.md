# ADR 0186 — Crewing and saddling: who paid, and what they count for

Date: 2026-10-14
Status: Accepted (owner decisions 2026-10-14, recorded under [Owner decisions](#owner-decisions-2026-10-14))
Sprint: S58 — Deck requests, October batch
Issues: #2704 (seam: creatures that saddled a Mount, CR 702.171c), #2705 (seam: saddle and crew "as though its power were 2 greater"). Registry rows: `saddlers-of-a-mount`, `saddle-crew-power-bonus`.
Builds on: #2695 (saddle, the [ADR 0071](0071-designations-that-switch-abilities-on.md) amendment of 2026-10-08), S27's crew (`effects/vehicles.go`), [ADR 0059](0059-turn-machinery.md) Decision 8 and [ADR 0108](0108-turn-scoped-effects-object-history-and-damage-shields.md) group H (per-object history on `TurnTally`), [ADR 0083](0083-token-abilities.md) (token templates with their own abilities), [ADR 0019](0019-structured-targeting.md) and its amendments (data clauses on `TargetSpec` judged beside `CardOK`).
Numbering: after `git fetch --all --prune`, every file `docs/decisions/` has held on any remote head was listed (`git log --remotes --name-only -- docs/decisions/`; 14 refs) and the files of every open PR were read (#3514, #3496). `develop` ends at 0184 and #3514 holds 0185, so this ADR takes **0186**, the number reserved for it.

## Context

Two seams, found while building saddle (#2695), share one place in the engine: the payment of a crew or saddle cost, `validateCrewCostLocked` and the tap loop after it in `game/activated.go`. One seam is about what the engine remembers from that payment. The other is about how much each creature counts toward it.

### The rules

Checked against `MagicCompRules 20260925.txt`.

- **CR 702.122a:** "Crew N" means "Tap any number of other untapped creatures you control with total power N or greater: This permanent becomes an artifact creature until end of turn."
- **CR 702.122b:** "A creature 'crews a Vehicle' when it's tapped to pay the cost to activate a Vehicle's crew ability."
- **CR 702.122c:** "A Vehicle is 'crewed by' a creature if that creature was tapped to pay the cost to activate that Vehicle's crew ability."
- **CR 702.171a:** "Saddle N" means "Tap any number of other untapped creatures you control with total power N or greater: This permanent becomes saddled until end of turn. Activate only as a sorcery."
- **CR 702.171b:** saddled is a designation that lasts until the turn ends or the permanent leaves the battlefield.
- **CR 702.171c:** "A creature 'saddles' a permanent as it's tapped to pay the cost to activate a permanent's saddle ability."
- **CR 702.184c** (station) is the only rule that says a static can change which number a tapped creature contributes: "Static abilities may modify the result of a station ability by causing it to use a characteristic other than the tapped creature's power". Crew and saddle have no such rule. The Pilot cards' own text is the instruction.
- **CR 208.3a**, with the Veteran Motorist example: a "whenever this creature crews a Vehicle" trigger resolves while the Vehicle is not yet a creature. The trigger goes on the stack above the crew ability, because the creature crews it as the cost is paid.
- **CR 400.7:** a creature that left the battlefield and came back is a new object, and did not crew or saddle anything.
- **CR 608.2h:** a sacrificed creature's power is read as it last existed (The Gitrog's ruling).

Rulings that settle the design (Scryfall, read 2026-10-14):

- **Fortune, Loyal Steed** (2024-04-12): "This ability will trigger whether or not Fortune is still on the battlefield. If it's not, you can still exile up to one creature that saddled it that turn." So the record of who saddled a Mount must outlive the Mount. Today it does not.
- **Fortune:** "The delayed triggered ability doesn't target any creature. You choose which other creature to exile, if any, as that ability is resolving."
- **Calamity, Galloping Inferno:** "You don't have to choose the same creature both times"; the tokens "were never declared as attackers"; you choose what each token attacks.
- **Dynamite Diver** (2025-02-07): "Saddling a Mount or crewing a Vehicle doesn't cause this creature's power to change."
- **Tapestry Warden** (2025-07-25), on station's version of the same idea: the ability doesn't change any creature's power, and other rules and effects that check power use the real value.

### What exists (develop at 8dd2c2e2c)

- **The payment.** `ActivateCatalogAbility` validates `ActivateAbilityParams.CrewIDs` with `validateCrewCostLocked` (`game/activated.go`), which sums `c.CurrentPower()` over the named creatures, refuses the Mount itself for a saddle cost, and returns `ErrInsufficientCrew` below the number. The tap loop after it sets `Tapped`, emits `EventTapCard` per creature and, for a saddle cost only, copies each creature into `saddlers`, which become `item.Paid.TappedOthers` (`PaidTap`, ID and epoch).
- **The saddle record.** `saddleEffect` (`effects/saddle.go`) hands those refs to `Game.SaddleForEffect` at **resolution**, which appends them to `Card.SaddledBy []ObjectRef` on the Mount (`game/card.go`). `Card.SaddledBy` is cleared with `Saddled` at cleanup (`clearSaddledLocked`) and when the Mount leaves the battlefield (`zone.go`, `entry_tail.go`). `Game.SaddlersOf(mountID)` returns the saddlers still on the battlefield as the same objects, and nothing at all when the Mount is not on the battlefield or not saddled. Rambling Possum is the only reader. Snapshot: `battlefield.cards[].saddledBy` (v8).
- **No crew record.** Nothing records which creatures crewed a Vehicle, and no event says a creature crewed or saddled anything.
- **History precedent.** `TurnTally` (`game/turn_tally.go`) holds the turn's per-object history as data: `Attacks []AttackRecord`, `Exerts []ExertRecord`, `DamageDealers []DamageDealtRecord`. It is replaced as the turn begins (`resetTurnTallyLocked`), copied by `cloneTurnTally` for undo, and snapshotted as `turnTally` (v8).
- **Target precedent.** `TargetSpec.RelativeToSource`, `CombatWithSource` and `ExiledWithSource` (`game/targets.go`) are data clauses about the SOURCE, judged in `sourceAdmits` beside `CardOK`, so the legal set, the announce check (CR 601.2c), the resolution re-check (CR 608.2b), the client's picker and the bot agree. The source is `TargetSource.ObjectRef()`, live or last-known.
- **A rule-modifying flag precedent.** `Characteristic.DamageByToughness` (`game/characteristic.go`) is set by a layer-6 static (Ghalta the Immovable) and read in one place (`combatDamagePower`). It is compared in `sameCharacteristic` (`game/layer_dependency.go`) and is not snapshotted: the effective characteristics are recomputed.
- **The other readers of crew power.** The enumerator's `crewPayment` (`legal/abilities.go`) picks one greedy set by `CurrentPower()`, and teamwork reuses it (`legal/teamwork_blight.go`). The wire sends `crew_cost` and `crew_options` (card IDs only) on `ActivatedAbilityView` (`protocol/view.go`, `crewOptions`), and the client sums `CardView.power` (`client/src/lib/crew.ts`, `CrewCostModal.svelte`). The heuristic prices each named crewer as a tapped blocker (`aiseat/heuristic/moves.go`).
- **Tokens.** A token with a printed ability is a `tokenTemplate` (`effects/token_catalog.go`) whose `Static` slot is registered under `game.TokenKey(slug)`; `Text` is the printed line.

### The waiting cards (Commander-legal unless noted)

Creatures that saddled a Mount (#2704):

- **Giant Beaver** {3}{G}, 4/4 Beaver Mount: "Vigilance. Whenever this creature attacks while saddled, put a +1/+1 counter on target creature that saddled it this turn. Saddle 3".
- **The Gitrog, Ravenous Ride** {3}{B}{G}, 6/5: "Trample, haste. Whenever The Gitrog deals combat damage to a player, you may sacrifice a creature that saddled it this turn. If you do, draw X cards, then put up to X land cards from your hand onto the battlefield tapped, where X is the sacrificed creature's power. Saddle 1".
- **Calamity, Galloping Inferno** {4}{R}{R}, 4/6: "Haste. Whenever Calamity attacks while saddled, choose a nonlegendary creature that saddled it this turn and create a tapped and attacking token that's a copy of it. Sacrifice that token at the beginning of the next end step. Repeat this process once. Saddle 1".
- **Fortune, Loyal Steed** {2}{W}, 2/4: "When Fortune enters, scry 2. Whenever Fortune attacks while saddled, at end of combat, exile it and up to one creature that saddled it this turn, then return those cards to the battlefield under their owner's control. Saddle 1".

Crewing and saddling as though power were greater (#2705):

- **Pilot token** (the "saddles" variant, made by most cards below): 1/1 colorless Pilot, "This token saddles Mounts and crews Vehicles as though its power were 2 greater."
- **Cloudspire Captain** {2}{W}, 2/3: "Mounts and Vehicles you control get +1/+1. This creature saddles Mounts and crews Vehicles as though its power were 2 greater."
- **Dynamite Diver** {R}, 1/1: the same clause; "When this creature dies, it deals 1 damage to any target."
- **Deathless Pilot** {1}{B}, 2/2: the same clause; "{3}{B}: Return this card from your graveyard to your hand."
- **Interface Ace** {1}{W}, 0/4 artifact creature: "This creature saddles Mounts and crews Vehicles using its toughness rather than its power. Whenever this creature becomes tapped during your turn, untap it. This ability triggers only once each turn."
- **Cloudspire Coordinator** {R}{W}, 3/1: "When this creature enters, scry 2. {T}: Create X 1/1 colorless Pilot creature tokens, where X is the number of Mounts and/or Vehicles that entered the battlefield under your control this turn. The tokens have 'This token saddles Mounts and crews Vehicles as though its power were 2 greater.'"
- **Defend the Rider** {G} instant: "Choose one — • Target permanent you control gains hexproof and indestructible until end of turn. • Create a 1/1 colorless Pilot creature token with '…'."
- **Valor's Flagship** {4}{W}{W}{W}, 7/7 Vehicle: "Flying, first strike, lifelink. Crew 3. Cycling {X}{2}{W}. When you cycle this card, create X 1/1 colorless Pilot creature tokens with '…'."
- **Back on Track** {4}{B} sorcery: "Return target creature or Vehicle card from your graveyard to the battlefield. Create a 1/1 colorless Pilot creature token with '…'."
- **Roadside Assistance** {2}{W} Aura: "Enchant creature or Vehicle. When this Aura enters, create a 1/1 colorless Pilot creature token with '…'. Enchanted permanent gets +1/+1 and has lifelink."
- **Country, Foul, Reef, Rocky and Wild Roads** (lands): "This land enters tapped unless you control a Mount or Vehicle. {T}: Add {W}/{B}/{U}/{R}/{G}. {1}{W}, {T}, Sacrifice this land: Create a 1/1 colorless Pilot creature token with '…'. Activate only as a sorcery."
- **Canyon Vaulter** {1}{W}, 3/1: "Whenever this creature saddles a Mount or crews a Vehicle during your main phase, that Mount or Vehicle gains flying until end of turn."
- **Reckless Velocitaur** {3}{R}, 3/3: "Whenever this creature saddles a Mount or crews a Vehicle during your main phase, that Mount or Vehicle gets +2/+0 and gains trample until end of turn."

Not on either row, but expressible once these land (none is catalogued today): the crew-only Pilot token ("This token crews Vehicles as though its power were 2 greater.") from **Shorikai, Genesis Engine**, **Prodigy's Prototype** and **Born to Drive**; **Hotshot Mechanic** ("crews Vehicles as though its power were 2 greater"); **Giant Ox** ("crews Vehicles using its toughness rather than its power"); the "creature that crewed it this turn" Vehicles **Turtle Van**, **Getaway Car**, **Subterranean Schooner**, **Golden Argosy** and **Luxurious Locomotive**; and the "whenever this creature crews a Vehicle" Pilots **Veteran Motorist**, **Gearshift Ace** and **Speedway Fanatic**. Mindlink Mech (becomes a copy of a crewer) and Adrestia ("if an Assassin crewed it") need more than this ADR and are left out.

## Options considered

### Where the record lives

- **A. Keep it on the Mount (`Card.SaddledBy`) and add `Card.CrewedBy`.** Smallest change. Rejected: a card-held record is cleared when the permanent leaves, so Fortune's ruling ("whether or not Fortune is still on the battlefield") and The Gitrog dying to combat damage while its trigger waits both read nothing. Keeping the slice on the card after it leaves would put history on a graveyard card, which every other per-object history in the engine avoids.
- **B. Walk the event log for `EventTapCard` events inside a crew activation.** Rejected: the log is not the rules state (`TurnTally` exists because log walks broke on tokens and departed objects, #596), and nothing in a tap event says which activation it paid for.
- **C. A turn tally of who crewed and saddled what, keyed by objects (chosen).** One record per creature per payment on `TurnTally`, naming the permanent and the creature as objects (instance ID and epoch). It outlives both, is reset with the turn, clones for undo, and snapshots as data. It is the shape `Attacks` and `Exerts` already have.

### When it is recorded

- **At resolution (today's saddle).** Rejected: CR 702.122b and 702.171c say a creature crews or saddles *as it is tapped to pay the cost*. A crew ability that is countered still had its creatures crew the Vehicle, and the "whenever this creature crews" trigger must go on the stack above the crew ability (CR 208.3a's example).
- **At payment (chosen).** In the tap loop that already taps each creature. Owner decision 1.

### How a creature counts for more than its power

- **A per-card hook in the validator** (a func keyed by oracle ID). Rejected: a granted or copied ability would not carry it, ability removal would not take it away, and it adds a closure route to the ratchet.
- **Change the creature's power during payment.** Rejected by the Dynamite Diver ruling: its power does not change.
- **A layer-6 rule on the creature's effective characteristics (chosen).** `DamageByToughness`'s shape: a static writes a fact about how the creature counts, and the payment reads it. A Pilot that loses all abilities stops counting extra (its catalog static is not applied, CR 613.1f); a copy of a Pilot counts extra; a Pilot token's ability comes from its token template.

## Decision

### 1. `TurnTally.Crews`: who crewed and saddled what this turn

```go
// CrewRecord is one creature tapped to pay one crew or saddle cost
// (CR 702.122b-c, CR 702.171c).
type CrewRecord struct {
    Saddle    bool      `json:"saddle,omitempty"` // false: crew
    Permanent ObjectRef `json:"permanent"`        // the Vehicle or Mount, as the object it was when the cost was paid
    Creature  ObjectRef `json:"creature"`         // the creature tapped, as the object it was when it tapped
}

// TurnTally
Crews []CrewRecord `json:"crews,omitempty"`
```

- **Written** in the tap loop of the activation path in `game/activated.go`, the loop that already sets `Tapped` on each ID `validateCrewCostLocked` returned. One record per creature, in payment order, for a crew cost (`ab.Cost.Crew > 0`) and a saddle cost (`ab.Cost.Saddle > 0`). The `Permanent` epoch is the source's epoch as the cost is paid. As `exertLocked` does, the slice is copied before the append so the undo clone never shares the backing array (#1238). A Heart of Kiran crew paid by removing a loyalty counter taps no creature and records nothing.
- **Reset** with the rest of the tally as the turn begins (`resetTurnTallyLocked`), and copied in `cloneTurnTally`.
- **Read** through two game methods, both under `g.mu`:
  - `CrewRecordsForEffect(permanent ObjectRef, saddle bool) []CrewRecord`: every record for that permanent object and kind, departed creatures included, in order.
  - `CreaturesThatCrewedForEffect(permanent ObjectRef)` and `CreaturesThatSaddledForEffect(permanent ObjectRef) []uuid.UUID`: the distinct creatures from those records that are still on the battlefield as the same object (`findBattlefieldCard` and an epoch match), in first-payment order. This is "creatures that crewed / saddled it this turn".
  The permanent is named by `ObjectRef`, never by a live lookup, so the answer survives the Vehicle or Mount leaving (Fortune's ruling). A Mount that left and came back is a new object with no saddlers (CR 400.7).
- **Card side** (`effects/saddle.go`, `effects/vehicles.go`): `SaddlersOfThis(ctx *Context) []uuid.UUID` and `CrewersOfThis(ctx *Context) []uuid.UUID` read through `ctx.SourceRef()`. A delayed trigger's `SourceObject` is its parent's source object (CR 603.7e, `ScheduleDelayedTriggerForEffect`), so Fortune's end-of-combat trigger reads Fortune's saddlers.
- **`Game.SaddlersOf(mountID)`** is kept as a wrapper: it resolves the Mount's current object and calls `CreaturesThatSaddledForEffect`. `effects.SaddlersOf(ctx, mount)` stays as an alias. Rambling Possum switches to `SaddlersOfThis`, so a Possum that left in response still offers its saddlers, as Fortune's ruling reads the rule.

**`Card.SaddledBy` stays.** It is in the v8 shape and a binary from before this ADR reads it, so `saddleEffect` keeps writing it unchanged. The new reader falls back to it in one case: when the tally holds no saddle record for that Mount object and the Mount is on the battlefield with a non-empty `SaddledBy`. That is a game restored mid-turn from a file written before this change. Nothing else reads `SaddledBy`.

### 2. Two events: a creature crews, a creature saddles

`EventCrewedVehicle` (`"crewed_vehicle"`) and `EventSaddledMount` (`"saddled_mount"`), emitted in the same tap loop, straight after each creature's `EventTapCard`, one per creature: `Actor` the activator, `CardID` the creature, `Target` and `Source` the Vehicle or Mount. They are emitted while the cost is paid, so a trigger on them is put on the stack above the crew or saddle ability (CR 603.3b, CR 208.3a).

`effects.WhenThisCrewsOrSaddles(label string, crew, saddle bool, effect)` builds "whenever this creature crews a Vehicle", "… saddles a Mount" or "… saddles a Mount or crews a Vehicle": it watches the kinds asked for, with `ev.CardID == source.InstanceID`. The built item carries `ev.Target` on `item.Params` (Caustic Bronco's carrier), so "that Mount or Vehicle" is the object the creature paid for. "During your main phase" is a condition in `AppliesTo` (`b14InYourMainPhase`), checked when the event happens; it is not an intervening "if".

Both kinds go on the public log's silence list (`protocol/log_event_kind_gate_test.go`) with `silentImpliedByAnotherLine`: the activation line and the taps already say it.

### 3. A target that must be a crewer or a saddler

`TargetSpec.PaidForSource` (`PaidForSourceKind`: `""`, `"crewed"`, `"saddled"`) is a data clause beside `ExiledWithSource`. It is judged in `sourceAdmits`: the candidate must be in `CreaturesThatCrewedForEffect` / `CreaturesThatSaddledForEffect` of `TargetSource.ObjectRef()`. With no source object it admits nothing. Being data, it rides a stack item and a paused prompt without a closure, and the announce check, the CR 608.2b re-check, the client's picker and the bot all read the same answer. `effects.SaddledThis(spec)` and `effects.CrewedThis(spec)` wrap a clause, as `ExiledWithThis` does:

```go
Targets: SaddledThis(TargetCreature("target creature that saddled it this turn")), // Giant Beaver
```

A non-targeted "choose a creature that saddled it" is a `ChooseCardsPrompt` over `SaddlersOfThis` at resolution (Rambling Possum's shape), never this clause.

### 4. What a creature counts for when it crews or saddles

```go
// CostPowerUse names the costs a CostPowerRule changes. A bitmask.
type CostPowerUse uint8
const (
    CostPowerCrew   CostPowerUse = 1 << iota // "crews Vehicles"
    CostPowerSaddle                          // "saddles Mounts"
)

// CostPowerRule is "this creature [saddles Mounts and] crews Vehicles
// as though its power were N greater" (Bonus N) or "… using its
// toughness rather than its power" (UseToughness).
type CostPowerRule struct {
    Uses         CostPowerUse
    Bonus        int
    UseToughness bool
}

// Characteristic
CostPower []CostPowerRule `json:",omitempty"`
```

- **Written** only by layer-6 statics (CR 613.1f), only ever appended to, like `DamageByToughness`. A creature's own clause is a `selfOnly` static on its card or token template, so "loses all abilities" takes it away. Added to `sameCharacteristic` (`slices.Equal`). It is not snapshotted: the effective characteristics are recomputed from the catalog.
- **Read** in one place, `func (c Card) CountedPower(use CostPowerUse) int`:
  1. Start from `PowerForComparison()` (layers and P/T counters, not clamped). If any rule for this use has `UseToughness`, start from `CurrentToughness()` instead.
  2. Add the `Bonus` of every rule for this use. Two rules add up: a Pilot token that also benefits from a second "as though" effect counts both.
  3. Clamp at zero. A creature never takes power away from a crew total, which is what `CurrentPower()`'s clamp does today. Owner decision 2.
  The order (toughness first, then the bonuses) is this ADR's reading; no ruling says how the two combine, and no Commander-legal card gives a creature both.
- **Callers:** `validateCrewCostLocked` sums `CountedPower(CostPowerCrew)` or `CountedPower(CostPowerSaddle)`. Nothing else in the engine changes: the creature's power is not changed (Dynamite Diver's ruling), so "where X is its power" still reads the real power.
- **Constructors** (`effects/vehicles.go`): `CrewsAsThoughPowerGreater(n int, uses game.CostPowerUse) game.StaticAbility` and `CrewsUsingToughness(uses game.CostPowerUse) game.StaticAbility`, each a layer-6 `selfOnly` static whose `Label` is the printed sentence.
- **Station** (CR 702.184c, Tapestry Warden, Stoic Star-Captain) is the same idea for a third cost, read at resolution through `PaidTapPowerForEffect` and frozen as last-known power. It is out of scope here; a later change adds `CostPowerStation` and freezes the counted value on `PaidTap`.

### 5. The Pilot tokens

Two token templates in `effects/token_catalog.go`, each a 1/1 colorless Pilot creature with one static:

- `pilot-saddles-and-crews`: `CrewsAsThoughPowerGreater(2, CostPowerCrew|CostPowerSaddle)`, `Text` "This token saddles Mounts and crews Vehicles as though its power were 2 greater."
- `pilot-crews`: `CrewsAsThoughPowerGreater(2, CostPowerCrew)`, `Text` "This token crews Vehicles as though its power were 2 greater."

The slugs are on-disk identities (a snapshot carries the token key) and are never renamed. `effects.PilotToken()` and `effects.CrewPilotToken()` mint them through `tokenFromCatalog`.

### 6. The wire, the enumerator, the bot and the client

- **Wire.** `ActivatedAbilityView.CrewPower map[string]int` (`crew_power`, omitempty) is sent beside `crew_options` on every crew and saddle ability: the counted power of each offered creature for that ability. It is absent on teamwork prompts, which count plain power. A client that ignores it sums `CardView.power` as today, which under-counts a Pilot and never over-counts. Documented in `docs/protocol.md`.
- **Enumerator.** `crewPayment(need int, except uuid.UUID, use game.CostPowerUse)` sorts and sums by `CountedPower(use)`; teamwork passes `0`, which means plain `CurrentPower()`. Among creatures that count the same, the one with the lower real power is taken first, so a Pilot token is tapped before a 3/3 that would otherwise attack.
- **Heuristic bot.** No new pricing: it already charges each named crewer as a tapped blocker (`moves.go`), and the enumerator now names Pilots. The MCP seat reads the same legal moves, and a card's ability text reaches it through `card`.
- **Client.** `crew.ts` reads `a.crew_power?.[id] ?? card.power` in `crewPower` and `crewAvailablePower`. `CrewCostModal.svelte` shows "counts as N" beside a creature whose counted power differs from its power, through a new label key in `labels.ts`.
- **Who crewed or saddled** is shown on the table (owner decision 3). `CardView.CrewedBy` (`crewed_by`) and `CardView.SaddledBy` (`saddled_by`), both `[]string` and omitempty, are set on a Vehicle or Mount on the battlefield to `CreaturesThatCrewedForEffect` / `CreaturesThatSaddledForEffect` of its current object: the creatures that crewed or saddled it this turn and are still on the battlefield as the same objects. Tapping to pay is public, so every viewer gets them. The face-down redaction clears both with `Saddled`. The client names those creatures on the permanent: the saddled badge's tooltip says who saddled it, and a crewed Vehicle gets a "crewed" badge whose tooltip says who crewed it, with the names taken from the board's card views. The wording goes through `labels.ts`. Hovering the badge does not highlight the creatures; that can come later if players ask.

### 7. Snapshot, restore, undo and the closure ratchet

- **Snapshot (v8, additive).** One new key, `turnTally.crews[]` (`saddle`, `permanent.id`, `permanent.epoch`, `creature.id`, `creature.epoch`), recorded in `internal/game/testdata/snapshot_shape/v8.txt` with `-update-shape`. Its absence is "nobody crewed or saddled anything this turn", which is right for a file written before it; the saddle reader's fallback to `Card.SaddledBy` covers the one turn in flight at a deploy. A binary from before it drops the key and reads `Card.SaddledBy`, which is still written. No fixture changes. `Characteristic.CostPower` is not on disk. `TargetSpec.PaidForSource` is a string field on a spec that already rides stack items and prompts; additive and `omitempty`. The `crewed_by` / `saddled_by` view fields are derived from the tally and are not on disk.
- **Restore points.** Nothing here adds a continuation. Fortune's and Calamity's delayed triggers use `ScheduleDelayedTrigger` bodies registered by key, as Dorothea and Reflection of Kiki-Jiki do. The Gitrog's and Calamity's choices are `ChooseCardsPrompt`s, already classified.
- **Undo.** `TurnTally` is part of the undo clone; the copy-before-append rule keeps the clone's slice its own. Undoing a crew activation removes its records with the taps.
- **Closure ratchet.** No new route from `Game` to a func or interface. `CrewRecord`, `CostPowerRule` and `PaidForSourceKind` are data; the two statics are catalog closures reached through the catalog key, as every static is. `closure_fields.txt` is unchanged.

## Delivery plan

Each slice is one PR into `develop`. Run Go only through `scripts/go-docker.sh`. Each slice that adds cards generates only those cards' oracle fixtures (`-update-oracle -oracle-ids=…`), refreshes the roadmap golden (`scripts/go-docker.sh go test ./internal/roadmap/ -update`), and moves its cards from the row's `Waiting` to `Examples`.

1. **`feat(server): record the creatures that crewed or saddled a permanent this turn`** (tier:2-standard; no dependency)
   - Scope: `game/turn_tally.go` (`CrewRecord`, `TurnTally.Crews`, `cloneTurnTally`), the tap loop in `game/activated.go`, the readers in `game/designations.go` (`CrewRecordsForEffect`, `CreaturesThatCrewedForEffect`, `CreaturesThatSaddledForEffect`; `SaddlersOf` rewritten over them with the `Card.SaddledBy` fallback), `effects/saddle.go` and `effects/vehicles.go` (`SaddlersOfThis`, `CrewersOfThis`), `effects/rambling_possum.go`, `snapshot_shape/v8.txt`, `CardView.CrewedBy` / `SaddledBy` in `protocol/view.go` (set from the readers, cleared by the face-down redaction) with their entries in `docs/protocol.md`, the saddle paragraph of `docs/adding-cards.md`, and the two registry rows' `EngineNotes`.
   - Acceptance tests: crewing a Vehicle with two creatures writes two crew records naming the Vehicle's and each creature's object; a saddle writes saddle records; a crew ability countered after payment still leaves its records (CR 702.122c); a crewer flickered after paying is not returned by `CreaturesThatCrewedForEffect` (CR 400.7) but is in `CrewRecordsForEffect`; a Mount that left the battlefield still answers `CreaturesThatSaddledForEffect` by its old `ObjectRef`, and its new object answers nothing; records are gone after the next turn begins; undo of the activation removes them and the clone does not share the slice; a game restored from a v8 file with `saddledBy` and no `crews` still gives Rambling Possum its saddlers; Rambling Possum bounced in response to its trigger still offers its saddlers; the v8 shape test passes with the new key; a crewed Vehicle's view carries `crewed_by` with both crewers, every seat and a spectator see it, it drops a crewer that left the battlefield, and it is absent the next turn; a saddled Mount's view carries `saddled_by`.

2. **`feat(client): show which creatures crewed or saddled a permanent`** (tier:1-mechanical; after 1)
   - Scope: `client/src/lib/protocol.ts` (`crewed_by`, `saddled_by` on `CardView`), `client/src/lib/components/board/Card.svelte` (the saddled badge's tooltip names the saddlers; a new crewed badge on a Vehicle with a non-empty `crewed_by`, its tooltip naming the crewers), new keys in `client/src/lib/labels.ts` and `docs/labels.md`, and a render test beside the existing badge tests. Names come from the board's card views; an ID with no card on the board is skipped.
   - Acceptance tests: a Mount with `saddled_by` of two creatures shows the saddled badge with both names in its tooltip; a Vehicle with `crewed_by` shows the crewed badge and names; a Vehicle without `crewed_by` shows no crewed badge; an ID missing from the board is left out rather than shown as an ID; a face-down permanent shows neither badge.

3. **`feat(server): crews-a-Vehicle and saddles-a-Mount triggers, with Canyon Vaulter and Reckless Velocitaur`** (tier:2-standard; after 1)
   - Scope: `EventCrewedVehicle` and `EventSaddledMount` in `game/events.go`, emitted in the tap loop after each `EventTapCard`; the two silence entries in `protocol/log_event_kind_gate_test.go`; `effects.WhenThisCrewsOrSaddles`; `cards/effects/canyon_vaulter.go` and `reckless_velocitaur.go` (Full). Registry: the two cards leave `saddle-crew-power-bonus`'s `Waiting`.
   - Acceptance tests: one event per creature, after its tap, with `CardID` the creature and `Target` the Vehicle; Canyon Vaulter crewing in your main phase puts its trigger above the crew ability, and it resolves first; the Vehicle has flying once crewed; the same crew in combat or on an opponent's turn does not trigger; Reckless Velocitaur saddling a Mount gives it +2/+0 and trample; a Vehicle crewed by two Vaulters gets two triggers.

4. **`feat(server): a target that must have crewed or saddled its source, and Giant Beaver`** (tier:2-standard; after 1)
   - Scope: `TargetSpec.PaidForSource` and its gate in `sourceAdmits` (`game/targets.go`), `effects.SaddledThis` / `CrewedThis` in `effects/targets.go`, the targeting paragraph of `docs/adding-cards.md`, `cards/effects/giant_beaver.go` (Full).
   - Acceptance tests: Giant Beaver's attack trigger offers exactly the creatures that saddled it and are still on the battlefield; a creature that saddled a different Mount is not offered; announcing an illegal target is refused; a saddler flickered in response makes the trigger fizzle (CR 608.2b); with no saddlers on the battlefield the trigger is removed for want of a target (CR 603.3d); the enumerator's and the view's candidate lists match the engine's.

5. **`feat(server): the gitrog, ravenous ride and fortune, loyal steed`** (tier:2-standard; after 1)
   - Scope: `cards/effects/the_gitrog_ravenous_ride.go` and `fortune_loyal_steed.go` (Full), with any small helper in a shared file. The Gitrog: a combat-damage-to-a-player trigger; a `ChooseCardsPrompt` (0..1) over `SaddlersOfThis` filtered to creatures its controller controls; sacrifice; X is the sacrificed creature's last-known power (Ghoulcaller Gisa's read); draw X; then `PutFromHandOntoBattlefield` with lands, tapped, `Max` X. Fortune: `AttacksWhileSaddled` schedules a `ScheduleDelayedTrigger{At: game.StepEndCombat}` body (Dorothea's shape) that chooses up to one saddler (not targeted), exiles Fortune if it is still the same object and the chosen creature, then returns those cards under their owners' control (the `flicker.go` return); tokens do not return.
   - Acceptance tests: The Gitrog, saddled by a 3-power creature, deals combat damage, sacrifices it and draws three, then puts up to three lands tapped; declining sacrifices nothing and draws nothing; X uses the power the creature had as it was sacrificed (a pumped creature counts its pump); Fortune with a saddler flickers both at end of combat; Fortune killed before end of combat still lets its controller flicker a saddler (the ruling); a token saddler does not return; choosing none flickers only Fortune.

6. **`feat(server): calamity, galloping inferno`** (tier:2-standard; after 1)
   - Scope: `cards/effects/calamity_galloping_inferno.go` (Full). Two rounds, each: a `ChooseCardsPrompt` (exactly one) over nonlegendary creatures in `SaddlersOfThis`; `TokenCopyTemplate` and `CreateTokensAttackingYourChoice` (Sandstorm Crasher's shape), tapped and attacking with the attack target chosen per token; a delayed sacrifice of that token at the beginning of the next end step (Reflection of Kiki-Jiki's shape). The second round starts from the first round's continuation, so it may choose the same creature.
   - Acceptance tests: two token copies enter tapped and attacking; their "whenever this attacks" triggers do not fire (CR 508.4); each can attack a different player; both are sacrificed at the next end step; a legendary saddler is not offered; with no nonlegendary saddler on the battlefield nothing happens; choosing the same creature twice makes two copies.

7. **`feat(server): count a creature's crew and saddle power with a modifier`** (tier:2-standard; no dependency)
   - Scope: `CostPowerUse`, `CostPowerRule`, `Characteristic.CostPower` (`game/characteristic.go`), `sameCharacteristic` (`game/layer_dependency.go`), `Card.CountedPower` (`game/card.go`), `validateCrewCostLocked`; `effects.CrewsAsThoughPowerGreater` and `CrewsUsingToughness`; `crewPayment` with its `use` parameter and tie-break (`legal/abilities.go`, `legal/teamwork_blight.go`); `ActivatedAbilityView.CrewPower` and its population beside `crewOptions` (`protocol/view.go`); `docs/protocol.md`. No cards in this slice, because the client does not count the bonus until slice 8.
   - Acceptance tests (with test-only cards built in the test file): a 1/1 with a +2 crew rule crews a crew-3 Vehicle alone; the same creature cannot saddle a Saddle 3 Mount when the rule says crew only; a 0/4 with `UseToughness` crews 4; the rule is gone after "loses all abilities"; the creature's `CurrentPower()` is unchanged throughout; a -1/1 with +2 counts 1; two +2 rules count +4; the enumerator offers a crew move naming the 1/1 alone; teamwork still counts plain power; `crew_power` carries 3 for the 1/1 on a crew ability and is absent on a teamwork prompt; the layer-dependency probe treats two passes with different `CostPower` as different.

8. **`feat(client): count crew and saddle power the way the server does`** (tier:1-mechanical; after 7)
   - Scope: `client/src/lib/protocol.ts` (`crew_power`), `client/src/lib/crew.ts`, `CrewCostModal.svelte`, a new key in `client/src/lib/labels.ts` and `docs/labels.md`, `crew.test.ts` and `costSheets.render.test.ts`.
   - Acceptance tests: `crewPower` and `crewAvailablePower` use `crew_power` when present and `power` otherwise; the modal enables Crew with a lone 1/1 counted as 3 on a crew-3 Vehicle; the "counts as 3" note shows only when the counted power differs; a teamwork prompt is unchanged.

9. **`feat(server): pilot tokens and the pilot creatures`** (tier:2-standard; after 7 and 8)
   - Scope: the two templates and `PilotToken` / `CrewPilotToken` (§5) in `effects/token_catalog.go` and `effects/tokens.go`; `cloudspire_captain.go` (with a layer-7c +1/+1 to Mounts and Vehicles you control), `dynamite_diver.go`, `deathless_pilot.go` (the graveyard-to-hand ability, Blessed Ghoul's shape) and `interface_ace.go` (the once-each-turn untap on becoming tapped during your turn, Dionus's shape), all Full.
   - Acceptance tests: each Pilot creature crews and saddles for its counted power; a minted Pilot token counts 3 for crew and saddle and survives a snapshot round trip with its ability; the crew-only token counts 1 for saddle; Interface Ace taps to crew on your turn and untaps once; Cloudspire Captain's anthem applies to a crewed Vehicle and a Mount and not to other creatures.

10. **`feat(server): the cards that make pilot tokens`** (tier:1-mechanical; after 9)
   - Scope: `defend_the_rider.go`, `valors_flagship.go` (cycling {X}{2}{W} and the cycle trigger, Shark Typhoon's shape), `back_on_track.go`, `roadside_assistance.go` (Enchant creature or Vehicle, Lightwheel Enhancements' shape), `cloudspire_coordinator.go` (X from `EnteredWithSubtypeThisTurn` for Mount plus Vehicle; no printed permanent is both), and the five Roads (enters tapped unless you control a Mount or Vehicle, Castle Ardenvale's shape), all Full. Registry: `saddle-crew-power-bonus` becomes `StatusImplemented` once its `Waiting` is empty.
   - Acceptance tests: each card makes the saddle-and-crew Pilot token; Valor's Flagship cycled for X=2 makes two; a Road enters untapped with a Vehicle in play and tapped without; Cloudspire Coordinator counts a Mount and a Vehicle that entered this turn and not one that entered last turn.

11. **`feat(server): the crew-record and crew-only pilot cards`** (tier:2-standard; after 1, 3, 4 and 9; owner decision 4)
    - Scope: Turtle Van, Getaway Car and Subterranean Schooner (`CrewedThis` targets), Golden Argosy (exile each creature that crewed it, return at the next end step), Luxurious Locomotive (a Treasure per creature that crewed it, counted from `CrewRecordsForEffect` by distinct creature object), Veteran Motorist, Gearshift Ace and Speedway Fanatic (`WhenThisCrewsOrSaddles`), Hotshot Mechanic and Giant Ox (`CostPowerCrew` rules), and the crew-only Pilot token makers Shorikai, Genesis Engine and Prodigy's Prototype. Born to Drive joins them if channel needs nothing new. Registry: a new row is not needed; the cards are catalogued directly.
    - Acceptance tests: one per card for its crew-specific clause, in the style of slices 3 and 4.

`saddlers-of-a-mount` becomes `StatusImplemented` when slices 4 to 6 have landed.

## Owner decisions (2026-10-14)

1. **A creature counts as having crewed or saddled when it is tapped to pay the cost** (CR 702.122b-c, 702.171c), not when the ability resolves. A countered crew or saddle ability still counts its creatures, and "whenever this crews" triggers go above the crew ability. This is Decision 1 and Decision 2.
2. **Negative power with a bonus: add the bonus to the unclamped power, then clamp at zero.** A -1/1 Pilot counts 1. Clamping first would make it count 2, stronger than printed. This is Decision 4.
3. **The table shows who crewed or saddled a permanent.** The Vehicle's or Mount's card view carries the creature IDs (`crewed_by`, `saddled_by`, slice 1), and a tier-1 client slice names them on the permanent (slice 2). It does not touch the crew picker, so it goes straight after slice 1. This is Decision 6. The draft had recommended leaving it out.
4. **Slice 11 is in the plan, as the last slice.** It covers the crew-record and crew-only Pilot cards that are on no waiting list. Luxurious Locomotive counts every distinct creature that crewed it this turn, departed ones included.

## Consequences

- One additive snapshot key, two event kinds, one target clause, one characteristic field, one wire map, two card-view lists and two token templates. No closure route, no schema bump, no fixture change.
- `Card.SaddledBy` becomes a compatibility mirror. It can be dropped at the next schema bump, with its fallback.
- A future "as though its power were greater" for station adds `CostPowerStation` and freezes the counted value on `PaidTap` (Tapestry Warden's last-known ruling).
- **The thing most likely to be got wrong later.** Reading `Card.SaddledBy`, or looking the Vehicle or Mount up on the battlefield, to answer "creatures that crewed/saddled it". Both forget a permanent that has left. Ask the tally by the source's `ObjectRef`.
