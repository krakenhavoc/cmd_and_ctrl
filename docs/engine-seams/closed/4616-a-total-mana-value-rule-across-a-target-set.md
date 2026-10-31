---
title: "A total mana value rule across a target set"
date: 2026-10-30
issues: [4616]
---
**A total mana value rule across a target set** (#4616, CR 601.2c, CR 608.2b, [ADR 0019 amendment 2026-10-30](decisions/0019-structured-targeting.md)) — "Up to two target creatures with total mana value 6 or less" (The Super Hero Civil War, chapter I).
- **`TargetSpec.Total`** (`game.TargetTotal`, built with `effects.TotalManaValueAtMost(n)` and attached with `.WithTotal`) is the third set rule beside `Different` and `Same`: the picks' mana values summed may not exceed the ceiling. A closed enum, so the closure ratchet gains no route. The announce gate sums the set, fillability counts the smallest values first, the retarget offer keeps to what fits beside the staying slots, the enumerator never extends a set past the ceiling, and the view ships `total: {label, max, values}` for the picker, which greys what would pass it and shows the running total.
- **Resolution:** the picks still individually legal are summed again, and when they no longer fit NONE of them is legal, so the chapter does not resolve. That is the printed ruling on the family (Nethroi, Reunion of the House: "the entire selection of targets is illegal"). A pick that left is dropped first and adds nothing. A card whose cost the engine cannot read fits no total.
- **Cards (1, Full):** The Super Hero Civil War.
