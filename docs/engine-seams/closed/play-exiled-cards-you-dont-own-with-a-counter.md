---
title: "Playing exiled cards you don't own that carry a counter"
date: 2026-10-06
issues: [2179]
---
**Playing exiled cards you don't own that carry a counter** (#2179): "during your turn, you may play cards you don't own with stash counters on them from exile, and mana of any type can be spent to cast those spells" is a derived standing cast permission (`Spec.CastPermissions`), read off the battlefield on every query and never stored. `PermissionFilter` gained `WithCounter` and `NotOwnedByHolder` (the owner clause lives in `permissionReachesPileLocked`, which knows the holder), and a `CastPermission` may now carry `TimingYourTurnOnly`, which `CastTimingOpenLocked` reads and which leaves the card's own timing in force. The marker is an ordinary counter that `MoveCard` clears when the card leaves exile, and the permission ends when its source leaves. Exile is a shared pile, so `CastPermissionOnCardForEffect` now also asks the derived permissions and the view stamps the holder's cast surface; `CastSpell`, the bot enumerator and the view agree. **1 card:** Tinybones, Bauble Burglar (new, complete).
