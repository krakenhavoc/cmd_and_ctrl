---
title: "Modes chosen at random"
date: 2026-10-14
issues: [3399]
---
**Modes chosen at random** (#3399, CR 700.2, CR 603.3c) — "choose one at random —", from the Tidus, Yuna's Guardian deck request (#3246). A `game.ModeSpec` was always chosen by a player; a trigger's modes go through the `mode_pick` prompt.
- **The flag.** `ModeSpec.AtRandom`, built with `effects.ChooseOneAtRandom(...)`. Honoured in `queueModePickLocked` (`game/mode_choice.go`): after the bullets with no legal target are dropped (CR 700.2, so a bullet that can't be taken is never drawn), `pickModeAtRandomLocked` draws one with `ChooseAtRandomForEffect` on the "pick" stream keyed to the controller and the source, records it as chosen, logs it (`option_chosen`, "<bullet> (at random)"), and goes straight into the CR 603.3d target walk. No `mode_pick` prompt exists, so nobody is asked; the bullet's own target clause is the only question, if it has one.
- **Determinism.** The draw is the keyed stream, so an undo rewinds the counter and the redo draws the same bullet. Nothing is added to the snapshot: the flag lives on the catalog's `ModeSpec`, which is re-derived on restore, and the picked mode rides `StackItem.Modes` as any chosen mode does.
- **Scope.** Triggered abilities only, `Min == Max == 1`, not repeatable, no count raise or "hasn't been chosen" memory: `effects.Register` panics otherwise, and on a spell's or an activated ability's `Modes`. No printed card needs more. "If you discarded a card this turn, you choose one instead" (Typhoid Mary, Fractured) is a conditional chooser and stays out.
- **Cards:** Cult of Skaro and Umaro, Raging Yeti (Full). Summon: Magus Sisters now waits only on shield counters (#2000).
