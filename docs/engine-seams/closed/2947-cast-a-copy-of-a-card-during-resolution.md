---
title: "Casting a copy of a card during resolution"
date: 2026-10-14
issues: [2947, 3258]
---
**Casting a copy of a card during resolution** (#2947, #3258, CR 707.12, [ADR 0066](decisions/0066-granted-cast-and-play-permissions.md)'s 2026-10-13 card-copy amendment) — `effects.CopyAndCast{Cards, Free, FollowUp}` over `game.CopyCardsAndOfferCastsForEffect`. Each card that really reached exile (or sits in a graveyard) gets a copy marked `Card.CardCopy` and its own cast permission, at its printed cost or `{0}`, at instant timing and lapsing on the holder's next priority pass. A copy not cast ceases to exist (CR 704.5e) and a permanent copy resolves as a token (CR 707.10b). "If you do" is a registered `game.RegisterCastFollowUp` body that runs once, only for a cast made through that offer. **Cards:** Mizzix's Mastery and Arcane Bombardment (#3258), and Kaervek, the Punisher, whose crime trigger exiles up to one black card from your graveyard and costs 2 life only when the copy is cast (Full, #2910).
