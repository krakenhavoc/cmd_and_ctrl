# Card-coverage plan: DMU (Dominaria United) and DMC

218 cards not yet in the catalog (181 `dmu`, 37 `dmc`; 211 normal layout, 7 Sagas). Checked against `origin/develop` at `abb3b8bb3`. Machine-readable plan: `plan.json`.

## Mechanics

| Mechanic | Exists | Evidence |
|---|---|---|
| Enlist | **no** | grep -rli enlist server/internal/game -> 0 files; in cards/effects only cenns_enlistment.go (a name) and one test. CR 702.154 / 508.1g read in cmdr/cr.txt. |
| Read ahead | yes | grep -rli 'read ahead' cards/effects -> read_ahead_test.go, the_cruelty_of_gix.go, the_elder_dragon_war.go, love_song_of_night_and_day.go; 'ReadAhead' in 10 files of internal/game (CR 702.155, 714.3b). |
| Kicker | yes | grep -rli kicker -> 29 files in internal/game, 135 in cards/effects (emblazoned_golem.go, vodalian_serpent.go, entry_counters.go for 'if it was kicked, it enters with'). |
| Domain | yes | No engine keyword is needed (ability word). Count exists only as two private copies: domainCountYouControl (leyline_binding.go:38, a CostQuery reader) and rfCreatureBBasicLandTypesAmong (reality_fracture_fra_creature_b_helpers.go:134). No shared exported helper: see s9. |
| Stun counters | yes | grep -rli stun -> 130 files in internal/game, 954 in cards/effects. |
| Compleated | yes | cards/effects/compleated.go, vraska_betrayals_sting.go; game/phyrexian_mana.go, cast_provenance.go ('two fewer loyalty' found in compleated.go). |
| Powerstone tokens | yes | grep -rli powerstone -> 30 files in cards/effects, 9 in internal/game. |
| Defiler cycle (optional 2 life for a coloured permanent spell) | **no** | grep 'As an additional cost to cast .* spells, you may pay' and 'reduces only the amount of' -> 0 files. game/granted_optional_cost.go (Silverquill's casualty grant) is the nearest seam; see s2. |
| Affinity for creatures | yes | salt_road_packbeast.go, leonardo_worldly_warrior.go, witherbloom_the_balancer.go. |
| Battle cry / Heroic / forestwalk / ward / 'can be your commander' | yes | battle_cry.go, hero_of_bladehold.go; heroic in 55 effect files; forestwalk in 6; 'can be your commander' in 7. |

## Seams (built first, one strong-model worker each, who also builds the cards)

| id | Seam | Size | Cards |
|---|---|---|---|
| s1 | Add enlist | medium | Argivian Cavalier, Balduvian Berserker, Barkweave Crusher, Benalish Faithbonder, Coalition Skyknight, Coalition Warbrute, Guardian of New Benalia, Hexbane Tortoise, Keldon Flamesage, Linebreaker Baloth, Yavimaya Steelcrusher |
| s2 | Let a permanent offer '2 life' on coloured permanent spells, with the linked coloured discount | small | Defiler of Dreams, Defiler of Faith, Defiler of Flesh, Defiler of Instinct, Defiler of Vigor |
| s3 | Add 'players can't pay life to cast spells or activate non-mana abilities' | small | Karn's Sylex |
| s4 | Add 'mana abilities of this land cost an additional 1 life' | small | Thran Portal |
| s5 | Report the life paid on an ability's activation event | small | Verrak, Warped Sengir |
| s6 | Remember that an opponent's creature or planeswalker was dealt excess damage this turn | small | Rith, Liberated Primeval |
| s7 | Add 'it can't phase in for as long as you control this' | small | The Phasing of Zhalfir |
| s8 | Add a this-turn form of 'activate loyalty abilities twice' | small | Urza Assembles the Titans |
| s9 | Promote one shared domain count (helper, not an engine gap) | small | Jodah's Codex, Llanowar Greenwidow, Radha's Firebrand, Yavimaya Sojourner, Sunbathing Rootwalla, Territorial Maro, Nishoba Brawler |

s3, s4 and s5 all touch how life is paid as a cost (`game/activated.go`, cost payment); run them one after another, or give them to one worker, to avoid conflicts. s2 should be tested against s3 once both land. s9 is a helper promotion, not an engine gap; it is a seam item only so the three domain batches do not each write the count.

## Batches

| n | Theme | Cards | After |
|---|---|---|---|
| 1 | Vanilla, French vanilla and one-line commons | 13 | - |
| 2 | Simple removal, tricks and stun | 13 | - |
| 3 | Simple creatures: one trigger or one activation | 13 | - |
| 4 | Cost reductions, graveyard abilities and 'this turn' conditions | 11 | - |
| 5 | Instants-and-sorceries matter | 9 | - |
| 6 | Kicker creatures, simple | 11 | - |
| 7 | Kicker with a delayed or replaced outcome | 10 | - |
| 8 | Kicker spells, part 1 | 9 | - |
| 9 | Kicker spells, part 2 | 8 | - |
| 10 | Domain: spells and enter triggers | 8 | s9 |
| 11 | Domain triggers, lands and green rares | 10 | s9 |
| 12 | Go-wide rares: lords, tokens and recursion | 10 | - |
| 13 | Rares with a choice or a continuous effect | 9 | - |
| 14 | Read-ahead Sagas and enchantments | 9 | s9 |
| 15 | Multicoloured and 'all colors' | 8 | - |
| 16 | Planeswalkers and artifact engines | 8 | - |
| 17 | Copy effects | 10 | - |
| 18 | Commander legends: token makers and tribal lords | 10 | - |
| 19 | Commander legends: combat and conditions | 10 | - |

Total: 29 cards in seams, 189 in 19 batches.

## Watch items (a precedent exists but the fit is not proven: stop and report if it does not fit)

- **Ramses, Assassin Lord** (19): the lose-the-game trigger exists (`zenos_yae_galvus.go`); 'attacked this turn by an Assassin you controlled' needs a per-defender attack record that `game/turn_tally.go` may not keep. The most likely late seam.
- **Karn, Living Legacy** (16): 'Pay any amount of mana' at resolution, and an emblem with an activated ability.
- **Stangg, Echo Warrior** (17): token copies of Auras and Equipment entering attached to a named token.
- **Unite the Coalition** (15): five picks of repeatable targeted modes.
- **Nemata, Primeval Warden** (12): reflexive 'When you do' off a dies-replacement.
- **Academy Loremaster** (13): a turn-long cost increase for one player. **Zur, Eternal Schemer** (13): animation with base P/T equal to mana value.
- **Monstrous War-Leech** (7): 'As this creature enters, if it was kicked, mill four cards.' **Volshe Tideturner** (6): mana spendable on a kicked spell.
- **Rona, Sheoldred's Faithful** (5): graveyard cast with an additional cost. **Ertai's Scorn** (2): a conditional {U}-only reduction.
- **Sivitri, Dragon Master** (16): an attack tax in life as a duration effect. **The Peregrine Dynamo** (17): the 'legendary source that's not a commander' filter. **Jasmine Boreal** (19): 'no abilities' on both a mana restriction and a block rule. **Threats Undetected** (11): pairwise different powers.

## New token rows (each in one batch or seam only)

Kavu 3/3 trample, all colors (15); Badger 3/3, Cat Warrior 2/2 forestwalk, Knight 2/2 protection from red, Wurm 4/4 (18); Ornithopter 0/2 (14); Stangg Twin (17); Phyrexian 2/2 black (s7); Bird 1/1 black flying 'can't block' if no row fits (12). The 4/4 red flying Dragon is used by Rohgahh (18) and Rith (s6) and the 4/4 Angel by Historian's Boon (14) and Jenson (15): both look to be existing rows; reuse, do not add.

## Cards not to build

None. The set has no acorn or joke text and nothing here needs an owner decision, so there is no 'owner decision' batch.
