---
title: "Untap restrictions, next-untap markers and stun counters"
date: 2026-09-17
issues: [751]
pr: 852
legacy_order: 98
---
**Untap restrictions, next-untap markers and stun counters** (#751, [ADR 0058](decisions/0058-doesnt-untap.md)): `Spec.UntapStepRestrictions` checks self, attached and filtered restrictions at the controller's untap step after layers. `DoesntUntapNextUntapStep` and `TapAndFreeze` record controller- or player-keyed markers that survive skipped steps and snapshots. Stun counters replace any attempted untap, including effect and sandbox untaps. The first wave adds the Monoliths, Mana Vault, Meekstone, Back to Basics, Intruder Alarm, freeze spells and stun cards, and completes Tangle. The board shows a tapped-only badge and hover explanation; bots and auto-tap account for the next missed untap. Exert remains a separate seam; choose-N untaps closed with #826 (below), and duration-bound locks with #1313 (below).
