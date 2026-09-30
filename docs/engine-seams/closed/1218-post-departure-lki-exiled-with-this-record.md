---
title: "Post-departure LKI / \"exiled with this\" record"
date: 2026-09-24
issues: [1218]
pr: 1402
legacy_order: 7
---
**Post-departure LKI / "exiled with this" record** (#1218) — closed on the 2026-09-24 re-triage (#1386): the row was fully stale. Its text, as it stood:

  > **What was missing:** **Closed for two of the three by #1218, and it turned out to need less than this row said.** The "exiled with this" record already existed by the time anyone looked again — `b27ExiledWith` (batch 27, after this row was written) reads it back off the event log for Duplicant, Angel of Serenity, Chrome Mox, Bag of Holding and Ossification — so Valakut Exploration and Currency Converter needed no new primitive at all: `b27ExiledWith` plus `game.CastPermission{Duration: WhileInZoneDuration()}` (the airbend primitive, exile_play.go) plus `PutCardsIntoGraveyardThenForEffect` (a new continuation sibling of `PutIntoGraveyardForEffect`, needed only so "deal that much damage" counts what actually landed after a CR 903.9 pause rather than before it, ADR 0013 §5t) covers both cards outright. The genuine gap was narrower than "no record of what a permanent exiled": it was **counters at departure**, which is not an exile question at all — `Characteristic` (the CR 603.10 LKI snapshot every leaves-the-battlefield trigger reads) deliberately excludes Counters ("belongs to other engine subsystems", characteristic.go), and `MoveCard`'s exit cleanup zeroes them a line after the snapshot is taken. `Game.lastKnownCounters` (triggers.go's `snapshotLKILocked`, alongside `lastKnownBattlefield`) and `LastKnownCountersForEffect` are the fix, kept on the GAME for the same reason the sibling maps are — the departing `Card` value is about to be overwritten out from under whatever holds a copy of it. It answers for a BYSTANDER's trigger too (The Ozolith is not the permanent that left), which its own `sourceLKI` could never have.
  >
  > **Cards waiting:** The Ozolith (#295) — counters-at-departure LKI, the actual new primitive. Shipped with no new primitive: Valakut Exploration (#307); Currency Converter (#387)

  The row is quoted from `docs/engine-seams.md` as of #1244, because it had already left the open table by the time #1402 moved the table into the registry; the two entries that follow this one (#1379 and #1396) closed what it still listed.
