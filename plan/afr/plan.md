# Card-coverage plan: AFR (Adventures in the Forgotten Realms) and AFC

227 cards not yet in the catalog (afr and its Commander companion afc), checked against `origin/develop`. 3 seams (14 cards) and 20 batches (213 cards). Every oracle id appears exactly once in `plan.json` (checked by script).

## Mechanics

| Mechanic | Exists | Evidence |
|---|---|---|
| Venture into the dungeon / dungeons | yes | No venture card is in the input (all already in the catalog). docs/adding-cards.md has 'The initiative and dungeons (ADR 0203, #4096)'; effects/venture_cards_b_test.go exists. |
| Rolling dice (d20 results tables, d4-d12) | yes | effects/random_effects.go: RollDice{Sides,N,IgnoreLowest,Then}, DiceTail, WheneverYouRollDice; game/dice_roll.go RollDiceThenForEffect (#4307). Precedents: hoarding_ogre.go (d20 table), diviners_portent.go ('roll a d20 and add'), reckless_endeavor.go ('roll two d8 and choose one result'), berserkers_frenzy.go. |
| Dice: 'roll that many dice plus one and ignore the lowest roll' (static replacement) | **no** | Only a one-shot scoped record exists: game/dice_roll.go ModRollExtraDieIgnoreOne (Squid Fire Knight, a chooser picks the roll to ignore). grep 'RepEventRollDice' shows no card-side static replacement; normalized grep of effects for 'plus one and ignore' and 'ignore the lowest roll' = 0 files. Seam s1. |
| Dice: natural result vs result after modifiers | **no** | CR 706.2 separates the natural result from the result. game/dice_roll.go:237 emits EventRollDie with Amount = the die face only; grep 'Natural/Modifier' in game/events*.go and effect_api_random.go = nothing; diviners_portent.go adds its modifier inside its own tail, after the event. Natural-result readers can use Amount/Sides today; a reader of the modified result (Farideh) cannot. Seam s1. |
| Dice rolled as a cost or as a permanent enters | **no** | Cost constructors in effects (grep '^func [A-Z]\w*Cost(') list no roll cost; game corpus grep 'roll a d[0-9]+ as a cost/RollAsCost/CostRoll' = 0; no effects file combines RollDice with EntersWith/AsEnters. Seam s2. |
| Class enchantments (levels, 'When this Class becomes level N') | yes | effects/artificer_class.go and the Bloomburrow Talents (alchemists_talent.go, builders_talent.go ...): 8 files quote 'Class becomes level'; docs/adding-cards.md 'Designations: Class levels' (#757). |
| Pack tactics (ability word) | yes | effects/werewolf_pack_leader.go and intrepid_outlander.go use packTacticsAttackingPower / packTacticsTotalPower. |
| 'Mana from a Treasure was spent' | yes | game/mana_spent.go ManaSpent view (FromTreasure), cast_provenance.go; effects/hired_hexblade.go (ETB check), kalain_reclusive_painter.go ('for each mana from a Treasure'). Not proven for an ACTIVATED ability's payment (Forsworn Paladin): watch item in batch 9. |
| Flavor words (Bardic Inspiration, Cone of Cold ...) | yes | No rules meaning (ability-word-like labels); nothing to build. |
| Ward (mana and non-mana costs) | yes | effects/ward.go; 14 files quote 'Ward—Discard a card'. |
| Monstrosity (Clay Golem) | yes | effects/monstrosity.go; 13 files quote 'becomes monstrous'. Only the 'Roll a d8' cost is missing (s2). |
| Melee (Wulfgar) | yes | game/melee.go; trigger doubler for 'a creature attacking causes a triggered ability ... to trigger' quoted in 23 effects files. |
| Phasing ('phases out') | yes | docs 'Phasing (#1199, CR 702.26)'; 19 effects files quote 'phases out' (galadriels_dismissal.go, guardian_of_faith.go). |
| Cascade granted to a spell (Wild-Magic Sorcerer) | yes | effects/maelstrom_nexus.go, imoti_celebrant_of_bounty.go ('has cascade'). 'First spell you cast from exile each turn' filter unproven: watch item in batch 16. |

## Seams (one strong-model worker each, before the batches that name them)

| id | Seam | Size | Cards |
|---|---|---|---|
| s1 | Add a static 'roll one more die and ignore the lowest' replacement and carry roll modifiers on the roll | medium | Pixie Guide, Barbarian Class, Farideh, Devil's Chosen, Netherese Puzzle-Ward, Critical Hit, The Deck of Many Things, Danse Macabre, Revivify, Song of Inspiration |
| s2 | Roll dice as an activation cost and as a permanent enters | medium | Clay Golem, Neverwinter Hydra |
| s3 | Give a token only the triggered abilities of the cards exiled to pay for it (Vecna) | small | The Book of Vile Darkness, Eye of Vecna, Hand of Vecna |

s2 is built after s1: both edit `server/internal/game/dice_roll.go`.

## Batches

| n | Theme | Cards | After |
|---|---|---|---|
| 1 | Vanilla, keyword and one-line ETB creatures | 13 | — |
| 2 | Simple creatures with one activated or triggered ability | 13 | — |
| 3 | One-shot instants and sorceries | 13 | — |
| 4 | Modal 'You ...' spells and other 'Choose one' instants | 12 | — |
| 5 | Modal ETB creatures and simple artifacts | 12 | — |
| 6 | Simple Equipment and 'is equipped' creatures | 13 | — |
| 7 | Life gain and 'if a creature died this turn' | 12 | — |
| 8 | Heavier Equipment and Equipment legends | 10 | — |
| 9 | Treasure-mana payoffs and pack tactics | 11 | — |
| 10 | d20 results tables, simple | 13 | — |
| 11 | Dice tables with heavier rows, and other dice sizes | 11 | — |
| 12 | 'Whenever you roll one or more dice', the Endeavors and multi-player rolls | 9 | s1 |
| 13 | Classes I: levels built from existing statics and triggers | 8 | — |
| 14 | Classes II and playing cards from exile | 8 | — |
| 15 | Planeswalkers and token-making legends | 8 | — |
| 16 | Copies, and casting cards that are not in your hand | 9 | — |
| 17 | Creature lands, phasing and tap-down | 10 | — |
| 18 | Dragons, ward and big creatures | 11 | — |
| 19 | X spells, target-set rules and one-off artifacts | 9 | — |
| 20 | Complex legends | 8 | — |

Batches 12 to 16, 19 and 20 are rules-heavy (8 or 9 cards). Card names per batch are in each batch's `cards` (oracle ids) and its `shared` note in `plan.json`.

## Watch items (a precedent exists but was not proven to fit: stop and report)

- Batch 4: You Find Some Prisoners (impulse from an opponent's library).
- Batch 6: Trickster's Talisman (granted ability sacrifices its grantor).
- Batch 7: Belt of Giant Strength (equip reduction reads the target's power); Dancing Sword (permanently stops being an Equipment); Galea (ability given to a spell cast from the library top); Vorpal Sword (Equipment gains a trigger until end of turn).
- Batch 8: The Book of Exalted Deeds (granted "can't lose" with no duration).
- Batch 9: Forsworn Paladin (Treasure mana spent on an activated ability); Mimic (nontoken Treasure must count).
- Batch 11: Delina ("You may roll again", CR 706.3c); Ebony Fly; Sword of Hours; Bag of Tricks.
- Batch 12: Chaos Dragon (each player rolls, then a per-player attack ban on a must-attack creature); Grave Endeavor.
- Batch 13: Fighter Class level 3 (a block requirement naming the triggering attacker).
- Batch 14: Bard Class level 2 (coloured-only reduction); Sorcerer Class level 2 (mana usable to gain a Class level); Monk Class level 3 (granted conditional cast-from-exile); Asmodeus (draw replaced by face-down exile); Share the Spoils ("whenever an opponent loses the game", one shared pile, per-player once-a-turn permission).
- Batch 15: Mordenkainen −10 ("Exchange your hand and library": no precedent found); Lolth's emblem; Grand Master of Flowers +1.
- Batch 16: Demilich (graveyard cast by exiling four instants/sorceries); Wild-Magic Sorcerer (first spell cast from exile has cascade); Extract Brain (cast from an opponent's hand); Xanathar (play the top of an opponent's library, cast ban).
- Batch 18: Green Dragon (repeating delayed trigger until end of turn); Inferno of the Star Mounts ("When its power becomes 20 this way").
- Batch 19: Ride the Avalanche ("When you next cast a spell this turn"); Hellish Rebuke (granted ability naming the caster).
- Batch 20: Orcus; Lorcan; Nihiloor; Froghemoth (target count from the damage event).

## Cards that should not be built

None. The set has no acorn or joke cards, and no card needs an owner decision, so there is no "owner decision" batch. Flavor words (Bardic Inspiration, Cone of Cold and so on) have no rules meaning and need nothing.

## Notes

- No venture or dungeon card is in the input: they are already in the catalog.
- Shared tokens are kept in one batch: Skeleton (batch 8), Spider (batch 15). Goblin, Wolf, Zombie, Treasure, Knight and Beast tokens are expected to exist; check the token rows before adding one.
- Batch 12 waits for s1 because `WheneverYouRollDice` changes there. Batches 10 and 11 use only the existing `RollDice` / `DiceTail` and do not wait.
