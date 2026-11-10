# Card-coverage wave plan: Commander precons (mkc, otc, lcc, dsc, blc, ecc, woc, drc, tdc)

536 cards not yet in the catalog. Checked against `fix-r2` (develop plus the roll-up). 8 seam items cover 13 cards and go first; the other 523 cards are in 52 batches of 8 to 13, paired for 26 workers. Machine-readable detail, including the greps behind each seam and the helpers each batch should reuse, is in `plan.json`.

## Seams (build before any card batch; the seam worker also builds these cards)

| id | seam | size | cards |
|---|---|---|---|
| s1 | Build clash (CR 701.30) and the 'win a clash' event | medium | 2: Hoarder's Greed, Marvo, Deep Operative |
| s2 | Open the keyword-action window on explore and give investigate an event | medium | 3: Topography Tracker, Hakbal of the Surging Soul, Erdwal Illuminator |
| s3 | Build leveler cards (level up, CR 702.87 / CR 711) | medium | 1: Coralhelm Commander |
| s4 | Let Grist be a creature card off the battlefield (and a legal commander) | medium | 1: Grist, the Hunger Tide |
| s5 | Grant evoke and offspring to spells as they are cast | medium | 2: Ashling, the Limitless, Zinnia, Valley's Voice |
| s6 | Grant a graveyard ability whose cost is the card's own mana value | small | 1: Graywater's Fixer |
| s7 | Manifest the top card of another player's library | small | 2: Thieving Amalgam, Orochi Soul-Reaver |
| s8 | Mark the cards a player surveilled into their graveyard this turn | small | 1: Eye of Duskmantle |

s6, s7 and s8 are small and touch different files (zone grants, manifest, the surveil route), so one worker could take all three. s2 changes how investigate is called; batches 31, 32 and 49 investigate through the existing card-side helper and are safe either way, but merge s2 before them if you can.

## Batches

| n | theme | cards | pair |
|---|---|---|---|
| 1 | Simple lands | 12 | 2 |
| 2 | Mana rocks, mana creatures and mana triggers | 10 | 1 |
| 3 | Simple creatures and small permanents | 11 | 4 |
| 4 | Choose a creature type | 9 | 3 |
| 5 | Merfolk tribal | 10 | 6 |
| 6 | Pirates | 10 | 5 |
| 7 | Dinosaurs: lords, cost reducers, simple triggers | 10 | 8 |
| 8 | Enrage, damage triggers and damage modifiers | 10 | 7 |
| 9 | Vampires: lords, tokens, Vampire Demon | 11 | 10 |
| 10 | Vampires: Blood tokens, partner pairs, Athreos | 10 | 9 |
| 11 | Faeries | 10 | 12 |
| 12 | Elementals, evoke and reveal-a-card lands | 11 | 11 |
| 13 | Deserts and creature lands | 9 | 14 |
| 14 | Desert payoffs and lands in the graveyard | 11 | 13 |
| 15 | Land recursion and self-mill for lands | 10 | 16 |
| 16 | Land ramp, landfall, untap lands | 9 | 15 |
| 17 | Squirrels, Rats and Saprolings | 10 | 18 |
| 18 | Bloomburrow legends, offspring and Food | 11 | 17 |
| 19 | Auras: simple | 10 | 20 |
| 20 | Aura payoffs and Auras entering without being cast | 9 | 19 |
| 21 | Equipment, artifact creatures, a Vehicle and base-P/T matters | 10 | 22 |
| 22 | Draw-matters, hand size and flash grants | 11 | 21 |
| 23 | +1/+1 counters: simple and adapt | 10 | 24 |
| 24 | +1/+1 counters: enters-with, moving, alliance | 10 | 23 |
| 25 | -1/-1 counters: placing them | 11 | 26 |
| 26 | -1/-1 counter payoffs and blight | 11 | 25 |
| 27 | Surveil basics and simple card selection | 10 | 28 |
| 28 | Surveil payoffs | 9 | 27 |
| 29 | Delirium | 12 | 30 |
| 30 | Reanimation and graveyard creatures | 12 | 29 |
| 31 | Investigate: making Clues | 10 | 32 |
| 32 | Clues as objects, detectives and exile-until-leaves | 8 | 31 |
| 33 | Outlaws, Treasure and Mercenaries | 11 | 34 |
| 34 | Goad and attack politics | 11 | 33 |
| 35 | Morph and megamorph: straightforward | 10 | 36 |
| 36 | Morph: megamorph payoffs, returns face down, spell theft | 9 | 35 |
| 37 | Manifest and face-up payoffs (Commander 2015/Khans) | 8 | 38 |
| 38 | Disguise, cloak and face-up payoffs (Karlov Manor) | 8 | 37 |
| 39 | Split, aftermath, flashback, jump-start and board wipes | 13 | 40 |
| 40 | Suspend, foretell, plot, miracle and casting from outside the hand | 12 | 39 |
| 41 | Spellslinger: storm, cascade and spell copies | 11 | 42 |
| 42 | Casting instants and sorceries from graveyards | 8 | 41 |
| 43 | Impulse play, cast-from-top and life-for-mana costs | 10 | 44 |
| 44 | Theft from libraries and hands | 8 | 43 |
| 45 | Ninjutsu and combat-damage theft | 8 | 46 |
| 46 | Encore, myriad and tokens that enter attacking | 12 | 45 |
| 47 | Token copies, token replacement and gifts | 10 | 48 |
| 48 | Monarch courts and clones | 10 | 47 |
| 49 | Parley, tempting offer, secret and group choices | 10 | 50 |
| 50 | Duskmourn enchantments, horrors and hideaway | 10 | 49 |
| 51 | Planeswalkers | 8 | 52 |
| 52 | Complex legends and one-offs | 9 | 51 |

