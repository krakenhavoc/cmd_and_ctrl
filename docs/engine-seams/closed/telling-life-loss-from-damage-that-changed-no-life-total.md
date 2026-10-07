---
title: "Telling life loss from damage that changed no life total"
date: 2026-10-07
issues: [2105]
---
**Telling life loss from damage that changed no life total** (#2105) — damage dealt to a player is life lost only when it costs life (CR 119.2, CR 120.3a). The player-damage path stamps the part that cost none on its `EventDealDamage` as `DamageNotLifeLoss`: all of it for a source with infect, or dealt as though it had infect (CR 702.90b), and all of it for a player whose life total can't change (CR 119.8). `Event.DamageLifeLoss()` is the life the damage cost, and `Amount` is still the damage, so "is dealt damage" readers are unchanged. `s22PlayerLostLife`, `b04OpponentLostLife`, The Master of Lake-Town's reader, the Ob Nixilis pre-filter and the turn tally's `LifeLost` read it, so Vilis, Exquisite Blood, Transcendence, Mindcrank, Valgavoth, Wound Reflection, Bloodchief Ascension, spectacle and every other "lost life this turn" card stop counting infect damage and damage to a locked life total. The field is additive in snapshot v7, and records the part NOT lost so that an event written before it reads as it did then. **Cards:** Lich's Mastery (Full).
