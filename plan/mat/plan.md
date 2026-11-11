# Card-coverage plan: mat (March of the Machine: The Aftermath)

40 cards not yet in the catalog, all `set: mat`, all `layout: normal`. Checked against `origin/develop`. 3 seam items (3 cards), 4 batches (37 cards).

## Mechanics

| Mechanic | Exists | Evidence |
|---|---|---|
| (no new set mechanic) | yes | March of the Machine: The Aftermath (40 cards in the input, all layout=normal, all set=mat) prints no keyword or keyword action of its own; it reuses the ones below. |
| Convoke | yes | grep 'func Convoke(' in server/internal/cards/effects/ -> helper exists; many catalog cards (adept_watershaper.go, appeal_to_eirdu.go). |
| Madness | yes | grep -i Madness in server/internal/cards/effects/ -> alternative_cost.go and cards; game.AltCostKeyMadness in server/internal/game/cast_permission.go. |
| Affinity (for Equipment) | yes | grep 'func AffinityFor(' -> exists; used for artifacts, creatures, Vulcans, Daleks. No 'Affinity for Equipment' card yet, but the helper takes the counted kind. |
| Ward (printed and granted) | yes | grep 'func Ward' -> Ward, WardMana, WardGranted, WardAttached; 'have ward' granted in flowering_of_the_white_tree.go, bronze_guardian.go. |
| Hexproof from monocolored | yes | grep 'hexproof from monocolored/multicolored' -> niv_mizzet_guildpact.go, dragonfire_blade.go. |
| Prowess granted to creatures you control | yes | grep 'ha(s/ve) prowess' -> bria_riptide_rogue.go, wizards_staff.go. |
| Constellation / Landfall / Undergrowth (ability words) | yes | grep 'func Landfall(' exists; 'Undergrowth' in necrotic_wound.go; enchantment-enters triggers are ordinary ETB triggers. |
| Jump-start on the card itself | yes | grep -i jump-start -> radical_idea.go, cmdr2S39JumpStart in cmdr2_s39_helpers.go (flashback offer plus a discard). |
| Jump-start GRANTED to cards in a graveyard (Niv-Mizzet, Supreme) | **no** | grep 'JumpStart/jump-start' in server/internal/game -> nothing. Lier grants flashback through CastPermission{AltCostKey:"flashback"}; cast_permission.go has no branch for a granted jump-start (discard as an additional cost) and PermissionFilter has ColorsAny but no 'exactly two colours'. Seam s2. |

## Seams (one strong-model worker each, before the batches; the worker also builds the card)

| id | Seam | Size | Cards |
|---|---|---|---|
| s1 | Carry the amount of mana a mana ability produced on its event | small | Tyvar the Bellicose |
| s2 | Grant jump-start to graveyard cards that are exactly two colours | small | Niv-Mizzet, Supreme |
| s3 | Replace damage that would remove a planeswalker's last loyalty counter | small | Deification |

No batch depends on a seam.

## Batches

| n | Theme | Count | After | Cards |
|---|---|---|---|---|
| 1 | Granted ward, entered-this-turn, simple spells and bodies | 10 | - | Coppercoat Vanguard, Gold-Forged Thopteryx, Leyline Immersion, Drannith Ruins, Samut, Vizier of Naktamun, Animist's Might, Blot Out, Filter Out, Ayara's Oathsworn, Markov Baron |
| 2 | Graveyard returns, died-this-turn counts, distributed counters, resolution choices | 9 | - | Cosmic Rebirth, Campus Renovation, Harnessed Snubhorn, Death-Rattle Oni, Feast of the Victorious Dead, Undercity Upheaval, Tolarian Contempt, Tranquil Frillback, Nahiri's Resolve |
| 3 | Casting and playing from library, graveyard and exile; legends that put cards into play | 9 | - | Sigarda, Font of Blessings, Danitha, New Benalia's Light, Nahiri, Forged in Fury, Kiora, Sovereign of the Deep, Kolaghan Warmonger, Nissa, Resurgent Animist, Pia Nalaar, Consul of Revival, The Kenriths' Royal Funeral, Arni Metalbrow |
| 4 | Copies, animation and layer work (rules-heavy) | 9 | - | Narset, Enlightened Exile, Nashi, Moon's Legacy, Rebuild the City, Calix, Guided by Fate, Vesuvan Drifter, Jolrael, Voice of Zhalfir, Spark Rupture, Urborg Scavengers, Tazri, Stalwart Survivor |

The set is small, so batches 1 to 3 run at 9 or 10 cards rather than 10 to 13.

## Watch items (a precedent exists but the fit is unproven: stop and report if it does not fit)

- **Leyline Immersion** (batch 1): a granted five-mana any-combination ability with a spend restriction.
- **Tranquil Frillback** (batch 2): Hawkeye, Master Marksman is the same sentence; check its helper takes three different targeted modes.
- **Nahiri's Resolve** (batch 2): an any-number choice on resolution feeding a delayed return of exactly those cards.
- **Nahiri, Forged in Fury** (batch 3): "You may cast Equipment spells this way without paying their mana costs" has no precedent; try a second permission with a cost and a subtype filter.
- **The Kenriths' Royal Funeral** (batch 3): a cost reduction counted from cards exiled with the source.
- **Arni Metalbrow** (batch 3): "attacks or enters attacking" has no trigger precedent, and `game.Event` has no entered-attacking field. If an ETB trigger reading the card's `AttackingTarget` does not work, this is a small seam.
- **Nashi, Moon's Legacy** (batch 4): `CopyAndCast` with a permanent-card copy (the copy becomes a token, CR 707.10f).
- **Vesuvan Drifter** (batch 4): becoming a copy of a card in the library until end of turn, with an exception.
- **Spark Rupture** (batch 4): no class-animation static exists; build from three generic statics. Check layer-cache invalidation on counter changes and damage to a creature-planeswalker.
- **Tazri, Stalwart Survivor** (batch 4): "any of this creature's colors" and "Activate only if this creature has another activated ability" have no precedent.

## Cards not to build

None. No acorn or joke text, and nothing needs an owner decision, so there is no "owner decision" batch.
