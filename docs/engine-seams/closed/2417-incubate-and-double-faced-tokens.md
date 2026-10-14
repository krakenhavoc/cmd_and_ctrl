---
title: "Incubate and double-faced tokens"
date: 2026-10-14
issues: [2417]
---
**Incubate and double-faced tokens** (#2417, CR 701.53, CR 111.10i, CR 707.8a, [ADR 0185](decisions/0185-double-faced-tokens-and-incubate.md); follows [ADR 0079](decisions/0079-transforming-a-permanent.md), which listed Incubator tokens as out of scope) - the Incubator is the first double-faced token: a colorless Incubator artifact with "{2}: Transform this token." whose back face is a 0/0 Phyrexian artifact creature named "Phyrexian Token".
- **Template.** `tokenTemplate.Back` (`effects/token_catalog.go`) is a token's second face. `tokenFromCatalog` builds `Layout`, `Faces` and `ActiveFace` from the pair, and the back face files any abilities under `<token key>#1`. `catalogKeyOf` keys a turned-over token that way, as a double-faced card's back face already keys as `<oracle id>#1`. `Transform` needed no token-specific code: `CanTransform` already keys on the layout and the face count.
- **Keyword action.** `effects.Incubate{Controller, N}` creates the token through the token-creation window with `+1/+1` counters as its entry clause, so Doubling Season and the other replacements see both the token and the counters. `Incubate 0` still makes the token. `PrintedPTKnown` on the template makes an Incubator that turns over with no counters a real 0/0 that CR 704.5f buries.
- **Copies.** `TokenCopyTemplate` now copies a token's key (it dropped it, so a copy of any token with a catalog ability lost the ability) and, for a double-faced token, both faces and which is up, without counters.
- **Cards.** Sunfall, Traumatic Revelation, Converter Beast, Eyes of Gitaxias, Merciless Repurposing and Excise the Imperfect, all complete.
- **Still open.** A token's graveyard face is still reset to its front by `MoveCard` (the dies triggers read last-known information, so nothing observable is wrong); Incubator art (the art pool takes only the `token` layout, #3511); copies of real double-faced permanents (#3512); Incubation Triformer's third side; "incubate N, X times"; and the Incubator-aware cards, each filed on #2417.
