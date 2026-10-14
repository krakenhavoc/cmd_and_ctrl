---
title: "Putting a counter on another permanent as this enters"
date: 2026-10-13
issues: [1893]
---
**Putting a counter on another permanent as this enters** (#1893, [ADR 0170](../../decisions/0170-choosing-other-permanents-as-a-permanent-enters.md), CR 614.1c / 614.12 / 614.13a) — Phylactery Lich, and Sin, Unending Cataclysm from the Tidus deck request (#3246). `Spec.AsEnters` ran after the permanent arrived and chose no permanent; ADR 0098's `EntryCardChoice` asked before it arrived but only revealed, discarded or sacrificed.
- **The action.** `EntryCardChoice.Action` gains `EntryCardChoose`: the engine does nothing to the named permanents, and the declaration's `Then` is the effect. Asked through a new prompt kind, `entry_choose_permanents`, which joins its three siblings in every table (entry kinds, card-set picks, the table gate, leave-game, the action router, the enumerator, the view). `AnyController` widens the candidates from the chooser's permanents to anyone's, and `AnyNumber` makes the floor zero. CR 614.13a's exclusion of the entering permanent is the existing rule.
- **`Then` sees the entry.** `Then` is now `func(g, ev, picked)` for every action, so it can give the entering permanent counters (`ev.AddCounterAtETB`, which Doubling Season and Hardened Scales see). What it does to other permanents settles inline with the `MustSettleNow` counter forms, because the entry cannot pause twice. This also unblocks Indominus Rex, Alpha's follow-on.
- **A suggested answer.** `EntryCardChoice.SuggestKey` names a function filed with `game.RegisterEntryChoiceSuggestion` (a key, not a closure, for the ADR 0041 ratchet), and its answer rides the prompt as `ChooseSuggested` (proliferate's field): the client pre-selects it, the enumerator always offers it, and the heuristic bot answers with it.
- **Restore points.** Like its siblings, the open prompt holds a resume frame and is not a restore point; the undo clone carries it. No snapshot field.
- **Cards:** Phylactery Lich (Full), Sin, Unending Cataclysm (Full).
