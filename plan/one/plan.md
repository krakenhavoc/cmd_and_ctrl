# Card-coverage plan: ONE (Phyrexia: All Will Be One) and ONC

233 cards not yet in the catalog (212 `one`, 21 `onc`), checked against `origin/develop`. 3 seam items (14 cards) and 20 batches (219 cards).

## Mechanics

| Mechanic | Exists | Evidence |
|---|---|---|
| Toxic | yes | grep -i toxic game: game/infect_wither_toxic.go (ToxicValue, ToxicTotal, CR 702.164), keywords.go; 33 effects files (skrelvs_hive.go, skrelv_defector_mite.go). |
| Proliferate (the action) | yes | game/proliferate.go, game/keyword_action.go (KeywordActionProliferate, CR 701.34); 101 effects files. |
| 'Whenever you proliferate' triggers | **no** | grep -ri 'you proliferate' game cards/effects: 0; events.go:888 states there is no EventProliferate. Seam s1. |
| Corrupted (ability word) | yes | Ability word, no rules meaning (CR 207.2c). Predicates rfReprintBCorrupted, seedcoreCorrupted; cards skrelvs_hive.go, the_seedcore.go, corrupted_conviction.go. |
| For Mirrodin! | **no** | grep -ri mirrodin game cards/effects: no keyword or helper. LivingWeapon (living_weapon.go) and token row '2/2 red Rebel' exist. Seam s2 (small). |
| Oil counters | yes | Plain named counters: vat_of_rebirth.go, mindsplice_apparatus.go, chocobo_knights.go, batch41_helpers.go; RemoveCountersFromThis / RemoveCountersAmong in activated.go. |
| Compleated and Phyrexian mana | yes | compleated.go Compleated() (CR 702.150a), game/phyrexian_mana.go, paid_cost.go PhyrexianLife; card vraska_betrayals_sting.go. Hybrid Phyrexian symbols are mentioned in game/mana.go and paid_cost.go but no catalog card uses one (watch item, batch 20). |
| Poison counters on players | yes | game/player_counters.go, game_end.go; ichor_rats.go, swamp_mosquito.go. |
| Affinity (for artifacts / for Equipment) | yes | AffinityFor(label, predicate) in self_cost_modifier.go; 72 effects files. No 'Affinity for Equipment' card yet, same helper with another predicate. |
| Battle cry, melee, lieutenant, experience counters (ONC) | yes | battle_cry.go BattleCry(); melee in game/keywords.go and glory_to_the_empire.go; lieutenant in 33 effects files; experience in game/counter_types.go. |
| Sphere land subtype | yes | the_seedcore.go and the_mycosynth_gardens.go (both Land — Sphere) are in the catalog; the subtype registry itself was not read. |

## Seams (one strong-model worker each, before any batch that names it)

| id | Seam | Size | Cards |
|---|---|---|---|
| s1 | Emit a proliferate event and add a 'whenever you proliferate' trigger | medium | Ezuri, Stalker of Spheres, Scheming Aspirant, Venser, Corpse Puppet, Voidwing Hybrid |
| s2 | Add the For Mirrodin! keyword helper | small | Barbed Batterfist, Dragonwing Glider, Goldwarden's Helm, Hexgold Halberd, Hexgold Hoverwings, Kemba's Banner, Mirran Bardiche, Sylvok Battle-Chair, Vulshok Splitter |
| s3 | Add an 'exile a spell you control' activation cost | small | Synthesis Pod |

## Batches

| n | Theme | Cards | After |
|---|---|---|---|
| 1 | Vanilla and keyword-only creatures | 13 | — |
| 2 | Sphere lands, Skullbombs and simple mana | 13 | — |
| 3 | Simple enters and dies triggers | 13 | — |
| 4 | Phyrexian Mites and poison-counter spells | 13 | — |
| 5 | Toxic matters and toxic grants | 12 | — |
| 6 | Proliferate creatures and an Aura | 10 | — |
| 7 | Corrupted commons and uncommons | 13 | — |
| 8 | Proliferate spells and simple instants and sorceries | 13 | — |
| 9 | Oil: noncreature-spell counters | 9 | — |
| 10 | Oil: enters with counters, remove a counter to activate | 13 | — |
| 11 | Oil: graveyard triggers and oil counts | 12 | — |
| 12 | Oil: rares and artifacts | 9 | — |
| 13 | Artifacts matter | 12 | — |
| 14 | Equipment matters | 11 | s2 |
| 15 | Remaining spells and Auras | 9 | — |
| 16 | Toxic and corrupted rares (rules-heavy) | 8 | — |
| 17 | Legends and Commander cards I | 9 | — |
| 18 | Legends II (rules-heavy) | 9 | — |
| 19 | Copies and casting from exile (rules-heavy) | 8 | s2 |
| 20 | Planeswalkers and the toxic Beast token (rules-heavy) | 10 | — |

## Watch items (a precedent exists; stop and report if it does not fit)

- **Batch 5** Plague Nurse: granted toxic must add to existing toxic. Necrogen Rotpriest: the extra poison counter is its own trigger.
- **Batch 12** Churning Reservoir: "an oil counter was removed ... this turn or a permanent with an oil counter on it was put into a graveyard this turn" has no turn tally.
- **Batch 15** Green Sun's Twilight: per-card "onto the battlefield or into your hand".
- **Batch 16** Geth's Summons (targets counted as you cast), Noxious Assault (delayed "whenever a creature blocks this turn"), Venerated Rotpriest, Kinzu, Glissa's Retriever.
- **Batch 17** Rhuk: moving all Equipment from a creature that died.
- **Batch 18** Melira (cap poison at one, then none this turn), Encroaching Mycosynth (types off the battlefield), Thrun (targeting restriction by colour of spell or source), Slobad, Geth, Ria Ivor.
- **Batch 19** Chiss-Goria (affinity on one spell cast from exile), Ixhel (face-down exile from each corrupted opponent, play with any colour).
- **Batch 20** Kaito (loyalty twice this turn, granted by a trigger), Viral Spawning (conditional flashback), Nahiri (forced attack on a player; hybrid Phyrexian mana), Lukka (hybrid Phyrexian mana).

Of these, Kaito and Churning Reservoir are the two most likely to turn out to be small seams.

## Shared sentences split across batches

- "an opponent has three or more poison counters" (corrupted): batches 7, 8, 16, 19, 20. The predicate already exists; nobody adds another.
- "Whenever you cast a noncreature spell, put an oil counter on this creature": batch 9 owns the helper; Mercurial Spelldancer (batch 19) writes it inline.
- For Mirrodin!: s2 owns the helper; batches 14 and 19 use it.
- Phyrexian Mite token (exists): batches 2, 4, 18. Phyrexian Golem and Phyrexian Goblin token rows exist: batches 3, 10, 13.
- New token rows, each in one batch only: Phyrexian Wurm X/X trample toxic 1 (16); 3/3 green Phyrexian Beast with toxic 1 (20); X/1 Phyrexian Horror (12); Drone with deathtouch, X/X Phyrexian Horror, emblem (20); The Hollow Sentinel (s1); X/X Golem with haste (12). The 2/2 white Cat row exists (17).

## Cards that should not be built

None. No acorn or joke text and nothing needing an owner decision, so there is no "owner decision" batch.
