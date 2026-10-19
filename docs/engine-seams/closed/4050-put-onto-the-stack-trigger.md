---
title: "Triggers when a spell or ability is put onto the stack"
date: 2026-10-19
issues: [4050]
---
**Triggers when a spell or ability is put onto the stack** (#4050, CR 603.3, CR 115.9a, CR 707.10) — "whenever a spell or ability is put onto the stack, if it has a single target" has five ways in, each a watchable event: `EventCast`, `EventSpellCopied`, `EventActivateAbility`, `EventTriggerPlaced` (#4099, the CR 603.3b drain) and the new `EventAbilityCopied`.
- **`EventAbilityCopied`.** A copy of an activated or triggered ability is created, not activated or triggered (CR 707.10a), so it fired neither of the other kinds. `createAbilityCopyLocked` (`game/ability_copy.go`) now emits it, with the copy's `StackItemID`, only while a permanent on the battlefield watches it (`battlefieldWatchesLocked`, the generalisation of `triggerPlacedWatchedLocked`), so a table without a watcher keeps its event log. It is a separate kind from `EventTriggerPlaced` on purpose: a copy carries the original's triggering event, so a watcher that filters a placed trigger by its cause (Aboleth Spawn) would copy the copy forever.
- **The intervening if.** "If it has a single target" is `StackItemTargetCountForEffect == 1` (CR 115.9a counts instances of "target", not objects), checked when the event happens and again as the trigger resolves, where the item may have left the stack or changed shape. The reselection is `ReselectTargetsAtRandomForEffect` (#1815) with no exclusion.
- **No loop.** Grip's own trigger targets nothing, so it never satisfies the condition, for itself or for a second Grip.
- **Cards:** Grip of Chaos (Full).
