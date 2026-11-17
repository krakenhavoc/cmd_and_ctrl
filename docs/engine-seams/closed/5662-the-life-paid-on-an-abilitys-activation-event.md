---
title: "The life paid, on an ability's activation event"
date: 2026-11-12
issues: [5662]
---
**The life paid, on an ability's activation event** (#5662, CR 603.4 / 107.4f / 119.4 / 605.3b / 707.10; follows [ability-copy](engine-seams/closed/1223-ability-copy-and-a-trigger-that-reads-the-event-that-fired.md)) — the announcement already totalled the life it paid on `PaidCost.LifePaid` (a printed "Pay N life", an X or counted life cost, and 2 for each Phyrexian symbol paid with life), but only the ability's stack item held it, so a trigger on `EventActivateAbility` could not ask "if life was paid to activate it" or "that much life". `Event.LifePaid` now carries that total on the activation event, stamped after every cost is paid. It is on the event, beside `Exhaust` and `XCost`, and not read back from the stack item, because the intervening "if" is rechecked as the trigger resolves (CR 603.4) and the ability may have left the stack by then.
- **The card side.** `YouActivatedAnAbilityPayingLife` (the When), `LifeWasPaidToActivateIt` (the recheck) and `MayPayThatLifeAgainToCopyIt` (the effect) are in `ability_copy_life_paid.go`. A player who can't pay that much life is not asked (CR 119.4, CR 119.8); the keyed answer re-checks the payment and copies the ability `Event.StackItemID` names with `CopyAbility`, new targets offered (CR 707.10c). The second payment is not an activation and the copy is not activated (CR 707.10), so neither triggers the watcher again. A mana ability emits `EventManaAbilityActivated` and never reaches it (CR 605.3b).
- **Snapshot.** `life_paid` is additive in v9 wherever a triggering event is stored. One new effect key, `confirm/pay-life-copy-that-ability`. No wire, client or closure field changes.

**Cards:** Verrak, Warped Sengir — Full.
