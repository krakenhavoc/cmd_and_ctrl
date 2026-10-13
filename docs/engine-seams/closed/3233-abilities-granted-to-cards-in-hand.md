---
title: "Abilities granted to cards in a hand, graveyard or exile"
date: 2026-10-13
issues: [3233]
---
**Abilities granted to cards in a hand, graveyard or exile** (#3233, CR 113.10 and 613.1f, [ADR 0093 amendment 2026-10-13](decisions/0093-abilities-granted-to-other-permanents.md)) — "Each creature card in your hand has ninjutsu {2}{U}{B}." (Satoru Umezawa).
- **A zone grant.** `StaticAbility.GrantZone` beside `GrantAbilities` names the zone whose cards receive the bundles; `AbilityGrant.Zone` says where a bundle's recipients are. A layer-pass step (`zoneGrantPassLocked`, after layer 2) stamps the same plain `GrantedAbility` data a permanent's grant is onto `Card.zoneGranted`. It is rebuilt, never snapshotted, and ends the pass after its source leaves or loses its abilities, or when the card changes zone.
- **One reader.** `activatedAbilityRows` reads it, so the activation path, the wire's `zone_abilities` row (with `granted_by`), the legal enumerator and the stack item's ref all see a granted hand ability. No closure, no new wire shape.
- **Cards (5):** Satoru Umezawa, Rhet-Tomb Mystic, Tectonic Reformation, Sedris, the Traitor King and Solemn Doomguide (all Full).
- **Not built:** triggered, static or mana abilities on cards off the battlefield, the command zone, and bundle abilities with a cost read off the recipient (Graywater's Fixer's encore {X}).
