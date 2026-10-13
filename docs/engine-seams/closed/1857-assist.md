---
title: "Assist"
date: 2026-10-13
issues: [1857]
---
**Assist** (#1857, CR 702.132a, CR 601.2f, CR 601.2h, CR 800.4f, [ADR 0150](decisions/0150-other-players-pay-any-player-offers-and-assist.md) PR 3) — `game/assist.go`.
- **The announcement.** `cast_spell` takes `assist_player` (`CastSpellParams.AssistPlayer`): another player still in the game, on a spell with the `assist` keyword (now canonical; the importer stamps it, so a printed assist works with no catalog entry) whose total cost has a generic part. Anything else is refused.
- **The park.** `castSpellLocked` validates the whole announcement, then, where #1397's commander-owner question parks, asks the helper a `pay_amount` (`resource: "mana"`, `unit: "assist"`, 0 up to the smaller of the total cost's generic part and what their pool and auto-tapper can make). Nothing is tapped, paid or moved. The generic is the TOTAL cost's (commander tax and increases raise it, reductions lower it). The prompt carries `PendingChoice.Assist` (caster, spell, name, cost, generic), public on the wire as `assist`. Its continuation rides the existing `chooseValueResume` frame, so no closure route was added.
- **The re-run.** The helper's answer re-runs the cast with the amount attached. `prepareAssistLocked` plans the helper's share and checks the caster's remaining cost before either is paid; if either is short nothing happens and the caster gets an `effect_error` line. The helper's mana joins `PaidCost.Mana`, and `PaidCost.AssistedBy` / `AssistPaid` record who helped and how much.
- **Leaving and withdrawing.** A helper who leaves pays 0 and the cast goes on unassisted; a caster who leaves has the prompt withdrawn. `cancel_assist` lets the caster withdraw it (owner decision 5).
- **Client.** The cast modal's Assist row picks the helper (bots marked "(always declines)"); the caster's dock shows "Waiting for … to help pay" with Cancel request; the helper gets the stepper.
- **Bot.** The heuristic helper answers 0, and the enumerator never offers `assist_player`.
- **Cards:** Fan Favorite loses its caveat (Full). The other 15 assist cards are #3204.
