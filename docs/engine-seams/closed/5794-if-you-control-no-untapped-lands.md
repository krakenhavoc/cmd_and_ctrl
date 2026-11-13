---
title: "If you control no untapped lands"
date: 2026-11-13
issues: [5794]
---
**If you control no untapped lands** (#5794, CR 603.4, 611.3a, 613.1f, 613.4c, 508.1f) — the Prophecy condition, as a static's "as long as" and as a trigger's intervening "if". One board read, `effects.NoUntappedLands(g, you)` (`cards/effects/no_untapped_lands.go`), over the shared "you control" walk and Lotus Vale's untapped-land test: no permanent you control is both a land and untapped, read through the layers. Four shapes on it: `SelfPumpWhileNoUntappedLands` (layer 7c), `SelfKeywordWhileNoUntappedLands` (layer 6), `WheneverThisAttacksIfNoUntappedLands` and `AtYourEndStepIfNoUntappedLands`, the last two `InterveningIf` over `CondNoUntappedLands`, which declares `game.FootprintCondBoard`. The statics stay current because the layer cache already drops on every tap and untap event; the one tap with no event is an attacker's own (CR 508.1f), so the statics declare `DependsOnEvents{EventAttack}` for an animated land that attacks. No new engine state, snapshot field, effect key or wire field.

**Cards:** Scoria Cat, Spur Grappler, Fen Stalker, Vintara Snapper, Keldon Berserker, Well of Life, Well of Discovery (Full).
