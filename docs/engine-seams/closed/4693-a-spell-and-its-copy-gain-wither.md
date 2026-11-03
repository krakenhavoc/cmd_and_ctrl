---
title: "A spell and its copy gain wither on the stack"
date: 2026-11-03
issues: [4693]
---
**A spell and its copy gain wither on the stack** (#4693, CR 702.80a, 702.80c, 120.3d, 120.3a, 611.2, 613.1f, 707.2, 707.10c, [ADR 0107 §3](decisions/0107-state-triggers-rebound-disturb-and-damage-prevention.md#3-rebound-1854)) — closed with no engine change: the pieces were already in place. A keyword given to a spell on the stack is ADR 0107 §3's stack step (`GrantKeywordsToSpellForEffect`, a ScopedEffect pinned to the spell, #1854), a copy that gains a keyword once it exists is Spider-Verse's `copyGrant` (`CopyLastKnownSpellGainingKeywordsForEffect`, #4460), and the damage tail reads a spell's keywords off the stack through `HasKeyword`, which sees a granted keyword beside the printed ones, so the damage is -1/-1 counters on a creature and life loss on a player. Neither grant is a copiable value, so a copy another effect makes has no wither; both are data, so a table with the spell and its copy on the stack, or with the copy's new-targets prompt open, is a restore point. The stack view shows the granted keyword. **Cards:** Spinerock Tyrant (Full).
