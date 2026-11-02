---
title: "Whenever a player loses life, that much, over simultaneous losses"
date: 2026-11-02
issues: [4460]
---
**Whenever a player loses life, that much, over simultaneous losses** (#4460) — the engine opens one event per damage source and per life change, but a printed "whenever a player loses life, … that much" is one trigger reading the TOTAL that player lost at the same time (CR 603.2c, CR 120.3a). `BatchTotalLifeLost`, the third reading of `AtBatchEnd` beside `BatchTotalDamage` and `BatchTotalEvents`, groups by the losing player and settles with the batch's running per-player life-lost total (`batchLifeLost`, which counts what was lost after every replacement), never less than the staged event's own loss. A card sets `AtBatchEnd: ThatMuchLifeLost()` and reads `item.Trigger.Event.Amount`, positive whichever event kind opened the batch. Two creatures connecting at once are one loss of their sum; two players losing together are two triggers. **Cards:** Hela, Mistress of Darkness // Dark Tutor (Full).
