---
title: "Keyword counters"
date: 2026-09-30
issues: [1753]
---
**Keyword counters** (#1753, [ADR 0101](decisions/0101-keyword-counters.md)): a flying, first strike, double strike, deathtouch, haste, hexproof, indestructible, lifelink, menace, reach, shadow, trample or vigilance counter gives its object that keyword (CR 122.1b). The engine reads the counter itself, as a layer-6 effect at the counter's own CR 613.7c timestamp (`Card.CounterStampedAt`, stamped by the one counter door), so a "loses all abilities" older than the counter leaves the keyword and a newer one takes it, a "can't have" beats it, proliferate renews it, and a card in another zone with a keyword counter has the keyword too. `b24KeywordCounterGrant` is gone: Vraska Joins Up and Sorin of House Markov lost their caveats, and Tekuthal, Solphim and Drivnod no longer carry a static for their own counters. Perennation joined the catalog. Still open: decayed, exalted and "hexproof from" counters, which wait on those keywords.
