# Card-coverage plan: MOM (March of the Machine) and MOC

248 cards not yet in the catalog (209 `mom`, 39 `moc`), checked against `origin/develop`. 6 seams (19 cards), 23 batches (229 cards). Details, greps and oracle ids are in `plan.json`.

## Mechanics

| Mechanic | Exists | Evidence |
|---|---|---|
| Battle — Siege (defense, protector, defeated → cast transformed) | yes | game/battle.go (PendingChoiceChooseProtector, stampBattleEntryLocked), cards/effects/battles.go (BattleSpec, DefeatedTrigger, SiegeDefeated); 4 Sieges in catalog: invasion_of_innistrad/karsus/new_phyrexia/tarkir.go. Attack on a battle: game/attack_target.go AttackTargetBattle. TargetAny already admits battles (targets.go:449). |
| Backup | yes | grep -ri backup cards/effects: Backup(name,n,grantKey,keywords...) in soc_b6_helpers.go, used by guardian_scalelord.go. No tag on the stack item says 'this is a backup ability' (grep IsBackup/AbilityTag/AbilityKeyword in game: 0) — see seam s5. |
| Incubate / Incubator tokens | yes | cards/effects/incubate.go and incubate_cards_a_helpers.go; 'transform target Incubator token' in elvish_vatkeeper.go, progenitor_exarch.go. |
| Convoke | partly (see seam) | Partly: effects.Convoke() (tap_cost.go), PaidCost.Convoked (game/paid_cost.go:360, read by lethal_scheme.go), SpellsYouCastHaveConvoke (convoke_grant.go). Missing: NextSpellPromise has Improvise but no Convoke (game/next_spell_promise.go), and nothing answers 'a spell that has convoke' (grep 'has convoke/spell with convoke' in effects: 0) — seam s1. |
| Transforming double-faced cards with a Phyrexian-mana transform ability | yes | '{N}: Transform this creature. Activate only as a sorcery' in concealing_curtains.go; 21 files with 'transforms into' triggers; Phyrexian mana in activated costs (grep '/p}:' : dismember.go, norns_annex.go, 6 files; ADR 0131). |
| Praetor → Saga ('exile, then return transformed'; chapter III returns front face up) | yes | elesh_norn_mother_of_machines.go + elesh_norn_the_argent_etchings.go, jin_gitaxias_progress_tyrant.go + _the_great_synthesis.go, sheoldred_the_true_scriptures.go; saga.go. |
| Kicker with a non-mana cost | yes | 'Kicker—Sacrifice' in additional_cost.go, falling_timber.go, eject_the_warp_core.go. |
| Cycling / basic landcycling | yes | grep Landcycling cards/effects: 28 files; Typecycling 46. |
| Stun counters, prowess, proliferate, ward, surveil, bolster, dash, parley, horsemanship, flashback grant | yes | Stun 180 files; Prowess 85; Proliferate 75; bolster.go; dash in ragavan_nimble_pilferer.go; parley in selvala_explorer_returned.go, cmdr2_s49_helpers.go; horsemanship in lu_xun_scholar_general.go; 'gains flashback until end of turn' 10 files (flashback.go). |
| Will of the Planeswalkers / planeswalk / chaos ensues / planar dice (MOC) | partly (see seam) | Partly: Planechase is in the engine (ADR 0216: game/planeswalk.go PlaneswalkForEffect, game/planechase_die.go RollPlanarDieForEffect, effects_planechase.go; voting ADR 0146). planes_b1_helpers.go says 'No card of the catalog yet says "chaos ensues" as an effect; the batch that adds one widens this shape'. No planar-die 'roll one more and ignore one' static (grep 'plus one and ignore': 0; game/dice_roll.go has only the one-shot ModRollExtraDieIgnoreOne) — seam s2. |

## Seams (one strong-model worker each, before the batches that name them)

| id | Seam | Size | Cards |
|---|---|---|---|
| s1 | Make 'has convoke' a readable property and let a next-spell promise grant convoke | medium | Wand of the Worldsoul, Flockchaser Phantom, Joyful Stormsculptor, Kasla, the Broken Halo, Saint Traft and Rem Karolus, Invasion of Segovia |
| s2 | Let a spell planeswalk or make chaos ensue after a vote, and add the planar-die extra roll | medium | Path of the Animist, Path of the Enigma, Path of the Ghosthunter, Path of the Pyromancer, Path of the Schemer, Ichor Elixir |
| s3 | Record and expose that a creature attacked a battle | small | War Historian, Thrashing Frontliner, War-Trained Slasher, Rampaging Geoderm |
| s4 | Add protection from colorless as a quality and as a choice | small | Angelic Intervention |
| s5 | Mark backup abilities on the stack so a trigger can see one targeting a creature | small | Mirror-Shield Hoplite |
| s6 | Track whether a player was dealt combat damage since their last turn | small | Invasion of Fiora |

