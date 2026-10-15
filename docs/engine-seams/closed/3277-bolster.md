---
title: "Bolster"
date: 2026-10-14
issues: [3277]
---
**Bolster** (#3277, [ADR 0152](decisions/0152-more-than-meets-the-eye-convert-and-living-metal.md), CR 701.39) — `effects.Bolster{N, Then}` chooses a creature the resolving player controls with the least toughness (read after the layers; the controller picks only on a tie, through the ordinary choose-cards prompt), puts N +1/+1 counters on it, and hands `Then` the chosen creature so the rest of the resolution can refer to it. With no creature nothing happens and `Then` does not run. Optimus Prime, Hero // Optimus Prime, Autobot Leader is complete: the back face grants the chosen creature trample and binds a one-shot delayed trigger (CR 603.7) to that object, whose convert is ignored if Prime turned over meanwhile (CR 701.28e).
