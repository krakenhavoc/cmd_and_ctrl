# Card-coverage plan: BRO (The Brothers' War) and BRC

237 cards not yet in the catalog (215 `bro`, 22 `brc`), checked against `origin/develop` at abb3b8bb3. Two seams (12 cards), 21 batches (225 cards). Machine-readable: `plan.json`.

## Mechanics

| Mechanic | Exists | Evidence |
|---|---|---|
| Prototype | **no** | grep -rni prototype in server/internal/cards, server/internal/game, docs/adding-cards.md, client/src/lib: only the card names Prodigy's Prototype, Bladegriff Prototype, U.S.S. Prometheus, Prototype. No keyword, no alternative-characteristics cast, no Layout=='prototype' handling. CR 702.160 / CR 718. |
| Powerstone tokens (restricted colourless mana) | yes | effects/tokens.go:332 PowerstoneToken(); game/mana_restriction.go:121 "can't be spent to cast a nonartifact spell" (#1727); CreateToken has Tapped. Used by cityscape_leveler.go, splitting_the_powerstone.go, gixs_caress.go. |
| Unearth | yes | effects/unearth.go Unearth(cost) / UnearthPaying; game/scoped_replacements.go exileInsteadOfLeaving (CR 702.84a). Granted to graveyard cards: sedris_the_traitor_king.go, solemn_doomguide.go. |
| Meld | yes | game/meld.go + effects/meld.go MeldWith / YouOwnAndControlThisAndANamed (ADR 0145, docs/adding-cards.md 'Meld'); urza_lord_protector.go is built; meld.go's header already names Mishra, Claimed by Gix's 'both attacking' condition. |
| Artifact offering (BRC, Blast-Furnace Hellkite) | **no** | grep -rnE '702\.48\|[Oo]ffering \(\|(artifact\|fox\|rat\|snake\|goblin\|moonfolk) offering\|difference in mana costs' in effects and game: nothing (no Kamigawa Patron is in the catalog). Nearest: Emerge (effects/alternative_cost.go:515) reduces by mana VALUE, gives no flash timing. |
| Encore, granted to graveyard cards (BRC, Wire Surgeons) | yes | effects/encore.go Encore(cost); graywaters_fixer.go grants encore to cards in a graveyard; effects/activated.go:769 ItsManaCost. |
| Overload (BRC, March of Progress) | yes | effects/alternative_cost.go:38 Overload(cost); 22 cards. |
| Affinity / Metalcraft / Fear / Goad (BRC returning words) | yes | effects/self_cost_modifier.go:52 AffinityFor(label, match); metalcraft in 13 card files (ability word, a count check); fear in game/ (10 files); goad in 19 game files. |
| Stun counters | yes | 'stun counter' in 98 effects files and 3 game files. |
| Mana spent to cast ({C}{C} spent, colours spent, amount spent) | yes | game/paid_cost.go SpentOfColor / ColorsSpentCount / ManaSpentCount; deceit.go deceitSpent(c) reads source.ManaSpentToCast().Count(c) >= 2; cast_trigger_entry_counters.go; aberrant_manawurm.go. |
| Phasing / 'Permanents can't phase in' (BRC) | yes | game/phasing.go:345 PermanentsCantPhaseIn; trek_b23_helpers.go:55. |
| End the turn | yes | docs/adding-cards.md 'Ending the turn (#2165, CR 724.1)'; 6 effects files. |

## Seams (built first, one strong-model worker each, who also builds the seam's cards)

| id | Seam | Size | Cards |
|---|---|---|---|
| s1 | Add prototype: cast a card with its alternative mana cost, colour, power and toughness | large | Blitz Automaton, Goring Warplow, Rust Goliath, Boulderbranch Golem, Combat Thresher, Cradle Clearcutter, Depth Charge Colossus, Fallaji Dragon Engine, Spotter Thopter, Iron-Craw Crusher, Autonomous Assembler |
| s2 | Add offering: sacrifice a permanent as you cast, reduce by its mana cost, cast at instant speed | medium | Blast-Furnace Hellkite |

Prototype is the only large piece: 18 cards have it. The seam worker builds the eleven whose other text is plain; batch 21 (after s1) holds the seven with copy, clone, animation or power-reading text, plus Mask of the Jadecrafter, which shares the X/X Golem token with Rootwire Amalgam.

## Batches

| n | Theme | Cards | After |
|---|---|---|---|
| 1 | Keyword creatures and plain pump/draw activations | 13 | — |
| 2 | Plain instants and sorceries | 13 | — |
| 3 | Powerstone makers I: enters, dies, or resolves and makes tapped Powerstones | 13 | — |
| 4 | Unearth I: artifact creatures with unearth and one plain ability | 12 | — |
| 5 | Soldiers I: Soldier tokens and Soldier-matters commons and uncommons | 12 | — |
| 6 | Mill-three-take-one cycle and creature-cards-in-graveyard counts | 13 | — |
| 7 | Second-card-each-turn and artifact-enters triggers | 12 | — |
| 8 | Sacrifice themes, the 3/3 Zombie artifact token and Scrap | 12 | — |
| 9 | Auras, an Equipment and exile-until-leaves enchantments | 11 | — |
| 10 | Graveyard hate, recursion and reanimation | 12 | — |
| 11 | Assembly-Workers and things that become creatures | 10 | — |
| 12 | Powerstone makers II: conditional and counted Powerstones | 11 | — |
| 13 | Unearth II: granted unearth and encore, play-one-of-the-exiled, enters-from-graveyard | 10 | — |
| 14 | Soldiers II: tap-three costs, attack triggers and legends | 9 | — |
| 15 | Library digs, piles and play-from-exile | 10 | — |
| 16 | Commands, sweepers and stack control | 10 | — |
| 17 | Mana spent to cast, and the legendary artifacts | 10 | — |
| 18 | Commander-deck and leftover one-offs | 8 | — |
| 19 | Copy effects | 8 | — |
| 20 | Meld pairs, the Saga and the planeswalker | 8 | — |
| 21 | Prototype II: prototype cards whose other text is rules-heavy | 8 | s1 |

### Cards by batch

- **1** Roc Hunter, Yotian Medic, Tomakul Honor Guard, Wing Commando, Dwarven Forge-Chanter, Fallaji Chaindancer, Retrieval Agent, Third Path Savant, Ambush Paratrooper, Argothian Sprite, Tyrant of Kher Ridges, Clay Revenant, Coastal Bulwark
- **2** Gaea's Gift, Machine Over Matter, Moment of Defiance, Scatter Ray, Union of the Third Path, Unleash Shell, Urza's Rebuff, Whirling Strike, Raze to the Ground, Shoot Down, Lay Down Arms, Gruesome Realization, Desynchronize
- **3** Stone Retrieval Unit, Argothian Opportunist, Koilos Roc, Powerstone Engineer, Horned Stoneseeker, Junkyard Genius, Great Desert Prospector, Geology Enthusiast, Excavation Explosion, Fallaji Excavation, Sibling Rivalry, Repair and Recharge, Hall of Tagsin
- **4** Ashnod's Harvester, Combat Courier, Mishra's Juggernaut, Reconstructed Thopter, Scrapwork Cohort, Scrapwork Mutt, Scrapwork Rager, Simian Simulacrum, Tocasia's Onulet, Yotian Frontliner, Heavyweight Demolisher, Artificer's Dragon
- **5** Aeronaut Cavalry, Air Marshal, Conscripted Infantry, Emergency Weld, Mass Production, Mishra's Onslaught, Veteran's Powerblade, Yotian Tactician, Zephyr Sentinel, Fortified Beachhead, Rescue Retriever, Siege Veteran
- **6** Airlift Chaplain, Blanchwood Prowler, Fallaji Archaeologist, Ravenous Gigamole, Tomakul Scrapsmith, Wasteful Harvest, Burrowing Razormaw, Gaea's Courser, Gixian Skullflayer, Gnarlroot Pallbearer, Battlefield Butcher, Deathbloom Ritualist, Overwhelming Remorse
- **7** Lat-Nam Adept, Thopter Mechanic, Gurgling Anointer, Evangel of Synthesis, Trench Stalker, Perimeter Patrol, Phalanx Vanguard, Thopter Architect, Yotian Dissident, Sarinth Steelseeker, Teething Wurmlet, Symmetry Matrix
- **8** Gixian Infiltrator, Goblin Blast-Runner, Kill-Zone Acrobat, Thraxodemon, Penregon Strongbull, Powerstone Fracture, Transmogrant Altar, Ashnod, Flesh Mechanist, Transmogrant's Crown, Supply Drop, Farid, Enterprising Salvager, Sardian Avenger
- **9** Aeronaut's Wings, Audacity, Draconic Destiny, Military Discipline, Mishra's Domination, Prison Sentence, Take Flight, Weakstone's Subjugation, Involuntary Cooldown, In the Trenches, Ashnod's Intervention
- **10** Carrion Locust, Gnawing Vermin, Hoarding Recluse, Swiftgear Drake, Keeper of the Cadence, Disciples of Gix, Hero of the Dunes, Recommission, No One Left Behind, Dreams of Steel and Oil, Loran, Disciple of History, Tocasia, Dig Site Mentor
- **11** Mine Worker, Power Plant Worker, Tower Worker, Mishra's Foundry, Levitating Statue, Alloy Animist, Mightstone's Animation, Tawnos's Tinkering, Argivian Avenger, Surge Engine
- **12** Arbalest Engineers, Thran Spider, Sarinth Greatwurm, Static Net, Slagstone Refinery, Wreck Hunter, Terisiare's Devastation, Visions of Phyrexia, Battery Bearer, Urza, Powerstone Prodigy, Mishra, Excavation Prodigy
- **13** Terisian Mindbreaker, Terror Ballista, Platoon Dispenser, Mishra, Tamer of Mak Fawa, Wire Surgeons, Meticulous Excavation, Mishra's Research Desk, Feldon, Ronom Excavator, Dredging Claw, Thran Vigil
- **14** Skystrike Officer, The Archimandrite, Harbin, Vanguard Aviator, Myrel, Shield of Argive, Recruitment Officer, Survivor of Korlis, Hajar, Loyal Bodyguard, Fallaji Vanguard, Citanul Stalwart
- **15** Forging the Anchor, Kayla's Reconstruction, Smelting Vat, Glint Raker, Hurkyl, Master Wizard, Sanwell, Avenger Ace, Hostile Negotiations, The Temporal Anchor, Queen Kayla bin-Kroog, Arms Race
- **16** Urza's Command, Kayla's Command, Gix's Command, Urza, Chief Artificer, Fade from History, The Fall of Kroog, Calamity's Wake, Defabricate, Over the Top, Hurkyl's Final Meditation
- **17** Bladecoil Serpent, Clay Champion, Steel Exemplar, Liberator, Urza's Battlethopter, Su-Chi Cave Guard, The Stasis Coffin, The Stone Brain, Urza's Sylex, Soul Partition, Fateful Handoff
- **18** Hexavus, Scavenged Brawler, Disciple of Caelus Nin, Rootpath Purifier, Urza's Workshop, Misery's Shadow, Sardian Cliffstomper, Spectrum Sentinel
- **19** March of Progress, Tawnos, Solemn Survivor, Tawnos, the Toymaker, Drafna, Founder of Lat-Nam, Urza, Prince of Kroog, Wondrous Crucible, Ashnod the Uncaring, Thopter Shop
- **20** Phyrexian Dragon Engine, Mishra, Claimed by Gix, Mishra, Lost to Phyrexia, Titania, Voice of Gaea, Argoth, Sanctum of Nature, Titania, Gaea Incarnate, The Brothers' War, Saheeli, Filigree Master
- **21** Arcane Proxy, Hulking Metamorph, Skitterbeam Battalion, Woodcaller Automaton, Phyrexian Fleshgorger, Steel Seraph, Rootwire Amalgam, Mask of the Jadecrafter

## Watch items (a precedent exists; not proven to fit; stop and report if it does not)

- **Rootpath Purifier** (18): "land cards in your library are basic". `StaticAbility.AffectsZones` (#3862, Maskwood Nexus) covers library cards, but `IsBasicLand` documents printed supertypes off the battlefield.
- **Ashnod the Uncaring** (19): "if one or more permanents were sacrificed to activate it" should be `PaidCost.Sacrificed` on the activated ability's stack item (Rings of Brighthearth trigger).
- **The Brothers' War** chapter II (20): forced attacks on a named player until your next turn; `MustAttack` (#3172) and Maximum Carnage are the nearest.
- **Mishra, Claimed by Gix** (20): the melded permanent enters tapped and attacking; `meld.go` has that only for tokens. **Mishra, Lost to Phyrexia**: choose three of six.
- **Titania, Voice of Gaea** (20): batched "land cards are put into your graveyard from anywhere" trigger.
- **Slagstone Refinery** (12) and **Urza's Sylex** (17): "put into exile from the battlefield" triggers; the one precedent is self-only and not from a cost.
- **Visions of Phyrexia** (12): "if you didn't play a card from exile this turn" as a per-turn fact. **Wreck Hunter** (12): counting cards put into a graveyard from the battlefield this turn.
- **The Temporal Anchor** (15): `EventScry.Amount` is the number bottomed; the card says "choose to".
- **Meticulous Excavation** (13): "if it has unearth" must see granted unearth. **Wire Surgeons** (13): encore cost equal to mana cost through `ItsManaCost`.
- **Disciple of Caelus Nin** (18): simultaneous phase-out of everything not chosen.
- **No One Left Behind** (10): cost reduction that depends on the target. **Tocasia, Dig Site Mentor** (10): targets with a total mana value cap.
- **Ashnod's Intervention** (9): the dies-or-exiled return, granted to a target.
- **Soul Partition** (17): the {2} tax applies only to an opponent's cast.

## Token rows and shared helpers (one owner each)

- Powerstone: exists (`PowerstoneToken()`), used by batches 3, 8, 12, 16, 17, 20. No new row.
- 1/1 colorless Soldier artifact: exists (`tokens_table.go`), batches 4, 5, 14. 2/2 green Bear and 1/1 Thopter with flying: exist.
- New rows: 3/3 colorless Zombie artifact and Scrap (batch 8); 2/2 colorless Construct artifact and the shared 0/0 Construct template lifted from `urzas_saga.go` (batch 16); X/X colorless Golem (batch 21).
- New helpers: mill-three-take-one (batch 6), "control creatures named A and B" (batch 11), discard-artifact-cards trigger (batch 12), exile-N-choose-one-to-play reuse (batch 13), "for each {C}{C} spent" (batch 17).

## Cards that should not be built

None. No acorn or joke text, and nothing needs an owner decision, so there is no "owner decision" batch.
