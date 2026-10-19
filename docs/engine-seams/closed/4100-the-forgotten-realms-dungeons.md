---
title: "The Forgotten Realms dungeons and venture into the dungeon"
date: 2026-10-19
issues: [4100]
---
**The Forgotten Realms dungeons and venture into the dungeon** (#4100, CR 309, CR 701.49, [ADR 0203 amendment 2026-10-19](../../decisions/0203-the-initiative-and-dungeons.md)) — Lost Mine of Phandelver, Dungeon of the Mad Mage and Tomb of Annihilation are `DungeonSpec` rows with their room abilities (`cards/effects/dungeons_forgotten_realms.go`), so a plain venture offers all three; `effects.VentureIntoTheDungeon` is the card verb. The bot prices every room and the dungeon choice.
- **Rules pieces.** Runestone Caverns' play permission lasts while the cards stay exiled (CR 611.2a); Veils of Fear and Sandfall Cell ask each player in turn order through keyed option picks (restore points); a `cantBeBlockedBySubtypes` block-rule mod (Kick in the Door); `Spec.VentureLimits` for "each opponent can't venture into the dungeon more than once each turn" (Keen-Eared Sentry, `game/venture_limit.go`).
- **Cards (41, all Full).** Every Commander-legal "venture into the dungeon" and "completed a dungeon" card but one. **Still open:** Dungeon Delver, which needs a granted trigger doubler (#4137).
