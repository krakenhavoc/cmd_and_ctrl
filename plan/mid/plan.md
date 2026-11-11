# Card-coverage plan: MID (Innistrad: Midnight Hunt) and MIC

184 cards not yet in the catalog (155 `mid`, 29 `mic`), checked against `origin/develop` at abb3b8bb3. 4 seams (4 cards), 17 batches (180 cards). Machine-readable: `plan.json`.

## Mechanics

| Mechanic | Exists | Evidence |
|---|---|---|
| Flashback | yes | effects/alternative_cost.go:111 Flashback(cost), :175 FlashbackSacrifice, :469 FlashbackTap; grep -rli flashback effects: 261 files. CR 702.34a. |
| Flashback with a counter-removal cost that sets X (Light Up the Night) | **no** | game/alternative_cost.go:342 AlternativeCost.RemoveCounters exists (#5390, Quilled Greatwurm) but its comment says effects.Register holds it to 'a fixed count'; grep -rniE 'loyalty counters from among\|Remove X loyalty' in effects and game: nothing. Seam s2. |
| Coven (ability word) | yes | effects/trek_b48_helpers.go:17 trekB48Coven(g, controller, source) reads three-or-more different powers through the layers; used by augur_of_autumn.go, redemption_choir.go; collectors_cage.go; rfCreatureCDifferentPowers (reality_fracture_fra_creature_c_helpers.go:242) counts different powers. |
| Decayed | yes | game/keywords.go:392 (CR 702.147a, #2650), game/layers.go:1296 foldDecayedLocked, game/keyword_counters.go CounterDecayed; token via TokenCard("2/2 black Zombie with decayed") in crowded_crypt.go:47; 15 effects files; granted decayed in ghouls_night_out.go. |
| Disturb | yes | effects/alternative_cost.go:139 Disturb(cost) (ADR 0107 §4, CR 702.146a); 66 effects files. No disturb card is in this input; only cards that REFER to disturb (Thraben Exorcism, Phantom Carriage, Shipwreck Sifters, Unblinking Observer). |
| Mana restricted to 'pay a disturb cost' (Unblinking Observer) | **no** | game/mana_restriction.go tag vocabulary is closed (lines 81-200): purpose:cast, cast:from-graveyard, cast:from-exile, type:, subtype:, ...; grep -nE 'alt:\|AltKey\|Alternative\|disturb' game/mana_restriction.go: nothing. Seam s4. |
| Daybound / nightbound, day and night | yes | docs/adding-cards.md 'Day and night, daybound and nightbound (ADR 0132, #2561, CR 731 / 702.145)'; 42 effects files. No daybound card is left in this input. |
| Transforming double-faced cards | yes | game/transform.go (ExileAndReturnTransformedForEffect :400, ReturnFromGraveyardTransformedForEffect :565); ReturnCard.Transformed (effects/primitives.go:1025, #1900); 'As this transforms into' in curse_of_leeches.go, sephiroth_one_winged_angel.go, zenos_yae_galvus.go. |
| Return transformed AND attached to a target (Vengeful Strangler) | **no** | ReturnFromGraveyardTransformedForEffect(cardID, controller, tapped, counters) takes no host; EntersAttachedTo exists only on SearchLibrary (primitives.go:1204, #3158); grep -rniE 'transformed under your control attached\|transformed.*attached\|attached.*transformed' effects game: nothing. Seam s3. |
| Curses (Aura Curse, enchant player) | yes | 12 effects files with 'Enchant player' (curse_of_fools_wisdom.go, curse_of_leeches.go, attachments.go); game.AttachForEffect(attachmentID, host TargetRef) game/attach.go:113; QueueAttachSourceToPlayerForEffect game/attach_transform.go:57. |
| 'If an opponent lost life this turn' (Vampires) | yes | 12 effects files: bloodtithe_collector.go, cindering_cutthroat.go:44 oneIfAnOpponentLostLifeThisTurn, cast_restriction.go, avatar_b34_helpers.go. |
| Adversary cycle: 'you may pay {cost} any number of times. When you pay this cost one or more times' | yes | bloodthirsty_adversary.go: MayPayX{Symbol,Goal,Then} + PayUnlessBranches + ReflexiveTrigger with Params.Amount (CR 603.12). |
| Investigate / Clue | yes | effects/tokens.go:192 ClueToken; docs/adding-cards.md 'Investigate (CR 701.16, #5425)'; 127 effects files. |
| 'Tap three untapped creatures you control' as an activation cost | yes | effects/activated.go:585-615 TapAnotherUntapped / TapAnUntapped / TapXUntapped over game.TapOthersCost; 10 effects files match 'tap (two\|three\|four\|five) untapped creatures you control' (harmonized_trio.go, aziza_mage_tower_captain.go). |
| Life total 'would be reduced to 0 or less' replacement (Enduring Angel) | **no** | grep -rniE 'reduced to 0 or less\|life total would\|LifeFloor' effects game: only game/life_floor.go (Angel's Grace: DAMAGE can't reduce life below 1, PlayerStatic.LifeFloorOne). Nothing replaces a life-total reduction from life loss or a payment, and nothing reads 'if it didn't transform this way'. Seam s1. |
| Phasing (Creeping Inn, Spectral Adversary) | yes | docs/adding-cards.md 'Phasing (#1199, CR 702.26)'; 46 effects files. |
| Mana spent to cast (Smoldering Egg) | yes | game/paid_cost.go ManaSpentCount; 15 effects files with 'amount of mana spent to cast' (aberrant_manawurm.go). |

The set's headline mechanics (daybound/nightbound, disturb, decayed, coven, flashback) are all on develop: later sets and the Commander waves brought them in, and no daybound or disturb card is left in the input. What is missing is four single-card pieces.

## Seams (built first, one strong-model worker each, who also builds the seam's card)

| id | Seam | Size | Cards |
|---|---|---|---|
| s1 | Replace a life total being reduced to 0 or less | medium | Enduring Angel |
| s2 | Add a flashback cost that removes X counters and sets the spell's X | small | Light Up the Night |
| s3 | Return a card to the battlefield transformed and attached to a target | small | Vengeful Strangler |
| s4 | Add a mana restriction tag for paying a disturb cost | small | Unblinking Observer |

No batch depends on a seam.

## Batches

| n | Theme | Cards | After |
|---|---|---|---|
| 1 | Plain creatures: enters, dies and pump activations | 13 | — |
| 2 | Plain instants and removal | 13 | — |
| 3 | Flashback I: one plain effect with flashback | 13 | — |
| 4 | Decayed Zombie tokens and Zombie tribal | 13 | — |
| 5 | Coven I: begin-combat grants and simple coven spells | 13 | — |
| 6 | Vampires: 'if an opponent lost life this turn' | 12 | — |
| 7 | Graveyard: exile a card, mill, and cards that name flashback or disturb | 12 | — |
| 8 | Auras, Equipment and plain artifacts and enchantments | 11 | — |
| 9 | Tap-three costs, spell triggers and this-turn counts | 13 | — |
| 10 | Flashback II: flashback spells with a choice, a copy or a lasting effect | 8 | — |
| 11 | Visions cycle and casting from your graveyard | 10 | — |
| 12 | Transforming cards I: a counter or a cost flips it | 8 | — |
| 13 | Coven II and 'creatures with different powers' choices | 8 | — |
| 14 | Curses and Lynde | 8 | — |
| 15 | Adversaries and single rares | 9 | — |
| 16 | Commander-deck Humans, Zombies and legends | 10 | — |
| 17 | Transforming cards II, copies and spell damage | 6 | — |

### Cards by batch

- **1** Dawnhart Rejuvenator, Gavony Silversmith, Snarling Wolf, Gavony Trapper, Bounding Wolf, Clarion Cathars, Novice Occultist, Loyal Gryff, Brood Weaver, Pestilent Wolf, Lambholt Harrier, Nebelgast Intruder, Odric's Outrider
- **2** Blessed Defiance, Flare of Faith, Play with Fire, Pack's Betrayal, Foul Play, Lunar Frenzy, Defend the Celestus, Fading Hope, Stolen Vitality, Defenestrate, Raze the Effigy, Sungold Barrage, Geistwave
- **3** Shadowbeast Sighting, Rise of the Ants, Dryad's Revival, Homestead Courage, Angelfire Ignition, Abandon the Post, Sacred Fire, Path to the Festival, Tapping at the Window, Arcane Infusion, Hungry for More, Winterthorn Blessing, Secrets of the Key
- **4** Diregraf Horde, Startle, Ghoulish Procession, Flip the Switch, Hobbling Zombie, Ghoulcaller's Harvest, Revenge of the Drowned, Rotten Reunion, No Way Out, Bladestitched Skaab, Tomb Tyrant, Crawl from the Cellar, Empty the Laboratory
- **5** Candlegrove Witch, Dawnhart Wardens, Stalwart Pathlighter, Duelcraft Trainer, Harvesttide Sentry, Candlelit Cavalry, Ritual Guardian, Might of the Old Ways, Ritual of Hope, Dawnhart Mentor, Candletrap, Contortionist Troupe, Duel for Dominance
- **6** Bat Whisperer, Arrogant Outlaw, Falkenrath Pit Fighter, Voldaren Ambusher, Mounted Dreadknight, Vampire Socialite, Famished Foragers, Falkenrath Perforator, Voldaren Stinger, Neonate's Rush, Slaughter Specialist, Bloodline Culling
- **7** Jack-o'-Lantern, Crossroads Candleguide, Soul-Guide Gryff, Dreadhound, Drownyard Amalgam, Organ Hoarder, Phantom Carriage, Shipwreck Sifters, Thraben Exorcism, Willow Geist, Locked in the Cemetery, Ardent Elementalist
- **8** Cathar's Call, Stuffed Bear, Bramble Armor, Howl of the Hunt, Hedgewitch's Mask, Silver Bolt, Necrosynthesis, Grafted Identity, Borrowed Time, Geistflame Reservoir, Avacyn's Memorial
- **9** Skaab Wrangler, Larder Zombie, Siege Zombie, Festival Crasher, Storm Skreelix, Purifying Dragon, Fleshtaker, Search Party Captain, Grizzly Ghoul, Diregraf Rebirth, Sunset Revelry, Burn Down the House, Briarbridge Tracker
- **10** Hallowed Respite, Dire-Strain Rampage, Storm the Festival, Turn the Earth, Can't Stay Away, Wake to Slaughter, Rite of Harmony, Siphon Insight
- **11** Visions of Glory, Visions of Dominance, Visions of Dread, Visions of Ruin, Visions of Duplicity, Patrician Geist, Ominous Roost, Rootcoil Creeper, Seize the Storm, Croaking Counterpart
- **12** Mystic Skull, Bereaved Survivor, Deathbonnet Sprout, Ecstatic Awakener, Mysterious Tome, Heirloom Mirror, Smoldering Egg, Ambitious Farmhand
- **13** Leinore, Autumn Sovereign, Sungold Sentinel, Sigarda, Champion of Light, Wall of Mourning, Sigarda's Vanguard, Sigardian Zealot, Celestial Judgment, Celebrate the Harvest
- **14** Curse of Unbinding, Curse of Conformity, Curse of Shaken Faith, Curse of Clinging Webs, Curse of Obsession, Curse of Silence, Curse of Surveillance, Lynde, Cheerful Tormentor
- **15** Primal Adversary, Intrepid Adversary, Spectral Adversary, Tainted Adversary, Gisa, Glorious Resurrector, Consuming Blob, Moonveil Regent, Slogurk, the Overslime, Sigarda's Splendor
- **16** Kyler, Sigardian Emissary, Heronblade Elite, Katilda, Dawnhart Prime, Somberwald Beastmaster, Eloise, Nephalia Sleuth, Prowling Geistcatcher, Moorland Rescuer, Sigardian Savior, Ravenous Rotbelly, Gorex, the Tombshell
- **17** Jerren, Corrupted Bishop, Hostile Hostel, Ludevic, Necrogenius, Flame Channeler, Rem Karolus, Stalwart Slayer, Shadow Kin

## Watch items (a precedent exists; stop and report if it does not fit)

- **Empty the Laboratory** (4): sacrifice X Zombies at resolution, then a counted reveal-until.
- **Wake to Slaughter** (10): an opponent chooses between two targeted cards (`extrapolate_the_impossible.go`).
- **Visions cycle** (11): the discount is written with `CastFromGraveyard()`, which is true for any graveyard cast, not only the flashback one.
- **Sigarda's Vanguard, Sigardian Zealot, Celestial Judgment** (13): a battlefield pick that must have pairwise different powers (`rip_spawn_hunter.go` does it for cards).
- **Lynde, Cheerful Tormentor** (14): returning a Curse from a graveyard attached to a named player.
- **Sigarda's Splendor** (15): "note your life total" needs a number on the permanent that an ability rewrites (`Card.ChosenNumber` is set on entry).
- **Katilda, Dawnhart Prime** (16): a granted mana ability that reads the colours of the creature it is granted to.
- **Moorland Rescuer** (16): a pick bounded by total power.
- **Flame Channeler, Rem Karolus** (17): "a spell deals damage" read off `ev.SourceLKI`; must tell a spell from a permanent.
- **Ludevic, Necrogenius / Olag** (17): a copy with exceptions applied as the card transforms.

## Shared code split across batches

- Coven (`trekB48Coven`): batches 5, 12 and 13. Batch 5 owns any rename; nobody adds a second function.
- The decayed Zombie token row: batch 4, reused by Tainted Adversary and Gisa in batch 15.
- "Card with flashback or disturb" predicate: written in batch 7, reused by batch 11.
- `Flashback(cost)`: batches 3, 10 and 11 (existing constructor, nothing new).

## Cards that should not be built

None. No acorn or joke text, and nothing that needs an owner decision, so there is no "owner decision" batch.
