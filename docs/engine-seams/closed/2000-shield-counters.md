---
title: "Shield counters"
date: 2026-10-14
issues: [2000]
---
**Shield counters** (#2000, CR 122.1c, [ADR 0173](decisions/0173-shield-counters.md)) — two engine-owned built-ins in `game/shield_counters.go`, one registration each, so any number of counters is a single replacement and a single prevention effect. A destruction by an effect is cancelled and a shield counter removed. The state-based-action sweep destroys through a new `sbaDestroyRoute` (`ReplacementEvent.StateBasedDestruction`), so lethal damage and deathtouch aren't replaced. Damage is prevented (CR 615), and removing the counter is the prevention's additional effect (CR 615.5): an ADR 0108 §8 `Then` owed once per recipient per damage instance, so damage from several sources at the same time costs one counter. Damage that can't be prevented is dealt and a counter is still removed (CR 615.12). A built-in has no object of its own, so it owes the follow-up on the damaged permanent (`ReplacementEffect.thenOnDamaged`). The permanent's controller orders them against any other applicable replacement (CR 616.1). **Cards:** Shield Broker, Protection Magic and Summon: Magus Sisters (Full).
