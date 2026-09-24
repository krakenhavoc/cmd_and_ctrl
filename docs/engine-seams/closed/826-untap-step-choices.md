---
title: "Untap-step choices"
date: 2026-09-18
issues: [826]
pr: 962
legacy_order: 96
---
**Untap-step choices** (#826, [ADR 0070](decisions/0070-untap-step-choices.md)): CR 502.3's "the active player determines which permanents they control will untap" is a real decision, and the untap step's turn-based action can pause for it. `Spec.UntapCaps` is "players can't untap more than N ⟨kind⟩ during their untap steps"; `Spec.UntapOptOuts` is "you may choose not to untap this during your untap step". Both families raise ONE prompt, the `untap_choice` kind, over the permanents actually in question — a cap that does not bind and an untapped opt-out permanent ask nothing. Several caps compose as one solver (`untapChoicePlan.legal`: at most N per cap, and no wasted untap, since untapping is otherwise mandatory), run through `ChooseCardsPrompt.Validate` so the submit path and the enumerator share one copy. Nothing untaps until the determination is complete, and `exitUntapStepLocked` is the step's one exit, called from the step-entry hook or from the prompt's continuation. Shipped on Winter Orb, Static Orb, Winter Moon (caps) and Rust Tick, Amber Prison (opt-outs, with their "for as long as this remains tapped" tap abilities declared as caveats). Still open: the per-source linked-object field those tap abilities need (ADR 0058 Decision 8's second bullet, designed in ADR 0070 Decision 6 and not built), Smoke and Stoic Angel (a restriction with a cost, not a cap), Dovin Baan's emblem, and exert.
