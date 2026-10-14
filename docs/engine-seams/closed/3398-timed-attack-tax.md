---
title: "Attack taxes that last for a while"
date: 2026-10-14
issues: [3398]
---
**Attack taxes that last for a while** (#3398, CR 508.1a, CR 611.2) — Summon: Yojimbo, from the Tidus deck request (#3246). Attack taxes (ADR 0080) were statics only, read off the defending player's battlefield; a resolving ability could not make one that lasts.
- **Storage.** A new scoped-effect mod kind, `ModAttackTax` (`game/attack_tax_duration.go`), pinned to the protected player (`Mod.Player`) and carrying the mana for one attacking creature (`Mod.Text`), timed by an ADR 0063 `Duration`. It reuses existing fields, so the snapshot gains no key; it is a new mod kind, so a binary from before it refuses a restore point naming it. Registration and restore both refuse a record with no player or an unparseable cost.
- **Reading.** `priceAttackDeclarationLocked` takes the live records as a second source beside the battlefield walk, asking the duration at read time so "until your next turn" ends as that turn begins. The declaration verb, the legal-move enumerator and the bot all go through the one price, so a timed tax stacks with Propaganda and is charged and refused exactly as it is. It defends the player, not their planeswalkers (the zero `AttackTaxScope`).
- **Card side.** `Game.AttackTaxForEffect(source, defender, mana, duration, label)`.
- **Not covered.** War Tax ("this turn, creatures can't attack unless their controller pays {X} for each attacking creature they control") taxes every attack, whoever is defending, with X paid into the activation; this seam protects a named player.
- **Cards:** Summon: Yojimbo (Full).
