# ADR 0101 — Keyword counters read by the engine

**Status:** Proposed · 2026-09-30 · Post-S30 — Rolling deck-driven catalog growth
**Issue:** [#1753](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1753). It relates to the Betor deck re-check on #1117, where Perennation is the last seam-blocked card.
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-09-30. I ran `git fetch --all --prune`, then read every `docs/decisions/` file name on every remote branch (41 heads). The highest number on any branch is **0098**. Numbers 0099, 0100 and 0102 are held for ADRs being written at the same time, so this one takes **0101**.
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
