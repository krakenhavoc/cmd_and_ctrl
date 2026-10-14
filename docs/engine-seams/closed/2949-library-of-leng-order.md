---
title: "Library of Leng's multi-card discard order"
date: 2026-10-14
issues: [2949]
---
**Library of Leng's multi-card discard order** (#2949, CR 401.4, [ADR 0088](decisions/0088-ordered-library-placement.md)'s 2026-10-14 amendment) — when one effect discards several cards and Library of Leng puts more than one on top of the library, the owner now arranges them. Once the batch's last leg has landed, `discardBatchLocked` raises the `put_in_library` prompt (placement `top`, reordered in place) over exactly the batch's cards that are the top of the library, and the effect's "then" waits for it. Discard events, "discarded this way" and payoffs are unchanged, and opponents do not learn the order. **Cards:** Library of Leng (Full; its ordering caveat is gone).
