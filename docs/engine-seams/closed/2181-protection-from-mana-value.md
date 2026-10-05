---
title: "Protection from mana value"
date: 2026-10-05
issues: [2181]
---
**Protection from mana value** (#2181): "protection from mana value N or less" (Reaver Titan) is a new `game.ProtectionQualityKind`, `ProtectionQualityManaValueAtMost`, with its own parser token. A source's mana value is not a characteristic a layer computes, so the source snapshot gained `Characteristic.SourceManaValue` (and `SourceManaValueKnown`, so an unreadable cost matches nothing). CR 202.3e: a spell counts the X it was cast with, through `TargetSource.X` at announce and re-check and `damageSourceLKILocked` for damage; a token that is no copy is 0 and a copy has the copiable cost's value. Every protection reader (targeting, attachment, damage, block) already asked `ProtectionQuality.Matches`, so none changed. Wire token `mana_value_at_most`; the bot reads the cost it already sees. Shipped on **Reaver Titan**.
