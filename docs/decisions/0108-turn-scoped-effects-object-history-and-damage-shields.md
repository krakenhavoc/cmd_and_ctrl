# ADR 0108 — Turn-scoped effects, object history and the rest of the damage shields

**Status:** Accepted · 2026-10-02 · S51 — Turn-scoped effects, object history, and the rest of the damage shields (tracker [#1908](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1908))
**Owner decisions:** 2026-10-02. All four questions are answered, each with the recommended option (a). See [Owner decisions](#owner-decisions-2026-10-02) at the end. The sections and the Delivery plan below are written as decided; the options not chosen are kept as considered options.
**Issues:** group A, turn-scoped `ScopedEffect` kinds: [#1886](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1886) (exile instead if it would die this turn), [#1887](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1887) (can't be regenerated this turn), [#1890](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1890) (damage doubled or tripled this turn), [#1823](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1823) (exile instead of your graveyard this turn). Group H, per-object history: [#1888](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1888) (echo), [#1882](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1882) (what a creature did during your last turn). Group J, the ADR 0107 §6 follow-ups: [#1904](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1904) (shields against a chosen source that are not one-use), [#1906](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1906) (a static prevention effect with an additional effect), [#1905](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1905) (damage from a chosen source dealt to something else instead), [#1889](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1889) (damage dealt as though its source had wither or infect).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-02. I ran `git fetch --all --prune` and read the `docs/decisions/` file names on all 36 remote heads (`origin/develop`, `origin/main` and 34 chore, docs, feat, fix, repro and wip branches). I also listed every name ever committed on any ref (`git log --all --name-only -- docs/decisions/`) and the files of every open PR. The highest number anywhere is 0107, and the only open PR touching `docs/decisions/` is the 2026-10-02 promotion (#1917), which adds no ADR. This one takes **0108**.
**Amends:** [ADR 0041](0041-game-persistence.md) Decision P8: §1, §2, §3, §4, §7 and §9 add `ScopedEffect` kinds, and this ADR is the single P8 amendment the tracker asks for. It also takes up [ADR 0107](0107-state-triggers-rebound-disturb-and-damage-prevention.md) §6: §7 to §10 are the follow-ups that section left open. A pointer line goes into ADR 0041 P8 and ADR 0107 §6 with the first implementation PR.
**Builds on:** [ADR 0107](0107-state-triggers-rebound-disturb-and-damage-prevention.md) (§5's prevention flag and rules gates, §6's source shield, `choose_source` and follow-up queue), [ADR 0063](0063-durations-and-control.md) (durations), [ADR 0059](0059-turn-machinery.md) Decision 8 (the attack record), [ADR 0056](0056-infect-wither-toxic.md) (damage results), [ADR 0018](0018-triggers-on-the-stack.md) §6 (pay-unless) and [ADR 0033](0033-ai-bot-seat.md) §1 (the legal-move enumerator).

This ADR was written plan-first. No engine code changed with it. The engine and card changes land in the PRs listed under [Delivery](#delivery).

---

## Context

The S51 triage (2026-10-01, at `eccf038b`) grouped ten open seams that are plain data on machinery the engine already has. This ADR re-checks each one on `origin/develop` at `d016badd`, sizes it by the printed cards that need it rather than by its `Waiting` list (ADR 0106 owner decision 6: a PR lands every card its seam unblocks), and designs all ten.

How the numbers were made:

- **Printed.** The Scryfall default-cards dump of 2026-09-24 (`data/scryfall/default-cards.json`), Commander-legal cards only, deduplicated by oracle ID, tokens and art cards removed: 31,830 cards. I counted every card whose oracle text needs the seam, by the keyword field or a regular expression over the text of every face, then read each hit and dropped the false positives. For example, "Echo of the First Murder" is an ability word, and the werewolves' "last turn" is about spells cast, not attacks.
- **Not in the catalog.** Compared against the oracle fixtures in `server/internal/cards/coverage/testdata/oracle/` on `origin/develop` at `d016badd` (3,310 files). Every catalogued card has one.
- **Waiting.** The `Waiting` list of the seam's row in `server/internal/roadmap/registry.go` at the same commit.
- **Alone.** I read each uncatalogued card's full text and asked whether everything except this seam already has a shape in the engine. As in ADR 0107, a ✓ is a card I found every other clause for, a ? is one I was unsure of, and the estimate is the ✓ count plus half the ? count. It is a planning number. Each PR verifies every card again.

The seams doc runs stale, and so did two claims in the issues. Both are corrected below: `TurnTally.Attacks` exists and is flushed at each turn (#1882), and the static follow-up family (#1906) is 32 cards, not 14. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, effective September 25, 2026). Rulings quoted below are the Scryfall rulings for the named card.

### Sizing

| § | Seam | Issue | Printed | Not in catalog | Waiting | Alone (est.) |
|---|---|---|---:|---:|---:|---:|
| 1 | Exile instead if it would die this turn | #1886 | 60 | 59 + 1 front face | 1 | ~54 |
| 2 | Can't be regenerated this turn | #1887 | 18 (4 shared with §1) | 18 | 1 | ~12 more |
| 3 | Damage doubled or tripled this turn | #1890 | 10 | 10 | 3 | ~8 |
| 4 | Exile instead of your graveyard this turn | #1823 | 4 | 4 | 4 | 4 |
| 5 | Echo, and three non-mana cumulative upkeeps | #1888 | 51 + 3 | 54 | 1 | ~49 + 3 |
| 6 | What happened during your last turn | #1882 | 6 | 6 | 2 | ~5 |
| 7 | Prevention shields that are not one-use | #1904 | 13 + ~78 | ~91 | 13 | ~10 + ~55 |
| 8 | A prevention effect with an additional effect | #1906 | 32 + 8 | 40 | 2 | ~29 + ~8 |
| 9 | Damage from a chosen source dealt to something else | #1905 | 14 | 14 | 14 | ~13 |
| 10 | Damage dealt as though its source had wither or infect | #1889 | 2 | 2 | 2 | 2 |

About 295 printed cards and about 250 that need nothing else. Notes on the counts:

- **§1.** Spikefield Hazard // Spikefield Cave is catalogued for its land back face only (`mdfc_lands.go`), so its front face, the spell, is manual today and lands here. Four cards print both §1 and §2: Carbonize, Disintegrate, Scorching Lava and Runesword. That is why §1 and §2 are one PR.
- **§3.** The ten are the doublers and triplers created by a resolving spell or ability: Blind Fury; Desperate Gambit; Goblin Goliath; Impulsive Maneuvers; Insult // Injury; Isengard Unleashed; Jeska, Thrice Reborn; Lightning, Army of One; Overblaze; and Quest for Pure Flame. The 30 static doublers (Furnace of Rath, Gisela and the rest) are ordinary `Spec` replacements and are not this seam.
- **§5.** 50 cards carry the Echo keyword, and Volcano Hellion prints "has echo {X}" without it. Three echo costs are not mana: Deepcavern Imp and Rakdos Headliner ("Discard a card") and Skizzik Surger ("Sacrifice two lands"). Owner decision 3 lands them, and with them the three cumulative upkeeps that print the same payments (Polar Kraken, Vexing Sphinx, Phyrexian Soulgorger).
- **§6.** Three cards ask what one creature did (Giant Turtle, Goblin Rock Sled, Tangle Kelp) and three ask whether a player attacked you (Avenge; O-Kagachi, Vengeful Kami; Weathered Sentinels). Arboria, Marchesa's back face and Premature Burial print "last turn" about other things and are not counted.
- **§7.** The 13 are the registry row's `Waiting` list. The same kind also covers two families the issue does not name: about 42 cards that prevent all the damage one *object* would deal this turn (Maze of Ith, Kor Haven, Dromoka's Command, Fend Off, Safeguard), and about 36 that prevent all damage dealt to a recipient this turn (Indestructible Aura, Shielded Passage, Safe Passage, Endure, Brace for Impact). None of them is catalogued. They need no engine change beyond §7's kind, so ADR 0106 decision 6 puts them in §7's card PR.
- **§8.** The issue lists 14. The full family is 32. The rest are the seven Phantom creatures, the six "prevent that damage and remove that many +1/+1 counters" creatures (Polukranos, Unchained, Protean Hydra, Ugin's Conjurant, Undergrowth Champion, Oathsworn Knight, Unbreathing Horde), Magma Pummeler, Angel of Suffering, Khârn the Betrayer, The Mindskinner and Weeping Angel. Every one is a prevention static whose additional effect CR 615.12 must still run. Owner decision 2 adds the eight existing scoped shields with a follow-up (Test of Faith, Temper, Candles' Glow and the rest).
- **§9.** About 60 more printed cards redirect damage without a chosen source (Pariah, Palisade Giant, the en-Kor creatures, Martyrdom). §9 builds the redirection mechanics they all need, so its PR checks each one and lands those that need nothing more.

---

## Group A — turn-scoped `ScopedEffect` kinds

Each of §1 to §4 is a continuous effect a resolving spell or ability creates (CR 611.2). ADR 0041 P8 already says such an effect is a `ScopedEffect` record, the kind names its reader, and a kind is added with the first card that needs it. These four sections add four kinds, one `AffectedScope` and no new record type.

### 1. Exile instead if it would die this turn (#1886)

#### What exists, what is missing

- `ModExileInsteadOfLeaving` (Whip of Erebos, unearth) replaces *every* battlefield exit, and its duration is indefinite and pinned (`scoped_replacements.go`). "If it would die" replaces only the move to a graveyard.
- `ModExileInsteadOfGraveyard` (Cosmic Intervention) is the right event but the wrong set: `ScopeYourPermanents`, plus a delayed return.
- Nothing pins a battlefield-to-graveyard replacement to one creature until end of turn, and no scope says "creatures".
- `DealDamageEachThenForEffect` reports one total, so an "each creature" spell cannot tell which creatures were dealt damage.

#### The rules

- **CR 700.4:** "dies" means "is put into a graveyard from the battlefield".
- **CR 614.1a:** "instead" makes this a replacement effect. **CR 611.2c:** it modifies no characteristics, so a "creature" set read live ("If a creature would die this turn", Flaying Tendrils) covers creatures that weren't there when it began. A record that names one creature names that object only (**CR 400.7**).
- **CR 514.2:** "this turn" ends at cleanup.
- **The Disintegrate ruling:** "The 'can't regenerate' is an effect of Disintegrate and not an effect of the damage. It works even if the damage is prevented or redirected." Disintegrate says "If it's a creature …". Demonfire says "If a creature dealt damage this way …", and that clause applies only to a creature that was dealt damage.

#### Decision

There is one reasonable design.

1. **The kind.** `exileIfWouldDie`, reader `replacement`, reads nothing. It applies to a `RepEventMove` from the battlefield to a graveyard whose card the record affects. It writes `NewZone = ZoneExile`.
2. **The sets.** The record can be pinned to one object (ScopeNone, with `PinnedTo` the duration, so a flicker ends it), use a new scope `ScopeCreatures` ("If a creature would die this turn": Cry of the Carnarium, Flaying Tendrils, Malicious Malfunction), or use the existing `ScopeOpponentsCreatures` (Malicious Eclipse). Both scopes are read live, as CR 611.2c says.
3. **Two helpers, by the printed wording.** `effects.ExileIfItWouldDieThisTurn(target)` registers unconditionally ("If that creature would die this turn", "If it's a creature"). `effects.ExileIfDealtDamageWouldDie` registers from the damage instruction's continuation only on a creature that took more than 0 damage ("If a creature dealt damage this way", "If a permanent dealt damage this way").
4. **Damage per recipient.** `DealDamageEachThenForEffect` gains a per-target sibling, `DealDamageEachEachThenForEffect`, whose continuation is told each recipient and what it took, for Anger of the Gods, Crush the Weak and the other "each creature" forms. It is the same sequential walk with one more argument.
5. **Commanders.** Nothing special. An exiled commander may go to the command zone by CR 903.9a, through the engine's existing path.
6. **What the table sees.** A chip on the creature: "Exiled if it dies this turn — Lava Coil". For a scoped record, a line on the game banner.

#### Cards

About 54 of 60. The ? cards are Betrayer's Bargain, Bouncer's Beatdown, Brutal Expulsion, Chandra, Awakened Inferno, Cry of the Carnarium ("creature cards … put there from the battlefield this turn"), Draconic Intervention, Faunsbane Troll, Mawloc, Runesword, The War Doctor, Torch the Tower and Wilt in the Heat. Demonfire, the row's `Waiting` card, lands Full.

### 2. A creature that can't be regenerated this turn (#1887)

#### What exists, what is missing

- `DestroyOptions.CantBeRegenerated` rides one destroy instruction to `ReplacementEvent.CantBeRegenerated`, and `regenerationShieldAppliesLocked` (`regeneration.go`) refuses the shield for that event. Nothing marks a creature for the rest of the turn.

#### The rules

- **CR 701.19c:** "Effects that say that a permanent can't be regenerated … cause regeneration shields to not be applied." The shield is not used up.
- **CR 701.19b:** a static ability that regenerates ("If this creature would be destroyed, regenerate it", Clergy of the Holy Nimbus) is a regeneration effect too, so the mark stops it.
- **CR 514.2, 400.7:** it ends at cleanup or when the object leaves.

#### Decision

There is one reasonable design.

1. **The kind.** `cantBeRegenerated`, reader `rule` (ADR 0107 §5's gate kind), reads nothing, pinned to the permanent with `PinnedTo`.
2. **The gate.** `permanentCantBeRegeneratedLocked(card)` reads the destroy instruction's flag and any live record. `regenerationShieldAppliesLocked` reads it, and so does a new helper that every catalog static regeneration uses. A census test fails a catalog replacement that answers a destruction with a regeneration without asking the gate, in the shape of ADR 0107 §5's prevention census.
3. **Helpers.** `effects.CantBeRegeneratedThisTurn(target)` and `effects.CantBeRegeneratedIfDealtDamage` (Incinerate, Flamebreak, Jaya Ballard), the latter built on §1's per-recipient continuation.
4. **Whippoorwill's "When the creature dies this turn, exile the creature"** is an event-delayed trigger (`DelayedTrigger.On`, `Duration`) on that object, which exists.
5. **What the table sees.** A chip: "Can't be regenerated this turn".

#### Cards

About 12 of the 14 that §1 does not already count. The ? cards are Bone Shaman (a granted ability that watches its own damage), Clergy and Knight of the Holy Nimbus ("Only your opponents may activate this ability": `ActivatedAbility.AnyPlayer` exists, but not an opponents-only form) and Runesword. Whippoorwill lands Full.

### 3. Damage doubled or tripled this turn (#1890)

#### What exists, what is missing

- Every doubler in the catalog is a battlefield `Spec` replacement that multiplies `ev.DamageAmount` (Angrath's Marauders, Fiery Emancipation, Twinflame Tyrant).
- A resolving spell cannot create one. The P8 replacement kinds prevent damage or redirect zone moves.

#### The rules

- **CR 614.1a, 616.1, 616.1f:** a multiplier is a replacement, and several are ordered by the affected player. The Insult // Injury ruling: "If you cast two Insults in one turn, damage dealt by sources you control this turn will be multiplied by 4."
- **CR 120.8:** a source that would deal 0 damage deals none, so nothing doubles it.
- **CR 611.2c:** "a source you control" is read as the damage would be dealt, so a source you gained control of after the spell resolved is doubled too.
- **CR 615.8's "next time"** (Desperate Gambit, Impulsive Maneuvers): the next *instance* of damage from that source, however many events it is.

#### Decision

There is one reasonable design. The fields below are the vocabulary of the ten cards.

1. **The kind.** `multiplyDamage`, reader `replacement`, not a prevention or redirection effect. Reads:
   - `Amount`, the multiplier (2 or 3);
   - which damage: `Objects` and `SourceZone` for one source (Overblaze's target, Jeska's creature, Desperate Gambit's chosen source, Impulsive Maneuvers' attacker), or a new `Sources` value, `yours` (Insult, Isengard, Goblin Goliath, Quest for Pure Flame) or `creatures` (Blind Fury), read through the source's last-known information, as Angrath's Marauders' `damageSourceControlledBy` already does;
   - to what: a new `Recipients` value, `opponents` (Goblin Goliath, Jeska), `opponentsAndTheirPermanents` (Isengard Unleashed), `playerAndTheirPermanents` with `Player` (Lightning, Army of One), or `creatures` (Blind Fury); empty is anything;
   - `CombatOnly`; and a new `Next` bool, for "the next time" (spent per instance, exactly as `ModPreventNextFromSource` is spent).
2. **Durations.** `UntilEndOfTurn`, or `UntilYourNextTurn` for Jeska and Lightning, Army of One, both of which exist.
3. **Desperate Gambit.** "Choose a source you control" is `choose_source` narrowed to the chooser's own sources (a controller filter on `ChooseSourcePrompt`). The losing flip's "prevent that damage" is ADR 0107's `PreventNextDamageFromSource{From: …}`, already shipped.
4. **What the table sees.** A banner line: "Your sources deal double damage this turn — Insult".

#### Cards

About 8 of 10. Overblaze (splice onto Arcane is not built) and Jeska, Thrice Reborn (counting commander casts from the command zone) are the ? cards. Insult // Injury, Isengard Unleashed and Desperate Gambit land Full.

### 4. Exile instead of your graveyard this turn (#1823)

#### What exists, what is missing

- `effects.GraveyardBecomesExile` is a battlefield static (Rest in Peace, Leyline of the Void), and it already replaces moves to a graveyard from anywhere, the stack included.
- `ModExileInsteadOfGraveyard` is battlefield-only and permanents-only.
- The other half of each card, "you may play lands and cast spells from your graveyard this turn", is a `ScopeStanding` `CastPermission` over the graveyard with an until-end-of-turn duration. That exists (`cast_permission.go`; Glacierwood Siege plays lands that way).

#### The rules

- **CR 614.1a.** It covers costs and effects alike. The Yawgmoth's Will ruling: "The second ability creates a replacement effect. It applies to both costs and effects", and "It will exile itself since it goes to the graveyard after its effect starts."
- "A card" is a card, so a token going to a graveyard is not exiled.
- **CR 616.1:** with buyback ("If you cast a Buyback spell … You get to choose") or flashback, the affected player orders the effects, through the existing CR 616 prompt.

#### Decision

There is one reasonable design.

1. **The kind.** `exileInsteadOfYourGraveyard`, reader `replacement`, `ScopeGame`, reads `Player`. It applies to a `RepEventMove` of a card (not a token) into `Player`'s graveyard from any zone, and writes `NewZone = ZoneExile`. It is a new kind rather than a scope of `ModExileInsteadOfGraveyard`, because that kind's `Then` and its battlefield-only reading would mean different things under different scopes.
2. **One helper for the pair.** `effects.YawgmothsWillThisTurn()` writes the record and the graveyard permission together, so neither half can ship without the other (the half that would be stronger than printed).

#### Cards

All four: Yawgmoth's Will, Gaea's Will, Magus of the Will and Walk-In Closet // Forgotten Cellar (the Room's door). Their row's `Waiting` list.

### Snapshot impact (group A)

- Four new kinds (`exileIfWouldDie`, `cantBeRegenerated`, `multiplyDamage`, `exileInsteadOfYourGraveyard`), one new scope (`creatures`) and three new `Mod` fields (`Sources`, `Recipients`, `Next`). Each is an on-disk identity, additive inside `scopedEffects` under schema v7 (ADR 0041 P10). An older binary refuses a file naming one with `ErrUnknownEffectKey`, which is the designed rollback case.
- `v7.txt` records the new `Mod` fields with `-update-shape`. Each PR writes one new fixture per kind into `testdata/snapshots/v7/`. The writer never touches an existing file.
- Every record has a non-layer mod, so it takes a `Seq`, and its `ReplacementEffectID` is minted from it (P8). No new closure route: the gather adapts the record as it adapts every other replacement kind.

---

## Group H — per-object history

### 5. Echo (#1888)

#### What exists, what is missing

- `QueueUpkeepPayUnlessForEffect` (`upkeep_pay_unless.go`) is "at the beginning of your upkeep, pay <cost> or <consequence>", and it holds the table in the step until it is answered (#997). `effects.CumulativeUpkeep` is the same shape for CR 702.24.
- The prompt parses a mana cost only (`ParseCost`).
- Nothing records when a permanent came under its controller's control. `SummonedThisTurn` is set on entry and on a control change (`layers.go`), but it is cleared at the controller's turn and holds no time.
- Echo is not in `canonicalKeywords`. Like cumulative upkeep and ward, it carries a cost, so it should not be: it is a `Spec` constructor.

#### The rules

- **CR 702.30a:** "At the beginning of your upkeep, if this permanent came under your control since the beginning of your last upkeep, sacrifice it unless you pay [cost]."
- **CR 603.4:** the intervening "if" is checked when the trigger would trigger and again on resolution.
- **CR 118.12a:** "sacrifice it unless you pay" is a choice to pay, made on resolution.
- **The Volcano Hellion ruling:** "The echo cost you pay is equal to your life total as the echo triggered ability resolves."

#### Decision

There is one reasonable design.

1. **The fact.** `Player.UpkeepsBegun int` goes up by one as that player's upkeep begins, before any "at the beginning of your upkeep" trigger is checked. `Card.ControlledSinceUpkeep int` is set to the controller's `UpkeepsBegun` whenever the permanent comes under that player's control: as it enters, and at each control change in the layer pass's control step.
2. **The condition.** "Came under your control since the beginning of your last upkeep" is `ControlledSinceUpkeep >= UpkeepsBegun - 1`, read at the trigger and again at resolution. It counts *upkeeps*, not turns, so a skipped upkeep is not a "last upkeep", and a permanent that entered in the untap step is charged once, at the upkeep that follows.
3. **The constructor.** `effects.Echo(cost)` builds the upkeep trigger, with the cost as a function of the board at resolution (Volcano Hellion's {X}). Declining sacrifices it (CR 118.12a). Paying emits `EventEchoPaid`, which is what Shah of Naar Isle's "When this creature's echo cost is paid" watches.
4. **Non-mana costs** (owner decision 3). The pay-unless prompt learns two payments besides mana: "discard N cards" and "sacrifice N permanents of a type". Each is a cost paid on resolution (CR 118.12a), chosen by the payer, and an unpayable one is declined. `effects.Echo` and `effects.CumulativeUpkeep` both accept them, so Polar Kraken ("Sacrifice a land"), Vexing Sphinx ("Discard a card") and Phyrexian Soulgorger ("Sacrifice a creature") land in the same PR. The enumerator offers the payer's legal picks through the existing card-set payload.
5. **What the table sees.** An "Echo due" marker on a permanent whose echo will trigger at its controller's next upkeep. It is derived from the two counters and never stored.
6. **The bot.** The heuristic pays when it can afford to and declines otherwise.

#### Snapshot impact

Two additive fields (`Player.UpkeepsBegun`, `Card.ControlledSinceUpkeep`) under v7, recorded with `-update-shape`. The prompt is the existing pay-unless.

#### Cards

About 49 of 51, plus the three cumulative upkeep cards. Volcano Hellion, the row's `Waiting` card, lands Full.

### 6. What happened during your last turn (#1882)

#### What exists, what is missing

- `TurnTally.Attacks` (ADR 0059 Decision 8) records every attack declared this turn, per object (`AttackRecord{Attacker, Epoch, Defender, PhaseID}`), and is read by `TimesAttackedThisTurn` and `ObjectAttackedThisTurn`.
- `resetTurnTallyLocked` replaces the whole tally when a turn begins, so nothing survives into the next turn.

#### The rules

- **CR 502.3:** effects can keep permanents from untapping (Goblin Rock Sled, Tangle Kelp). **CR 508.1c:** "can't attack if" is a restriction (Giant Turtle).
- **CR 508.4:** a creature put onto the battlefield attacking never "attacked", so only declarations count, which is what the tally records.
- **CR 400.7:** a creature that left and came back has no history.
- **The Giant Turtle ruling:** "It only cares if it attacked on *your* last turn, and not your opponent's."

#### Decision

There is one reasonable design.

1. **The record.** `Player.LastTurnAttacks []AttackRecord`. As a turn ends, the ended turn's `Attacks` are copied onto its active player, replacing that player's previous record. An empty list is written too, because "your last turn" is the most recent one you took, extra turns included.
2. **Readers.** `AttackedDuringControllersLastTurn(card)` matches the object (instance and epoch) in its current controller's record. `AttackedYouDuringTheirLastTurn(player, you)` matches any record whose `Defender` is the player `you` (an attack on a planeswalker or battle is not an attack on the player, as `AttackedPlayersThisTurn` already reads it).
3. **The cards.** Goblin Rock Sled and Tangle Kelp are an `UntapStepRestriction`. Giant Turtle is a layer-6 "can't attack" static with a condition. The condition changes only as a turn begins, and `onTurnBeganLocked` already invalidates the layer cache there. Avenge is a cost reduction, O-Kagachi an intervening "if", and Weathered Sentinels an attack-target permission.

#### Snapshot impact

One additive field, `Player.LastTurnAttacks`, recorded with `-update-shape`.

#### Cards

About 5 of 6. Weathered Sentinels ("can attack players who attacked you … as though it didn't have defender", a per-target waiver) is the ? card. Goblin Rock Sled and Giant Turtle land Full.

---

## Group J — the rest of the damage shields

### What ADR 0107 left, and what it built

ADR 0107 §6 shipped `ModPreventNextFromSource`: a source pin (CR 400.7, with `SourceZone` for a permanent spell, CR 609.7a), a protected player, permanent or nothing, the CR 615.9 recheck (`Queries`), a one-use spend marked per event batch (`SpentBatch`), and a follow-up body (`Then`) queued per (record, batch) in `Game.preventionFollowUps` and run once with the instance's total (CR 615.5). §5 shipped `ReplacementEffect.Prevention` and `RedirectsDamage`, the CR 615.12 settle step (`settleUnpreventableLocked`), and `ModDamageCantBeRedirected`. A census test (`TestDamageReplacementsDeclareWhetherTheyPrevent`) already requires a replacement that rewrites `DamageTarget` to declare `RedirectsDamage`. No catalog card declares it yet.

### Shared machinery: the damage instance

Three of the four sections, and owner decisions 1 and 4, turn on the same question: which damage events happen at the same time? The engine opens one event per (source, recipient), and ADR 0107 used the event batch as its stand-in for "at the same time". It admits the batch is too wide: "it deals 2 damage to you. Then it deals 2 damage to you" is one batch, so the next-damage shield prevents both.

Owner decision 4 answers it with a `DamageInstance` ID:

1. **The stamp.** Each damage instruction (one `DealDamage…` call, including the whole of an `…Each…` walk) and each combat damage step takes the next ID from a game counter and stamps it on every damage event it opens. It rides the event and its tail, so a CR 616 pause resumes with the instance it was opened in. The stamp itself is transient, never captured.
2. **The readers.** The ID is the unit for §7's `divide_shield` (decision 1), for §8's one application per recipient or per source (the Phantom Centaur and Nine Lives rulings), for §10's life check (the Phyrexian Unlife ruling), and for "the next time" on every kind (`preventNextFromSource`, and the `Next` field of `multiplyDamage` and `redirectDamage`).
3. **ADR 0107 §6's known limit is fixed.** `preventNextFromSource` is spent per instance rather than per event batch, so "it deals 2 damage to you. Then it deals 2 damage to you" inside one resolution is two instances, and a next-damage shield prevents only the first. The follow-up queue groups by instance as well. The flush points stay where ADR 0107 put them, and the queue also flushes as an instance ends.
4. **What reaches a snapshot.** Only the references: a new `Mod.SpentInstance` on a spent "next time" record, and a new `PreventionFollowUp.Instance`. Both are additive. The counter is not a top-level key. It is cloned with the game, and restore sets it to the largest instance a restored record names, as it does `scopedEffectSeq`. `SpentBatch` and `Batch` stay readable, so a file written before reads as it did. The PR checks that an older v7 binary refuses a file carrying either new field, as ADR 0107 PR 3 checked for stack pins.
5. **Where it lands.** In its own PR, PR 0, before every PR that reads it (see Delivery).

### 7. Prevention shields that are not one-use (#1904)

#### What is missing

- Every scoped prevention kind is one-use against a source (`preventNextFromSource`), charged against a recipient (`preventDamage`), or combat-only (`preventCombatDamage`). Nothing prevents *all* damage from a source this turn ("Prevent all damage a source of your choice would deal this turn", Pay No Heed; "Prevent all combat damage that would be dealt by target creature this turn", Maze of Ith), *all* damage to a recipient this turn (Indestructible Aura), or the next N damage from a source (Healing Grace).
- Nothing halves damage (Dark Sphere's "prevent half that damage, rounded down").

#### The rules

- **CR 615.1a, 615.7:** a charged shield prevents each 1 damage until it is reduced to 0. "If damage would be dealt to the shielded permanent or player by two or more applicable sources at the same time, the player or the controller of the permanent chooses which damage the shield prevents." Owner decision 1 makes that a prompt.
- **CR 615.9, 609.7b:** a property shield rechecks its source's properties as the damage would be dealt.
- **CR 107.1a:** a fractional result is rounded the way the card says.
- **CR 615.13:** "Some triggered abilities trigger when damage that would be dealt is prevented. Such an ability triggers each time a prevention effect is applied to one or more simultaneous damage events." That is Samite Ministration's "Whenever damage from a black or red source is prevented this way this turn".

#### Decision

1. **The kind.** `preventFromSource`, reader `replacement`, declares `Prevention`. It reads the same source and protection fields as `preventNextFromSource` (`Objects`, `SourceZone`, `Queries`, `Player`, `Types`, or a pinned protected permanent), plus `CombatOnly`, `Then` and `Amount`. `Amount` 0 is "all damage this turn": the shield is never spent. `Amount` N is CR 615.7's charge, replaced copy-on-write as it is spent, as `preventDamage` is (P8's one record that changes).
2. **The source can be chosen, targeted, or not named at all.** A chosen source comes from `choose_source` as it does today. A targeted source is pinned at resolution with no prompt (Maze of Ith, Fend Off, Kor Haven). A shield with no source and only a protected recipient is "prevent all damage that would be dealt to X this turn". `Queries` alone is "damage black sources and red sources would deal" (Prismatic Strands). "Prevent all combat damage that would be dealt to and dealt by that creature" is two mods on one record.
3. **Overlap with `preventDamage`.** A charged shield with no source stays `preventDamage`. Register refuses a `preventFromSource` with an `Amount` and no source, so the two kinds never describe the same shield.
4. **Dark Sphere.** A new `Mod.Half` on `preventNextFromSource`: it prevents half the instance's damage, rounded down as printed (CR 107.1a). It is still spent by that instance.
5. **Samite Ministration** is a follow-up `Then` that, when the prevented source was black or red, puts a triggered item on the stack with the amount. That is New Way Forward's reflexive-trigger shape, already shipped. It runs once per application to simultaneous events, as CR 615.13 says, because the follow-up queue already groups by instance.
6. **The split choice** (owner decision 1). When one damage instance would meet a charged shield across two or more events, and the charge is less than their total, the protected player divides the charge among those events before any of them is applied. It covers every charged shield: the existing `preventDamage` (Mending Hands), `preventFromSource` with an `Amount`, and §9's charged redirection (Harm's Way, Shining Shoal). This is a new pending choice, `divide_shield`, on the existing distribution payload. When the charge covers the total, nothing is asked. The heuristic shields the player first, then the creature closest to lethal. The choice is made before the events are applied, which needs the instance's events known up front, so it builds on PR 0.

#### Cards

- **The 13 members.** About 10: Auriok Replica, Burrenton Forge-Tender, Consulate Surveillance, Dark Sphere, Healing Grace, Pay No Heed, Prahv, Spires of Order, Rith's Charm, Samite Ministration and Shieldmage Advocate. The ? cards are Mourner's Shield ("shares a color with the exiled card"), Protective Sphere (the colours of the mana spent on its activation, which `StackItem.Paid` may carry) and Refraction Trap (its alternative cost).
- **The families.** About 78 more print a shape the kind covers, and none is catalogued. I estimate about 55 need nothing else. Channel Harm, Comeuppance and Judgment of Alexander ("by sources you don't control") need a controller filter that `PermanentQuery` does not have. They land here only if that filter turns out to be one field.

### 8. A static prevention effect with an additional effect (#1906)

#### What is missing

- A `Spec` replacement prevents by cancelling or reducing the event in its `Replace`. Under unpreventable damage, `settleUnpreventableLocked` skips the `Replace`, so an additional effect written inside it (Nine Lives's counter, a Phantom's counter removal) is skipped too. CR 615.12 says it must still happen.
- The follow-up queue is keyed on a `ScopedEffect`'s `Seq`, so a static has no way to queue one.

#### The rules

- **CR 615.5:** the additional effect happens immediately after the prevention, and may count what was prevented.
- **CR 615.12:** under unpreventable damage, a prevention effect is still applied, prevents nothing, and its additional effects still happen.
- **The rulings fix the unit**, and they disagree by wording:
  - Phantom Centaur ("If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter"): "If this creature would be dealt damage from multiple sources at the same time … all of the damage is prevented and only one +1/+1 counter is removed." It also says: "If damage that would be dealt to this creature can't be prevented, the damage is dealt and a +1/+1 counter is removed."
  - Nine Lives ("If a source would deal damage to you, prevent that damage and put an incarnation counter"): "If more than one source deals damage to you at once, prevent the damage from each of them and put that many incarnation counters on Nine Lives." Also: "If damage that a source would deal to you can't be prevented, you still put an incarnation counter on Nine Lives."

#### Decision

1. **The declaration.** `ReplacementEffect.Then string` (a registered body key, never a closure) and `ReplacementEffect.ThenPer`, which is `recipient` or `source`. `Then` requires `Prevention`, and Register refuses it otherwise.
2. **The unit, by the printed wording.** "If damage would be dealt to X" is one application per recipient per instance: `ThenPer: recipient` (the Phantoms, Vigor, Panther Habit). "If a source would deal damage to X" is one per source per instance: `ThenPer: source` (Nine Lives, Swans of Bryn Argoll). A card file picks the one its text says. The helper takes the printed subject, so the choice can't be forgotten.
3. **The amount.** The apply loop records the event's amount before the `Replace` and after it (all of it when the event is cancelled), and queues the difference as the follow-up's `Prevented`. The `Replace` itself only prevents. A census test fails a `Prevention` static whose `Replace` changes anything but the event.
4. **CR 615.12.** `preventionAppliedToUnpreventableLocked` queues a static's follow-up with zero, exactly as it does a scoped shield's.
5. **The queue.** `PreventionFollowUp` gains the static's identity: the source object (instance and epoch), the replacement's slot, and the per-unit key (recipient or damage source). This is additive. `Seq` stays zero for a static.
6. **The existing scoped shields** (owner decision 2). `preventDamage` and `preventCombatDamage` read `Then` as well, through the same queue and the same instance key as the scoped `preventNextFromSource`. Every scoped prevention kind then carries its CR 615.5 additional effect the same way, and its CR 615.12 zero-prevented run too.

#### Snapshot impact

Additive fields on `PreventionFollowUp`, recorded with `-update-shape`. `Then` and `ThenPer` are catalog data, not captured.

#### Cards

About 29 of 32. The ? cards are Hostility, Purity and Vigor (the Incarnations' "shuffle it into its owner's library"), Khârn the Betrayer (a control change as an additional effect), Weeping Angel and Jared Carthalion. Immortal Coil and Nine Lives, the row's `Waiting` cards, land Full. Owner decision 2 adds about 8 more: Acolyte's Reward, Candles' Glow, Inkshield, Sacred Boon, Scars of the Veteran, Temper, Test of Faith and Vengeful Archon.

### 9. Damage from a chosen source dealt to something else (#1905)

#### What is missing

- No replacement can change a damage event's recipient in a way the engine honours. The damage tail's kind is about the target (`damageTailPlayer`, `damageTailPermanent`, `damage_tail.go`) and is fixed when the event is opened, so a `Replace` that rewrites `DamageTarget` from a player to a creature would land through the player path.
- No scoped kind redirects damage.

#### The rules

- **CR 614.9:** a redirection replaces damage dealt to one recipient with the same damage dealt to another. "If one of those permanents is no longer on the battlefield when the damage would be redirected, or is no longer a battle, creature, or planeswalker … the effect does nothing. If damage would be redirected to or from a player who has left the game, the effect does nothing."
- **CR 614.5, 616.1f, 616.2:** the modified event can meet effects that now apply (the new recipient's protection, a shield on it), and the redirection does not apply to its own result again.
- **CR 120.4b:** results are processed after replacements and prevention, so the redirected damage keeps its source's deathtouch, lifelink, infect and wither and, in combat, stays combat damage. A commander's redirected combat damage counts toward the new recipient's CR 903.10a total.
- **The Harm's Way ruling:** "If the chosen source would simultaneously deal damage to multiple permanents you control … Harm's Way will redirect just 2 of that damage. … You choose which 2 damage is redirected." Owner decision 1's `divide_shield` covers it.

#### Decision

1. **The primitive.** `redirectDamageEventLocked(ev, to)` rewrites `DamageTarget` and the tail's kind together. It does nothing if `to` fails CR 614.9's test at that moment. Every redirection uses it, scoped or static. The census test already requires `RedirectsDamage`. It now also fails a `Replace` that writes `DamageTarget` by hand.
2. **The kind.** `redirectDamage`, reader `replacement`, declares `RedirectsDamage`, so `ModDamageCantBeRedirected` and Lava Burst's mark stop it (ADR 0107 §5). It reads:
   - the source fields of §7 (`Objects`, `SourceZone`, `Queries`), optional for the en-Kor shields;
   - the protected recipient (`Player`, `Types`, or a pinned permanent), optional (Opal-Eye: "would deal damage this turn");
   - how long it lasts: `Next` (one instance), `Amount` (the next N, charged), or neither (all this turn: Kor Chant, Oracle's Attendants);
   - where to: a new `Mod.To` (one `ObjectRef`, epoch-pinned, so a creature that left is gone, CR 614.9), or a new `Mod.ToSourceController` bool (Reflect Damage, Aegis of Honor), read through the source's last-known information as the damage would be dealt;
   - `Then`, for Eye for an Eye.
3. **Eye for an Eye** ("instead that source deals that much damage to you and Eye for an Eye deals that much damage to that source's controller") is a `redirectDamage` record with no `To`. The event is unchanged and `Then` runs with the amount. It does not declare `RedirectsDamage`, because nothing is dealt instead to another player.
4. **Statics.** Pariah's "All damage that would be dealt to you is dealt to enchanted creature instead" is a `Spec` replacement whose `Replace` calls the primitive. That is all the other redirection cards need from this PR.

#### Cards

About 13 of 14. Nova Pentacle ("target creature of an opponent's choice") and Shaman en-Kor are the ? cards. Harm's Way and Shining Shoal land with `divide_shield` (owner decision 1). The PR checks the other ~60 redirection cards and lands those that need only the primitive.

### 10. Damage dealt as though its source had wither or infect (#1889)

#### What is missing

The damage results (`DamageResultSource{Infect, Wither, ToxicTotal}`) are snapshotted off the source onto the tail (`damageTail.result`) and read as the damage lands. Nothing else contributes to them.

#### The rules

- **CR 120.3b, 120.3d, 702.80a, 702.90b, 702.90c:** infect makes damage to a player poison and damage to a creature -1/-1 counters, and wither makes damage to a creature -1/-1 counters, from the source's controller.
- **CR 613.11:** "dealt as though its source had" is a rules effect, read as the damage is processed into its results (CR 120.4c). It gives the source no ability.
- **The Phyrexian Unlife ruling:** it "won't affect damage that reduces your life total from a positive number to 0 or less … The next time you're dealt damage, it will be dealt as though its source had infect." So the life total is read once for damage dealt at the same time.

#### Decision

1. **The gate.** `damageResultAsThoughLocked(ev)` ORs `Wither` or `Infect` into the tail's result as the damage lands. It asks the battlefield statics `Spec.DamageAsThough{Wither, Infect, ToYou, WhileAtOrBelowZeroLife}`, dropped when `CatalogAbilityKey` answers empty.
2. **Unlife's condition** is read once per damage instance, against the player's life total as the instance began (owner decision 4), so simultaneous damage that takes you from a positive total to 0 or less is all life loss, as the ruling says.
3. **Nothing else changes.** A "whenever a creature with infect deals damage" trigger does not see it, because the source still lacks the ability.

#### Cards

Both: Everlasting Torment and Phyrexian Unlife. Their other lines shipped with ADR 0107 PR 6 and ADR 0057.

### Snapshot impact (group J)

- Two new kinds (`preventFromSource`, `redirectDamage`) and four new `Mod` fields (`Half`, `Next` if §3 has not landed, `To`, `ToSourceController`), on-disk identities under v7, refused by older binaries with `ErrUnknownEffectKey`. One fixture per new kind and per charged split goes into `v7/`.
- `PreventionFollowUp` gains additive fields (§8).
- `divide_shield` (owner decision 1) is an ordinary pending choice.
- The `DamageInstance` stamp is transient. Only `Mod.SpentInstance` and `PreventionFollowUp.Instance` name an instance, and restore derives the counter from them (Shared machinery, owner decision 4).
- No new closure route. `ReplacementEffect.Then` is a string, and `redirectDamageEventLocked` is engine code, not state.

---

## Shared machinery

- **One source vocabulary.** `Objects` + `SourceZone` + `Queries` mean the same on `preventNextFromSource`, `preventFromSource`, `redirectDamage` and `multiplyDamage`, and `damageFromChosenSourceLocked` (CR 400.7, 609.7a) is the one reading of them. So a permanent spell chosen as a source covers the permanent it becomes, on every kind.
- **One spend rule.** "The next time" is spent per damage instance on every kind (`SpentInstance`). A charge is copy-on-write on every kind, and divided by `divide_shield` when it meets several events of one instance.
- **The enumerator.** Every new prompt (`divide_shield`, and §3's narrowed `choose_source`) is a pending choice with the shared payloads, so `internal/legal` and the client's card grid and distribution UI apply. ADR 0033 §1 holds.
- **The client.** Every new record gets a chip on the object it pins or a line on the game banner. The ADR 0107 §5 banner line is the model.

## Delivery

Each PR lands its engine change and its cards together, test first. Each flips its registry row to implemented (or partial), adds a closed-seam fragment under `docs/engine-seams/closed/`, moves any card that needs more onto a `Waiting` list with the reason, and adds a registry row for any untracked family it finds. Every PR verifies every card against its full oracle text (ADR 0106 decision 6).

| PR | Engine | Cards (est.) | Needs |
|---|---|---:|---|
| 0 | The damage instance: the stamp, the counter, `Mod.SpentInstance`, `PreventionFollowUp.Instance`; `preventNextFromSource` and the follow-up queue move onto it (ADR 0107 §6's known limit) | none new; the 0107 §6 shield cards are re-tested | — |
| 1 | §1 + §2: `exileIfWouldDie`, `cantBeRegenerated`, `ScopeCreatures`, the per-recipient damage continuation, the regeneration gate and its census | ~66 (Demonfire, Whippoorwill, Lava Coil, Disintegrate, Incinerate, Anger of the Gods …) | — |
| 2 | §3: `multiplyDamage`, `Sources`, `Recipients`, `Next`, the controller filter on `choose_source` | ~8 (Insult // Injury, Isengard Unleashed, Desperate Gambit …) | PR 0 (`Next`) |
| 3 | §4: `exileInsteadOfYourGraveyard` and the paired helper | 4 (Yawgmoth's Will, Gaea's Will, Magus of the Will, Walk-In Closet // Forgotten Cellar) | — |
| 4 | §5: `UpkeepsBegun`, `ControlledSinceUpkeep`, `effects.Echo`, `EventEchoPaid`, the marker; the pay-unless prompt's "discard N" and "sacrifice N permanents of a type" payments (owner decision 3) | ~49 echo cards, plus Polar Kraken, Vexing Sphinx and Phyrexian Soulgorger | — |
| 5 | §6: `Player.LastTurnAttacks` and its two readers | ~5 (Goblin Rock Sled, Giant Turtle, Tangle Kelp, Avenge, O-Kagachi) | — |
| 6 | §7 engine: `preventFromSource`, `Half`, and `divide_shield` for every charged shield, `preventDamage` included (owner decision 1) | ~10 (the 13 members) | PR 0 |
| 7 | §7 cards: the object-source and recipient families | ~55 (Maze of Ith, Kor Haven, Dromoka's Command, Indestructible Aura …) | PR 6 |
| 8 | §8: `ReplacementEffect.Then` / `ThenPer`, the queue's static key, the apply-loop amount and its census; `Then` on `preventDamage` and `preventCombatDamage` (owner decision 2) | ~29 statics (Immortal Coil, Nine Lives, the Phantoms, Polukranos …), plus ~8 scoped shields (Test of Faith, Temper, Candles' Glow …) | PR 0, PR 6 |
| 9 | §9: `redirectDamageEventLocked`, `redirectDamage`, `To`, `ToSourceController` | ~13 of the 14, plus the other redirection cards that need nothing else | PR 0, PR 6 (`divide_shield`) |
| 10 | §10: `Spec.DamageAsThough` and its gate | 2 (Everlasting Torment, Phyrexian Unlife) | PR 0 |

PR 0 introduces the damage instance and goes first: it is small, it lands no cards of its own, and PRs 2, 6, 8, 9 and 10 read it. PRs 1, 3, 4 and 5 do not need it and can start at once, in parallel with PR 0. PR 2 starts after PR 0, or starts at once and rebases onto it before review. PRs 1, 2 and 3 each add cases to the same switches in `scoped_replacements.go` and `scoped_effects.go`, so whichever merges second rebases, but that conflict is mechanical. In group J, PR 6 goes next, as the tracker orders. It is the first PR that reads the instance and the one that adds `divide_shield`. PR 7 follows PR 6. PRs 8 and 9 follow PR 6, and PR 10 needs only PR 0; the three can go in parallel with each other. PR 8 and PR 9 both touch the apply loop in `unpreventable_damage.go`, so the second rebases. The tracker's order (#1904, #1906, #1905, #1889) is the merge order if they conflict.

## Consequences

- A resolving spell can now create any of the replacements the catalog prints as statics: exile instead of dying, exile instead of a graveyard, a multiplier, a source-keyed shield and a redirection. All of them are records with a duration, so a table holding one is a restore point.
- A redirection is honoured all the way through the damage tail, and one function does it.
- A prevention static can have an additional effect, and CR 615.12 runs it.
- The engine remembers two facts across a turn boundary: when a permanent came under its controller's control, measured in upkeeps, and what each player attacked with during their last turn.
- "At the same time" for damage is the damage instance, not the event batch, and ADR 0107 §6's known limit is gone.
- A charged shield or redirection that meets several simultaneous events is divided by the player who is protected, as CR 615.7 says, on every charged kind.

## Out of scope

- The 30 static damage multipliers (§3 note), already ordinary `Spec` replacements.
- "Only your opponents may activate this ability" (Clergy and Knight of the Holy Nimbus), splice onto Arcane (Overblaze), an opponent choosing a target (Nova Pentacle), and the "by sources you don't control" controller filter, unless it proves to be one field (§7).
- The werewolves' and Arboria's "last turn", which are about spells and permanents, not attacks.
- Gisela's static "prevent half that damage, rounded up" (a static, and not one of the members).

---

## Questions for the owner (answered)

These are the questions as asked. The owner's answers follow.

1. **A charged shield that meets several damage events at once (§7, §9; CR 615.7, the Harm's Way ruling).** "Prevent the next 3 damage" and "the next 2 damage … is dealt to any target instead" can meet damage to several recipients, or from several sources, at the same time.
   - **(a) Recommended:** the protected player divides the charge among those events, in a new `divide_shield` prompt asked before any of them is applied, and only when the charge is less than their total. It applies to every charged shield, including the existing `preventDamage` (Mending Hands). This is what CR 615.7 says ("the player or the controller of the permanent chooses which damage the shield prevents") and what the Harm's Way ruling says ("You choose which 2 damage is redirected").
   - (b) Spend the charge in the engine's event order, as `preventDamage` does today. It needs no prompt, and it takes a choice the rules give the player.
2. **"Prevented this way" on the existing scoped shields (§8).** Test of Faith, Temper, Candles' Glow, Sacred Boon, Scars of the Veteran, Acolyte's Reward, Inkshield and Vengeful Archon are `preventDamage` or `preventCombatDamage` shields with a follow-up.
   - **(a) Recommended:** in PR 8, those two kinds read `Then` too, through the same queue. About 8 more cards land, and every scoped prevention kind then carries its CR 615.5 additional effect the same way.
   - (b) Only the new kinds and the statics get `Then`. The eight wait on a new row.
3. **Echo costs that aren't mana (§5).** Deepcavern Imp and Rakdos Headliner print "Echo—Discard a card", and Skizzik Surger "Echo—Sacrifice two lands". The pay-unless prompt parses mana only.
   - **(a) Recommended:** PR 4 teaches the prompt "discard N cards" and "sacrifice N permanents of a type" as payments. The three echo cards land, and so do the cumulative upkeep cards that print the same payments (Polar Kraken, Vexing Sphinx, and Phyrexian Soulgorger's "sacrifice a creature"). Each payment is a cost paid on resolution (CR 118.12a), so this matches the rules.
   - (b) Mana echo only. The three cards wait on the echo row.
4. **What "at the same time" means for damage (Shared machinery, §8, §10).** The engine has no unit narrower than the event batch, and a whole resolution is one batch.
   - **(a) Recommended:** each damage instruction and each combat damage step stamps a transient `DamageInstance` ID on the events it opens. That ID is the unit for Q1's split, for §8's one-counter-per-application rule (the Phantom Centaur ruling), and for Phyrexian Unlife reading your life once per instance (its ruling). It also fixes ADR 0107 §6's known limit, where two separate instances inside one resolution spend one next-damage shield.
   - (b) Keep the event batch. It is less work, and two separate damage instructions in one resolution stay one instance: a Phantom loses one counter for both, a next-damage shield prevents both, and Phyrexian Unlife reads your life once for both.

---

## Owner decisions, 2026-10-02

The owner answered the four questions on 2026-10-02. Every answer was the recommended option (a).

1. **A charged shield that meets several damage events at once.** The protected player divides the charge through a new `divide_shield` prompt, asked only when the charge is less than the total of the events it meets. It applies to every charged shield, including the existing `preventDamage`, `preventFromSource` with an `Amount`, and §9's charged redirection (§7 decision 6; Delivery PR 6).
2. **"Prevented this way" on the existing scoped shields.** `preventDamage` and `preventCombatDamage` read `Then` in Delivery PR 8, and the eight existing shields with a follow-up land: Acolyte's Reward, Candles' Glow, Inkshield, Sacred Boon, Scars of the Veteran, Temper, Test of Faith and Vengeful Archon (§8 decision 6).
3. **Non-mana echo costs.** Delivery PR 4 teaches the pay-unless prompt "discard N cards" and "sacrifice N permanents of a type" as payments. Deepcavern Imp, Rakdos Headliner and Skizzik Surger land, and so do the cumulative upkeep cards with the same payments: Polar Kraken, Vexing Sphinx and Phyrexian Soulgorger (§5 decision 4).
4. **What "at the same time" means for damage.** Each damage instruction and each combat damage step stamps a transient `DamageInstance` ID on the events it opens. It is the unit for decision 1's split, §8's follow-up granularity, §10's life check and every "next time" spend, and it fixes ADR 0107 §6's known limit. It lands in Delivery PR 0, before every PR that reads it (Shared machinery, the damage instance).

---

## Amendment 2026-10-05 — §7: shields against sources described by what they aren't, their power, combat status or controller (#2026)

Delivery PR 7b left 28 shields whose sources a `PermanentQuery` list can't describe: "non-Spider creatures", "nongreen creatures", "creatures other than target creature", "creatures with power 3 or less", "attacking creatures", "unblocked creatures", "colorless sources", "creatures with no +1/+1 counters on them", "without trample", "sources you don't control", "creatures your opponents control", "creatures target opponent controls", and Benalish Missionary's "target blocked creature". This amendment is how §7's `preventFromSource` names them.

### The rules

- **CR 609.7b, 615.9:** a shield against sources with certain properties rechecks the properties when the source would deal damage. If they no longer match, the damage isn't prevented and the shield isn't used up.
- **CR 611.2c** fixes the affected set of a resolved continuous effect only when the effect modifies characteristics or changes control. A prevention effect does neither. So "creatures with power 3 or less" is every creature that has power 3 or less as it would deal damage, including one that entered after the spell resolved. Every card's ruling agrees (Hindervines, Vine Snare, Tanglesap, Moonmist, Hunter's Ambush, Lithomancer's Focus, Haze Frog, Inspire Awe, Encircling Fissure).
- **CR 608.2h:** a source that has left the battlefield is read as it last existed there. Heavy Fog's ruling applies it to combat status: a departed creature's noncombat damage is prevented "if it was an attacking creature at the time it left". Frontline Strategist's applies it to creature types.
- **CR 509.1h:** an attacking creature becomes blocked or unblocked as blockers are declared, and stays so until it is removed from combat or the combat phase ends. "A creature remains blocked even if all the creatures blocking it are removed from combat."
- **CR 105.2c:** a colorless object has no color.
- **CR 603.2c:** an ability triggers once each time its trigger event occurs, and repeatedly when one event contains several occurrences.

### Decision

1. **The shape.** One plain-data `game.DamageSourceFilter`, in `Mod.SourceFilter` (at most one entry), beside the record's `Queries`. Only `preventFromSource` reads it. It is refused on every other kind, beside a pinned source, and with `AndDealtBy`. Its fields:
   - `Except`: queries the source must match none of. "Non-Spider creatures" is `Queries {creature}` plus `Except {Subtypes: Spider}`. A changeling is every creature type (CR 702.73a), so it is a Spider. `Except` may ask `Enchanted`, which is read off the battlefield (Inspire Awe).
   - `ExceptObjects`: objects pinned as the shield resolves (CR 400.7). These are the only sets a card locks in itself: Haze Frog's "other creatures" pins the Frog, and Terrifying Presence's "other than target creature" pins the target. The same card as a new object is "other".
   - `PowerBounded` and `PowerAtMost`: "with power N or less", with counters included.
   - `Colorless`, and `NoCounters` (counter kinds the source must have none of).
   - `Combat`: `attacking` or `unblocked`. A closed vocabulary.
   - `Controller`: `notYou`, `opponents`, or `player` with `ControllerPlayer`. A closed vocabulary, relative to the record's controller. `notYou` and `opponents` are the same test in this free-for-all engine. They are separate words because the cards print separate words.
2. **When it is read.** Every time the shield meets a damage event, never as it resolves, off one view of the source (`damageSourceViewLocked`):
   - **On the battlefield:** its characteristics as the event was opened (`ReplacementEvent.SourceLKI`, the reading every property shield already shares). Its own facts are read now: power with counters, counters, attacking, unblocked (`UnblockedAttackerForEffect`), enchanted, and which object it is.
   - **Left the battlefield this turn, and not since cast:** the permanent as it last existed there, controller and characteristics included (CR 608.2h). `PermanentInfo` gained `Attacking` and `Enchanted` for this.
   - **Anything else** (a spell, an emblem, a card in another zone): its characteristics only. It is not attacking, has no counters and is not enchanted.

   A filtered shield reads its `Queries` off the same view, so "attacking creatures" asks both halves of one source. A source the engine can't see matches nothing, the weaker direction.
3. **Blocked.** `game.BlockedAttackerForEffect` is CR 509.1h's blocked reader, beside `UnblockedAttackerForEffect`. It needs an attacking creature, a completed declaration, and the blocked record (or a staged block). `effects.BlockedCreature` targets with it (Benalish Missionary).
4. **A follow-up per source.** "If damage from a creature source is prevented this way, Comeuppance deals that much damage to that creature" and Judgment of Alexander's "Whenever damage from a creature is prevented this way … to that creature" each name one creature. The follow-up queue grouped a scoped shield's applications per record and instance, so three attackers stopped in one combat damage step were one entry. `Mod.ThenPer` (`ThenPerSource`, the static's unit from §8) now keys a source shield's entry on the damage source too: one application per source, which is CR 603.2c's one trigger per occurrence. It is refused on every other kind and without `Then`. The follow-up also carries the source's card types as it would have dealt the damage (`PreventionFollowUp.SourceTypes`, handed to the body as `Event.LastKnownTypes`).
5. **Undergrowth's additional cost** is a plain optional additional cost (`effects.OptionalAdditionalMana`, CR 118.8b), not a kicker, so nothing that asks "was it kicked" sees it.

### Snapshot impact

- New `Mod` keys `sourceFilter` and `thenPer`. A binary from before this amendment refuses a file carrying either (an unknown mod key, ADR 0041 P4). It does not drop them, because a shield without its filter would prevent more than printed. The restore check also walks the filter's own keys, its `Except` queries' keys and its pinned objects' keys, so a key a later binary adds is refused too. Its vocabularies (`combat`, `controller`, `thenPer`) are checked as keys.
- `PreventionFollowUp.sourceTypes`, refused by an older binary for a file of the current schema, like every follow-up field.
- `PermanentInfo.attacking` and `.enchanted` on `lastKnownPermanents`. An older binary drops them. Their only readers are filtered shields, which that binary refuses anyway.
- All additive under v7 (`-update-shape`). No fixture changes. No new closure route: the filter is plain data.

### Cards

25 ship Full: Al-abara's Carpet, Arachnogenesis, Benalish Missionary, Comeuppance, Deep Wood, Fog of War, Frontline Strategist, Galadhrim Ambush, Harmless Assault, Haze Frog, Heavy Fog, Hindervines, Hunter's Ambush, Inspire Awe, Judgment of Alexander, Lithomancer's Focus, Moonmist, Obscuring Haze, Repel the Abominable, Tanglesap, Terrifying Presence, Thwart the Enemy, Undergrowth, Vine Snare and Winds of Qal Sisma. Deep Wood and Heavy Fog's "only if you've been attacked this step" reads the turn's attack declarations in the current combat phase.

### What is still out

- **Channel Harm** waits on `prevention-follow-up-choice` (#2040). Its shield is `notYou`, but "you may have Channel Harm deal that much damage" is a choice made as the damage is prevented (CR 615.5). A follow-up can only queue a prompt that is answered after the damage's spell or ability has finished.
- **Encircling Fissure** waits on awaken (#2411, new row `awaken`). Its shield is the `player` controller test.
- **Snag** waits on an alternative cost that discards a card (#2412, new row `discard-alternative-cost`). Its shield is `unblocked`.

---

## Amendment 2026-10-06 — §7: shields for every creature, your creatures, or players (#2045)

Delivery PR 7 left twelve shields whose protected set §7's fields can't describe. `preventFromSource` protects a player (with `Types` for "you and <types> you control", which always protects you too), pinned permanents, or everything. It can't say "to creatures" (Forfend, Blinding Fog), "to creatures you control" without you (Divine Light, Sivvi's Ruse, Loyal Unicorn, Summon: Alexander's chapters I and II, Surge of Salvation), "to players" (Commencement of Festivities, Defend the Hearth, Chameleon Blur), "to Dogs you control" (Pack Leader) or "to artifact creatures" (Ethersworn Shieldmage). This amendment is how it names them.

### The rules

- **CR 615.1:** prevention effects "apply continuously as events happen—they aren't locked in ahead of time".
- **CR 611.2c** fixes the affected set of a resolved continuous effect only when it modifies characteristics or changes control. Otherwise the effect "modifies the rules of the game, so it can affect objects that weren't affected when that continuous effect began". A prevention effect does neither, so "creatures you control" is every creature you control as the damage would be dealt: one that enters later, or that you gain control of, is protected, and one you lose control of is not. Blinding Fog, Surge of Salvation and Loyal Unicorn also grant a keyword, which modifies characteristics, so that half of each card is fixed as it resolves.
- **CR 615.9:** a property shield rechecks the source as it would deal damage. Surge of Salvation's "black and/or red sources" and Chameleon Blur's "creatures" are this recheck, beside the recipient set.
- **CR 615.11:** "the next N damage that would be dealt to each of a number of untargeted creatures" creates one shield per creature as it resolves. That is not one shared charge, so a charged shield refuses a set.

### Decision

1. **The shape.** One plain-data `game.DamageRecipientFilter`, in `Mod.RecipientFilter` (at most one entry). Only `preventFromSource` reads it, on an unpinned record (`ScopeGame`) with no `Player`, no `Types`, no `AndDealtBy` and no charge. Registration and restore refuse it anywhere else. Its fields:
   - `Players`: every player.
   - `Permanents`, narrowed by `Match` and `Controller`.
   - `Match`: queries the permanent must match **every** one of. The engine's other query lists are any-of, but "artifact creatures" is a permanent that is both. "Creatures" is one query, `{creature}`; "Dogs" is `{Subtypes: Dog}`, so a changeling counts.
   - `Controller`: `you`, a closed vocabulary relative to the record's controller.
2. **When it is read.** Every time the shield meets a damage event (`nextFromSourceProtectsLocked`), never as it resolves. A player is matched by identity, and a permanent as it is on the battlefield now, by its effective characteristics and controller. A recipient the engine can't see is not protected, the weaker direction.
3. **Composition.** The set sits beside every source field: a chosen source, `Queries` and #2026's `SourceFilter`. Surge of Salvation and Chameleon Blur are one record each, so each is one prevention effect (CR 614.5's one opportunity, CR 615.13's one application).
4. **The card side.** `effects.ShieldCreatures`, `ShieldCreaturesYouControl`, `ShieldPlayers`, `ShieldPermanents(queries…)` and `ShieldPermanentsYouControl(queries…)` are new `ShieldTarget`s. Only `PreventDamageFromSource` reads them. A next-damage shield, which protects one thing, refuses them.

### Snapshot impact

- New `Mod` key `recipientFilter`. A binary from before this amendment refuses a file carrying it (an unknown mod key, ADR 0041 P4). It does not drop it, because the shield would then protect everything. The restore check also walks the filter's own keys and its `Match` queries' keys, and checks its `controller` vocabulary and the record's scope.
- Additive under v7 (`-update-shape`). No fixture changes. No new closure route: the filter is plain data.

### Cards

All twelve ship Full: Blinding Fog, Chameleon Blur, Commencement of Festivities, Crystal Fragments // Summon: Alexander, Defend the Hearth, Divine Light, Ethersworn Shieldmage, Forfend, Loyal Unicorn, Pack Leader, Sivvi's Ruse and Surge of Salvation. Sivvi's Ruse's free cast is the Nemesis cycle's condition, `effects.AnOpponentControlsAAndYouControlA`. Crystal Fragments flips with ADR 0079's exile-and-return, and Summon: Alexander enters with its first lore counter (CR 714.3a).

### What is still out

Nothing on the row. The seam is closed.

---

## Amendment 2026-10-07 — §7: durations of "this combat", "until end of combat" and "while this Saga remains" (#2027)

Delivery PR 7b left five cards whose shield or pump does not last "this turn": Sewers of Estark, Suppressor Skyguard and Winter's Chill prevent combat damage "this combat", Glyph of Destruction's +10/+0 lasts until end of combat, and Old Fat Spider Can't See Me's chapter II lasts "for as long as this Saga remains on the battlefield". With an additional combat phase "this combat" and "this turn" differ, so none of them could ship as "this turn".

### The rules

- **CR 611.2a, 611.2b:** a continuous effect from a resolved spell or ability lasts as long as it says. An effect whose duration is already over as it would begin never begins.
- **CR 511.3:** as the end of combat step ends, creatures are removed from combat, and effects that last "until end of combat" expire. **CR 724.2d** says the same for an effect that ends the combat phase.
- **CR 506.1** and the turn plan (ADR 0059 Decision 3): a turn can hold more than one combat phase, and each is its own phase instance.

### Decision

1. **A seventh duration kind, `UntilEndOfCombat`.** It names the combat phase INSTANCE it was made in, by the turn's identity and the phase instance (`Duration.CombatTurn` is `Turn.Seq`, `Duration.CombatPhase` is `Turn.PhaseID`), both stamped by `Game.UntilEndOfCombatDuration`. Appended after `UntilYourNextEndStep`, so every stored kind keeps its meaning.
2. **One expiry rule, a pure read of the cursor.** `durationExpiredLocked` says the effect is over when the cursor is not inside that combat phase. There is no flag to clear. `advanceCursorLocked` runs the ordinary sweep after every move out of a combat step, which is the one seam every step transition takes, `EndCombatPhaseForEffect` included (it walks the cursor out step by step), and any later cleanup or turn start would catch a turn that ended early.
3. **No combat, no effect.** `UntilEndOfCombatDuration` returns false outside a combat phase. A card registers nothing then, which is CR 611.2b's reading of an effect that has already ended as it begins.
4. **"While this Saga remains" needed no new duration.** `ForAsLongAs` with `WhileSourceOnBattlefield` (ADR 0063) already says it, keyed on the Saga as the object it is now (CR 400.7). What the damage shield lacked was a way to take either duration: `DamageShield.Duration`, a non-zero `Duration`, replaces the default. It is refused beside `UntilYourNextTurn` and when `Duration.Problem` names anything. On the card side, `effects.PreventDamageFromSource.Lasts` is `ShieldThisTurn` (the default), `ShieldThisCombat` or `ShieldWhileSourceRemains`. `effects.DurationUntilEndOfCombat` builds the duration for a `ScopedEffectFor` (Glyph of Destruction's +10/+0).
5. **Validation.** `Duration.Problem` refuses the two combat fields on any other kind, so a binary that ignores them cannot read a duration differently from the one that wrote it.

### Snapshot impact

- New `Duration` keys `CombatTurn` and `CombatPhase`, omitted when zero, so every stored duration reads the same. Additive under v7 (`-update-shape`); the shape file gains the two keys at every route to a `Duration`.
- The new kind is an integer one past the last. A binary from before this amendment refuses a file carrying it (`Duration.Known`, ADR 0041 P4) rather than restoring an effect that never ends. No closure route: the duration is plain data.
- No fixture changes.

### Cards

Four ship `full`: Glyph of Destruction (a CR 603.7 delayed trigger destroys the Wall as the object it is now, via the new `destroy/the-object` body), Old Fat Spider Can't See Me, Sewers of Estark and Suppressor Skyguard (its "another opponent who isn't being attacked" is an intervening if, checked on trigger and again on resolution, CR 603.4).

### What is still out

- **Winter's Chill** moves to a new row, `x-capped-by-a-count` (#2581). Its "this combat" prevention is expressible now. "X can't be greater than the number of snow lands you control" is not: the engine bounds X only through a cost, so the card would cast for any X, stronger than printed.
