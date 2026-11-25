---
title: "Bands with other"
date: 2026-11-25
issues: [6142]
---
**Bands with other** (#6142, CR 702.22b-c, j-k) — "bands with other [quality]" is an ability token with its quality in it, read off the layered ability list like banding (game/bands_with_other.go): `game.BandsWithOtherLegendary` ("bands with other legendary creatures", Cathedral of Serra and the Legends lands) and `game.BandsWithOtherNamed(name)` ("bands with other creatures named Wolves of the Hunt", Master of the Hunt's token). The attack declaration accepts a band of [quality] creatures anchored by a [quality] creature that has the ability (`bandsWithOtherBandLocked`, CR 702.22c), and the damage division changes hands when the blockers (or the blocked attackers) hold both such an anchor and another [quality] creature (`anyHasBandingLocked`, CR 702.22j-k). `RemoveKeywordsMod(game.KeywordBanding)` ("loses banding", Tolaria) strips every bands-with-other token too (CR 702.22b); "loses all 'bands with other' abilities" (Shelkin Brownie) is `RemoveKeywordPrefixesMod(game.KeywordBandsWithOtherPrefix)`. **Cards**: built by their OLD batches.
