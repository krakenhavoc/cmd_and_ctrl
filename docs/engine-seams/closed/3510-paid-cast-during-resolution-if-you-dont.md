---
title: "A paid cast from exile during a resolution, with an \"if you don't\""
date: 2026-10-14
issues: [3510]
---
**A paid cast from exile during a resolution, with an "if you don't"** (#3510, CR 608.2g, CR 601.2g; [ADR 0099](decisions/0099-discover.md) amendment 2026-10-14) — Chandra, Torch of Defiance: "Exile the top card of your library. You may cast that card. If you don't, Chandra deals 2 damage to each opponent."
- **The window.** `Game.ExileTopThenPaidCastOrElseForEffect` exiles the top card and asks a `may_cast` under `paid_or_else`. Accepting grants ADR 0099 §4's pass-closed window over the card at its printed cost: `TimingFlash`, `CastOnly`, `LapseOnPass: LapseStaysInExile`.
- **"If you don't".** `CastPermission.IfNotCast`, a registered body beside `FollowUp`, runs when a pass-closed permission lapses unused (the holder's pass or the end-of-turn sweep), and on the decline. A land, an empty library or a commander sent home asks nothing and runs it at once. A cast spends the grant and runs nothing.
- **Client and bot.** The client words the prompt as a paid cast with the server's decline label and opens the cast chain on accept. The enumerator offers both answers (the decline always legal) and then the paid cast; the heuristic accepts, which is never worse than declining.
- **Cards:** Chandra, Torch of Defiance (Full); Cecily, Haunted Mage (Full, on the #3270 pick, #2110).
