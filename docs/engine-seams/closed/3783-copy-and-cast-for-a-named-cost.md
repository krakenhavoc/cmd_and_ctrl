---
title: "Cast a card copy for a named alternative cost"
date: 2026-10-15
issues: [3783]
---
**Cast a card copy for a named alternative cost** (#3783, [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md) amendment 2026-10-15) — "you may cast the copy by paying {3} rather than paying its mana cost". `CardCopyOffer.Cost` already carried any price; `effects.CopyAndCast` could only name "{0}" (`Free`) or the printed cost, and now takes `Cost: "{3}"`. **Cards:** Blue Mage's Cane (Full): a granted attack trigger on the Job select shape, targeting up to one instant or sorcery card in the defending player's graveyard, exiling it, and offering a copy of what reached exile for {3}.
