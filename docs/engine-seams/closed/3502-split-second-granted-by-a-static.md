---
title: "Split second granted to spells by a static"
date: 2026-10-14
issues: [3502]
---
**Split second granted to spells by a static** (#3502) — `Spec.SplitSecondGrants` declares a permanent's "<spells> you control have split second" (`game.SplitSecondGrant`, a `PermissionFilter` and the oracle sentence). It is read LIVE off the battlefield, never stamped at cast (CR 611.3a, CR 702.61a): `recomputeSplitSecondLocked` counts a spell on the stack as having split second when its own keyword or sandbox flag says so (`StackItem.SplitSecond`, unchanged) or when a source its controller controls grants it right now (`spellGrantedSplitSecondLocked`, keyed by `CatalogAbilityKey`, so a source that lost its abilities grants nothing). A battlefield zone move or a control change under a spell already on the stack re-runs it, so a granting permanent that leaves takes split second off the spell and one that arrives gives it. The stack chip reads `SpellHasSplitSecondForEffect`, so the wire's `split_second` and `split_second_active` agree. Amends ADR 0007's 2026-09-24 note. **Cards:** Samut, Tyrant of Naktamun (Full).
