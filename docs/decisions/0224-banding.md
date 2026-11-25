# ADR 0224 — Banding: attacking in a band, blocking a band, and who divides the damage (CR 702.22)

**Status:** Accepted · 2026-11-19 · S82 — Card coverage: sets outside Standard, newest first. The owner reviews it before the PR merges; the two recommendations flagged in §8 are the ones to look at.
**Issues:** [#5723](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5723) (seam: banding), found while adding the first Limited Edition Alpha batch ([#5642](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5642)). Cards: Benalish Hero, Mesa Pegasus, Timber Wolves, Helm of Chatzuk.
**Numbering:** after `git fetch --all --prune`, `git ls-tree --name-only <ref> docs/decisions/` over all 272 remote branches. The highest number anywhere was **0223**, so this is **0224**.
**Builds on:** [ADR 0045](0045-combat-restrictions.md) (the block declaration and its lock-in, #830 and #1279; a creature blocking several attackers, Decisions 60-63, #1706), [ADR 0147](0147-auto-assign-combat-damage.md) (the canonical damage split and the client's auto-assign), the attack declaration's lock-in (#859, `game/attackers.go`) and the bulk declaration (#318).

---

## Context

### The rules

From the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`):

- **CR 702.22a:** banding is a static ability that modifies the rules for combat.
- **CR 702.22b:** "bands with other" is a special form of banding; losing banding loses every "bands with other".
- **CR 702.22c:** as a player declares attackers, they may declare that one or more attacking creatures with banding and up to one attacking creature without banding are all in a band. Any number of bands, each creature in at most one. Defending players can't declare bands.
- **CR 702.22d:** all creatures in an attacking band attack the same player, planeswalker or battle (also **CR 802.3b**).
- **CR 702.22e:** once announced, a band lasts for the rest of combat, even if a member loses banding.
- **CR 702.22f:** an attacking creature removed from combat is removed from its band.
- **CR 702.22g:** banding shares and removes no abilities; the members are separate permanents.
- **CR 702.22h:** if an attacking creature becomes blocked by a creature, each other creature in its band becomes blocked by that same blocking creature. The rulebook's example: a band of a flier and a swampwalker, the defender blocks the flier, and the swampwalker becomes blocked too.
- **CR 702.22i:** if one member of a band would become blocked due to an effect, the entire band becomes blocked.
- **CR 702.22j:** in the combat damage step, if an attacking creature is being blocked by a creature with banding, the **defending** player (rather than the active player) chooses how the attacking creature's damage is assigned, dividing it as they choose among any creatures blocking it. An exception to CR 510.1c.
- **CR 702.22k:** if a blocking creature is blocking a creature with banding, the **active** player (rather than the defending player) chooses how the blocking creature's damage is assigned, dividing it as they choose among any creatures it's blocking. An exception to CR 510.1d.
- **CR 702.22m:** multiple instances are redundant. (The rule has no 702.22l.)
- **CR 508.1e:** if any chosen attacker has banding or "bands with other", the active player announces which creatures, if any, are banded with which. It comes after the requirements (508.1d) and before the taps (508.1f) and the costs (508.1h).

The issue text read 702.22j as the attacking player's division and 702.22k as the defending player's. The pinned text is the other way round, as above, and this ADR and the code follow the text.

### What the engine had

- The attack declaration is staged and locked in (#859): `declare_attacker` stages one creature, `declare_attackers` stages a set and locks it in at once, and `commitAttackDeclarationLocked` announces it. The client's attack flow stages locally and sends one `declare_attackers` (#3266).
- A creature can block several attackers (#1706): `Card.BlockingTarget` holds the first, `Card.AlsoBlocking` the rest, and the combat damage steps read `BlockedAttackers()`. A blocker of two or more divides its damage through the CR 510.1c prompt turned round (`BlockerDivides`).
- The block declaration is announced once per defender, at `commitBlockDeclarationLocked`, which emits one `EventBlock` per pair (with the pair's place, so "whenever this creature blocks" triggers once) and one `EventBecomesBlocked` per newly blocked attacker.
- The CR 510.1c prompt (`PendingChoiceDamageAssignment`) checks the answer against `PendingChoice.Chooser`, not against the source's controller, and the bots answer it with the canonical split the enumerator lists first (ADR 0147).

## Decisions

### 1. A band is a game-level map, attacker to band key

`Game.attackBands map[uuid.UUID]uuid.UUID`: each banded attacker's instance ID to its band's key, the instance ID of the band's first member in the declaration. Two attackers are in one band when their keys are equal; a key is compared, never resolved, so a band whose first member has left keeps its key.

- **Membership is read live.** `bandMatesLocked(id)` returns the other members still on the battlefield and still attacking. That is CR 702.22f for every exit at once: a zone change mints a new instance ID, and an attacker whose `AttackingTarget` was cleared is not attacking. `removeFromCombatLocked` and phasing's `forgetCombatRecordLocked` also delete the entry, so a creature taken out of combat and later put back attacking does not rejoin.
- **It lasts the combat.** Cleared with the rest of combat (`clearAttackAnnouncementsLocked`), and nothing else rewrites it, so a member that loses banding stays in the band (CR 702.22e).
- **Snapshot:** `GameSnapshot.AttackBands`, `json:"attackBands,omitempty"`, additive within `SnapshotSchemaVersion` 9 and recorded in the snapshot shape file. A file written before it restores with no bands, which is the only kind of combat there was. A binary before it that drops the key restores a mid-combat band as unbanded attackers: weaker, never wrong for a game that had no banding cards. No fixture changes, no effect keys. Cloned with the undo stack.

**Alternatives.** A field on `Card` (`Card.Band`) would have needed clearing at every one of the eight places that write `AttackingTarget` directly, and the drift tests for `Card` and `CardSnapshot` besides; the map has one write and one clear. A fresh `uuid.New()` per band was rejected because the engine's events must replay deterministically; the first member's ID is deterministic.

### 2. Bands are announced on `declare_attackers`, as a `band` number per entry

`AttackDeclaration.Band int` (wire: `attackers[].band`, omitempty). The entries of one `declare_attackers` that share a positive number are one band; 0 or absent is none; a group of one entry is no band (it changes nothing). Negative is refused as a malformed action.

- Bands are judged on the **eligible** entries, after the existing silent skip of tapped, sick and restricted creatures, so a skipped creature is no member. They are judged **all or nothing**, like the CR 508.1c count limits and the attack tax: a band that breaks CR 702.22c-d refuses the whole declaration and nothing is declared, tapped or paid. Order: after the requirement check and the exert and enlist validation, before the tax, matching CR 508.1d → e → h.
- **Refusals** are a `*game.BandRefusedError` wrapping `game.ErrIllegalBand`, sent as `illegal_attack` with reason `band`, the first creature of the band in `card_id`, and a sentence:
  - two creatures without banding: "A band can have only one creature without banding, and Grizzly Bears and Hill Giant both lack it."
  - split across defenders: "Creatures in a band must attack the same player, planeswalker or battle, and Benalish Hero and Grizzly Bears don't."
  - different controllers (an admin driving two seats): "Creatures in a band must all be controlled by the attacking player."
  - a band sent while confirming a declaration another player chose (#4154): "Bands are announced by the player declaring the attackers, not as the attacks are confirmed."
- Banding is read off the effective characteristic with fresh layers, so printed, imported and granted banding are the same.
- **The per-creature verb announces no band.** `declare_attacker` ignores `Band`. Re-pointing a banded attacker at a different target with it (the sandbox's re-declare) takes that attacker out of its band, since CR 702.22d forbids the band splitting; the rest stay banded. A CR 508.7 reselect leaves bands as they are (CR 702.22e: the band lasts).
- The wire's card view gets `attack_band` (omitempty): the band's number, 1, 2, … in battlefield order of each band's first member, on each attacker in a band that still has two members in combat. A number rather than the key, so the view never names an instance ID it would otherwise hide. Public: the band is announced to the table.

**Alternatives.** A separate `band_attackers` verb over already-staged attackers would have fitted the per-creature flow, but `declare_attackers` locks the declaration in as it stages it, so a band sent after it would be late, and the client's attack flow sends `declare_attackers` anyway. A `band_with: <id>` on each entry was rejected because a band is a set, and pairwise links can describe an illegal shape (A with B, B with C, C with nobody) that the number cannot.

### 3. A blocker of one member blocks the band, at the block lock-in

`commitBlockDeclarationLocked` first widens each announced blocker's set to the bands of the attackers it blocks (`expandBandBlocksLocked`): declared attackers first, band-mates after. Everything downstream is the existing #1706 machinery for a creature blocking several attackers.

- **Legality is judged on the declared pairs**, which is the rule's own example: the swampwalker becomes blocked by a creature that could never have been declared against it. A blocker's capacity is not consulted for the band-mates either; they are blocked by the band, not declared.
- **Events.** One `EventBlock` per pair, the band-mates' at a place past 1, so "whenever this creature blocks" triggers once (CR 509.3a) and "becomes blocked by a creature" triggers for every member it now blocks (CR 509.3d). One `EventBecomesBlocked` for each member that was unblocked (CR 509.3c).
- **Effects.** `BecomeBlockedForEffect` (Fog Patch) blocks every member of a band it names one of (CR 702.22i), and a creature put onto the battlefield blocking a member blocks the band (`stampEntryBlockerLocked`, CR 702.22h).
- The widening happens at the lock-in, not as the defender clicks: a staged block shows the declared pair, and the board shows the band through `attack_band`, so the defender knows what blocking one member means before they finish.

### 4. Banding changes who divides the damage, and nothing else about the prompt

Both divisions reuse the CR 510.1c prompt. The frame gains `DamageAssignmentFrame.Banding` (`json:",omitempty"`, additive, recorded in the shape file); the wire's `damage_assignment` gains `banding`. `ResolveDamageAssignment` already checks `PendingChoice.Chooser`, so changing the chooser is the whole of the hand-off.

- **CR 702.22j.** In the attacker loop of `assignAndDealCombatDamageLocked`: when any creature blocking the attacker has banding *as damage is assigned*, the **defending player** of the attack (falling back to the banding blocker's controller) divides its damage among the creatures blocking it. With one blocker there is nothing to divide and the whole power goes on it with no prompt. Trample is off (the division is "among any creatures blocking it"), so nothing tramples over a banding block, and Wolverine's "as though it weren't blocked" is not offered.
- **CR 702.22k.** In `queueBlockerDamageDivisionLocked`: when any attacker the blocker blocks has banding as damage is assigned, the **active player** divides the blocker's damage among the creatures it blocks.
- Because both read banding as damage is assigned, a band whose members have all lost banding is blocked as a band (CR 702.22e) but its blocker's controller divides, as CR 702.22k says.
- The source's controller still owns everything else the frame carries: lifelink, commander damage, infect. Only the chooser changes.
- A chooser who leaves the game drops the prompt as it did before: under CR 702.22j the blockers left with them, and under CR 702.22k the attackers did.

**The canonical split.** For a banding frame, `legal.CanonicalDamageSplit` is **protective**: it loses the chooser the least. Damage on a creature it can't destroy (indestructible, protected from the source) first; otherwise spread short of lethal when the power allows, else all of it on the one creature worth least, since killing one absorbs any amount. `covers_lethal` is never set on a banding frame, so the client's "Auto-assign combat damage" never sends a division of the player's own creatures unasked. The enumerator still lists the other divisions after it.

### 5. The client: a Band toggle in the Choose attackers picker, and a BAND badge (minimal)

What ships is the smallest affordance that lets a human form a band and see one:

- **Forming a band.** The Choose attackers picker (`AttackDeclarationModal`, opened from the attack row's "Choose attackers…") shows a **Band** toggle on every row whenever a creature there has banding (`card.abilities` includes `banding`). The checked rows toggled on attack as one band at that defender (`band: 1`). A band with two creatures without banding shows the engine's sentence under the rows and holds the confirm; the server checks it again. The toggle's accessible name is the contract label `L.bandAttacker` ("Band <name>").
- **Seeing a band.** An attacker with `attack_band` wears a `BAND n` badge, top centre, with a tooltip saying that blocking one blocks them all. `Banding` joins the keyword badge row.
- **The division.** The damage-assignment sheet says when banding has handed the division over ("is blocked by 2 of your creatures, and one of them has banding, so you divide its 4 damage among them however you like. None of it tramples over.").

**Minimal, and why it is not silently wrong:** the picker forms one band per defender per declaration. The engine takes any number; two bands at one defender, or a band formed by clicking creatures on the board, are not offered by this client yet. Clicking creatures stages them unbanded, exactly as before, and the board shows no badge, so nothing looks banded that isn't. A follow-up can add a board affordance.

### 6. The bot: never bands; blocks a band only to survive; divides protectively

- The bot never forms a band: the enumerator offers per-creature attacks, which announce none.
- The block scorer prices a block as one attacker against its blockers, which is the wrong picture for a band (the blocker takes every member's damage, and the attacking player divides its damage). So the bot blocks a banded attacker only when it is otherwise dead this combat (`desperate`), and then counts the damage the block stops from the whole band. `aiseat/heuristic/banding.go`.
- A banding division handed to the bot is answered with the protective canonical split (§4).

**Alternative.** Teaching the block scorer to price a band as one gang attacker was rejected for now: it needs the attacking player's division modelled too, and the beta's banding cards are four Alpha commons.

### 7. Out of scope

- **"Bands with other"** (CR 702.22b and the second half of 702.22c, Legends' "bands with other legends" and friends). It is not a canonical keyword, so a card that prints it stays flagged by the coverage importer and is never catalogued as Full. It would need a quality per instance and a second band shape; nothing in the beta's decks asks for it yet.
- Bands on the per-creature `declare_attacker` verb (§2), and more than one band per defender in the client (§5).

### 8. Recommendations for the owner

1. **Minimal client (§5).** Recommended: ship the picker toggle now and add a board affordance (select creatures, then "Band") only if players ask. The alternative is to hold the PR for a board flow.
2. **Bot policy (§6).** Recommended: never band, block a band only to survive. The alternative, a scorer that prices bands, is a bot-strategy change worth its own issue.

## Consequences

- `PrintedKeywords: []string{game.KeywordBanding}` is enforced, and "gains banding" is an ordinary keyword grant (`GrantKeywordUntilEOT`, `KeywordGrant`). Benalish Hero, Mesa Pegasus, Timber Wolves and Helm of Chatzuk ship `Full`. A creature whose only text is banding needs no card file.
- `banding` is a canonical keyword, a coverage mechanic row (`cards/coverage/caveats.go`) and a roadmap item.
- New wire fields, all omitempty: `declare_attackers` `attackers[].band`; card view `attack_band`; `damage_assignment.banding`; error reason `band` on `illegal_attack`.
- New snapshot fields, additive in v9: `attackBands`, and `Banding` on a damage-assignment frame (pending and held). No fixture edited.

## Amendment (2026-11-25): Defensive Formation and Cunning Giant (#6192)

Two printed cards move the damage assignment the same way banding does, and the owner decided (2026-11-24) to build them with the prompt addressed to the other seat.

- **Defensive Formation**: "Rather than the attacking player, you assign the combat damage of each creature attacking you. You can divide that creature's combat damage as you choose among any of the creatures blocking it." These are the words of CR 702.22j, so the division is §4's banding division: `effects.Spec.YouAssignAttackersDamage` (→ `game.CardDef`, `CatalogYouAssignAttackersDamage`) makes the player a creature is attacking, when they control such a permanent, the chooser of its division among two or more blockers, with no trample over and no "as though unblocked". The frame carries `Banding` (the rules and the bot's protective split are banding's) and `DefenderAssignsBy`, the permanent's name, for the sheet's text. A creature attacking that player's planeswalker or battle is not "attacking you". An unblocked attacker, or one blocked by one creature, has nothing to divide (CR 510.1b-c).
- **Cunning Giant**: "If this creature is unblocked, you may have it assign its combat damage to a creature defending player controls." `effects.Spec.UnblockedAssignsToACreature` (→ `CatalogUnblockedAssignsToACreature`, read live off the attacker). An unblocked attacker with it, whose defending player controls a creature, is asked in the combat damage step with the existing `damage_assignment` prompt: `blocker_card_ids` are those creatures, `as_though_unblocked` is set (the player is the other answer) and the new `one_creature` frame flag says the answer is all of `attacker_power` on exactly ONE of them, or all of it as `trample_to_player`; a division is refused. The bot's canonical answer is the player (what the creature would do anyway); the move list offers each creature after it; `covers_lethal` is never set, so the browser never sends it unasked.
- **Wire**, omitempty: `damage_assignment.defender_assigns_by`, `damage_assignment.one_creature`. **Snapshot**, additive in v9: `DefenderAssignsBy` and `OneCreature` on a damage-assignment frame.
- **Client**: the sheet names the permanent for a Formation division, and for a one-creature prompt says "one creature, or all of it to the defending player" and holds the confirm on a division (`oneCreatureSplit`).

Cards: Defensive Formation (OLD b720), Cunning Giant (b733), built by their batches.
