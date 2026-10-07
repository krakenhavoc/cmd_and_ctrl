# ADR 0130 — Exert: an optional cost to attack, a skipped untap, and the triggers that watch it

**Status:** Accepted (owner answers 2026-10-07) · 2026-10-07 · S68 — Cost components and alternative costs
**Issues:** [#2048](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2048) (seam: exert, CR 701.43; Oketra's Avenger). Related: [#2061](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2061) (Combat Celebrant, an S58 deck request that waits on this row), [#2029](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2029) (Telekinesis: "next two untap steps", the duplicate-skip limit), [#2441](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2441) / PR [#2478](https://github.com/krakenhavoc/cmd_and_ctrl/pull/2478) ("Choose attackers…").
**Owner decisions:** the owner's answers to this ADR's five questions on 2026-10-07, each the recommended option. All five are quoted under [Owner decisions](#owner-decisions-2026-10-07) and are binding. The options not chosen are kept under [Questions for the owner (answered)](#questions-for-the-owner-answered).
**Numbering:** 0128 (`0128-playmats.md`) is the highest number on any remote head (`origin/develop`, `origin/main`, `origin/docs/issue-audit`, `origin/feat/2515-playmats-three-slots-fit`, `origin/feat/750-conditional-block-restrictions`, `origin/feat/playmats`, `origin/fix/caddy-reload-admin-off`, `origin/wip/836-one-click-default`, `pr/2326`). 0129 is being written at the same time for energy (#1995), so this ADR takes **0130**.
**Builds on:** [ADR 0058](0058-doesnt-untap.md) (the next-untap marker, `UntapSkip`, and its 2026-09-23 amendment), [ADR 0059](0059-turn-machinery.md) Decision 8 (`TurnTally.Attacks`), [ADR 0080](0080-attack-taxes.md) (a cost paid by the declaration verb), [ADR 0108](0108-turn-scoped-effects-object-history-and-damage-shields.md) §7 (`preventFromSource`, Oketra's Avenger's shield), [ADR 0109](0109-rule-gates-land-types-mana-and-cost-components.md) (cost components), [ADR 0111](0111-action-dock.md) (small prompts in the dock), [ADR 0125](0125-a-walkthrough-that-keeps-up.md) §2 (labels), [ADR 0126](0126-bots-that-play-their-decks.md) §6 (the declared `purpose`).

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

Exert is printed on 36 Commander-legal cards, and none is catalogued. Two are waiting on this row: Oketra's Avenger ("You may exert this creature as it attacks. When you do, prevent all combat damage that would be dealt to it this turn") and Combat Celebrant, from the S58 deck requests ("If this creature hasn't been exerted this turn, you may exert it as it attacks. When you do, untap all other creatures you control and after this phase, there is an additional combat phase").

Every claim below was checked on `origin/develop` at `0e20f52b1`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`). Exert has not moved: it is still CR 701.43 in that edition.

### Owner decisions (2026-10-07)

The owner answered this ADR's five questions on 2026-10-07, each with the recommended option:

1. **Every path asks** (question 1, §7). A click on an exert creature offers "Attack" and "Attack and exert" in the dock, the context menu has both rows, "Choose attackers…" has an Exert toggle on each row, and "Attack with all" opens that picker when an exert creature could attack. There is no ADR 0127 preference for the question yet.
2. **Chosen at the declaration, paid at the lock-in** (question 2, §2). The exert is paid when the declaration locks in, in the same event batch as the attacks.
3. **Exert as an activation cost is in scope** (question 3, §4). The cost component lands in PR 4 with its 8 cards.
4. **The heuristic prices exerting** (question 4, §9). The gain comes from declared purposes plus a new `pump` field. The cost is the creature's next-turn attack and block value, and zero when exerting is free. It is measured with the arena report.
5. **Every exert card the engine can fully support** (question 5, §11). A card that needs another seam goes on that seam's `Waiting` list.

### The rules

- **CR 701.43a:** "To exert a permanent, you choose to have it not untap during your next untap step."
- **CR 701.43b:** "A permanent can be exerted even if it's not tapped or has already been exerted in a turn. If you exert a permanent more than once before your next untap step, each effect causing it not to untap expires during the same untap step."
- **CR 701.43c:** "An object that isn't on the battlefield can't be exerted."
- **CR 701.43d:** "'You may exert [this creature] as it attacks' is an optional cost to attack (see rule 508.1g). Some objects with this static ability have a triggered ability that triggers 'when you do' printed in the same paragraph. These abilities are linked. (See rule 607.2h.)"
- **CR 508.1** declares attackers as one turn-based action, in order. **508.1a** chooses the attackers; **508.1d** says a player is not required to pay a cost to attack in order to obey a requirement; **508.1f** taps them; **508.1g** "If there are any optional costs to attack with the chosen creatures (expressed as costs a player may pay 'as' a creature attacks), the active player chooses which, if any, they will pay"; **508.1h** locks the total cost in; **508.1j** pays it; **508.1k** makes them attacking creatures; **508.1m** "Any abilities that trigger on attackers being declared trigger." **508.2** then gives the active player priority.
- **CR 607.2h:** a triggered ability printed in the same paragraph as a static ability is linked to it, and "refers only to actions taken as a result of the static ability". **CR 603.11** describes the paragraph form.
- **CR 502.3:** effects can keep permanents from untapping during the untap step.
- **CR 508.4** and **CR 508.3a:** a creature put onto the battlefield attacking was never declared as an attacker.
- **CR 702.20b:** attacking doesn't cause a creature with vigilance to tap.
- **CR 500.8:** extra phases are added directly after the specified phase.
- **CR 602.2b** and **CR 601.2h:** an activated ability's costs are paid like a spell's. **CR 605.1a**, **605.3b:** a mana ability doesn't use the stack.

The rulings (Amonkhet, 2017-04-18; Hour of Devastation, 2017-07-14; Sandstorm Crasher, 2024-11-08) settle the edges:

1. You can't exert a creature unless an effect lets you. "Tap it and it doesn't untap" (Decision Paralysis) is not exert.
2. If the exerted creature is already untapped during your next untap step (vigilance, or something untapped it), exert's effect expires having done nothing.
3. If you gain control of another player's creature until end of turn and exert it, it untaps during that player's untap step. The skip is keyed to **your** untap step (CR 701.43a), not to the creature's controller.
4. You exert as you declare the creature as an attacker. You can't exert later in combat, and a creature put onto the battlefield attacking can't be exerted. Triggers on exerting resolve before blockers are declared.
5. A creature whose "when you do" trigger targets can be exerted even with no legal target (Glorybringer).
6. "Whenever you exert a creature" (Battlefield Scavenger) triggers on that creature or any other creature you exert.
7. An activated ability's "Exert this creature" cost can be paid even if the creature was exerted earlier in the turn. Exerting it several times keeps it tapped through one untap step only (CR 701.43b).
8. Combat Celebrant: exerting several Celebrants in one combat gives that many additional combat phases, but each Celebrant can be exerted only once per turn. The additional phase happens even if Celebrant dies.

### The 36 cards

**Exert as it attacks (28).** Each is a static ability, usually with a linked "when you do" trigger.

| Shape | Cards |
|---|---|
| Self pump or keyword | Bitterblade Warrior, Emberhorn Minotaur, Glory-Bound Initiate, Gust Walker, Hooded Brawler, Khenra Scrapper, Nef-Crop Entangler, Rhet-Crop Spearmaster, Rhonas's Stalwart, Themberchaud |
| Shield | Oketra's Avenger |
| Board effect | Ahn-Crop Champion (untap all other creatures), Combat Celebrant (that, plus an extra combat), Tah-Crop Elite (team +1/+1) |
| Targeted | Ahn-Crop Crasher (can't block), Devoted Crop-Mate (reanimate), Glorybringer (4 damage), Hydra Trainer (+X/+X), Sandstorm Crasher (attacking token copy) |
| Card or mana advantage | Anep, Vizier of Hazoret (impulse two), Champion of Rhonas (creature from hand), Clockwork Droid (unblockable, scry 1), Watchful Naga (draw) |
| No linked trigger; "Whenever you exert a creature" payoff | Battlefield Scavenger (rummage), Resolute Survivors (drain 1), Rohirrim Chargers (fetch Equipment), Trueheart Twins (team +1/+0), Vizier of the True (tap a creature) |

**Exert as an activation cost (8).** Angel of Condemnation, Basri, Tomorrow's Champion, Fervent Paincaster, Hope Tender, Pride Sovereign and Steward of Solidarity have an activated ability; Arena of Glory (a land) and Oasis Ritualist have a mana ability.

No Commander-legal card says "if it was exerted" or "if it's exerted". The only condition on the record is Combat Celebrant's "if this creature hasn't been exerted this turn".

### What exists

- **The skip.** `Game.SkipNextUntapForEffect(cardID, player)` (`server/internal/game/untap.go`) appends one `UntapSkip{Player: player}` to `Card.NextUntapSkips`. A non-nil `Player` names that player's untap step, which is exactly ruling 3. `consumeUntapSkipsLocked` uses the marker up at that player's next untap step, whether or not the permanent was tapped, which is exactly ruling 2. The marker is cloned, snapshotted (`untapSkipSnapshot`), projected on the wire (`CardView.no_untap.next`) and dropped on a zone change. It is already used by `untap_restrictions.go`.
- **The duplicate rule.** A second identical marker for the same player is dropped. For Telekinesis that is a bug (#2029: "next two untap steps" counts as one). For exert it is the rule: CR 701.43b says every exert before your next untap step expires at that same step.
- **The declaration.** `declare_attacker` (one creature) and `declare_attackers` (a set) only stage an attack on `Card.AttackingTarget`. `commitAttackDeclarationLocked` (`attackers.go`, #859) locks the declaration in and emits one `EventAttack` per creature in one event batch, which is what keeps a whole declaration one occurrence for `OncePerBatch` triggers. Both verbs already carry a cost: ADR 0080's attack tax (`auto_tap`, `locked_sources`, `phyrexian_life`), paid by the verb.
- **The enumerator.** `legal.combatMoves` offers one move per (attacker, target), and the policy composes a declaration one move at a time.
- **The client.** The per-creature click and the context menu's "Declare attacker" send `declare_attacker`. "Attack with all" and "Choose attackers…" (`attackAll.ts`, `AttackDeclarationModal.svelte`, PR #2478) send `declare_attackers`.
- **The record.** `TurnTally.Attacks` records every declaration this turn by object and epoch (CR 400.7). Nothing records an exert.
- **The effects.** `preventFromSource` (Oketra's Avenger), untapping all your creatures, and additional combat phases (`extra_phases.go`, Relentless Assault) exist.

What is missing: a way to choose to exert, a record of the exert, an event to trigger on, and a cost component for activated abilities.

---

## Decision

### 1. The exert primitive

One engine function, `exertLocked(card, player, how)`, is the only way a permanent is exerted. It:

1. refuses a card that is not on the battlefield (CR 701.43c);
2. calls `SkipNextUntapForEffect(card, player)`, keyed to the **exerting player** (CR 701.43a, ruling 3);
3. appends an `ExertRecord{Object, Epoch, Player, PhaseID}` to a new `TurnTally.Exerts`;
4. emits `EventExert{Actor: player, CardID: card, Source: card, Target: defender}`. `Target` is the attack target when the creature was exerted as it attacked, and nil when it was exerted to pay an activation cost.

**Reuse ADR 0058's marker (the only option considered).** The marker already has the right key, the right expiry and the right duplicate rule. A second exert before the untap step adds nothing (CR 701.43b). An exerted creature with vigilance, or one already held tapped, still records the marker and it expires harmlessly (ruling 2). Seedborn Muse untaps an exerted creature during another player's untap step, correctly, because the marker names only the exerter's step.

**#2029's warning.** Whoever fixes #2029 must not make the marker count. Two exerts are one skip, while Telekinesis's "next two untap steps" is two. #2029's fix should add a count to its own entries (or a second kind of entry) and leave the plain marker, which exert uses, idempotent. A test in PR 1 pins it: exert the same creature twice, and it untaps at the exerter's second untap step.

### 2. Choosing to exert as it attacks

**The catalog declares it.** `effects.Spec.ExertOnAttack *ExertOnAttack` is the static ability "You may exert this creature as it attacks". It carries an optional `Unless func(g, source) bool` for Combat Celebrant's "if this creature hasn't been exerted this turn". The bridge reads it off the creature's **effective** abilities, so a creature that has lost its abilities (Humility) can't be exerted.

**The engine records the choice at the declaration and pays it at the lock-in.**

- **(a) Chosen (owner decision 2).** Both verbs take `exert: true` per attacker. The verb validates it (the creature has the ability now, and `Unless` is false) and stages it on `Card.ExertOnAttack`, beside `AttackingTarget`. `commitAttackDeclarationLocked` pays every staged exert (§1), then emits the `EventAttack`s, in one event batch. So the exerts are paid before the creatures become attacking (CR 508.1j before 508.1k), the triggers they cause are harvested with the attack triggers (CR 508.1m), and they go on the stack together at CR 508.2, before blockers (ruling 4). A creature re-pointed before the lock-in keeps its choice. A staged attack cleared before the lock-in (`clear_combat`, undo) never exerted anything. Once the declaration is locked in, the exert is paid, like the tap.
- (b) Pay at the verb, as ADR 0080 pays the tax. It is smaller. But the exert event would be emitted before the declaration finishes, so its triggers would not share the attack batch (the #859 bug, again), and clearing a staged attack would have to undo an exert.

**A refusal names its reason.** `exert: true` on a creature that can't be exerted refuses the whole action with `ErrCantExert`, also in the bulk verb. The bulk verb skips an *ineligible attacker* silently (#318), but an exert flag the creature doesn't have is a client bug, and silently dropping it would attack without the cost the player chose.

**A requirement never forces an exert (CR 508.1d).** A creature that must attack can attack without exerting, and the requirement check ignores the flag.

**Only declared attackers can be exerted.** A creature put onto the battlefield attacking never passes through the declaration (`stampEntryAttackerLocked`), so it can't be exerted (CR 508.4, ruling 4). Nothing extra is needed.

### 3. The triggers

**"When you do" (linked, CR 607.2h).** A helper, `effects.WhenExerted(row)`, builds a triggered ability on `EventExert` whose `CardID` is the source and whose `Target` is non-nil (the creature was exerted by its own "as it attacks" ability). That is how the trigger "refers only to actions taken as a result of the static ability": no catalog card exerts itself any other way. The trigger is an ordinary catalog row with a declared `Effect`, so it is stamped and restorable like any other (AGENTS.md, "abilities on the stack"). A targeted one with no legal target is removed from the stack as usual, and the creature stays exerted (ruling 5).

**"Whenever you exert a creature."** A helper, `effects.WheneverYouExert(row)`, triggers on any `EventExert` whose `Actor` is the source's controller and whose exerted permanent was a creature as it was exerted (Arena of Glory, a land, does not count). It fires for the creature itself, for any other creature, and for a creature exerted to pay a cost (ruling 6).

**"Hasn't been exerted this turn."** `Game.ExertedThisTurn(card)` reads `TurnTally.Exerts` by object and epoch, so a creature that left and came back is a new object (CR 400.7). It is Combat Celebrant's `Unless`. No card asks "if it was exerted", so nothing else reads the record yet.

### 4. Exert as an activation cost

- **(a) Chosen (owner decision 3).** A cost component, `Exert bool`, on `game.AbilityCost` and on `effects.ManaAbilityCost`, following ADR 0109's cost components. It is paid with the other costs (CR 602.2b, 601.2h) by calling `exertLocked(source, activator, cost)`. For Arena of Glory and Oasis Ritualist it is paid inside a mana ability (CR 605.3b), and any "whenever you exert" trigger waits for the next priority, as every trigger from a mana ability does. A creature already exerted can pay it again (ruling 7). The ability row on the wire says "Exert" in its cost, and the auto-tapper never activates an exert mana ability on its own, because an exert is a cost the player chooses. PR 4 flags such a source in the planner's `tapSource`, as `Sacrifices` flags a Treasure, so that only the player picks it. The 8 cards land in PR 4.
- (b) Exert as it attacks only. The 8 cards get their own registry row and wait.

### 5. The wire

- **Actions.** `declare_attacker` gains `exert` (bool). `declare_attackers` gains `exert` on each `attackers[]` entry. Both are absent from every payload built before this ADR, and false means what it always meant.
- **What the client may offer.** `LegalSourceView` (the legal-action digest) gains `exert_on_attack: true` on a creature that may be exerted as it attacks right now. The server computes it; the client never reads oracle text for it.
- **What happened.** The exerted creature's `no_untap.next` already names the exerter's untap step, so the board shows "won't untap" with no new field. The game log gets one line, "<player> exerted <card>", from `EventExert`.
- **`docs/protocol.md`** documents the two params, the digest field and the event kind.

### 6. The legal enumerator

- **(a) Chosen.** For an attacker with `exert_on_attack`, `combatMoves` offers two moves per target: the plain attack and the same attack with `exert: true`, labelled "Attack <target> with <card> and exert it (it won't untap during your next untap step)". A `Move`'s `Params` stays exactly the payload that performs it. The exert move is not gated on the linked trigger having a target (ruling 5).
- (b) One move with a follow-up choice. It would be the only attack move that needs a second round trip, and a policy that composes a declaration one move at a time would have to handle a choice in the middle of it.

The MCP seat and the model tiers see the twin move through the same list, and `boardtext` already renders `no_untap`.

### 7. The client

The choice is the player's (CR 508.1g), so no click path makes it for them.

- **(a) Chosen (owner decision 1).** Every path asks.
  - **A click on an exert creature** (ADR 0117's click to act) opens a two-button choice in the dock: "Attack" and "Attack and exert", with the linked trigger's text as the hint.
  - **The context menu** gets "Declare attacker and exert" beside "Declare attacker".
  - **"Choose attackers…"** gets an Exert toggle on each row whose creature has `exert_on_attack`, off by default.
  - **"Attack with all"** opens "Choose attackers…" instead of sending, when any eligible attacker has `exert_on_attack`, so the player sees the toggles. With no such creature it is one click, as now.
  - The labels go in `labels.ts` (ADR 0125 §2), and the attack hint learns about exert in the same PR.
- (b) The one-click paths never exert. Exert only through the picker and the context menu. Fewer clicks for a player who never exerts, but "Attack with all" silently chooses "no" for every exert creature.
- (c) (a), plus an ADR 0127 Ask / Always / Never preference per card for the exert question. It is the most convenient once a player knows their deck. It can follow later without changing (a).

### 8. The snapshot (schema v7, additive)

Three additions, recorded with `-update-shape`. No bump, and no fixture is touched.

1. `Card.ExertOnAttack` (`"exertOnAttack"`, omitted when false): a staged choice not yet paid. A restore point can be written between the verb and the lock-in, so it must be captured.
2. `TurnTally.Exerts` (`"exerts"`, omitted when empty): the turn's `ExertRecord`s.
3. The event kind `"exert"` in the event log.

The marker itself is the existing `nextUntapSkips` entry. An older v7 binary reading a newer file drops the two new keys. That is the rollback case: the worst outcome is that a staged exert from the instant of the restart is not paid, and Combat Celebrant could be exerted a second time that turn. A new `PendingChoiceKind` is not needed, so the closure ratchet is unchanged.

### 9. The heuristic bot (ADR 0126)

In PR 1 the heuristic never picks an exert move, so nothing changes for bots until PR 3 prices it. The `random` policy may pick one, which is legal.

- **(a) Chosen (owner decision 4).** Price it. Exert when `gain > cost`:
  - **gain**: the linked row's declared `purpose` (ADR 0126 §6), plus the purpose of every "whenever you exert a creature" row the bot controls. A self pump is priced by evaluating the attack again with the pumped stats. For that, `PurposeView` gains one additive field, `pump: {power, toughness, keywords}`, declared on the linked row. A trigger with no legal target is worth nothing.
  - **cost**: zero when the exert is free: the creature has vigilance, already won't untap during the bot's next untap step, or the attack is a lethal push. Otherwise the creature's attack value next turn plus its blocking value across the opponents' turns, since it stays tapped through them.
  - Combat Celebrant is priced by the extra combat: the bot's other untapped creatures' attack value in a second combat.
  - It is measured with ADR 0052's arena report, and must not lose to the S66 heuristic.
  - *As amended on 2026-10-07* ([below](#amendment-2026-10-07-exert-rows-and-what-they-declare)): the rows say they are exert rows, and `PurposeView` gains six fields, not one.
- (b) Simple rules. Exert only when free, on a lethal push, or when the trigger draws or makes a card. Smaller, and right most of the time. It undervalues pumps and Combat Celebrant.
- (c) Bots never exert. The 36 cards are worse in a bot's deck than printed. It is the right default for PR 1 only.

### 10. Multiple combat phases, and creatures that don't untap anyway

- **A second combat** (Combat Celebrant, Relentless Assault). An attacker in the new combat is a new declaration, so it can be exerted again (CR 701.43b): one more `ExertRecord`, one more `EventExert`, and the same single marker. Combat Celebrant's `Unless` stops Celebrant itself from being exerted twice in a turn. Each exerted Celebrant adds its own phase (CR 500.8, ruling 8), and the phase is added even if Celebrant has died, because the trigger is already on the stack.
- **A creature that won't untap anyway** (a stun counter, a static restriction, a hold, a marker already there) can still be exerted (CR 701.43b). The "when you do" trigger and the payoffs fire as usual. The marker expires at the exerter's untap step having done nothing (ruling 2). The heuristic counts the cost as zero (§9).
- **Vigilance.** The creature attacks untapped and stays untapped, so it can block during the opponents' turns. The marker expires unused unless something taps it first.
- **A borrowed creature.** Exerting a creature you control until end of turn names your untap step, so it untaps during its owner's (ruling 3), with no special case.

### 11. The catalog

- **PR 1** (with the engine): **Oketra's Avenger** and **Combat Celebrant** (the two waiting cards), **Glorybringer** (a targeted linked trigger, ruling 5) and **Resolute Survivors** (a "whenever you exert" payoff with no linked trigger). Together they cover every engine shape. Each lands `Full` if its full text is met.
- **PR 4**: the 8 activation-cost cards (§4, owner decision 3).
- **PR 5**: the other 24 attack-exert cards (owner decision 5). Each is checked against its full text at that PR. A card whose "when you do" needs a seam that doesn't exist yet (likely candidates: Rohirrim Chargers' reveal until an Equipment and attach it, Sandstorm Crasher's attacking token copy, Hydra Trainer's count of counters) goes on that seam's `Waiting` list, never shipped without it.

The `exert` registry row closes with PR 1, with a fragment in `docs/engine-seams/closed/`. `docs/adding-cards.md` gets an "Exert" subsection (both helpers, the `Unless` hook and the cost component) with a link in AGENTS.md §7.

---

## Tests

PR 1 adds Go tests for these:

1. Exert as it attacks records one marker keyed to the exerter, a `TurnTally.Exerts` entry and one `EventExert` in the same batch as the `EventAttack`s.
2. The creature doesn't untap at the exerter's next untap step and untaps at the one after.
3. A second exert before that step leaves one marker, and the creature untaps one step later, not two (CR 701.43b; the #2029 guard).
4. A vigilant creature, exerted, is untapped through the turn and the marker expires unused.
5. A borrowed creature, exerted, untaps during its owner's untap step.
6. A staged exert that is cleared, or undone, before the lock-in records nothing.
7. `exert: true` on a creature without the ability is refused, whole batch included.
8. A requirement to attack is met without exerting.
9. A creature put onto the battlefield attacking has no `exert_on_attack` and can't be exerted.
10. Oketra's Avenger's shield, Glorybringer with no legal target (exert still paid), Resolute Survivors on another creature's exert, and Combat Celebrant's extra combat, its untap, and its refusal on the second combat.
11. The enumerator offers the twin moves, and the heuristic picks neither exert move in PR 1.
12. The shape guard records the three additions. A restore point with a staged exert, and one with `Exerts`, restores and round-trips.

PR 2 adds vitest for the dock choice, the picker toggle and "Attack with all" opening the picker, and an e2e spec that exerts Oketra's Avenger. PR 3 carries the ADR 0052 arena report.

## Delivery

Each PR goes into `develop`, Sprint S68, Issue #2048.

| PR | What | Needs |
|---|---|---|
| 0 | **This ADR.** Docs only. | — |
| 1 | **Server: exert as it attacks.** §1–3, §5, §6, §8; the heuristic ignores exert moves; the four cards in §11; the registry row closed; `docs/protocol.md` and `docs/adding-cards.md`. | 0 |
| 2 | **Client: choosing to exert.** §7, labels and hints, vitest, an e2e spec, the nightly E2E on the branch. | 1 |
| 3 | **Bots: pricing an exert.** §9, the `pump` purpose on PR 1's cards, the arena report. | 1 |
| 4 | **Server: exert as a cost** (§4) and its 8 cards. | 1 |
| 5 | **Server: the other attack-exert cards** (§11). | 1 (3 for their `purpose`) |

PRs 1 and 2 should reach `main` in the same promotion, so a human player at a production table can always choose to exert a card the catalog marks `Full`. PRs 3, 4 and 5 can run in parallel after PR 1.

## Consequences

- Exert is one primitive. The attack path and the cost path differ only in when they call it.
- The attack declaration can carry an optional cost (CR 508.1g). Exert is the only one printed, so `ExertOnAttack` is a specific field rather than a general "optional attack cost" list. If another one is printed, it is a sibling field and one more branch at the lock-in.
- The untap marker's duplicate rule is now load-bearing for a rule (CR 701.43b), and #2029's fix must not change it.
- "Attack with all" is one click more for a player with an exert creature on the board (owner decision 1).

## Out of scope

- Granting exert to a creature that doesn't print it. No card does.
- Exerting outside combat or outside an activation cost. No card does (ruling 1).
- An ADR 0127 preference for the exert question (§7 (c)). Not now (owner decision 1); it can follow.

---

## Amendment (2026-10-07): exert rows and what they declare

PR 3 found two gaps in §9 before any code was written, and the owner answered both on 2026-10-07.

**1. The bot could not tell which rows are exert rows.** `ability_rows` carried each row's kind, label and purpose, and nothing said a triggered row was a linked "when you do" or a "whenever you exert" payoff. The owner chose (a): the two helpers stamp their rows. `effects.WhenExerted` sets a new `game.TriggeredAbility.Exert` field to `"linked"` and `effects.WheneverYouExert` sets it to `"payoff"`. The field reaches the wire as `AbilityRowView.exert`, absent on every other row. It is catalog data set by the helper that declares the row, not text parsing, and the engine never reads it. Rejected: (b) treating a row that declares `pump` as the linked row, which finds no payoff rows and no linked row that isn't a pump.

**2. No purpose field described the gains of PR 1's four cards.** Oketra's Avenger's shield, Glorybringer's 4 damage, Combat Celebrant's extra combat and Resolute Survivors' drain had no field, so §9's "one additive field" would have priced three of the four at zero. The owner chose (a): six additive fields on `Purpose` and `PurposeView`, each declared by hand on the card, the way `death_payoff` is. This widens §9's "one additive field, `pump`" to these six:

| Field | Meaning | Declared on |
|---|---|---|
| `pump` | `{power, toughness, keywords}` the row gives its own source until end of turn | the PR 5 self-pump cards |
| `extra_combat` | additional combat phases it adds | Combat Celebrant 1 |
| `prevent_combat_damage_to_self` | it prevents all combat damage that would be dealt to its source this turn | Oketra's Avenger |
| `damage_to_creature` | damage it deals to one target creature | Glorybringer 4 |
| `damage_each_opponent` | damage it deals to each opponent | Resolute Survivors 1 |
| `life_gain` | life its controller gains | Resolute Survivors 1 |

None of these existed on `Purpose` under another name. `DiscardPayoff.DamageEachOpponent` is the same amount per discarded card, on a discard payoff only, so it is not reused; the top-level field takes its name. The registration guard refuses `pump` and `prevent_combat_damage_to_self` off a triggered or activated row (each is about the ability's own source), a `pump` that gives nothing, and a negative amount. Rejected: (b) `pump` and `extra_combat` only, and (c) `pump` alone.

The pricing sits behind `Config.PriceExert`, which is off in `BaselineConfig()`, so `heuristic-baseline` stays frozen (ADR 0126 §9). The arena gains a second arena-only contestant, `heuristic-noexert` (today's heuristic with `PriceExert` off), so the exert pricing can be measured alone, and a synthetic deck, `exert-battle`, because no curated deck holds an exert card (ADR 0126 §8). The numbers are under [Measurements](#measurements).

## Measurements

### PR 3: pricing an exert (2026-10-07)

`DefaultConfig()` gains `PriceExert` on and `ExertCostWeight` 1.00; `BaselineConfig()` turns both off. No other weight moved.

**Arena.** No curated deck holds an exert card, so the run uses the synthetic `exert-battle` deck: `BattleDeck`'s curve in red and white with 4 Oketra's Avenger, 4 Combat Celebrant, 3 Glorybringer and 4 Resolute Survivors. Two `heuristic` seats against two `heuristic-noexert` seats (today's heuristic with `PriceExert` off), every seat on `exert-battle`: `boteval arena --seats heuristic,heuristic,heuristic-noexert,heuristic-noexert --decks exert-battle,exert-battle,exert-battle,exert-battle --games 96 --rotate --seed N`, for seeds 1, 97 and 193. 288 games, 0 stalls, 0 rejected moves, turns p50 14.

| Seeds | `heuristic` | Wilson 95% interval | `heuristic-noexert` | Wilson 95% interval |
|---|---|---|---|---|
| 1–96 | 45 of 192, 23.4% | 18.0%–29.9% | 51 of 192, 26.6% | 20.8%–33.2% |
| 97–192 | 54 of 192, 28.1% | 22.2%–34.9% | 42 of 192, 21.9% | 16.6%–28.2% |
| 193–288 | 57 of 192, 29.7% | 23.7%–36.5% | 39 of 192, 20.3% | 15.2%–26.6% |
| pooled | 156 of 576, 27.1% | 23.6%–30.9% | 132 of 576, 22.9% | 19.7%–26.5% |

The pricing does not lose to the heuristic without it. Each block's `heuristic` upper bound stays above 25%, the bar ADR 0126 §8 sets for a sub-PR, and two of the three blocks point clearly its way. Pooled, it does not clear 25% on its own.

Exert attacks (the Cards section's new `exert` action), pooled:

| Card | `heuristic`: games used of offered | windows taken of offered | `heuristic-noexert` |
|---|---|---|---|
| Glorybringer | 219 of 318 | 423 of 1678 | never |
| Resolute Survivors | 124 of 255 | 201 of 1733 | never |
| Oketra's Avenger | 114 of 272 | 171 of 1923 | never |
| Combat Celebrant | 62 of 275 | 84 of 1219 | never |

**Curated decks.** None holds an exert card, so no attack there has an exert twin. `exertTwin` finds nothing, and `decideAttack` and `raceAttack` choose exactly as before. The S66 runs are not repeated.

**Suite.** `boteval suite run --policy heuristic`: 37 of 37 agree, every tag at 100%. No position was added.

## Questions for the owner (answered)

These are the questions as asked. The owner chose the recommended option, (a), for each one on 2026-10-07; see [Owner decisions](#owner-decisions-2026-10-07).

1. **How a player chooses to exert (§7; CR 508.1g).**
   - **(a) Recommended:** every path asks. A click opens "Attack" / "Attack and exert" in the dock, the context menu has both rows, "Choose attackers…" has an Exert toggle per row, and "Attack with all" opens the picker when an exert creature could attack. Nothing chooses for the player, at the price of one more click for an alpha strike with an exert creature on the board.
   - (b) The one-click paths never exert, and exert is chosen only in the picker or the context menu. Faster, but "Attack with all" quietly answers "no" for the player.
   - (c) (a), plus an Ask / Always / Never preference per card now (ADR 0127). Most convenient, more work in PR 2.
2. **When the exert is paid (§2; CR 508.1j, 508.1m).**
   - **(a) Recommended:** chosen at the verb, paid at the declaration's lock-in, with the attacks in one event batch. The triggers go on the stack with the attack triggers, and a cleared or undone attack exerted nothing.
   - (b) Paid by the verb, like the attack tax. Smaller, but the exert triggers miss the attack batch, and clearing a staged attack has to undo an exert.
3. **Exert as an activation cost (§4; CR 602.2b).**
   - **(a) Recommended:** in scope. A cost component on activated and mana abilities, and 8 more cards in PR 4. One primitive serves both.
   - (b) Attack exert only. The 8 cards wait on a new registry row.
4. **How the heuristic bot decides (§9; ADR 0126).**
   - **(a) Recommended:** price it. Gain from declared purposes (with a new `pump` field), cost as the creature's next-turn attack and block value, zero when free. Measured in the arena. It plays the cards as well as the heuristic can, and is the most work.
   - (b) Simple rules: exert only when free, on a lethal push, or for card advantage. Smaller, and it undervalues pumps and Combat Celebrant.
   - (c) Bots never exert. Nothing to build, and the cards are weaker in a bot's hands than printed.
5. **How many cards (§11).**
   - **(a) Recommended:** the four proof cards in PR 1, then every other exert card in PRs 4 and 5 whose text the engine can meet, each checked; the rest go on their seam's `Waiting` list. The mechanic is useful the day it lands.
   - (b) Only Oketra's Avenger and Combat Celebrant, the two waiting cards. The rest stay on the backlog for deck requests to pull in.
