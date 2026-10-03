# ADR 0109 — Rule gates, land types, mana and cost components

**Status:** Accepted · 2026-10-02 · S52 — Rule gates, land types, mana and cost components (tracker [#1909](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1909))
**Owner decisions:** 2026-10-02. All four questions are answered, each with the recommended option (a). See [Owner decisions](#owner-decisions-2026-10-02) at the end. The sections and the Delivery plan below are written as decided; the options not chosen are kept as considered options.
**Issues:** group B, land types and durations: [#1881](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1881) (a land that becomes a basic land type for a while), [#1604](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1604) (counter-held durations, "loses all land types", Teferi's Talent), [#1894](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1894) (for as long as you control this and it remains tapped). Group I, rule gates: [#1895](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1895) (players can't play lands), [#1899](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1899) (emblems that stop players casting spells), [#1885](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1885) (cards in graveyards can't be targeted). Group C, cost components and mana: [#1902](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1902) (costs from a library, or onto it), [#1862](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1862) (an ability that reads the card discarded to pay for it), [#1842](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1842) (a target bounded by the counters removed), [#1556](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1556) (riot), [#1552](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1552) (granted mana-spent readers).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-02. I ran `git fetch --all --prune` and read the `docs/decisions/` file names on all 37 remote heads (`origin/develop`, `origin/main` and 35 chore, docs, feat, fix, repro and wip branches). I also listed every name ever committed on any ref (`git log --all --name-only -- docs/decisions/`) and the files of every open PR. The highest number anywhere is 0108, and no open PR touches `docs/decisions/`. This one takes **0109**.
**Amends:** [ADR 0063](0063-durations-and-control.md) (§2 and §3 add duration conditions and a conjunction), [ADR 0041](0041-game-persistence.md) Decision P8 (§1, §2 and §4 add `ScopedEffect` kinds), [ADR 0093](0093-abilities-granted-to-other-permanents.md) Decision 10 (granted loyalty abilities and the CR 614.12 look-ahead come into scope, owner decisions 1 and 2), [ADR 0040](0040-mana-pipeline.md) (§11's granted readers) and [ADR 0100](0100-delve-either-or-and-variable-sacrifice-costs.md) (`PaidCost.Discarded` on activated abilities). A pointer line goes into each with the first implementation PR that touches it.
**Builds on:** [ADR 0107](0107-state-triggers-rebound-disturb-and-damage-prevention.md) (state triggers, the stack keyword pass, the rules gates), [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) (any-player activation; owner decision 6, a PR lands every card its seam unblocks), [ADR 0102](0102-entering-under-another-players-control.md) (the copy look-ahead in the entry window), [ADR 0066](0066-granted-cast-and-play-permissions.md) (play permissions), [ADR 0058](0058-doesnt-untap.md) (untap holds), [ADR 0057](0057-win-and-lose-by-effect.md) (game-end gates) and [ADR 0033](0033-ai-bot-seat.md) §1 (the legal-move enumerator).

This ADR was written plan-first. No engine code changed with it. The engine and card changes land in the PRs listed under [Delivery](#delivery).

---

## Context

The S52 triage (2026-10-01, at `eccf038b`) grouped eleven open seams: two land-type and duration gaps, three rule gates, and six cost and mana gaps. #1600's mana-rock leftovers were in the same milestone; they are not designed here (see [Out of scope](#out-of-scope)). This ADR re-checks each seam on `origin/develop` at `2ac93592`, sizes it by the printed cards that need it rather than by its `Waiting` list (ADR 0106 owner decision 6), and designs all eleven.

How the numbers were made:

- **Printed.** The Scryfall default-cards dump of 2026-09-24 (`data/scryfall/default-cards.json`), Commander-legal cards only, deduplicated by oracle ID, tokens and art cards removed: 31,830 cards. I counted every card whose oracle text needs the seam with a regular expression over the text of every face, then read each hit and dropped the false positives. For a cost, I matched only the text before the colon of an activated ability.
- **Not in the catalog.** Compared against the oracle fixtures in `server/internal/cards/coverage/testdata/oracle/` at `2ac93592` (3,321 files).
- **Waiting.** The `Waiting` list of the seam's row in `server/internal/roadmap/registry.go` at the same commit.
- **Alone.** I read each uncatalogued card's full text and asked whether everything except this seam already has a shape in the engine. A ✓ is a card I found every other clause for, a ? is one I was unsure of, and the estimate is the ✓ count plus half the ? count. It is a planning number. Each PR verifies every card again.

The seams doc and the issues ran stale in several places. Each correction is stated where it matters, and these are the ones that change a design:

- **#1899 is not already a `CastBanRule`.** `CastBanRule` (`game/cast_ban.go`) can say "can't cast noncreature spells", but it is a `PlayerStatic` written onto the banned player. An emblem written that way would be one frozen record per opponent, invisible on the board, and would outlive its owner leaving the game (CR 800.4a), because `game/leave_game.go` sweeps the leaver's emblems but not the statics they wrote onto others. §5 explains.
- **Granted loyalty abilities almost work.** `ActivateCatalogAbility` (`game/activated.go`) already reads granted rows, and it enforces CR 606.3 and 606.6 on whatever row it gets. Teferi's Talent's file says nothing can grant an activated ability to the permanent it is attached to; `effects.GrantAbilitiesToAttached` has done that since ADR 0093 PRs 1-3. What blocks the card is ADR 0093 Decision 10's policy and one plumbing gap (§2).
- **#1552's three cards are catalogued**, each with a caveat whose stated reason has gone stale: Lux Artillery says spent colours are not recorded (they have been since #761), Coin of Mastery says mana sources are not snapshotted (they have been since #1212), and Satoru says a free cast and a paper-paid cast can't be told apart (`ManaOnPaper` and `ManaSpent.None()` do that). Opal Palace, which the registry calls uncatalogued, is catalogued with a caveat.
- **Riot already ships once, as an approximation.** Uncivil Unrest gives every nontoken creature the +1/+1 counter and never offers haste (`uncivil_unrest.go`). The riot row does not mention it.
- **Spreading Seas is not catalogued.** The land-type row names it as a user of `effects.SetsBasicLandType`. The users are Magus of the Moon and Song of the Dryads.
- **`SetsBasicLandType` removes too much.** `Characteristic.SetSubtypes` replaces *every* subtype, so Dryad Arbor under Magus of the Moon loses "Dryad". CR 305.7 replaces land types only. §1 fixes it.
- **The restore check has a gap.** `Duration.Known()` and the unknown-field scan run on `scopedEffects` and `delayedTriggers`, but not on the durations inside `castPermissions`, `seats[].statics` and `cards[].nextUntapSkips[].while`. An older binary reads an unknown condition there as "while the source is on the battlefield". §2 and §3 close it.
- **The world rule (CR 704.5k) is not implemented.** Three catalogued world enchantments carry it as a caveat (Caverns of Despair, Concordant Crossroads, Forsaken Wastes). §8 lands it.
- **A copied ability loses its payment facts.** `ability_copy.go:copiedPaidCost` keeps the counters removed and the life paid, but drops `Discarded`, `Exiled`, `Sacrificed` and `ReturnedAttacking`, so a copied Land's Edge ability would see no card. §8 fixes it.
- **Narset Transcendent's emblem is her −9, not her −7.**

Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, effective September 25, 2026).

### Sizing

| § | Seam | Issue | Printed | Not in catalog | Waiting | Alone (est.) |
|---|---|---|---:|---:|---:|---:|
| 1 | A land that becomes a basic land type for a while | #1881 | 33 | 33 | 3 | ~30 |
| 2 | Counter-held durations, and losing all land types | #1604 | 19 | 19 | 2 | ~16 |
| 3 | For as long as you control this and it remains tapped | #1894 | 8 | 8 | 1 | ~7 |
| 4 | Players can't play lands | #1895 | 14 | 14 | 1 | ~11 |
| 5 | Emblems that restrict players | #1899 | 3 | 3 | 1 | ~2 |
| 6 | Cards in graveyards can't be targeted | #1885 | 5 | 5 | 1 | ~5 |
| 7 | Costs from a library, or onto it | #1902 | 10 | 10 | 2 | ~9 |
| 8 | The card discarded to pay, and the world rule | #1862 | 29 + 3 caveats | 29 | 1 | ~19 |
| 9 | A target bounded by counters removed, or by X | #1842 | 9 | 8 + 1 caveat | 1 | ~8 |
| 10 | Riot and unleash | #1556 | 29 + 1 caveat | 29 | 2 | ~27 |
| 11 | Granted mana-spent readers | #1552 | 10 | 6 + 4 caveats | 3 | ~9 |

The table's seams come to about 170 printed cards and about 145 that need nothing else, against the tracker's "about 70 plus riot and the mana-spent readers". The difference is mostly families the issues do not name: the 22 other world permanents (§8), unleash (§10), the group and other-duration forms of §1, and the X-bound targets of §9. The owner's decisions add three groups the table leaves out:

- the four other Talents, with Teferi's Talent's caveat (decision 2, §2): 5 cards;
- 17 "discard a card at random" costs (decision 3, §7), about 15 of which need nothing else;
- about 60 cards that already have every shape they need but share a PR's code (decision 4), about 49 of which need nothing else.

**With the decisions, the sprint is about 250 printed cards and about 215 that need nothing else.** Notes on the counts:

- **§1.** 21 cards make one target land a basic land type until end of turn: Deepwood Elder, Dream Thrush, Dreamwinder, Floodchaser, Grixis Illusionist, Jinx, Kavu Recluse, Kukemssa Serpent, Moonbow Illusionist, Mystic Compass, Pixie Illusionist, Reef Shaman, Sea Snidd, Shimmering Mirage, Slimy Kavu, Streambed Aquitects, Tidal Warrior, Tideshaper Mystic, Tundra Kavu, Unstable Frontier and Zombie Trailblazer. The issue's list of 21 has Navigator's Compass instead of Deepwood Elder. Four set the type of a group of lands (Elsewhere Flask, Terraformer, Nightcreep, Vision Charm). Two add types for a turn (Navigator's Compass, Energybending). Six use another duration: "until this creature leaves the battlefield" (Gaea's Liege, Graceful Antelope), "for as long as this creature remains on the battlefield" (Tide Shaper), "until its controller's next untap step" (Orcish Farmer) and indefinitely (Cyclopean Giant, Thelonite Monk).
- **§2.** 17 cards print "for as long as that \<object\> has a \<kind\> counter on it" on a resolved effect, Ultima included: Aquitect's Will, Aven Mimeomancer, Cyclopean Tomb, Dread Wight, Eluge, Immortal Obligation, Liege of the Tangle, Makeshift Mannequin, Mathas, Minas Morgul, Obsidian Fireheart, Promise of Loyalty, Quicksilver Fountain, Sauron, Dino Devotee, Shield Broker, Ultima and Xolatoyac. Two more statics lose all land types (Lithoform Blight, Alpine Moon). The 28 statics that print "has \<keyword\> as long as it has a counter on it" (the Myojin, Lightwalker) are ordinary conditional statics and are not this seam.
- **§3.** Five cards print "for as long as you control this and this remains tapped": Seasinger, Helm of Possession, Hivis of the Scale, Rubinia Soulsinger and Willow Satyr. Old Man of the Sea prints "for as long as this creature remains tapped and that creature's power remains less than or equal to this creature's power". Zygon Infiltrator and Braided Net print "for as long as *that* creature remains tapped", about the target. 33 more print the single "for as long as this remains tapped" (owner decision 4).
- **§8.** Seven activated abilities read the card their discard cost discarded: Land's Edge, Hisoka, Minamo Sensei, Mercurial Chemister, Moria Scavenger, Necromancer's Stockpile, Slumbering Tora and Volrath the Fallen. 26 Commander-legal permanents have the world supertype. Three are catalogued with the caveat, and 23 are not (Land's Edge is one of them).
- **§10.** 13 cards print riot and 14 print unleash. Rhythm of the Wild and Spider-Punk grant riot, Tesak grants unleash, and Domri, Chaos Bringer's mana gives a creature spell riot.

---

## Group B — land types and durations

### 1. A land that becomes a basic land type for a while (#1881)

#### What exists, what is missing

- `effects.SetsBasicLandType(applies, types, subtypes)` (`effects/attachments.go`) is a layer-4 static with `RemovesAbilities: true`. Its users are Magus of the Moon and Song of the Dryads. Layer 4 clears the rules-text abilities and layer 6 grants land after it, so they survive. The basic land type's mana ability is not written anywhere: `manaAbilityRows` (`game/granted_abilities.go`) adds `intrinsicLandManaAbilities(c)` (`game/mutations.go`) from the effective subtypes, even on a permanent that has lost all its abilities.
- `Characteristic.SetSubtypes` (`game/characteristic.go`) replaces every subtype and clears `AllCreatureTypes`. That is too much for CR 305.7.
- A resolved effect can only add a subtype. `ModAddSubtypes` is the only layer-4 subtype kind (`modKinds` in `game/scoped_effects.go`). The others are `addTypes`, `removeTypes` (card types only) and `allCreatureTypes`. The Legend of Kyoshi and Sealock Monster use `ModAddSubtypes`.
- The engine has no list of land types. `landTypeMana` (`game/mutations.go`) knows the five basic land types, `game.AllCreatureTypes` knows the CR 205.3m creature types, and the Urza's subtypes appear only as card-file literals.
- A choice at resolution exists. `effects.PickOption` (`effects/resolution_choice.go`, `PendingChoiceOptionPick`) can ask "Plains, Island, Swamp, Mountain or Forest". There is no basic-land-type helper on top of it.

#### The rules

- **CR 305.7:** an effect that sets a land's subtype to one or more basic land types removes its old land types, all abilities generated from its rules text and any copiable effects, and gives it each new type's mana ability. It "doesn't remove any abilities that were granted to the land by other effects", and it doesn't change card types or supertypes. A land that gains a type "in addition to its other types" keeps everything (also CR 205.1b).
- **CR 205.1a:** setting a subtype replaces the existing subtypes "from the appropriate set", so a Dryad Arbor that becomes an Island keeps the creature type Dryad.
- **CR 205.3i:** the 17 land types are Cave, Desert, Forest, Gate, Island, Lair, Locus, Mine, Mountain, Plains, Planet, Power-Plant, Sphere, Swamp, Tower, Town and Urza's.
- **CR 613.1d, 611.2a, 611.2c:** a type change is layer 4. A resolved effect lasts as long as it says, and the set of lands it affects is fixed when it begins.

#### Decision

There is one reasonable design.

1. **The vocabulary.** `game.LandTypes` lists CR 205.3i's 17 types, and `game.IsLandType` answers for one. `game.BasicLandTypes` is the first five of them. These are the shared machinery for §2 as well.
2. **The kind.** `setBasicLandTypes`, layer 4, `removes: true`, reads `Mod.Subtypes`. It removes every land type from the affected permanent, keeps every other subtype (CR 205.1a), and adds the named basic land types. Its mana abilities come from the existing intrinsic path, which needs no change.
3. **The fix.** `SetsBasicLandType` uses the same land-types-only replacement. Dryad Arbor keeps "Dryad" under Magus of the Moon. The fix is in the same PR because it is the same function.
4. **The choice.** `effects.ChooseBasicLandTypeThen(chooser, options, then)` is `PickOption` over `BasicLandTypes`, or over a narrower list (Tundra Kavu's "a Plains or an Island"). Vision Charm's "a land type and a basic land type" asks twice, the first time over all 17.
5. **Durations and sets.** These are the existing ones. A single target is pinned (`PinObject`). "Each land you control" and "all lands" are pinned at resolution as a set (CR 611.2c). "Until this creature leaves the battlefield" is `DurationWhileSourceRemains`. "Until its controller's next untap step" is `UntilYourNextTurn` read for the land's controller, because the untap step is the first step of that player's turn. "Indefinitely" is `IndefiniteDuration()` pinned to the land.
6. **Adding a type for a while** (Navigator's Compass, Energybending) is the existing `ModAddSubtypes` with an until-end-of-turn duration. It needed only the choice.
7. **What the table sees.** A chip on the land: "Island until end of turn — Tidal Warrior".

#### Snapshot impact

One new kind, `setBasicLandTypes`, an on-disk identity inside `scopedEffects`, additive under schema v7 (ADR 0041 P10). It reads the existing `subtypes` field, so no `Mod` field is added. An older binary refuses a file that names it with `ErrUnknownEffectKey`, which is the designed rollback case. One fixture goes into `testdata/snapshots/v7/`.

#### Cards

About 30 of 33. The ? cards are Deepwood Elder (X targets from an activated {X}), Vision Charm (the second mode's choice over all land types) and Gaea's Liege, Orcish Farmer and Cyclopean Giant (each for its other text). Dreamwinder, Floodchaser and Kukemssa Serpent, the row's `Waiting` cards, land Full. Under owner decision 4, the 12 static land-type Auras and enchantments that `SetsBasicLandType` already expresses land too: Spreading Seas, Sea's Claim, Evil Presence, Tainted Well, Lingering Mirage, Contaminated Ground, Lush Growth, Convincing Mirage, Phantasmal Terrain, Blood Moon, Harbinger of the Seas and Illusionary Terrain.

### 2. Counter-held durations, and losing all land types (#1604)

#### What exists, what is missing

- `Duration` (`game/duration.go`) has one `Condition` field. Its values are `WhileSourceOnBattlefield`, `WhileYouControlSource`, `WhileYouControlSourceOnceItLands` and `WhileSourceRemainsTapped`. They are bare ints in the snapshot, and the comment says they are "appended, never inserted". Nothing reads a counter.
- `durationExpiredLocked` is the one expiry function. The layer listener already bumps the layer version on `EventCounterPlaced`, which also fires on removal (`game/layer_listener.go`). So the sweep would see a counter leave without new plumbing.
- Ultima's layer-6 half exists: `LoseAllAbilitiesMod()` plus a granted "{T}: Add {C}" bundle (`TapForManaGrant`, used by Urza's Saga), written as one record by `effects.GrantAbilitiesFor{…, Also}`, which applies `Also` before the grant (ADR 0046 §2).
- No kind removes land types, and there is no land-type list to remove (§1).
- **Teferi's Talent** is catalogued with a caveat. Its draw trigger works. Its granted −12 is missing, and three things stand in the way:
  - ADR 0093 Decision 10 put "granted loyalty abilities" out of scope because no waiting card needed them.
  - The engine would mostly handle one. `ActivateCatalogAbility` reads `ActivatedAbilitiesWithOrigins` (own and granted), `ActivationAbilityOf` sets `Loyalty` from the cost, and the CR 606.3 and 606.6 checks run on any row. `checkGrantAbilities` does not refuse a loyalty cost.
  - `CreateEmblemForEffect` (`game/emblem.go`) keys the emblem on the *source card's* catalog key. A granted −12's source is the planeswalker, not the Talent, and no grantor reaches effect code (`AbilityOrigin.GrantedBy` exists only in the origins list).

#### The rules

- **CR 611.2b:** a "for as long as" duration ends the moment its condition stops holding. If it never starts, the effect does nothing, and it does not come back.
- **CR 205.3i, 613.1d:** "loses all land types" is a layer-4 removal of the 17 land types. **CR 613.1f:** "loses all abilities and has …" is layer 6, removal before the grant in timestamp order.
- **CR 606.3:** one loyalty ability of a permanent per turn, at sorcery speed. A granted loyalty ability is still a loyalty ability of that permanent.
- **CR 114.2:** "[Player] gets an emblem" puts it in that player's command zone. The emblem's text is the granted ability's, so it is the grantor's (the Talent's) emblem.

#### Decision

1. **The condition.** `WhilePinnedHasCounter` (value 4), with a new `Duration.CounterKind`. It holds while the pinned object (instance and entry stamp) is on the battlefield with at least one counter of that kind. It is checked in `durationConditionHoldsLocked` beside the others.
2. **The kind.** `loseLandTypes`, layer 4, reads nothing. It removes every subtype for which `IsLandType` is true. The static form, `effects.LosesAllLandTypes(applies)`, serves Lithoform Blight and Alpine Moon.
3. **Ultima** is one record: `loseLandTypes`, `loseAllAbilities` and the {C} grant, pinned to the land and timed by `WhilePinnedHasCounter{blight}`.
4. **Granted loyalty abilities** (owner decision 2). ADR 0093 Decision 10's loyalty exclusion is lifted:
   - A granted loyalty row is an ordinary granted activated row. It shares the host's CR 606.3 once-per-turn count, because the engine already counts per permanent.
   - The stack item names its grantor. `ActivateCatalogAbility` copies the row's `AbilityOrigin.GrantedBy` onto the item, and `effects.CreateEmblem{}` keys on the grantor when there is one. This is the only place that needs it today.
   - `Register` needs no change, because `checkGrantAbilities` already accepts a loyalty cost. A test asserts that a granted −N can't be activated with fewer than N loyalty counters (CR 606.6).
5. **What the table sees.** The land's chip reads "No land types, no abilities, '{T}: Add {C}' while it has a blight counter — Ultima". A granted loyalty row shows in the planeswalker's menu like any granted row.

#### Snapshot impact

- `WhilePinnedHasCounter` is a new condition value and `CounterKind` a new `Duration` field. Both are additive under v7. In `scopedEffects` and `delayedTriggers` an older binary refuses them: `Known()` refuses the value and the unknown-field scan refuses the field. Shared machinery 2 closes the gap for the other three places a `Duration` is stored.
- `loseLandTypes` is a new kind, refused the same way. One fixture per kind goes into `v7/`.
- The grantor on a stack item is a new top-level `StackItem` field. Every binary since ADR 0041 phase 3 refuses an unknown stack-item field in a current-schema file, so it needs no bump. Record it with `-update-shape`.

#### Cards

- **Counter-held.** About 14 of the 17. The ? cards are Cyclopean Tomb (its leaves trigger tracks which counters it put), Eluge (a cost reduction that counts flooded lands), Immortal Obligation and Promise of Loyalty (per-player attack restrictions), Quicksilver Fountain (each player targets their own land) and Dread Wight. Dread Wight writes the new condition into an untap hold, so it lands only after PR 2 has reached `main` (Shared machinery 2).
- **Losing land types.** Ultima, Lithoform Blight and Alpine Moon, all three.
- **Granted loyalty** (owner decision 2). Teferi's Talent becomes Full, and Elspeth's, Liliana's, Rowan's and Vivien's Talent land. Rowan's Talent copies the ability (Rings of Brighthearth's shape), so it is the ? card.

### 3. For as long as you control this and it remains tapped (#1894)

#### What exists, what is missing

- `game.ForAsLongAsYouControlDuration(source, player)` (`WhileYouControlSource`; card side `effects.DurationWhileYouControlSource`) and `game.ForAsLongAsSourceTappedDuration(source)` (`WhileSourceRemainsTapped`) each exist. Each returns "never started" when its condition is already false. `durationConditionHoldsLocked` tests one or the other, never both.
- The only users are untap holds: `TapAndHoldWhileThisRemainsTapped` (Rust Tick, Amber Prison) and `TapAndHoldWhileYouControlThis` (Ty Lee, Tidebinder Mage). No control-change card uses either.
- `effects.GainControl{Target, Controller, Duration}` steals for any duration. The untap opt-out (`mayChooseNotToUntapSelf`, `game/untap_choice.go`) and Seasinger's state trigger (`effects.WhenYouControlNo`) exist.
- ADR 0058 already noted that Willow Satyr, Rubinia and Hivis "need a condition that is both of these at once".

#### The rules

- **CR 611.2b:** the effect ends when the condition stops holding. With two conditions joined by "and", it ends when either stops.
- **The Old Man of the Sea wording** adds a comparison: "that creature's power remains less than or equal to this creature's power", read continuously.

#### Decision

There is one reasonable design.

1. **The conjunction.** `Duration.Also []DurationCondition`. Every condition in it must hold as well as `Condition`. A list rather than a combined enum value, because three of these cards need three different pairs, and a combined value per pair would grow without end.
2. **Two more conditions**, appended after §2's:
   - `WhilePinnedRemainsTapped` (5): the pinned object stays tapped (Zygon Infiltrator's and Braided Net's "for as long as that creature remains tapped").
   - `WhilePinnedPowerAtMostSource` (6): the pinned object's power is at most the source's, both read live (Old Man of the Sea).
3. **Constructors.** `effects.WhileYouControlThisAndItRemainsTapped(ctx)` builds the Seasinger pair. It returns "never started" if either half is already false, so the card registers nothing (CR 611.2b).
4. **One sweep.** All conditions are read in `durationConditionHoldsLocked`, so the existing tap, untap and counter bumps re-sweep them.

#### Snapshot impact

`Also` is a new `Duration` field and the two conditions are new values, all additive under v7 and refused by older binaries in `scopedEffects` and `delayedTriggers`, exactly as §2's are. The control change itself is the existing `setController` record.

#### Cards

About 7 of 8: Seasinger, Helm of Possession, Hivis of the Scale, Rubinia Soulsinger, Willow Satyr, Old Man of the Sea and Zygon Infiltrator. Braided Net (craft) is the ? card. Seasinger, the row's `Waiting` card, lands Full. Under owner decision 4, the 33 single-condition "remains tapped" cards land with them (Vedalken Shackles, Endoskeleton, the Couriers, Mana Leech and the rest). About 25 need nothing else. Thran Weaponry waits on echo (ADR 0108 PR 4), and Entrancing Lyre on §9's power bound.

---

## Group I — rule gates

The three gates share one shape. Each is one function, read at the action, by the legal-move enumerator and by the view, so a card the gate refuses is never offered (ADR 0033 §1, ADR 0105). None of them writes state, except §4's "this turn" form.

### 4. Players can't play lands (#1895)

#### What exists, what is missing

- `CastGateLocked` (`game/cast_gate.go`) says outright: "A land PLAY is not a cast (CR 305.1, CR 116.2a) and is not gated here." `castSpellLocked` calls it only for a nonland (`game/mutations.go`).
- A land is played by two paths, and both read `LandDropsRemainingLocked` (`game/land_drops.go`):
  - the land branch of `castSpellLocked`, which also takes plays from other zones through `CastPermissionForClaimLocked` (Glacierwood Siege, play from exile, play from the top of the library);
  - `CanPlayLandDuringResolutionForEffect` and `PlayLandDuringResolutionForEffect` (`game/hideaway.go`), for a land played while a spell resolves. The registry row misses this path.
- The enumerator's `landPlayMove` (`legal/cast.go`) checks only timing and drops. The view's `castableNow` (`protocol/view.go`) does the same, and `cant_cast` is stamped only on nonlands, so a refused land gets the generic tooltip ("Can't play a land now", `client/src/lib/dragCast.ts`).
- "This turn" cast bans are `CastBanRule` `PlayerStatic`s (`game/cast_ban.go`, `effects.RestrictCasting`). They do not cover lands.

#### The rules

- **CR 305.1, 116.2a:** playing a land is a special action, not a cast.
- **CR 101.2:** "can't" beats "can". Its own example: "You may play an additional land this turn" loses to "You can't play lands this turn". So the gate is asked before the drop count, and extra drops do not lift it.
- **CR 305.2a:** lands played during a resolution count against the drops, and a ban stops them too.
- **CR 206.3a** lists every name City in a Bottle means.

#### Decision

There is one reasonable design.

1. **The static.** `Spec.LandPlayRestrictions []game.LandPlayRestriction`, each with `{Label, Forbids func(LandPlayQuery) bool, ActiveWhen}`, mirroring `CastRestriction`. `LandPlayQuery` carries the game, the land, the player, the source permanent and the zone the land is played from. `Register` refuses an empty label or a nil `Forbids`, as it does for cast restrictions. Constructors: `PlayersCantPlayLands`, `YouCantPlayLands`, `OpponentsCantPlayLandsFrom(zones…)` (Tomik) and `CantPlayLandsNamed(names)` (City in a Bottle, with `game.ArabianNightsNames`, CR 206.3a's list, which the cast half reads too).
2. **The "this turn" form.** A new `ScopedEffect` kind, `cantPlayLands`, reader `rule`, `ScopeGame`, reads `Player` (Turf Wound, Solfatara, Pardic Miner, Moonhold). It is a `ScopedEffect` rather than a fifth `CastBanRule` kind on purpose. An older binary does not check `CastBanKind` on restore and would quietly ban nothing, while it refuses an unknown mod kind with `ErrUnknownEffectKey`.
3. **The gate.** `LandPlayGateLocked(player, card, fromZone)` walks the battlefield restrictions (through `CatalogAbilityKey`, so a land-play restriction that loses its abilities stops), emblems (§5) and the live `cantPlayLands` records, and returns a `*CantPlayLandError{Reason, Source}`. It is called by:
   - the land branch of `castSpellLocked`, before the drop count;
   - both hideaway entry points;
   - the enumerator's `landPlayMove`;
   - the view's `castableNow`. The view also stamps the reason on the land's existing `cant_cast` field, so the client's tooltip says "Players can't play lands — Territorial Dispute" with no client change beyond reading the field for lands.
4. **What the table sees.** The tooltip above. A banner line for a "this turn" ban: "Bob can't play lands this turn — Turf Wound".

#### Snapshot impact

One new kind, `cantPlayLands`, refused by older binaries, with one fixture. The statics are catalog data.

#### Cards

About 11 of 14: Aggressive Mining, City in a Bottle, Moonhold, Pardic Miner, Rock Jockey, Solfatara, Territorial Dispute, Turf Wound and Ward of Bones need nothing else. The ? cards are Damping Engine (its "sacrifice a permanent to ignore this effect until end of turn"), Experimental Frenzy (playing from the top of the library, which the PR checks against ADR 0066's permissions), Shaman's Trance (casting from other players' graveyards), Tomik (needs §6 too; it lands in whichever PR merges second) and Worms of the Earth ("Lands can't enter the battlefield" is a separate prohibition; see Out of scope). City in a Bottle, the row's `Waiting` card, lands Full.

### 5. Emblems that restrict players (#1899)

#### What exists, what is missing

- `effects.EmblemSpec` (`effects/emblem.go`) has `Static`, `Triggered`, `UntapStep`, `DrawStep` and `ActivationTimings`. It has no cast restriction slot. `CastGateLocked` walks only the battlefield, so an emblem's "can't cast" would be dead text.
- Emblems are cards in `Player.Emblems`. Five readers walk them (layers, triggers, the untap step, the draw step and activation timing). The activation-timing walk (#1275) is the model.
- `game/game_end_gates.go` says that emblems "would be a third source" of game-end gates and that "the reader gains a loop when that card is catalogued". That card is Gideon of the Trials.
- Narset Transcendent's other two abilities:
  - The +1 is `TakeFromLibraryToHand{Cards: LookAtTopOfLibraryForEffect(p, 1), Match: …, Max: 1, Optional: true, Reveal: true}`.
  - The −2 needs a "from your hand" form of `effects.WhenYouNextCast`. Its condition `youNextCast` does not check the zone. Its body must read the spell from `ctx.PayloadCards()`, because a delayed-trigger item has no `Trigger` and `ThatSpellGains` reads `ctx.Trigger()`. It then calls `GrantKeywordsToSpellForEffect` with rebound.

#### The rules

- **CR 114.4:** the abilities of emblems function in the command zone. **CR 114.2:** an emblem is owned and controlled by the player who got it.
- **CR 101.2:** the ban beats any permission to cast.
- **CR 800.4a:** when a player leaves the game, every object they own leaves with them, emblems included.

#### Decision

1. **Emblem slots follow the gates.** `EmblemSpec` gains `CastRestrictions`, `LandPlayRestrictions`, `GameEndGates` and `UntapCaps`, the same types as the `Spec` slots of the same names. `buildEmblemDef` copies each one onto the emblem's `CardDef`, and `checkEmblemSpec` counts them as content.
2. **Each gate walks emblems.** `CastGateLocked`, §4's `LandPlayGateLocked`, `forEachGameEndGateLocked` and the untap-cap reader each gain one more walk, over `p.Emblems.Cards`, with the emblem as the source, so `OpponentsCantCast` reads its controller from the emblem.
3. **Why not `CastBanRule`.** The tracker suspected that Narset's emblem was already a `CastBanRule`. It can say the same words, but a `CastBanRule` lives on the banned player, so:
   - the emblem would be one record per opponent, frozen when created;
   - the board would show no emblem;
   - CR 800.4a would not end it, because nothing sweeps the statics a leaving player wrote onto others.

   The emblem is the object CR 114.4 says the ability is on, so the gate reads it there.
4. **Narset's −2.** There is a new delayed condition key, `cast/you-next-cast-from-hand`, which tests `ev.OldZone == ZoneHand`. There is a new delayed body key that reads the spell from the payload and grants rebound. Both keys are on-disk identities.

#### Snapshot impact

None for the slots: emblems are already captured (`PlayerSnapshot.Emblems`), and their abilities are catalog data. The two delayed keys are new vocabulary, and an older binary refuses a file that names an unknown delayed key.

#### Cards

About 2 of 3. Narset Transcendent, the row's `Waiting` card, lands Full. Dovin Baan's emblem ("Your opponents can't untap more than two permanents") is the ? card: an `UntapCap` scoped to opponents may need a field. Gideon of the Trials waits on ADR 0108 PR 6 for its +1 ("prevent all damage target permanent would deal"). Its emblem is ready after this PR.

### 6. Cards in graveyards can't be targeted (#1885)

#### What exists, what is missing

- `canBeTargetedBy(g, c, zone, src)` (`game/keywords.go`) returns true at once for any zone but the battlefield. It then reads shroud, hexproof (with the #1560 bypass) and protection.
- Every target check reaches it through two choke points in `game/targets.go`: `specMatchesLocked`, which enumerates targets for `LegalTargetsForEffect`, the enumerator and the view, and `specMatchLocked`, which is behind `targetLegalLocked` at announce (CR 601.2c) and at resolution (CR 608.2b). A cost payment passes `targeting=false`, because paying a cost is not targeting.
- `game/hexproof_bypass.go` is the precedent for a table-wide static read live at targeting, with no state.

#### The rules

- **CR 601.2c:** a player announces "an appropriate object or player" for each target, obeying any rule that says something can't be chosen. **CR 608.2b:** a target that has become illegal is illegal at resolution. **CR 101.2:** "can't" wins.

#### Decision

There is one reasonable design.

1. **The static.** `Spec.TargetingRestrictions []game.TargetingRestriction{Label, Zone, Forbids func(TargetingQuery) bool}`. `TargetingQuery` carries the candidate card, its zone and the spell or ability's controller. Constructors: `CardsInGraveyardsCantBeTargeted()` and `OpponentsCantTarget(zones, match)` (Tomik: "lands on the battlefield and land cards in graveyards … your opponents control").
2. **The read.** `canBeTargetedBy` asks `targetingRestrictedLocked(c, zone, src)` before its battlefield guard. That covers announce, the CR 608.2b re-check, the enumerator and the view at once, through the two existing choke points.
3. **What the table sees.** A graveyard card is simply not offered. The graveyard viewer shows a banner line: "Cards in graveyards can't be targeted — Ground Seal".

#### Snapshot impact

None. The restriction is read live off the battlefield.

#### Cards

All five: Dennick, Pious Apprentice // Dennick, Pious Apparition (the row's `Waiting` card, Full), Ground Seal, Silent Gravestone, Underworld Cerberus, and Tomik (after §4).

---

## Group C — cost components and mana

### Shared machinery: a payment fact reaches the effect

§7, §8 and §9 are the same idea three times. A cost's payment is a fact the ability's effect or its targets may read (CR 400.7j: "If the cost of a spell or ability causes an object to move to a public zone, that spell or ability's effects can find that object"). Each fact rides `StackItem.Paid` (`game/paid_cost.go`), which is already captured. So:

- the activated-ability payer (`activateCatalogAbilityLocked`) records every card its cost moved, as the spell payer already does;
- `copiedPaidCost` (`game/ability_copy.go`) copies every recorded fact, so a CR 707.10 copy reads what the original paid;
- a new `PaidCost` field is avoided where an existing one fits. Unknown keys inside `paid` are *not* refused on restore (the stack-item scan stops at the top level), so a new one would be dropped silently by an older binary.

### 7. Costs from a library, or onto it (#1902)

#### What exists, what is missing

- `game.AbilityCost` (`game/activated.go`) has `Tap`, `TapOthers`, `Waterbend`, `SacrificeSelf`, `SacrificeOther`, `Mana`, `Life`, `LifeFrom`, `Loyalty`, `Crew`, `RemoveCounters`, `AddCounter`, `MinX`, `DiscardSelf`, `DiscardCards`, `ReturnToHand`, `ExileSelf` and `ExileCards`. It has nothing that moves a card from a hand to a library, or exiles from a library. `ExileCostZoneSupported` allows only a hand or a graveyard.
- Cost choices are named at announce in `ActivateAbilityParams` (`discard_ids`, `exile_ids` and the rest, decoded in `actions/actions.go`), because CR 602.2b makes activation one indivisible step.
- The enumerator (`legal/abilities.go`) offers one payment per cost component. The view stamps `DiscardCost*` and `ExileCost*`. The client chains pickers in `Board.svelte:handleActivateAbility`.
- The closure ratchet counts `AbilityCost` routes to a func (`census:IntrinsicAbilityCards`, ceiling 49, `closure_fields_test.go`). A component with a `Match` func would be a new route over the ceiling.
- Both waiting cards' shields shipped in ADR 0107 PR 7 (`effects.PreventNextDamageFromChosenSource`). The cost is all they lack.
- There is no "discard at random" cost (`game.DiscardCost` has `N`, `Label` and `Match`).

#### The rules

- **CR 602.2b, 601.2h:** costs are paid in any order, except that costs with random elements, or that move objects from the library to a public zone, are paid after all the others.
- **CR 118.3:** a cost can't be paid without the resources. "Exile the top four cards of your library" can't be paid from a library of three.
- **CR 400.7j:** the effect can find the cards the cost exiled (Phyrexian Devourer's "the exiled card's mana value", Storm Elemental's "if the exiled card is a snow land").
- **CR 701.9b:** a random discard is not the player's choice.

#### Decision

1. **Exile the top N.** `AbilityCost.ExileFromLibraryTop int`. There is nothing to choose, so it carries no func, no route, no params field and no wire field. It is validated with the other costs (`len(Library) >= N`, CR 118.3), paid last (CR 601.2h) through `routeCardToZoneLocked` with `MustSettleNow` and the commander answer, and recorded on the existing `PaidCost.Exiled`.
2. **A card from your hand on top of your library.** `AbilityCost.PutFromHandOnLibraryTop int`. All three printed cards say "a card", so it has no predicate and no route.
   - It needs one new params and wire field, `top_ids`.
   - Its candidates are the hand minus the source, and a card can't also pay a discard or an exile.
   - It is paid through `routeCardToZoneLocked` to the top of the library.
   - The activator still knows the card it put there, so `KnownBy` keeps them.
   - It is not a discard, so it fires no discard trigger and no madness, and it is not `DiscardCost`.
3. **Discard at random** (owner decision 3). `DiscardCost.Random bool`, a bool on a struct already on the ratchet's list, so no new route. The engine chooses the card with the game's seeded random source when the cost is paid, last (CR 601.2h), and records it on `PaidCost.Discarded`. The client shows a confirm, not a picker, and the enumerator offers one payment.
4. **The enumerator, the view and the client.** One payment each: the top N needs a gate only, and the hand card is `cheapestFuelFirst` over the hand. The view stamps a count and a label for each new component, and the client reuses the discard modal for the hand pick with its own heading.

#### Snapshot impact

None. The library cards ride the existing `Exiled`, and the random discard rides `Discarded`.

#### Cards

- **The 10.** About 9: Penance and Seasoned Tactician (the row's `Waiting` cards, Full), Leashling, Arc-Slogger, Whirling Catapult, Royal Herbalist, Storm Elemental, Phyrexian Devourer and Hidden Retreat. Hidden Retreat's "prevent all damage that would be dealt by target instant or sorcery spell this turn" is ADR 0108 §7's `preventFromSource`, so it lands after ADR 0108 PR 6. Thought Lash is the ? card: its cumulative upkeep exiles from the library, a pay-unless payment ADR 0108 PR 4 does not teach.
- **Discard at random** (owner decision 3). About 15 of 17: Mage il-Vec, Frenetic Ogre, Canyon Drake, Coral Helm, Meteor Storm, Pardic Swordsmith, Stormbind, Amok, Pyromania, Hell-Bent Raider, Dwarven Strike Force, Draconian Cylix, Ogre Shaman, Pardic Lancer, and Pyromancy and Stormscale Anarch, which read the discarded card through §8. Barbarian Bully ("unless a player has this creature deal 4 damage to them") is the ? card.

### 8. The card discarded to pay, and the world rule (#1862)

#### What exists, what is missing

- `PaidCost.Discarded` is set only by the spell path (`additional_cost.go:paidWithBranchAndDiscards`) and read by `Context.Discarded()`. Its one reader is Grab the Prize.
- `activateCatalogAbilityLocked` pays a discard with `payAbilityDiscardsLocked` and records nothing. Its neighbours `paid.Exiled` and `paid.Sacrificed` are recorded. That is why Ellie and Alan and Holistic Wisdom (exile readers) and Greater Good and Birthing Pod (sacrifice readers) are catalogued and no discard reader is.
- `copiedPaidCost` drops `Discarded` (Shared machinery).
- **The world rule is not implemented.** `stateBasedActionsLocked` (`game/mutations.go`) has no CR 704.5k arm. "World" is parsed as a supertype, so `Card.HasSupertype("world")` answers.
- `Card.EnteredBattlefieldAt` is a wall-clock stamp (`timeNowUnixNano`). It is neither unique nor shared by a simultaneous move, so it can't decide CR 704.5k's tie.
- Any-player activation exists (`ActivatedAbilityShape.AnyPlayer`, ADR 0106). Land's Edge's discard comes from the activator's hand, which is right.

#### The rules

- **CR 400.7j:** the effect finds the card the cost moved.
- **CR 704.5k:** if two or more permanents have the world supertype, all but the one that has had it for the shortest time go to their owners' graveyards, and on a tie for the shortest time they all do. **CR 205.4f:** any world permanent is subject to it.
- No effect grants the world supertype, so "had the supertype for the shortest time" is "entered the battlefield most recently".

#### Decision

There is one reasonable design.

1. **Record it.** `activateCatalogAbilityLocked` appends the paid discards to `paid.Discarded`, as the spell path does. "Discard this card" (cycling's `DiscardSelf`) is recorded too, because it is a discarded card and CR 400.7j does not distinguish them.
2. **Copy it.** `copiedPaidCost` copies `Discarded`, `Exiled`, `Sacrificed` and `ReturnedAttacking`, so every recorded fact survives a copy.
3. **The world rule.** `Card.EntryOrdinal int64` is taken from a game counter once per move that puts permanents onto the battlefield. Every card a single simultaneous move puts there shares it (Replenish), and every later move gets a strictly larger one. The state-based action keeps the world permanent with the largest ordinal when exactly one has it, and puts every other one, all of them on a tie, into its owner's graveyard. It is not a choice, so it needs no pending choice. The three caveats go.

#### Snapshot impact

`Card.EntryOrdinal` is an additive card field, recorded with `-update-shape`. The game counter is not a top-level key. It is cloned with the game, and restore sets it to the largest ordinal any card names, as it does `scopedEffectSeq`. An older binary drops the field, which is harmless, because it has no world rule to read it. `Discarded` is already in the shape.

#### Cards

- **The 7 discard readers,** all of them: Land's Edge (the row's `Waiting` card, Full), Hisoka, Minamo Sensei, Mercurial Chemister, Moria Scavenger, Necromancer's Stockpile, Slumbering Tora and Volrath the Fallen.
- **The world permanents.** The three caveats go: Caverns of Despair, Concordant Crossroads, Forsaken Wastes. The other 22 uncatalogued world permanents are checked, and about 12 need nothing else. The rest are the old rules-heavy ones (Chaosphere, Storm World, Teferi's Realm, Revelation, Mystic Decree and others).

### 9. A target bounded by counters removed, or by X (#1842)

#### What exists, what is missing

- `effects.RemoveCountersXFromThis(kind, min)` is `CounterRemovalCost{Variable: true}`. The count is announced in `ActivateAbilityParams.CounterCounts` and recorded as `paid.CountersRemoved` (read by `Context.CountersRemoved()`). It is validated before targets are bound, so it is known in time (CR 601.2c).
- The announced-X bounds are `TargetSpec.ManaValueAtMostX` and `ManaValueEqualsX`, bound by `bindStepsX` at announce and again for the CR 608.2b re-check in `clauses.go:itemAnnouncedClauses`. They are mana value only. `checkNoXBound` (`effects/registry.go`) refuses one unless the cost `DemandsX()`, and a counter removal does not.
- No bound is on power or toughness. `PowerLE(n)` is a fixed predicate.
- The enumerator's X ladder supports only a mana {X} (`legal/abilities.go`). The view stamps `ManaValueAtMostX` but not `ManaValueEqualsX`, so the client offers Lazav a superset. The server still refuses a wrong pick. This PR fixes both.
- Ruthless Technomancer's caveat says its reanimation is missing because "Sacrifice X artifacts" can't be paid. Since ADR 0100 that cost exists (`SacrificeCountFromX`), and the remaining gap is the power bound.

#### The rules

- **CR 107.3a, 601.2b:** X in a cost is announced while casting or activating, and is the announced value on the stack. **CR 602.2b** applies this to activation.
- **CR 400.7j and Simic Manipulator's wording:** "the number of +1/+1 counters removed this way" is a fact of the payment, known before targets are chosen.
- **CR 608.2b:** the bound is checked again at resolution.

#### Decision

There is one reasonable design.

1. **One bound, two inputs.** The unexported `xBound` becomes a bound with a statistic (`manaValue`, `power`, `toughness`), a comparison (`atMost`, `equals`) and a source:
   - `announcedX`, the item's `XValue`, which a mana {X}, a variable sacrifice and a variable tap already set;
   - `countersRemoved`, the item's `Paid.CountersRemoved`.

   The card-side flags are `ManaValueAtMostX`, `ManaValueEqualsX`, `PowerAtMostX` and `ToughnessAtMostX`, plus `BoundByCountersRemoved`, which switches the source.
2. **Binding.** At announce, from the value the payment just computed. At resolution, from the item. Nothing new is stored, because the `TargetSpec` is rebuilt from the catalog row.
3. **The registry check.** `checkNoXBound` accepts `BoundByCountersRemoved` when the cost has a variable counter removal, and refuses it otherwise.
4. **The enumerator.** For a counter-sourced bound, the innermost loop over counter payments filters targets per payment, through `TargetsWithinXForEffect`, as the cast path does. For a "≤" bound the largest payment admits the most, so the existing "offer the largest" stays right.
5. **The view and the client.** The view stamps the bound's statistic and comparison and a `powers` (or `toughnesses`) map. `client/src/lib/targeting.ts:withinX` reads the X from the counter payment when the bound says so.

#### Snapshot impact

None.

#### Cards

About 8 of 9: Simic Manipulator (the row's `Waiting` card, Full), Quillmane Baku (mana value from counters), Killing Glare, Minamo Sightbender, Aryel, Knight of Windgrace, Finale of Eternity, Entrancing Lyre (with §3's untap hold) and Ruthless Technomancer (the caveat goes). Gang Up is the ? card, because assist is not built. Board-count bounds (Beguiler of Wills, Legacy's Allure, Goma Fada Vanguard) are live predicates over the board and are not this seam.

### 10. Riot and unleash (#1556)

#### What exists, what is missing

- Neither riot nor unleash is in `canonicalKeywords` (`game/keywords.go`), and nothing in `game/` reads either.
- The entry replacement window (`gatherActiveReplacementsLocked`, `game/replacements.go`) collects the entering card's own replacements through `CatalogKey(entering)`, which folds in copy grants but not layer-6 grants. The entering card has no layer pass yet. The one look-ahead that exists is for a card entering as a copy (ADR 0102 decision 6). A static on another permanent that grants an ability is never seen by a card entering under it, and the ETB harvest catches up only after landing (`game/triggers.go`, CR 603.6a).
- The in-window question shapes are `Optional`, `EntryLifeCost`, `EntryCardChoice`, `CopySelector` and `EntryController`. `Optional` cannot carry riot: "no" runs no replacement, but riot's "no" grants haste. `EntryController` is the precedent for a mandatory entry choice with a default.
- A permanent can gain haste indefinitely as data: a `ScopedEffect` with `AddKeywordsMod("haste")`, pinned (`game/suspend.go`, `game/earthbend.go`).
- Uncivil Unrest gives the counter every time (`b28PermanentsYouControlEnterWithACounter`). "Creature spells you control can't be countered" exists (`SpellsYouControlCantBeCountered`, Prowling Serpopard).
- A paused entry window is not a restore point (`replacementResume` is in the continuation census). Only the result has to persist.

#### The rules

- **CR 702.136a:** riot is a static ability: "You may have this permanent enter with an additional +1/+1 counter on it. If you don't, it gains haste." **CR 702.136b:** multiple instances each work separately.
- **CR 702.98a:** unleash is two statics: "You may have this permanent enter with an additional +1/+1 counter on it" and "This permanent can't block as long as it has a +1/+1 counter on it."
- **CR 614.1c, 614.12:** both are entry replacements. Which apply is decided by "the characteristics of the permanent as it would exist on the battlefield", taking into account its own statics and "continuous effects that already exist and would apply to the permanent". So Rhythm of the Wild's grant gives a creature riot as it enters, and a creature entering under Dress Down ("Creatures lose all abilities") has no riot to use. **CR 614.12a:** the choice is made before it enters.
- **CR 400.7a:** a keyword granted to the creature spell carries onto the permanent (Domri, Chaos Bringer's "it gains riot").

#### Decision

1. **The look-ahead** (owner decision 1). `entryLookAheadLocked(entering, ev)` computes the entering permanent's layer-4 and layer-6 result as it would exist on the battlefield. It is a dry run of the layer pass over the board plus the entering card, under its would-be controller, with its stack-granted keywords. Nothing is written. It reads keywords (counting instances, for CR 702.136b) and ability removal, and nothing else. The copy look-ahead becomes its first step.
2. **Keyword-shaped entry replacements.** Riot and unleash join `canonicalKeywords`. The gather adds one entry replacement per instance the look-ahead reports, so printed riot plus Rhythm of the Wild asks twice, and a creature that would enter without abilities is not asked at all.
3. **The riot question.** A new in-window kind, `entry_riot`, mandatory, with two answers:
   - the counter, which adds one to `EntersWithCounters`;
   - haste, which sets a new `ReplacementEvent.EntersWithHaste`. After landing it becomes an `addKeywords` haste record pinned indefinitely to the permanent, the earthbend shape.

   An entry that can't pause takes the counter. That is the outcome every riot creature in the catalog gets today, and it is the default `EntryController` uses for the same situation.
4. **Unleash** is the existing `Optional` entry counter. Its second half is a can't-block rule read live by the block gate, for a permanent with the unleash keyword and a +1/+1 counter.
5. **Grants.** Rhythm of the Wild and Spider-Punk are a layer-6 keyword grant (`b16GrantKeywords`) of riot, and Tesak of unleash. Uncivil Unrest moves to the same grant, and its caveat goes.
6. **The bot.** It takes haste when the creature enters during its controller's turn before combat damage and could attack, and the counter otherwise. It always takes the unleash counter unless the creature is its only untapped potential blocker on an opponent's turn.
7. **What the table sees.** A two-button prompt: "+1/+1 counter" or "Haste". A permanent that took haste shows the keyword chip, labelled "Riot".

Considered and not chosen (question 1, option b): printed riot read off the card's printed keywords, and each grantor carrying its own entry replacement in Uncivil Unrest's shape.

#### Snapshot impact

None new. The haste is an existing `addKeywords` record. `entry_riot` lives only in the paused window, which is not a restore point.

#### Cards

About 27 of 29: the 13 riot creatures (Spider-Punk is the ? card, for "Spells and abilities can't be countered", which nothing gates for an ability), the 14 unleash creatures, Rhythm of the Wild (the row's `Waiting` card, Full) and Domri, Chaos Bringer (after §11's spend grant). Uncivil Unrest's caveat goes.

### 11. Granted mana-spent readers (#1552)

#### What exists, what is missing

- The payment record (#761), the source snapshot (#1212, `ManaToken.SourceKinds`, carried by `Card.Provenance`, CR 400.7d) and spend riders (#1547) all exist. `ManaSpent` (`game/mana_spent.go`) answers `CountFrom(kinds)` and `None()`.
- `EntryCastCountsForEffect(ev)` (`game/entry_counters.go`) is the only reader a catalog replacement can call inside the entry window. Its `CastCounts` has `X`, `Kicked`, `ColorsSpent`, `ManaSpent` and `Delved`, and no per-source-kind count.
- Sunburst is a printed `EntryCountersFromCast` clause (`effects.SunburstCounters`, Etched Oracle), not a keyword. `applyCastEntryCountersLocked` reads printed clauses only.
- A keyword granted to a spell exists (`GrantKeywordsToSpellForEffect`, `effects.ThatSpellGains`, ADR 0107). It is re-pinned to the permanent with `PinnedToEpoch(IndefiniteDuration())`, so a duration it had is lost on entry.
- Satoru's trigger is `OncePerBatch` over a per-card predicate that accepts only uncast creatures (`b30NontokenCreatureYouControlEnteredUncast`).

#### The rules

- **CR 702.44a, 702.44b:** sunburst works as the object enters from the stack as a resolving spell, with a +1/+1 counter (a creature) or a charge counter (otherwise) per colour of mana spent. **CR 702.44d:** multiple instances each work separately.
- **CR 400.7a:** "it gains sunburst" given to the spell carries onto the permanent. **CR 400.7d:** a permanent's ability can read the mana spent to cast the spell it was.
- **CR 603.4:** Satoru's "if none of them were cast or no mana was spent to cast them" is an intervening "if" over the whole set, checked when it triggers and again on resolution.

#### Decision

There is one reasonable design.

1. **The spent record in the window.** `EntrySpentForEffect(ev) ManaSpent` returns the entering spell's whole spend, so a replacement on another permanent can ask `CountFrom(ManaSourceArtifact)` (Coin of Mastery) or count Treasure mana (Kalain).
2. **Sunburst is a keyword.** It joins `canonicalKeywords`, and `applyCastEntryCountersLocked` reads `HasKeyword(spell, "sunburst")` on the resolving stack card, so stack grants count, once per instance. Etched Oracle moves to it. Lux Artillery is `ThatSpellGains{Keywords: {"sunburst"}}` from its cast trigger. Lux's grant arrives after payment, so the auto-payer does not spread colours for it. That is correct (the colours are already spent) and is noted in the card's tooltip.
3. **A spell grant keeps its duration.** `GrantKeywordsToSpellForEffect` takes a `Duration`, and the re-pin to the permanent keeps its kind. This serves the haste-until-end-of-turn cards: Generator Servant and Carnelian Orb (as spend riders), and Tyvar Kell's emblem.
4. **Satoru** reads the batch. The condition is "no creature in the batch was cast with mana spent" (`!Spent().Known() || Spent().None()` per creature, `ManaOnPaper` answering "unknown"), checked when it triggers and again on resolution.
5. **Opal Palace's count** is a rider with a registered count key, the number of times the commander has been cast from the command zone.

#### Snapshot impact

None for the readers. A stack-pinned record with a non-indefinite duration is the existing shape. A new spend-rider kind for haste until end of turn is new vocabulary: the PR checks that an older v7 binary refuses a file naming it, as ADR 0107 PR 3 checked for stack pins, and adds the refusal if it does not.

#### Cards

About 9 of 10. Lux Artillery, Coin of Mastery and Satoru (the row's `Waiting` cards) lose their caveats. Kalain, Reclusive Painter, Generator Servant, Carnelian Orb of Dragonkind, Tyvar Kell and Freestrider Commando land. Opal Palace's caveat goes. Primeval Spawn is the ? card, for its leaves trigger's free casts. Under owner decision 4, the 15 printed sunburst cards land with the keyword (Engineered Explosives, Pentad Prism, Suntouched Myr, Infused Arrows and the rest).

---

## Shared machinery

1. **The land-type list** (§1, §2). `game.LandTypes`, `IsLandType` and `BasicLandTypes`, from CR 205.3i. Both new layer-4 kinds and the fixed `SetsBasicLandType` read them.
2. **Durations restore safely everywhere** (§2, §3). `Duration` grows `CounterKind` and `Also`, and three conditions. PR 2 extends `Duration.Known()` and the unknown-field scan to every place a duration is stored: `castPermissions`, `seats[].statics` and `cards[].nextUntapSkips[].while`, as well as `scopedEffects` and `delayedTriggers`. Until a binary with that check is the oldest one deployed, no card writes a new condition outside `scopedEffects` and `delayedTriggers`. The only card in this ADR that would is Dread Wight's untap hold, and it lands after PR 2 reaches `main`.
3. **One gate function per rule** (§4, §5, §6). Each gate is asked by the action, the enumerator and the view, with the same refusal text. Each walks battlefield statics, emblems and its live records. ADR 0033 §1 holds: the bot's move list and the client's highlights come from the same function.
4. **A payment fact reaches the effect** (§7, §8, §9). See the shared section in group C.
5. **The entry look-ahead** (§10, and §11 through the stack keywords). One function computes what an entering permanent would be, and the gather derives keyword-shaped entry replacements from it.
6. **Stale notes fixed with the code.** Each PR corrects the comments and registry notes its section found stale: Teferi's Talent, Lux Artillery, Coin of Mastery, Satoru, Ruthless Technomancer, the land-type and riot rows, Opal Palace's registry note, `cast_ban.go`'s header, `land_drops.go`'s "no catalog card declares AdditionalLandPlays", and `AnyCastRestrictionsForEffect`'s unused "fast negative".

## Delivery

Each PR lands its engine change and its cards together, test first. Each flips its registry row to implemented (or partial), adds a closed-seam fragment under `docs/engine-seams/closed/`, moves any card that needs more onto a `Waiting` list with the reason, and adds a registry row for any untracked family it finds. Every PR verifies every card against its full oracle text (ADR 0106 decision 6). The card estimates count only the cards that need nothing else, and include the cards owner decisions 2, 3 and 4 add (named in brackets).

| PR | Engine | Cards (est.) | Needs |
|---|---|---:|---|
| 1 | §1: CR 205.3i list, `setBasicLandTypes`, the `SetsBasicLandType` fix, `ChooseBasicLandTypeThen`, the chip | ~41 (~30, and ~11 static land-type cards, decision 4) | — |
| 2 | §3 and §2's condition: `Duration.Also`, `CounterKind`, the three conditions, the restore check on every stored duration | ~45 (~7, ~13 counter-held, and ~25 single "remains tapped", decision 4) | — |
| 3 | §2's `loseLandTypes` kind and its static | 3 (Ultima, Lithoform Blight, Alpine Moon) | PR 1, PR 2 |
| 4 | §2, owner decision 2: granted loyalty abilities, the grantor on the stack item | 5 (Teferi's Talent and the four other Talents) | — |
| 5 | §4: `LandPlayRestrictions`, `cantPlayLands`, `LandPlayGateLocked` at all four callers, the view's reason | ~11 | — |
| 6 | §5 and §6: the emblem slots and walks, Narset's two keys, `TargetingRestrictions` | ~7 (Narset, Dennick, Ground Seal, Silent Gravestone, Underworld Cerberus …) | Tomik needs PR 5 |
| 7 | §8 and §9: discards recorded on abilities, `copiedPaidCost`, the world rule and `EntryOrdinal`, the generalised target bound | ~27 (~7 discard readers, ~12 world permanents, ~8 bounded targets) | — |
| 8 | §7: `ExileFromLibraryTop`, `PutFromHandOnLibraryTop`, `top_ids`; owner decision 3's `DiscardCost.Random` | ~24 (~9, and ~15 random-discard cards, decision 3) | PR 7 (Pyromancy and Stormscale Anarch read the discard) |
| 9 | §10: the look-ahead, riot and unleash, `entry_riot`, `EntersWithHaste`, the bot arm | ~27 | — (Domri needs PR 10) |
| 10 | §11: `EntrySpentForEffect`, sunburst as a keyword, spell grants with a duration, Satoru's batch read | ~22 (~9, and ~13 sunburst cards, decision 4) | — (Domri needs PR 9) |

**The tracker's single PR for #1902, #1862 and #1842 is corrected to two.** #1862 and #1842 go together in PR 7: both make a payment fact reach the effect or the targets, neither needs a new wire field, and both touch `activateCatalogAbilityLocked`'s payment block and `itemAnnouncedClauses`. #1902 is PR 8 on its own. It is the only one with a new wire field, a new client picker and a new enumerator arm, and owner decision 3 more than doubles its cards. The two PRs touch the same payment block, so whichever merges second rebases. The conflict is mechanical.

**Amendment (PR 5, 2026-10-02): Tomik lands in PR 6.** Tomik's text needs both gates: the land-play restriction from §4 and the targeting restriction from §6. PR 5 builds the first and not the second, and a Tomik with half its text would be a card the engine plays wrongly, so PR 5 leaves it out (on the §6 row's `Waiting` list) and PR 6 lands it with `OpponentsCantPlayLandsFrom`, which PR 5 provides.

**Amendment (PR 6, 2026-10-02): what PR 6 found.** `TargetingRestriction` carries `Zones`, a list, rather than §6's single `Zone`, because Tomik's one clause names two zones (the battlefield and the graveyards). Dovin Baan's emblem needed no new `UntapCap` field: `Applies` already receives its source, which for an emblem is the emblem, so "your opponents" is `source.Controller != activePlayer`; Dovin lands Full. `AnyCastRestrictionsForEffect` had no caller left and is deleted rather than taught about emblems. Gideon of the Trials moves to the chosen-source shield row (#1904) for its +1.

**Order and parallelism.** PRs 1, 2, 4, 5, 6, 7, 9 and 10 can start at once. PR 3 follows PRs 1 and 2. PR 8 follows PR 7. PR 6 lands Tomik only if PR 5 has merged, and otherwise PR 5 lands it. Domri lands in whichever of PRs 9 and 10 merges second. PRs 1 and 3 both add cases to `modApply` in `scoped_effects.go`, and PRs 2 and 3 both touch `duration.go`. The second to merge rebases. Dread Wight lands in a small follow-up after PR 2 reaches `main` (Shared machinery 2). Cards that wait on S51 land after the S51 PR they need: Thran Weaponry (ADR 0108 PR 4), Hidden Retreat and Gideon of the Trials (ADR 0108 PR 6).

## Consequences

- A resolved effect can set a land's basic land type, or strip its land types, for any duration, and the static form keeps a land's other subtypes as CR 205.1a says.
- A duration can depend on a counter, on the pinned object being tapped, on a power comparison, and on several conditions at once. Every stored duration is now refused on restore if the binary can't read it.
- Land plays have a gate, as casting does. Emblems are read by every gate whose `Spec` slot they carry. Cards in a zone can be protected from targeting.
- An activated ability's payment reaches its effect and its targets as a spell's does, and survives a copy. The world rule is applied.
- An entering permanent is judged as it would exist on the battlefield. Riot, unleash and sunburst are keywords the engine honours, granted or printed.
- About 215 more cards need nothing else: about 145 from the seams, and about 70 from owner decisions 2, 3 and 4 (5, about 15 and about 49).

## Out of scope

- **#1600** (Lion's Eye Diamond, Throne of Eldraine, The Soul Stone, Crystal Skull, Isu Spyglass). It stays with irobinson010, who landed its Chromatic Orrery and Coveted Jewel items in #1927.
- **"Lands can't enter the battlefield"** (Worms of the Earth). It is a prohibition on a zone change, not a land play. It gets its own registry row.
- **"Spells and abilities can't be countered"** for abilities (Spider-Punk). Nothing gates countering an ability. Its row already says so.
- **Assist** (Gang Up), **"unless a player has this creature deal damage to them"** (Barbarian Bully), and **cumulative upkeep paid from the library** (Thought Lash).
- **Bounds read off the board** ("power less than or equal to the number of creatures you control"). They are live predicates that the engine already expresses.
- **The 28 counter-conditioned self statics** ("has trample as long as it has a +1/+1 counter on it"). They are ordinary conditional statics.

---

## Questions for the owner (answered)

These are the questions as asked. The owner's answers follow.

1. **Riot under a granted or removed ability (§10; CR 614.12).** Rhythm of the Wild grants riot to a creature that is not yet on the battlefield. A creature entering under Dress Down has no abilities to use.
   - **(a) Recommended:** build the CR 614.12 look-ahead. The entry window computes what the permanent would be on the battlefield, from its own statics, the statics already in play and the keywords its spell was granted. Riot, unleash and sunburst entry replacements come from that result. Printed riot plus Rhythm asks twice (CR 702.136b), a creature entering under Dress Down is not asked, and Domri's mana-granted riot works. This is what CR 614.12 says.
   - (b) Each grantor carries its own entry replacement, as Uncivil Unrest does today, and printed riot is read off the card. It is smaller. A creature entering under Dress Down still gets its riot, and a creature with printed riot under Rhythm is asked once, not twice.
2. **Granted loyalty abilities (§2; ADR 0093 Decision 10; CR 606.3).** Teferi's Talent's −12 is a loyalty ability an Aura grants. ADR 0093 put granted loyalty abilities out of scope because no waiting card needed them. Now one does, and four other Talents print the same shape.
   - **(a) Recommended:** lift that exclusion. A granted loyalty row is an ordinary granted activated row and shares the planeswalker's one-per-turn count (CR 606.3). The stack item carries its grantor, so an emblem it creates is the grantor's (CR 114.2). Teferi's Talent becomes Full, and Elspeth's, Liliana's, Rowan's and Vivien's Talent land. That is PR 4.
   - (b) Keep the exclusion. Teferi's Talent keeps its caveat, and the other four Talents wait on the row.
3. **"Discard a card at random" as a cost (§7; CR 601.2h, 701.9b).** 17 Commander-legal activated abilities print it (Pyromancy, Stormscale Anarch, Mage il-Vec, Ogre Shaman …), and none is catalogued. It is the same payment block as #1902's two costs.
   - **(a) Recommended:** PR 8 adds `DiscardCost.Random`. The engine picks the card at payment, after every other cost, as CR 601.2h orders, and records it for §8's readers. About 15 cards land.
   - (b) Only #1902's two costs. The random-discard cards get a registry row and wait.
4. **Cards that need nothing new but share a PR's code (§1, §3, §11).** About 60 cards already have every shape they need, and each PR touches their helper anyway: the 12 static land-type Auras and enchantments (`SetsBasicLandType`, which PR 1 fixes), the 33 single "for as long as this remains tapped" cards (the duration PR 2 extends), and the 15 printed sunburst cards (which PR 10 moves onto the keyword).
   - **(a) Recommended:** each PR lands the ones that need nothing else, about 49, verified against their full text like the rest. The changed helper is then tested against every card that uses it.
   - (b) The PRs land only the cards their seams unblock, and the families stay on the backlog.

---

## Owner decisions, 2026-10-02

The owner answered the four questions on 2026-10-02. Every answer was the recommended option (a).

1. **Riot under a granted or removed ability.** The entry window builds the CR 614.12 look-ahead, `entryLookAheadLocked`, which computes the entering permanent as it would exist on the battlefield. Riot, unleash and sunburst entry replacements come from that result: printed riot under Rhythm of the Wild asks twice (CR 702.136b), a creature entering under Dress Down is not asked, and Domri's mana-granted riot works. Uncivil Unrest moves onto the keyword grant and loses its caveat (§10 decisions 1 to 5; Delivery PR 9, with sunburst's reading in PR 10).
2. **Granted loyalty abilities.** ADR 0093 Decision 10's loyalty exclusion is lifted. A granted loyalty row shares the planeswalker's CR 606.3 count, and the stack item carries its grantor, so an emblem it creates is the grantor's. Teferi's Talent becomes Full, and Elspeth's, Liliana's, Rowan's and Vivien's Talent land (§2 decision 4; Delivery PR 4).
3. **"Discard a card at random" as a cost.** `DiscardCost.Random` is added. The engine picks the card at payment, after every other cost (CR 601.2h), and records it on `PaidCost.Discarded` for §8's readers. About 15 of the 17 cards land (§7 decision 3; Delivery PR 8).
4. **Cards that need nothing new but share a PR's code.** Each PR lands the cards of its family that need nothing else, about 49 in all, each verified against its full text: about 11 static land-type Auras and enchantments in Delivery PR 1, about 25 single "for as long as this remains tapped" cards in PR 2, and about 13 printed sunburst cards in PR 10 (§1, §3 and §11 Cards).