Order runs from independent and simple (lands, tribal lords) to rules-heavy (face-down, theft, planeswalkers, complex legends). Batches 35 to 38, 42, 44, 45, 51 and 52 are kept at 8 to 10 on purpose.

## Helpers that only one batch may write

- New tokens: Sand Warrior and Plant Warrior (14), Ox (20), Hamster (21), Scarecrow (26), 3/3 Dinosaur (7), Pirate and Skeleton Pirate (6), goaded Pirate and Bird (34), Storm Crow and Dragon Elemental (41), Lightning Rager (34), Tiny (32), Octopus gift (47), Wrenn's Treefolk and Garruk's Wolf (51).
- Keywords with no helper yet, one card each: afterlife (3), vanishing (17), graft (24), jump-start (39). Gift an Octopus and gift an extra turn are both in 47. Parley is lifted into one helper in 49.
- Three sentences cross batches and are flagged in `shared` so the second worker greps first: "discard two unless you discard a <type> card" (21, 27, 51), the face-down exile and any-type-mana theft sentence (44, 45), and the graveyard free-cast sentence (42, 46).

## Watch items (not seams: a precedent or a generic hook exists, but the worker should stop and report if it does not fit)

- Umbra Mystic (20): umbra armor is a replacement on each Aura's own Spec; granting it to other Auras is untested.
- Boltbender (38): `retarget.go` retargets one spell or ability, the card retargets any number.
- Selvala, Explorer Returned (49): parley inside a mana ability.
- Evercoat Ursine (50): two hideaway instances on one permanent.
- Grothama and Illusionist's Gambit (52): per-player damage totals on one permanent; an extra combat with a per-combat attack requirement.
- Feather, Radiant Arbiter and Forger's Foundry (41), Finale of Promise (42), Esix (47): long but each clause has a precedent.

## Cards not to build

None. There are no silver-border, Archenemy-only or joke cards in the list, so `plan.json` has no owner-decision batch. One owner decision sits inside seam s4: Grist can only be a commander if `deck/validate.go` stops requiring a Legendary Creature type line or the "can be your commander" text.
