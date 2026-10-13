---
title: "Crimes at announcement"
date: 2026-10-13
issues: [2944]
---
**Crimes at announcement** (#2944, [ADR 0157](decisions/0157-crimes-at-announcement.md), CR 700.13) — a qualifying cast, activation or triggered-ability placement marks one existing `EventBecomesTarget` with `CrimeCommitted` regardless of its number of targets. Opponents' stack abilities are classified by their controller, independently of their source permanent. Creating a copy or changing an existing object's targets retains ordinary becomes-target events for ward and other target triggers, but commits no new crime. Magda and Lazav use the shared crime predicate. **Forsaken Miner** is added with a graveyard-only trigger, optional payment at resolution and a check that the same graveyard object is still present before returning it. The additive event annotation preserves existing snapshot histories; there is no snapshot schema change.
