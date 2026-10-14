---
title: "Effects that set a creature's types"
date: 2026-10-13
issues: [2144]
---
**Effects that set a creature's types** (#2144, CR 205.1a, CR 613.1d): "becomes a blue Frog" (Polymorphist's Jest) and "it becomes a Halfling Scout" (Frodo, Sauron's Bane) are `game.SetCreatureTypesMod`, a layer-4 `ScopedEffect` mod. The object's creature types are replaced and "is every creature type" goes with them; its card types, supertypes and other subtypes stay, so an artifact creature is still an artifact. It is separate from `SetTypesMod` (The Irencrag), which sets the card types too. Shipped on **Polymorphist's Jest** and **Frodo, Sauron's Bane**, both Full.
