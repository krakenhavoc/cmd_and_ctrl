---
title: "Counter removal as an activated-ability cost"
date: 2026-09-16
issues: [625]
pr: 699
legacy_order: 112
---
**Counter removal as an activated-ability cost** (#625) - `AbilityCost.RemoveCounters`, written `RemoveCountersFromThis(kind, n)` or `RemoveCountersFrom(kind, n, label, preds…)`, with kind `""` meaning "a counter" of any kind ([ADR 0020](decisions/0020-activated-abilities.md) addendum). Paid at announce, not replaceable, and not a loyalty activation. A printed 0/0 that pays with its last counter dies, and an uncoded `*` creature does not, by the toughness rule in [ADR 0007 §7](decisions/0007-stack-foundation.md) (#683, and #690/#691 for the rest of that rule: `game.Card.ToughnessIsKnown` is the one gate CR 704.5f has). Shipped on Heart of Kiran (its "rather than pay" crew, as a second ability entry), Dragon's Hoard, Mikaeus, the Lunarch, Benevolent Hydra and Fain, the Broker. The rest of the counter-cost row above is still open: adding a counter, mana abilities, and removals split across permanents.
