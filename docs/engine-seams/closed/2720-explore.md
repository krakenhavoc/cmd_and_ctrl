---
title: "Explore"
date: 2026-10-08
issues: [2720]
---
**Explore** (#2720, CR 701.44, CR 111.10s) — `Game.ExploreForEffect(source, explorer ObjectRef, controller, then)` is the keyword action (`game/explore.go`).
- **The steps.** It reveals the top card. A land goes to hand through `routeCardToZoneLocked`, so a commander in the library still gets its CR 903.9 question. Otherwise a +1/+1 counter goes on the explorer through `AddCounterByThenForEffect`, so counter replacements settle first. Then the general confirm prompt asks whether to put the revealed card into the graveyard, and the answer is checked against the library as it is then. It is not a surveil prompt, so "whenever you surveil" payoffs don't fire.
- **The event.** `EventExplored` is its own kind and is emitted when the process completes on every branch (CR 701.44b): an empty library, a departed explorer and a declined question included. It is silent in the log, because the reveal, the move and the counter are already lines.
- **A departed explorer.** One that has left (CR 701.44c), or a flickered new object (CR 400.7), still reveals but gets no counter.
- **Card side** (`cards/effects/explores.go`): `Explores{Explorer, Then}`, `targetCreatureYouControlExplores`, and `MapToken()` (CR 111.10s), registered in the token catalog.
- **Cards:** Get Lost and Lodestone Needle // Guidestone Compass, both Full. Lodestone Needle leaves the craft row's waiting list.
- **Not built:** CR 701.44d's APNAP order for simultaneous explores, and a CR 614 window on the action (Topography Tracker).
