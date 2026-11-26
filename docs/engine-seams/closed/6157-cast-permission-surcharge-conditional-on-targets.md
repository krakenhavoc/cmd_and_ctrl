---
title: "A cast-permission surcharge conditional on the cast spell's targets"
date: 2026-11-25
issues: [6157]
---
**A cast-permission surcharge conditional on the cast spell's targets** (#6157, CR 601.2c, CR 601.2f) — `CastPermission.CostMoreUnlessTargetsYourCreature` (an additive snapshot field) is "If that spell doesn't target a creature you control, it costs {N} more to cast this way": `CostMore` read with the announced targets, since CR 601.2c announces them before CR 601.2f totals the cost. `Game.TargetSurchargeFor` is the one reader, adding N unless one of the targets is a creature on the battlefield the caster controls. It joins the permission's `CostMore` in `grantIncreaseLocked`, so CastSpell, the auto-tap preview and the view price it alike, and the enumerator prices such a permission per target set (as it does a target-reading cost modifier), so a move offered at a target is one the caster can afford there. **Cards:** Mavinda, Students' Advocate (Full).
