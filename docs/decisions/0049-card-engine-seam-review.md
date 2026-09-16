# ADR 0049 — Architecture review of the card ↔ engine seam (September 2026)

**Status:** Accepted · 2026-09-15 · Discussions [#557](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/557) (findings), [#558](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/558) (D1), [#559](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/559) (D2), [#560](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/560) (D3), [#561](https://github.com/krakenhavoc/cmd_and_ctrl/discussions/561) (D4, open) · Tracking issue #585

## Context

The card roadmap paused after batch 36 with 1,534 specs on `main` and
about 2,400 more scheduled. Before those went in, the question was
whether the way cards plug into the engine is the way to keep building:
cards should hook in where things happen, carry flags for the hooks they
use so the engine consults only relevant cards, wire new references into
existing machinery rather than adding parallel paths, and never need the
same change in twelve places.

The review (#557) found that the design already exists — `effects.Spec`
slots are the flags, `Watches` on triggers and replacements is the
pre-filter, and the event log is the hook surface — and that neither
runtime nor compile cost is a problem: on an 80-permanent board a tap
event costs about 6 µs through the whole listener chain and a full layer
recompute about 46 µs; an incremental build after touching one card file
takes under half a second.

What had drifted was the card layer. A clone detector over the effects
package found 87 exact-duplicate groups, 218 same-code-different-literal
groups and 326 structural groups (about 8,700 redundant lines) across
3,708 function and closure bodies. 757 batch-prefixed helper functions
existed against 46 shared primitives; 64 were referenced by nothing and
532 by exactly one file. 262 of the 559 files declaring a trigger were
one trigger wrapping one primitive. The root cause was the parallel-agent
rule "never edit shared files", which made every batch grow its own
vocabulary. Separately, `Spec.OnETB` ran off the stack and was used for
printed "When ~ enters" triggers in nine files.

## Decisions

### D1 — Authoring model (#558, option A)

Stay in Go. Add constructor functions returning `game.TriggeredAbility`
for the printed shapes (`WhenThisEnters(...)`, `WheneverYouCast(...)`,
`AtYourUpkeep(...)`); move shared code into mechanic-named, append-only
vocabulary files; run a promotion pass that merges the duplicate groups
and drops the `bNN` prefix; add a CI gate that fails a PR introducing a
new exact duplicate of six or more lines. A data-driven card format was
considered and deferred: it would add a second interpreter and a
1,500-file conversion for a benefit the constructors mostly deliver.
Issues #579, #583, #584.

### D2 — Engine seams that delete card-side copies (#559, all approved)

1. A typed per-player `TurnTally` on `Game`, reset with the existing
   per-turn tallies, replacing ~30 helpers that rescan the event log
   (#586).
2. `OncePerBatch` on `TriggeredAbility` plus a `BatchID` on events,
   replacing four hand-rolled "one or more" dedup helpers (#587).
3. Let the harvester watch the generic `EventStepBegan`; retire the
   per-step event kinds (#588).
4. A token table on `TokenSpec` replacing 111 near-identical
   constructors (#581).
5. Delete the 64 unreferenced helpers; run golangci-lint in CI (#582).
6. `docs/engine-seams.md` as the seam registry, and a `seam` label
   (#580).

### D3 — ETB paths (#560, option A) — this PR

`Spec.OnETB` is renamed `Spec.AsEnters` and is the CR 614.12 slot only:
"As this permanent enters, choose …" effects, which are not triggers and
correctly happen off the stack. Every printed "When ~ enters" declares a
`Triggered` ability watching `EventETB`, so it uses the stack and can be
answered. The two remaining tap-after-entry workarounds use
`SelfEntersTapped()`. `game.ETBEffectHook` keeps its name until D4
replaces the hook variables. Issue #578.

### D4 — Plumbing (#561, option A, approved 2026-09-16) — issue #622

One precomputed `game.CardDef` per card, built at `effects.Register`;
`game.CatalogLookup` is the only hook the catalog sets. The per-slot
variables remain as defaults that read the `CardDef`, because two dozen
test files stub them individually; nothing in the catalog assigns them.
The review's benchmark is committed as `bench_test.go`.

## Consequences

- The batch brief and AGENTS.md §7 change under D1: "append to a
  vocabulary file" replaces "never edit shared files"; the `bNN` prefix
  goes.
- The roadmap resumes at batch 37 after the constructors (#579) land, so
  new batches write constructors rather than closures.
- Nine cards that resolved their ETB instantly now give every player a
  response window; three edicts marked complete were wrong before and
  are right after.
- Things the review looked at and rejected, so they are not re-proposed:
  splitting the effects package (compile times say no), a per-event-kind
  index of battlefield permanents (invalidation cost exceeds the
  microseconds saved), an instance-ID index over zones (too many
  slice-moving writers).
