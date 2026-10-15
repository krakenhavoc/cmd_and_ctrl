---
title: "A value recorded as you cast a spell"
date: 2026-10-15
issues: [3781]
---
**A value recorded as you cast a spell** (#3781, Lifestream's Blessing of the Cloud, Ex-SOLDIER deck #3656; CR 601.2i, CR 707.10, [ADR 0021](decisions/0021-additional-costs.md)'s 2026-10-15 amendment).
- **Declaration.** `Spec.CastTimeValues` is a keyed list of `game.CastTimeValue{Key, Value}` (`game/cast_time_values.go`). `Register` refuses a blank key, a nil `Value` or a duplicate key.
- **The stamp.** `CastSpell` runs `stampCastTimeValuesLocked` after every cost is paid and before `EventCast`, so nobody can respond before the value is fixed. A creature that dies, arrives or grows in response does not change it.
- **The memory.** `StackItem.CastValues`, carried by `Clone`, the snapshot (`castValues`, additive in v9) and a spell copy (CR 707.10). Read at resolution with `ctx.CastValue(key)`.
- **Cards:** Lifestream's Blessing (Full).
