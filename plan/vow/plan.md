# Card-coverage plan: VOW (Innistrad: Crimson Vow) and VOC

206 cards not yet in the catalog (182 `vow`, 24 `voc`). Checked against `origin/develop`. 3 seams (14 cards), 18 batches (192 cards). No card is held back for an owner decision: the input has no acorn or joke text.

## Mechanics

| Mechanic | Exists | Evidence |
|---|---|---|
| Blood tokens | yes | grep 'Blood' cards/effects/tokens.go: BloodToken() / printedBloodToken (CR 111.10g); 'Blood token' in 8 non-test effect files (glass_cast_heart.go, transmutation_font.go, kamber_the_plunderer.go ...). 'Whenever you sacrifice a <Food/Clue/Treasure>' has 13 precedent files (rapacious_guest.go, lonis_genetics_expert.go); SacrificeN / SacrificeSubtype in activated.go. |
| Cleave | yes | grep 'Cleave': effects.Cleave(cost, targets) in cards/effects/alternative_cost.go, AlternativeCost.Targets in game/alternative_cost.go (CR 702.148); catalog cards wash_away.go, dread_fugue.go, inspired_idea.go, alchemists_gambit.go. A {X} alternative cost is covered by game/cast_cost_test.go TestAlternativeCostAndTheValueOfX. |
| Training | yes | grep 'Training(': cards/effects/training.go Training(label) (CR 702.149a), used by hopeful_initiate.go and warriors_resolve.go. 'trains' (CR 702.149c) has no hook: grep 'this creature trains' and 'Trained|EventTrain' in game and effects found nothing; one card (Savior of Ollenbock), watch item in batch 9. |
| Exploit | yes | grep 'Exploit': cards/effects/exploit.go Exploit(cardName, payoff BodyRef) (CR 702.110), used by overcharged_amalgam.go. No engine event: grep 'Exploit' in internal/game found 0 files, so another permanent cannot watch an exploit (seam s2). |
| Transforming double-faced cards | yes | grep 'Transform' in internal/game: LayoutTransform, TransformPermanentForEffect, ReturnFromGraveyardTransformedForEffect, ExileAndReturnTransformedForEffect (game/transform.go); 232 effect files; 'transforms into' triggers in afflicted_deserter.go, ashling_rekindled.go. |
| Daybound / nightbound | yes | grep 'Daybound': KeywordDaybound in internal/game, docs/adding-cards.md 'Day and night' (ADR 0132). No card of the input uses it. |
| Disturb | yes | grep 'Disturb': 35 effect files, game tests TestDisturbCastsTheBackFaceFromTheGraveyard. No card of the input uses it. |
| Soulbond (VOC) | **no** | grep -rni 'soulbond' over server/internal, docs and client/src: 0 hits. grep 'PairedWith|paired with another': 0 hits (the 11 'paired' hits in effects are unrelated prose). Seam s1. |
| Melee (VOC) | yes | grep 'Melee': KeywordMelee, resolveMeleeItems in internal/game; 'have melee|has melee' in glory_to_the_empire.go, titania_proud_pummeler.go. |
| Affinity for Spirits (VOC) | yes | grep 'AffinityFor': cards/effects/self_cost_modifier.go AffinityFor(label, match), 40 files. |
| Partner with (VOC) | yes | grep 'PartnerWith': 21 effect files; docs/adding-cards.md 'Partner with (CR 702.124j, #2142)'. |

## Seams

| id | Seam | Size | Cards |
|---|---|---|---|
| s1 | Add soulbond: pairing two creatures | large | Breathkeeper Seraph, Doom Weaver, Imperious Mindbreaker, Mirage Phalanx, Thundering Mightmare |
| s2 | Make exploiting an event other permanents can watch | small | Skull Skaab, Diver Skaab, Fell Stinger, Graf Reaver, Mindleech Ghoul, Repository Skaab, Rot-Tide Gargantua, Stitched Assistant |
| s3 | Return a card transformed, attached to a named host | small | Radiant Grace |

Each seam is built by one strong-model worker, who also builds its cards. No batch waits on a seam.

## Batches

| n | Theme | Cards | After |
|---|---|---|---|
| 1 | Simple creatures, tricks and removal | 13 | - |
| 2 | Simple creatures with one trigger or condition | 13 | - |
| 3 | Blood tokens: plain makers | 13 | - |
| 4 | Blood tokens: sacrificing and counting them | 11 | - |
| 5 | Blood and Vampires: counted amounts and rules-heavy cards | 9 | - |
| 6 | Auras, Equipment, artifacts and life gain | 13 | - |
| 7 | Creature-type words: Wolves, Vampires, Zombies, Spirits | 13 | - |
| 8 | Cleave | 8 | - |
| 9 | Training and +1/+1 counters | 11 | - |
| 10 | Graveyard value: Zombies, self-mill and recursion | 13 | - |
| 11 | Green: Insects, lands and toughness | 12 | - |
| 12 | Spirits: tokens, melee and partners | 10 | - |
| 13 | Tempo, tappers and mass effects | 11 | - |
| 14 | Double-faced cards: one transform condition | 8 | - |
| 15 | Double-faced cards: rules-heavy | 8 | - |
| 16 | Cards exiled with a permanent, and cards in exile | 10 | - |
| 17 | Copies and reanimation | 8 | - |
| 18 | Planeswalker, play-from-exile and big rares | 8 | - |

## Watch items (a precedent exists; stop and report if it does not fit)

- Batch 4, Anje, Maid of Dishonor: the cost "Sacrifice another creature or a Blood token" as one either-or sacrifice.
- Batch 5, Midnight Arsonist: "up to X target artifacts without mana abilities"; no card has that filter, `game.ManaAbilitiesOf` can serve.
- Batch 6, Boarded Window: "dealt 4 or more damage this turn" needs combat plus noncombat damage to a player.
- Batch 8, Lantern Flare: a cleave cost with {X}; tested in `game`, used by no catalog card.
- Batch 9, Savior of Ollenbock: "Whenever this creature trains" (CR 702.149c), built like `exploit.go`'s reflexive trigger.
- Batch 15, Jacob Hauken, Inspector: face-down exile from hand, and a free play once each turn across both faces.
- Batch 16, Circle of Confinement: "the same name as a card exiled with this enchantment".
- Batch 17, Donal, Herald of Wings (a creature-spell copy with an exception) and Olivia, Crimson Bride (an indefinitely granted state trigger).

## Cards not to build

None. Every card is in a seam or a batch.
