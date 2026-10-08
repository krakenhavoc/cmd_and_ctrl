# ADR 0101 — Keyword counters read by the engine

**Status:** Accepted · 2026-09-30 · Post-S30 — Rolling deck-driven catalog growth
**Issue:** [#1753](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1753). It relates to the Betor deck re-check on #1117, where Perennation is the last seam-blocked card.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-09-30. I ran `git fetch --all --prune`, then read every `docs/decisions/` file name on every remote branch (41 heads). The highest number on any branch is **0098**. Numbers 0099, 0100 and 0102 are held for ADRs being written at the same time, so this one takes **0101**.
**Owner decisions:** 2026-09-30. All five recommendations were accepted; see Owner decisions below.
**Amendments:** 2026-10-08, exalted as a keyword and its counter ([#2538](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2538)). Accepted (owner answers 2026-10-08); see Amendments at the end.
**Builds on:** [ADR 0014](0014-combat-keywords.md) (the closed keyword table), [ADR 0046](0046-layer-6-authoritative.md) (layer 6 is the one ability list), [ADR 0067](0067-layer-dependency-ordering.md) (CR 613.6 silencing), [ADR 0038](0038-protection-style-keywords.md)'s 2026-09-28 amendment (the "can't have" strip), and [ADR 0041](0041-game-persistence.md) (the snapshot shape rule).

---

## Context

### The gap

CR 122.1b: "A keyword counter on a permanent or on a card in a zone other than the battlefield causes that object to gain that keyword." The engine reads no keyword counters of its own. A `deathtouch` entry in `Card.Counters` is storage. It grants nothing unless a card on the battlefield carries a static that says so.

### The workaround, and where it is wrong

Five catalogued cards carry the rule themselves with `b24KeywordCounterGrant(kw)` (`cards/effects/batch24_helpers.go`). It is a layer-6 static on the card that PLACED the counters, granting `kw` to every **creature** with a `kw` counter.

| Card | Counter | Status today |
|---|---|---|
| Vraska Joins Up | deathtouch, on each creature you control | `CompletenessCaveats`: the counters stop working once Vraska leaves |
| Sorin of House Markov | lifelink, from his −6 | `CompletenessCaveats`: the same gap once Sorin leaves |
| Tekuthal, Inquiry Dominus | indestructible, on itself | `CompletenessFull` |
| Solphim, Mayhem Dominus | indestructible, on itself | `CompletenessFull` |
| Drivnod, Carnage Dominus | indestructible, on itself | `CompletenessFull` |

The roadmap row calls the three Dominus cards exact because the counter and the grant sit on the same permanent. They are exact on an ordinary board, but not in three corners:

1. **The grant is an ability of its source, and the counter is not.** If Tekuthal loses all abilities to an effect with an EARLIER timestamp than the counter, the counter should still grant indestructible (CR 613.7c, below). The static is silenced by CR 613.6, so the counter does nothing.
2. **The grant has the source's timestamp, not the counter's.** CR 613.3 orders layer 6 by timestamp. A "loses indestructible" effect that arrived between Tekuthal entering and the counter being placed should lose to the counter. It sorts after the grant instead.
3. **Creatures only.** `b24KeywordCounterGrant` checks `target.IsCreature()`. Perennation returns any permanent card with a hexproof and an indestructible counter. An indestructible counter on an enchantment would do nothing.

Perennation has no permanent to carry the grant at all, so it stays out of the catalog.

### What the rules say

I checked every rule below against the Comprehensive Rules text in the pinned edition.

- **CR 122.1b.** The keywords a keyword counter can be are "flying, first strike, double strike, deathtouch, decayed, exalted, haste, hexproof, indestructible, lifelink, menace, reach, shadow, trample, and vigilance, as well as any variants of those keywords". Protection is not on the list. Neither are shield or stun counters, which are CR 122.1c and 122.1d.
- **CR 613.1f.** "Layer 6: Ability-adding effects, keyword counters, ability-removing effects, and effects that say an object can't have an ability are applied." A keyword counter is applied in layer 6, in the same bucket as grants and removals.
- **CR 613.3.** Within layers 2–6, characteristic-defining abilities apply first, then all other effects in timestamp order. A keyword counter is not a CDA, so it takes its place by timestamp.
- **CR 613.7c.** "Each counter receives a timestamp as it's put on an object or player. If that object or player already has a counter of that kind on it, each counter of that kind receives a new timestamp identical to that of the new counter." There is one timestamp per KIND on an object, and it moves forward on every placement.
- **CR 113.11.** An object that "can't have" an ability loses it, and "it's also impossible for an effect **or keyword counter** to add that ability to the object."
- **CR 122.2.** Counters do not survive a zone change; they cease to exist (CR 400.7).
- **CR 701.34a.** Proliferate gives "each one additional counter of each kind that permanent or player already has". That is a placement, so it gives the kind a new timestamp under CR 613.7c.
- **CR 702.1c.** A grant of counters "of that keyword" grants each appropriate variant. Hexproof's variant is "hexproof from [quality]" (CR 702.11d).
- **CR 708.2a.** A face-down permanent is a 2/2 with no text. Turning a permanent face down is not a zone change, so its counters stay on it.

The Ikoria release notes, where keyword counters were introduced (2020-04-10), add two points:

- "For determining the interaction of continuous effects, the timestamp of a keyword counter on an object is the most recent time that any counter of that kind was put on that object. Removing a keyword counter doesn't change the timestamp of any remaining counters."
- "Among the keywords that a keyword counter can grant, multiple instances of any one of these keywords are redundant."

The Elspeth Resplendent ruling (2022-04-29) draws the contrast from the other side: "'Shield' is not an ability that creatures have and shield counters are not keyword counters. If a creature with a shield counter loses its abilities, the shield counter will still protect it as normal." A keyword counter does not get that protection.

**Does "loses all abilities" remove a counter's keyword?** It depends on timestamps. Both effects are in layer 6 and neither is a CDA, so CR 613.3 orders them:

- **Removal earlier than the counter.** The removal applies, then the counter adds its keyword. The permanent **has** the keyword. An example is a Turn to Frog cast before a flying counter is put on the Frog.
- **Removal later than the counter.** The counter adds its keyword, then the removal takes it away. The permanent **does not** have the keyword. An example is a flying counter first, then Turn to Frog. The counter stays on the permanent (CR 122.1b concerns the keyword, not the counter), and it grants again only if a later counter of that kind renews its timestamp.
- **Dependency.** Neither effect depends on the other under CR 613.8a. Removing abilities does not change whether the counter exists or what it applies to. So timestamp order is the whole answer.
- **"Can't have"** beats the counter whatever the timestamps (CR 113.11).

### How many cards

A scan of the Scryfall default-cards dump found **128 distinct cards** whose oracle text names a keyword counter, **117** of them Commander-legal. The scan deduplicated by oracle ID and skipped art-series, token and Un-set printings. It counted choice lists ("a counter from among flying, first strike, lifelink, or vigilance") as well as single counters.

| Kind | Cards |
|---|---|
| flying | 34 |
| indestructible | 23 |
| lifelink | 19 |
| trample | 17 |
| vigilance | 13 |
| deathtouch | 12 |
| menace | 11 |
| first strike | 9 |
| reach | 9 |
| double strike | 4 |
| hexproof | 4 |
| shadow | 3 |
| haste | 2 |
| decayed | 1 (Rot-Curse Rakshasa) |
| exalted | 1 (Emissary of Soulfire) |

Six of the 128 are in the catalog: the five in the table above, and Mondrak, Glory Dominus. Mondrak's counter ability is blocked by its sacrifice cost, not by this seam. Examples from the other 122 include Perennation, the five Myojin of Kamigawa, Luminous Broodmoth, Crystalline Giant, Grimdancer, Elspeth Resplendent, Kathril, Unbreakable Bond, Arwen, Mortal Queen, and the Ikoria "-bonder" cycle.

### What the engine already has

- **Storage.** `Card.Counters map[string]int`, string-keyed. Every placement and removal goes through one door, `applyCounterByLocked` (`game/mutations.go`). That includes `AddCounter`, `AddCounterForEffect`, entry counters (`applyEntryCountersLocked`), proliferate (`applyProliferateLocked` calls `AddCounterByForEffect` once per kind), and counter-removal costs. The door emits `EventCounterPlaced` for removals too, and `layer_listener.go` already bumps the layer version on it. Two sites write the map directly: the CR 704.5q +1/+1 / −1/−1 cancel (`mutations.go`) and `Game.Clone`. Neither touches a keyword kind.
- **Clearing.** `MoveCard` drops `Counters` on every zone change (CR 122.2).
- **Timestamps.** `EnteredBattlefieldAt`, `AttachedAt` and `FaceTurnedAt` are wall-clock nanoseconds (`timeNowUnixNano`). `Card.layerTimestamp()` is the latest of the three. Layer 6 sorts its bucket by `ContinuousEffect.Timestamp()` with a stable sort (`applyLayerLocked`).
- **Layer 6.** `printedCharacteristic` puts printed keywords in the baseline. Grants append through `AppendKeywordAbility`, which dedupes every keyword except the cumulative ones (prowess and toxic, neither of which is a keyword counter). A `RemovesAbilities` effect empties the list in its own timestamp slot. `enforceCantHaveLocked` strips "can't have" tokens straight after the bucket (#1651).
- **Silencing.** `applyOneEffectLocked` silences an effect only when `effectSourceLocked` names a live battlefield source that has lost its abilities. An effect with no such source is never silenced.
- **The keyword table.** `canonicalKeywords` (`game/keywords.go`) is closed: a keyword joins it in the same change that teaches the engine to honour it. Of CR 122.1b's fifteen, **thirteen are in it**. Decayed and exalted are not. Nor is any "hexproof from [quality]" variant, which ADR 0038 §6 refused on purpose.
- **Wire.** `CardView.counters` already ships every counter. `CardView.abilities` is the effective list, so whatever layer 6 adds becomes a badge (`KeywordBadgeRow.svelte`). The client's `counterTypes.ts` draws an unknown counter kind as a neutral pip showing its first three letters.

---

## Decisions

### 1. The counter kind IS the keyword token, from one closed table

A keyword counter is a `Card.Counters` entry whose kind is a canonical keyword token, spelled exactly as `canonicalKeywords` spells it: `"flying"`, `"first strike"`, `"indestructible"`. The catalog already spells them this way (Vraska places `"deathtouch"` and Sorin places `"lifelink"`). There is no name mapping to maintain.

The set of kinds that count is a new closed table beside the keyword table, `keywordCounterKinds` in `game/keyword_counters.go`. It holds CR 122.1b's list **intersected with** `canonicalKeywords`:

> flying, first strike, double strike, deathtouch, haste, hexproof, indestructible, lifelink, menace, reach, shadow, trample, vigilance

- **One predicate:** `game.IsKeywordCounter(kind) bool`. Nothing else decides the question.
- **Named constants:** `CounterFlying` through `CounterVigilance` join `counter_types.go`, so a card file writes `game.CounterIndestructible` rather than a string literal.
- **Two guard tests.** One pins the table to CR 122.1b's list, so a kind outside 122.1b cannot join. The other pins every entry to `canonicalKeywords`, so the engine never grants a keyword it does not enforce (ADR 0014's closedness rule).
- **Decayed and exalted stay out.** Each joins `keywordCounterKinds` in the same change that adds its keyword to `canonicalKeywords`. Until then Rot-Curse Rakshasa and Emissary of Soulfire stay out of the catalog for the missing keyword, not for this seam.
- **"Hexproof from [quality]" stays out.** It would join as a parameterised family (`"hexproof from white"`), the way protection did, once hexproof-from itself is enforced. Kathril, Aspect Warper can still ship with a caveat for the variants.
- **One keyword, however many counters.** Two flying counters grant flying once. `AppendKeywordAbility` already dedupes, and no kind in the table is cumulative.
- **Unknown kinds do nothing.** A counter kind that is not in the table (a misspelling, `"Flying"`, or a homebrew kind) round-trips as storage and grants nothing, as today.

### 2. Every keyword kind gets a timestamp (CR 613.7c)

`Card` gains one field:

```go
// CounterStampedAt is the CR 613.7c timestamp of each KEYWORD counter
// kind on this object: the moment the most recent counter of that kind
// was put on it. Removing counters does not change it. nil means none.
CounterStampedAt map[string]int64 `json:"counterStampedAt,omitempty"`
```

- **Written in one place.** `applyCounterByLocked` stamps `timeNowUnixNano()` when `delta > 0` and `IsKeywordCounter(name)`. It deletes the entry when the kind's count reaches zero. Every door that places counters (entry counters, proliferate, Doubling Season's doubled placement, a sandbox edit) goes through it, so all of them stamp.
- **Keyword kinds only.** A +1/+1 counter has no layer-6 meaning, and stamping it would add snapshot noise for nothing.
- **Cleared with the counters.** `MoveCard` clears it wherever it clears `Counters` (CR 122.2, 400.7).
- **Copied by `Game.Clone`**, so undo keeps the timestamp.
- **Not a copiable value** (CR 707.2). A Clone does not copy counters, and does not copy their timestamps either.
- **No stamp.** A card that has a keyword counter but no stamp (a restore point from before this change, or a test fixture that writes the map directly) is ordered at `layerTimestamp()`. That is the earliest moment the counter could have been placed. See Decision 7.

### 3. One more source in the layer-6 gather

`activeStaticAbilitiesLocked` gains one source list, `keywordCounterEffectsLocked()`. It sits beside the scoped effects, emblems, declared-zone statics and mana riders, and it is not a second pass.

For each battlefield permanent and each keyword kind in its `Counters`, it contributes one `keywordCounterEffect`:

- **Layer:** `Layer6Ability`.
- **Timestamp:** the kind's `CounterStampedAt`, or the fallback from Decision 2.
- **AppliesTo:** that one object, identified by instance ID.
- **Apply:** `c.Abilities = AppendKeywordAbility(c.Abilities, kind)`.
- **`RemovesAbilities` and `ContinuesAfterRemoval`** are both false.

It has **no source permanent** in `effectSourceLocked`'s sense, so the CR 613.6 silencing never touches it. A counter is not an ability of the permanent, which is exactly the difference from the `b24` grant. Everything else follows from the existing machinery, with no new rule code:

- **"Loses all abilities"** is ordered against the counter by timestamp in the one stable sort, as Context sets out. Tests pin both orders.
- **"Can't have"** strips the keyword afterwards, because `enforceCantHaveLocked` runs after the whole bucket (CR 113.11).
- **A granted-then-lost keyword composes by timestamp.** Examples are Equipment's "loses flying" and a Frogify placed before or after the counter.
- **Face-down permanents** keep their counters' keywords. The counter effect does not go through the catalog, so `CatalogKey`'s empty answer for a face-down permanent does not hide it. A face-down 2/2 with a flying counter flies.
- **Phased-out permanents** are not in `g.Battlefield`, so they contribute nothing, as with every other effect.
- **Last-known information.** `snapshotLKILocked` copies the effective characteristics, so a creature that died with a deathtouch counter still had deathtouch as it last existed. The #1396 departed-source damage path reads that.
- **Every reader.** `HasKeyword`, `ProtectionQualities`, the combat, targeting and destruction paths, `internal/legal`, the bot and the wire all read the one effective list (ADR 0046). None of them changes.

The CR 613.8 dependency probe runs on layer 4 only (ADR 0067), and a keyword counter has no dependency to find (Context), so the probe is not involved. `st.track` (the fast path for boards with no ability removal) is unaffected, because a counter effect removes nothing.

The cost is one walk over the battlefield's counter maps per recompute, and a table lookup per kind. Recomputes already happen on every counter change.

### 4. Off the battlefield: read, but only through the same predicate

CR 122.1b also covers "a card in a zone other than the battlefield". Such a card only has counters if an effect put them there, because CR 122.2 wipes them on the way in. Examples are a suspended card's time counters, or an effect that exiles a card "with a flying counter on it". None of the 128 cards is known to do this with a keyword counter.

`forEachAbilityToken`'s off-battlefield branch gains one loop over `c.Counters` filtered by `IsKeywordCounter`, after the printed keywords. The view's off-battlefield ability list reads the same helper, so the badge and the rule cannot disagree. There is no timestamp question off the battlefield: no layer pass runs there, and the only competing effects are the card's own printed keywords. **Owner question 3** asks whether this ships now or waits for a card that needs it.

### 5. Proliferate, removal costs and the SBA need nothing new

- **Proliferate** adds one counter of each kind through `AddCounterByForEffect`, so a proliferated keyword counter gets a fresh timestamp (CR 613.7c, 701.34a). If a later "loses all abilities" had taken the counter's keyword away, proliferating the counter gives the keyword back. That is the rules' answer, and it falls out of Decision 2.
- **Counter-removal costs** already exist: Tekuthal's "remove three counters from among" and Arwen's removal of her own indestructible counter. Removing the last counter of a kind removes the keyword on the next recompute, because the removal emits `EventCounterPlaced` and the layer version is bumped. Removing SOME counters of a kind leaves the timestamp alone (release notes).
- **CR 704.5q** cancels only +1/+1 against −1/−1, and never touches a keyword kind. Its direct map write (#1630's subject) needs no stamp handling. Decision 2 names `applyCounterByLocked` as the only door that stamps.
- **The public log** already narrates a counter landing (#1021). A keyword counter is an ordinary counter there.

### 6. The wire and the client: a badge appears, and the pip gets a style

No new wire field. The keyword arrives in `CardView.abilities` like any grant, so the badge row shows it. The counter still arrives in `CardView.counters`, so the pip still shows the count. Both are true statements about the permanent: it has a flying counter, and it has flying.

The one client change is cosmetic. `counterTypes.ts` gains pinned styles for the thirteen keyword kinds, so a flying counter's pip reads "Fly" in a keyword colour rather than a neutral "fly". The server's `counter_types.go` constants are the source of truth, as that file already says.

A badge carries no mark saying it came from a counter. Whether the owner wants one is **owner question 2**.

### 7. Snapshot compatibility

- **Additive field.** `Card.CounterStampedAt` is additive, within schema v7. The shape record `testdata/snapshot_shape/v7.txt` is updated with `-update-shape` in the same change. There is no `SnapshotSchemaVersion` bump.
- **Older restore points.** A file written before the change restores with no stamps. Its keyword counters start granting their keyword at once, ordered at `layerTimestamp()`. That differs from the pre-change engine only where the counter had no grant before, and there the old answer was the declared-weaker one. The one board where the fallback timestamp can be wrong has a keyword counter placed AFTER an ability-removal effect, both already on the table before the deploy. The restored order then says the removal wins. That is the rarer order, and it errs toward losing the keyword. I accept it rather than bump the schema.
- **The corpus.** `TestWriteSnapshotCorpus` gains one v7 board with keyword counters and a later "loses all abilities" effect. The writer may add a board to an existing version without a bump (ADR 0041 phase 3, owner decision 6). The ordering therefore survives a restore.
- **Rollback.** An older binary reads a newer file and ignores `counterStampedAt`, because Card fields are not checked strictly. The strict check covers only effect records and stack items. That binary still carries the `b24` statics in its own catalog, so a rollback is as correct as that binary ever was.
- **No closure is added.** The counter effect is built at each recompute from data, so `closure_fields.txt` does not change.

### 8. The workaround is deleted in the same change

`b24KeywordCounterGrant` and its five uses are deleted with the engine change.

- **Vraska Joins Up** and **Sorin of House Markov** lose their caveats and become `CompletenessFull`. Their tests that pinned "the counters go inert when the source leaves" are inverted to pin that the counters keep working.
- **Tekuthal, Solphim and Drivnod** drop the static. Their doc comments stop naming `b24`. A test is added for the corner the workaround got wrong: the Dominus loses all abilities first, then gets its counter, and is still indestructible.
- **`putCounterOnSourceWhileOnBattlefield`** and its comment in `helpers.go` stay, minus the sentence about pairing with `b24`.
- **The clone baseline** is regenerated if the deletion shrinks a group, as `TestNoNewExactClonesInTheCatalog` asks.

Keeping the statics would be harmless, because `AppendKeywordAbility` dedupes them. But it would leave five card files describing a rule the engine owns, which is the stale-caveat problem AGENTS.md §7 warns about.

### 9. Cards and docs

- **Perennation** joins the catalog: return the permanent card, then put a hexproof counter and an indestructible counter on it. It uses the same shape as Rakdos Joins Up's `b33ReanimateChosenWithCounters`, generalised to named kinds. It works for any permanent card, not only creatures. That closes #1117's last seam-blocked card.
- **The roadmap row** `keyword-counters` flips to implemented, with a `docs/engine-seams/closed/1753-keyword-counters.md` fragment. `Waiting` is emptied.
- **`docs/adding-cards.md`** gets a short "A keyword counter needs no grant" paragraph under "Adding a combat-keyword card". It says to put the counter with `AddCounter{Kind: game.CounterFlying}` and write nothing else. Decayed, exalted and hexproof-from are named as not yet counted. The layer table's "+1/+1 counter math stays in `CurrentPower`" row is unchanged: P/T counters remain layer 7d and are not touched here.
- **The other 122 cards** are unblocked **as far as this seam goes**. Each still needs its other clauses checked by the batch that builds it: mutate, the Myojin's "if you cast it from your hand", Crystalline Giant's random choice, and so on. This ADR does not claim any of them are buildable.

---

## Consequences

- One engine-owned rule replaces five per-card statics, and it is right in the three corners the statics got wrong: ability removal on the source, timestamp order, and noncreature permanents.
- Keyword counters work with no card on the battlefield to carry them. That is the whole of Perennation, and the caveats on Vraska Joins Up and Sorin of House Markov.
- `Card` gains one additive snapshot field, stamped at the one counter door.
- A permanent with a keyword counter shows both the pip and the keyword badge.
- Decayed, exalted and hexproof-from counters stay storage until their keywords are enforced. The table's guard test makes that a deliberate step rather than an oversight.

## Out of scope

- **Mutate and merged permanents** (Ikoria's other mechanic). Many keyword-counter cards are also mutate cards; they wait on mutate.
- **Decayed, exalted and "hexproof from" as keywords.** Each is its own seam, and its counter follows it into `keywordCounterKinds`.
- **Player counters.** CR 122.1b is about objects. Player "keyword" effects (a player with hexproof) are `player_statics.go`'s.
- **Shield, stun and other rules-read counters.** They are CR 122.1c onward, and already engine-owned where built.

---

## Open questions for the owner

1. **Delivery.** Should this ship as one PR, with the engine read, the `b24` deletion and caveat clears (Decision 8), the roadmap flip and Perennation? Or as two, with the engine PR followed by a card PR carrying Perennation? I recommend one PR, because Perennation is a single small card and the seam's `Waiting` list would otherwise point at a card that is not there yet.
2. **Badge provenance.** Should a keyword that comes from a counter be marked on the wire (say `counter_keywords: ["flying"]`) so the client can draw its badge differently? I recommend no. The counter pip already says where it came from, and a second field is a second answer to "does it have flying".
3. **Off-battlefield counters (Decision 4).** Ship the off-battlefield read now, or wait for a card that puts a keyword counter on a card outside the battlefield? I recommend shipping it now: it is one loop behind the same predicate, and CR 122.1b names it.
4. **The restore fallback (Decision 7).** Is ordering an unstamped keyword counter at its permanent's own timestamp acceptable for restore points written before the change? The alternative is a schema bump with a migration that cannot know the true timestamp either. I recommend the fallback.
5. **Decayed and exalted.** Should either keyword be scheduled so its counter can join the table? Rot-Curse Rakshasa and Emissary of Soulfire are one card each. I recommend waiting for a deck that asks.

---

## Owner decisions (2026-09-30)

The owner accepted all five recommendations.

1. **Delivery: one PR.** It contains the engine read, the `b24` deletion with the caveat clears, the roadmap flip, and Perennation.
2. **Badge provenance: no `counter_keywords` wire field.** The counter pip says where the keyword came from.
3. **Off-battlefield counters: ship the read now** (Decision 4).
4. **Restore fallback: accepted** (Decision 7). An unstamped keyword counter is ordered at its permanent's own timestamp. There is no schema bump.
5. **Decayed and exalted wait** for a deck that asks for them.

---

## Amendments

### 2026-10-08 (#2538): exalted becomes a keyword, and its counter joins the table (Accepted)

**Status:** Accepted (owner answers 2026-10-08). The owner chose the recommended option on every question; see "Owner answers (2026-10-08)" at the end of this amendment.
**Reconsiders:** owner decision 5 ("Decayed and exalted wait for a deck that asks for them"), for exalted only. On 2026-10-08 the owner agreed to reconsider it for Emissary of Soulfire, which S68's energy sweep found.
**Registry row:** `exalted-counters` (`server/internal/roadmap/registry.go:2132`).
**Line numbers** below are on `develop` at `62fa97c64`.

#### Why now

Emissary of Soulfire reads: "Pay {E}{E}: Put an exalted counter on target creature you control. Activate only as a sorcery." Energy is built (ADR 0129), so the counter is the only missing piece. Decision 1 keeps the counter out until exalted is in `canonicalKeywords`.

Making exalted a keyword buys more than the one card:

- **Printed exalted.** 35 Commander-legal cards print exalted as a keyword (Scryfall's `keywords` array, deduplicated by oracle ID). #2538's count of 42 is every card whose text names exalted. Only Noble Hierarch and Ignoble Hierarch are in the catalog. The deck importer filters Scryfall's keywords through `CanonicalKeywords` (`deck/deck.go:345`), so it drops "Exalted" today, and the other 33 have no exalted at all.
- **Keyword-only cards.** 15 of the 35 print nothing but canonical keywords once exalted is one: Akrasan Squire, Aven Squire, Court Archers, Duskmantle Prowler, Ethercaste Knight, Goblin Champion, Guardians of Akrasa, Knight of Glory, Knight of Infamy, Outrider of Jhess, Rhox Charger, Servant of Nefarox, Sigiled Behemoth, Sigiled Paladin and Waveskimmer Aven. `NeedsCatalogEffect` skips a line that is only canonical keywords (`game/coverage.go:160`), so these 15 move from `manual` to `no_effect` with no catalog entry.
- **Granted exalted.** Six Commander-legal cards GRANT exalted: Sublime Archangel ("Other creatures you control have exalted"), First Sliver's Chosen, Merchant of Truth, Rashel, Fist of Torm, Zarda, the Power Princess, and Rammas Echor, Ancient Shield. A constructor on `Spec.Triggered` cannot reach a granted instance. A token can, through `KeywordGrant` and `TribalKeywordGrant` (`cards/effects/tribal.go:176`, `:196`) and the other layer-6 grants.

ADR 0014's amendment of 2026-09-24 made this argument for prowess, and ADR 0106 §3 made it for evolve.

#### What the rules say

I checked every rule below against the Comprehensive Rules effective September 25, 2026.

- **CR 702.83a.** "Exalted is a triggered ability. 'Exalted' means 'Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.'"
- **CR 702.83b.** "A creature 'attacks alone' if it's the only creature declared as an attacker in a given combat phase. See rule 506.5."
- **CR 506.5.** "A creature attacks alone if it's the only creature declared as an attacker during the declare attackers step."
- **CR 113.2c.** "If an object has multiple instances of the same ability, each instance functions independently." CR 702.83 has no rule of its own about multiple instances, so this is the rule that makes two exalted abilities two triggers. The registry row's "one trigger per instance" was attributed to CR 702.83b in the brief for this amendment. That is wrong: 702.83b defines "attacks alone". Sublime Archangel's reminder text says the same thing: "If a creature has multiple instances of exalted, each triggers separately."
- **CR 603.3a.** "A triggered ability is controlled by the player who controlled its source at the time it triggered, unless it's a delayed triggered ability."
- **CR 508.3a.** An ability that reads "Whenever [a creature] attacks" triggers only if that creature is declared as an attacker. It "won't trigger if a creature is put onto the battlefield attacking."
- **CR 122.1b** lists exalted among the keywords a keyword counter can be.
- **CR 613.7c** gives every counter of one kind one timestamp, renewed on each placement. Decision 2's stamp already does this for exalted counters.

Two rulings bear on the design:

- **Emissary of Soulfire (2024-06-07):** "A creature with multiple exalted counters will have that many instances of exalted." So exalted counters are **not** redundant. Two statements stop being true once exalted joins: Decision 1's "One keyword, however many counters", and the comment on `keywordCounterEffect.Apply` (`game/keyword_counters.go:154`, "CR 122.1b's keywords are all redundant when repeated"). The Ikoria release note they rely on predates exalted counters.
- **Sublime Archangel (2012-07-01):** "count the number of instances of exalted among permanents you control. After those abilities resolve, that's how many times the creature will get +1/+1."

#### How keyword triggers work today

Exalted can be built on the machinery that prowess, evolve and annihilator already use. Each of the three is derived from the effective ability list and has no catalog row.

1. **The token is the declaration.** `canonicalKeywords` holds `KeywordProwess` (`game/keywords.go:224`), `KeywordEvolve` and `KeywordAnnihilator` (`:242`). The deck importer stamps a printed instance from Scryfall, a token template's `Keywords` carries it, and a layer-6 grant appends it.
2. **Cumulative, so instances survive.** `KeywordIsCumulative` (`game/infect_wither_toxic.go:163`) names all three, so `AppendKeywordAbility` (`:135`) keeps every granted instance instead of deduping it. For a cumulative token, `mergePrintedKeywords` (`game/characteristic.go:458`) takes the higher of the catalog's count and the import's count, never the sum. The importer counts a doubled printed line such as "Prowess, prowess" (#1510).
3. **One trigger per instance.** `keywordTriggersFor` (`game/prowess.go:144`) walks the list with `forEachAbilityToken` (`game/keywords.go:565`), the same walk `HasKeyword` uses. It returns one `TriggeredAbility` per instance. `triggersOf` (`game/designations.go:388`) puts them in front of the catalog's rows, so every harvest sees them (`game/triggers.go:680`).
4. **The same harvest as a catalog trigger.** `harvestMatchLocked` (`game/triggers.go:787`) treats a derived trigger exactly as it treats a catalog one: suppression, trigger doublers (`triggerDoublersLocked`, `game/trigger_doubling.go:180`), the CR 603.3d target check and the APNAP queue. `stampTriggerContext` (`game/triggers.go:927`) puts the triggering event on the item.
5. **A keyed item, not a catalog ref.** Each `Build` calls `NewKeyedTriggeredItem` (`game/triggers.go:80`) with a body registered under an on-disk key: `"prowess/pump"` (`game/prowess.go:100`) or `"annihilator/sacrifice"` (`game/annihilator.go`). A table with one of these on the stack is a restore point. An older binary refuses the file with `ErrUnknownEffectKey`, which is the designed rollback case.
6. **Ability removal and face-down come free.** A CR 613.1f "loses all abilities" empties the list in its own timestamp slot, and the triggers go with it. A face-down permanent has no tokens, so it has no keyword triggers.

Exalted today is the other shape. `effects.Exalted()` (`cards/effects/exalted.go:30`) is a catalog `TriggeredAbility` with a declared `Effect`. Its stack item is therefore stamped `Body: "catalog/triggered"`, with an `AbilityRef` naming the card's row (ADR 0041 P9). It has **two** production call sites, `noble_hierarch.go:21` and `ignoble_hierarch.go:21`. The brief for this amendment counted five. The other hits are comments and the shared `attackedAlone` helper's two readers, `derelict_attic_widows_walk.go:32` and `disturb_batch_b_helpers.go:19`. Those two are not exalted, and they keep working.

#### The plan

##### A1. `KeywordExalted` joins `canonicalKeywords`, as a cumulative keyword

- **The token.** `const KeywordExalted = "exalted"` goes in a new `game/exalted.go`. Its entry in `canonicalKeywords` gets the kind of comment the other triggered keywords carry.
- **Cumulative.** `KeywordIsCumulative(KeywordExalted)` is true, citing CR 113.2c. Two grants are two abilities, and a printed exalted under Sublime Archangel's grant is two.
- **No importer change.** The deck importer already filters against the table. Its line scan already counts repeats of a cumulative keyword. Urza's Dark Cannonball's "Exalted, exalted" is the only printed double, and it is an Un-card.

##### A2. One derived trigger per instance

`keywordTriggersFor` gains a fourth count. Each exalted instance contributes one copy of a package-level `exaltedTrigger`, built like `prowessTrigger`:

- **Keyword:** `KeywordExalted`.
- **Watches:** `EventAttack`. That event is emitted only for a declared attacker (CR 508.3a).
- **AppliesTo:** two tests.
  - The attacker is controlled by the source's controller (`ev.Actor == source.Controller`).
  - Exactly one creature was declared as an attacker. That count moves into the game package as `AttackedAlone(g)`, the body of today's `effects.attackedAlone` (`cards/effects/exalted.go:61`). `DeclareAttackers` stamps every attacker before it announces any `EventAttack`, so a lone declaration reads as alone. The effects package's two other readers call the exported function.
- **Build:** `NewKeyedTriggeredItem(source, exaltedLabel, exaltedPumpBody, EffectParams{Object: <the attacker's ObjectRef>})`.
  - The attacker is fixed when the ability triggers, by instance ID and battlefield-entry stamp. An attacker that leaves and comes back in response gets nothing (CR 400.7).
  - That is what `exaltedPumpTheLoneAttacker` does today through `item.Trigger.Event.CardID`. Carrying the attacker in `Params.Object` means the body reads one field instead of the trigger context.
- **The body,** `"exalted/pump"`: a layer-7c +1/+1 until end of turn on the pinned attacker. It is registered as a scoped-effect data record, as `resolveProwess` does (`game/prowess.go:111`).
- **Commutes:** true, for the reason #1511 gives for prowess.
  - Every exalted instance reads only its pinned attacker and writes only a +1/+1 to it. A batch of nothing but exalted triggers therefore needs no CR 603.3b ordering prompt.
  - A batch that also holds Rafiq of the Many's or Battlegrace Angel's own "attacks alone" trigger still asks, because those do not commute.

**The controller of the trigger** is its source's controller when it triggered (CR 603.3a).

- `NewTriggeredItem` already sets `Controller` from `source.Controller` (`game/triggers.go:63`), and the harvest reads the battlefield as it stands when the event fires.
- Exalted on a permanent you control triggers only for your own lone attacker. In a Commander game only the active player attacks, so only the active player's exalted triggers.
- A permanent whose control changed, such as a Noble Hierarch taken with Threaten, triggers for its new controller.

**Noncreature sources** need nothing extra. Cathedral of War is a land, and Angelic Benediction and Finest Hour are enchantments. `forEachAbilityToken` reads any permanent's list, so their printed exalted works like a creature's.

##### A3. Several instances: printed, granted and counters

Every instance is one entry in the effective ability list, and the trigger count is the entry count. The three sources combine like this:

- **Printed.** The importer's stamp and the catalog's `PrintedKeywords` merge to the higher count (`mergePrintedKeywords`). A card that prints exalted once has one instance, whether or not its catalog entry also declares it.
- **Granted.** Each grant appends one instance, because the keyword is cumulative. With two Sublime Archangels out, each other creature you control has two instances and each Archangel has two: its own printed one and the other's grant.
- **Counters.** `keywordCounterEffect` gains a count field.
  - For a cumulative kind, `Apply` appends `Counters[kind]` instances. For every other kind it appends one, as now.
  - `keywordCounterTokens` (`game/keyword_counters.go:199`), the off-battlefield half, does the same, so the badge and the rule agree.
  - All counters of one kind share one CR 613.7c timestamp. A later "loses all abilities" removes every counter instance at once, and a placement or proliferate after it brings them all back. That is Decision 2's stamp, unchanged.

The comment on `keywordCounterEffect.Apply` is corrected to cite the Emissary ruling. Decision 1's bullet "One keyword, however many counters" now reads "one keyword per kind, however many counters, except exalted, which gives one instance per counter (ruling of 2024-06-07)".

##### A4. Migrating the two call sites without double triggers

The risk is a card that has both the catalog row and the token, which would trigger twice. Four steps prevent it:

1. **Delete the constructor.** `Exalted()` is deleted from `cards/effects/exalted.go`, along with `exaltedPumpTheLoneAttacker`. `attackedAlone` either becomes a one-line call to `game.AttackedAlone`, or is deleted and its callers repointed.
2. **Change the two cards.** Noble Hierarch and Ignoble Hierarch drop `Triggered: []game.TriggeredAbility{Exalted()}` and declare `PrintedKeywords: []string{game.KeywordExalted}`. Every prowess and evolve card already has this shape (`monastery_mentor.go`, `fathom_mage.go`). With the importer's stamp beside it, `mergePrintedKeywords` keeps one instance, not two.
3. **Add a guard test.** It fails on any catalog row that watches `EventAttack` and has a label starting "Exalted". It also asserts that each Hierarch harvests exactly one trigger for a lone attack.
4. **Fix the comments that call exalted a constructor:** `provoke.go:19` ("for Exalted's reason"), `shared_animosity.go:16` and `goblin_rabblemaster.go:28`. Provoke's reason for staying out of the table loses exalted as its example.

##### A5. `CounterExalted` joins `keywordCounterKinds`, and the guard tests follow

- **The constant.** `CounterExalted = "exalted"` is added to `game/keyword_counters.go` and to `keywordCounterKinds` (`:65`).
- **The two guard tests.**
  - `TestKeywordCounterKindsAreCR1221b` needs no change, because exalted is in CR 122.1b.
  - `TestKeywordCounterKindsAreCanonicalKeywords` drops "exalted" from the kinds that must stay out (`game/keyword_counters_test.go:55`).
  - The count assertion at `:443` becomes 14, "CR 122.1b's fifteen less decayed".
- **A new test pins the ruling.** Two exalted counters on one creature are two instances and two triggers. So are one exalted counter and a printed exalted.
- **The file comment** changes from "Two of CR 122.1b's fifteen are out" to one.

##### A6. Emissary of Soulfire

- **The card file,** `cards/effects/emissary_of_soulfire.go`, has two abilities:
  - the ETB "you get {E}{E}{E}", with ADR 0129's helper;
  - the activated "Pay {E}{E}: Put an exalted counter on target creature you control", with sorcery timing (CR 602.5d) and a "creature you control" target.
- **`CompletenessFull`,** if tests pin both abilities and an attack with two counters.
- **The registry row.** `exalted-counters` flips to implemented, with a `docs/engine-seams/closed/2538-exalted-counters.md` fragment, then `go test ./internal/roadmap/ -update`.
- **`docs/adding-cards.md`.** The "A keyword counter needs no grant" paragraph stops naming exalted as not counted. It gains one sentence: exalted counters are cumulative.

#### Snapshot and restore

- **No new Card field.** The token lives in `Abilities` and `Keywords`, which are already persisted. The counter is a `Counters` entry and its stamp is in `CounterStampedAt`, both from ADR 0101. `testdata/snapshot_shape/v7.txt` does not change.
- **A new effect key, `"exalted/pump"`.** It is on-disk vocabulary like `"prowess/pump"`: never renamed and never reused. An older binary refuses a newer restore point that has one waiting on the stack, with `ErrUnknownEffectKey`, and keeps the file. That is the designed rollback case, and within schema v7 it needs no bump.
- **A Hierarch trigger written before the change.**
  - **The problem.** Suppose a restore point from today's binary has a Hierarch's exalted trigger on the stack or queued. It names `catalog/triggered` with `AbilityRef{key: <the Hierarch's key>, slot: "triggered", ref: "own:0", name: "Exalted — +1/+1 until end of turn"}`. After A4 that row is gone, so `restoreCatalogAbility` (`game/ability_ref.go:445`) reaches the owner's Q3 outcome: the item becomes a manual item with no effect, the Hierarch is flagged `AbilitiesLostOnRestore` for the rest of the game, and the boot logs an ERROR.
  - **The size.** The window is narrow: from a lone attack until the trigger resolves, at the moment of a deploy. The flag, though, lasts the whole game.
  - **A fix.** Owner question 3 asks whether to close it with a one-entry restore alias. On a lost `catalog/triggered` ref whose name is exactly the retired exalted label, restore rewrites the item to `"exalted/pump"`, taking `Params.Object` from the item's carried `Trigger.Event`, and does not flag the card.
- **No frozen fixture is affected.** No file under `internal/game/testdata/snapshots/` names exalted. A new v7 board with a derived exalted trigger on the stack is added beside the others; the writer never touches an existing file.
- **`closure_fields.txt` does not change.** The derived trigger is package-level data, and its item is keyed.

#### Bot valuation

- **`keywordTable`** (`aiseat/heuristic/score.go:218`) gains `"exalted": 0.30`, summed per instance as prowess is. The value is low because the bonus applies only to a lone attack, and the table is board-blind on purpose. A creature with exalted counters becomes worth a little more, so the existing `targetsValue` prices Emissary's activation.
- **`redundantKeywordCounter`'s list** (`aiseat/heuristic/purpose.go:392`) must **not** gain exalted, because a second exalted counter is a second instance, not a redundant one. A test pins that an exalted-counter activation on a creature that already has exalted is not penalised.
- **Emissary's activation declares no `Purpose`.** `game.Purpose` has no field for "put a keyword counter on target creature", and adding one for a single card is the wrong trade. The bot prices the activation as an undeclared one (`ActivateBase` plus the target's value), less the energy it spends (ADR 0129 §7).
- **Out of scope:** teaching the attack planner to attack alone to collect exalted. The bot does not see the Hierarchs' exalted today either, so nothing regresses. Changing the planner would need ADR 0052's arena measurement. See owner question 4.

#### Decayed stays out, as its own seam

Decayed should **not** ride this change. CR 702.147a: "Decayed represents a static ability and a triggered ability. 'Decayed' means 'This creature can't block' and 'When this creature attacks, sacrifice it at end of combat.'"

- **Its consumer is different.** Decayed is a block restriction folded in after the layer pass (like unleash's `foldUnleashLocked`), plus a sacrifice at end of combat. This change builds neither.
- **Its cards are mostly tokens.** 19 Commander-legal cards print decayed or make something with it, and none is catalogued. Most of them create decayed Zombie tokens: Ghoulish Procession, Wilhelt, the Rotcleaver, Tainted Adversary, Diregraf Horde and Falcon Abomination, among others. Rot-Curse Rakshasa is the only card that uses the counter.
- **Recommendation.** File a `tier:3-design` seam issue for decayed with those 19 cards. Keep `keywordCounterKinds` at fourteen kinds until it lands; the counter is then a one-line follow-up. See owner question 5.

#### Delivery

One PR holds A1 to A5 (the engine, the migration, the counter and the tests), plus Emissary of Soulfire and the registry flip.

- **Testing.** The PR is engine work: a derived trigger, a new stack-item key, and a change to the layer-6 count. So it runs the branch E2E and the real-dump audits.
- **The coverage census** is left for CI to publish. It should show the 15 keyword-only cards moving to `no_effect`.

#### Owner questions

1. **Exalted as a canonical keyword with a derived trigger (A1 and A2), replacing `effects.Exalted()`?** Recommended: yes. The alternative is to keep the constructor and add a separate "granted exalted" static that builds triggers. That leaves the 33 uncatalogued printed cards without exalted, and makes granted and printed exalted two mechanisms that must agree.
2. **Exalted counters count one instance per counter (A3), per the Emissary of Soulfire ruling of 2024-06-07, by giving `keywordCounterEffect` a count for cumulative kinds?** Recommended: yes. One instance per kind would make a second Emissary activation on the same creature do nothing, which is weaker than printed.
3. **A Hierarch trigger on a restore point from before the change.**
   - (a) Recommended: a one-entry restore alias. It rewrites a lost `catalog/triggered` ref named "Exalted — +1/+1 until end of turn" to `"exalted/pump"`, so the trigger resolves and the Hierarch is not flagged for the rest of the game.
   - (b) Accept the existing Q3 outcome (a manual item, `AbilitiesLostOnRestore` and an ERROR log) for a window of one combat at one deploy.
4. **Bot scope.**
   - (a) Recommended: `keywordTable` gains `exalted` at 0.30 per instance, `redundantKeywordCounter` leaves it out, and the attack planner does not change in this PR.
   - (b) As (a), plus a separate bot PR, measured in the arena, that teaches the attack planner to value attacking alone with exalted.
5. **Decayed.**
   - (a) Recommended: it stays out of this change. File its own seam issue (19 cards, mostly decayed Zombie tokens), and its counter follows that seam.
   - (b) Build decayed in the same PR.
6. **Delivery.**
   - (a) Recommended: one PR with the engine change, the Hierarch migration, `CounterExalted` and Emissary of Soulfire, as ADR 0101 shipped Perennation with its seam.
   - (b) Two PRs: the engine and the migration first, then the card.

#### Owner answers (2026-10-08)

The owner chose option (a), the recommended one, on every question.

1. **Canonical keyword: yes.** Exalted joins `canonicalKeywords` as a cumulative keyword with a derived trigger (A1 and A2), and `effects.Exalted()` is deleted.
2. **One instance per counter: yes.** `keywordCounterEffect` gets a count for cumulative kinds, per the Emissary of Soulfire ruling (A3).
3. **A one-entry restore alias.** A lost `catalog/triggered` ref named "Exalted — +1/+1 until end of turn" is rewritten to `"exalted/pump"` on restore, and the Hierarch is not flagged.
4. **Keyword value only.** `keywordTable` gains `exalted` at 0.30 per instance, `redundantKeywordCounter` leaves it out, and the attack planner does not change.
5. **Decayed stays out.** It has its own seam issue, [#2650](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2650), and its counter follows that seam.
6. **One PR.** It holds the engine change, the Hierarch migration, `CounterExalted` and Emissary of Soulfire.
