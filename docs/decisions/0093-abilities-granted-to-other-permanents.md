# ADR 0093 — Abilities granted to other permanents: a layer-6 grant writes a catalog bundle key onto the recipient

**Status:** Proposed · 2026-09-24 · S38 — Layers: dependency ordering, ability grants, ability removal
**Amended:** 2026-09-24 — the owner answered four of the five open questions; see "Owner decisions".
**Issue:** [#754](https://github.com/krakenhavoc/cmd_and_ctrl/issues/754) (this seam — the public roadmap's
top missing seam, `abilities-granted-to-other-permanents` in `server/internal/roadmap/registry.go`)
**Numbering:** swept with the AGENTS.md §4 check on 2026-09-24 — `git fetch origin`, then every
`docs/decisions/009*` file in the history of every remote head (`git log --all --remotes --name-only`,
413 remote branches). The highest number present anywhere is **0092**
(`0092-public-roadmap-and-site-portal.md`); `0093` appears on no branch.

**Related:** [ADR 0046](0046-layer-6-authoritative.md) (layer 6 is authoritative; `CatalogAbilityKey`),
[ADR 0067](0067-layer-dependency-ordering.md) (CR 613.6 / 613.8; the layer-4 dependency pass),
[ADR 0043](0043-copy-effects.md) and its #665 amendment (the `AbilityGrant` bundle and the composite
catalog key this reuses), [ADR 0064](0064-emblems.md) (a second `CardDef` that is not a card's),
[ADR 0083](0083-token-abilities.md) (token templates in the catalog), [ADR 0035](0035-until-end-of-turn-effects.md)
and [ADR 0063](0063-durations-and-control.md) (scoped statics and durations), [ADR 0041](0041-game-persistence.md)
(the continuation census), [ADR 0040](0040-mana-pipeline.md) (which named Chromatic Lantern and
The World Tree as blocked here).

---

## Context

CR 113.10 (pinned edition, effective August 7, 2026):

> Effects can add or remove abilities of objects. An effect that adds an ability will state that the
> object "gains" or "has" that ability, or similar.

CR 613.1f puts ability-adding effects in layer 6. The engine does layer 6 for **keywords only**.
`Characteristic.Abilities` is a `[]string`, and every other kind of ability reaches the engine
through a `Catalog*` hook keyed by the permanent's own catalog key. Checked on `develop` (`97438db2`):

- `ActivatedAbilitiesForCard` (`game/activated.go:623`) returns the instance list or
  `CatalogActivatedAbilities(CatalogAbilityKey(c))`. `ManaAbilitiesForCard` (`game/mutations.go:6055`)
  does the same, plus CR 305.6's intrinsic land abilities. `TriggersForCard`
  (`game/designations.go:296`) is the trigger harvest's one door. None of them can see an ability
  another permanent gave this one.
- The one catalog-level grant, `Spec.Grants` / `AbilityGrant` (#665, `cards/effects/ability_grant.go`),
  is a **copy effect's** "except it has …" clause. It rides `Card.GrantedAbilities`, a *copiable*
  printed field, and `CatalogKey` turns it into a composite key, `"<oracle>|grant:<name>"`, that
  `catalogDef` answers by merging the bundle into the card's definition (`game/copy_grants.go`).
  It is the right lookup and the wrong home for a layer-6 grant, which CR 707.2 says is **not**
  copied. It also has no mana slot.
- Four catalogued cards fake a triggered grant source-side: Dionus, Elvish Archdruid; Agent of the
  Iron Throne; `WardAttached` and `WardGranted` (#626). The granting object carries the trigger and
  watches the recipient. What this gets wrong: who controls the trigger, the recipient losing its
  abilities, and the recipient's last-known information.
- Springleaf Parade puts the mana ability on its own token template. That can't reach a permanent
  that already exists.

### What the printed cards need

A heuristic scan of the Scryfall dump found **664** commander-legal oracle cards with a quoted
`has "…"` / `have "…"` / `gains "…"` grant. It matched the regex on oracle text, deduplicated by
oracle ID and skipped art-series and token placeholders. Classified by the quoted ability:

| Granted ability | Cards | Examples |
|---|---|---|
| triggered only | 248 | Feign Death, Thornbite Staff (half), Dionus |
| activated only | 215 | Necrotic Sliver, Squirrel Nest, Retraction Helix |
| mana only | 118 | Cryptolith Rite, Chromatic Lantern, Gemhide Sliver, Rishkar |
| static only | 81 | "has 'this creature gets +1/+1 for each …'" |
| mixed | 2 | Urza's Saga (mana, then activated) |

By recipient, the common shapes are a **class** ("creatures you control", "all Slivers", "lands you
control", often conditional: The World Tree's six lands, Rishkar's counter), the **attached host**
(Equipment, Auras), and a **target, until end of turn** (about 130 rows). A few are self-grants from
a resolving ability with no stated duration (Urza's Saga's chapters). Thirteen of these cards are
already catalogued by other routes, among them the source-side approximations above, token templates
that print the ability (Awakening Zone, Pawn of Ulamog), and the two #665 copy grants (Phantasmal
Image, Sakashima the Impostor).

The registry's Waiting list, checked card by card:

| Card | Needs | Verdict |
|---|---|---|
| Cryptolith Rite, Chromatic Lantern, Jaheira, Insidious Roots, Great Divide Guide, Rishkar, The World Tree | static mana grant to a class | this ADR |
| Gemhide Sliver | mana grant to **every** player's Slivers | this ADR |
| Urza's Saga | self-grant from a chapter, no duration (mana, then activated) | this ADR (Decision 7) |
| Ultima, Origin of Oblivion | removal + mana grant "for as long as that land has a blight counter" | this ADR, plus a new `DurationCondition` |
| Rhythm of the Wild | **riot** (CR 702.136a), a keyword with a CR 614.12 entry choice | **misattributed** — riot is a keyword grant, which already works; riot itself does not exist |
| Marvin, Murderous Mimic | "has all activated abilities of …" | **deferred** (Decision 10) |

## Decision 1 — A grant is a catalog bundle, and a layer-6 grant writes its KEY onto the recipient's `Characteristic`

We reuse #665's bundle unchanged in shape and extend it with two slots:

```go
type AbilityGrant struct {
    Key       string                 // catalog-wide, "<card>/<what>", as permanent as an oracle ID
    Triggered []game.TriggeredAbility
    Static    []game.StaticAbility   // copy grants only; refused for a layer-6 grant (Decision 10)
    Activated []ActivatedAbility
    Mana      []ManaAbility          // NEW
    Text      string                 // NEW — the quoted ability as printed, for the wire (Decision 8)
}
```

The bundle stays **static catalog data**, registered under `game.GrantKey(Key)` in the same `defs`
map. What is new is the layer-6 output that names it:

```go
// on game.Characteristic
GrantedAbilities []GrantedAbility   // layer 6 output, rebuilt every pass

type GrantedAbility struct {
    Key    string    // game.GrantKey(bundle)
    Source uuid.UUID // the granting object — for the wire label and the stale-ref check, never for rules
}
```

A granting static's `Apply` appends one entry. It works exactly the way a keyword grant appends a
string to `Characteristic.Abilities`: in its timestamp slot of the layer-6 bucket.

Why this and not the alternatives the issue listed:

- **Not closures on `Characteristic` or `Card`.** A `Characteristic` is rebuilt every recompute and
  is never snapshotted, so it could hold closures. But its last-known-information copy is read by
  the dies harvest, and the lookup would need a second, closure-shaped seam beside `catalogDef`.
  A key keeps the single lookup #665 built.
- **Not "grant source key + ability kind + index + timestamp" resolved against the granting card's
  spec.** That makes the recipient's abilities depend on finding the grantor. The grantor is often
  gone by the time the ability matters (a dies trigger; an until-end-of-turn grant from a spell in
  the graveyard). A bundle key is resolvable with nothing but the catalog.
- **Not a second `Card` field beside `Card.GrantedAbilities`.** That field is copiable (CR 707.9a)
  and a layer-6 grant is not (CR 707.2: "Other effects … are not copied"). Keeping the layer-6 grant
  on the `Characteristic` makes the difference structural: `CopiableValuesOf` never reads the
  layered result, so nothing has to remember to exclude it.

Nothing new is persisted for a **static** grant. The grantor's static is catalog code, the
`Characteristic` is re-derived after restore, and the census gains no row. Duration grants are
Decision 7.

## Decision 2 — ONE key: `CatalogAbilityKey` composes the base, the copy grants and the layered grants

`CatalogAbilityKey(c)` is already the single "what does this permanent DO" accessor (ADR 0046 §1)
behind about thirty production call sites, and every ability reader reaches it through `TriggersForCard`,
`ActivatedAbilitiesForCard`, `ManaAbilitiesForCard`, the cost-modifier gather and the rest. So it
becomes the composition:

```
CatalogAbilityKey(c) = base  +  Card.GrantedAbilities (copy grants)  +  effective.GrantedAbilities (layer 6)
where base is "" when the object's own abilities are removed (AbilitiesRemoved) or it is face down
```

and `catalogDef` answers the composite as it does today, merging each bundle. Four changes to the
existing code:

1. **An empty base keeps its layered grants.** `catalogKeyWithGrants` returns `""` for an empty
   base today. It must not, for three real cases: a face-down 2/2 under Cryptolith Rite (CR 708.2a
   silences the card's *text*, not the effects on it), an uncatalogued imported creature (the common
   case — a vanilla bear under Cryptolith Rite), and a permanent whose own abilities were removed
   before the grant arrived (Decision 3). The composite form for these is `"|grant:<key>…"`, and
   `mergedCatalogDef` already handles a nil base lookup.
2. **`mergedCatalogDef` merges `ManaAbilities`.** #665 left mana out on purpose. The reason it gave,
   that mana is read "off the card object as well as the catalog", is Decision 5's job.
3. **`CatalogKey` does not change.** It stays the identity / printed surface: the `auto` badge and
   `game.Unimplemented`, the ETB hook, printed keywords, cast-time slots. A creature with a granted
   ability is not "a catalog card", and the ADR 0037 unimplemented signal must not flip because an
   enchantment is on the table.
4. **The key is composed once per layer pass, not per call.** Cryptolith Rite with twenty creatures
   breaks `mergedCatalogDef`'s assumption: "a handful of permanents in the rarest game … Deliberately
   NOT cached." The pass stores the composed key on the `Characteristic` (unexported), so
   `CatalogAbilityKey` stays a field read. The merged `CardDef` is memoised per game, keyed by the
   composite string and dropped at each recompute. It is not a process-lifetime cache: #665's
   objection about test stubs swapping `CatalogLookup` still holds. The implementation PR carries a
   benchmark, as ADR 0067 did: harvest and view cost for 40 creatures under one grant, against
   `develop`.

The last-known-information harvests read the same composition from the snapshot: `harvestLTB`
already holds `lastKnownBattlefield[id]`, a full `Characteristic`, and the simultaneous-exit batch
holds a `Card` whose `effective` is intact. So a creature that dies under Feign Death has the dies
trigger its snapshot says it had (CR 603.10a). `harvestLTB`'s hand-rolled "CatalogKey plus
AbilitiesRemoved" becomes one helper, `abilityKeyFromLKI(identity, lki)`, and it is the only other
place the composition is written.

## Decision 3 — Removal clears layered grants in its own slot; a later grant survives; a removal on the GRANTOR wins by dependency

**On the recipient (CR 613.6, CR 113.10c).** A `RemovesAbilities` effect already empties
`Characteristic.Abilities` and sets `AbilitiesRemoved` in its timestamp slot. It now also empties
`GrantedAbilities`. A grant sorted after it appends to the emptied slice and survives, which is
exactly how granted keywords behave today (ADR 0046 §2). `AbilitiesRemoved` narrows in meaning from
"has no abilities" to "**its own** abilities (printed, copy-granted, token-template) are gone". Every
reader of `HasLostAllAbilities` / `AbilitiesRemoved` is audited in PR 1. The early `return nil`
in `ActivatedAbilitiesForCard` and the `declared = nil` arm in `ManaAbilitiesForCard` become
"skip the own half, keep the granted half".

**CR 305.7 falls out.** `SetsBasicLandType` removes in **layer 4** (ADR 0067 §3), before any layer-6
grant. So Chromatic Lantern keeps working on a land under Blood Moon, which is the rule's last
sentence ("doesn't remove any abilities that were granted to the land by other
effects").

**On the grantor (CR 613.8a).** A Kenrith's Transformation on Gemhide Sliver removes Gemhide's
static, and so the grant, from **every** Sliver, whatever the timestamps. The grant's *existence*
depends on the removal (CR 613.8a(b)), so the removal applies first. Today the layer-6 bucket orders
a removal against the sources it silences by **timestamp only** ("IN the removal's bucket, timestamp
order decides", `applyOneEffectLocked`). That is already wrong for keyword lords: a Lord of Atlantis
under a later-timestamped removal still grants islandwalk. This ADR fixes it for both, with the
narrowest dependency rule that covers it:

> In the layer-6 bucket, a removal effect is applied before every effect whose **source** it
> applies to. If two removals each apply to the other's source, that is a dependency loop
> (CR 613.8b): timestamp order, which is today's answer for two Song of the Dryads.

This is a structural pre-sort of the layer-6 bucket, keyed on "does removal R apply to source S".
It is not ADR 0067's general trial application. ADR 0067 asked for "a catalogued pair that wants it
and a benchmark" before adding another dependency-ordered bucket: the pair is Gemhide Sliver plus
Kenrith's Transformation (and Lord of Atlantis plus the same), and the benchmark is Decision 2's.
Layer 1-5 effects of a silenced grantor are untouched: ADR 0067 §2 stands.

Other layer-6 dependencies between two grants (a grant whose "applies to" reads another grant's
output) are **not** ordered. Timestamp order applies, as it does today, and no catalogued pair
reaches it.

## Decision 4 — "This creature" is the host, by construction

Every reader walks the **recipient's** key, so a granted ability is handed the recipient as its
`source` everywhere:

- A granted trigger's `AppliesTo(ev, source, …)` sees the host. `NewTriggeredItem(source, …)` makes
  the host's controller the trigger's controller (CR 603.3a, CR 113.8). "When this creature dies"
  compares against the host's ID.
- A granted activated or mana ability's source is the host. `{T}` taps the host, "Sacrifice this
  permanent" sacrifices the host (Necrotic Sliver), and `NewContext(g, item)` resolves "this
  creature deals 1 damage" to the host (Thornbite Staff).
- The activator is the host's controller (CR 602.2: "only an object's controller … can activate its
  activated ability"). `ActivateCatalogAbility` already refuses anyone else with
  `ErrCardCallerMismatch`, so another player's Necrotic Sliver is theirs to use. The grantor's
  controller gets nothing.
- **CR 302.6 is free.** The sickness check reads the source's `SummonedThisTurn` and haste on the
  activation path, and the source is the host. So Cryptolith Rite on a creature cast this turn
  cannot tap for mana, a hasty one can, and Chromatic Lantern's lands are never "sick".
- CR 113.10a activation instructions ("Activate only as a sorcery") are fields on the bundle's
  shapes and come along with it. CR 607.1a (an ability granted within another ability is "printed
  on" the object for linked-ability purposes) needs nothing either: the linked pair lives in one
  bundle and both halves see the same host.

Two things a bundle may not declare, refused by `Register`:

- **`ActiveWhen`.** A designation is the *host's*. Gate the grant, not the granted ability: the
  grantor's own `StaticAbility` takes `ActiveWhen`.
- **`Zones` other than the battlefield.** The layer pass only reaches permanents. Grants to cards in
  a hand or graveyard are Decision 10.

## Decision 5 — Readers return own + intrinsic + granted, in that order; every row carries a stable `ref`

`ActivatedAbilitiesForCard` returns the own list (the instance list or the base catalog entry,
designation-gated as today) followed by the granted bundles' activated abilities, in layer-6 order.
`ManaAbilitiesForCard` returns declared own, then intrinsic (CR 305.6, deduplicated against the own
half only, as today), then granted. An instance list no longer hides the catalog's granted half:
"a grant must hide neither", per the issue.

**Granted abilities are always last**, so a grant appearing never renumbers anything that was
already there. A grant disappearing can still renumber later grants, and the wire indexes are
positional (`ability_index`, `ActivateCatalogAbility(…, index, …)`, `ActivateManaAbility(…, abilityIdx, …)`).
So every row gets a **ref**: `own:<i>`, `land:<colour>` or `grant:<bundle>:<i>:<n>` (the nth
occurrence of that bundle on this object).

- `ActivatedAbilityView`, `ManaAbilityView` and `legal.Move` carry `ref` beside `index`.
- The client and the bot send it back.
- The engine checks it against the row at `index` **before** anything is validated or paid. A
  mismatch is a new `ErrStaleAbilityRef`, and it costs nothing.
- An absent ref is accepted as today, for one release of client skew.

This is the #544 rule. A stale move must be **refused**, never silently fired on whichever ability
now sits at that index. And a refusal after the board changed is not a wedge, because the next
enumeration is fresh. It also closes a latent version of the same bug: a Class levelling up can
already shift a designation-gated own list between view and announce.

The accessors being the choke point is what makes the rest free. The legal enumerator
(`legal/abilities.go`), `CanActivateAbilities` / `CanActivateManaAbilities` (Arrest keeps working on
granted abilities, because they are the host's), the board-wide activation gate (Pithing Needle,
Linvala, Cursed Totem: the source is the host), `producible_mana.go`'s CR 106.7 "could produce"
(Exotic Orchard sees a Lantern'd land's any-colour), the cost-modifier readers, the view and the bot
all read through them.

**Identical grants are not merged.** Two Cryptolith Rites give a creature two instances of the
ability (CR 113.2c allows multiple instances). For a mana ability this is harmless; for a trigger it
is the rules (two Dionuses, two untaps). One known weakness: two instances of the same bundle share a
`(source, label)` tally key, so "only once each turn" on a doubled granted trigger is counted
together. That is weaker than printed, no catalogued card reaches it except a doubled Dionus, and it
is declared, not modelled.

## Decision 6 — The auto-tapper plans every plannable ability of a permanent as mutually exclusive alternatives

`autoTapAbilityFor` takes the **first** plannable mana ability of each permanent. Under Cryptolith
Rite, a Llanowar Elves would only ever be planned for `{G}`, and a Forest under Chromatic Lantern
only for `{G}`. #779 already gives the solver mutually exclusive candidates for one source (one per
colour of "N mana of any one colour"). PR 2 extends that to one candidate set **per plannable
ability**, so the solver may pick the granted any-colour ability when the cost needs it.

Until PR 2 lands, the gap is declared at the ADR level, as ADR 0074 §7 declared the triggered-mana
surplus: the planner may fail to find a payment that exists. That is weaker than printed and safe,
and a manual tap always works.

**Creatures are planned last (owner decision, 2026-09-24).** Cryptolith Rite turns every creature
into a mana source, and an auto-paid cast must not tap the creatures the player meant to attack or
block with. So a **creature's granted** mana ability is planned only as a last resort. It comes after
every land, mana rock and other ordinary source, in the same late tier #1215 gave Treasures and other
sacrifice-cost sources. A granted mana ability on a noncreature permanent (a Forest under Chromatic
Lantern) is an ordinary source and is planned normally. The per-ability candidate fix above stays in
PR 2 either way.

## Decision 7 — Grants from a resolving spell or ability are DATA: `ScopedGrant`

Until-end-of-turn grants (Feign Death, Retraction Helix, Fake Your Own Death), "for as long as"
grants (Ultima) and no-duration self-grants (Urza's Saga's chapters, CR 611.2a) are continuous
effects created by a resolution. Today that means a `ScopedStatic`, whose `Ability` is two closures.
That makes the snapshot census non-empty for as long as the effect lives: Urza's Saga would block
restore points for three turns, and an indefinite grant for the whole game.

A grant is fully described by data, so it gets its own registry:

```go
type ScopedGrant struct {
    Key       string       // game.GrantKey(bundle)
    Affected  []ObjectRef  // CR 611.2c: the set, locked at creation — {instance, entered-at}
    Source    ObjectRef    // for the wire label; LKI is enough
    Timestamp int64        // CR 613.7b
    Duration  Duration     // ADR 0063: plain data, one expiry function
    Label     string
}
```

`activeStaticAbilitiesLocked` adapts each live `ScopedGrant` into an engine-built layer-6 effect.
This is the pattern `controlStatic` uses for layer 2, and it sorts by timestamp with everything
else. `ClearExpiredScopedStaticsLocked` sweeps `ScopedGrant` through the same
`durationExpiredLocked`. The snapshot carries it, the schema version is bumped, and the drift test
classifies every field as `carried`. The census gains no row. A grant's object set is pinned
(ADR 0063 Decision 6), so a flickered target is a new object without it (CR 400.7). Ultima's "for as
long as that land has a blight counter" needs one more `DurationCondition`, on the pinned object's
counters, added in PR 4.

**Amended 2026-09-24 (#1497, owner decision 1 on ADR 0041's phase 3 amendment):** `ScopedGrant` is
not a registry of its own. It is the `grantAbilities` mod of ADR 0041 phase 3's `ScopedEffect`
record (`server/internal/game/scoped_effects.go`), which landed first and already gives it the pinned
set, the timestamp, the `Duration`, the sweep, the layer-pass adapter and `carried` in the drift
plan. PR 4 adds the `grantAbilities` kind to that vocabulary; see
[ADR 0041](0041-game-persistence.md), "Amendment, 2026-09-24 — phase 3: effects as data".

**Not a delayed trigger.** The issue offered #663's event-conditioned delayed trigger as an
alternative route for "until end of turn, target creature gains 'When this creature dies …'". It is
the wrong model. A granted trigger is an **ability of the creature**: a later Darksteel Mutation
removes it, it is gone if the creature phases or flickers, and it goes on the stack under the
creature's controller. A delayed trigger has none of those properties.

## Decision 8 — The wire: granted rows are labelled, and the bundle's text ships

- `ActivatedAbilityView` / `ManaAbilityView` gain `ref` (Decision 5) and
  `granted_by: {id, name} | null`.
- `CardView` gains `granted_abilities: [{text, source_id, source_name}]`. This covers granted
  **triggers**, which have no row and would otherwise be invisible. `AbilityGrant.Text` is the source
  of the text, for ADR 0083 Decision 6's reason: a trigger's `Label` is a log line, and the recipient
  has no printing that says it. #665 recorded this gap for copy grants ("A granted ability's TEXT is
  not projected onto the wire"); copy grants pick it up too.
- All of it is public. The grantor is on the battlefield or is a spell that resolved in front of
  everyone.

**Click behaviour (owner decision, 2026-09-24).** A permanent that has a granted mana or activated
ability opens its ability menu or picker on **left-click**. This extends #368's rule (a land with a
non-mana activated ability is clicked for that ability) to granted abilities on any permanent. The
sandbox tap stays in the context menu. There is **never a silent default** between two mana
abilities:
- a creature under Cryptolith Rite opens the picker;
- a Birds of Paradise under the Rite, with its own ability and the granted one, opens the picker;
- a Forest under Chromatic Lantern opens the picker;
- a land enchanted by Squirrel Nest opens its menu.

The client derives this from `granted_by` on the rows. The server needs no new flag.

**Rows (working default, not decided).** Until the owner answers the open question below, each
grantor gets one row, labelled with the grantor's name: two Cryptolith Rites mean two rows,
"from Cryptolith Rite". There is no board-level marker.

`aiseat` needs nothing. It reads the view and the move list, and still imports nothing from
`internal/game`.

## Decision 9 — Card-side vocabulary

Declared on the granting card, in the vocabulary a printed ability already uses:

```go
Grants: []AbilityGrant{{
    Key:  "cryptolith-rite/any-color",
    Mana: []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{W|U|B|R|G}", Label: "Add one mana of any color"}},
    Text: "{T}: Add one mana of any color.",
}},
Static: []game.StaticAbility{
    GrantAbilities(creaturesYouControl, "cryptolith-rite/any-color"),   // a class (the AppliesTo shape)
},
```

| Printed | Constructor |
|---|---|
| "Creatures you control have …", "Lands you control have …", conditional classes | `GrantAbilities(appliesTo, key)` |
| "All Slivers have …", "Sliver creatures you control have …" | `TribalAbilityGrant(TribeFilter{…}, key)` — the `TribalKeywordGrant` twin |
| "Equipped creature has …", "Enchanted land has …" | `GrantAbilitiesToAttached(key)` — the `GrantToAttached` twin |
| "Until end of turn, target creature gains …" | `GrantAbilitiesUntilEOT{Targets, Key}.Apply(ctx)` |
| "… gains …" for another duration, or none | `GrantAbilitiesForDuration{Targets, Key, Duration}.Apply(ctx)` |
| "loses all abilities and has …" (Ultima) | `LoseAllAbilities().AndGrant(key)` — one effect, one timestamp, as ADR 0046 §2 requires |

`TestEveryGrantKeyResolves` (the `TestEveryTokenKeyResolves` twin) fails the build if a constructor
names a key no card registered, or if a layer-6 grant names a bundle with a `Static` slot.

## Decision 10 — Out of scope, stated

- **Granted static abilities** (81 cards). The layer pass gathers every static *before* it walks
  layer 1, so a static that only exists after layer 6 is never gathered. Doing it properly needs a
  second gather after layer 6, restricted to layer 7. A granted static that would apply in layers 1-5
  is a CR 613 question this engine should not open for 81 mostly-niche cards. `Register` refuses the
  combination until an ADR amendment does it.
- **Granted replacement effects and the CR 614.12 look-ahead**, including riot (Rhythm of the Wild).
  Riot is a keyword, so granting it already works once the keyword exists. What is missing is riot
  and "abilities it would have on the battlefield" as it enters. PR 1 moves Rhythm of the Wild off
  this seam in the registry.
- **"Has all activated abilities of …"** (Marvin, Necrotic Ooze, Drana and Linvala, Hazel's
  Brewmaster's caveat). The grant's content is another object's text, computed each pass, and it has
  no bundle key. This needs its own ADR. PR 1 moves Marvin to a seam of its own.
- **Granted triggered MANA abilities** (CR 605.1b on a recipient) and **granted loyalty abilities**.
  `AbilityGrant` gets no `ManaTriggers` slot. No card on the Waiting list needs either.
- **Grants to cards off the battlefield** ("creature cards in your graveyard have …"). Cast
  permissions are ADR 0066's; anything else waits for a card.
- **Grants to players.** Player keywords are ADR 0072's.
- **CR 113.11 "can't have".**
- **Migrating `WardGranted` / `WardAttached`.** Ward's cost is a parameter a bundle can hold, so it
  *could* move. The observable difference (who controls the ward trigger, and a warded recipient
  that later loses its abilities) is small, and #626 declared it. It is a follow-up, not a
  prerequisite.

## Consequences

- One lookup seam, still. #665's composite key becomes the home of every granted ability, copy or
  layer 6. The difference between the two is where the name lives: `Card` (copiable) or
  `Characteristic` (layered).
- `Characteristic` grows `GrantedAbilities` and a composed key. `ScopedGrant` is a new persisted
  registry, with a snapshot schema bump and drift-test rows.
- `AbilitiesRemoved` changes meaning to "own abilities removed". Every reader is audited in PR 1.
- **A behaviour change on existing cards.** A layer-6 removal on a keyword lord now strips the lord's
  grant from everything it applied to, whatever the timestamps (Decision 3). That is rules-correct
  and is pinned by a test.
- The wire gains `ref`, `granted_by` and `granted_abilities`. The activate verbs gain an optional
  `ref`.
- The coverage guard learns the mechanic. A caveat naming "gains" / "has '…'" on a card that now
  declares a layer-6 grant fails the build, as with every other mechanic probe.

## Proposed PR split

1. **The seam, no cards.** `Characteristic.GrantedAbilities`; the composed `CatalogAbilityKey`
   (empty-base, face-down, LKI); `mergedCatalogDef` merges mana; the memo and the benchmark;
   removal clears grants; the layer-6 grantor-dependency pre-sort (with the Lord of Atlantis
   regression test); the `AbilitiesRemoved` reader audit; readers returning own + intrinsic +
   granted; `ref` on the views, `legal.Move` and both activate verbs; `ErrStaleAbilityRef`;
   `granted_by` / `granted_abilities` on the wire; `AbilityGrant.Mana` / `.Text`;
   `TestEveryGrantKeyResolves`; registry moves for Rhythm of the Wild and Marvin. Tested with
   fixture bundles.
2. **Static and attached grants, mana and activated — the first slice players see (owner decision,
   2026-09-24).** The constructors, the auto-tapper's per-ability candidates with creatures' granted
   mana in the last tier (Decision 6), and client rendering with the left-click picker (Decision 8). Cards: Cryptolith Rite, Chromatic
   Lantern, Gemhide Sliver, Manaweft Sliver (lifting `batch31_test.go`'s hold), Rishkar, Jaheira,
   Insidious Roots, Great Divide Guide, The World Tree, Paradise Mantle, Necrotic Sliver, Squirrel
   Nest. Springleaf Parade moves off its token-template workaround.
3. **Granted triggers.** LKI dies triggers through `abilityKeyFromLKI`; Thornbite Staff (both
   halves); and Dionus and Agent of the Iron Throne moved off their source-side approximations onto
   the seam (owner decision, 2026-09-24). Tests pin the two behaviour changes: after a control
   change the trigger belongs to the recipient's controller, and a recipient whose own abilities are
   removed later loses the grant.
4. **Duration grants.** `ScopedGrant` with persistence and the drift test;
   `GrantAbilitiesUntilEOT` / `ForDuration`; the blight-counter `DurationCondition`. Cards: Feign
   Death, Fake Your Own Death, Malakir Rebirth (the spell face), Retraction Helix, Urza's
   Saga, Ultima.

#754 stays open until PR 4 merges, per the claim comment.

## Test plan (across the PRs)

- `game/granted_abilities_test.go`:
  - A grant on an uncatalogued bear and on a face-down 2/2.
  - Removal before and after a grant, in both orders (CR 613.6).
  - Blood Moon plus a Lantern-style grant (CR 305.7).
  - A removal on the grantor strips the grant whatever the timestamps, and two mutual removals fall
    back to timestamp order (CR 613.8a/b).
  - The same fixture on a keyword lord.
  - A token copy and a Clone of a granted creature do not copy the grant (CR 707.2).
  - A granted `{T}` ability refused on a summoning-sick host and allowed with haste (CR 302.6).
  - Another player's Sliver activated by its controller only.
  - A granted ETB trigger fires for a creature entering under an existing grant (CR 603.6a).
  - A granted dies trigger fires from LKI in a single death and in a simultaneous wipe (CR 603.10a).
- `game/granted_abilities_ref_test.go`: a grant vanishing between view and announce is
  `ErrStaleAbilityRef` with nothing paid, and an absent ref is accepted.
- `legal/granted_abilities_test.go`: every enumerated granted move dispatches (`dispatchAll`), and a
  removed grant is not offered (#544).
- `game/autotap_granted_test.go` (PR 2):
  - Llanowar Elves under Cryptolith Rite pays `{U}`.
  - A cost that lands, rocks and Treasures can pay taps no creature for granted mana.
  - A creature's granted mana is used only when nothing else can pay.
  - A Forest under Chromatic Lantern is planned as an ordinary source.
- Client (PR 2): left-click on a permanent with a granted mana or activated ability opens the menu
  or picker, including a creature under Cryptolith Rite, a Birds with both kinds and a Squirrel
  Nest land; the sandbox tap is still in the context menu.
- Card tests for Dionus and Agent of the Iron Throne (PR 3): the trigger's controller after the
  recipient changes control, and the grant lost to a later removal on the recipient.
- `game/scoped_grant_test.go`: until-end-of-turn expiry at cleanup, the pin on a flicker, and a
  snapshot round trip with an empty census (PR 4).
- `protocol/granted_abilities_view_test.go`: `granted_by`, `ref` and `granted_abilities` on every
  viewer.
- Card tests per PR, asserting behaviour rather than labels.

## Owner decisions (2026-09-24)

The owner answered four of the five open questions this ADR was proposed with. The ADR stays
**Proposed** until the owner accepts it.

1. **Slice 1.** PR 2 as proposed. Mana and activated static grants (about a dozen cards) are the
   first slice players see, and triggers follow in PR 3.
2. **Click behaviour.** A permanent with a granted mana or activated ability opens its ability menu
   or picker on left-click. This is #368 extended to granted abilities, and the sandbox tap stays
   in the context menu. It covers a permanent with its own and a granted mana ability (always the
   picker, never a silent default) and Squirrel Nest's land. See Decision 8.
3. **Auto-tapper.** A creature's granted mana abilities are planned only as a last resort, in the
   late tier #1215 gave Treasures. The per-ability fix stays in PR 2. See Decision 6.
4. **Dionus and Agent of the Iron Throne.** They move onto the seam in PR 3, with tests pinning the
   changed control and removal behaviour.

## Open questions for the owner

This is a product decision the code and the rules do not settle. The ADR does not answer it.

1. **Presentation of granted abilities.** Should identical grants from two sources (two Cryptolith
   Rites) show as one row or two? Should a permanent carrying a granted ability get a board-level
   marker, or is the labelled row in its menu enough? *Working default until decided:* one row per
   grantor, labelled with the grantor's name, and no board marker (Decision 8).
