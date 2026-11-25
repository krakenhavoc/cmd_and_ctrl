---
title: "A draw knows the cycling ability that made it"
date: 2026-11-25
issues: [6169]
---
**A draw knows the cycling ability that made it** (#6169, CR 702.29a, CR 614.1a) — an activated ability item carries `StackItem.Cycling`, stamped from `ActivatedAbilityShape.Cycling` at activation and kept by a copy (CR 707.10). A draw made while such an item resolves, by its controller, names the cycled card on its replacement event (`ReplacementEvent.DrawCycled`; uuid.Nil for every other draw), carried through an amount replacement's descendants, a paused window and a parked `DrawDone` (`DrawDone.Cycled`). A draw replacement reads it for "if a cycling ability of another nonland card would cause you to draw a card", and a draw-instead body reads the cycled card's characteristics from it. **Cards:** Unpredictable Cyclone (Full).
