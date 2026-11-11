# Card-coverage plan: SNC (Streets of New Capenna) and NCC

251 cards not yet in the catalog (188 `snc`, 63 `ncc`), checked against `origin/develop` at abb3b8bb3. 4 seams (4 cards) and 23 batches (247 cards). Every oracle id appears exactly once (checked by script).

## Mechanics

| Mechanic | Exists | Evidence |
|---|---|---|
| Connive (CR 701.50) | yes | grep Connive: 8 files in internal/game, 61 in effects; effects/connive.go has Connive and connive X (raffine_scheming_seer.go). |
| Casualty N (CR 702.153) | yes | game/casualty.go + effects/casualty.go Casualty(n); 22 effects files. Granting it exists too (SpellsYouCastHave, silverquill_the_disputant.go). |
| Casualty X (Ob Nixilis, the Adversary) | no | grep -i 'Casualty X' and 'CasualtyX' over effects and game: 0 hits. Casualty(n) is a fixed PowerGE(n); resolveCasualty copies with no exception. Seam s2. |
| Blitz (CR 702.152) | yes | game/blitz.go + effects/blitz.go Blitz(cost); 44 effects files; no printed-blitz card is in this input. |
| Blitz granted to spells (Henzie) | no | grep 'has blitz', 'GrantBlitz': 0 hits; effects/granted_alternative_cost.go:383 (checkGrantedAlternativeCosts) refuses a granted offer with o.Blitz. Seam s1. |
| Alliance (ability word) | yes | 'Alliance —' in 26 effects files (bosss_chauffeur.go etc.); it is an enters trigger for another creature you control. |
| Hideaway N (CR 702.75) | yes | effects/hideaway.go func Hideaway; docs/adding-cards.md 'Adding a hideaway card (S43+, ADR 0091)'. |
| Shield counters (CR 122.1c) | yes | CounterShield in 47 effects files (counter_shields.go); 'shield counter' in 24. |
| Tri-colour 'exile this card from your hand: target land gains ... until this card is cast from exile' cycle | yes | grep 'until this card is cast from exile': emrakul_the_exigent_doom.go, game/duration.go UntilCardCast (#3520). |
| 'Five or more mana values among cards in your graveyard' (unnamed SNC threshold) | no | grep 'five or more mana values' and 'different mana values among cards in your graveyard': 0 hits. Not an engine seam: a count over the graveyard, one new helper in batch 12. |
| Fish token with "This token can't be blocked" | no | grep 'Fish creature token with': only reef_worm.go (a different Fish). New token row, all four cards in batch 7. |
| Citizen token (1/1 green and white) | yes | 'Citizen creature token' in 9 effects files (bosss_chauffeur.go, caldaia_guardian.go, grand_crescendo.go). |
| Treasure-mana and colours-of-mana read on an activation | yes | game/activated.go #3855 ('an ability that reads the mana spent to activate it'); effects/context.go ManaSpent/ColorsSpent; hired_hexblade.go for the cast side. |
| NCC returning keywords: encore, spectacle, scavenge, parley, split second, escape, support, manifest, addendum, aftermath, affinity, dethrone, melee, replicate, suspend, goad, proliferate, phasing | yes | each grepped in effects: encore 22 files, spectacle 11, scavenge 53, parley 6, split second 19, escape 53, support 21, manifest 77, addendum 2 (unbreakable_formation.go), aftermath 13, affinity 61, dethrone 2 (scourge_of_the_throne.go), melee 19, replicate 22 (granted: hugh_first_of_many.go), suspend 44, goad 91, proliferate 75, phase out 37. |

## Seams

| id | Seam | Size | Cards |
|---|---|---|---|
| s1 | Grant blitz to spells you cast, priced at their mana cost | medium | Henzie "Toolbox" Torre |
| s2 | Build casualty X with a modified copy | medium | Ob Nixilis, the Adversary |
| s3 | Add 'untap N tapped creatures you control' as an activation cost | small | Halo Fountain |
| s4 | Let a cast permission ask for a creature sacrifice in addition | small | Maestros Ascendancy |

Each seam was listed only after a search for a precedent found none; the greps are in `plan.json` (`checked`).

## Batches

| n | Theme | Cards | After | Card names |
|---|---|---|---|---|
| 1 | Vanilla, French vanilla and one-line creatures | 13 | - | Midnight Assassin, Fleetfoot Dancer, Crooked Custodian, Wrecking Crew, Chrome Cat, Jewel Thief, Glittermonger, Backup Agent, Speakeasy Server, High-Rise Sawjack, Obscura Initiate, Cabaretti Initiate, Riveteers Initiate |
| 2 | Simple instants and sorceries | 13 | - | Strangle, Antagonize, Daring Escape, Refuse to Yield, Demon's Due, Deal Gone Bad, Ready to Rumble, For the Family, Case the Joint, Prizefight, Buy Your Silence, Run Out of Town, Incandescent Aria |
| 3 | Sac-lands, plain Equipment and small artifacts | 12 | - | Racers' Ring, Tramway Station, Botanical Plaza, Waterfront District, Skybridge Towers, Jackhammer, Gilded Pinions, Quick-Draw Dagger, Ominous Parcel, Halo Scarab, Goldhound, Cement Shoes |
| 4 | Connive creatures and attachments | 11 | - | Raffine's Informant, Psionic Snoop, Echo Inspector, Revel Ruiner, Hypnotic Grifter, Raffine's Silencer, Illuminator Virtuoso, Security Bypass, Psychic Pickpocket, Obscura Interceptor, Mask of the Schemer |
| 5 | Casualty spells | 12 | - | Make Disappear, A Little Chat, Light 'Em Up, Grisly Sigil, Dig Up the Body, Rob the Archives, Rooftop Nuisance, Join the Maestros, Illicit Shipment, Cut Your Losses, Cut of the Profits, Audacious Swap |
| 6 | Shield counters | 11 | - | Boon of Safety, Swooping Protector, Brokers Veteran, Wingshield Agent, Voice of the Vermin, Disciplined Duelist, Rhox Pummeler, Dapper Shieldmate, Rigo, Streetwise Mentor, Sanctuary Warden, Family's Favor |
| 7 | Fish token and kinds of counters | 10 | - | Reservoir Kraken, Sleep with the Fishes, Expendable Lackey, Exotic Pets, Perrie, the Pulverizer, Bribe Taker, Agent's Toolkit, Aven Courier, Denry Klin, Editor in Chief, Contractual Safeguard |
| 8 | Counters matter (Brokers) | 10 | - | Metropolis Angel, Celestial Regulator, Skyboon Evangelist, Kros, Defense Contractor, Brokers Ascendancy, Park Heights Maverick, Revelation of Power, Backstreet Bruiser, Brokers Initiate, First Responder |
| 9 | Alliance and Citizens I | 13 | - | Celebrity Fencer, Attended Socialite, Social Climber, Elegant Entourage, Rumor Gatherer, Sizzling Soloist, Civil Servant, Darling of the Masses, Take to the Streets, Angelic Observer, Ceremonial Groundbreaker, Citizen's Crowbar, Gathering Throng |
| 10 | Citizens II and Cabaretti | 11 | - | Warm Welcome, Luxurious Libation, Exhibition Magician, Courier's Briefcase, Stimulus Package, Prosperous Partnership, Scepter of Celebration, Brazen Upstart, Park Heights Pegasus, Bess, Soul Nourisher, Civic Gardener |
| 11 | Treasure | 11 | - | Ognis, the Dragon's Lash, Hoard Hauler, Black Market Tycoon, Jetmir's Fixer, Most Wanted, Jolene, the Plunder Queen, Vazi, Keen Negotiator, Capenna Express, Spiteful Repossession, Paragon of Modernity, Boxing Ring |
| 12 | Mana values in your graveyard | 10 | - | Sanguine Spy, Tainted Indulgence, Graveyard Shift, Sewer Crocodile, Aven Heartstabber, Snooping Newsie, Syndicate Infiltrator, All-Seeing Arbiter, Oskar, Rubbish Reclaimer, Maestros Initiate |
| 13 | Sacrifice and Riveteers | 12 | - | Body Dropper, Forge Boss, Pyre-Sledge Arsonist, Riveteers Ascendancy, Crew Captain, Freelance Muscle, Mr. Orfeo, the Boulder, Cutthroat Contender, Dusk Mangler, Bellowing Mauler, Incriminate, Industrial Advancement |
| 14 | Removal and conditional-cost spells | 11 | - | Unleash the Inferno, Corpse Explosion, Hostile Takeover, Structural Assault, Out of the Way, Knockout Blow, Whack, Torch Breath, Even the Score, Body Count, Majestic Metamorphosis |
| 15 | Charms and Confluences | 11 | - | Brokers Charm, Maestros Charm, Riveteers Charm, Obscura Charm, Cabaretti Charm, Riveteers Confluence, Cabaretti Confluence, Obscura Confluence, Brokers Confluence, Maestros Confluence, Cleanup Crew |
| 16 | Graveyard recursion | 11 | - | Patch Up, Rogues' Gallery, Extraction Specialist, Jailbreak, Corpse Appraiser, Body Launderer, Cormela, Glamour Thief, Grime Gorger, Syrix, Carrier of the Flame, Next of Kin, Dodgy Jalopy |
| 17 | Attack triggers and combat (rules-heavy legends) | 11 | - | Kamiz, Obscura Oculus, Kitt Kanto, Mayhem Diva, The Beamtown Bullies, Phabine, Boss's Confidant, Shakedown Heavy, Maestros Diabolist, Ballroom Brawlers, Vivien's Stampede, Indulge // Excess, Arc Spitter, Scuttling Butler |
| 18 | Opponent choices and group effects | 9 | - | Protection Racket, Make an Example, Master of Ceremonies, Lord Xander, the Collector, Nimble Larcenist, Sky Crier, Mage's Attendant, Vampire Scrivener, Change of Plans |
| 19 | Hideaway and the exile-from-hand land cycle | 8 | - | Masked Bandits, Rakish Revelers, Spara's Adjudicators, Glamorous Outlaw, Shattered Seraph, Smuggler's Buggy, Wiretapping, Widespread Thieving |
| 20 | Exile-and-play, free casts and cast permissions | 10 | - | Urabrask, Heretic Praetor, Xander's Pact, Meeting of the Five, Evelyn, the Covetous, Evolving Door, Rocco, Cabaretti Caterer, Falco Spara, Pactweaver, Skyway Robber, Cryptic Pursuit, Sinister Concierge |
| 21 | Spell copies and granted spell keywords | 9 | - | Anhelo, the Painter, Threefold Signal, Parnesse, the Subtle Brush, Errant, Street Artist, Storm of Forms, Brass Knuckles, Flawless Forgery, Swindler's Scheme, Spellbinding Soprano |
| 22 | Copy effects, Auras and odd permanents | 9 | - | Undercover Operative, Cephalid Facetaker, Life of the Party, In Too Deep, Jinnie Fay, Jetmir's Second, Public Enemy, Hold for Ransom, False Floor, Turf War |
| 23 | Planeswalkers and exile-until permanents | 9 | - | Elspeth Resplendent, Vivien on the Hunt, Lagrella, the Magpie, Mysterious Limousine, Soul of Emancipation, Toluz, Clever Conductor, Crash the Party, Cabaretti Ascendancy, Obscura Ascendancy |

No batch depends on a seam. Batches 1 to 3 are simple and independent; 17 to 23 are rules-heavy (8 to 11 cards).

## Watch items (a precedent exists, fit not proven: stop and report if it does not fit)

- Batch 4: 'When it connives this way' reflexive trigger (Psychic Pickpocket, Obscura Interceptor).
- Batch 5: Audacious Swap's owner-casts-free permission.
- Batch 9: Sizzling Soloist's 'attacks during its controller's next combat phase if able'.
- Batch 11: mana spent to ACTIVATE (Jetmir's Fixer, Paragon of Modernity); Vazi's Treasure-mana trigger and Treasures created this turn; Boxing Ring's 'fought this turn' (no tally, read EventFight from EventsThisTurn).
- Batch 12: Oskar's cast-the-discarded-card permission.
- Batch 13: Dusk Mangler's ordered entry trigger (the registry's either-or row holds it for exactly this).
- Batch 14: Even the Score's coloured reduction.
- Batch 16: Scavenge on a Vehicle (Dodgy Jalopy); Grime Gorger's one-card-of-each-type pick.
- Batch 19: a colour-choice mana ability granted until the card is cast from exile (the five-card cycle); Wiretapping's first draw of the draw step.
- Batch 20: Falco Spara (RemoveCountersInAddition 1), Urabrask's opponent draw replacement, Meeting of the Five's spend restriction, Skyway Robber's escape-exiled cards.
- Batch 21: Anhelo's 'first instant or sorcery spell each turn' on a granted casualty.
- Batch 22: Jinnie Fay's two-kind optional token replacement; Public Enemy's attack requirement.

## New token rows (each in one batch only)

- Fish 1/1 blue, can't be blocked: batch 7. Ogre Warrior 4/3 black: batch 5. Wizard 1/1 blue with a sacrifice ability: batch 18. Cat 2/2 green haste and Dog 3/1 green vigilance: batch 22.
- Citizen, Treasure, Devil, Angel, Rhino Warrior and Spirit rows already exist on develop.

## Cards that should not be built

None. The input has no acorn or joke cards and nothing that needs an owner decision, so there is no 'owner decision' batch.
