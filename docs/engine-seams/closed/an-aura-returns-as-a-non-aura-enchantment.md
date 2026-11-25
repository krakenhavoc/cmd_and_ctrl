---
title: "An Aura returns as a non-Aura enchantment"
date: 2026-11-25
issues: [6235]
---
**An Aura returns as a non-Aura enchantment** (#6235, CR 303.4f, 205.1a, 613.1d, 613.1f) — Takklemaggot's "that creature's controller chooses a creature that this card could enchant … If they don't, return this card to the battlefield under your control as a non-Aura enchantment. It loses "enchant creature" and gains …". ADR 0210's entry overwrite (`game.EntersAs`) gains `LoseSubtypes` (layer 4), `LoseKeywords` (layer 6, keyword families) and `ChosenPlayer` (set on the permanent as it lands, "that player"); an entry that loses "Aura" is no Aura entering, so CR 303.4f asks nothing. `game.CreaturesAuraCardCouldEnchantForEffect` is "a creature that this card could enchant" for an Aura card off the battlefield (no targeting). Card helpers: `effects.AuraReturnsOnHostDeath(name, grantKey)` (the dies trigger and the chooser's prompt) and `effects.ChosenPlayerUpkeepDamageGrant(key, n, name)` (the gained trigger). **Cards:** Takklemaggot (b869) is built by its OLD batch.
