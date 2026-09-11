# ADR 0042 — The public card catalogue, and declaring completeness

**Status:** accepted
**Date:** 2026-09-11
**Sprint:** Post-S30 — Rolling deck-driven catalog growth
**Issue:** [#410](https://github.com/krakenhavoc/cmd_and_ctrl/issues/410); follow-up to [#338](https://github.com/krakenhavoc/cmd_and_ctrl/issues/338) / [#350](https://github.com/krakenhavoc/cmd_and_ctrl/issues/350)

## Context

We want a public page that publishes every card the engine has wired
up — "all the cards we have configured and should be working in their
entirety."

The last clause is the hard part. Not every registered card is fully
implemented. Since S14 the catalog has carried a convention of
*declared simplifications*: a card ships with a comment explaining the
clause the engine does not model (AGENTS.md §7). A page that listed a
partially-implemented card as "working" would be worse than no page at
all, because it makes a promise the engine does not keep.

So the page has to distinguish, visibly, between cards implemented in
full and cards that work with a caveat. The question was what to
detect that from.

## The signal we had was not good enough

The obvious answer is to parse the existing comments. Surveying all
381 card files says no:

- **The wording is free-form.** "No simplification." / "No
  simplifications." / "No other simplification." / "No simplification
  remains." all appear, and so do at least six ways of announcing a
  gap ("Sandbox simplification:", "DECLARED SIMPLIFICATION —", "Shape
  notes and simplifications:", …).

- **Silence is ambiguous, and it is the majority case.** 219 of the
  504 registry entries sit in files that never mention the subject.
  Reading them, Mind Stone and Blaze turn out to be complete
  implementations whose authors saw nothing worth saying — and there
  is no way to tell those from a card nobody ever checked. A parser
  must guess, and either guess is a lie.

- **The notes go stale in both directions.** This is what
  [#338](https://github.com/krakenhavoc/cmd_and_ctrl/issues/338) and
  [#350](https://github.com/krakenhavoc/cmd_and_ctrl/issues/350) were
  about. Reviewing the 87 files that do mention a simplification,
  **24 of them turn out to have no current gap at all** — Awakening
  Zone, Birds of Paradise, Wash Away and others carry paragraphs whose
  own last sentence says the note outlived the fix. A parser keyed on
  "this file says 'simplification'" marks all twenty-four as broken.

Parsing comments would therefore publish a page that is wrong in both
directions, and would enshrine the staleness that two issues have
already been spent on.

## Decision

**1. Completeness becomes a field on `Spec`, and its zero value means
"unreviewed".**

```go
Completeness Completeness   // CompletenessUnreviewed | Full | Caveats
Caveats      []string       // player-facing, required iff Caveats
```

Three states, not two. "Nobody has checked this" is a real answer and
it gets its own bucket, on the wire and on the page. The zero value
carries it so that a spec that forgets the field, a card file
copy-pasted from another, and a card written on a branch that predates
this ADR all land on the conservative answer without anyone thinking
about it.

`Register` panics on a `CompletenessCaveats` with no `Caveats`, and on
`Caveats` set under any other value. Those are the two mistakes a
machine can catch; both fail at boot rather than in a browser.

**2. `Register` deliberately does NOT reject an undeclared spec.**

The stricter design was considered and rejected. Cards arrive in
batches of thirty from several branches at once; a hard gate turns
every one of those into a merge-blocking chore, and the predictable
outcome is a reviewer stamping `CompletenessFull` to make the build go
green. An unreviewed card that says it is unreviewed is honest; a card
falsely marked complete is the failure this whole mechanism exists to
prevent. The lazy path is made safe and visible rather than fast and
wrong.

**3. The backfill declares only what a human actually decided.**

The 504 registry entries were classified as follows, and the page's
numbers are exactly this and nothing more:

| Bucket | Entries | Where it came from |
| --- | ---: | --- |
| `CompletenessFull` | 141 | 84 files whose comment plainly declares "No simplification", plus the 24 reviewed files found to have no current gap |
| `CompletenessCaveats` | 144 | 63 of the 87 reviewed files, each read against its code and given a one-sentence player-facing caveat, plus 3 cycle/helper files adjudicated by hand |
| `CompletenessUnreviewed` | 219 | everything else — left at the zero value |

(`len(effects.All())` reports 505 inside the test binary: `flicker_test.go`
registers a probe spec. 504 is the production catalog, and what the
page serves.)

Nothing was inferred. A card is `Full` only where someone looked and
said so. The 219 unreviewed entries are not a defect in this ADR; they
are the pre-existing state of the catalog, now visible instead of
implicit.

**4. Prose comments stay exactly where they are.**

The field is not a replacement for the doc comment — it is the
machine-readable half. The comment keeps the engineering reason, at
length, for the next person editing the file; the field carries the
verdict and one sentence a player can act on. `Spec.Caveats` is tested
to be player-facing prose (a sentence, no engine jargon).

## The page

- `GET /catalog` — public JSON, `server/internal/catalog`. Joins
  `effects.All()` against the Scryfall index on `oracle_id`.
- `GET /catalog/image/{id}` — public card art, scoped to catalogue
  cards only.
- `#/catalog` in the client, listed in App.svelte's `isPublic` set.

Three things worth recording about the shape:

**Its own package.** The join needs both `internal/cards` and
`internal/cards/effects`, and effects is otherwise blank-imported
exactly once (by `main`) so that nothing depends on it. A new package
keeps that single import explicit and keeps a public read-only surface
out of the lobby's session-shaped handler file.

**Why the image route is not `/cards/{id}/image`.** That route is
session-gated and should stay that way: it accepts any Scryfall UUID
in the dump, so unauthenticated it would be a general-purpose image
proxy for ~35k cards. `/catalog/image/{id}` accepts only the
representative printing of a card the catalog registers — a few
hundred UUIDs, fixed at build time, already published by
`GET /catalog`. That is exactly the set the page renders, so opening
it adds no reachable surface. It reuses the same `ImageCache`, so both
routes share one disk cache and one SSRF-guarded fetch.

**Not a dev feature.** No `requireDev`, no `appenv.Features` entry.
This ships in production, which is the whole point.

## Consequences

- The `@api path` matcher in `deploy/Caddyfile`, the proxy map in
  `client/vite.config.ts` and `API_PATH` in
  `client/src/sw/service-worker.js` all gain `/catalog`. All three in
  this PR, as the Caddyfile comment demands.
- `cards.Index` gains a `byOracle` map and `FindByOracleID`. There was
  no oracle-id lookup before; the catalog is keyed by oracle_id, so
  the join needed one.
- `isPlayablePrint` now also rejects the `"Card"` / `"Card // Card"`
  placeholder type lines. AGENTS.md §7 warns about these and the
  layout/set_type switches did not catch all of them.
- **60 modal double-faced cards register only their land back**
  (`<oracle_id>#1`) and nothing automates the creature front. No card
  file can declare that gap, because no card file owns that face, so
  the catalog derives it: a card with an unautomated printed face gets
  an automatic caveat naming it. This was invisible before the page
  existed.
- Adding a card without touching `Completeness` still works, and
  publishes as unreviewed. Reviewing the 219 is ordinary follow-up
  work, one card at a time, and the page is the worklist.
- `GET /catalog` rebuilds the whole response per request: 504 entries,
  228 KB raw and 53 KB after Caddy's gzip, with a five-minute browser
  cache. Measured, not estimated. At this size a cache would be a
  worse bug (stale after a deploy) than the pass is a cost.

## Alternatives rejected

- **Parse the doc comments.** See above: wrong in both directions on
  real data.
- **A side table keyed by oracle ID, in one file.** Smaller diff, no
  conflict with concurrent card branches. Rejected because a second
  place to update is a second place to forget, which is precisely the
  #338/#350 failure mode. The per-file diff turned out to be cheap
  anyway: one-file-per-card means 173 single-hunk edits that merge
  cleanly against branches adding new files.
- **Two buckets, folding unreviewed into "has caveats".** Tempting —
  it is conservative and it makes the page read better. Rejected
  because it is also false: it tells a player Mind Stone has a
  known gap when it does not, and it erases the distinction that makes
  the remaining work legible.
