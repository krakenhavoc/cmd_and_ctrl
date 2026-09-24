---
title: "Life-change effect-path replacement routing"
date: 2026-09-17
issues: [482]
pr: 790
legacy_order: 129
---
**Life-change effect-path replacement routing** (#482) - every writer of a life total now runs the CR 614 window and lands in one tail (`server/internal/game/life_tail.go`): the public `ChangePlayerLife` sandbox verb, `ChangePlayerLifeForEffect` (every catalog `GainLife`, drain and pay-life cost) and the CR 702.15b lifelink credit. Damage deliberately does NOT fire it — damage reduces life directly once the DAMAGE replacements have settled (CR 120.3). Rhox Faithmender now doubles a `GainLife` and its own lifelink; Bloodletter of Aclazotz, Alhammarret's Archive and Angel of Vitality can be written.
