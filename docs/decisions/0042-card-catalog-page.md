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

## Amendment (2026-09-24, #1276): the wiring is now checked against the printed card

`Completeness` and `Caveats` say what is wired. Nothing checked that the
wiring matched the card. The Wandering Emperor shipped from S27 to #1208
with three loyalty abilities at the wrong costs, one of them invented,
and four tests asserted the wrong card. The coverage package couldn't see
it because it reads the catalog, not the card.

`server/internal/cards/coverage/oracle.go` is the check.
`TestAbilitiesMatchOracleText` runs it in the default `go test`. For
every `Spec.Activated` ability it checks two things. **Cost:** the
label's cost prefix, or its keyword form ("Crew 3"), is printed on the
card. **Effect:** at least half of the label's words appear in an oracle
line printed at that cost. It also checks the reverse. **Missing:** every
loyalty line and every non-mana `<cost>:` line on the Spec's face has a
registered ability, or a caveat that quotes the cost ("the −7"), or a
pinned row in `knownOracleMismatches`. The effect half is what catches
the Emperor. All three of her costs are printed on the card, so a
cost-only check passes a permuted set.

Three decisions:

- **The oracle text is a checked-in, generated fixture.** CI has no
  Scryfall dump. `testdata/oracle_text.json` holds the printed text of
  every catalogued oracle ID, one card per line (~520 KB; one file per
  card since the #1542 amendment below). It is
  generated through `cards.Index.FindByOracleID`, the join the server
  itself uses, so placeholder printings never win. The regen command is
  `CMDCTRL_SCRYFALL_DUMP=… go test ./internal/cards/coverage/ -run
  TestOracleFixtureIsCurrent -update-oracle`. It uses its own flag
  because `-update` also rewrites the census block, which CI owns.
  `TestOracleFixtureIsCurrent` compares the fixture with the dump and
  runs nightly in `e2e-nightly.yml`, where the dump is available. A Spec
  that registers an activated ability but has no fixture entry fails
  at PR time, not nightly.
- **Normalisation, not equality.** Reminder text is dropped. U+2212 is
  treated as `-`. The card's name, its short name and "this
  creature"-style self-references all become `~`. Loyalty brackets,
  leading ability words and station thresholds are stripped. A modal
  line absorbs its bullets. Labels abbreviate, so the effect score is
  word containment, and the floor is 0.5. On the catalog, every correct
  label scored ≥ 0.8 and the Emperor's three scored 0.14, 0.33 and 0.20.
- **Out of scope, on purpose.** Mana abilities: their labels carry no
  cost, and basic lands have no Spec. Abilities granted in quotes. The
  missing-ability check on a second face that has no Spec of its own.

The first run over 2,380 Specs (548 abilities) produced 17 findings on
14 cards. Three of those cards had label typos (two findings each), and
the labels are fixed: Brass Squire and Vexing
Puzzlebox omitted `{T}`, and Beledros Witherbloom's label wasn't the
printed line. The other 11 are pinned. One is Heart of Kiran's
alternative crew, which the card prints as prose. Ten are unregistered
abilities on caveated cards. Five of those have a cost shape that
now exists (#1381).

## Amendment (2026-09-24, #1542): one oracle-text file per card

The fixture is no longer one file. It is a directory,
`server/internal/cards/coverage/testdata/oracle/`, with one generated
`<oracle_id>.json` per catalogued base oracle ID. Each file holds exactly
the row `oracle_text.json` held for that card, plus a newline. The split
was converted byte for byte, and the loaded map was checked to be
identical.

Why it changed: since #1496, every catalogued card needs text, so every
card PR regenerated the one file. Two PRs adding different cards touched
the same path. They had to merge one at a time, each regenerated on the
tip (#1491, #1496, #1509). Per-card files remove that conflict. A PR adds
only its own cards' files.

What each test does now:

- `TestOracleFixtureIsCurrent` is still dump-gated and nightly. It reports
  a missing, changed or non-canonical file, and a file for a card no
  longer catalogued. `-update-oracle` writes every file and deletes those
  stale ones. It never deletes the file of a catalogued card the dump
  doesn't know. A new flag, `-oracle-ids=<id>,…`, narrows both the check
  and the write to the named cards. That is the command a card PR runs,
  so it cannot touch another card's file.
- `TestOracleFixtureCoversRegistry` still runs without the dump. It fails
  when a registered card has no file, and it prints the exact
  `-oracle-ids` command for the missing IDs.
- The loader rejects any file that is not `<lower-case uuid>.json`. A
  skipped file would be a card whose text silently stopped being checked.
- `coverage.PrintedOracle()` embeds the directory (`testdata/oracle/*.json`).

Cost, measured on 2,430 files: loading from disk takes about 23 ms in
tests, against about 4 ms for the single file. The embedded parse takes
about 6 ms once per binary. We did not shard by ID prefix, because a
shard is a file that two card PRs can still share.
