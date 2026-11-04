---
title: "Protection that doesn't remove this Aura"
date: 2026-11-04
issues: [4843]
---
**Protection that doesn't remove this Aura** (#4843, CR 702.16n, CR 704.5m, [ADR 0072 amendment 2026-11-04](decisions/0072-protection.md)) — protection was a bare, deduplicated token, so the CR 704.5m check could not tell a white Aura's own protection from white from another source's and put the Aura in the graveyard. The exemption now rides on the token: `game.ProtectionSparingAura(tok, auraID)` mints "protection from white (doesn't remove <id>)" and `ProtectionSparingAuras` the "doesn't remove Auras" form. `ParseProtectionQuality` reads it into `ProtectionQuality.SparesAura` / `SparesAuras` and keeps `Printed` as the quality, and only `ProtectedFromAttachment`, the attachment check, reads it. Two instances are two tokens, so a second source of the same protection still removes the Aura. A card uses `effects.ProtectionFromChosenColorSparingThisAura()`. **Cards:** Flickering Ward (Full); Ward of Lights moved onto it (it fell off its creature when it chose white).
