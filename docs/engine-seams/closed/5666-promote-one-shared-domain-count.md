---
title: "Promote one shared domain count"
date: 2026-11-12
issues: [5666]
---
**Promote one shared domain count** (#5666, CR 207.2c, CR 305.6) - not an engine gap: the count existed twice as private helpers (`domainCountYouControl`, `rfCreatureBBasicLandTypesAmong`). `cards/effects/domain.go` now exports `DomainCount`, `DomainCostAmount`, `DomainCostReduction` (a spell or an activated ability's cost clause) and `DomainPT` (a layer 7a characteristic-defining power and toughness), and Leyline Binding and Fblthp, Knows the Way moved onto it. Built on it: Jodah's Codex, Llanowar Greenwidow, Radha's Firebrand, Yavimaya Sojourner, Sunbathing Rootwalla, Territorial Maro, Nishoba Brawler. No snapshot, wire or effect-key change.
