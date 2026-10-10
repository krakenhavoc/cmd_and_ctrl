# ADR 0147 — Auto-assign combat damage, and assigning it on the blockers

**Status:** Accepted · 2026-10-10 · S59 — Automated table: clicks that act, payment that counts
**Issue:** [#2956](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2956).
**Owner decisions:** the issue's own: a gameplay setting, "Auto-assign combat damage", **on by default**; when the damage covers lethal for every blocker, no prompt; otherwise a prompt pre-filled with the canonical split, with damage ticked on the blockers themselves; `ResolveDamageAssignment` stays the authority. The issue left one question open, client or server, and this ADR answers it.
**Numbering:** after `git fetch --all --prune`, every ADR file on any remote ref (`git log --remotes --name-only -- docs/decisions/`) and on every local branch was listed, and no open pull request adds one. The highest number anywhere was **0146**, so this is **0147**.
**Builds on:** [ADR 0045](0045-combat-restrictions.md) (the #1706 division prompt), [ADR 0110](0110-remember-me.md) §4 (synced settings), [ADR 0111](0111-action-dock.md) (the dock and its sheets), [ADR 0127](0127-answering-repeated-prompts-for-you.md) (a notice after an automatic answer), [ADR 0143](0143-gameplay-settings-overhaul.md) (the Gameplay tab's Prompts section).

---

## Context

**The rules.** CR 510.1c: a creature blocked by two or more creatures assigns its combat damage "divided as its controller chooses among" them. CR 510.1d does the same for a creature blocking two or more attackers. CR 702.19b: an attacker with trample may assign the rest to the player, planeswalker or battle it is attacking once every blocker has been assigned lethal damage. Lethal damage takes account of damage already marked, and is 1 from a deathtouch source (CR 702.19c, 702.2c), as `aiseat/heuristic/trample.go` and `ResolveDamageAssignment` already count it. Since #2692 the engine has no damage assignment order: any split that adds up is legal, and trample's lethal check is the one constraint.

**What exists.** `legal.canonicalDamageAssignment` (`server/internal/legal/choices.go`) builds the one split the enumerator offers, which is what the bots answer with. It kills the most the power can buy (`KillingSet`), and a trampler with enough power assigns lethal to every blocker and puts the rest over. `aiseat/heuristic/trample.go` estimates the same thing from the board for the attack decision. The client's sheet (`ChoicePromptModal.svelte`, the `damage_assignment` branch) started every blocker at 0 and the trample row at 0, with ▲/▼ order buttons left over from before #2692. The report on #2956: Vivi Ornitier (44, trample) blocked by creatures with lethal 4, 3 and 2. The player had to type 35 into the trample box.

## Decision

### 1. The server ships the canonical split on the prompt

`canonicalDamageAssignment` becomes `legal.CanonicalDamageSplit(g, choice)`, exported, and the enumerator's move is built from it, so there is one split. It returns the shares, the trample amount, each blocker's lethal damage and `CoversLethal`: the power is enough to assign lethal to every blocker.

`PendingChoiceView.damage_assignment` gains three additive fields, built from it in `viewOfPendingChoices`:

- `suggested`: `{assignments: [{blocker_id, amount}], trample_to_player}`, the resolve_choice payload's own shape;
- `lethal`: each blocker's lethal damage, indexed like `blocker_card_ids`;
- `covers_lethal`.

"Covers lethal" sums each blocker's lethal damage as the split counts it, so deathtouch counts 1 per blocker and damage already marked reduces it. A 5-power deathtouch trampler blocked by three 4/4s covers lethal (3, not 12): 1 to each blocker and 2 to the player. A 4-power trampler blocked by a 2/4 with 3 damage marked and a 1/1 covers lethal too (1 + 1): 1 to each and 2 over.

The client never works out lethal damage. Deathtouch, damage already marked and which blockers die stay on the server, next to the validator that checks them.

One behaviour change in the split: power that covers lethal for every blocker now assigns lethal to every blocker **with or without trample**. Before, a non-trampler did that only when `KillingSet` chose all of them, which it does not for a blocker worth nothing to kill (indestructible, or protected from the attacker); the leftover then went on that blocker. Now the leftover goes on the last blocker, as the issue asks. Either is legal and neither changes what dies.

### 2. The client sends it: auto-assign is a client setting

`gameplay.autoAssignCombatDamage`, default `true`, synced (ADR 0110 §4), in Settings → Gameplay → Prompts. When it is on and the viewer's next prompt is a damage assignment with `covers_lethal`, `DamageAutoAssign.svelte` sends `suggested` as the answer, once per prompt. `ChoicePromptModal` holds its sheet back meanwhile. The dock shows a notice for six seconds, "Vivi Ornitier: 4 to Solphim, 3 to Professional Face-Breaker, 2 to Corsair Captain, 35 to Bob (trample)", with **Undo** and **Always ask** (which turns the setting off). The resolve_choice is an ordinary commit, so the log shows the damage as it always has.

If the prompt is still pending four seconds after the answer was sent (the server refused it, or an undo brought it back), the sheet asks instead, and it is not sent again.

**Why the client and not the server.** ADR 0127 put standing answers on the server, because a closed tab would otherwise stall a run of prompts nobody is watching. That argument is weaker here. A damage assignment is asked once per blocked attacker, in a combat the player started a moment ago, so the tab is open. In return, a client toggle needs no new action, no `Player` field, no snapshot field and no reconcile loop, and the server's answer is the same one either way. If a player ever wants it while away, moving the send into the room is a small follow-up: the split is already the server's.

### 3. The prompt, when it asks

When the setting is off or the damage falls short of lethal for every blocker, the sheet opens **on the suggested split**: trample's leftover is on the player, not 0. The ▲/▼ order buttons are gone, because there has been no order since #2692. Each row has − and + beside its number, and the blocker's lethal damage.

While the sheet is up, the board draws the same numbers on the cards (`DamageStepper.svelte`, mounted by `Card`): each blocker carries its damage with − and +, marked once it is lethal, and the attacker carries what is left to assign and, with trample, the amount going over, with its own − and +. One store (`damageAssignment.ts` `boardDamageAssign`) holds the split, the pattern #2880 set for board picks (`boardChoicePick.ts`): the sheet publishes it, a press on the board updates it, and the sheet copies it back. The sheet's count and its **Deal damage** read it.

+ on a blocker takes from the unassigned damage first, then from the trample, so a full split can be moved one point at a time. − frees a point. While damage goes over and a blocker is short of lethal, **Deal damage** is disabled and the sheet says which blocker is short; the server would refuse that split anyway.

A blocker drawn as part of a stacked token group gets no stepper; the sheet still has its row.

### 4. What does not change

`ResolveDamageAssignment` checks every answer as before. The bots answer with the enumerator's move, which is the same split. The MCP seat sees the same legal move.

## Consequences

- The reported case is no clicks at all, and the canonical split is one click when the setting is off.
- The wire grows by three omitempty fields on one prompt kind. `docs/protocol.md` lists them.
- A future server-side auto-answer can reuse `CanonicalDamageSplit` unchanged.
