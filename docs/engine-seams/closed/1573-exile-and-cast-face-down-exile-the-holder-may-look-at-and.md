---
title: "Exile-and-cast: a face-down exile the permission holder may look at, and mana of any type"
date: 2026-09-24
issues: [1573]
pr: 1586
---
**Exile-and-cast: a face-down exile the permission holder may look at, and mana of any type** (#1573, CR 406.3 / 106.1b, [ADR 0066's 2026-09-24 amendment](decisions/0066-granted-cast-and-play-permissions.md)). There are two gaps in the "exile it and you may cast it" family.

**1. A face-down exile the holder may look at.** `FaceDownPermitted` (`"permitted"`) is an exile-side face-down kind. Its viewer is the holder of the cast permission over the card. `GrantCastPermissionForEffect` stamps the look on such a card and on no other face-down kind, so Necropotence's `exiled` card stays unreadable. `Game.ExileTopFaceDownWithPermissionForEffect` picks the top N cards up front and routes each one through the shared exit primitive. It grants from the continuation, so a commander whose owner declines the command zone is still exiled and granted. The existing non-knower redaction handles the rest: the owner and the other seats get a card back with no `exile_play` and no cast surface.

**2. Mana of any type.** `CastPermission.AnyType` means "mana of any type can be spent". Colorless is a type and not a color, so it folds `{C}` requirements into generic as well as the colored ones. `spendAsThoughAny` applies the fold in the one pricer, so the hand payment, the auto-tapper, the preview, the bot enumerator and `cast_prices` all agree. On the wire it appears as `exile_play.any_type`, which is set together with `any_color`.

**Cards now `full`:** Hostage Taker, Gonti, Night Minister and Outrageous Robbery. Tracker [#885](https://github.com/krakenhavoc/cmd_and_ctrl/issues/885); deck tracker [#1565](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1565).
