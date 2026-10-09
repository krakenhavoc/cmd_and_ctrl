---
title: "Two mana of different colors"
date: 2026-10-09
issues: [2558]
---
**Two mana of different colors** (#2558, CR 106.1a, CR 605.3b, CR 106.12a) — `game/mana_different_colors.go` is the rules side, `effects.DifferentColors(n)` the card side.
- **The grammar.** `"{W|U|B|R|G:2}"` is ONE produced-mana slot (`ProducedManaEntry.Distinct`) whose N mana must be N different colours. Two independent pipes, `"{W|U|B|R|G}{W|U|B|R|G}"`, would allow {U}{U}, which is stronger than printed (#259). The parser refuses a count below two, amounts, colourless, and fewer options than N; exactly N options expands to fixed slots. A reader that has not learned the field sees one ordinary pick and adds one mana, the weaker direction.
- **Named up front** (#1443). The view publishes one `color_options` list per mana and `different_colors: true`; `validateUpfrontManaColors` refuses a repeated colour before anything is paid. The client offers only pairs of different colours and never the stepper.
- **The prompt.** Without colours up front the activation queues ONE `mana_pick` carrying `ManaDifferent`. Each answer is struck from the next question's options and recorded in `ManaChosen`, and the choice gets a fresh ID. Nothing is added until the last answer, then all N together (CR 605.3b), so a "tapped for mana" trigger fires once (CR 106.12a). `ManaDifferent`, `ManaChosen` and `ManaLabel` are additive snapshot fields, so a restore with the pick half answered resumes it.
- **The auto-tapper.** `appendTapSource` offers one candidate per set of N colours (ten for a pair), mutually exclusive like a Gilded Lotus's per-colour candidates, and the plan carries the booked set to the executor (`plannedTap.DifferentColors`, `expandDifferentColors`). An effect's "add two mana of different colors" in auto-tap mode picks greedily without repeating (`pickDifferentColors`). A board with a mana-production replacement that would change the slot declines the source (the player taps it by hand).
- **Cards (4, all Full):** Firemind Vessel, Guild Globe, Component Pouch, and Interplanar Beacon, whose {1}, {T} planeswalker filter is restored and its caveat removed.
