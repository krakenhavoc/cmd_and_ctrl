---
title: "Split second read from the card"
date: 2026-09-24
issues: [1519]
pr: 1532
---
**Split second read from the card** (#1519, CR 702.61 / 702.61b / 708.4 / 707.2, [ADR 0007 amendment 2026-09-24](decisions/0007-stack-foundation.md) Decisions 11–12) — the rule was built in S13.1 and every consumer asked `Game.SplitSecondActive`, but the only writer was the sandbox `split_second` cast flag no client sends, so a Krosan Grip could be answered. Split second is now a `canonicalKeywords` token: the deck importer keeps Scryfall's "Split second", and `castHasSplitSecond` (`game/split_second.go`) stamps `StackItem.SplitSecond` at announce from `HasKeyword` on the announce copy, ORed with the flag (one writer) and refused for a face-down cast. `CastTimingOpenLocked` and `ActivationTimingOpenLocked` refuse under split second, so `castable_here` and an instant-speed row's `timing_closed` agree with the enumerator; mana abilities, triggers, foretell and turning face up still work. Krosan Grip ships `full`; Sudden Shock, Sudden Death, Sudden Edict and Sudden Spoiling join the catalog. Angel's Grace (#749) declares the keyword and drops its split-second caveat. The importer confirms the token against a bare keyword line (`narrowVariantKeywords`), so The Fearsome Flock's "split second level up" is not stamped on the creature.
