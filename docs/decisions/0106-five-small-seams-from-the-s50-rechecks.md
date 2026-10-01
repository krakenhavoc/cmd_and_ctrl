# ADR 0106 — Five small seams from the S50 re-checks

**Status:** Proposed · 2026-10-01 · S50 — Seams from the deck re-checks (tracker [#1784](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1784))
**Issues:** [#1793](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1793) (abilities any player may activate), [#1794](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1794) (a creature that can't attack its owner), [#1805](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1805) (evolve), [#1806](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1806) (a permanent that makes your other spells uncounterable), [#1807](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1807) (targets from a single graveyard).
**Numbering:** checked with the AGENTS.md §4 sweep on 2026-10-01. I ran `git fetch --all --prune` and read every `docs/decisions/` file name on all 36 remote heads (`origin/develop`, `origin/main` and 34 feature, chore, docs, fix, repro and wip branches) and all 522 local branches, and listed every name ever committed on any of them (`git log --all --name-only -- docs/decisions/`). The highest number anywhere is 0105. No open PR carries an ADR. This one takes **0106**.
**Builds on:** [ADR 0033](0033-ai-bot-seat.md) §1 (the legal-move enumerator), [ADR 0105](0105-legal-action-highlights.md) (the `legal_actions` digest), [ADR 0102](0102-entering-under-another-players-control.md) (Xantcha's entry), [ADR 0104](0104-gaining-control-of-a-spell.md) (a spell's controller and its caster), [ADR 0045](0045-combat-restrictions.md) and its 2026-09-24 amendment (attack requirements), [ADR 0080](0080-attack-taxes.md), [ADR 0014](0014-combat-keywords.md)'s 2026-09-24 amendment (prowess, the first keyword trigger), [ADR 0019](0019-structured-targeting.md)'s 2026-09-24 amendment (the target set rule), [ADR 0072](0072-protection.md)'s 2026-09-22 amendment (player statics), [ADR 0093](0093-abilities-granted-to-other-permanents.md) (ability refs), [ADR 0009](0009-smart-priority-autopass.md) (autopass) and [ADR 0041](0041-game-persistence.md) (restore points).

This ADR was written plan-first. No engine or card code changes until the owner answers the questions at the end.

---

## Context

The S50 deck re-checks and the discover batch (#1112) left five small seams. Each blocks one named card, and each is a single rule the engine does not yet express. They are grouped here because they are small, and because two of them (#1793, #1794) are the last two things Xantcha, Sleeper Agent waits on.

The seams doc runs stale, so every claim below was checked in the code on `origin/develop` at `79a87b57`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260819.txt`, effective August 7, 2026). Card counts come from the Scryfall dump of 2026-09-24, counted by oracle ID, Commander-legal only.

Two things the registry says are wrong, and the PRs below correct them:

- The `evolve` row says "CR 702.100b compares the entering creature's power and toughness". It doesn't. The comparison is in **CR 702.100a**, and it is checked twice because of **CR 603.4** (the intervening "if"). 702.100b defines "evolves".
- The `cant-be-countered-grant` row says a static can't reach a spell "because the layer system does not apply to objects on the stack (ADR 0012)". Since ADR 0104 the layer pass does have a stack step, for layer 2. But that is beside the point. "Spells you control can't be countered" is not a layer effect at all (§4).

---

## 1. Abilities any player may activate (#1793)

### What exists, what is missing

- **The activation path refuses a non-controller.** `activateCatalogAbilityLocked` (`game/activated.go`) returns `ErrCardCallerMismatch` when `source.Controller != playerID` on the battlefield. `actions.Dispatch` does not gate `activate_ability` on the controller, so this is the only refusal.
- **Everything after that check already takes the activator.** The board-wide gate (`ActivationGateLocked`), the timing read (`ActivationTimingOpenLocked`), `ab.Condition(g, playerID, cardID)` and the cost payment all take `playerID`. The stack item is pushed with `Controller: playerID`. So "you" in the effect is already the activator, and "Xantcha's controller" is already `ctx.SourcePermanent().Controller` (CR 608.2h last-known information if Xantcha has left).
- **The enumerator skips other players' permanents.** `activatedMoves` (`legal/abilities.go`) does `if source.Controller != e.seat { continue }`.
- **The view stamps rows for the controller only.** `ActivatedAbilityView`'s `condition_unmet` and `timing_closed` are "evaluated once with the ability's CONTROLLER as 'you' and sent to every viewer".
- **The client shows nothing on a permanent you don't control.** `canOverride` (`contextMenu.logic.ts`) gives a non-controller an empty menu, and the click router answers `"none"`.
- **`ActivatedAbility` has no field for who may activate.**

### The rules

- **CR 602.2:** "Only an object's controller (or its owner, if it doesn't have a controller) can activate its activated ability unless the object specifically says otherwise."
- **CR 602.1b:** text after the colon "may state which players can activate that ability … This text is not part of the ability's effect. It functions at all times."
- **CR 602.1a:** "An ability's activation cost must be paid by the player who is activating it."
- **CR 602.2a / 113.8:** the ability's controller on the stack "is the player who activated it". **CR 109.5:** for an activated ability, "you" "is the player who activated the ability".
- **CR 602.5b:** a restriction on use ("Activate only once each turn") "continues to apply to that object even if its controller changes".
- **CR 113.7a:** once activated, the ability exists independently of its source.
- Xantcha's rulings (2018-07-13, quoted in #1793): whoever activated it draws, and Xantcha's controller may activate it too.

### Options

- **A. A declaration on the ability (recommended).** `ActivatedAbility.AnyPlayer bool`, mirrored on `game.ActivatedAbilityShape`. The controller check becomes "the controller, or anyone if `AnyPlayer`". Everything downstream already reads the activator. This is CR 602.1b exactly: the permission is part of the ability, so a copy of the ability or a layer-6 grant of it carries the permission with it, and an ability-removal effect removes it.
- **B. A per-player permission** ("player P may activate abilities of permanent X"), in the shape of a cast permission. It fits no printed card. Every one says it of its own ability, and a separate record could outlive or miss the ability it describes.

### Decision (recommended: A)

1. **The declaration.** `AnyPlayer bool` on the catalog row. Register refuses it on a mana ability (the mana path is separate, and Mana Cache, the one legal any-player mana ability, also needs "only during their turn before the end step"). Register also refuses it with a `{T}`, `{Q}`, loyalty, crew or self-sacrifice cost component. No printed any-player card has one, and who may tap another player's permanent would be a rule nobody has tested.
2. **The gate.** One function, `mayActivateLocked(player, source, ab)`, is used by the activation path, the enumerator and the view. On the battlefield it answers `source.Controller == player || ab.AnyPlayer`. Off the battlefield the CR 108.4a owner check is unchanged, because no printed any-player ability functions from a hidden zone. Arrest's `CanActivateAbilities` still applies to every player: it restricts the object, not the activator.
3. **What the activator does.** Costs, X, targets and modes are the activator's (CR 602.1a, 601.2c via 602.2b). The stack item's `Controller` and `BaseController` are the activator (CR 113.8). The once-per-turn and exhaust tallies stay keyed by object and label, not by player (CR 602.5b).
4. **The enumerator.** `activatedMoves` visits every battlefield permanent. For one the seat does not control, it offers only `AnyPlayer` rows, through the same body (`abilityMovesForSource`), so the timing, condition, gate and cost checks are the activator's.
5. **The digest and the view.** ADR 0105's digest needs no change. It is built from the seat's own move list, and keyed by source instance ID, so another player's Xantcha appears under the activator's key with `abilities: ["own:0"]`. The digest stays "own seat only": it describes the viewer's moves, not the permanent's controller's. `ActivatedAbilityView` gains `any_player: true`. The shared row's stamps stay the controller's. A non-controller's verdict (is this row usable by *me* right now) comes from the digest (`legalGate`), and the activator's charged price rides the per-seat carrier the way cast stamps do.
6. **The client.** A permanent with an `any_player` row is clickable by every seat. The click opens the ability popover, which lists only the `any_player` rows for a non-controller. The rows are enabled exactly when the digest lists their ref. ADR 0105's bolt pip appears on an opponent's permanent when the digest says so. The log line names both players: "Bob activated Alice's Xantcha, Sleeper Agent".
7. **Autopass and the bot** are open questions 1 and 2. An any-player ability on the table is a legal move for every seat at every priority window. Without a decision, smart autopass would hold every seat with three mana open (`hasResponse` counts any activate move in `legal_moves`), and the heuristic would pump an opponent's Flailing Ogre because `ActivateBase` is positive for any activation.

### Snapshot impact

None. The flag is catalog data and is not serialised. The stack item already records its controller and its `AbilityRef`, and restore rebuilds the effect from the catalog row.

### Cards

Xantcha, Sleeper Agent, once §2 lands too: **Full**. 38 other Commander-legal cards print "Any player may activate this ability"; none is catalogued. Open question 6 asks how many join.

---

## 2. A creature that can't attack its owner (#1794)

### What exists, what is missing

- **Requirements exist; per-target restrictions do not.** `Characteristic.AttackRequirements` carries CR 508.1d requirements, and `AttacksEachCombat()` writes Xantcha's first half. The CR 508.1c restrictions are a `Restriction` bit (can't attack at all), count limits (`AttackLimit`) and taxes (`AttackTax`). None names a defender by its relation to the attacker.
- **Target legality ignores the attacker.** `canAttackTargetLocked(attackerController, target)` and `AttackTargetsForEffect(attackerController)` take only the controller. The declaration verbs (`mutations.go`), the CR 508.1d reach search (`attackRequirementCandidatesLocked`), the enumerator (`legal/combat.go`) and the view all list targets per seat, not per creature.
- **The client already gates per attacker.** Since ADR 0105 sub-PR 5, a click on a defender is checked against the selected attacker's digest `attack_targets`. A narrower enumerator reaches the client with no client change to the gate itself.

### The rules

- **CR 508.1c:** "If any restrictions are being disobeyed, the declaration of attackers is illegal."
- **CR 508.1d:** requirements are counted only "without disobeying any restrictions". So "attacks each combat if able" yields: with only its owner and the owner's planeswalkers to attack, Xantcha does not attack. A battle the owner protects is still a legal target, and attacking it obeys the requirement.
- **CR 508.4c:** a creature "put onto the battlefield attacking … isn't affected by requirements or restrictions that apply to the declaration of attackers". **CR 508.7b:** the same while reselecting what a creature attacks.
- **CR 108.3 / 111.2:** the owner is who started the game with the card, or a token's creator. A Clone copying Xantcha is owned by the Clone's owner.

### Options

- **A. A data restriction on the characteristic (recommended).** `Characteristic.AttackTargetRestrictions []AttackTargetRestriction{Source, SourceName, NotOwner, NotOwnersPlaneswalkers}`, written by an ordinary layer static like `AttackRequirements`, and only ever appended to. One predicate, `canAttackTargetWithLocked(attacker, target)`, is the old controller check plus this list. Every declaration-time caller moves to it.
- **B. A `Restriction` bit, `CantAttackOwner`.** It is smaller, but a bit cannot say "and planeswalkers its owner controls", which Xantcha has and Alexios does not. It also cannot be granted by a resolved effect (Elrond's "This creature can't attack its owner"), which ADR 0041 phase 3 says must be data.

### Decision (recommended: A)

1. **The predicate.** `canAttackTargetWithLocked(attacker *Card, target)` refuses an attack on `attacker.Owner`, and, with `NotOwnersPlaneswalkers`, on a planeswalker the owner controls. Battles are never refused by it. The owner is read live, so a copy uses its own owner.
2. **Who calls it.** Both declaration verbs, `attackRequirementCandidatesLocked` (so CR 508.1d's maximum is computed over restriction-legal targets), and the enumerator's attack loop, which walks a new per-attacker `AttackTargetsForAttackerForEffect`. The digest's `attack_targets` is built from those moves, so it follows.
3. **Who does not call it.** `attack_reselect.go` (CR 508.7b) and the "put onto the battlefield attacking" path (CR 508.4c) keep the controller-only check, because the rules exempt them.
4. **The view.** `Turn.AttackTargets` stays per seat: it prices taxes and limits, which are per target. The per-attacker answer is the digest's. The card gets a restriction chip naming the owner (open question 3).
5. **Ability removal.** Xantcha's own static is applied only while `CatalogAbilityKey` answers (CR 613.1f), exactly as its "attacks each combat" requirement is. A creature that loses all abilities may attack its owner.

### Snapshot impact

`AttackTargetRestrictions` reaches the snapshot only through `lastKnownBattlefield`, as `AttackRequirements` does. It is additive. The shape file is updated with `-update-shape` under the current schema version.

### Cards

Xantcha (with §1). Alexios, Deimos of Kosmos still waits on "can't be sacrificed", which the engine does not have. Elrond of the White Council waits on secret council and on a `ScopedEffect` mod kind that grants this restriction. Neither is in scope.

---

## 3. Evolve (#1805)

### What exists, what is missing

- **Nothing for evolve.** It is not in `canonicalKeywords` (`game/keywords.go`), so the deck importer drops Scryfall's "Evolve" and no trigger exists.
- **The model to copy exists.** Prowess (`game/prowess.go`) is a canonical token. `keywordTriggersFor` turns each token into one engine `TriggeredAbility` with a keyed body (`prowess/pump`), so a printed, token-template or layer-6 granted prowess all work without a catalog entry, and a table with one on the stack is a restore point.
- **Last-known information exists.** `PermanentForEffect` and `ctx.TriggeringPermanent()` read the entered creature live while it is on the battlefield and as it last existed once it has left (#1379).
- Dinosaur Egg is catalogued with the caveat "Evolve isn't implemented".

### The rules

- **CR 702.100a:** "Evolve is a triggered ability. 'Evolve' means 'Whenever a creature you control enters, if that creature's power is greater than this creature's power and/or that creature's toughness is greater than this creature's toughness, put a +1/+1 counter on this creature.'"
- **CR 702.100b:** "A creature 'evolves' when one or more +1/+1 counters are put on it as a result of its evolve ability resolving."
- **CR 702.100c:** "A creature can't have a greater power or toughness than a noncreature permanent."
- **CR 702.100d:** "If a creature has multiple instances of evolve, each triggers separately."
- **CR 603.4:** the intervening "if" is checked when the event occurs and again on resolution; if false then, the ability does nothing.
- **CR 603.6a:** permanents entering together see each other enter.
- **CR 608.2h:** an object no longer in its zone is read by its last-known information.

### Options

- **A. An engine keyword trigger, like prowess (recommended).** `evolve` joins `canonicalKeywords` and is cumulative (702.100d). `keywordTriggersFor` returns one evolve trigger per token. A printed, granted (Tyranid Prime, Propagator Drone) or token-carried evolve works with no catalog entry.
- **B. A catalog constructor, `effects.Evolve()`, on `Spec.Triggered`.** Each evolve card needs an entry, and a granted evolve reaches nothing. This is less faithful to how the cards are printed and granted.

### Decision (recommended: A)

1. **The trigger.** It watches `EventETB`. `AppliesTo` holds when the entered permanent is a creature under the source's controller, the source is a creature, and the entered creature's power or toughness (counters included, among them any it entered with) is greater than the source's. Comparing with a noncreature on either side is false (702.100c).
2. **The resolution check.** A keyed body `evolve/grow` re-runs the same comparison (603.4). It reads the entered creature through the item's carried trigger object, live or last-known (608.2h). It reads the evolving creature live, and does nothing if that object has left or changed (CR 400.7, the prowess rule). If the check holds, it puts one +1/+1 counter through the ordinary counter path, so Hardened Scales and "can't have counters" apply.
3. **"Evolves".** The body emits `EventEvolved{CardID}` when at least one counter was actually put on (702.100b). Renegade Krasis and Watchful Radstag watch it.
4. **Commuting.** Evolve triggers do not commute the way prowess triggers do. Each one changes the power and toughness another one's condition reads, so ordering is a real choice (CR 603.3b) and the prompt is kept.

### Snapshot impact

No new fields. `evolve/grow` is a new keyed body: an on-disk identity, never renamed. An older binary refuses a file naming it with `ErrUnknownEffectKey`, which is the designed rollback case.

### Cards

- Dinosaur Egg drops its caveat: **Full**.
- Seven creatures whose only text is evolve plus canonical keywords play with no catalog entry: Adaptive Snapjaw, Battering Krasis, Clinging Anemones, Cloudfin Raptor, Crocanura, Gluttonous Slug and Shambleshark.
- 14 more Commander-legal evolve cards need entries for their other lines, and Propagator Drone grants evolve. Open questions 5 and 6 decide how many land.

---

## 4. A permanent that makes your other spells uncounterable (#1806)

### What exists, what is missing

- **One gate, two sources.** `spellCantBeCounteredLocked` (`game/cant_be_countered.go`) reads the spell's own printed rider (`CatalogCantBeCountered`) and the mana that paid for it (#1547). Every counter verb goes through it, and the view's chip (#1553) reads it.
- **No source on the battlefield.** Chimil, the Inner Sun is catalogued with the caveat "Chimil doesn't stop your spells from being countered."
- **The pattern for the rest exists.** `Player.Statics` holds abilities a player has for a duration (Teferi's Protection, Emergence Zone's timing). Its rule is that a *derived* statement is read off the battlefield on every query, and only a *granted-for-a-duration* one is stored. `PermissionFilter` is a data spell filter already used by cast bans and timing statements.

### The rules

- **CR 613.11:** "Some continuous effects affect game rules rather than objects. … These effects are applied after all other continuous effects have been applied."
- **CR 611.3a:** a static's effect "isn't 'locked in'; it applies at any given moment to whatever its text indicates."
- **CR 611.2c:** a resolving spell's effect that does not modify characteristics or control "modifies the rules of the game, so it can affect objects that weren't affected when that continuous effect began." Veil of Summer covers a spell cast after it resolved.
- **CR 101.2:** "can't" beats "can". **CR 701.6a:** to counter is to remove from the stack.
- ADR 0104's split: "Spells you **control**" is the spell's current controller; "spells you **cast**" (Cunning Nightbonder, Thryx) is the caster (`StackItem.BaseController`, zero meaning `Controller`).

So a "can't be countered" static is a rule-modifying effect, read when something tries to counter. It is not a characteristic of the spell, and it does not belong in the layer pass.

### Options

- **A. A third and fourth source at the gate (recommended).** The derived source is a catalog declaration on a permanent, read off the battlefield at the gate. The granted source is a `PlayerStatic` payload for "this turn" grants. The gate is CR 613.11's own shape, and it is the gate every counter verb already uses.
- **B. Stamp a "can't be countered" bit on stack items in the layer pass's stack step.** It treats a rules effect as a characteristic (wrong under 613.11). It widens ADR 0104's layer-2-only stack step for no rule's sake. It still needs a separate answer for "you cast".

### Decision (recommended: A)

1. **The derived source.** `Spec.SpellsCantBeCountered []CounterShield{Label, Whose, Spell}`:
   - `Whose` is `YouControl` (the zero value, and the narrower reading), `YouCast` or `AnyPlayer` ("Creature spells can't be countered", Gaea's Herald).
   - `Spell` is a predicate over the spell card on the stack (creature, green, power 5 or greater, Dragon).

   It is read through an accessor that drops a shield when `CatalogAbilityKey` answers empty (CR 613.1f) or the permanent is phased out, as `AttackTaxesForCard` does.
2. **The granted source.** A `PlayerStatic` payload, `CantBeCountered{Filter PermissionFilter, Others bool}`, with a CR 611.2 duration, for "Spells you control can't be countered this turn" (Veil of Summer) and "Other spells you control …" (Bound // Determined: the resolving spell itself is excluded). It is in scope only if open question 4 says so.
3. **The gate.** `spellCantBeCounteredLocked` asks all four sources. The view's chip and every counter verb follow with no change.
4. **What it doesn't do.** "Can't be countered" does not make a spell an illegal target. Counterspell can still target it and does nothing on resolution. That is already how the printed rider behaves.

### Snapshot impact

The derived source is catalog data, so none. The granted source, if in scope, adds one additive field to `seats[].statics[]`, recorded with `-update-shape` under the current version.

### Cards

Chimil drops its caveat: **Full**. About 20 Commander-legal permanents print a "… spells (you control) can't be countered" static, and three resolving cards grant it for a turn. None but Chimil is catalogued. Open questions 4 and 6 set the scope.

---

## 5. Targets that must come from a single graveyard (#1807)

### What exists, what is missing

- **The difference rule exists; its opposite does not.** `TargetSpec.Different *TargetDifference` (`game/target_set.go`) is a key no two picks may share. It has four readers: the announce gate, the CR 608.2b re-check (`setRuleConflictLocked`), the enumerator (`legal/cast.go`) and the client picker (`targeting.ts`, from the view's `different: {label, keys}`). It also feeds `fillableCountLocked` and the retarget narrowing (`withoutSetRuleConflictsLocked`).
- **"From a single graveyard" needs every pick to share a key:** the graveyard's owner.

### The rules

- **CR 601.2c** (via 602.2b for abilities): the player announces the targets, and the chosen set must meet the targeting text.
- **CR 400.3 / 404.1:** a card in a graveyard is always in its owner's graveyard, so the key is the card's `Owner`.
- **CR 608.2b:** each target is re-checked at resolution. A card that left and came back is a new object and illegal on its own.
- **CR 115.7d / 115.7e:** when choosing new targets, a changed target "must not cause any unchanged targets to become illegal", and "only the final set of targets is evaluated". So moving every target to another graveyard at once is legal.

### Options

- **A. A sibling set rule, `TargetSpec.Same *TargetSameness{Label, Key}` (recommended).** The same four readers each gain a "share" branch. The wire gains `same: {label, keys}` beside `different`. It is additive, and no existing clause changes.
- **B. Generalise `TargetDifference` into one `TargetSetRule` with a kind.** It means the same work plus renaming the type, the wire field and three catalog constructors, with no behaviour gained.
- **C. Ask for a graveyard first, then for cards from it.** The graveyard's owner becomes an extra announced choice that the card does not print. The graveyard's owner is not a target, and nothing printed asks anyone to choose a graveyard. It also does not match CR 115.7e's retarget, where only the final set is judged. It is less faithful.

### Decision (recommended: A)

1. **The constructor.** `effects.FromASingleGraveyard()` returns a `TargetSameness` keyed on the card's owner, labelled "come from a single graveyard".
2. **The four readers:**
   - The announce gate refuses a set with two keys.
   - The re-check judges the surviving picks. With an owner key they cannot diverge, but the branch is written generally, with the weaker reading `setRuleConflictLocked` already uses.
   - The enumerator builds sets within one key group.
   - The picker greys candidates outside the first pick's group.
3. **Counts.** `fillableCountLocked` answers the largest key group, so an exact "four target cards" (Pestilent Cauldron) is not offered when no graveyard holds four.
4. **Retargeting.** The offer narrows against the keys of slots that stay. `retargetCheckLocked` judges the final set (115.7e), so changing every slot to another graveyard is accepted.
5. **Not in scope.** Costs that exile or move "two cards from a single graveyard" (Night Soil, Jötun Grunt) are costs, not targets.

### Snapshot impact

None. The clause is rebuilt from the catalog row on restore. A `pick_target` pause carries no live frame after a restore, and the announce gate still enforces the rule on the answer, as it does for `Different`.

### Cards

Digsite Conservator: **Full** (its discover half is ADR 0099's). 24 other Commander-legal cards target "from a single graveyard", and two more pay a cost from one; none is catalogued. Several need nothing else, for example Decompose, Carrion Beetles, Rag Dealer, Famished Ghoul, Griffnaut Tracker and Scarab Feast (open question 6).

---

## Shared machinery

- **Xantcha needs §1 and §2 together.** The card lands only in the second of the two PRs, and then as **Full**. Control comes from ADR 0102's entry (control: an opponent), ownership from CR 108.3 (owner: the player whose deck it came from). §1 reads "may this player activate", where the controller is just one of the players who may. §2 reads the owner, never the controller.
- **The enumerator is the source of truth for both.** Per ADR 0033 §1 and the #544 rule, every gate is one function shared by the engine verb and `internal/legal`: `mayActivateLocked` for §1 and `canAttackTargetWithLocked` for §2. ADR 0105's digest is built from the enumerator, so the client's ready pips (§1) and per-attacker attack rings (§2) follow without a second rule in the client.
- **The client's popover on another player's permanent** (§1) is the one new client surface. §2's client work is a chip.
- **Snapshots.** Only §2 (one `lastKnownBattlefield` field), §3 (one keyed body) and optionally §4 (one player-static field) touch the snapshot. All three are additive or keyed within the current schema version, and none adds a closure route.

## Delivery

Each PR lands its engine change and its cards together, test first. Each also flips its registry row to implemented, adds a closed-seam fragment under `docs/engine-seams/closed/`, and corrects the row's stale notes (see Context). They are independent except where noted.

1. **Single graveyard (#1807).** The sameness rule, its four readers, the wire field and the picker. Digsite Conservator (Full), plus whichever plain exilers open question 6 admits.
2. **Uncounterable (#1806).** The derived shield, the granted payload if open question 4 says so, and the gate. Chimil (Caveats → Full), plus the statics question 6 admits.
3. **Evolve (#1805).** The token, the trigger, `evolve/grow`, `EventEvolved`, and the keyword-only creatures. Dinosaur Egg (Caveats → Full), plus the catalog cards questions 5 and 6 admit.
4. **Can't attack its owner (#1794).** The restriction, the predicate and its callers, the per-attacker target list and the chip. No card lands yet: Xantcha also needs PR 5.
5. **Any player may activate (#1793).** The flag, the gate, the enumerator, the view field, the client popover, and the autopass and bot rules from open questions 1 and 2. **Xantcha, Sleeper Agent** lands here (Full), plus whichever any-player cards question 6 admits.

PRs 1–3 are small and touch nothing in combat or priority, so they can go in parallel. PR 4 goes before PR 5 so that Xantcha can land in PR 5.

## Consequences

- An activated ability is no longer always its controller's to activate. Any new code that asks "may this player activate this" must go through `mayActivateLocked`.
- Attack target legality becomes a question about a creature, not just a seat. `AttackTargetsForEffect(seat)` stays for the per-target view (taxes, limits). Legality per attacker comes from the new per-attacker list.
- Evolve is the second engine-derived keyword trigger. Printed evolve creatures with no other text need no catalog entry.
- "Can't be countered" has four sources, read at one gate. It stays outside the layer pass, by CR 613.11.
- The target set rule gains its opposite, and every future "from a single …" or "that share a …" clause uses it.

## Out of scope

- Any-player **mana** abilities (Mana Cache), and any-player abilities on a spell (Lightning Storm).
- A ScopedEffect mod kind that grants "can't attack its owner" (Elrond), and "can't be sacrificed" (Alexios).
- "The next spell you cast this turn can't be countered" (Insist, Overmaster, Mistrise Village, Savage Summoning) and "Target spell can't be countered" (Vexing Shusher). These are a one-use promise and a mark on one stack object, each its own shape. They would go on the `cant-be-countered-grant` row's `Waiting` list.
- "Can't be countered by blue or black spells" (Autumn's Veil), which depends on the counterer.
- Costs from a single graveyard (Night Soil, Jötun Grunt).

---

## Open questions for the owner

1. **Autopass and another player's Xantcha (#1793).** While Xantcha is out, every seat with {3} open has a legal instant-speed activation at every priority window.
   - **(a) Recommended:** an ability on a permanent you don't control does not count as a response for smart autopass. You see its ready pip only on frames where you stop anyway.
   - (b) It counts like your own abilities, so every such seat stops at every window.
   - (c) A per-player setting that defaults to (a).
2. **The bot and other players' abilities (#1793).**
   - **(a) Recommended:** the heuristic activates an ability on a permanent it does not control only when the row declares a purpose for the activator. Xantcha's would be "you draw a card; its controller, an opponent, loses 2 life". Rows with no purpose are never chosen. The model tiers see the move labelled with the permanent's controller.
   - (b) Score it like the bot's own abilities. The generic `ActivateBase` is positive, so the bot would pump an opponent's Flailing Ogre or put counters on an opponent's Feral Hydra.
   - (c) Never activate another player's ability.
3. **Showing Xantcha's restriction (#1794).**
   - **(a) Recommended:** a restriction chip on the card, "Can't attack Alice", naming the owner, beside the existing greyed target ring.
   - (b) Only the greyed ring and the refusal toast.
4. **What makes spells uncounterable (#1806).**
   - **(a) Recommended:** the printed statics (Chimil and about 20 others, "you control", "you cast" and "any player" forms), **and** the "this turn" grants (Veil of Summer, Bound // Determined, Domri, Anarch of Bolas's +1) as a player static. Both are data, read at the same gate.
   - (b) Printed statics only. The "this turn" family waits on the row.
   - (c) (a) plus the one-use "next spell you cast" promise and "target spell can't be countered". These are two more shapes, each with its own duration and ending.
5. **How far evolve goes (#1805).**
   - **(a) Recommended:** the keyword, its trigger and the CR 702.100b "evolves" event in one PR. Granted evolve (Tyranid Prime, Propagator Drone) and "whenever this creature evolves" (Renegade Krasis, Watchful Radstag) work from day one.
   - (b) The keyword and trigger only. The "evolves" cards wait.
6. **How many cards each PR lands.**
   - **(a) Recommended:** the named waiting card, plus every other card of that seam that needs no other missing primitive. Each is verified against its full oracle text in the PR, and any that turns out to need more goes on a `Waiting` list with the reason. The pools are about 38 any-player cards, 24 single-graveyard cards, 15 evolve cards with other text (Propagator Drone included) and about 20 uncounterable statics.
   - (b) Only the named waiting cards (Xantcha, Digsite Conservator, Dinosaur Egg's and Chimil's caveats). The rest are listed on each row's `Waiting` list for later batches.
