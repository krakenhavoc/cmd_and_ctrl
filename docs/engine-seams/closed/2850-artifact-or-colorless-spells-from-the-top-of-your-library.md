---
title: "Artifact and colorless spells from the top of your library"
date: 2026-10-09
issues: [2850]
---
**Artifact and colorless spells from the top of your library** (#2850) — `game.PermissionFilter` gained `ArtifactOrColorlessOnly`, the OR of "is an artifact" and "is colorless" read from the card's effective characteristics. Paired with `NonLandOnly` it is Mystic Forge's "artifact spells and colorless spells": a colored artifact and a colorless non-artifact both qualify, a land never does (a colorless land is a land play, not a spell). The field is additive snapshot data. Shipped on Mystic Forge, whose `{T}, Pay 1 life: Exile the top card of your library` pays both costs at announcement and exiles the current top card as it resolves, and on Deathbringer Thoctar in the same change.
