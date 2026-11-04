---
title: "Blight N as part of an activated ability's cost"
date: 2026-11-03
issues: [4684]
---
**Blight N as part of an activated ability's cost** (#4684, CR 701.68a, 701.68b, 602.2b, 614.16, [ADR 0020](decisions/0020-activated-abilities.md) amendment 2026-11-03) — `game.AbilityCost.Blight` is "Blight N" in a cost, built by `effects.BlightN(n)`: the activator names, in `blight_ids`, any creature they control that can have counters put on it (`BlightOptionsForEffect`, the cast path's walk), the source included. The counters go on through `blightLocked`, marked `CounterFromCost`, so Doubling Season doesn't double them and Vizier of Remedies applies (CR 614.16). With no creature to blight the ability can't be activated (CR 701.68b), and a creature that dies of the counters still paid. A graveyard ability blights a creature on the battlefield. The row ships `blight` / `blight_options`; the client asks which creature after the unattach pick; the enumerator offers one move per creature, a survivor first, and `legal.MoveCost.Blight` prices the counters for the bot. **Cards:** Gristle Glutton, Sting-Slinger, Evershrike's Gift and Spiral into Solitude — Full.
