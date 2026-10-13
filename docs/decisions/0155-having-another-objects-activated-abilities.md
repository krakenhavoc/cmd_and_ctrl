# ADR 0155 — Having another object's activated abilities: a layer-6 borrow that names the lent row

**Status:** Accepted (owner decisions 2026-10-13) · proposed 2026-10-13 · S53 — Faces, zones, and shared payment (ADR first)
**Issue:** [#1557](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1557) (seam: "has all activated abilities of …").
Registry row: `all-activated-abilities-of` in `server/internal/roadmap/registry.go`, already pointed at #1557.
**Numbering:** assigned by the S53 orchestrator. `origin/develop` (`d66aec82`) has no `0155`, and no remote
branch carries a `0155-*` file.
**Builds on:** [ADR 0093](0093-abilities-granted-to-other-permanents.md) (layer-6 grants, the composed ability
key, ability refs, Decision 3's removal pre-sort, Decision 10, which sent this case here),
[ADR 0067](0067-layer-dependency-ordering.md) (CR 613.8 ordering), [ADR 0046](0046-layer-6-authoritative.md)
(layer 6 is authoritative), [ADR 0041](0041-game-persistence.md) phase 3 (`ScopedEffect` mods and P9's
`AbilityRef` on stack items), [ADR 0066](0066-granted-cast-and-play-permissions.md)'s 2026-10-02 and 2026-10-07 amendments
(`AnyColorSpendStatic`, `Card.ExiledWith`), [ADR 0071](0071-designations-that-switch-abilities-on.md) (designation-gated rows).

---

## Context

### The rules

- **CR 113.10:** effects can add abilities to objects. **CR 613.1f:** "Ability-adding effects, keyword
  counters, ability-removing effects" apply in layer 6.
- **CR 602.1 / CR 605.1a:** an activated ability has a cost and an effect. A mana ability is an activated
  ability that meets 605.1a's criteria. So "all activated abilities" includes mana abilities. Manascape
  Refractor and Mirran Safehouse exist for exactly that. Loyalty abilities are activated abilities too
  (CR 606), which is why Scheming Fence has to say "except for loyalty abilities".
- **CR 113.6:** an ability works only in the zone it says it works in. Necrotic Ooze gets a lent cycling or
  unearth ability, but can never use it from the battlefield.
- **CR 201.5b:** "If an ability of an object refers to that object by name, and an object with a different
  name gains that ability, each instance of the first name in the gained ability that refers to the first
  object by name should be treated as the second name." Quicksilver Elemental's reminder text says the
  same thing.
- **CR 607.5:** "If an object acquires a pair of linked abilities as part of the same effect", they stay
  linked to each other on the new object.
- **CR 613.8a/b:** dependency. A borrow depends on every layer-6 effect that adds or removes abilities on
  one of its lenders, because that effect changes what the borrow does. If two borrows lend to each other,
  that is a loop and timestamp order decides.
- **CR 608.2h / CR 611.2c:** for an effect created by a resolving ability ("gains … until end of turn"), the
  lent set and the affected set are fixed once, when the effect begins. Quicksilver Elemental keeps what it
  gained after the target leaves.
- **CR 113.7a:** an activated ability on the stack exists apart from its source. A lender leaving later does
  not change an ability that has already been activated.

### The cards

The Scryfall dump (`data/scryfall/default-cards.json`, refreshed 2026-09-24) was scanned for "has / have /
gains … activated abilit(y|ies) of", deduplicated by oracle ID, with tokens and art cards skipped. That
finds **27 cards. 26 are commander-legal**. Vazal, the Compleat is not. One of the 26 is catalogued: Hazel's
Brewmaster, with a caveat for this exact clause (`cards/effects/hazels_brewmaster.go`). So this seam
**unblocks 25 new cards and lifts 1 caveat**. The registry's `Unblocks: 1` and its three-card Waiting list
are stale.

| Where the lent abilities come from | Duration | Cards |
|---|---|---|
| Permanents on the battlefield, read through a predicate | static | Marvin, Murderous Mimic; Experiment Kraj; Robaran Mercenaries; Drana and Linvala; Manascape Refractor; Sharkey, Tyrant of the Shire |
| Cards in graveyards, read through a predicate | static | Necrotic Ooze; Mirran Safehouse; Thranduil, the Elvenking; Trazyn the Infinite |
| Cards exiled with the source (CR 607.2a), sometimes also needing a counter | static | Dark Impostor; Patchwork Crawler; Myr Welder; Hazel's Brewmaster (to Foods); Agatha's Soul Cauldron (to creatures with +1/+1 counters); Steward of the Harvest (to creatures); Territory Forge; Rex, Cyber-Hound (brain counters); Mairsil, the Pretender (cage counters, each once a turn); The Enigma Jewel // Locus of Enlightenment (craft materials, each once a turn) |
| The top card of your library, while revealed | static | Conspicuous Snoop; Skill Borrower |
| A permanent chosen as it enters | static | Scheming Fence |
| A target, fixed at resolution | until end of turn | Quicksilver Elemental; Havengul Lich; Grell Philosopher (to each Horror) |

Seven of these also let you spend mana "as though it were mana of any color" (Drana and Linvala, Manascape
Refractor, Scheming Fence, Agatha's Soul Cauldron) or "of any type" (Sharkey), or let you spend blue mana
that way (Quicksilver Elemental, Grell Philosopher). Three also stop the lender's own abilities: Drana and
Linvala, Sharkey, and Scheming Fence.

Out of the family: The Questing Role's "has all the abilities of Questing Beast" (ADR 0093's 2026-10-08
amendment). That is every ability, not only activated ones, and the gap there is the bundle's missing
block-rule slot.

### What the engine has today (checked on `develop` `d66aec82`)

- **One accessor per kind of ability row.** `activatedAbilityRows` and `manaAbilityRows`
  (`game/granted_abilities.go`) are the only bodies behind `ActivatedAbilitiesForCard`,
  `ActivatedAbilitiesWithOrigins`, `ManaAbilitiesForCard` and `ManaAbilitiesWithOrigins`. They return the
  object's own rows, then its intrinsic CR 305.6 land rows (mana only), then each layer-6 grant's rows, with a
  parallel `AbilityOrigins` list. The legal enumerator (`legal/abilities.go`), `MayActivate`
  (`game/any_player_activation.go`), the activation verbs, the auto-tapper, the views and the bot all read
  through them. A new half added there reaches every reader.
- **A grant is a bundle key.** `Characteristic.GrantedAbilities []GrantedAbility{Key, Source, Activator}` is
  written by `staticContinuousEffect.Apply` from `StaticAbility.GrantAbilities` (`game/layers.go`).
  `catalogDef` merges each bundle into the composite `CatalogAbilityKey`, and that is why a grant's triggers
  and 7c statics also work. A borrow has no bundle. The lent set is "whatever these other objects have right
  now". That is the gap the issue names.
- **Refs.** Each row has a stable ref: `own:<i>`, `land:<colour>` or `grant:<bundle>:<i>:<n>`. An activation
  whose ref no longer matches its index is refused with `ErrStaleAbilityRef` and pays nothing (ADR 0093
  Decision 5, the #544 rule).
- **Stack items name a catalog row, not a source.** `activatedAbilityRefFor` (`game/ability_ref.go`) stamps
  the item with `AbilityRef{Key, Slot, Ref, Name}`, where `Key` is the catalog key whose list holds the row.
  Restore and the `catalog/activated` body read that key through the catalog alone, with no source object.
  An instance-carried row (a Food's ability) is not stamped and stays an unkeyed closure, which the census
  counts.
- **Layer 6.** `applyBucketLocked` (`game/layer_dependency.go`) runs ADR 0093 Decision 3's
  `orderRemovalsBeforeTheirVictimsLocked`, then applies the bucket in timestamp order. The layer pass then
  runs `enforceCantHaveLocked` and `grantedStaticEffectsLocked` (`game/layers.go`, `layerPassWithLocked`).
  ADR 0067's trial-application dependency ordering is on for layer 4 only (`dependencyOrderedLayers`).
  `layerPassState.removedAt` records the bucket in which each object lost its abilities, but not the
  removal's timestamp.
- **Invalidation.** Any graveyard change bumps the layer version unconditionally (`layer_listener.go`, #1117).
  Leaving exile bumps it only while some static declares `DependsOnExile`. Nothing bumps it when the top of
  a library changes.
- **Exile links.** `Card.ExiledWith` (`game/card.go`) is CR 607.2a's object link. Only replacement-effect
  exiles stamp it (Valgavoth). Exiles made by a resolving ability (Duplicant, Angel of Serenity) are found by
  `b27ExiledWith` (`cards/effects/batch27_helpers.go`), which walks `g.Events`. That is fine for one
  resolution, and far too slow to run on every layer pass.
- **Mana-spend riders.** `AnyColorSpendStatic{Label, Whose, Covers func(ManaSpendContext) bool}`
  (`game/spend_any_color.go`) is read at payment, and `Covers` narrows it. `ManaSpendContext`
  (`game/mana_restriction.go`) carries the object's types and colours, but not which ability row is being
  paid for.
- **Duration effects as data.** ADR 0041 phase 3's `ScopedEffect` with the `grantAbilities` mod
  (`game/scoped_effects.go`) is the precedent for a resolved grant that keeps the table a restore point.

## Decisions

### 1. A borrow is a declaration on a layer-6 static, beside `GrantAbilities`

```go
// on game.StaticAbility (Layer6Ability only)
Borrows *AbilityBorrow

type AbilityBorrow struct {
    // Lenders lists the objects whose activated abilities are lent, read on every layer pass.
    // The recipients are the static's AppliesTo, as with any grant.
    Lenders func(g *Game, source *Card) []*Card
    // Except leaves out a kind of row: BorrowExceptLoyalty (Scheming Fence),
    // BorrowExceptMana (Sharkey). Zero lends everything.
    Except BorrowExcept
    // EachOncePerTurn is Mairsil's and Locus's "You may activate each of those abilities only
    // once each turn".
    EachOncePerTurn bool
    // Label is the printed clause, for the wire and the ability rows.
    Label string
}
```

This follows the reason ADR 0093's PR 1 amendment gave for `GrantAbilities`. The engine has to know the
static's source, and a test has to be able to read the declaration. `Register` refuses `Borrows` on any
layer except 6. It also refuses it inside a granted bundle, because a bundle that lends abilities would put a
borrow inside a borrow. The card side gets constructors in the `GrantAbilities` style:

| Printed | Constructor |
|---|---|
| "~ has all activated abilities of [permanents matching P]" | `effects.HasActivatedAbilitiesOf(BattlefieldLenders(P))` |
| "… of all [cards matching P] in [your / all] graveyards" | `effects.HasActivatedAbilitiesOf(GraveyardLenders(whose, P))` |
| "… of all cards exiled with ~" (plus a counter, optionally) | `effects.HasActivatedAbilitiesOf(ExiledWithLenders(P))` |
| "[Recipients] have all activated abilities of …" | `effects.GrantActivatedAbilitiesOf(appliesTo, lenders)` |
| "… of the top card of your library" | `effects.HasActivatedAbilitiesOf(TopOfLibraryLender(P))` |
| "… of the chosen permanent" | `effects.HasActivatedAbilitiesOf(ChosenPermanentLender())` |

The `Lenders` closure is catalog code and is never stored. A static borrow persists nothing, just like a
static grant.

### 2. What a borrow writes: one entry per lent ROW, named by the catalog row it came from

```go
// on game.Characteristic — layer-6 output, rebuilt every pass, NOT composed into CatalogAbilityKey
BorrowedAbilities []BorrowedAbility

type BorrowedAbility struct {
    Slot   string    // AbilitySlotActivated or "mana"
    Kind   LentKind  // LentCatalog, LentLand, LentInstance
    Key    string    // LentCatalog: the catalog key whose list holds the row — a card's own key
                     // (CatalogKey, face-aware) or a bundle's GrantKey
    Index  int       // LentCatalog / LentInstance: the row's index in that list (declared, pre-gating)
    Color  string    // LentLand: the CR 305.6 colour
    Lender uuid.UUID // the object it was read off this pass: wire label, instance rows, 201.5b
    Source uuid.UUID // the borrowing static's source (granted_by)
    Activator uuid.UUID
    Borrow string    // the static's identity (source + label), for EachOncePerTurn and the riders
}
```

Three reasons it is shaped this way:

- **It is not composed into `CatalogAbilityKey`.** Only activated and mana rows are lent. Merging a
  lender's whole `CardDef` into the borrower's key, the way a grant merges a bundle, would also lend its
  triggers, its statics and its cost modifiers. A field of its own, read only by the two row accessors,
  keeps every other reader blind to it, and that is the rule.
- **One entry per row, not one per lender.** A borrower has the abilities the lender has *right now*. That
  means the lender's ADR 0071 designation gate (a Class's level rows), its #1859 per-row removals, and
  Except filters such as "except mana abilities" are all settled once, when the pass writes the entry. The
  readers then look up `CatalogActivatedAbilities(Key)[Index]` and nothing else.
- **Named by catalog row, not by lender.** That makes the ref and the stack item (Decision 5) work without
  the lender. It also means two Llanowar Elves in two graveyards are the same row twice, which is CR 113.2c's
  two instances.

`activatedAbilityRows` and `manaAbilityRows` gain a fourth half after the granted half. The order is own,
intrinsic, granted, borrowed. So a borrow appearing never renumbers an existing row. `AbilityOrigin` gains
`Borrowed bool` and `LentBy uuid.UUID` beside `GrantedBy`.

### 3. Reading a lender's rows

- **A lender on the battlefield** is read from its *finished* layer-6 rows, using the same accessors with
  origins. Each origin turns into an entry:
  - `own:<i>` becomes `{LentCatalog, CatalogKey(lender), i}`;
  - `grant:<b>:<i>:<n>` becomes `{LentCatalog, b, i}`;
  - `land:<c>` becomes `{LentLand, c}`;
  - an instance-carried row becomes `{LentInstance, Lender, i}`;
  - a row the lender itself borrowed is copied as it is, so it can be lent again.

  So Marvin gets a Llanowar Elves' mana ability, a Cryptolith Rite grant on a lender, and a Necrotic Ooze's
  lent rows.
- **A lender anywhere else** (graveyard, exile, library) has no layered characteristic. It is read from
  `CatalogKey(card)`, which is face-aware, so a modal double-faced card in a graveyard lends its front face.
  It is also read through the card's designation gate, and through its printed land types for intrinsic
  mana rows. That is how Mirran Safehouse with a Forest card in a graveyard gets "{T}: Add {G}".
- **An uncatalogued lender lends nothing.** The engine does not know its abilities. The borrower's menu shows a disabled "<lender>: not automated" line (owner decision 2).
- **A face-down lender** lends nothing, for CR 708.2a's reason.

### 4. Layer 6 ordering (CR 613.8): borrows run last in the bucket, ordered among themselves

This answers the issue's question about how a borrow interacts with the #1551 removal pre-sort.

1. **Borrows apply after every other layer-6 effect**, as one step at the end of `applyBucketLocked`. They
   run before `enforceCantHaveLocked` and `grantedStaticEffectsLocked`. Every ordinary layer-6 effect that
   touches a lender (a grant, a removal, a keyword) changes what the borrow does, so under CR 613.8a the
   borrow depends on it and waits. This is exact, not a heuristic. The borrow reads the lender's finished
   rows, and that is what dependency asks for.
2. **Among borrows,** B goes before A when B's recipients include one of A's lenders. A cycle is CR 613.8b's
   loop, and timestamp order decides. The relation is read off the borrows' own `AppliesTo` and `Lenders`
   once per pass, as `orderRemovalsBeforeTheirVictimsLocked` reads its own relation. It is not ADR 0067's
   trial application. There are at most a handful of borrows on any board.
3. **A removal on the borrower's SOURCE wins, whatever the timestamps.** This is ADR 0093 Decision 3,
   unchanged. When the borrow step runs, `applyOneEffectLocked` finds the source silenced and skips it.
   Darksteel Mutation on Necrotic Ooze ends the borrow, whichever came first.
4. **A removal on a RECIPIENT that is not the source** keeps timestamp order (CR 613.6), the same as a
   grant. Hazel's Brewmaster lends to Foods. A Food hit by a later "loses all abilities" loses the lent rows.
   One hit earlier keeps them, because the borrow happens after it. Running borrows last would break the
   first case, so `layerPassState` records the latest removal *timestamp* per object beside `removedAt`.
   The borrow step skips any recipient whose recorded removal is later than the borrow's own timestamp.
   This costs nothing on a board with no removal (`st.track`).
5. **Self-borrowers are most of the list.** For them, point 3 decides everything and point 4 never applies.

Two Experiment Krajs each with a counter (or a Kraj and a Marvin) form a loop. The earlier one takes the
other's rows. The later one takes the earlier one's rows, and those include its own rows lent back. Under
CR 113.2c that is two instances of the same ability, which is harmless. The engine keeps both rows (owner
decision 4).

### 5. Refs, the stack and restore

- **Wire refs:**
  - `borrow:<key>:<i>:<n>` for a catalog row;
  - `borrow:land:<colour>:<n>` for an intrinsic land row;
  - `borrow:inst:<lender>:<i>:<n>` for an instance row.

  `<n>` counts earlier entries with the same row. `staleAbilityRef` is unchanged, so a lender that leaves
  between view and announce gives `ErrStaleAbilityRef` and pays nothing. If an identical row is still lent
  from elsewhere (the second Llanowar Elves), the ref still matches, and that is the same ability.
- **The stack item names the lender's catalog row.** `activatedAbilityRefFor` stamps
  `AbilityRef{Key: entry.Key, Ref: own:<i> or grant:<b>:<i>:0, Name}` for a borrowed catalog row. That is
  the row `resolveActivatedAbilityRef` already resolves. So `catalog/activated`, the restore path, the
  Q2/Q3 name check and the boot report work unchanged, and the item does not depend on the lender still
  existing (CR 113.7a). An instance row stays unkeyed and is counted by the census, exactly as the
  lender's own activation would be.
- **No snapshot schema change for static borrows.** `Characteristic` is embedded by value in the snapshot
  as last-known information, so `BorrowedAbilities` gets a `carried` row in `snapshot_drift_test.go`, beside
  `GrantedAbilities`.

### 6. "This creature" is the borrower; names are rewritten for the wire

ADR 0093 Decision 4 holds by construction. A row's `Effect` reads `item.SourceCardID`, which is the
borrower. So `{T}` taps the borrower, "Sacrifice this creature" sacrifices the borrower, the borrower's
controller activates the ability (CR 602.2), and CR 302.6's sickness check reads the borrower. Under CR
201.5b, the lender's name in a row's `Label` is replaced with the borrower's name in the view and the log,
using the lender's printed name, which the entry knows. Linked pairs (CR 607.5) work by construction, because
the engine keys "exiled with this" on the source, and the source is the borrower. Myr Welder's imprint pair,
lent to Necrotic Ooze, links on the Ooze. One known weakness: two lenders with the same linked pair share
one link record on the borrower. That is declared, and no card in the family reaches it.

### 7. The riders

- **"You may spend mana as though it were mana of any color to activate those abilities."** This is an
  `AnyColorSpendStatic` on the borrower with a `Covers` that accepts only an activation of a borrowed row
  from that borrow. `ManaSpendContext` gains `AbilityOrigin` (set by the activation and mana-ability payment
  sites, zero everywhere else). Agatha's Soul Cauldron's version covers "abilities of creatures you control",
  which is a plain `Covers` on the type. Quicksilver Elemental's and Grell Philosopher's "blue mana as though
  any color" is one more `AnyColorSpendStatic.From` colour filter.
- **"Activated abilities of [lenders] can't be activated"** (Drana and Linvala, Sharkey, Scheming Fence) is
  the existing board-wide activation gate (`CanActivateAbilities` / `CanActivateManaAbilities`, with
  Pithing Needle's predicate shape). The gate looks at the *source object*, which is the lender, so the
  borrower's copies are not stopped. That matches the rulings.
- **"Each of those abilities only once each turn"** (Mairsil, Locus) is a per-turn tally keyed by
  `(borrower instance, Borrow, Key, Index)`. `MayActivate` and the mana path both check it. Two identical
  lent rows count as one ability, which matches the Mairsil ruling.
- **`Except`** is applied when the entry is written (Decision 2).

### 8. Lender sets and invalidation

| Lenders | Read from | Bump |
|---|---|---|
| Battlefield predicate (Marvin, Kraj, Robaran, Drana and Linvala, Manascape, Sharkey) | `g.Battlefield`, layers 1-5 settled (control, name, types) | already: every battlefield move, counter and control change bumps |
| Graveyards (Ooze, Safehouse, Thranduil, Trazyn) | each player's graveyard | already: #1117's unconditional graveyard bump |
| Exiled with the source (eight cards) | `Card.ExiledWith == {source, epoch}`, plus counters | `DependsOnExile` (exists), plus a counter-change bump for an exiled card when a static declares it |
| Top of library (Snoop, Skill Borrower) | `Library.Top()` when revealed | **new** `DependsOnLibraryTop`, which follows the `DependsOnHandSize` pattern |
| Chosen as it enters (Scheming Fence) | the as-enters choice record | already: the choice is stamped when the permanent enters |

The exile row needs one engine change. A resolving ability's exile has to **stamp `Card.ExiledWith`**, as a
replacement exile already does, through `ExileLinkedForEffect(source, card)`. Walking `b27ExiledWith`'s event
log on every pass is not an option. The Duplicant family and Angel of Serenity can move onto the stamp in
the same PR. This is a small amendment to ADR 0066's 2026-10-07 amendment, and it is needed before any
"exiled with" lender ships.

### 9. Borrows from a resolving ability are data: a `borrowAbilities` mod

Quicksilver Elemental, Havengul Lich and Grell Philosopher lock in the lent set at resolution (CR 608.2h,
CR 611.2c). That is ADR 0093 PR 4's `grantAbilities` situation, so it is one more `ScopedEffect` mod kind:
`borrowAbilities` (`game.ModBorrowAbilities`, layer 6). Its `Mod.Borrows []LentRow{Slot, Kind, Key, Index,
Color}` is captured from the target when the effect resolves. The adapter writes them as `BorrowedAbility`
entries in the record's timestamp slot, so CR 613.6 timestamp order applies against removals on the
recipient. A `LentInstance` row can't be captured as data, so the mod refuses it. That loses a creature
token's instance-carried ability, which is declared. The kind string is an on-disk identity and the field is
additive within v7, which is ADR 0093 PR 4's precedent. An older binary refuses the file (ADR 0041 P4).
Restore checks each `Key` against the catalog and refuses with `ErrUnknownEffectKey`, the same as
`grantAbilities`.

### 10. Wire and client

- `ActivatedAbilityView` and `ManaAbilityView` already carry `ref` and `granted_by`. A borrowed row also gets
  `borrowed_from: {id, name} | null`, which is the lender. It is `null` for a non-controller viewer when the
  lender is hidden from them. No lender in the current family is hidden: graveyards and the exiles are
  public, and the library top is revealed. Even so, the projection asks the redaction helper rather than
  assuming that. The label is CR 201.5b's rewritten label.
- `GrantedAbilitiesOf` / `granted_abilities` on `CardView` get one line per borrowing static: its `Label`
  ("Has all activated abilities of all creature cards in all graveyards"). There is no per-row text, because
  the rows themselves are in the ability menu.
- **Client:** ADR 0093 Decision 8's left-click rule already covers it. A permanent with a granted or
  borrowed row opens its menu or picker. The menu groups borrowed rows under "From <lender>" and collapses
  identical rows (owner decision 3). `abilityRef.ts` learns the `borrow:` prefix.

### 11. The bot

Nothing new in `aiseat`. The enumerator (`legal/abilities.go`) reads the rows through the accessors, so
borrowed rows are offered, and the `dispatchAll` test holds them to the verbs. ADR 0126's heuristic prices a
row by its `Purpose` and its cost. A lent row brings its `Purpose` with it. The auto-tapper treats a borrowed
mana row like a granted one: a creature's borrowed mana is planned in the last tier (ADR 0093 Decision 6),
and Manascape Refractor's is an ordinary source. No combo planning (Necrotic Ooze loops) is in scope.

## Proposed PR split

1. **The seam, no cards.** `StaticAbility.Borrows`, `Characteristic.BorrowedAbilities`, the fourth half of
   both accessors, the `borrow:` refs, `activatedAbilityRefFor` for borrowed rows, the borrow step in layer 6
   with its own ordering and the removal timestamp, `ManaSpendContext.AbilityOrigin`, the per-turn tally,
   `borrowed_from` and the label rewrite, the drift-test row, and `Register`'s refusals. Tested with fixture
   lenders. Includes the benchmark ADR 0067 asks for: a 40-permanent board, with and without a battlefield
   borrower.
2. **Battlefield and graveyard lenders, the first slice players see** (owner decision 1). Marvin, Experiment
   Kraj, Robaran Mercenaries, Drana and Linvala, Manascape Refractor, Sharkey, Necrotic Ooze, Mirran
   Safehouse, Thranduil, Trazyn the Infinite. Ten cards, plus the client grouping.
3. **Exiled-with lenders.** `ExileLinkedForEffect` and the stamp. Dark Impostor, Patchwork Crawler, Myr Welder,
   Hazel's Brewmaster (caveat lifted), Agatha's Soul Cauldron, Steward of the Harvest, Territory Forge, Rex,
   Mairsil, and Locus of Enlightenment (its "copy each non-mana ability" half uses the existing ability-copy
   seam, and it ships with a caveat if that does not fit). Ten cards, one caveat lifted.
4. **The rest.** The `borrowAbilities` mod (Quicksilver Elemental, Havengul Lich, Grell Philosopher),
   `DependsOnLibraryTop` (Conspicuous Snoop, Skill Borrower), and Scheming Fence. Six cards.

#1557 stays open until PR 4 merges. Each PR moves its cards off the registry row's Waiting list. The row
closes with PR 4.

## Test plan

- `game/borrowed_abilities_test.go` (PR 1, fixture lenders):
  - a lent `{T}` row taps the borrower, is refused while the borrower is summoning sick, and is allowed with
    haste;
  - the borrower's controller activates it, and the lender's controller can't;
  - a lent mana row produces mana, and an intrinsic land row is lent from a land card in a graveyard;
  - a lender leaving between view and announce gives `ErrStaleAbilityRef` with nothing paid, and a second
    identical lender keeps the ref valid;
  - an activated borrowed ability resolves after its lender has left (CR 113.7a);
  - a snapshot taken with it on the stack restores through `catalog/activated`.
- Layer ordering (PR 1):
  - a removal on the borrower's source beats the borrow in both timestamp orders (CR 613.8a);
  - a removal on a non-source recipient keeps timestamp order in both orders (CR 613.6);
  - a grant and a removal on a lender are visible to the borrow whatever the timestamps;
  - two mutual borrowers settle in timestamp order with no growth across passes (CR 613.8b);
  - a token copy of a borrower copies the static, not the lent rows (CR 707.2).
- `legal/borrowed_abilities_test.go`: every enumerated borrowed move dispatches (`dispatchAll`), and a row
  whose lender has gone is not offered.
- Riders: any-colour spend covers only the borrowed rows of that borrow, Drana and Linvala's gate stops the
  lender but not the copy, and Mairsil's once-a-turn counts two identical rows as one.
- `protocol/borrowed_abilities_view_test.go`: `borrowed_from`, the 201.5b label and `ref` for every viewer.
- Card tests per PR, asserting behaviour rather than labels, and one arena smoke run with a Marvin deck to
  confirm the bot activates lent rows without stalling.

## Alternatives considered

- **Compose the lender's key into the borrower's `CatalogAbilityKey`, as a grant does.** This was rejected
  because it would lend triggers, statics, replacement effects and cost modifiers. To stop that, every one of
  roughly thirty readers would need a filter. Decision 2's separate field is read by two functions.
- **One entry per lender (`{Lender uuid}`), resolved live by the readers.** This was rejected for three
  reasons. The readers would have to find a card in another zone on every row read. The lender's designation
  gate and per-row removals would be re-derived at read time. And a stack item or ref naming a lender would
  stop working when the lender leaves, which happens constantly for graveyard lenders.
- **Synthesize a bundle key per lender (`grant:borrow:<oracle>`)** and register a bundle on the fly. This was
  rejected because the catalog is static data checked at boot (`TestEveryGrantKeyResolves`). A bundle made
  at runtime from another card's rows would also bring that card's triggers in through `catalogDef`.
- **ADR 0067 trial application for layer 6.** This was rejected because the dependency is structural ("does
  B touch A's lender") and readable without trials. Turning on O(N²) trials for every layer-6 bucket costs
  every board for the sake of a few cards.
- **Snapshot the lent rows when the borrower enters.** This is wrong under the rules. Static borrows change
  with the lender set (Necrotic Ooze gains a row the moment a creature dies).

## Out of scope

- Granting triggered, static or loyalty-only abilities of another object ("has all the abilities of"). That
  is the Questing Role, and ADR 0093's amendment covers it.
- Vazal, the Compleat (not commander-legal).
- Infinite-loop shortcuts for Necrotic Ooze combos (CR 732). The engine plays them one activation at a time,
  as it does for any other loop.
- Lending from an uncatalogued card (owner decision 2: a disabled "not automated" line only).

## Consequences

- `Characteristic` grows `BorrowedAbilities`, and the layer pass grows one step at the end of the layer-6
  bucket, plus a removal timestamp in `layerPassState`. A board with no borrower pays one nil-slice test per
  accessor call and one bool per pass.
- The ability row accessors have four halves. Every reader is unchanged.
- A resolving ability's exile stamps `Card.ExiledWith` (PR 3), and the effects package's event-log helper
  shrinks.
- `ManaSpendContext` carries the row being paid for. That is useful beyond this seam: "spend this mana only to
  activate abilities of …" can read it too.
- Twenty-five cards and one caveat, across three card PRs.

## Questions for the owner (answered 2026-10-13)

The owner accepted every recommendation on 2026-10-13. Each answer is recorded under its question.

1. **Which slice ships first?**
   - (a) Battlefield and graveyard lenders: Marvin, Necrotic Ooze, Experiment Kraj, Drana and Linvala and six
     more. This needs no new zone link.
   - (b) Exiled-with lenders first: Hazel's Brewmaster's caveat, Rex, Mairsil. This needs the
     `Card.ExiledWith` stamp first.
   - (c) Both in one PR.

   **Recommendation: (a).** It is the issue's named cards and ten cards on the seam alone. (b) follows
   directly.

   **Decision:** (a). Battlefield and graveyard lenders ship first, as PR 2. Exiled-with lenders follow
   in PR 3.
2. **What does a borrower get from a lender the catalog does not automate** (an imported card with no entry)?
   - (a) Nothing. The borrower's menu shows a disabled line, "<lender>: not automated", so the player knows to
     resolve it by hand.
   - (b) A sandbox row, "Activate <lender>'s ability (manual)", that puts a manual stack item naming the
     lender.
   - (c) Nothing, and no line.

   **Recommendation: (a).** It is honest and cheap. (b) invents a manual verb no other uncatalogued ability
   has.

   **Decision:** (a). An uncatalogued lender lends nothing, and the borrower's menu shows a disabled
   "<lender>: not automated" line.
3. **How are borrowed rows presented?**
   - (a) One row per lent ability, grouped under "From <lender>". Identical rows from different lenders
     (two Llanowar Elves in graveyards) collapse to one row in the client.
   - (b) One row per lent ability, ungrouped and uncollapsed, as ADR 0093's working default does for grants.
   - (c) One row per lender, which opens a sub-menu.

   **Recommendation: (a).** Necrotic Ooze in a creature-heavy graveyard can have twenty rows, and grouping
   is what makes that readable. Collapsing is display only: the engine keeps both rows, and the first ref is
   sent.

   **Decision:** (a). Rows are grouped under "From <lender>", and identical rows are collapsed in the client
   only.
4. **Mutual borrowers lend a creature its own abilities back** (two Experiment Krajs, a Kraj and a Marvin).
   - (a) Keep the duplicate rows, as CR 113.2c's two instances. The client's collapse from question 3 hides
     them.
   - (b) Drop a lent row whose catalog row the recipient already has as its own.

   **Recommendation: (a).** It is the rules answer. (b) is visible only to Mairsil's once-a-turn count, where
   it would be wrong.

   **Decision:** (a). The duplicate rows are kept as CR 113.2c instances, and the client's collapse hides
   them.
5. **The "can't be activated" half of Drana and Linvala, Sharkey and Scheming Fence:** ship it with its card
   (a), or leave it as a caveat at first (b)?

   **Recommendation: (a).** The gate exists (Pithing Needle's shape), and the cards are unfair without it.

   **Decision:** (a). The activation gate ships with each of the three cards, with no caveat.
