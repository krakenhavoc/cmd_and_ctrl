---
title: "Harmonize"
date: 2026-10-16
issues: [3931]
---
**Harmonize** (#3931, CR 702.180, [ADR 0135](decisions/0135-alternative-costs-that-tap-discard-awaken-and-emerge.md) amendment 2026-10-16) — "Harmonize [cost]" casts the card from your graveyard for that cost while tapping up to one untapped creature you control, reduces the cost by the tapped creature's power, and exiles the spell any time it would leave the stack (CR 702.180a). `effects.Harmonize(cost)` is flashback's graveyard offer whose card component is `game.HarmonizeTapCost()`, a tap of up to one creature (`TapOthersCost.UpTo`), flagged `AlternativeCost.ReducedByTappedPower`.
- **The price.** Emerge's shape with power for mana value: the creature is chosen with the cost and tapped as it is paid (CR 702.180b), so CR 601.2f totals the cost while it is on the battlefield. `CostQuery.AltTappedPower` is its power, clamped at zero, filled by the one pricer from `alt_cost_ids`; the reduction comes off the generic part, then off an announced {X} (#2701), never below {0}. Nature's Rhythm's and Zenith Festival's {X} harmonize costs buy a larger X for the same mana.
- **The payment.** Naming none pays the full cost, so the offer is never withheld for want of a creature. The creature is tapped with the spell on the stack by the tap-others payer (no summoning-sickness check, CR 302.6) and kept away from the auto-tapper.
- **Views.** `tap_options` (`min` 0, `max` 1), `reduces_by_power` and per-candidate `tap_prices`; the client's tap picker shows each creature's power and price, accepts none, and is skipped when there is nothing to tap.
- **The bot.** The enumerator prices each creature on its own, offers the affordable ones that save mana ranked by the policy's tap price less the mana saved, then the empty payment when it is affordable.
- **Granted.** A cast permission keyed `harmonize` (`effects.GrantHarmonizeToCard`, Songcrafter Mage) synthesises the same tap and reduction.
- **Guards.** `effects.Register` holds the flag and `UpTo` together over one creature and refuses `UpTo` on every other tap cost.
- **Cards** (12, all Full): Channeled Dragonfire, Glacial Dragonhunt, Mammoth Bellow, Nature's Rhythm, Roamer's Routine, Songcrafter Mage, Synchronized Charge, Unending Whisper, Ureni's Rebuff, Wild Ride, Winternight Stories and Zenith Festival.
