---
title: "Kicker {X}, with a floor of 1"
date: 2026-10-13
issues: [3232]
---
**Kicker {X}, with a floor of 1** (#3232, CR 107.3a / 601.2b / 702.33, [ADR 0073 amendment 2026-10-13](decisions/0073-optional-additional-costs-and-the-cast-gate.md)): `AdditionalCost.MinX` ("X can't be 0" on an optional cost, built with `effects.KickerX("{X}", 1)`) is refused at the announce gate under the floor, an {X} kicker that is not paid announces no X at all (a non-zero unkicked X is refused), the bot enumerator raises its X floor for the kicked expansion, the view ships `optional_costs[].x_min`, and the client's X picker opens only once the kicker is chosen, at the floor. **Cards:** Thieving Skydiver, Verdeloth the Ancient, Kangee, Aerie Keeper (Full). Emblazoned Golem stays out: its "no more than one mana of each color may be spent on X" cap has no shape.
