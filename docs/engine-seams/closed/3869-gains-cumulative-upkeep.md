---
title: "Testing for and granting cumulative upkeep"
date: 2026-10-16
issues: [3869]
---
**Testing for and granting cumulative upkeep** (#3869, CR 702.24) — another effect can ask whether a permanent has cumulative upkeep, and give it one. Every cumulative upkeep trigger carries `game.KeywordCumulativeUpkeep`, stamped by `effects.CumulativeUpkeepPaying`, and `game.HasCumulativeUpkeep` counts the instances through the trigger reader that sees both printed and ADR 0093 granted abilities; the target clause is `effects.DoesNotHaveCumulativeUpkeep()`. A granted cumulative upkeep is an ordinary bundle, `effects.CumulativeUpkeepGrant(key, cost)` or `CumulativeUpkeepGrantPaying` with `LifePayment`. Its source is the host, so the age counters, the prompt and the sacrifice are the host's, and two instances trigger separately and each count every age counter (CR 702.24b). **Cards:** Balduvian Shaman, Mana Chains, Breath of Dreams, Dreams of the Dead and Decomposition.
