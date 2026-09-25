---
title: "Landwalk"
date: 2026-09-17
issues: [705]
pr: 705
legacy_order: 103
---
**Landwalk** (#705) - `islandwalk`, `swampwalk`, `forestwalk`, `mountainwalk`, `plainswalk` and `nonbasic landwalk` are canonical keyword tokens, read in the evasion slot of `Game.BlockPairRefusalLocked` (`game/block_legality.go`, `game/landwalk.go`), the one block-pair check `DeclareBlocker`, the legal-move enumerator and the #328 auto-pass signal all call ([ADR 0045](decisions/0045-combat-restrictions.md) addendum, Decisions 7-10). The check is a method on the game, lock-held and read-only (`TestBlockLegalityDoesNotMutate`), and derives the defending player from the attack (a planeswalker's controller, a battle's protector). Both sides are read after layers: a granted landwalk through `HasKeyword`, a land's types through `IsLand` / `HasSubtype` / the new `Card.HasSupertype`, so Urborg switches swampwalk on. A refused block comes back as the `illegal_block` error with a `reason` token and a server-built sentence. Imported cards that print landwalk are enforced from the same change. Shipped on Lord of Atlantis (both caveats gone), Elvish Champion and Goblin King (their landwalk granted, now complete), Master of the Pearl Trident, Cold-Eyed Selkie and Trailblazer's Boots. The rarer variants (snow swampwalk, legendary landwalk, desertwalk) join the token table with their first card. Protection's blocking half (#662) took its reserved slot in the same function; block rules (#750) still have theirs.
