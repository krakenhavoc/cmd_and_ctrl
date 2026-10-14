---
title: "Goad for the rest of the game, one opponent per permanent"
date: 2026-10-14
issues: [3522]
---
**Goad for the rest of the game, one opponent per permanent** (#3522, [ADR 0045](decisions/0045-combat-restrictions.md) amendment 2026-10-14) — `Goad.ForRestOfGame` is a goad with no end of its own: the turn-start sweep skips it, a later "until your next turn" goad by the same player never shortens it, and it ends only when the permanent leaves the battlefield (CR 400.7). It survives a control change, because the goad names the goader. `Game.GoadForRestOfGameForEffect` is the door; the flag is an additive snapshot key (`forRestOfGame`). Handing N permanents to N different opponents needed no new prompt: a chain of single-pick choose-player prompts through a registered option-pick continuation (so a table waiting on one is still a restore point), each leaving out the opponents already chosen, and an obvious last pick is made without asking. **Cards:** Dack Fayden, Helping Hand (Full).
