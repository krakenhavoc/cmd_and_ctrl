---
title: "Permanents and spells of a colour are colorless sources of damage"
date: 2026-11-25
issues: [6230]
---
**Permanents and spells of a colour are colorless sources of damage** (#6230, CR 105.2c, 120.2, 702.16e) — Ghostly Flame's "Black and/or red permanents and spells are colorless sources of damage". A world rule the permanent carries on itself, `Characteristic.ColorlessDamageSourceColors` (WUBRG letters), read where a damage source's characteristics are taken (`game/damage_source_colorless.go`): `prepareDamageEventLocked` rewrites the event's `SourceLKI` with no colour, so protection and every prevention shield see a colourless source; `damageSourceViewLocked` does the same for a departed permanent's last-known record; and `damageSourceCandidatesLocked` doesn't offer such a source to "a red source of your choice". The objects keep their colours for everything else. Card helper: `effects.ColorlessDamageSources(colors, label)`. **Cards:** Ghostly Flame (b848) is built by its OLD batch.
