---
title: "Per-card entry counters on a return, and loyalty counters on entry"
date: 2026-10-21
issues: [4193]
---
**Per-card entry counters on a return, and loyalty counters on entry** (#4193, [ADR 0032 amendment 2026-10-21](decisions/0032-planeswalkers.md)): a batch put onto the battlefield can give each card its own counters by what the card would be on the battlefield (`ZoneEntryOptions.EntersWithCountersIfType`, judged by the CR 614.12 entry look-ahead after the window settles, so a copy it enters as counts), and the return stays one event (CR 603.6a). A planeswalker given loyalty counters as it enters now keeps its printed loyalty too (CR 306.5b): the drain stamps the starting loyalty before the entry's counters instead of the stamp reading them as the starting loyalty. **Cards:** Semester's End (Full).
