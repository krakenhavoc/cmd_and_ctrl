# ADR 0107 — State triggers, rebound, disturb and damage prevention

**Status:** Accepted · 2026-10-01 · S50 — Seams from the deck re-checks (tracker [#1784](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1784))
**Owner decisions:** 2026-10-01. All six questions are answered, each with the recommended option (a). See [Owner decisions](#owner-decisions-2026-10-01) at the end. The sections and the Delivery plan below are written as decided; the options not chosen are kept as considered options.
**Issues:** [#1858](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1858) (state triggers), [#1854](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1854) (rebound), [#1855](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1855) (disturb), [#1853](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1853) (damage that can't be prevented), [#1860](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1860) (the next damage from a source), [#1879](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1879) (a creature that can't attack unless the defending player controls something, §2) and [#1880](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1880) (players can't gain life, §5 owner decision 5).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-01. I ran `git fetch --all --prune` and read the `docs/decisions/` file names on all 35 remote heads (`origin/develop`, `origin/main` and 33 chore, docs, feat, fix, repro and wip branches). I also listed every name ever committed on any ref (`git log --all --name-only -- docs/decisions/`). The highest number anywhere is 0106, and no open PR carries an ADR. This one takes **0107**.
**Builds on:** [ADR 0106](0106-five-small-seams-from-the-s50-rechecks.md) (§2's attack-target restrictions, §4's one-gate rules effect, and owner decision 6), [ADR 0104](0104-gaining-control-of-a-spell.md) (the stack step of the layer pass), [ADR 0079](0079-transforming-a-permanent.md) (transform), [ADR 0066](0066-granted-cast-and-play-permissions.md) (cast permissions), [ADR 0072](0072-protection.md) (protection's prevention), [ADR 0033](0033-ai-bot-seat.md) §1 (the legal-move enumerator) and [ADR 0041](0041-game-persistence.md) (restore points).

This ADR was written plan-first. The engine and card changes land in the PRs listed under Delivery.

---

## Context

S50 left twenty open seam rows. ADR 0106 owner decision 6 says each PR lands every card its seam unblocks. So a seam is worth the number of printed cards that need it, not the one or two names on its `Waiting` list. This ADR sizes all twenty, picks the six that unblock the most cards for the least engine work, and designs those six.

How the numbers were made:

- **Printed.** I used the Scryfall default-cards dump of 2026-09-24, kept Commander-legal cards only, and deduplicated by oracle ID. I counted every card whose oracle text needs the seam, by keyword field or by a regular expression over the full text of every face. I then read every hit and dropped the false positives (for example, the eleven Licids that "lose this ability" belong to a different seam, and "target opponent discards a card at random" is not a random retarget).
- **Not in the catalog.** I compared against the oracle fixtures in `server/internal/cards/coverage/testdata/oracle/` on `origin/develop` at `9e11d658`. Every catalogued card has one.
- **Alone.** I read each uncatalogued card's full text and asked whether everything except this seam already has a shape in the engine. A ✓ means I found every other clause in code or in an existing catalog card. A ? means I was not sure. The estimate is the ✓ count plus half the ? count. It is a planning number. Each PR verifies every card again (ADR 0106 decision 6).

The seams doc runs stale. Every "missing" claim below was checked in the code at `9e11d658`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`, effective September 25, 2026).

### Sizing

| Seam | Issue | Printed | Not in catalog | Alone (est.) | Engine work | Verdict |
|---|---|---:|---:|---:|---|---|
| State triggers | #1858 | 40 | 40 | ~18 (~31 with §2) | medium | **§1** |
| Can't attack unless defending player controls X | #1879 | 34 | 34 | ~32 (14 shared with §1) | small | **§2** |
| Rebound | #1854 | 38 | 38 | ~36 (with the granters, decision 3) | small (printed), medium (granted) | **§3** |
| Disturb | #1855 | 25 | 25 | ~20 | small | **§4** |
| Damage that can't be prevented | #1853 | 31 | 29 + 2 caveats | ~24 (with "can't gain life", decision 5) | medium | **§5** |
| The next damage from a source | #1860 | 45 | 45 | ~43 (with "prevented this way", decision 4) | medium | **§6** |
| Assist | #1857 | 16 | 15 + 1 caveat | ~15 | medium-large: a second payer mid-cast, and a client prompt | deferred |
| Discarded by an opponent | #1861 | 16 | 16 | ~14 | medium: the cause's controller on the discard, and a redirect through the entry pipeline | deferred |
| More Than Meets the Eye, convert, living metal | #1856 | 15 | 15 | ~4 | large: three mechanics, and every card has more | deferred |
| Targets relative to the source | #1863 | 12 | 12 | ~10 | small | deferred: next cheap batch |
| An ability that reads the card discarded for it | #1862 | 9 | 9 | ~6 | tiny: one line plus last-known information | deferred: next cheap batch |
| Paradigm | #1866 | 5 | 3 + 2 caveats | ~5 | large: casting a copy of a card in exile | deferred |
| Triggers once per counter | #1841 | 4 | 4 | ~4 | small | deferred |
| Auras on a graveyard card | #1780 | 4 | 4 | ~3 | large: attachment off the battlefield | deferred |
| Next creature spell promises | #1852 | 3 | 3 | ~3 | small | deferred |
| Losing one printed ability | #1859 | 2 | 2 | 2 | medium | deferred |
| Random retarget | #1815 | 2 | 2 | ~1 | medium | deferred |
| Target bounded by counters removed | #1842 | 1 | 1 | 1 | small | deferred |
| Unremovable stun counters | #1824 | 1 | 1 | 1 | small | deferred |
| Abilities of a spell on the stack | #1865 | 1 | 1 | 1 | large | deferred |
| Any-player mana abilities | #1864 | 1 | 1 | 1 | medium | deferred |

Notes on the counts:

- **State triggers.** 14 of the 40 also print "can't attack unless defending player controls an Island" (Sea Serpent and the rest). Nothing in the engine has that restriction (`restrictions.go` has only attack taxes). That is why §2 exists.
- **Rebound.** 34 cards print the keyword. Four more grant it to a spell: Cast Through Time, Narset Transcendent, Ojer Pakpatiq, Deepest Epoch and Taigam, Ojutai Master.
- **Can't be prevented.** Banefire and Bonecrusher Giant are catalogued with caveats for this gap, and both drop them. Five of the cards (Skullcrack, Call In a Professional, Sunspine Lynx, Leyline of Punishment and Everlasting Torment) also print "players can't gain life". That is a separate gap: Screaming Nemesis and Sulfuric Vortex carry it as a caveat, and 21 uncatalogued Commander-legal cards print it. Owner decision 5 puts it in §5's PR, tracked on #1880.
- **Next damage from a source.** The registry row names only Mercenaries ("the next time this creature would deal damage to you"). The rule it waits on, CR 615.8, is also the rule behind the whole Circle of Protection family, which chooses its source as the shield is made (CR 609.7a, 615.9). Nine of the 45 also do something with "the damage prevented this way" (CR 615.5). Owner decision 4 puts that in §6.
- **Lose one printed ability** counts only Glittering Lion and Glittering Lynx. **Trigger per counter** counts the four cards with no "only once each turn" limit (Bloodcrazed Hoplite, Fathom Mage, Flourishing Defenses, Sigurd, Jarl of Ravensthorpe). **Target bounded by counters removed** counts only Simic Manipulator: the other four "removed this way" cards bound an amount of damage, not a target.

### Why these six

The six chosen seams cover 199 distinct printed cards, and an estimated 172 of them need nothing else under the owner's decisions. Each one either reuses machinery the engine already has, or adds one gate in the shape ADR 0106 already used:

- §2 is one more field on ADR 0106 §2's `AttackTargetRestriction`.
- §3 and §4 are an exile route like an Adventure's, a delayed trigger, and a cast permission like suspend's or a defeated Siege's.
- §5 is a rules effect read at one gate, as ADR 0106 §4 was.
- §6 is one more scoped prevention mod kind beside the two that exist.

The deferred seams are either small in cards (ten of them unblock five cards or fewer) or large in engine (assist needs a second player to pay in the middle of a cast; More Than Meets the Eye is three mechanics for 15 cards that all print more besides). Targets relative to the source (#1863) and the card discarded to an ability (#1862) are cheap. They are the next batch's natural opening PR, and they are left out only to keep this ADR to six sections.

---

## 1. State triggers (#1858)

### What exists, what is missing

- **Every trigger watches an event.** `TriggeredAbility` is harvested from `EventKind`s (`triggers.go`, `harvestMatchLocked`). Nothing asks a condition over the board on its own.
- **The place to ask exists.** `runStateChecksLocked` (`mutations.go`) is the engine's one pairing of CR 704.3's state-based actions and the CR 603.3 trigger drain. It runs every time a player would receive priority.
- **The latch needs no new state.** CR 603.8's "doesn't trigger again until the ability has resolved, has been countered, or has otherwise left the stack" can be read off the stack and the trigger queue. Both already name each item's source object and its catalog row (`AbilityRef`, ADR 0041 tier 4).
- No card in the catalog prints a state trigger, so nothing has to migrate.

### The rules

- **CR 603.8:** a state trigger triggers "as soon as the game state matches the condition", goes on the stack "at the next available opportunity", and "doesn't trigger again until the ability has resolved, has been countered, or has otherwise left the stack. Then, if the object with the ability is still in the same zone and the game state still matches its trigger condition, the ability will trigger again." Its example has a hand that is empty only for a moment, in the middle of a resolution, and the ability still triggers.
- **CR 603.2:** an ability triggers when a game event *or game state* matches its trigger.
- **CR 603.4:** an intervening "if" is checked when the ability triggers and again on resolution (Veiled Crocodile's "if this permanent is an enchantment").
- **CR 704.3 / 117.5:** state-based actions are checked, then waiting triggers go on the stack, then the check repeats.
- **CR 400.7:** an object that changed zones is a new object, so the latch is per object.

### Options

- **A. A state condition on the trigger, checked after every event and in the CR 704.3 loop (chosen, owner decision 1).** `TriggeredAbility.State` is a condition over the board, the source and its controller. A trigger with a `State` watches no event. The engine asks every battlefield permanent's state triggers after each event it emits and in each pass of `runStateChecksLocked`. That is what makes CR 603.8's own example come out right: a hand that is empty for a moment during a resolution triggers "Whenever you have no cards in hand".
- **B. The same condition, checked only in the CR 704.3 loop.** It is simpler, and it only misses states that come and go within one resolution. None of the 40 cards is likely to notice, but the rule's own example does.
- **C. Approximate each card with event triggers** (watch a land leaving, a counter being removed). This is what the registry row warns against. It triggers once per event instead of once per state, and it misses every way a state can arise that the card file did not think of.

### Decision (A, owner decision 1)

1. **The declaration.** `TriggeredAbility.State func(g *Game, source *Card, controller uuid.UUID) bool`. It is catalog data, rebuilt from the row like `AppliesTo`, so the ADR 0041 closure ratchet gains no route. Register refuses a row that sets both `State` and `Watches`. A state trigger must also declare its `Effect` (ADR 0041 tier 4-final), so its stack item is keyed and a table with one waiting is a restore point.
2. **The check.** `stateTriggersLocked` walks the battlefield's state triggers (an index built in the layer pass, so the common case of no state triggers costs nothing). For each one whose condition holds and that is not latched, it queues an ordinary pending trigger with the permanent as the source. It is called from the event emitter after the zone-walk harvest, and from each pass of `runStateChecksLocked` before the drain.
3. **The latch.** A state trigger is latched while any stack item or pending trigger names the same source object (`ObjectRef`, CR 400.7) and the same `AbilityRef`. It is derived, never stored. When the item leaves the stack, the next check re-reads the condition, as CR 603.8 says.
4. **Ability removal.** A permanent that has lost its abilities (CR 613.1f) has no state triggers, through the same `CatalogAbilityKey` read every other catalog trigger uses.
5. **Targets and choices.** A state trigger is an ordinary triggered ability once it is queued. Deadly Designs' "destroy up to two target creatures" picks targets as it is put on the stack (CR 603.3d), through the existing path.
6. **Shared helpers.** `effects.WhenYouControlNo(query)` ("When you control no Islands, sacrifice this creature") and `effects.WhenThereAreNo(query)` ("When there are no creatures on the battlefield") cover about 29 of the 40 cards. A counter threshold ("When there are five or more plot counters on this enchantment") gets a third helper.

### Snapshot impact

None. The latch is derived from the stack and the trigger queue, which are already in the snapshot. A queued or stacked state trigger is a keyed catalog item.

### Cards

About 18 land with §1 alone: Barbarian Outcast, Covetous Dragon, Dark Depths, Deadly Designs, Emperor Crocodile, Homarid, Last Laugh, Lurebound Scarecrow, Nine Lives, Serendib Djinn, Skeleton Ship, Synod Centurion, Task Mage Assembly, Tethered Griffin and others, each verified in the PR. 13 more need §2 as well, and land with whichever of the two PRs merges second. Giant Shark ("a creature that has been dealt damage this turn"), Immortal Coil ("for each 1 damage prevented this way") and Rune-Tail, Kitsune Ascendant (a flip card) go on the row's `Waiting` list.

---

## 2. A creature that can't attack unless the defending player controls something (#1879)

### What exists, what is missing

- **ADR 0106 §2 built per-target attack restrictions.** `Characteristic.AttackTargetRestrictions []AttackTargetRestriction{Source, SourceName, NotOwner, NotOwnersPlaneswalkers}`, and one predicate, `canAttackTargetWithLocked(attacker, target)`. It is asked by both declaration verbs, by the CR 508.1d search and by the enumerator. The ADR 0105 digest's `attack_targets` and the client's per-attacker gate follow from it.
- **Nothing says "unless defending player controls".** `restrictions.go` lists attack taxes as the only "can't attack unless" shape. 34 Commander-legal cards print this restriction, and none is catalogued.

### The rules

- **CR 508.1c:** restrictions include effects that say a creature "can't attack unless some condition is met". A declaration that disobeys one is illegal.
- **CR 508.5 / 508.5a:** "defending player" means the player the creature is attacking, the controller of the planeswalker it is attacking, or the protector of the battle it is attacking. In multiplayer it is one specific player, determined separately for each attacking creature.

So in Commander, Sea Serpent may attack any opponent who controls an Island, and that opponent's planeswalkers and the battles they protect. It may not attack anyone else. The restriction is per target, exactly the shape ADR 0106 §2 built.

### Options

- **A. One more field on the existing restriction (chosen, owner decision 2).** `AttackTargetRestriction.DefenderMustControl *PermanentQuery`, where `PermanentQuery` is a closed data struct (any of: card types, subtypes, supertypes, colours, a keyword, "enchanted"). The predicate works out the defending player of each target (CR 508.5) and refuses the target if that player controls no matching permanent.
- **B. A restriction bit per land type.** That covers Island, Swamp, Forest and Mountain. It does not cover "a snow land", "a blue permanent", "an enchantment or an enchanted permanent" (Godhunter Octopus) or "a creature with flying" (Lurking Green Dragon), all of which are printed.

### Decision (A, owner decision 2)

1. **The query is data.** The restriction reaches the snapshot through `lastKnownBattlefield`, so it cannot hold a func. `PermanentQuery` is plain fields. Every printed condition in the 34 is one of: a land subtype, "a snow land", a colour, an enchantment or an enchanted permanent, or a creature with flying.
2. **The predicate.** `canAttackTargetWithLocked` gains one branch. It reads the board live at declaration, which is the only time CR 508.1c asks (an Island that leaves after attackers are declared changes nothing).
3. **The view.** The chip from ADR 0106 §2 decision 4 gets a second sentence: "Can't attack Bob: he controls no Island". The greyed target ring needs no change.
4. **Helper.** `effects.CantAttackUnlessDefendingPlayerControls(query)`, as an ordinary layer-6 static, applied only while `CatalogAbilityKey` answers (CR 613.1f).

### Snapshot impact

One additive field inside `lastKnownBattlefield`'s characteristic, recorded with `-update-shape` under the current schema version.

### Cards

About 18 land with §2 alone: Armored Galleon, Deep-Sea Serpent, Dreamwinder, Ethereal Whiskergill, Floodchaser, Godhunter Octopus, Hammerhead Shark, Lurking Green Dragon, Red Cliffs Armada, Sea Monster, Sealock Monster, Serpent of the Endless Sea, Slipstream Eel, Steam Frigate, Vodalian Serpent, Whimwader, Wu Warship and Zhou Yu, Chief Commander. Goblin Rock Sled and Veiled Serpent are checked in the PR. The 14 that also need §1 land with whichever PR merges second.

---

## 3. Rebound (#1854)

### What exists, what is missing

- **Rebound isn't in `canonicalKeywords`.** The deck importer drops it, and a rebound spell goes to the graveyard.
- **Every piece it needs exists somewhere else:**
  - The resolution route `routeStackCardToGraveyardLocked` (`mutations.go`) already replaces "put into its owner's graveyard as it resolves" three ways: flashback (CR 702.34a), buyback (CR 702.27a) and an Adventure's exile (CR 715.3d). The Adventure arm exiles the card and hangs a grant on the route's continuation, which is rebound's shape.
  - `StackItem.CastFromZone` records the zone a spell was cast from.
  - `DelayedTrigger{At: StepUpkeep, ControllerTurnOnly: true}` is "at the beginning of your next upkeep". `delayed.go` already names rebound as its use.
  - Suspend's free cast (`offerSuspendedCastLocked`, `grantSuspendedFreeCastLocked`) is CR 608.2g's "cast it without paying its mana cost" during a resolution: a `{0}`-priced per-card `CastPermission` with flash timing and the CR 107.3b X lock.
- **A spell cannot gain an ability on the stack.** The stack step of the layer pass (ADR 0104, `stackControlPassLocked`) applies layer 2 only, and `stackPinProblem` refuses any other mod on a stack pin. Taigam, Ojer Pakpatiq, Narset and Cast Through Time all give rebound to a spell, which is layer 6 (CR 613.1f).

### The rules

- **CR 702.88a:** rebound "represents a static ability that functions while the spell is on the stack and may create a delayed triggered ability". It means "If this spell was cast from your hand, instead of putting it into your graveyard as it resolves, exile it and, at the beginning of your next upkeep, you may cast this card from exile without paying its mana cost."
- **CR 702.88b:** that cast follows the rules for alternative costs (CR 601.2b, 601.2f–h).
- **CR 702.88c:** multiple instances are redundant.
- **CR 608.2n:** an instant or sorcery is put into its owner's graveyard as the final part of its resolution. Rebound replaces only that, so a countered or fizzled rebound spell goes to the graveyard as usual.
- **CR 603.7a / 603.7c:** the delayed trigger is created on resolution. If the card has left exile by the upkeep, the trigger finds nothing (CR 400.7).
- **CR 608.2g:** a spell cast during a resolution is cast without anyone receiving priority.
- **CR 613.1f:** "ability-adding effects" are layer 6. "That spell gains rebound" and "Instant and sorcery spells you control have rebound" are ability-adding effects on a spell.

### Options for the printed keyword

There is one reasonable design. Rebound joins `canonicalKeywords`, and `routeStackCardToGraveyardLocked` gains a fourth arm, after flashback.

### Options for granted rebound

- **A. Widen the stack step of the layer pass to layer-6 keyword grants (chosen, owner decision 3).** `stackPinProblem` admits `ModAddKeywords` on a stack pin. The stack step applies those records and battlefield statics over spells ("Instant and sorcery spells you control have rebound") to each spell's keywords, in timestamp order, as the battlefield pass does. Taigam's "that spell gains rebound" is then an ordinary `ScopedEffect` pinned to the spell, and it ends with the object (CR 400.7). This is the rule as written. It also gives the next stack keyword grant (storm or split second, for example) a place to land.
- **B. A rebound-only mark on `StackItem`, plus a battlefield static read at resolution.** It is smaller, and it is the shape ADR 0106 §4 chose for "can't be countered". But "can't be countered" is a rules effect (CR 613.11) and rebound is an ability (CR 702.88a), so this would treat an ability grant as a marker.
- **C. Printed rebound only.** The four granters wait on the row.

### Decision (printed keyword; A for granted, owner decision 3)

1. **The keyword.** `rebound` joins `canonicalKeywords`. Its presence is read from the spell's keywords at resolution: printed, or granted by the stack step.
2. **The resolution arm.** If the spell resolved, it has rebound, it is not a copy, and `item.CastFromZone == ZoneHand`, the route's destination becomes exile. Its continuation schedules a delayed trigger for the spell's controller with `At: StepUpkeep` and `ControllerTurnOnly: true`, pinned to the exiled card's `ObjectRef`. The order against the other arms follows the existing comment's reasoning: flashback first, because CR 702.34a replaces every exit (and a flashed-back spell was not cast from a hand anyway). Rebound and buyback replace the same event, so CR 616.1 gives the choice to the spell's controller. The engine asks only when both apply, which no printed card does.
3. **The upkeep cast.** A keyed body, `rebound/cast`, offers the cast through the same `may_cast` prompt suspend uses. It grants a `{0}`-priced per-card permission with flash timing, so a sorcery can be cast in the upkeep (CR 608.2g). Declining leaves the card in exile.
4. **Copies.** A copy of a rebound spell was not cast (CR 707.10), so it was not cast from a hand, and it ceases to exist off the stack (CR 707.10a). The arm skips it, as the flashback arm does.
5. **Granted rebound.** Taigam's and Ojer Pakpatiq's "that spell gains rebound" write a `ScopedEffect` pinned to the spell with `ModAddKeywords{"rebound"}`. Narset's −2 is a delayed trigger on `EventCast` ("when you next cast an instant or sorcery spell from your hand this turn") that writes the same record. Cast Through Time is a static over spells, applied in the stack step.

### Snapshot impact

- `rebound/cast` is a new keyed body: an on-disk identity, never renamed. An older binary refuses a file naming it with `ErrUnknownEffectKey`, which is the designed rollback case.
- A stack pin may now carry `ModAddKeywords`. The record shape is unchanged. The PR checks that an older binary refuses such a file rather than restoring a spell without its rebound (the `stackPinProblem` check has to run on restore as well as at registration). That is the same rollback case.

### Cards

About 32 of the 34 printed rebound spells, each verified in the PR. Blossoming Calm ("you gain hexproof until your next turn") and World at War (an additional combat and main phase after the second main phase) are the two I could not confirm. Cast Through Time, Narset Transcendent, Ojer Pakpatiq, Deepest Epoch and Taigam, Ojutai Master land too.

---

## 4. Disturb (#1855)

### What exists, what is missing

- **Casting a back face exists.** A defeated Siege's "cast it transformed" (CR 310.12b) is a `CastPermission` with `Faces: []int{1}`, and `faceForCastLocked` (`face.go`) honours it. The card is on the stack back face up, and it stays that way onto the battlefield (CR 712.14a). `MoveCard` turns it front face up in every other zone (CR 712.8a).
- **The mana value is already right.** `manaCostForValue` (`cost_modifier.go`) computes a back face's mana value from the front face (CR 712.8c, 712.8e).
- **Graveyard alternative costs exist.** `AlternativeCost.FromZone` is what makes flashback and escape "a price plus a place" (`alternative_cost.go`).
- **The back face's "If [this] would be put into a graveyard from anywhere, exile it instead" is one field away.** `effects.GraveyardBecomesExile` is Rest in Peace and Leyline of the Void. A self-only version is Blightsteel Colossus's shape.
- **Missing:** an alternative cost that names a face, and disturb's keyword.

### The rules

- **CR 702.146a:** "Disturb [cost]" means "You may cast this card transformed from your graveyard by paying [cost] rather than its mana cost." It is on the front face.
- **CR 702.146b:** a resolving spell cast using disturb enters the battlefield back face up.
- **CR 712.8c:** a spell cast transformed has only its back face's characteristics, and its mana value comes from the front face.
- **CR 601.2b, 601.2f–h:** disturb is an alternative cost, so it can't be combined with another (flashback, for example), and additional costs and commander tax still apply.
- **CR 614.1a:** the back face's exile clause is an ordinary replacement effect. It is read from whichever face is up, so it applies on the stack and the battlefield and not in any other zone (CR 712.8a).

### Decision

There is one reasonable design, so there is no question.

1. **The cost.** `AlternativeCost.CastsFace int` (zero meaning the front face, as today). A disturb offer is `{Key: "disturb", FromZone: ZoneGraveyard, CastsFace: 1, ManaCost: <the disturb cost>}`. `faceForCastLocked` accepts the named face when that offer is claimed. An offer that names a face the card lacks is refused at Register.
2. **The constructor.** `effects.Disturb("{1}{U}")` returns the offer and lists the graveyard in `Spec.CastableZones`. Register checks that the card's layout is `transform`.
3. **The back face's exile clause.** `GraveyardBecomesExile{SelfOnly: true}` on the back face's entry (`<oracle_id>#1`). Because it is read only while that face is up, a disturbed Aura that is put into a graveyard from the battlefield or the stack is exiled. The same card discarded from a hand goes to the graveyard and can be disturbed again.
4. **Keyword.** `disturb` joins `canonicalKeywords` so the importer keeps it. It has no engine behaviour of its own: the card file's offer is the behaviour, as with flashback.
5. **The client.** The graveyard cast menu already lists alternative costs by label ("Flashback {R}"). Disturb shows as "Disturb {1}{U}", and the card preview shows the back face.

### Snapshot impact

None. `StackItem.AltCost` already carries the key, and the face is carried by `Card.ActiveFace`.

### Cards

About 20 of 25. Covert Cutpurse ("a creature you don't control that was dealt damage this turn") and Dennick, Pious Apprentice ("Cards in graveyards can't be the targets of spells or abilities") need more. Brine Comber, Faithbound Judge and Katilda are checked in the PR. Malevolent Hermit // Benevolent Geist, the row's `Waiting` card, lands Full: its back face's static is the counter gate ADR 0106 PR 2 built.

---

## 5. Damage that can't be prevented (#1853)

### What exists, what is missing

- **Every prevention effect always works.** `builtin_replacements.go` says CR 615.12 "is not modelled anywhere in this engine". Banefire and Bonecrusher Giant carry caveats for it.
- **Prevention effects are not marked as prevention.** A prevention effect is a `ReplacementEffect` that watches damage and cancels or reduces it. Protection's built-in, the two scoped shields (`ModPreventDamage`, `ModPreventCombatDamage`) and catalog statics such as Glittering Lion's all do that. Nothing tells them apart from a replacement that doubles or redirects damage. CR 615.12 needs that distinction.
- **The pattern exists.** ADR 0106 §4 put a rules effect, "can't be countered", at one gate with several sources (`spellCantBeCounteredLocked`). Unpreventable damage is the same kind of effect.

### The rules

- **CR 615.1a:** effects that use the word "prevent" are prevention effects.
- **CR 615.12:** "If unpreventable damage would be dealt, any applicable prevention effects are still applied to it. Those effects won't prevent any damage, but any additional effects they have will take place. Existing damage prevention shields won't be reduced by damage that can't be prevented."
- **CR 615.12a:** a prevention effect is applied to a given unpreventable damage event only once.
- **CR 702.16e:** protection prevents damage, so it is a prevention effect too, and unpreventable damage gets through it.
- **CR 613.11 / 611.3a:** "Damage can't be prevented" is a rules effect. As a static it applies to whatever its text names at each moment. From a resolved spell ("this turn") it covers damage dealt later in the turn (CR 611.2c), and it ends at cleanup (CR 514.2).

### The forms the cards print

| Form | Printed example | Source of the rule |
|---|---|---|
| This spell's or ability's own damage | Combust, Pinpoint Avalanche, Banefire (X ≥ 5), Urza's Rage (kicked) | The damage instruction |
| A source's damage | Excruciator, Malignus | A static on the source |
| All damage | Leyline of Punishment, Sunspine Lynx | A static on the battlefield |
| Combat damage, or combat damage by your creatures | Frenzied Baloth, Questing Beast | A static on the battlefield |
| All damage this turn | Skullcrack, Flaring Pain, Stomp, Wild Slash | A rule grant from a resolved spell |
| Damage to one creature this turn | Whippoorwill, Lava Burst | A grant pinned to that creature |

### Options

- **A. A rules gate read when damage is dealt, plus a prevention flag on replacements (chosen).** Every prevention effect declares itself as one. One gate, `damageUnpreventableLocked(ev)`, asks every source in the table above. While a damage event is unpreventable, the replacement loop applies each prevention effect once (CR 615.12a), lets it run its additional effects, and keeps the damage and the shield's charge unchanged.
- **B. Skip prevention effects entirely for unpreventable damage.** It is smaller, but it is wrong under CR 615.12: a prevention effect's additional effects would be skipped and its shield would never be "applied".

### Decision (A)

1. **The flag.** `ReplacementEffect.Prevention bool`, set by every prevention effect: the protection built-ins, both scoped shield kinds, §6's new kind and every catalog replacement that prevents. A test runs every catalog replacement that watches damage against a probe event. It fails when one cancels or reduces the damage without saying whether it is a prevention effect (a redirection that cancels and re-deals declares that it is not). It is the census that stops a new card shipping a prevention effect that ignores the rule.
2. **The gate.** `damageUnpreventableLocked(ev)` asks, in order:
   1. the damage event's own mark (`DealDamage{Unpreventable: true}` stamps it on the event);
   2. the source's static (`Spec.DamageCantBePrevented` with the shape "by this object"), read by last-known information through `ev.SourceLKI`, as protection's check is;
   3. battlefield statics (all damage, combat damage, or combat damage dealt by creatures the static's controller controls), dropped when `CatalogAbilityKey` answers empty;
   4. the turn's rule grants: a game-scoped `ScopedEffect` with a new mod kind, `ModDamageCantBePrevented`, swept at cleanup. A pinned version covers "damage that would be dealt to that creature this turn".
3. **The replacement loop.** When the gate says yes, a `Prevention` replacement is applied once. Its `Replace` is skipped. If it has an additional effect, that effect runs with zero damage prevented (CR 615.12, 615.5). A charged shield's charge is not reduced, and a one-use shield (§6) is not used up. A non-prevention replacement (doubling, redirection) is unaffected.
4. **"Can't be dealt instead to another permanent or player"** (Lava Burst, Whippoorwill) is a redirection ban beside the prevention clause. It ships in the same PR if the redirection replacements prove to be as easy to flag as the prevention ones. Otherwise those two cards go on the row's `Waiting` list.
5. **What the table sees.** The stack chip and the log say "damage can't be prevented" for a marked spell. A live turn grant is a line on the game banner ("Damage can't be prevented this turn — Skullcrack").
6. **Players can't gain life** (owner decision 5, #1880). It ships in the same PR. It is the same kind of effect, a rules effect at one gate, with the same three durations (a static, a turn grant, and Screaming Nemesis's rest of the game). CR 119.7 is its only extra rule. The gate is asked at every life gain, before the replacement window, so "If you would gain life" replacements never see a gain that can't happen (CR 119.10).

### Snapshot impact

- `ModDamageCantBePrevented` and `ModCantGainLife` are new mod kinds: on-disk identities, refused by an older binary with `ErrUnknownEffectKey`. That is the rollback case.
- The per-event mark is transient and never captured.
- The `Prevention` flag is catalog data.

### Cards

About 20 of 31, counting Banefire and Bonecrusher Giant dropping their caveats. With "can't gain life" (decision 5), Skullcrack, Call In a Professional, Sunspine Lynx and Leyline of Punishment land too, and Screaming Nemesis drops its caveat. The other "can't gain life" cards on #1880 land in the same PR where they need nothing else. Volcano Hellion (echo {X} where X is your life total), Everlasting Torment (all damage as though it had wither) and Spider-Punk (riot) go on `Waiting` lists for their own blockers.

---

## 6. The next damage from a source (#1860)

### What exists, what is missing

- **Two shield kinds exist.** `ModPreventDamage` is CR 615.7's charged shield ("prevent the next 3 damage"), keyed on the damage target. `ModPreventCombatDamage` is Fog's. Both are scoped-effect records with replacement readers (`scoped_replacements.go`), so a table with one is a restore point.
- **The source-keyed, one-use shield is deliberately unbuilt.** `effects/prevention.go` says the Circle of Protection form "is not built: no catalogued card prints it, and ADR 0041 P8 adds a kind only with the first card that needs it". 45 printed cards need it.
- **Nothing lets a player choose a source.** There is no pending choice over "a permanent, a spell on the stack, or a face-up object in the command zone".
- **"The damage prevented this way" cannot be read.** A prevention effect reports nothing about what it prevented.
- **A stale citation.** `applyScopedReplacementLocked` calls its N-shield arithmetic "CR 615.8's". It is CR 615.7's. The PR corrects the comment.

### The rules

- **CR 615.8:** "Some prevention effects generated by the resolution of a spell or ability refer to the next time a specific source would deal damage. These effects prevent the next instance of damage from that source, regardless of how much damage that is."
- **CR 615.9 / 609.7b:** a shield against "a source of a player's choice with certain properties" rechecks the properties when the source would deal damage. If they no longer match, the damage isn't prevented and the shield isn't used up.
- **CR 609.7a:** the player may choose a permanent, a spell on the stack, any object referred to by an object on the stack, by a prevention or replacement effect waiting to apply, or by a delayed trigger waiting to trigger, or a face-up object in the command zone. A source need not be able to deal damage. "The source is chosen when the effect is created."
- **CR 615.5:** a prevention effect's additional effect may refer to the amount prevented, and it takes place immediately after the prevention.
- **CR 616.1:** the affected player orders competing prevention effects. That is already how the scoped shields report no controller.
- **CR 400.7:** a chosen permanent that changes zones is a new object, so a shield against it stops applying.

### Options

- **A. One new scoped prevention kind, a source choice, and a "prevented this way" follow-up (chosen, owner decision 4).** `ModPreventNextFromSource` holds the source (a chosen object, or "this object" for Mercenaries), the protected player or object, an optional property recheck, and an optional follow-up body key. The follow-up runs with the amount prevented (CR 615.5). Reverse Damage, Deflecting Palm, Bone Mask and six others need it.
- **B. The shield and the source choice only.** The nine "prevented this way" cards would have waited on the row.

### Decision (A, owner decisions 4 and 6)

1. **The mod kind.** `ModPreventNextFromSource`, with four fields:
   - `Source` pins one object (CR 400.7).
   - `Protected` is a player, a pinned permanent, or nothing ("would deal damage this turn", Awe Strike and Dazzling Reflection).
   - `Query` is the CR 615.9 recheck ("a red source"), the same data struct as §2's.
   - `Then` is a registered body key (the `ModExileInsteadOfGraveyard` precedent), run after the prevention with the amount prevented.

   It is spent whole by the first damage event it prevents, whatever the amount (CR 615.8). It is not spent by an event whose source fails the recheck, or by unpreventable damage (§5).
2. **Choosing a source.** A new pending choice, `choose_source`, made as the shield is created (CR 609.7a). It lists the permanents on the battlefield, the spells on the stack, and the face-up objects in the command zone, filtered by the card's property ("a black or red source", "a creature of the chosen type"). Per owner decision 6 it also lists everything else CR 609.7a allows: any object referred to by an object on the stack, by a prevention or replacement effect waiting to apply, or by a delayed trigger waiting to trigger, even if that object has left the zone it was in.
3. **Helpers.** `effects.PreventNextDamageFromChosenSource(query, protected)` covers the Circle of Protection, Rune of Protection and Pentagram families. `effects.PreventNextDamageFromThis()` covers Mercenaries. Each takes an optional `Then`.
4. **The bot.** The heuristic answers `choose_source` with the legal source controlled by an opponent that has the most power, and otherwise the first listed. The model tiers see the list with each source's controller.

### Snapshot impact

A new mod kind with plain-data fields, recorded under the current schema version. `Then` names a registered body, and a restore point naming an unknown kind or body is refused (`ErrUnknownEffectKey`). The choice is an ordinary pending choice.

### Cards

About 43 land. The plain shields: the seven Circles of Protection, the seven Runes of Protection, Circle of Solace, Story Circle, Prismatic Circle, Greater Realm of Preservation, Haazda Shield Mate, Invulnerability, Kithkin Armor, Martyr's Cause, Mercenaries, Penance, Pentagram of the Ages, the two Pilgrims, Righteous Aura, Samite Blessing, Sanctum Guardian, Seasoned Tactician, Charm Peddler, Circle of Despair and Dazzling Reflection. The "prevented this way" follow-up (decision 4) adds Awe Strike, Bone Mask, Cho-Arrim Alchemist, Deflecting Palm, Honorable Passage, Intervention Pact, New Way Forward, Reverse Damage and Shadowbane. Rhystic Circle and Desperate Gambit are checked in the PR.

---

## Shared machinery

- **`PermanentQuery`** is shared by §2 (the defending player's permanents), §6 (the source recheck) and §1's helpers ("When you control no Islands"). It is plain data so that it can live in a characteristic and in a scoped effect. It is the first closed permanent filter in `game/`. `PermissionFilter` is its spell-side sibling.
- **§5 and §6 share the prevention loop.** The `Prevention` flag is §5's. §6's shield must declare it and must not be spent by unpreventable damage, so §6 goes after §5 or is rebased onto it.
- **§3 and §4 share nothing in code**, but both add an arm to how a spell is cast or put away. Each is small enough to review on its own.
- **The enumerator.** Nothing here adds a new kind of move except §4's graveyard cast (an existing alternative-cost move) and §6's `choose_source` answer. The ADR 0033 §1 rule (one gate shared by the verb and `internal/legal`) holds for §2's predicate exactly as for ADR 0106 §2's.

## Delivery

Each PR lands its engine change and its cards together, test first. Each also flips its registry row to implemented (or partial), adds a closed-seam fragment under `docs/engine-seams/closed/`, adds the registry rows #1879 and #1880 need, and corrects the stale notes listed above. Every PR lands every card its seam unblocks, each verified against its full oracle text. Any that needs more goes on a `Waiting` list with the reason (ADR 0106 decision 6).

1. **State triggers (#1858).** `TriggeredAbility.State`, the check, the derived latch and the three helpers. About 18 cards.
2. **Can't attack unless the defending player controls something (#1879).** `PermanentQuery`, the restriction field, the predicate branch and the chip sentence. About 18 cards. The 14 cards that need both PR 1 and PR 2 land in whichever merges second.
3. **Rebound (#1854).** The keyword, the resolution arm, the delayed trigger and `rebound/cast`. About 32 cards.
4. **Granted rebound (#1854).** The layer-6 stack step for keyword grants. It lands Cast Through Time, Narset Transcendent, Ojer Pakpatiq and Taigam. It needs PR 3.
5. **Disturb (#1855).** `AlternativeCost.CastsFace`, `effects.Disturb`, the self-only exile replacement and the keyword. About 20 cards.
6. **Damage that can't be prevented (#1853).** The `Prevention` flag and its census, the gate and its four sources, the CR 615.12 loop, `ModDamageCantBePrevented`, and "players can't gain life" (#1880: its gate and `ModCantGainLife`). About 24 cards from #1853, plus the #1880 cards that need nothing else.
7. **The next damage from a source (#1860).** `ModPreventNextFromSource`, `choose_source` (engine, view, client picker and bot answer), the helpers and the "prevented this way" follow-up body. About 43 cards. It needs PR 6.

PRs 1, 2, 3, 5 and 6 touch different code and can go in parallel. PR 4 follows PR 3, and PR 7 follows PR 6.

## Consequences

- Triggers can watch a game state as well as an event. Any future "when there are no …" or "when you have no …" card uses `TriggeredAbility.State`, and nothing stores a latch.
- Attack legality can depend on what the defending player controls. It is still one predicate per attacker and target.
- A spell can gain an ability on the stack. The stack step of the layer pass is layers 2 and 6, keywords only.
- An alternative cost can cast a card's other face. A future "cast it transformed" keyword (More Than Meets the Eye) uses `CastsFace`.
- Every prevention effect is marked as one, and a test keeps it that way. "Can't be prevented" is read at one gate, like "can't be countered".
- A shield can be keyed on a source as well as a target, and a prevention effect can report what it prevented.

## Out of scope

- The fifteen deferred seams in the sizing table.
- "Can't be dealt instead to another permanent or player" if it turns out not to be a flag (§5 decision 4).
- Assist (#1857), which needs a second payer in the middle of a cast. It deserves its own ADR, with the client prompt.

---

## Questions for the owner (answered)

These are the questions as asked. The owner's answers follow.

1. **When are state triggers checked (§1, #1858)?**
   - **(a) Recommended:** after every event the engine emits, and in each pass of the CR 704.3 loop. A state that lasts only a moment during a resolution still triggers, as in CR 603.8's own example.
   - (b) Only in the CR 704.3 loop, when a player would receive priority. It is simpler, and it misses states that come and go within one resolution.
2. **"Can't attack unless defending player controls an Island" (§2).**
   - **(a) Recommended:** ship it in this ADR as one more field on ADR 0106 §2's per-target restriction. It lands about 32 cards, 14 of which also need §1's state triggers.
   - (b) Leave it for later. The 14 serpents wait, and only about 18 state-trigger cards land.
3. **Rebound given to a spell (§3, #1854).**
   - **(a) Recommended:** widen the stack step of the layer pass to layer-6 keyword grants, so "that spell gains rebound" and "spells you control have rebound" are ordinary ability-adding effects (CR 613.1f). It lands Taigam, Ojer Pakpatiq, Narset Transcendent and Cast Through Time.
   - (b) A rebound-only mark on the stack item, plus a battlefield static read at resolution. It is smaller, but it treats an ability as a rules marker.
   - (c) Printed rebound only. The four granters wait.
4. **"The damage prevented this way" (§6, #1860).**
   - **(a) Recommended:** the new shield reports the amount it prevented, and a follow-up body runs with it immediately afterward (CR 615.5). It lands nine more cards: Reverse Damage, Deflecting Palm, Bone Mask and six others.
   - (b) The shield only. The nine cards wait on the row.
5. **"Players can't gain life" alongside unpreventable damage (§5, #1853).**
   - **(a) Recommended:** ship it in the same PR, as a second rules gate with the same three durations. Skullcrack, Call In a Professional, Sunspine Lynx and Leyline of Punishment then land Full. It also opens the other 17 uncatalogued "can't gain life" cards and Screaming Nemesis's caveat to later batches.
   - (b) Leave it to its own row. The four cards wait.
6. **What a player may choose as "a source of your choice" (§6, #1860).**
   - **(a) Recommended:** everything CR 609.7a allows: permanents, spells on the stack, face-up objects in the command zone, and objects that are referred to by something on the stack, by a waiting prevention or replacement effect, or by a waiting delayed trigger.
   - (b) Permanents, spells on the stack and the command zone only. This covers nearly every real choice and makes a simpler picker, but it is narrower than the rule.

---

## Owner decisions, 2026-10-01

The owner answered the six questions on 2026-10-01. Every answer was the recommended option (a).

1. **State-trigger timing.** State triggers are checked after every event the engine emits and in each pass of the CR 704.3 loop, so a momentary state during a resolution triggers (§1, decision 2).
2. **Can't attack unless the defending player controls something.** It ships in this ADR as §2, one data field on ADR 0106 §2's per-target restriction. Tracked on #1879.
3. **Granted rebound.** The stack step of the layer pass is widened to layer-6 keyword grants, so "that spell gains rebound" and "spells you control have rebound" are ordinary ability-adding effects (CR 613.1f). Taigam, Ojer Pakpatiq, Narset Transcendent and Cast Through Time land in Delivery PR 4.
4. **"The damage prevented this way."** The source shield reports the amount it prevented, and its follow-up body runs immediately afterward with that amount (CR 615.5). The nine follow-up cards land in Delivery PR 7.
5. **"Players can't gain life."** It ships in the same PR as damage that can't be prevented (Delivery PR 6), as a second rules gate with the same three durations. Tracked on #1880.
6. **"A source of your choice."** `choose_source` offers everything CR 609.7a allows: permanents, spells on the stack, face-up objects in the command zone, and objects referred to by an object on the stack, by a waiting prevention or replacement effect, or by a waiting delayed trigger.
