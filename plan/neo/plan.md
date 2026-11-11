# Card-coverage plan: NEO (Kamigawa: Neon Dynasty) and NEC

225 cards not yet in the catalog (201 `neo`, 24 `nec`), checked against `origin/develop`. 4 seams (5 cards), 20 batches (220 cards). Details, greps and oracle ids are in `plan.json`.

## Mechanics

| Mechanic | Exists | Evidence |
|---|---|---|
| Ninjutsu | yes | git grep -il ninjutsu origin/develop: 33 files in effects, 30 in internal/game; effects/ninjutsu.go has Ninjutsu(cost) (ninja_of_the_deep_hours.go uses it). Only the cost-modifier keyword bit is missing (seam s2). |
| Channel | yes | effects/channel.go (channelOnResolve) and boseiju_who_endures.go; game/cost_modifier.go:814 prices a channel card in hand. 13 new cards reuse it. |
| Reconfigure | yes | 26 effects files (acquisition_octopus.go, blade_of_the_oni.go ...); docs/adding-cards.md 'Reconfigure (CR 702.151, #2639)'. No reconfigure card is left in the input. |
| Modified (CR 700.9) | yes | '\bmodified\b' in 56 effects files (arana_heart_of_the_spider.go, attachments.go, batch07_helpers.go); 27 game files match Modified. |
| Saga that exiles itself and returns transformed | yes | effects/transform.go:30,101 ExileAndReturnTransformed; fable_of_the_mirror_breaker.go is on develop. |
| Samurai/Warrior 'attacks alone' | yes | 'attacks alone' in 32 effects files (agent_13_sharon_carter.go, agents_of_shield.go: 'Whenever a creature you control attacks alone'); triggers_common.go:1356 ThisAttackedAlone. No 'Samurai or Warrior' card yet, so the subtype filter is new card-side code for one batch. |
| Crew / Vehicles | yes | 47 game files; CrewCost/CrewEffect (getaway_car.go), CrewedThis target wrapper, EventCrewedVehicle (events.go:1126). 'Whenever this Vehicle becomes crewed' (CR 702.122e) is NOT there (seam s1). |
| Pilot token 'crews Vehicles as though its power were 2 greater' | yes | 17 effects files (back_on_track.go, born_to_drive.go); Characteristic CountedPower in game/activated.go:2917. |
| Compleated | yes | effects/compleated.go, vraska_betrayals_sting.go; docs 'Compleated (CR 702.150a, ADR 0183, #2946)'. |
| Shrines (count of Shrines you control) | yes | Sanctum of Fruitful Harvest is cast in avatar_b3_cards_test.go; subtype counts are ordinary. Go-Shintai cards themselves are new. |
| Keyword counters (indestructible, menace, deathtouch, reach ...) | yes | 52 effects files (aragorn_company_leader.go, arwen_mortal_queen.go); 17 game files (card.go, characteristic.go). |
| Affinity (artifacts, enchantments) | yes | 'affinity for' in 45 effects files; game/cost_modifier.go:39. |
| Marches: exile any number of <colour> cards from hand, {2} less each | yes | effects/additional_cost.go:372 ExileAnyNumberFromHandCost + CostsLessPerExiledFromHand (March of Swirling Mist, #3175); hand_exile_x_costs_test.go. |

No headline mechanic of the set is missing: later sets and Commander decks already brought in ninjutsu, channel, reconfigure, modified, transforming Sagas, crew, compleated and keyword counters. The seams are four narrow gaps.

## Seams (built first, one strong-model worker each, who also builds the cards)

| id | Seam | Size | Cards |
|---|---|---|---|
| s1 | Add 'whenever this Vehicle becomes crewed' (CR 702.122e) | medium | Mobilizer Mech, Mindlink Mech |
| s2 | Let cost modifiers recognise a ninjutsu ability | small | Silver-Fur Master |
| s3 | Add a 'can't crew Vehicles' restriction | small | Intercessor's Arrest |
| s4 | Record damage one source dealt to one player this turn | small | Hidetsugu Consumes All // Vessel of the All-Consuming |

## Batches (one worker each, in order)

| n | Theme | Cards | After |
|---|---|---|---|
| 1 | Vanilla, keyword and one-line creatures and Vehicles | 13 | - |
| 2 | Simple instants and sorceries | 13 | - |
| 3 | Simple enters / dies / activated creatures | 13 | - |
| 4 | Artifact synergies, small creatures and Equipment | 12 | - |
| 5 | Vehicles with triggers and artifact engines | 13 | - |
| 6 | Ninjutsu creatures | 12 | - |
| 7 | Channel cards | 13 | - |
| 8 | Artifact-and-enchantment payoffs, Samurai tokens | 12 | - |
| 9 | Samurai or Warrior attacks alone | 12 | - |
| 10 | Modified matters I | 13 | - |
| 11 | Modified matters II and counter movers | 11 | - |
| 12 | Auras | 12 | - |
| 13 | Vehicle type-changing, animation and artifact recursion | 8 | - |
| 14 | Marches, Invokes and X costs | 9 | - |
| 15 | Trickier instants and sorceries | 11 | - |
| 16 | Shrines and Myojin | 11 | - |
| 17 | Sagas I: simple chapters, keyword back faces | 8 | - |
| 18 | Sagas II: board effects and counting back faces | 7 | - |
| 19 | Sagas III: rules-heavy chapters, and Satsuki | 7 | - |
| 20 | Planeswalkers and complex legends | 10 | - |

No batch depends on a seam. Batches 13 to 20 are the rules-heavy ones (7 to 11 cards).

## Watch items (a precedent exists but was not proven to fit: stop and report)

- Batch 4: Futurist Operative; Bronze Cudgels; Eater of Virtue; Kami of Transience
- Batch 5: Oni-Cult Anvil (artifacts leaving as a batch, once each turn)
- Batch 7: Mirrorshell Crab (counter an ability unless its controller pays); Roaring Earth ({X} channel)
- Batch 11: Dramatist's Puppet; Ascendant Acolyte; Kosei, Penitent Warlord
- Batch 12: Crackling Emergence and Harmonious Emergence (destroy replacement that sacrifices the Aura)
- Batch 13: Mechtitan Core (token-leaves delayed return); Greasefang, Okiba Boss
- Batch 14: March of Reckless Joy (play up to two); March of Burgeoning Life; Invoke Calamity; Scrap Welder; Tameshi, Reality Architect
- Batch 15: Storyweave (next entering batch gets counters); Collision of Realms; Smoke Spirits' Aid; Discover the Impossible; Invoke Despair
- Batch 16: Myojin of Grim Betrayal
- Batch 19: Etching of Kumano; Living Breakthrough (cast ban by mana value); The Dragon-Kami Reborn; O-Kagachi Made Manifest
- Batch 20: Hinata, Dawn-Crowned (cost per target on every spell); Thousand-Faced Shadow; Hidetsugu, Devouring Chaos; Risona

The most likely of these to turn into a small seam: Storyweave, Hinata, Tameshi, Invoke Calamity and the two Emergence Auras.

## Cards that should not be built

None. The input has no acorn, joke or owner-decision card, so there is no "owner decision" batch.

## Notes

- Tokens are `TokenCard("...")` strings on develop, so the 1/1 colorless Spirit made in batches 4, 10, 11, 16, 18 and 19 and the 2/2 red menace Spirit in batches 11 and 12 need no shared row. Every Samurai token is in batch 8; the named tokens (Tamiyo's Notebook, Keimi, Mechtitan, Smoke Blessing, the Shrine token) are each in one batch.
- Sentences split across batches are named in both `shared` notes: "As long as enchanted permanent is a creature" (7 and 12), the modified predicate (10, 11, 19), the look-and-pick sentence (3, 8 and 15).
- Rule numbers checked in `cmdr/cr.txt`: 702.122e (becomes crewed), 700.9 (modified).
