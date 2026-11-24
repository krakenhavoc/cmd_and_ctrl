---
title: "One announced X across an activation cost, and a trigger counted by its event's X"
date: 2026-11-24
issues: [6085]
---
**One announced X across an activation cost, and a trigger counted by its event's X** (#6085) — an activation cost with {X} beside a non-mana X ("{X}, {T}, Remove X charge counters", "{X}, {T}, Tap X untapped artifacts") pays every X with the one announced value (CR 107.3a). `effects.RemoveCountersXFromThisSameX` (`CounterRemovalCost.CountIsX`) removes the announced X counters, and Register accepts {X} beside `TapXUntapped`; the client shows the X picker as the payment and the enumerator offers the largest payable X. `effects.UpToXTargetsFromTheEventAmount` (`TriggeredAbility.XFromEventAmount`) sizes a trigger's CountFromX clause by the triggering event's amount (the cycling X on EventCycle), with no targets at X = 0, never an unbounded Max (#259). Data only, no snapshot change. **Cards:** Talon of Pain, Necron Overlord, Rampaging War Mammoth (Full).
