---
title: "Uncopyable abilities and the as-this-transforms-into hook"
date: 2026-09-24
issues: [1574]
---
**Uncopyable abilities and the as-this-transforms-into hook** (#1574, CR 707.10 / CR 712.18, [ADR 0043 amendment 2026-09-24](decisions/0043-copy-effects.md) Decision 20, [ADR 0079 amendment 2026-09-24](decisions/0079-transforming-a-permanent.md) Decision 9) — "This ability can't be copied" is `ActivatedAbility.Uncopyable`, stamped onto the stack item at activation (`StackItem.Uncopyable`, carried by Clone and the snapshot) and refused by `CopyAbilityForEffect`, the one door every ability copy comes through, so Lithoform Engine, Strionic Resonator, Rings of Brighthearth and another Gogo copy nothing. Like "can't be countered", it narrows no target clause. `Spec.AsTransformsInto` is a face's "As this permanent transforms into <face>" clause; the in-place transform verb runs it for the face now up whatever effect asked for the transform, reading `CatalogAbilityKey` so an ability removal still applies (CR 712.18), and "exile it, then return it transformed" does not run it. **Gogo, Master of Mimicry** and **Sephiroth, One-Winged Angel** now ship `full`.
