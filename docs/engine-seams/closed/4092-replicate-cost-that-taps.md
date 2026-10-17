---
title: "A replicate cost that taps a permanent"
date: 2026-10-17
issues: [4092]
---
**A replicate cost that taps a permanent** (#4092, [ADR 0129](decisions/0129-energy-getting-and-paying-it.md) amendment 2026-10-17, CR 702.56a / 601.2h / 302.6) — replicate's only non-mana component was energy, and ADR 0073 §4 let only a mana-only optional cost repeat, so Psionic Ritual's "Replicate—Tap an untapped Horror you control" had no shape. `effects.ReplicateTapping(label, max, preds…)` is the optional cost keyed `replicate` with #3868's `TapOthersCost`, one permanent per payment; `effects.Register` allows a tap component on that optional cost alone, never beside a mandatory tap cost, teamwork or escalate's taps.
- **The cast path.** N payments are N plan entries, and `game.planTapOthers` sums them into one demand, so the existing validator and payer check and tap the whole `teamwork_ids` list at once (each once, the caster's, untapped, matching without targeting) with the spell on the stack, recorded on `PaidCost.TappedOthers`. A Horror that arrived this turn pays (CR 302.6).
- **The wire and the bot.** `optional_costs[i].tap_others_label` / `tap_others_options` (one payment's count as `min` / `max`, `max_times` capped at what the options cover); the client's tap picker asks for that count times the payments claimed. The enumerator names that many off the engine's walk and drops a set the seat cannot pay. No snapshot change.

**Cards:** Psionic Ritual, Full.
