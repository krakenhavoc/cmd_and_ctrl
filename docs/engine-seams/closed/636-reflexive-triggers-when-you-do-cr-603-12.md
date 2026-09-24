---
title: "Reflexive triggers, \"when you do\" (CR 603.12)"
date: 2026-09-17
issues: [636]
pr: 795
legacy_order: 106
---
**Reflexive triggers, "when you do" (CR 603.12)** (#636) - a trigger CREATED by a resolving effect, written `ReflexiveTrigger{Label, Targets, Optional, Cards, Effect}.Apply(ctx)` (or `WhenYouDo(label, effect)` for the plain case) from inside the parent ability's `Effect`, on the branch where the condition actually held. It goes through `Game.QueueReflexiveTriggerForEffect` into the harvester's own dispatch, so it gets the CR 603.3d target pick when it is put on the stack, the CR 603.3d drop when nothing is legal, a CR 603.5 "you may" if it prints one, and a response window above its parent. The payload rides `StackItem.Payload` (`ctx.PayloadCards()`), which is what keeps the `Effect` capture-free across `Clone` / undo. Cards that folded the follow-up into the parent and declared a caveat are converted: Ziatora, the Incinerator; Invasion of Tarkir; Generous Plunderer; Riveteers Overlook and the four b08 Overlook lands (Brokers Hideout, Cabaretti Courtyard, Maestros Theater, Obscura Storefront). Eden, Seat of the Sanctum was the one card here waiting on a SECOND seam — a free yes/no a resolving ability can ask its controller — and #796 closed it (see below); Eden is now `CompletenessFull`. Still waiting: Caesar, Legion's Emperor (modal triggers); Invasion of New Phyrexia (#626), whose back face needed emblems too until #623 landed them (see Closed seams). See [ADR 0026](decisions/0026-delayed-triggers.md) (2026-09-17 amendment) and AGENTS.md §7.
