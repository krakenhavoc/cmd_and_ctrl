---
title: "Casting one spell from among several cards"
date: 2026-10-02
issues: [1729]
---
**Casting one spell from among several cards** (#1729, CR 611.2c) — `game.CastPermission.CastsLeft` is how many spells a permission may still open, zero being no limit. A cast the permission is the reason for spends one (`consumeLimitedGrantLocked`, `castUsesGrantLocked`), and the permission ends at zero; a cast the card's own text allows, such as a printed flashback, spends nothing, which is Locke's ruling. A card can now hold two stored permissions and the caster chooses: `CastPermissionForClaimLocked` picks the permission whose `AltCostKey` the cast claims, and `CastOffersForLocked` lists the second permission's offer, so the view and the bot enumerator offer it too. Snapshot: one additive `castPermissions[]` key, `castsLeft`. **Cards:** Court of Locthwain (Full) — the card it exiles is playable while exiled with mana of any type, and on a monarch turn one spell from among the cards exiled with it is free, claimed as an alternative cost so a paid cast leaves the free one for another card; Locke, Treasure Hunter (Caveats: a milled card a replacement exiled instead is not covered).
