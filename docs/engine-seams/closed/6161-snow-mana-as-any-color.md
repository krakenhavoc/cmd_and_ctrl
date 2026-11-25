---
title: "Spend mana from snow sources as though it were any colour"
date: 2026-11-25
issues: [6161]
---
**Spend mana from snow sources as though it were any colour** (#6161, CR 609.4b, CR 107.4h) — `CastPermission.AnyColorFromSnow` narrows the "spend mana as though it were mana of any color" clause by where the mana came from. With it, each coloured symbol of the cast's cost is marked `ColorRequirement.SnowAny` (`widenSnowAny`) and also takes any mana a snow source produced. Other mana pays as printed, and a {C} or an {S} is not widened. Because the answer depends on the mana's source, the slot is paid in the coloured passes through `ColorRequirement.PaidBy(color, snow)`, the same question {S} asks. So the pool solver, the auto-tapper, the auto-tap top-up and the move enumerator all read it the same way. **Cards:** Draugr Necromancer (Full).
