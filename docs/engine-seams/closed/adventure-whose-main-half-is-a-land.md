---
title: "An Adventure whose main half is a land"
date: 2026-10-05
issues: [2176]
---
**An Adventure whose main half is a land** (#2176, CR 715.3d, CR 305.2) — `grantAdventureCastFromExileLocked` now writes the exile grant as a play permission, not a cast-only one, when the card's main half is a land (`adventureMainHalfIsLand`), because the printed text says "play". The land play from exile reuses the existing land branch of `CastSpell`, so it spends the turn's land drop, needs the controller's own sorcery-speed window, and is offered by the view (`castable_here`) and the bot enumerator only while both hold. A countered Adventure still goes to the graveyard with no grant. **Cards:** Jidoor, Aristocratic Capital // Overture, Zanarkand, Ancient Metropolis // Lasting Fayth, Value Town // Take a Trip to... and Ishgard, the Holy See // Faith & Grief (Full). **Not shipped:** Lindblum, Industrial Regency (a token with its own trigger) and Midgar, City of Mako (an optional sacrifice with an "if you do" at resolution).
