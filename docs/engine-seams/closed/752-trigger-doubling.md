---
title: "Trigger doubling"
date: 2026-09-17
issues: [752]
pr: 846
legacy_order: 93
---
**Trigger doubling** (#752, [ADR 0018 addendum](decisions/0018-triggers-on-the-stack.md#addendum-2026-09-17-trigger-doubling-cr-6032d--accepted)): `Spec.TriggerDoublers` counts extra instances at the event harvest, using current or last-known characteristics and simultaneous-exit snapshots. Each instance has its own optional/target prompt and is attributed on the stack and prompt. Counts add (two doublers give three triggers); delayed/reflexive/manual triggers are excluded, while evoke and Saga abilities participate where their causes match. Panharmonicon, Teysa Karlov and Isshin are the reference cards; Cloud, Midgar Mercenary gains its missing ability. The accepted `OncePerBatch` limitation remains: a filter that matches only a later member of an already-open batch can undercount. Other first-wave cards remain follow-ups under #752; trigger suppression and copying an ability remain separate seams.
