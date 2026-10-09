---
title: "Protection from a card type chosen as it enters"
date: 2026-10-08
issues: [2742]
---
**Protection from a card type chosen as it enters** (#2742, CR 702.16, CR 614.12, [ADR 0072](decisions/0072-protection.md)) — "choose a card type" is the Sieges' CR 614.12 option pick, offered over `game.ChoosableCardTypes` (CR 205.2a's nine traditional card types) and stored on `Card.ChosenOption`, which already has the right lifecycle: per instance, carried by the snapshot and clone, cleared when the permanent leaves, not copiable. No new field and no snapshot change. The protection is resolved from the permanent that holds the answer, not from each protected object. A layer-6 static appends `game.ProtectionFromCardType(answer)` ("protection from instants") to each creature its controller controls. The player half is the `Spec.PlayerKeywords` placeholder `game.ProtectionFromTheChosenCardType`, which `playerAbilityTokensLocked` replaces with the same token as it walks the granting permanent. Unanswered, neither half grants anything. The grammar learns "kindred". **Card:** Serra's Emissary (Full).
