---
title: "Creatures that could have attacked this turn"
date: 2026-11-25
issues: [6211]
---
**Creatures that could have attacked this turn** (#6211, CR 508.1a, 508.1c) — `TurnTally.CouldAttack` is every creature OBJECT that could have been declared as an attacker in a declare attackers step this turn, written as each attack declaration passes its checkpoint, beside the CR 508.1d had-to-attack record (game/could_attack.go): the declared attackers, and every other creature the active player controls that `AttackerEligible` (the one copy of the CR 508.1a/c test the verb and the enumerator share) would have let them declare. Read through `Game.CouldHaveAttackedThisTurnForEffect`: a creature that was tapped, summoning sick or under a "can't attack" restriction at every declaration, one that entered or changed hands afterwards, and one that left and came back (CR 400.7) could not attack; a turn with two combats keeps both declarations. Season of the Witch's "destroy all untapped creatures that didn't attack this turn, except for creatures that couldn't attack" is this reader beside `AttackedThisTurn`. **Cards:** Season of the Witch, built by its OLD batch (b862).