## Batches

| n | Theme | Cards | After |
|---|---|---|---|
| 1 | Vanilla, keyword and one-line creatures | 12 | — |
| 2 | Landcycling, Auras, Equipment and small artifacts | 12 | — |
| 3 | Backup granting keywords only | 11 | — |
| 4 | Backup granting a triggered ability | 7 | — |
| 5 | Backup granting static, activated and dies abilities; +1/+1 counter cards | 10 | — |
| 6 | Convoke: plain creatures and spells | 13 | — |
| 7 | Convoke: creatures that convoked it, and rules-heavier convoke spells | 8 | — |
| 8 | Simple instants | 11 | — |
| 9 | Knights and the 2/2 Knight token | 11 | — |
| 10 | The 1/1 blue-red Elemental token and red spells | 10 | s1 |
| 11 | Phyrexians, Incubator tokens and graveyard counters | 12 | — |
| 12 | Battle-matters: protector, battle targets and battle cards | 12 | — |
| 13 | Dig, loot and graveyard utility creatures and Auras | 12 | — |
| 14 | Commander-deck utility: blink, parley, artifact tokens | 9 | — |
| 15 | Phyrexian transformers I: simple back faces | 10 | — |
| 16 | Phyrexian transformers II: legends and heavier back faces | 9 | — |
| 17 | Sieges I: simple enters effect, simple back face | 10 | — |
| 18 | Sieges II: tokens, Equipment, Vehicle and sorcery back faces | 8 | — |
| 19 | Planeswalkers, token copies and free casts from graveyards | 8 | — |
| 20 | Team-up legends I: combat, sacrifice and recursion | 10 | — |
| 21 | Team-up legends II: casting from elsewhere, draws and mana | 9 | — |
| 22 | Praetor Sagas and transformed-permanent cards | 7 | — |
| 23 | Sieges III: rules-heavy, and Begin the Invasion | 8 | — |

## Watch items (a precedent exists, fit not proven: stop and report if it does not fit)

- Batch 2: Flywheel Racer's "Activate only if this permanent is a creature"; Tidal Terror's optional tap of two other creatures at resolution.
- Batches 4 and 5: backup grant bundles carrying a trigger, a block rule (Chomping Kavu) or a damage-assignment static (Streetwise Negotiator); "this creature" must mean the creature that gained the ability.
- Batch 7: Ancient Imperiosaur entering with counters per convoking creature; Cut Short's "planeswalker that was activated this turn".
- Batch 11: Phyrexian Censor's "one non-Phyrexian spell each turn" needs a filtered per-turn count; the cast tally has totals and noncreature only.
- Batch 12: Nahiri's Warcrafting's excess damage on a battle or planeswalker.
- Batch 16: Heliod, the Warped Eclipse (cards opponents have drawn this turn); Rona, Tolarian Obliterator.
- Batch 17: Awakened Skyclave's "it's a land in addition to its other types" on a back face cast from exile.
- Batch 20: Goro-Goro and Satoru and Zurgo and Ojutai (batched combat-damage triggers that filter or name their sources).
- Batch 21: Errant and Giada's "spells with flash or flying" filter; Omnath, Locus of All's black unspent mana.
- Batch 22: The Great Work III (cast from any graveyard), The Grand Evolution III (granted fight ability), "transformed permanent" on tokens.
- Batch 23: Lazotep Convert entering as a copy when cast transformed; Zilortha's group form; Begin the Invasion's several protector prompts.

## Cross-batch notes

- The Phyrexian transform sentence is printed on all 19 cards of batches 15 and 16. Both write the activated row inline; neither adds a helper.
- The Backup helper (`soc_b6_helpers.go`) is used unchanged by batches 3, 4 and 5; seam s5 may add a tag to it.
- "Creature, planeswalker, or battle" is printed in batch 12 and on Shatter the Source (batch 6), which builds it inline.
- Joyful Stormsculptor (s1) and batch 10 make the same existing 1/1 blue-red Elemental token; batch 10 runs after s1.

## Owner decisions

No card is proposed as do-not-build, so there is no "owner decision" batch. One question rides on seam s2: the five Path cards' "Will of the Planeswalkers" and Ichor Elixir only do anything in a game with a planar deck. Planechase is in the engine (ADR 0216), so they are planned as a seam; in an ordinary Commander game the vote happens and nothing follows (Ichor Elixir is a {4} rock for {C}{C}). If the owner would rather not spend a strong-model slot on Planechase text, drop s2: the six cards then wait.
