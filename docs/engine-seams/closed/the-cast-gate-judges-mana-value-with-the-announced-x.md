---
title: "The cast gate judges mana value with the announced X"
date: 2026-11-24
issues: [6085]
---
**The cast gate judges mana value with the announced X** (#6085) — a "can't cast spells with mana value greater than N" restriction judged an {X} spell with X as 0, so a large X slipped through. `CastQuery.XValue` now carries the announcement, and a `CastRestriction` may declare a `ManaValueCeiling`: the gate refuses a cast whose mana value, counting the announced X (CR 202.3e, 601.3a), is above it, and `CastGateXCeilingForEffect` derives the largest X every reader uses (the bot's move list, the client's `x_max`, the auto-tap preview). Catalog data only, no snapshot change. **Cards:** Lavinia, Azorius Renegade (Full).
