---
title: "Casting a creature from your graveyard by foraging"
date: 2026-11-23
issues: [5184]
---
**Casting a creature from your graveyard by foraging** (#5184, CR 701.61a, 601.2b, 601.2h) — the last open part of the Forage row. Osteomancer Adept's ability grants two cast permissions until end of turn over creature cards in its controller's graveyard (`game.CastPermission`, ADR 0066), one per way to forage: the printed mana cost plus three other graveyard cards exiled (`ExileOtherFromGraveyard`), or the printed cost plus a Food sacrificed (`SacrificeFood`). Both are marked `CastPermission.Forage`, so paying either emits `game.EventForage` and "whenever you forage" sees it. **Card now Full:** Osteomancer Adept. Recorded by the November 2026 docs sync, which found the card registered Full and the row still partial.
