---
title: "Protection from each of your opponents"
date: 2026-10-08
issues: [2745]
---
**Protection from each of your opponents** (#2745, CR 702.16i over CR 702.16k, [ADR 0072](decisions/0072-protection.md)) — the closed protection grammar gains `ProtectionQualityOpponents`, printed "each of your opponents" and spelled with `game.ProtectionFromEachOfYourOpponents`. Like the chosen-player quality it tests the source's controller, against `ProtectionQuality.Holder`, which the reader fills in: a player's own seat through `bindPlayerProtectionQuality` (on `PlayerProtectedFromLocked` and the targeting walk), a permanent's controller through `bindProtectionQuality`. An unbound quality, or a source with no controller, matches nothing. So damage, targeting and enchanting all read it with no new consumer. Wire: `ProtectionView.kind` `"opponents"`, with `value` the seat that has it. **Card:** Absolute Virtue (Full), a `Spec.PlayerKeywords` grant like Aegis of the Gods.
