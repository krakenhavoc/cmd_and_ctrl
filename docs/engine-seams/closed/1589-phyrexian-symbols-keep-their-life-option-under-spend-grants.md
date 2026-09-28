---
title: "Phyrexian symbols keep their life option under spend grants"
date: 2026-09-28
issues: [1589]
---
**Phyrexian symbols keep their life option under spend grants** (#1589, CR 107.4c / 107.4f, [ADR 0066's 2026-09-28 amendment](decisions/0066-granted-cast-and-play-permissions.md)). Under "spend mana as though it were mana of any color" (Breeches) or "mana of any type" (Hostage Taker, Gonti), `asAnyColorCost` and `asAnyTypeCost` used to fold a Phyrexian symbol into generic mana, so its "or 2 life" half was lost. The view still stamped `phyrexian_symbols` from the printed cost, so the client offered a life payment the engine refused. Both folds now keep the slot Phyrexian and mark it `ColorRequirement.AnyMana`: the life half stays, and any mana pays the mana half. Hybrid Phyrexian symbols work the same way. Every payment path asks one predicate, `ColorRequirement.Admits`, so the hand payment, the auto-tapper, the preview, the bot enumerator and `cast_prices` agree. A widened slot is paid last, so it can't take the token a narrower slot needed. `{C}` under the any-colour grant still needs colorless mana. Tracker [#887](https://github.com/krakenhavoc/cmd_and_ctrl/issues/887).
