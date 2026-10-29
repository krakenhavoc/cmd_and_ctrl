---
title: "Storied, the enduring story, and recruit"
date: 2026-10-29
issues: [4475]
---
**Storied, the enduring story, and recruit** (#4475, CR 702.195, CR 701.70a, [ADR 0096](decisions/0096-the-monarch-from-a-card-effect.md)'s 2026-10-29 amendment and [ADR 0020](decisions/0020-activated-abilities.md)'s) — The Hobbit's two new mechanics, with every card that waited on them.
- **The enduring story** is the city's blessing with a different count. The blessing's sweep became one table of earned designations (`game/earned_designations.go`): ascend counts every permanent (ten), storied counts artifacts, Sagas and/or legendary permanents, each once (three). `Player.EnduringStory` is never cleared. It rides the snapshot and the wire, the layer pass bumps on `EventEnduringStory`, and the client shows a book marker on the player panel. Card vocabulary: `YouHaveAnEnduringStory`, `WhileEnduringStory`, `SelfPumpWhileEnduringStory`, `SelfKeywordWhileEnduringStory`, `DoesntUntapUnlessEnduringStory`.
- **Recruit** is `effects.Recruit`: draw, then a keyed discard (a restore point), and a nonland card discarded makes a 1/1 white Human Soldier.
- **"You may pay {0} rather than pay the equip cost of the first equip ability you activate each turn"** is a derived twin row per equip ability (`game/equip_cost_alternative.go`). The twin's whole cost is the alternative and it keeps everything else of its row. It is addressed by `AltEquipIndexBase + index` and an `alt-equip:` ref, and counted by `PlayerTurnTally.EquipActivations`.
- **Cards (19, all Full):** Ori, Keeper of Songs; Óin the Brave; Bombur, Gentle Dreamer; Dáin, Lord of the Iron Hills; Thorin Oakenshield; Fíli the Pathfinder; Balin, Loremaster; Bifur, Melodic Rider; Kíli the Resourceful; Lake-town Lookout; Long Lake Nuisance; Patient Instructor; Esgaroth Garrison; Great Gilded Boat; The Queen of Dale; Bard's Company; Sound the Trumpets; Celebrate the Mountain-king; The Mountain-king's Return.
