---
title: "Eminence triggers from the command zone"
date: 2026-10-09
issues: [2802]
---
**Eminence triggers from the command zone** (#2802, [ADR 0140](decisions/0140-eminence-statics-from-the-command-zone.md) amendment 2026-10-09; CR 113.6 / 603.4 / 108.4) — "Whenever …, if [this] is in the command zone or on the battlefield, …". The command zone is a declared trigger zone, walked per seat by the declared-zone harvest (`triggerZonesOfKindLocked`) only for the event kinds a registered card watches from there, with the zone's owner as "you". `effects.EminenceTrigger(...)` declares `{ZoneBattlefield, ZoneCommand}` (the one declaration in which the battlefield is named beside another zone, accepted by `game.TriggerZonesUnsupported`) and checks the intervening "if" again on resolution, so a commander cast in response does nothing. It composes with `OncePerBatch`, `Targeting` and the step and ETB constructors. **Cards**, all Full: Edgar Markov, The Ur-Dragon (its eminence is ADR 0140's cost form; its attack count is shared with The Ur-Sphinx), Sidar Jabari of Zhalfir, Arahbo, Roar of the World and Inalla, Archmage Ritualist. Eminence replacements are not built.
