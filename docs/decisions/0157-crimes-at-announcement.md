# ADR 0157 — Crimes at announcement

Date: 2026-10-13
Status: Accepted
Sprint: S58 — Deck requests, October batch (follow-up)
Issues: #2944, #2910

## Context

CR 700.13 defines a crime when a player casts a spell, activates an
ability, or puts a triggered ability on the stack with at least one
qualifying target. `EventBecomesTarget` also occurs when a copy is created
or an existing object's targets change. Using that event for crimes
incorrectly triggers Magda and Lazav without a new announcement and misses
opponents' stack abilities, which have no card-backed target object.

## Decision

1. Add `EventCrimeCommitted`, once per qualifying announcement, with the
   announcing controller in `Actor`, the source in `Source`, and the announced
   object in `StackItemID`. No qualifying-target count or persistent tally is
   needed. A separate later announcement is another crime.
2. The common announcement helper checks the complete target set while the
   announcement exists, then emits ordinary becomes-target events and the
   single crime event. It is called only by successful casts, catalog and
   sandbox activations, and actual placement of targeted triggered abilities.
   Trigger placement keeps the existing APNAP batch boundary and deferred
   target announcement flag. Copies and retargeting retain their ordinary
   becomes-target events and never call the crime announcement helper.
3. Qualifying targets are opponents, permanents/spells/abilities controlled
   by opponents, and cards in opponents' graveyards. Stack metadata is checked
   before card zones because an ability is a stack object without a card.
   Own objects and cards outside those zones do not qualify. The test is
   performed once over the entire chosen set, before event listeners run.
4. Catalog cards use the shared `YouCommittedCrime` predicate. Magda and
   Lazav migrate to the new event; Forsaken Miner supplies a graveyard proof
   card using the existing resolution-time payment and object-identity guard.
5. This is an additive event kind using existing serialized event fields.
   No snapshot structure, unfinished choice frame or schema version changes.
   Historical becomes-target events are not replayed as crimes on restore.

## Validation

Cover copied spells/abilities and retargeting without a crime, multiple
qualifying target slots with one crime, separate announcements, every target
category, own objects, and both harvested and sandbox trigger placement.
Preserve becomes-target events and ward; verify graveyard-only Miner triggers,
optional payment at resolution, and a source leaving and returning as a new
object. Run the affected engine/catalog suites and fixture/coverage guards.

## Number reservation

Checked all 14 remote refs after `git fetch --all --prune` on 2026-10-13.
0157 was the first unused number, with permanently unused 0005, 0024, 0029
and 0030 excluded. No existing ADR was renumbered.
