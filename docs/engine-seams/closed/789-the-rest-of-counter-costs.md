---
title: "The rest of counter costs"
date: 2026-09-18
issues: [789]
pr: 958
legacy_order: 67
---
**The rest of counter costs** (#789, [ADR 0020](decisions/0020-activated-abilities.md) addendum 2026-09-18) - one `game.CounterRemovalCost` with TWO owners: `AbilityCost.RemoveCounters` and, new, `ManaAbilityShape.RemoveCounters`. It covers five printed shapes — from the source, from one other permanent you control, of any kind, a count the activator announces (`RemoveCountersXFromThis`, reaching the effect through `StackItem.Paid.CountersRemoved` and `ManaAbilityShape.ProducedForPaid`), and a removal split across permanents (`RemoveCountersAmong`, validated as a set like a crew payment). `AbilityCost.AddCounter` is the other direction, Devoted Druid's "put a -1/-1 counter on this creature": paid at announce, not replaceable (CR 614.16 — a counter-doubling replacement applies only to a counter placed by an effect), and refused by one `canPlaceCounterLocked` predicate (CR 118.3). The auto-tapper plans a counter-cost mana source only when it can both decide and afford the cost, so a Vivid land out of charge counters is not a five-colour source. Shipped on Vivid Creek, Vivid Grove, Ramos Dragon Engine, Mage-Ring Network, Devoted Druid, Hopeful Initiate and Iron Spider (whose caveat is cleared). The one shape it left open — an any-kind removal split across permanents — is closed by #943 below, and the counter-cost row has come out of the open table.
