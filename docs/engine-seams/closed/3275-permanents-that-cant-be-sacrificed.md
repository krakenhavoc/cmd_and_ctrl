---
title: "Permanents that can't be sacrificed"
date: 2026-10-14
issues: [3275]
---
**Permanents that can't be sacrificed** (#3275, CR 101.2 / CR 118.3 / CR 701.21a) — one predicate, `permanentCantBeSacrificedLocked` (`game/cant_be_sacrificed.go`), reads two forms: the static `Characteristic.CantBeSacrificed`, set by the layer-6 `effects.CantBeSacrificed()` and gone with the permanent's abilities (CR 613.1f), and `ModCantBeSacrificed`, a scoped rules gate pinned to one object until cleanup (`effects.CantBeSacrificedThisTurn`; CR 400.7, 514.2). A sacrifice cost can't be paid with such a permanent: the shared cost validator refuses it (`ErrCantBeSacrificed`), and the view's options, the bot enumerator, the auto-tapper, alternative and "sacrifice all" costs and ward payments leave it out. An effect can't choose it: every edict, ADR 0175's ordered collector and `effects.SacrificeChoice` skip it. And the sacrifice route leaves it where it is, announcing nothing, so a delayed "sacrifice it", "sacrifice all", evoke and the Saga rule do nothing to it. **Cards:** Slicer, Hired Muscle // Slicer, High-Speed Antagonist (Full) and Alexios, Deimos of Kosmos (Full).
