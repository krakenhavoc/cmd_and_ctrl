---
title: "Vengeance (reminder text; not yet in the pinned CR)"
date: 2026-11-02
issues: [4460]
---
**Vengeance (reminder text; not yet in the pinned CR)** (#4460) — a Vengeance sorcery is "Cast only from your graveyard. Then exile it." The keyword is not in the pinned Comprehensive Rules (2026-09-25), so no rule number is cited and the engine implements exactly the reminder text. `isVengeanceHalf` reads it off the second face of an adventure-layout card whose type line is "Sorcery — Vengeance"; `castsOnlyFromGraveyard` joins it to the aftermath half (CR 702.127a) in `aftermathZoneRule` and in the spell's exile-as-it-leaves-the-stack stamp, so the half is refused from hand and from every other zone, opened in the graveyard, and exiled however it leaves the stack (resolved, countered or fizzled). `castAsAdventure` refuses it: the card is not an Adventure (CR 715), so its creature half is never castable from the exile it lands in. The same predicate serves the engine, the bot enumerator and the view. **Cards:** Hela, Mistress of Darkness // Dark Tutor (Full).
