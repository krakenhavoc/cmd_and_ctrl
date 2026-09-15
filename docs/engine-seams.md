# Engine seams

A seam is an engine gap a card in the catalog (`server/internal/cards/effects`)
is waiting on: an event kind, a cost component, a prompt shape, or a
primitive that `server/internal/game` does not yet provide. A seam is not a
bug: bugs found by the roadmap sweep (issues #446, #478, #482, #489, #492)
are tracked and fixed on their own, separately from this list.

This file is the seam registry for the card-coverage roadmap (Discussion #559 item 6, issue #580). It replaces
the seam lists that used to live only in agent prompts and in scattered
batch-issue comments. The rule going forward: every batch PR appends its
own skipped cards to the table below, merging into an existing row when the
seam is the same one already listed, and a seam moves from Open to Closed
on the day the engine PR that closes it lands.

The table covers every seam reported against **two or more** waiting cards
across the first pass over the first 3600 cards (batches 01-36, issues
#294-#313 and #383-#399). A seam that blocks exactly one card today is
noted on that card's own batch issue rather than duplicated here; it will
earn a row here the moment a second card waits on it.

## Open seams

| Seam | What is missing | Cards waiting | Count | Tracked |
|---|---|---|---:|---|
| Modal / multi-target clause machinery | prompt: a `Modes` slot on `TriggeredAbility` (and spells) for choose-one-or-more, with per-mode or per-slot target constraints; `Register` panics once more than one mode carries its own target | Black Market Connections (#294); Nesting Grounds (#297); Archdruid's Charm (#298); Decimate (#301); Monument to Endurance (#301); Prismari Command (#302); Ram Through (#302); Junji, the Midnight Sky (#304); +18 more | 26 | |
| Generic step-entry trigger event | event kind: a triggerable event for the untap step, beginning of combat, the postcombat main phase, or end of combat (upkeep, draw step, precombat main and end step are the only steps with an event today) | Black Market Connections (#294); Seedborn Muse (#295); Unwinding Clock (#297); Ripples of Undeath (#297); Bender's Waterskin (#298); Mesmeric Orb (#302); Intruder Alarm (#308); Drumbellower (#310); +7 more | 15 | #588 |
| Counter-count cost component | cost component: `AbilityCost` / `ManaAbilityCost` cannot add or remove a counter as part of paying a cost | Walking Ballista (#296); Devoted Druid (#302); Tome of Legends (#383); Steelbane Hydra (#384); Vat of Rebirth (#386); Khalni Heart Expedition (#386); Scholar of New Horizons (#387); Fertilid (#390); +7 more | 15 | |
| Put-from-hand onto the battlefield | prompt + primitive: a pick-from-hand prompt with a continuation and a hand-to-battlefield move (casting or playing a permanent card from hand for free) | Growth Spiral (#294); Archaeomancer's Map (#299); Kodama of the East Tree (#301); Ghalta, Stampede Tyrant (#302); Sneak Attack (#306); Cultivator Colossus (#307); Terrain Generator (#310); Arboreal Grazer (#313); +7 more | 15 | |
| Opponent's non-mana choice at resolution | prompt: an opponent-facing choice or free yes/no at resolution; only `PayUnless` / `MayPay` (mana-only) and trigger prompts exist | Braids, Arisen Nightmare (#295); Torment of Hailfire (#298); Chain of Vapor (#301); Combustible Gearhulk (#302); Charismatic Conqueror (#302); Painful Quandary (#303); Tempt with Bunnies (#311); Druid of Purification (#385); +5 more | 13 | |
| Reveal-from-hand entry choice | prompt: a permanent's own entry replacement offering to reveal a card from hand to change how it enters; only the pay-life variant exists | Choked Estuary (#295); Foreboding Ruins (#295); Port Town (#295); Fortified Village (#295); Frostboil Snarl (#295); Furycalm Snarl (#295); Game Trail (#296); Shineshadow Snarl (#296); +4 more | 12 | |
| Tap-another-permanent cost component | cost component: `AbilityCost` / `ManaAbilityCost` cannot require tapping an untapped creature or permanent you control other than itself | Relic of Legends (#297); Springleaf Drum (#297); Clock of Omens (#308); Uthros Research Craft (#309); Susur Secundi, Void Altar (#310); Survivors' Encampment (#386); Dawnsire, Sunstar Dreadnought (#387); Holdout Settlement (#390); +1 more | 9 | |
| Ability granted to another permanent by a static | primitive: `ManaAbilitiesForCard` (and the activated/triggered equivalents) reads only a permanent's own list or catalog entry; nothing lets a static grant an ability to another permanent | Cryptolith Rite (#298); Rishkar, Peema Renegade (#300); Jaheira, Friend of the Forest (#301); Marvin, Murderous Mimic (#302); Insidious Roots (#306); Great Divide Guide (#388); Gemhide Sliver (#393); Ultima, Origin of Oblivion (#393); +1 more | 9 | |
| Per-player "cast as though it had flash" | primitive: a per-player permission letting spells be cast at instant speed; only the printed-flash keyword path exists | Emergence Zone (#297); Borne Upon a Wind (#298); High Fae Trickster (#300); Alchemist's Refuge (#302); Liberator, Urza's Battlethopter (#304); Shimmer Myr (#305); Vedalken Orrery (#306); Gandalf the White (#387); +1 more | 9 | |
| Trigger doubling ("triggers an additional time") | primitive: a static that makes a triggered ability trigger an additional time per event (the Panharmonicon shape) | Panharmonicon (#295); Annie Joins Up (#302); Elesh Norn, Mother of Machines (#302); Teysa Karlov (#304); Isshin, Two Heavens as One (#306); Echoes of Eternity (#306); Yarok, the Desecrated (#385); Gandalf the White (#387) | 8 | |
| Effect-reachable RNG / die roll | primitive: the persisted seeded RNG is private to shuffle/persistence code; no `RollDieForEffect`-style method lets a card effect roll dice or flip a coin | Ancient Copper Dragon (#298); Ancient Gold Dragon (#312); Wyll's Reversal (#388); Deadbridge Chant (#394); Reckless Endeavor (#395); The Gold Saucer (#395); Hoarding Ogre (#397); Vexing Puzzlebox (#397) | 8 | |
| Token-creation replacement event kind | event kind: creating a token is not a replaceable event, so nothing can double or intercept "create a token" | Academy Manufactor (#295); Anointed Procession (#296); Parallel Lives (#296); Peregrin Took (#299); Xorn (#301); Stridehangar Automaton (#308); Divine Visitation (#310) | 7 | |
| Once-per-turn trigger / activation tally | primitive: a per-turn tally for triggers or activations that fire or activate at most once a turn (`Game.TurnTally` lands with #586; the cards still need writing) | Morbid Opportunist (#295); Welcoming Vampire (#296); Terrasymbiosis (#300); Tocasia's Welcome (#300); Monument to Endurance (#301); Quirion Ranger (#313); Wirewood Symbiote (#396) | 7 | #586 |
| Discard-a-card cost on an activated ability | cost component: `AbilityCost` has no discard-a-card shape | Fauna Shaman (#306); Tortured Existence (#307); Cryptbreaker (#311); Survival of the Fittest (#384) | 4 | |
| "Can't cast" restriction gate | primitive: no catalog hook can prevent a player from casting a spell under a static condition (non-legendary-sorcery cases) | Deafening Silence (#311); Archon of Emeria (#312); Rule of Law (#395); Eidolon of Rhetoric (#396) | 4 | |
| Ability activatable from a non-battlefield zone | primitive: activated abilities are reachable from the battlefield only; hand and graveyard activation have no path | Simian Spirit Guide (#296); Elvish Spirit Guide (#299); Reassembling Skeleton (#299); Drownyard Temple (#311) | 4 | |
| X on an activated ability | cost component: `AbilityCost` has no `{X}` component | Treasure Vault (#296); Blast Zone (#308); Gogo, Master of Mimicry (#396); Magus of the Candelabra (#398) | 4 | |
| Search / library-interaction primitives | primitive: no search-replacement or depth limit on a library search, no mill-as-a-replacement-event kind, no non-owner searcher, no library-top cast path | Aven Mindcensor (#304); Bruvac the Grandiloquent (#312); Bribery (#390); Conspicuous Snoop (#394) | 4 | |
| Layer invalidation on hand / life / attack / graveyard state | primitive: hand-size, life-total, attack-declaration and graveyard-arrival changes do not bump `layerVersion`, so statics keyed on them go stale | Psychosis Crawler (#296); Ohran Frostfang (#296); Wonder (#300); Serra Ascendant (#302) | 4 | |
| Ability-suppression static | primitive: "loses all abilities" only strips keywords; nothing suppresses a permanent's triggered, static or activated abilities, and a later grant can re-add a stripped keyword | Archetype of Imagination (#310); Tishana's Tidebinder (#388); Torpor Orb (#396); Dress Down (#399) | 4 | |
| Return-a-permanent-to-hand cost component | cost component: `AbilityCost` has no "return a permanent you control to its owner's hand" shape | Quirion Ranger (#313); Master Transmuter (#384); Wirewood Symbiote (#396) | 3 | |
| Legendary-sorcery cast-condition hook | primitive: no "you may cast this only if..." cast-condition hook for legendary sorceries | Urza's Ruinous Blast (#307); Jaya's Immolating Inferno (#383); Primevals' Glorious Rebirth (#387) | 3 | |
| "Can't be activated" activation gate | primitive: neither `ActivateCatalogAbility` nor `ActivateManaAbility` can be told an ability can't be activated (the Null Rod / Stony Silence shape) | Collector Ouphe (#385); Linvala, Keeper of Silence (#386); Cursed Totem (#386) | 3 | |
| Pick-from-exile / pick-from-graveyard prompt with continuation | prompt: a spell-side pick from exile or from a graveyard with a continuation, distinct from the sacrifice/discard pickers | Lord of the Void (#395); Necromantic Selection (#395); Skullwinder (#397) | 3 | |
| Life-change effect-path replacement routing | primitive: `ChangePlayerLifeForEffect` and `applyLifelinkLocked` bypass the `RepEventLife` replacement pipeline that the public `ChangePlayerLife` runs | Bloodletter of Aclazotz (#300); Alhammarret's Archive (#301); Angel of Vitality (#312) | 3 | |
| Permanent-duration continuous-effect registry | primitive: only a turn-scoped continuous-effect registry exists; nothing tracks an indefinite-duration continuous effect | Tree of Perdition (#310); Tree of Redemption (#394); Rise and Shine (#397) | 3 | |
| Post-departure LKI / "exiled with this" record | primitive: no log-walked record of what a permanent exiled or carried before it left, or of counters it held at departure | The Ozolith (#295); Valakut Exploration (#307); Currency Converter (#387) | 3 | |
| Choose-a-player / choose-an-opponent prompt | prompt: no generic "choose a player" or "choose an opponent" prompt for an ability, as opposed to a spell's player-targeting | Gluntch, the Bestower (#391); Skullwinder (#397); Victory Chimes (#399) | 3 | |
| Statics read from a non-battlefield zone | primitive: the layer engine gathers statics from the battlefield only; a static on a graveyard or otherwise off-battlefield card is never applied | Anger (#295); Wonder (#300); Brawn (#310) | 3 | |
| Ability-copy (copying an ability item, not a spell) | primitive: `CopySpellForEffect` copies spells only; copying an activated or triggered ability item is out of scope | Strionic Resonator (#298); Peter Parker's Camera (#395); Gogo, Master of Mimicry (#396) | 3 | |
| Discard cost on a non-activated-ability surface | cost component: a discard-a-card cost on a mana ability or inside an entry replacement; `ManaAbilityCost` has no discard component | Mox Diamond (#295); Skirge Familiar (#389) | 2 | |
| Mana-ability trigger carrying produced colour | primitive: `EventManaAdded` carries no colour, so nothing can trigger off "mana of a particular color was added" | Forsaken Monument (#300); Ultima, Origin of Oblivion (#393) | 2 | |
| Trigger/target clause reading the triggering event's data | primitive: a triggered ability's target clause is static and cannot read the event that fired it (the entering permanent, the dying card's mana value) | Scrap Trawler (#300); Cloudstone Curio (#387) | 2 | |
| "Whenever an opponent activates an ability" watch | event kind: `triggerHarvester.OnEvent` deliberately skips `EventTrigger`, so nothing can watch an opponent's activation | Harsh Mentor (#384); Runic Armasaur (#385) | 2 | |
| Extra turns primitive | primitive: extra turns do not exist | Time Warp (#311); Time Stretch (#390) | 2 | |
| Opponent picks from a revealed set | prompt: an opponent choosing from a set of cards you reveal, with a continuation | Gifts Ungiven (#307); Intuition (#311) | 2 | |
| Sacrifice cost of more than one permanent | cost component: `SacrificeOther` pays exactly one permanent; no shape for "sacrifice two/N other creatures" as a cost | Priest of Forgotten Gods (#305); Kuldotha Forgemaster (#387) | 2 | |
| Resolution-time "choose N of your own permanents" with continuation | prompt: a resolution-time multi-permanent self-choice (sacrifice or exile "any number") with a continuation, distinct from the single-permanent sacrifice prompt | Scapeshift (#304); Bane of Bala Ged (#395) | 2 | |
| Stack-item retarget | primitive: rewriting a stack item's target(s) at resolution | Imp's Mischief (#298); Spellskite (#302) | 2 | |
| Counter-to-hand / counter-to-zone effect surface | primitive: `counterSpellLocked` accepts a destination zone but no `*ForEffect` wrapper exposes it, so a spell can only counter into the graveyard | Reprieve (#298); Narset's Reversal (#299) | 2 | |
| Draw-replacement count | primitive: no draw replacement with a count ("draw two instead"), and no per-draw-step tally for "except the first" | Teferi's Ageless Insight (#298); Alhammarret's Archive (#301) | 2 | |
| Mana-production replacement event | event kind: no replaceable "mana is produced" event for "produces twice as much instead" | Nyxbloom Ancient (#302); Mana Reflection (#309) | 2 | |
| Mana-spent-to-cast record on the stack item | primitive: what mana paid for a spell is not recorded on its stack item | Liberator, Urza's Battlethopter (#304); Vexing Bauble (#304) | 2 | |
| Cost modification for activated abilities / granted alternative costs | primitive: cost modifiers and granted alternative costs apply to spells only; an activated ability never passes through CR 601.2f-style cost modification | Training Grounds (#302); Hunting Velociraptor (#399) | 2 | |
| A non-owner choosing among another player's permanents | prompt: `ResolveSacrificeChoice` requires the chooser to be the controller; nothing lets one player choose among a different player's permanents | Tragic Arrogance (#388); Gluntch, the Bestower (#391) | 2 | |

## Closed seams

- **AddMana primitive** (PR #341, batch 01) - gave a resolving spell a path to add mana to a pool; unblocked Dark Ritual.
- **DelayedTrigger.ControllerTurnOnly / until-end-of-turn statics** (#314) - unblocked Return of the Wildspeaker and Boros Charm in batch 01, and later The Wandering Emperor's minus ability.
- **EventAttack** (S22) - "whenever ~ attacks" is now a real trigger condition.
- **EventBecomesTarget** (S22) - "whenever ~ becomes the target of a spell or ability" is now a real trigger condition.
- **CopySpell** (#419) - a spell-copy primitive; unblocked Dualcaster Mage in batch 05.
- **Attachments / CR 122.10 "modified"** (#379, alongside #380) - Equipment and Auras read through `AttachedTo`; unblocked Kodama of the West Tree in batch 07.
- **Indestructible keyword** (#380).
- **Boardwipe primitives (`DestroyAllMatching` and the simultaneous-exit family)** (#382).
- **Proliferate** (#381).
- **Flashback and other-zone casting via `Spec.CastableZones`** (S29, #411).
- **Legend rule** (#418).
- **SearchLibrary's graveyard destination** - `searchDestZoneLocked` gained a graveyard case inside batch 02's own PR (#351); unblocked Entomb and Buried Alive.
- **Structured targeting and the zone browser** (S20 and S18.5) - "pick a card from a zone" is no longer a blocker; `TargetCardInGraveyard`-style target clauses exist and are answered by clicking the card in the zone browser.
- **Convoke and waterbend on a spell** (S22, `Spec.TapCost`) - the activated-ability version of both is still open; see the cost-component seams above.
- **Draw-step trigger event** (`EventBeginDrawStep`, S22) - exists; Howling Mine (#299) and Kami of the Crescent Moon (#309) were skipped against an older tree and can be written.
- **CR 903.9's command-zone offer on an effect-side bounce** (#539) - closed while batch 36 was in flight; Sanctum of Eternity's test now covers both directions.
