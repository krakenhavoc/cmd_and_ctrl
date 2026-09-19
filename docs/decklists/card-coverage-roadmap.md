# The card-coverage roadmap — the next 6000 Commander cards

Source: **`edhrec_rank`** on every card in the local Scryfall bulk dump
(`data/scryfall/default-cards.json`, refreshed 2026-09-09). Lower rank
means more played. Nothing here was scraped, recalled, or guessed — one
pass over the dump and one probe of the live registry produced every
number below.

This is the whole-format sequel to
[the top-100 staples triage](top-100-commander-staples.md): that file
ranked the first 100 cards of the gap and shipped 32 of them, this one
ranks the next **6000** and splits them into 60 tracked batches of 100.
Tracking issue: **#293**.

The 6000 were ranked in three passes, far apart in catalog terms and
identical in method. **Batches 01–20** (ranks 9–2234) came from the
first pass, against a 288-card catalog. **Batches 21–40** (ranks
2235–4253) came from the second, against a 504-spec catalog on
`origin/main` at `459dea6`. **Batches 41–60** (ranks 4254–6289) came
from the third, against an 844-spec catalog on `origin/main` at
`b5a3055` — the same dump, the same filters, the same detectors, each
pass continuing the same ranked list from where the previous one
stopped. No card appears in more than one batch.

## Progress

The map is being walked. This section is the running tally; the
per-card detail lives on each batch's issue, and every simplification
is a comment on its card file.

| Batch | Issue | Registered | Skipped (declared) | Still blocked | Notes |
|---|---|---:|---:|---:|---|
| 01 | #294 | **29** | 2 | 69 | first pass — the "no new machinery" group, plus 3 cards two engine changes unblocked |
| 02 | #295 | **38** | 19 | 43 | landed in three streams — #351 (25), #358 (11), #356's mana pipeline (Boros Signet, Mox Opal) |
| 03 | #296 | **37** | 11 | 52 | PR #361 — the "no new machinery" group, one pass |
| 04 | #297 | **29** | 9 | 62 | PR #362 — the "no new machinery" group plus Graven Cairns, unblocked by #356 |
| 05 | #298 | **23** | 12 | 65 | PR #437 — the "no new machinery" group plus Dualcaster Mage, unblocked by #419's spell copies |
| 06 | #299 | **29** | 6 | 65 | PR #439 — the "no new machinery" group, one pass |
| 07 | #300 | **28** | 12 | 60 | PR #438 — the "no new machinery" group plus Kodama of the West Tree, unblocked by #379's attachments |
| 08 | #301 | **34** | 9 | 57 | PR #441 — the "no new machinery" group (43 of the 100) |
| 09 | #302 | **27** | 16 | 56 | PR #442 — the group was 44; Fumigate was already on `main` from #382 |
| 10 | #303 | **35** | 3 | 62 | PR #443 — the "no new machinery" group (38 of the 100) |
| 11 | #304 | **26** | 8 | 62 | PR #444 — the group was 38; 4 of it (Bane of Progress, Cleansing Nova, Vorinclex, Kambal) were already on `main` |
| 12 | #305 | **28** | 6 | 64 | PR #445 — the group was 36; 2 of it (Grave Titan, Reverberate) were already on `main` |
| 13 | #306 | **33** | 9 | 57 | PR #472 — the group was 43; In Garruk's Wake was already on `main` from #382 |
| 14 | #307 | **33** | 9 | 55 | PR #475 — the group was 45; Pawn of Ulamog, Crux of Fate and Mazirek were already on `main` |
| 15 | #308 | **27** | 8 | 65 | PR #476 — the "no new machinery" group (35 of the 100) |
| 16 | #309 | **34** | 4 | 62 | PR #477 — the "no new machinery" group (38 of the 100) |
| 17 | #310 | **29** | 7 | 62 | PR #479 — the group was 38; Korvold and Evacuation were already on `main` |
| 18 | #311 | **36** | 6 | 55 | PR #480 — the group was 45; Dragon Fodder, Solitude and Nevinyrral's Disk were already on `main` |
| 19 | #312 | **22** | 5 | 73 | PR #481 — the "no new machinery" group (27 of the 100) |
| 20 | #313 | **27** | 2 | 71 | PR #483 — the "no new machinery" group (29 of the 100) |
| 21 | #383 | **33** | — | — | PR #485 — the "no new machinery" group (33 of the 37 it listed) |
| 22 | #384 | **28** | — | — | PR #486 — the "no new machinery" group (28 of the 34 it listed) |
| 23 | #385 | **14** | — | — | PR #487 — the "no new machinery" group (14 of the 21 it listed) |
| 24 | #386 | **30** | — | — | PR #488 — the "no new machinery" group (30 of the 40 it listed) |
| 25 | #387 | **19** | — | — | PR #490 — the "no new machinery" group (19 of the 32 it listed) |
| 26 | #388 | **28** | — | — | PR #491 — the "no new machinery" group (28 of the 35 it listed) |
| 27 | #389 | **27** | — | — | PR #493 — the "no new machinery" group (27 of the 29 it listed) |
| 28 | #390 | **27** | — | — | PR #494 — the "no new machinery" group (27 of the 37 it listed) |
| 29 | #391 | **26** | — | — | PR #495 — the "no new machinery" group (26 of the 33 it listed) |
| 30 | #393 | **27** | — | — | PR #496 — the "no new machinery" group (27 of the 32 it listed) |
| 31 | #394 | **17** | — | — | PR #497 — the "no new machinery" group (17 of the 24 it listed) |
| 32 | #395 | **20** | — | — | PR #498 — the "no new machinery" group (20 of the 30 it listed) |
| 33 | #396 | **26** | — | — | PR #528 — the "no new machinery" group (26 of the 34 it listed) |
| 34 | #397 | **32** | — | — | PR #532 — the "no new machinery" group (32 of the 42 it listed) |
| 35 | #398 | **23** | — | — | PR #538 — the "no new machinery" group (23 of the 28 it listed) |
| 36 | #399 | **27** | — | — | PR #542 — the "no new machinery" group (27 of the 33 it listed) |
| 37–40 | #400–#403 | 0 | 0 | — | not started |
| 41–60 | #448–#467 | 0 | 0 | — | not started — ranked against `b5a3055`, 2026-09-13 |

### Where the work actually is — measured 2026-09-18

The Progress table above says what has been *registered*. It does not say
what is *available*, and the two have drifted apart because a batch only
ever harvests its "no new machinery" group.

**Read the status comment on a batch issue, not the issue body.** Every
one of the sixty issues carries a per-card `Status 2026-09-1x` comment
checked against `develop`, and that comment — not the body — is the
current triage. The bodies were written against the frozen batch-01
detector set and still file cards under the five mechanics above; the
status comments already correct for that, list each card's *other*
blockers in parentheses, and carry a refutation checklist. A re-sort of
the bodies alone overcounts, because dominant-blocker filing hides
second blockers.

**What the first six batches actually measured.** Batches 37–43 were
worked on 2026-09-18 against today's `develop`:

| Batch | Scoped as writable | Registered | Rate |
|---|---:|---:|---:|
| 38 | 62 | 40 | 65% |
| 39 | 62 | 31 | 50% |
| 40 | 59 | 33 | 56% |
| 41 | 71 | 46 | 65% |
| 42 | ~40 | 29 | ~72% |
| 43 | 66 | 47 | 71% |
| **total** | **~360** | **226** | **~63%** |

**So roughly 63% of a "ready" card actually gets written, not the
umbrella's 86%.** Every miss but two was a *hidden second blocker*
rather than a mechanic anyone thought was missing — cards whose real
blocker is an unimplemented keyword (level up, prowess, ninjutsu,
persist, cycling, offspring) are the worst case, because a keyword's
absence never surfaces in a dominant-blocker filing at all.

Two cards went the other way and shipped *better* than filed: Bygone
Colossus (warp is implemented end to end, #324) and Decoction Module
(energy is a real player counter already).

Plan against ~60%, and expect the skip list to be as useful an output as
the cards — it is what tells you which primitive to build next. Batch 43
named three mechanics the detector set has no blocker tag for at all
(`floating-triggers`, `constrained-mana-production`,
`conditional-alternative-costs`), plus `attack-requirements`: CR 508.1d
"attacks each combat if able" has no home, because `game.Restriction`
carries prohibition bits only. Goblin Rabblemaster is a skip rather than
a caveat for that reason — dropping a requirement removes a DRAWBACK,
which is the #259 direction.

**A batch worked after a mechanic lands is worth more than the same batch
worked before it.** #937 (modal and multi-target clauses) merged while
batch 43 was in flight and converted three planned skips into the
catalog's first modal triggered ability, first multi-predicate trigger
and first modal activated ability. Other batches holding cards filed
under modal / multi-target are worth a re-look now.

<!-- BEGIN GENERATED CATALOG CENSUS — regenerate with: go test ./internal/cards/coverage/ -update -->

**The catalog, as measured on this commit.** Not typed by hand and not
derived from what a batch believes it landed — `TakeCensus` in
`server/internal/cards/coverage` reads `effects.All()`, and
`TestRoadmapCensusIsCurrent` fails the build when this block and the
registry disagree.

| Measured | Count |
|---|---:|
| Registry keys (`len(effects.All())`) | **2039** |
| — whole cards (bare `oracle_id`) | **1973** |
| — back faces (`<oracle_id>#1`) | 66 |
| Declared `full` | 1560 |
| Declared `caveats` | 408 |
| Declared `unreviewed` | 71 |

A back face is usually half a card: the modal-DFC land cycle registers
only its sixty land backs, and those cards are still gap cards on their
batch issues. The exception is the Sieges, whose fronts are registered
too — a Siege is one whole card spread over two keys. Whole cards is
still the number to quote when someone asks how many cards the engine
automates; it undercounts by the number of Sieges.

<!-- END GENERATED CATALOG CENSUS -->

Everything from here to the end of this section is **history**: each
figure was true of the commit it names and is deliberately left alone.
`-update` rewrites the generated block above and nothing else, so a
regeneration can never quietly retcon a past measurement into
today's.

**Catalog: 438 → 504**, measured with `len(effects.All())` minus the
flicker probe — 438 on `origin/main` at `e5fc440`, 504 with #361 and
#362 merged together on a scratch branch, where the full server suite
passes. (The test binary reports 439 and 505; the probe is explained
below.) Re-probed on `origin/main` at **`459dea6`** for the batch 21–40
ranking: **504** again, both PRs now merged.

**Then 610 → 690** with batches 05–07: 610 on `origin/main` at
`31aba35` after the S23–S28 sprint work and the catalog publishing
(#412) landed, 690 with #437, #438 and #439 merged together on a
scratch branch where the full suite passes and no oracle ID registers
twice. (Test binary: 611 and 691.)

**Then 690 → 844** with batches 08–12, each measured on `origin/main`
after the one before it merged: 724 (#441), 751 (#442), 786 (#443),
812 (#444), 840 (#445). Re-probed on `origin/main` at **`b5a3055`**:
**844** shipped specs, the test binary reporting 845 with the flicker
probe.

**844 specs is not 844 cards.** Sixty of them are MDFC *back faces*,
registered under the composite key `<oracle_id>#1` that
`game.CatalogKey` produces for face 1 — the modal-DFC land cycle, whose
front faces are deliberately **not** registered
(`TestBackFaceSpecsAreKeyedByFace` pins exactly that). So the catalog
holds **784 whole cards plus 60 land backs**, and the 60 half-covered
MDFCs are still gap cards: Ondu Inversion // Ondu Skyruins sits in
batch 20 at rank 2203 even though Ondu Skyruins is registered. The gap
pass keys on the bare `oracle_id`, so it counts them the same way.

Quoting the spec count as a card count overstates coverage by the full
sixty — about 7% at this size, and ~13% back when the catalog held 504.
The two numbers have to be carried together or the smaller one gets
dropped.

**How complete those 844 are, measured the same way** (`Spec.Completeness`,
flicker probe excluded). On `origin/main` at `b5a3055`: **323 full, 187
with declared caveats, 334 unreviewed**. After the audit in this branch:
**535 full, 248 with declared caveats, 61 unreviewed**.

`CompletenessUnreviewed` is the zero value and means nobody has audited
the card against its printed text — it is a third state, not a synonym
for "works", and anything publishing the catalog has to show it as one
(see `server/internal/cards/effects/completeness.go`). The 61 that
remain are honestly unreviewed, not assumed good; they are the residue
of a pass that ran out of time, and the next pass should start there.

Two things that audit turned up are worth carrying forward:

- **Every one of the 187 existing caveats was re-checked, and only one
  described a gap that had closed** — Darksteel Citadel's, and only
  partly. That is the #412 sweep working: the 24-of-87 stale rate it
  found has not come back.
- **Indestructible is half enforced, and that is a live bug rather than
  a declared simplification.** S25 (#77) gated `DestroyPermanentForEffect`
  and the damage state-based actions, but not the MASS destroy path
  (`game.DestroyPermanentsForEffect`), so every board wipe in the
  catalog kills indestructible permanents. Twenty-four specs now declare
  it; grep the catalog for `board wipe ("destroy all")` and delete them
  with the fix. The note lives on `DestroyAllMatching` in
  `cards/effects/mass.go`.
- **Caveats can be wrong without being stale.** Krenko, Tin Street
  Kingpin's said "you get no Goblins"; a probe test showed the engine
  makes two, one MORE than last-known information would. A caveat that
  understates a gap in the stronger-than-printed direction is the worst
  kind, and no count-based check can find it.

Always MEASURE this line, never derive it. The batch-01 entry derived
"319" by adding up what it believed had landed when the registry
actually held 320, and the count has since been moved by three
concurrent work streams that no single batch author could see. That
warning is now enforced rather than merely written down: the generated
block above is produced by `TakeCensus` in
`server/internal/cards/coverage`, and `go test ./internal/cards/coverage/`
fails the build whenever this document and `effects.All()` disagree.
It also reports whole cards and back faces as separate numbers,
because conflating them is the second way this section went wrong.

Batch 01 moved play-rate coverage by **+2 in the top 100** (Dark
Ritual, Arcane Denial), **+17 in the top 200**, **+29 in the top 300**;
batch 02 adds **+18 more in the top 300** and **+20 between rank 301
and 360**.

### Batch 01 — what shipped

26 of the 28 "no new machinery" cards, in rank order: Arcane Denial,
Mana Drain, Victimize, Spectator Seating, Vault of Champions, Farewell,
Undergrowth Stadium, Dreamroot Cascade, Tireless Provisioner,
Stormcarved Coast, Spire Garden, Buried Ruin, Impact Tremors,
Phyrexian Tower, Storm-Kiln Artist, Stroke of Midnight, Rapid
Hybridization, Professional Face-Breaker, Bountiful Promenade,
Fyndhorn Elves, Ornithopter of Paradise, Karplusan Forest, Terminate,
Rockfall Vale, Mirkwood Bats, Brushland. That completes three land
cycles — all ten bond lands, all ten painlands, six of the ten
slowlands.

Three more from the blocked groups, because the blockers moved:

- **Dark Ritual** (rank 33, mana pipeline). A spell that adds mana had
  no way to reach a pool — every mana in the engine came from
  `ActivateManaAbility`. `AddManaForEffect` on `*Game` and the
  `AddMana` primitive (`add_mana.go`) are the whole change; Mana Drain's
  refund rides the same helper.
- **Return of the Wildspeaker** and **Boros Charm** (until EOT). #314
  landed `BoostUntilEOT` / `GrantKeywordUntilEOT` after the triage was
  written, and both cards are those two primitives on modes.

Two engine seams grew to make the ready cards honest:

- **`DelayedTrigger.ControllerTurnOnly`** — "at the beginning of
  **your** next main phase" (Mana Drain) as opposed to "the next
  turn's upkeep" (Arcane Denial). Without it the refund fired on the
  next opponent's main phase and emptied unused.
- **`triggerAlreadyPendingFrom`** — the "whenever **one or more**
  creatures you control deal combat damage" dedup (Professional
  Face-Breaker). The engine emits one damage event per creature; the
  helper declines the second while the first trigger is still queued.
  Without it the card ships stronger than printed, which is the #259
  direction.

### Batch 01 — simplifications, all weaker than printed

- **Arcane Denial** — "may draw up to two" is "draws two", and both
  draws ride one delayed trigger (two simultaneous triggers under one
  controller would pop a meaningless ordering prompt every upkeep).
- **Mana Drain** — fires at your next **precombat** main phase; a
  Drain cast in your own precombat main waits a turn instead of paying
  out in that turn's postcombat main.
- **Victimize** — the sacrifice is an additional cost to cast, not a
  resolution-time action (no sacrifice-then-continue prompt exists).
  The engine validates targets before it pays the cost, so the
  sacrificed creature can never be one of the two targets — same as
  printed. The returned creatures enter untapped and are tapped a beat
  later.
- **Tireless Provisioner** — always a Treasure, never a Food (no
  resolution-time option prompt for a trigger).
- **Storm-Kiln Artist** — "cast or **copy**": no spell-copy event
  exists, so copies make no Treasure.
- **Boros Charm** — the indestructible mode is inert (nothing in the
  destroy path reads the keyword); the label says so.

### Batch 01 — skipped from the "ready" group, with the missing seam

The mechanical triage was optimistic about two:

- **Black Market Connections** — needs a beginning-of-main-phase
  trigger event (only upkeep and end step exist) **and** a
  resolution-time "choose one or more" prompt for a trigger (the modal
  machinery is cast-time only).
- **Growth Spiral** — "you may put a land card from your hand onto the
  battlefield" needs a pick-from-hand prompt with a continuation and a
  hand-to-battlefield move; neither exists. Shipping the draw alone
  would be a misleading cantrip, the Chromatic Lantern reasoning.

And from the until-EOT group, still blocked despite #314: Heroic
Intervention (hexproof and indestructible are both inert, so the whole
card would be a no-op), Rogue's Passage ("can't be blocked" is a combat
restriction), Toxic Deluge ("pay X life" as an additional cost has no
shape), Akroma's Will (the "choose both if you control a commander"
clause plus two inert keywords), Teferi's Protection and The One Ring
(the protection family, #95).

### Two engine bugs batch 01 found, both fixed with a regression test

- **Impulse exile took the bottom card.** `ExileTopWithPermissionForEffect`
  read `Library.Cards[0]`; the library's top is the **last** element
  (`PopTop`, and so every draw and mill). Every earlier test seeded a
  one-card library, where the two coincide — Ragavan and Breeches had
  been exiling the wrong end of the library since S21.
- **A token entering never invalidated the layer cache.**
  `CreateTokenForEffect` emits no `EventZoneMove`, which was the only
  entry event the layer listener watched, so a Goblin made under
  Glorious Anthem stayed 1/1 until something unrelated forced a
  recompute, and tokens never got a CR 613 timestamp. The listener now
  treats `EventTokenCreated` as an entry.

### Batch 02 — what shipped, in three streams

38 cards of the 100 at `edhrec_rank` 238–360, landed by **three work
streams, two of which were the same issue worked in parallel without
either session knowing**. Recording that here because the near-miss is
the reusable lesson, not the trivia:

- **First pass — PR #351, 25 cards.** The "no new machinery" group.
- **Second pass — this branch, 11 cards.** What the first pass did not
  reach: Ketria Triome, Darksteel Citadel, Entomb, Buried Alive,
  Seething Song, Mana Geyser, Harrow, Diabolic Intent, Gray Merchant of
  Asphodel, Craterhoof Behemoth, Decanter of Endless Water. Nine of
  them because the triage had filed them under a blocker that had
  already moved; two (Entomb, Buried Alive) because they needed a
  one-case engine fix that ships with them.
- **Third stream — #356, 2 cards.** The mana-pipeline work picked up
  Boros Signet and Mox Opal on its way past. Nobody planned that as
  part of #295.

**Six oracle IDs were written twice into differently-named files** —
`original_duals.go` / `original_dual_lands.go` and `bounce_lands.go` /
`karoo_lands.go`. Git reports **no conflict** when the filenames
differ, and `Register()` panics on a duplicate oracle ID, so the
collision would have been a hard boot failure discoverable only by
running the tests. The second pass deleted its own copies and took the
merged ones.

The one place this did NOT happen is `slowlands.go`, because both
sessions appended rows to the existing cycle table instead of creating
a second file. **That is the rule the near-miss argues for: a card that
belongs to an existing cycle goes in that cycle's table, never in a new
file of its own.** The cycle ended at ten rows with no duplicates.

**Four land cycles closed or extended.** The last four slowlands
(Sundown Pass, Haunted Ridge, Overgrown Farmland, Deathcap Glade)
finish that cycle at ten; four original duals (Underground Sea,
Volcanic Island, Tropical Island, Tundra), two tri-lands (Arcane
Sanctum, Jungle Shrine), two bounce lands (Simic Growth Chamber,
Golgari Rot Farm), plus Ketria Triome, Seat of the Synod, Darksteel
Citadel and Scavenger Grounds.

**Spells and permanents.** Infernal Grasp, Withering Torment, Baleful
Strix, Entomb, Buried Alive, Damn, Snap, Seething Song, Mana Geyser,
Harrow, Diabolic Intent, Gray Merchant of Asphodel, Archmage Emeritus,
Syr Konrad the Grim, Lotus Cobra, Guardian Project, Loran of the Third
Path, Avenger of Zendikar, Craterhoof Behemoth, Decanter of Endless
Water.

**Ten the #295 triage filed as blocked (or as would-ship-stronger) were
writable after all** — five because the blocker moved in the days
before the batch, five because they were mis-filed. Nine of the ten are
the second pass's whole contribution; only **Damn** was also caught by
#351. A triage written days before a batch is worked is the thing to
re-check first, and this is how much it was worth here:

- **Decanter of Endless Water** — filed "player / game-rule statics".
  `Spec.NoMaxHandSize` landed with Thought Vessel (#338).
- **Craterhoof Behemoth** — filed "cost modification, until EOT". The
  until-EOT half is S32's `BoostUntilEOT` / `GrantKeywordUntilEOT`;
  the cost half was never real.
- **Damn** — filed "would ship stronger than printed". Overload is a
  real alternative cost now, and the "can't be regenerated" clause is
  vacuous in an engine with no regeneration, so nothing is missing.
- **Seething Song** and **Mana Geyser** — filed "mana pipeline".
  `AddMana` from batch 01 is the whole requirement; Mana Geyser's
  derived count was always ordinary Go.
- **Harrow** and **Diabolic Intent** — filed "cost modification". Both
  are CR 601.2f *additional* costs, and `SacrificeCost` has existed
  since S21.
- **Gray Merchant of Asphodel** — filed "cost modification". Devotion
  only reads mana costs; it modifies nothing.
- **Ketria Triome** — filed "other-zone casting" for its cycling.
  Raffine's Tower already ships without cycling as a declared
  simplification; the same call applies.
- **Darksteel Citadel** — filed "protection / prevention". Two of its
  three lines are live; indestructible is declared in
  `PrintedKeywords` and inert until #176, which is the weaker
  direction and allowed. *(Update 2026-09-16: indestructible has been
  enforced since #380, so all three lines are live.)*

### Batch 02 — the 19 declared skips, each with its missing seam

- **A card choice inside a replacement effect** (7): Mox Diamond, and
  the six reveal-lands — Choked Estuary, Foreboding Ruins, Port Town,
  Fortified Village, Frostboil Snarl, Furycalm Snarl. All are "as this
  enters, you may reveal / discard <a card you pick>", and the
  replacement pipeline is synchronous with no per-card prompt.
  `ReplacementEffect.Optional` is a bare yes/no and carries a hazard
  besides — see below.
- **`ManaAbilityCost` had no mana component** (5): the filter lands —
  Cascade Bluffs, Flooded Grove, Fetid Heath, Twilight Mire, Rugged
  Prairie. The triage called these ready; they were blocked on exactly
  what Boros Signet and Skycloud Expanse were blocked on.

  **This skip was already stale when it was written.** #356 (the mana
  pipeline) landed `ManaAbilityCost.Mana` while this branch was in
  flight — `signets.go` uses it for Boros Signet's "{1}, {T}: Add
  {R}{W}" — so the stated blocker is gone. What remains unverified for
  the filter lands is only the three-way OUTPUT choice ("Add {U}{U},
  {U}{R}, or {R}{R}"), which is a different question from the cost.
  **These five are the obvious next pickup**, and they are deliberately
  not taken here: the mana pipeline is another session's lane and this
  batch has already collided once. Recording the staleness rather than
  leaving it to rot is the #350 lesson.
- **No replacement on *triggering*** (1): Panharmonicon. There is no
  `RepEvent` for an ability triggering.
- **No replacement on *token creation*** (1): Academy Manufactor. Same
  shape, different event that does not exist.
- **A per-turn tally** (1): Morbid Opportunist's "this ability triggers
  only once each turn". Shipping without the limiter would be stronger
  than printed, so it is skipped rather than shipped.
- **Leave-the-battlefield counter LKI** (1): The Ozolith needs the
  counters a permanent *had* as it left, and CR 400.7 has already
  cleared them by the time a watcher sees the event.
- **No untap-step trigger** (1): Seedborn Muse. `EventStepTransition`
  is deliberately not in the public event log, so only a replacement
  can see it — and a replacement that does not cancel re-fires until
  the 32-iteration cap. *(Closed by #74: the answer was that it is
  not a trigger at all — see `game/untap.go`.)*
- **An opponent-paid *choice* that is not a mana cost** (1): Braids,
  Arisen Nightmare. `PayUnless` prices a decision in mana; "each
  opponent may sacrifice a permanent sharing a card type" has no
  shape.
- **A static that applies from the graveyard** (1): Anger. `Spec.Static`
  applies only while the card is on the battlefield (CR 113.6 default).

### Batches 03 and 04 — one pass each, in parallel

Both were written by forked agents in isolated worktrees at the same
time as each other, with the batch 02 collision fresh: no engine edits,
every batch-specific package-level name prefixed `b03` / `b04`, cycle
tables edited by adding rows only, and cross-batch oracle IDs checked
against `main` before pushing. The two branches merge cleanly and pass
the suite together; the cost of the discipline is a few oddly named
files (`b03_tri_lands.go`, `b03_bounce_lands.go`, `b03_original_duals.go`,
and batch 04's four one-card land files) that want folding into the
cycle tables once both are on `main`.

**Batch 03 — 37 of 48 (PR #361).** Three cycle tables: six tri-lands
(Crumbling Necropolis, Nomad Outpost, Mystic Monastery, Opulent Palace,
Frontier Bivouac, Seaside Citadel), the six Alpha duals #351 did not
take (Badlands, Scrubland, Bayou, Taiga, Plateau, Savannah), four
bounce lands (Dimir Aqueduct, Orzhov Basilica, Izzet Boilerworks,
Gruul Turf). Twenty-one singles: Land Tax, Shamanic Revelation, Go for
the Throat, Rampaging Baloths, Great Furnace, Demolition Field,
Guttersnipe, Warren Soultrader, High Market, Red Elemental Blast,
Geier Reach Sanitarium, Sram, Entish Restoration, Urza's Cave,
Dispatch, Expedition Map, Sheoldred the Apocalypse, Living Death,
Mental Misstep, Pyroblast, Fabricate.

Declared weaker: Entish Restoration's sacrifice as an additional cost
(Victimize's posture); Mental Misstep's Phyrexian pip paid with {U}
only; Warren Soultrader's "another" enforced by name; Geier Reach's
loot seat by seat; the bounce lands inherit the Chancery's
choice-as-target. One stronger corner, declared: Land Tax's
intervening-if is checked at trigger time only — the same posture
every intervening-if card in the catalog takes.

Skipped (11), one seam each: Anointed Procession and Parallel Lives (no
token replacement-event kind — the Academy Manufactor gap); Game Trail
and Shineshadow Snarl (reveal-from-hand entry choice); Welcoming
Vampire (once-per-turn tally — the Morbid Opportunist gap); Walking
Ballista (no counter-removal cost, no enters-with-X for a permanent
spell); Treasure Vault (no X on an activated ability); Simian Spirit
Guide (no ability activatable from hand); Forgotten Ancient (the
upkeep counter-move is a promptless choice and the cast half alone is
a fraction); Psychosis Crawler (a hand-size CDA needs layer
invalidation on draw and discard, or its toughness goes stale in the
stronger direction); Ohran Frostfang (an "attacking creatures" static
needs a layer bump on `EventAttack` and `ClearCombat`).

**Batch 04 — 29 of 38 (PR #362).** Talisman of Curiosity and Talisman
of Resilience as rows; Sandsteppe Citadel; three bounce lands (Boros
Garrison, Rakdos Carnarium, Selesnya Sanctuary); Ancient Den; Graven
Cairns, the first filter land, writable since #356 gave
`ManaAbilityCost` a mana component; and twenty-one singles: Tatyova,
Accursed Marauder, Chandra's Ignition, Diabolic Tutor, Bedevil, Field
of the Dead, Soul Warden, Adeline, Rise of the Dark Realms, Sanguine
Bond, Avacyn's Pilgrim, Exquisite Blood, Terror of the Peaks, Chord of
Calling, Gitaxian Probe, Strip Mine, Hedron Archive, Cathars' Crusade,
Elemental Bond, Beastmaster Ascension, Wheel of Fortune.

Declared weaker: Terror of the Peaks' targeting tax is absent (#93);
Gitaxian Probe's {U/P} is charged as {U} (no payment path reads the
Phyrexian flag); Adeline's tokens always attack the player (no
planeswalker-attack path); the bounce lands inherit the Chancery's
choice-as-target.

Skipped (9): Necroblossom Snarl and Vineglimmer Snarl (reveal-from-hand
entry choice — eight cards across three batches now wait on this one
seam); Relic of Legends and Springleaf Drum ("tap an untapped creature
you control" as a mana-ability cost component); Emergence Zone
(per-player "cast as though it had flash" permission — the Teferi
gap); Unwinding Clock (untap-step trigger event — the Seedborn Muse
gap); Ripples of Undeath (beginning-of-main-phase trigger event plus a
pay-then-pick continuation — the Black Market Connections gap);
Nesting Grounds (per-slot target clauses on one ability plus a
counter-kind pick); Maskwood Nexus (changeling on the characteristic).

**The seams, ranked by cards they would unlock across batches 01–04:**
reveal-from-hand entry choice (8), token replacement-event kind (3),
once-per-turn trigger tally (2), untap-step trigger event (2),
main-phase trigger event (2), tap-another-permanent cost component
(2), per-player cast permission (1), X on an activated ability (1),
counter-removal cost (1), ability from hand (1), layer invalidation on
hand size or combat state (2).

### Batches 05, 06 and 07 — stopped, rebased, finished

Started as three parallel forks on the same day as 03 and 04, stopped
mid-way for usage, and finished by three fresh agents working from a
written brief (`data/roadmap/BRIEF.md`, local only) against a `main`
that had moved seventy commits in between — equipment and auras
(#379), indestructible (#380), the boardwipe primitives (#382),
proliferate (#381), flashback (#411), spell copies (#419), the legend
rule (#418). The re-triage against those seams is why each batch has a
card the original triage called blocked. The fresh-agent pass cost
roughly half a fork per batch, which is the cheaper shape to keep.

**Batch 05 — 23 of 35 (PR #437).** Talisman of Impulse and Savage
Lands as rows; twenty-one singles including Ruinous Ultimatum (on
#382's `DestroyAllMatching`, one simultaneous event), Scrawling
Crawler, and **Dualcaster Mage** — a targeted ETB trigger over #419's
`CopySpell`, the only re-triage unlock. Skipped (12), one seam each:
Greater Good and Altar of Dementia (the sacrificed creature's LKI is
not on the activated item); Reprieve (no counter-to-hand surface, and
counter-then-return would fire Syr Konrad — stronger); Strionic
Resonator (ability copies are outside `CopySpellForEffect`); Imp's
Mischief (no stack-item retarget); Teferi's Ageless Insight (draw
replacement has no count); Archdruid's Charm (per-slot mode targets);
Ancient Copper Dragon (no RNG surface for effects); Torment of
Hailfire (opponent's non-mana choice); Cryptolith Rite (mana ability
granted through a static); Borne Upon a Wind (flash permission);
Bender's Waterskin (untap-step event).

**Batch 06 — 29 of 35 (PR #439).** Twenty-seven from the first pass —
which had left them without tests or completeness declarations, and
with Riveteers Overlook searching even when bounced in response
(stronger than printed; fixed) — plus Springbloom Druid and Bloodchief
Ascension. Declared weaker: Into the Flood Maw has no gift; Riveteers
Overlook is one trigger rather than a reflexive one; Spelunking lacks
its land-from-hand rider; Starting Town rounds; Warstorm Surge has no
cross-card LKI; Springbloom Druid's land is chosen at trigger time;
Whip of Erebos's exile clause misses a bounce (see the engine note).
Skipped (6): Howling Mine (no draw-step trigger event); Archaeomancer's
Map (no put-land-from-hand prompt — the ETB alone is half the card);
Narset's Reversal (no counter-to-zone surface); Elvish Spirit Guide and
Reassembling Skeleton (no ability from a non-battlefield zone);
Peregrin Took (token-creation replacement event).

**Batch 07 — 28 of 40 (PR #438).** Twenty-five from the first pass,
plus Ghost Quarter, Field of Ruin, Tempt with Discovery (the first
tempting offer — a sequential search continuation chain) and **Kodama
of the West Tree**, whole now that "modified" can read Equipment and
Auras through #379's `AttachedTo`. Declared weaker: Sai's two-artifact
draw omitted; Myr Retriever's "another" checked at resolution;
Splendid Reclamation taps a beat after entry; Ayara reads stamped token
colours; Tempt counts only the opponents who took a land. Skipped (12):
Terrasymbiosis and Tocasia's Welcome (once-per-turn tally); Wonder
(statics gathered from the battlefield only); Tainted Pact (iterated
yes/no at resolution); Bloodletter of Aclazotz (life-loss replacement —
see the engine note); High Fae Trickster (per-player flash); Rishkar
(mana ability granted through a static); Scrap Trawler (trigger target
parameterised by the event); Thousand-Year Elixir (ability-only
haste); Burgeoning (no land-play event); Reconnaissance (no
remove-from-combat primitive); Forsaken Monument (triggered mana
ability; `EventManaAdded` carries no colour).

**The seams, re-ranked across batches 01–07:** reveal-from-hand or
put-from-hand entry / resolution choice (11 — the eight reveal-lands
plus Archaeomancer's Map, Burgeoning, Spelunking's rider), once-per-turn
trigger tally (4), token-creation replacement event (4), abilities
activated from a non-battlefield zone (4 — Simian and Elvish Spirit
Guide, Reassembling Skeleton, Anger-style statics aside), untap-step
trigger event (3), per-player flash permission (3), mana ability
granted through a static (2), main-phase trigger event (2),
tap-another-creature cost (2), sacrificed-cost LKI on the item (2),
counter-to-zone surface (2), draw-step event (1). Everything else is a
single card.

### Batches 08–20 — the first pass over the first 2000 is complete

Thirteen batches, one fresh agent each, one PR each (#441–#445, #472,
#475–#477, #479–#481, #483), run back to back over 2026-09-12/13 from
the written brief in `data/roadmap/BRIEF.md` (local only). Every
batch's "no new machinery" group was worked once; what each shipped,
caveated and skipped is on its issue and in its PR body. The
first-2000 tally: **every batch's ready group has been walked**, and
what remains of each batch is the cards the original triage filed
under a named mechanic.

The skips from those thirteen batches cluster, and the clusters are
the build order for the mechanics that would unlock the most
already-triaged cards (counts are approximate, from the issue
comments; a card can sit in two):

- **Modal triggers and richer mode clauses** (~12): a triggered
  ability with "choose one" (Junji, Atsushi, Glissa Sunslayer, Rankle,
  Aether Channeler), per-mode target slots on "choose two or more"
  (Cryptic Command, Casualties of War, Prismari Command), per-slot
  predicates on one mode (Bushwhack, Ram Through, Archdruid's Charm),
  repeatable modes (Fiery Confluence).
- **Activated-ability cost components that don't exist** (~11):
  remove a counter (Devoted Druid, Walking Ballista), discard a card
  (Fauna Shaman, Tortured Existence, Trading Post's fourth ability),
  tap another creature (Springleaf Drum, Relic of Legends, Susur
  Secundi, Uthros Research Craft, Clock of Omens), return a land
  (Quirion Ranger).
- **Trigger and token replacement** (~9): "triggers an additional
  time" (Panharmonicon, Teysa Karlov, Isshin, Echoes of Eternity,
  Annie Joins Up, Elesh Norn) and "create twice that many" (Xorn,
  Divine Visitation, Stridehangar Automaton, Peregrin Took).
- **Put or pick a card from hand at resolution** (~8): Kodama of the
  East Tree, Ghalta, Cultivator Colossus, Sneak Attack, Terrain
  Generator, Arboreal Grazer, Stoneforge Mystic's second ability,
  Eureka Moment's and Broken Bond's riders.
- **Step-entry trigger events that don't exist** (~7): ~~untap step
  (Seedborn Muse, Unwinding Clock, Bender's Waterskin,
  Drumbellower)~~ — **closed by #74**, and not as a trigger event:
  the untap step's turn-based action now takes contributions
  (`Spec.UntapStep`), and the per-permanent `EventUntapCard` is
  emitted from the step so Mesmeric Orb's "whenever a permanent
  becomes untapped" is writable. Quest for Renewal dropped its
  untap caveat with it. Remaining: draw step (Howling Mine, Kami of
  the Crescent Moon), main phase (Ripples of Undeath, Black Market
  Connections).
- **Per-player "cast as though it had flash"** (~7): Emergence Zone,
  Alchemist's Refuge, Vedalken Orrery, Shimmer Myr, Borne Upon a Wind,
  High Fae Trickster, Liberator.
- **Once-per-turn trigger tally** (5): Morbid Opportunist, Welcoming
  Vampire, Terrasymbiosis, Tocasia's Welcome, Monument to Endurance.
- **An opponent's non-mana choice at resolution** (5): Torment of
  Hailfire, Combustible Gearhulk, Charismatic Conqueror, Painful
  Quandary, Chain of Vapor.
- **A mana ability granted through a static** (5): Cryptolith Rite,
  Jaheira, Insidious Roots, Rishkar, Marvin.
- **Life-change replacement through the effect path** (#482, 4) and
  **mana-production replacement** (2: Nyxbloom Ancient, Mana
  Reflection).
- Singles worth naming: attack- and block-count restrictions (Silent
  Arbiter, Crawlspace), a die roll over the seeded RNG (Ancient
  Copper Dragon, Ancient Gold Dragon), a mill replacement (Bruvac), a
  search replacement (Aven Mindcensor).

**X on an activated ability is closed.** `game.AbilityCost` reads an
`{X}` out of its mana component and carries a `MinX` floor for "X
can't be 0"; the value is announced at CR 602.2b, locked onto the
stack item, and read back through `ctx.X()` exactly as a spell's is.
Treasure Vault, Helm of Obedience and Soothsaying ship with it. Blast
Zone is now a counter question rather than an X one — its "{X}{X},
{T}: Put X charge counters" is writable today, and what is left is
the third ability's "destroy each nonland permanent with mana value
equal to the number of charge counters".

Two neighbouring seams stayed open, and it is worth saying which:

- **A variable-count sacrifice cost.** Ruthless Technomancer's
  "Sacrifice X artifacts" is an X that is not in the mana component;
  `AbilityCost.SacrificeOther` still names exactly one permanent.
- **A mill replacement.** Bruvac still cannot be written, but the
  reason moved: `MillToZoneForEffect` routes every card through the
  CR 614 pipeline (so Leyline of the Void and the CR 903.9 commander
  redirect both apply to a mill today). What is missing is a
  replacement event kind for the mill AMOUNT — "if a player would
  mill one or more cards, they mill twice that many instead" is not a
  per-card zone move.

Three engine bugs the sweep found were filed rather than fixed on the
spot: **#446** (mass destroy bypasses indestructible), **#478** (a
fetched permanent whose entry queues a replacement prompt is stranded in
the library; **fixed**, S39), **#482** (effect-side life changes skip the
life-replacement pipeline; Rhox Faithmender is wrong today). The agents also recorded, on their issues,
the smaller gaps they worked around: attacking taps without a tap
event, dies-trigger LKI without counter math, entry replacements that
cannot read X, a printed 0/0 the toughness SBA never sweeps, tokens
skipping the entry pipeline, and `EventBlock` firing per blocker.

### What batches 06 and 07 found in the engine

- **FIXED (#482) — the effect-side life change skipped replacement
  effects.** `ChangePlayerLifeForEffect`, which every catalog drain and
  lifegain goes through, did not run the `RepEventLife` pipeline that
  the public `ChangePlayerLife` does, and neither did the lifelink
  credit — so Rhox Faithmender doubled a hand-typed life total and
  nothing else. Every writer now runs the CR 614 window and lands in
  one tail (`game/life_tail.go`), the CR 616 pause included.
  Bloodletter of Aclazotz, Alhammarret's Archive and Angel of Vitality
  can be written. `DealDamageToPlayerForEffect` still emits no
  life-loss `EventChangeLife` of its own: damage reduces life directly
  (CR 120.3) and the aristocrats cards read `EventDealDamage` for it.
- **OPEN — a bounce skips the leaves-the-battlefield replacements.**
  `BounceToHandForEffect` bypasses the CR 614 pipeline that destroy,
  sacrifice and the SBA exits go through, so a "if it would leave the
  battlefield, exile it instead" replacement never sees a bounce.
  Scenario: reanimate a creature with Whip of Erebos, bounce it in
  response to the end-step trigger — it goes to hand, not exile.
  Declared as a Whip caveat.

### What batch 04 found in the engine

- **OPEN — `triggerAlreadyPendingFrom` is not enough for attack
  triggers.** The batch 01 helper checks only `PendingTriggers`, which
  is exactly right for combat damage (every creature's damage event
  fires inside one mutation). `DeclareAttacker` is different: it runs
  the state checks after each single declaration, which drains the
  queue onto the stack, so a "whenever you attack" card written
  against the helper fires once per attacker — Adeline with two
  attackers made six Humans instead of three. Batch 04 uses a wider
  `b04TriggerPendingOrOnStack` that also scans `StackMeta`; it should
  become the shared helper when the batches are folded. Residual gap,
  weaker direction: an attacker declared after the trigger has already
  resolved fires it again.

### What batch 02 found in the engine

- **FIXED — a library search could not put a card into a graveyard.**
  `searchDestZoneLocked` handled hand, battlefield and library and
  returned `ErrZoneNotFound` for everything else, so Entomb and Buried
  Alive emitted an effect error and found nothing. The generic
  `MoveCard` branch already handled the destination correctly; only the
  zone lookup was missing. One case added, with tests on both cards.
- **OPEN — `ReplacementEffect.Optional` has no `entryResumable`
  guard.** `EntryLifeCost` checks `ev.entryResumable` before queuing
  its prompt (`entry_choice.go`), because only the land-play path can
  resume a paused battlefield entry. The `Optional` branch of the same
  apply-loop queues unconditionally and returns
  `errReplacementPending`. An `Optional` self-replacement on a
  permanent entering by any other route would therefore pause with no
  resume path. No catalog card can reach it today, which is why it is
  reported rather than patched — but it is the reason the six
  reveal-lands were skipped rather than written against `Optional`.
- **STALE COMMENT — `azorius_chancery.go`** still says a catalog
  replacement "cannot fire on its own source's entry at all". That was
  true when written and stopped being true with the Temple cycle, which
  added the entering card's own replacements as a third gathering
  block. Three land cycles depend on the behaviour it says is
  impossible. `karoo_lands.go` records the correction; the Chancery
  itself still taps via `OnETB` and could now use `SelfEntersTapped()`.

## The audited catalog count — 288 (at the time of the audit)

Not 289, and not the number `grep -c 'OracleID:'` reports. This is the
number the 2000-card gap below was computed against; the running
count is in the Progress section above.

Cards register from **tables inside `for` loops** — `temples.go`,
`check_lands.go`, `battle_lands.go`, `bond_lands.go`, `shocklands.go`,
`painlands.go`, `talismans.go`, the fetchland and Talisman cycles — so one
literal `Register(Spec{` site can be ten cards. Counting `Register` sites
is not counting cards; `len(effects.All())` is. That undercount has
produced two wrong numbers in the last week (191 against a real 201, then
225 against a real 261).

There is a second, subtler correction on top of it. `len(effects.All())`
measured **from a test** reports **289**, because `flicker_test.go`'s
`init()` registers a non-card probe spec keyed `test-flicker-etb-probe`
to count direct `OnETB` hook fires. It is test-binary-only and correct to
exist — it just is not a card. The shipped registry holds **288**, and
all 288 resolve to real, Commander-legal Scryfall oracle IDs. No tokens,
no helpers, no strays.

## Method — reproducible

### 1. The registry, honestly

A throwaway test inside the toolchain container, deleted before commit:

```go
// server/internal/cards/effects/zz_probe_test.go
package effects

import ("os"; "sort"; "testing")

func TestZZProbeCatalog(t *testing.T) {
	all := All()
	lines := make([]string, 0, len(all))
	for _, s := range all {
		lines = append(lines, s.OracleID+"\t"+s.Name)
	}
	sort.Strings(lines)
	f, _ := os.Create("/w/catalog_probe.txt")
	defer f.Close()
	for _, l := range lines {
		f.WriteString(l + "\n")
	}
	t.Logf("REGISTERED_COUNT=%d", len(all))
}
```

```bash
IMG=$(docker images --format '{{.Repository}}:{{.Tag}}' \
      | grep '^vsc-cmd_and_ctrl-' | grep -v features | head -1)
docker run --rm -v "$PWD":/w -w /w/server "$IMG" \
  bash -lc 'go test ./internal/cards/effects/ -run TestZZProbeCatalog -v'
# REGISTERED_COUNT=289   → 288 cards + the flicker probe
```

### 2. The gap, in one pass

One Python pass over the 629 MB dump, ~5 s. **Load it once** — the file is
a single-line JSON array, so decode it incrementally with
`json.JSONDecoder().raw_decode` rather than shelling out per card.

- **Skip placeholder printings.** `type_line == "Card // Card"`, and
  layouts `art_series` / `token` / `double_faced_token` (plus `emblem` /
  `vanguard` / `scheme` / `planar`). **6,456 printings rejected**, 2,712
  of them on the `"Card // Card"` type line alone. These mis-resolve real
  single-faced cards as multi-face; the trap has caused a wrong triage
  twice and it is the reason #278 carries a warning about it.
- Dedupe by `oracle_id`, preferring a `layout: normal` printing, carrying
  `edhrec_rank` and Commander legality across from any printing that has
  them. 117,738 printings → **34,936 distinct oracle IDs**.
- Keep `legalities.commander == "legal"` → 31,830. Drop basic lands →
  31,824. Drop the 288 already registered → **31,536 missing**, of which
  **31,474 carry an `edhrec_rank`**.
- Sort ascending by rank, take 2000, chunk into 20 batches of 100.

The top 2000 spans **rank 9 to rank 2,234** — near the top of the format
the catalog already owns most slots (74 of the top 100), so 2000 cards
only consumes ~2,234 rank positions.

**The second pass, for batches 21–40**, is the same pass with the same
filters, run against the 504-spec registry at `459dea6`: 117,738
printings → 34,936 distinct oracle IDs → 31,830 Commander-legal →
31,824 after basics → **31,380 unregistered**, of which **31,318 carry
an `edhrec_rank`**. Then the 2000 oracle IDs already assigned to
batches 01–20 are removed — by ID, not by rank cutoff, so a tie at the
boundary cannot land a card in two batches — leaving **29,472**. The
first 2000 of those are batches 21–40 and span **rank 2,235 to rank
4,253**: 2,019 rank positions for 2,000 cards, so out here the catalog
owns almost nothing and the batches run nearly one card per rank.

1,846 of the 2000 cards in batches 01–20 are still unregistered; those
stay on their original batch issues rather than being re-ranked. That
is deliberate — re-ranking would silently move cards between issues
that people are already working.

**The third pass, for batches 41–60**, is again the same pass with the
same filters, run against the registry at `b5a3055`. That probe
returned **845 entries**, which is *not* 845 cards. One is the flicker
probe, a synthetic spec registered from `flicker_test.go` and present
only in the test binary. Of the 844 that remain, **60 are MDFC back
faces** keyed `<oracle_id>#1` whose front faces are deliberately
unregistered (`TestBackFaceSpecsAreKeyedByFace` pins the key shape),
leaving **784 whole cards**. The flicker probe's key is not an oracle
ID, so it never matched a dump card and the gap arithmetic below is
unaffected either way. The counts: 117,738 printings → 34,936 distinct
oracle IDs → 31,830 Commander-legal → 31,824 after basics → **31,040
unregistered**, of which **30,978 carry an `edhrec_rank`**. Then the
4000 oracle IDs already assigned to batches 01–40 are removed — again
by ID, not by rank cutoff — leaving **27,432**. The first 2000 of those
are batches 41–60 and span **rank 4,254 to rank 6,289**: 2,036 rank
positions for 2,000 cards.

3,546 of the 4000 cards in batches 01–40 are still unregistered at
`b5a3055`, and they stay on their original batch issues for the same
reason. 27,432 ranked, Commander-legal, unregistered cards remain after
batch 60, so the dump is nowhere near exhausted — the roadmap stops at
6000 because that is what has been split into batches, not because the
list runs out.

### 3. The mechanic triage

Each of the 6000 is checked against the primitive set that actually exists
on `main` — the `effects.Spec` slots (`Targets`, `Modes`, `Static`,
`Replacements`, `Triggered`, `Activated`, `ManaAbilities`,
`AdditionalCost`, `AlternativeCosts`, `TapCost`, `PrintedKeywords`) and the
`*ForEffect` helpers on `*game.Game` — using regexes over oracle text plus
Scryfall's `keywords`, `layout` and `type_line`.

**This triage is mechanical, and that is its limit.** It is a starting
point good enough to schedule work against, not a verdict: a card marked
"no new machinery" can still turn out to need a seam that does not exist,
and the card-file author is the final arbiter. Where it errs it errs
toward optimism, because the detectors key off printed text and the engine
gaps that bite are usually the unprinted ones.

**Batches 21–40 and 41–60 ran the batch 01–20 detector set unchanged**,
on purpose: the three passes are one ranked list, and a "ready today"
count computed against a different detector set would not be comparable
across the boundaries. The cost is that the detectors go stale as
primitives land, and **as of 2026-09-18 five whole blocker groups are
over-counted in every batch issue, 01–60 alike**:

| Detector group | Shipped as | Cards it still counts as blocked |
|---|---|---:|
| Cost modification, alternative casts | #93 (S28), 2026-09-11 | 658 |
| Until-end-of-turn continuous effects | #279 / #314 (S32) | 540 |
| Attachments — Equipment and Auras | #280 → S33 (`game/attach.go`) | 332 |
| Mana pipeline | #352 (S32), `ManaAbilityCost.Mana` and friends | 234 |
| Card-type completeness | #92 (S27), `game/activated.go`, `game/loyalty_test.go` | 203 |

**1,967 cards across the 6000 sit in a blocked group whose blocker has
shipped.** That is more than the whole registered catalog, and it is
invisible to anyone reading a batch issue, because the issue still
files those cards under a mechanic that does not exist any more.

**Cost modification is the one to watch**, and it is the one an earlier
revision of this paragraph missed: it is the #1 blocker in all three
ranking tables, and it closed on 2026-09-11 — the same day batches
21–40 were ranked, two days before 41–60 were. It was stale before the
ink was dry on the issues that cite it.

Four more groups are PARTLY unblocked and need reading card by card:
protection (hexproof, shroud, ward, indestructible and damage
prevention shipped; protection itself is #662, copying is #665/#666),
deferred combat keywords (landwalk #705 and changeling shipped; infect
is #748, prowess is #706), multi-face (the face model and MDFC picker
shipped; the transform verb is #343, adventure is #719) and non-hand
casting (flashback and escape shipped; cycling #660, foretell #658,
suspend #659, madness #657 remain).

Every batch issue 01–60 now carries a comment with its own re-sorted
ready group — the issue's own card lists, re-filed against what is on
`develop`, not a fresh detection pass. The batch-02 lesson applies with
full force out here: re-check the triage before working a batch,
because the blockers move faster than the issues do.

## What the catalog covered at the audit

288 cards, by shape. This is the audit snapshot batches 01–20 were
computed against (batches 21–40 against 504, batches 41–60 against
844), kept for the shape it shows; the live count is in the Progress
section.

| Slice | Cards |
|---|---:|
| Lands | 83 |
| Creatures (bodies, ETB / dies payoffs) | 50 |
| Other (enchantments, artifacts, utility) | 49 |
| Ramp / mana rocks | 37 |
| Removal / interaction | 36 |
| Card draw / selection | 33 |

By play rate: **74 of the top 100**, 123 of the top 200, 143 of the top
300, 168 of the top 500, 196 of the top 1000, 219 of the top 2000.

Lands are over-represented on purpose. The land cycles are where one
helper buys ten cards — `SelfEntersTappedUnless` alone covers the
checklands, battle lands and bond lands — and #267/#268 finished the
painlands, Talismans and shocklands on the same logic.

Recent batches, newest first: **#271** tap-as-cost (convoke / waterbend),
**#269** airbend, **#268** shocklands, **#267** mana-ability riders and
`ManaAbilityCost.Life`, **#261** the top-100 staples batch (32 cards, 20
of them conditional-dual lands), **#258** Hashaton, on top of the S19–S21
engine work: triggered abilities, activated abilities, modal spells,
structured targeting, additional and alternative costs.

## The ranked missing mechanics — the key result

Across the first 2000 cards (batches 01–20); the second and third 2000
get their own tables below. **"Unlocks alone"** counts cards where the
named mechanic is the *only* missing piece — build it and those cards become
writable that day. **"Appears in"** counts every card that needs it at
all, whether or not something else also blocks. **"Dominant blocker for"**
is the grouping the batch issues use.

Ordered by "unlocks alone", because that is the column that answers
"what should we build next".

| Rank | Missing mechanic | Unlocks alone | Appears in | Dominant blocker for | Tracking |
|---:|---|---:|---:|---:|---|
| 1 | Cost modification, alternative casts and costs computed at activation | **120** | 241 | 205 | #93 — **shipped** (S28) |
| 2 | Mana pipeline — restricted / derived mana, mana from a spell, gated or scaled mana abilities | **98** | 183 | 115 | #352 — **shipped** (S32) |
| 3 | Protection / hexproof / ward / indestructible / shroud, damage prevention, copying | **78** | 242 | 119 | #662 / #665 / #666 ¹ |
| 4 | Until-end-of-turn continuous effects (turn-scoped statics) | **59** | 232 | 171 | #279 — **shipped** (S32) |
| 5 | Casting and playing from zones other than hand (flashback, escape, cycling, foretell, impulse) | **59** | 106 | 95 | — |
| 6 | Attachments — Equipment and Auras | **47** | 105 | 104 | #280 — **shipped** (S33) |
| 7 | Library-top placement and ordered look (Brainstorm / tutors / Ponder) | **45** | 98 | 61 | — |
| 8 | Player-scoped and game-rule effects (hand size, extra turns / combats / land drops, command zone) | **39** | 95 | 45 | — |
| 9 | Deferred combat keywords (infect, persist, undying, exalted, landwalk, changeling…) | **28** | 96 | 28 | #705 / #706 / #748 ¹ |
| 10 | Keyword actions with no primitive (proliferate, surveil, explore, connive, amass…) | **24** | 54 | 28 | — |
| 11 | Exile-and-return (blink) and exile-until-leaves | **24** | 52 | 29 | — |
| 12 | "As this enters, choose …" — creature type / colour / name a card | **23** | 47 | 24 | — |
| 13 | Attack / block restrictions and taxes (can't be blocked, attack taxes, must attack) | **21** | 77 | 32 | — |
| 14 | Multi-face cards (MDFC / transform / adventure / split ~~/ class / case~~) | **18** | 69 | 69 | #343 / #719 ² |
| 15 | Card-type completeness — planeswalkers, sagas, vehicles, battles, classes | **13** | 49 | 44 | #92 — **shipped** (S27) |
| 16 | Layer-4 type-changing statics feeding mana derivation (Urborg / Yavimaya / Blood Moon) | **9** | 29 | 14 | — |
| 17 | Per-player / per-turn tallies (storm, second-spell, cast counts, lifegain counts) | **9** | 23 | 9 | — |
| 18 | Change of control (gain control, exchange control) | **7** | 16 | 8 | #756 ³ |
| 19 | Shuffle a card or permanent into a library | **5** | 10 | 6 | — |
| 20 | Recurring self-drawbacks (cumulative upkeep, echo, fading, vanishing, doesn't untap) | **4** | 9 | 4 | — |
| 21 | Table-state mechanics (monarch, initiative, day/night, The Ring, dungeons, speed) | **4** | 6 | 4 | — |
| 22 | Face-down permanents (morph, manifest, disguise, cloak, mutate) | **3** | 14 | 13 | #95 |
| 23 | Regeneration, phasing, totem armor | **2** | 14 | 3 | #667 ¹ |
| 24 | Counters on players (energy, experience, poison, rad, ticket) | **0** | 10 | 1 | — |

¹ *Update 2026-09-16:* these rows used to say #176 (the deferred
combat keywords umbrella) and, on the protection rows, #95 (S30); both
are closed. The counts are from the original pass and are unchanged;
only the **Tracking** column moved. Protection is
[#662](https://github.com/krakenhavoc/cmd_and_ctrl/issues/662); the
hexproof, shroud, ward, indestructible and damage-prevention parts of
that row have all shipped (#353, #421/#433 and #647, #380, #420), and
copying's open pieces are #665 and #666. In the combat-keywords row,
landwalk is [#705](https://github.com/krakenhavoc/cmd_and_ctrl/issues/705)
and prowess is [#706](https://github.com/krakenhavoc/cmd_and_ctrl/issues/706);
changeling shipped in S26 (#404), and persist, undying, exalted
and the other legacy keywords have no tracker and are built on demand,
when a card needs one. *Update 2026-09-17:* infect, with wither and
toxic, now has one,
[#748](https://github.com/krakenhavoc/cmd_and_ctrl/issues/748), filed
from the card-coverage audit. In the last row, regeneration is
[#667](https://github.com/krakenhavoc/cmd_and_ctrl/issues/667); phasing
and totem armor are on demand too. The same footnote applies to the
two tables below.

² *Update 2026-09-16:* these rows used to say #278, the multi-face
spike, which is closed. It delivered ADR 0034 (#290); the face model
and the MDFC face picker shipped in #357, and #574 later let an effect
cast a transform back face. What is still open in the row is the
transform verb, [#343](https://github.com/krakenhavoc/cmd_and_ctrl/issues/343),
and adventure's exile-then-cast,
[#719](https://github.com/krakenhavoc/cmd_and_ctrl/issues/719). Split
fusing, prepare, flip and meld stay declared simplifications in
`deck/validate.go`, with no tracker. `class / case` is struck from the
label because neither layout carries `card_faces` (ADR 0034's first
correction, for `class`; `case` checks out the same against the dump),
so neither is multi-face. The counts are from the original pass, under
the old label, and are unchanged. The same footnote applies to the two
tables below.

³ *Update 2026-09-17:* this row used to say #76, the S24 attachments
tracker, which is closed. S24 shipped layer 2 for a control Aura only
(Mind Control, [ADR 0036](../decisions/0036-attachments.md) decision
17). *Update 2026-09-18:* [#756](https://github.com/krakenhavoc/cmd_and_ctrl/issues/756)
and its prerequisite [#755](https://github.com/krakenhavoc/cmd_and_ctrl/issues/755)
shipped ([ADR 0063](../decisions/0063-durations-and-control.md)): any
spell or ability can now gain control of a permanent for any CR 611.2
duration, and exchange control (CR 701.12). The cards in this row are
unblocked and still have to be written one at a time — Act of Treason,
Agent of Treachery, Sower of Temptation and Switcheroo landed with the
engine work. The counts are from the original pass and are unchanged.
The same footnote applies to the two tables below.

**769 of the 2000 (38%) need no new machinery at all.** That is the most
actionable number in this document: there is more than a sprint of
card-writing available before the next primitive has to land, and it is
spread across every batch rather than bunched at the front.

Three readings worth pulling out:

- **Cost modification (#93) is the biggest single unlock in the format** —
  120 cards on its own, 241 touched. It is also not one thing: cost
  reducers and increasers are a `CostModifier` hook on the S15 cost
  computation; free casts (Force of Will, Flawless Maneuver, the "if you
  control a commander" cycle) are CR 118.9 alternative costs, which
  `AlternativeCosts` almost covers already; kicker and cascade are
  separate again. Splitting #93 by sub-mechanic would probably move the
  first tranche inside one sprint.
- **The mana pipeline is second — tracked as #352 since S32.** 98 cards
  on its own. Four sub-gaps, same as the top-100 triage found: colours
  derived from the board (Exotic Orchard, Reflecting Pool, Chrome Mox,
  Mox Amber), spend restrictions (Cavern of Souls, Delighted Halfling),
  an activation gate (`Temple of the False God`, Mox Opal), and a mana
  component inside a *mana* ability's cost — which is the entire Signet
  cycle, cards that look trivial and were, until #352, unwritable.

  All four shipped in S32 (ADR 0040): `ManaAbilityCost.Mana`,
  `ManaAbility.Condition`, `ManaAbility.ProducedFunc` and
  `ManaAbility.Restrictions`, the last enforced at **spend** time via
  `ManaSpendContext` — the production half alone would have made every
  restricted-mana card stronger than printed, the #259 direction. 20
  cards registered against it (the ten Signets, Cabal Coffers, Gaea's
  Cradle, Temple of the False God, Mox Opal, Exotic Orchard, Reflecting
  Pool, Mox Amber, Delighted Halfling, Shrine of the Forsaken Gods,
  Eldrazi Temple) plus a fix to Fellwar Stone, whose declared
  five-colour simplification was over-permissive. Of the 96 cards still
  naming it as their dominant blocker, at least 29 now need no engine
  work at all — the ten Odyssey filter lands, the seven Verges, the four
  Tainted lands and a dozen more are pure data.
- **Until-end-of-turn (#279) is the clearest case of leverage over
  count.** Only 59 cards name it as their sole blocker, but it appears in
  **232** — nearly one card in eight. It is in flight in S32 and the
  ranking supports that.

### The same ranking over the second 2000 (batches 21–40)

Same detectors, ranks 2,235–4,253.

| Rank | Missing mechanic | Unlocks alone | Appears in | Dominant blocker for | Tracking |
|---:|---|---:|---:|---:|---|
| 1 | Cost modification, alternative casts and costs computed at activation | **110** | 282 | 225 | #93 — **shipped** (S28) |
| 2 | Protection / hexproof / ward / indestructible / shroud, damage prevention, copying | **93** | 248 | 149 | #662 / #665 / #666 ¹ |
| 3 | Deferred combat keywords (infect, persist, undying, exalted, landwalk, changeling…) | **64** | 214 | 64 | #705 / #706 / #748 ¹ |
| 4 | Until-end-of-turn continuous effects (turn-scoped statics) | **63** | 270 | 180 | #279 — **shipped** (S32) |
| 5 | Attachments — Equipment and Auras | **55** | 123 | 117 | #280 — **shipped** (S33) |
| 6 | Mana pipeline — restricted / derived mana, mana from a spell, gated or scaled mana abilities | **47** | 113 | 59 | #352 — **shipped** (S32) |
| 7 | Casting and playing from zones other than hand (flashback, escape, cycling, foretell, impulse) | **40** | 116 | 93 | — |
| 8 | Player-scoped and game-rule effects (hand size, extra turns / combats / land drops, command zone) | **35** | 86 | 46 | — |
| 9 | Library-top placement and ordered look (Brainstorm / tutors / Ponder) | **31** | 91 | 40 | — |
| 10 | Attack / block restrictions and taxes (can't be blocked, attack taxes, must attack) | **30** | 105 | 42 | — |
| 11 | Card-type completeness — planeswalkers, sagas, vehicles, battles, classes | **26** | 87 | 72 | #92 — **shipped** (S27) |
| 12 | Exile-and-return (blink) and exile-until-leaves | **23** | 58 | 32 | — |
| 13 | Table-state mechanics (monarch, initiative, day/night, The Ring, dungeons, speed) | **17** | 27 | 17 | — |
| 14 | "As this enters, choose …" — creature type / colour / name a card | **17** | 35 | 18 | — |
| 15 | Keyword actions with no primitive (proliferate, surveil, explore, connive, amass…) | **16** | 54 | 23 | — |
| 16 | Multi-face cards (MDFC / transform / adventure / split ~~/ class / case~~) | **14** | 72 | 72 | #343 / #719 ² |
| 17 | Counters on players (energy, experience, poison, rad, ticket) | **13** | 38 | 16 | — |
| 18 | Per-player / per-turn tallies (storm, second-spell, cast counts, lifegain counts) | **10** | 38 | 13 | — |
| 19 | Layer-4 type-changing statics feeding mana derivation (Urborg / Yavimaya / Blood Moon) | **9** | 39 | 15 | — |
| 20 | Shuffle a card or permanent into a library | **7** | 21 | 7 | — |
| 21 | Change of control (gain control, exchange control) | **6** | 25 | 8 | #756 ³ |
| 22 | Recurring self-drawbacks (cumulative upkeep, echo, fading, vanishing, doesn't untap) | **6** | 7 | 6 | — |
| 23 | Face-down permanents (morph, manifest, disguise, cloak, mutate) | **4** | 23 | 19 | #95 |
| 24 | Regeneration, phasing, totem armor | **1** | 13 | 7 | #667 ¹ |

**660 of the second 2000 (33%) need no new machinery** — down from 38%
in the first 2000, which is the expected shape: the deeper into the
format you go, the weirder the card. Two rows move enough to matter:

- **Deferred combat keywords jump from 9th to 3rd** — 64 sole
  blockers against 28 in the first 2000. Ranks 2,000–4,000 is where
  infect, persist, undying, exalted, landwalk and changeling live, and
  the bucket buys more cards down here than the mana pipeline does.
  *(Update 2026-09-16: the bucket's umbrella, #176, is closed. Landwalk
  is #705, prowess is #706, changeling shipped in S26, and the rest are
  built on demand; see note ¹ under the first table. Update 2026-09-17:
  infect, wither and toxic are #748.)*
- **Attachments (#280) move from 6th to 5th and the count rises** — 55
  sole, 123 touched. Equipment and Auras are a mid-rarity staple shape,
  not a top-of-format one.

The mana pipeline drops from 2nd to 6th, and library-top from 7th to
9th — both because the first 2000 front-loaded the tutors and the rocks.

### The same ranking over the third 2000 (batches 41–60)

Same detectors, ranks 4,254–6,289.

| Rank | Missing mechanic | Unlocks alone | Appears in | Dominant blocker for | Tracking |
|---:|---|---:|---:|---:|---|
| 1 | Cost modification, alternative casts and costs computed at activation | **114** | 296 | 228 | #93 — **shipped** (S28) |
| 2 | Until-end-of-turn continuous effects (turn-scoped statics) | **77** | 300 | 189 | #279 — **shipped** (S32) |
| 3 | Deferred combat keywords (infect, persist, undying, exalted, landwalk, changeling…) | **73** | 224 | 73 | #705 / #706 / #748 ¹ |
| 4 | Protection / hexproof / ward / indestructible / shroud, damage prevention, copying | **57** | 216 | 106 | #662 / #665 / #666 ¹ |
| 5 | Attachments — Equipment and Auras | **49** | 114 | 111 | #280 — **shipped** (S33) |
| 6 | Mana pipeline — restricted / derived mana, mana from a spell, gated or scaled mana abilities | **40** | 98 | 60 | #352 — **shipped** (S32) |
| 7 | Library-top placement and ordered look (Brainstorm / tutors / Ponder) | **37** | 108 | 53 | — |
| 8 | Attack / block restrictions and taxes (can't be blocked, attack taxes, must attack) | **35** | 114 | 44 | — |
| 9 | Casting and playing from zones other than hand (flashback, escape, cycling, foretell, impulse) | **31** | 111 | 85 | — |
| 10 | Keyword actions with no primitive (proliferate, surveil, explore, connive, amass…) | **26** | 67 | 30 | — |
| 11 | Player-scoped and game-rule effects (hand size, extra turns / combats / land drops, command zone) | **26** | 64 | 31 | — |
| 12 | Exile-and-return (blink) and exile-until-leaves | **22** | 65 | 28 | — |
| 13 | Card-type completeness — planeswalkers, sagas, vehicles, battles, classes | **19** | 106 | 87 | #92 — **shipped** (S27) |
| 14 | Per-player / per-turn tallies (storm, second-spell, cast counts, lifegain counts) | **16** | 40 | 16 | — |
| 15 | Layer-4 type-changing statics feeding mana derivation (Urborg / Yavimaya / Blood Moon) | **14** | 39 | 17 | — |
| 16 | Change of control (gain control, exchange control) | **13** | 26 | 14 | #756 ³ |
| 17 | Table-state mechanics (monarch, initiative, day/night, The Ring, dungeons, speed) | **13** | 25 | 14 | — |
| 18 | "As this enters, choose …" — creature type / colour / name a card | **11** | 23 | 12 | — |
| 19 | Multi-face cards (MDFC / transform / adventure / split ~~/ class / case~~) | **10** | 71 | 71 | #343 / #719 ² |
| 20 | Counters on players (energy, experience, poison, rad, ticket) | **9** | 34 | 13 | — |
| 21 | Face-down permanents (morph, manifest, disguise, cloak, mutate) | **6** | 37 | 32 | #95 |
| 22 | Shuffle a card or permanent into a library | **5** | 14 | 5 | — |
| 23 | Recurring self-drawbacks (cumulative upkeep, echo, fading, vanishing, doesn't untap) | **4** | 9 | 4 | — |
| 24 | Regeneration, phasing, totem armor | **0** | 7 | 4 | #667 ¹ |

**673 of the third 2000 (34%) need no new machinery** — flat against
the second 2000's 33%, so the "it gets harder the deeper you go" curve
has levelled off by rank 4,000. The order at the top is stable; one row
moves enough to matter:

- **Until-end-of-turn (#279) climbs from 4th to 2nd**, and it now has
  the widest reach of any mechanic in the table — **300 of 2000 cards
  touch it**, more than cost modification does. The combat tricks and
  the temporary pumps that fill ranks 4,000–6,000 are exactly its
  shape. It stays the single best ratio of engine work to cards.

Protection slides 2nd to 4th and card types 11th to 13th; library-top
and combat restrictions each gain two places. Regeneration / phasing is
the only mechanic in the table that unlocks *nothing* on its own out
here — 0 sole blockers against 1 in the second 2000 — though it is
still a co-blocker on 7 cards.

## Known traps across the 6000

Orthogonal to the blockers: a card can be implementable today and still be
one of these. The table is the first 2000; the second 2000 follows it.
The third 2000 (batches 41–60) carries the same five traps in the same
order of size — 391 of its 2000 cards trip at least one: deterministic
search pick 106, per-player tally 101, opponent-paid cost 97,
intervening-if 82, stronger-than-printed 31 — and the per-batch lists
are on each batch issue.

| Trap | Cards | What it means |
|---|---:|---|
| `search_chooser` | 138 | rides the catalog-wide **deterministic search pick** until a search chooser lands |
| `opponent_paid` | 106 | needs an **opponent-paid cost or an opponent's choice** (PayUnless / MayPay plumbing) |
| `intervening_if` | 70 | carries an **intervening-if clause** — checked on announce, not re-checked on resolution |
| `per_player_tally` | 68 | needs a **per-player tally** kept across the turn (cast counts, life gained, cards drawn, deaths) |
| `stronger_than_printed` | 27 | would ship **stronger than printed** — the drawback half has no seam (the #259 rule) |

Over the second 2000 the same five traps come out at
`opponent_paid` **107**, `search_chooser` **104**, `per_player_tally`
**86**, `intervening_if` **58**, `stronger_than_printed` **23** —
`opponent_paid` overtaking `search_chooser` is the only reordering, and
the tutor density falling off with rank is why.

`stronger_than_printed` is the one that changes decisions rather than
effort. The rule from #259: a simplification that makes a card **weaker**
than printed is acceptable and must be declared; one that makes it
**stronger** is not, and the card gets skipped instead.

## The 60 batches

| Batch | Rank range | Ready today | Dominant blocking mechanic | Runner-up | Issue |
|---|---|---:|---|---|---|
| 01 | 9–237 | 28 | mana pipeline (16) | cost modification (8) | #294 |
| 02 | 238–360 | 46 | cost modification (12) | mana pipeline (9) | #295 |
| 03 | 361–471 | 48 | mana pipeline (10) | cost modification (9) | #296 |
| 04 | 472–577 | 38 | until EOT (11) | other-zone casting (8) | #297 |
| 05 | 578–693 | 35 | cost modification (13) | other-zone casting (8) | #298 |
| 06 | 694–797 | 35 | cost modification (13) | mana pipeline (10) | #299 |
| 07 | 798–901 | 40 | mana pipeline (9) | cost modification (8) | #300 |
| 08 | 902–1004 | 43 | protection / prevention (8) | cost modification (8) | #301 |
| 09 | 1005–1110 | 44 | until EOT (10) | cost modification (9) | #302 |
| 10 | 1111–1213 | 38 | cost modification (13) | until EOT (12) | #303 |
| 11 | 1214–1315 | 38 | cost modification (11) | protection / prevention (9) | #304 |
| 12 | 1316–1420 | 36 | cost modification (11) | until EOT (9) | #305 |
| 13 | 1421–1524 | 43 | protection / prevention (9) | cost modification (8) | #306 |
| 14 | 1525–1625 | 45 | cost modification (15) | attachments (9) | #307 |
| 15 | 1626–1726 | 35 | until EOT (9) | cost modification (9) | #308 |
| 16 | 1727–1829 | 38 | cost modification (9) | protection / prevention (8) | #309 |
| 17 | 1830–1929 | 38 | cost modification (11) | other-zone casting (9) | #310 |
| 18 | 1930–2031 | 45 | until EOT (12) | cost modification (6) | #311 |
| 19 | 2032–2132 | 27 | until EOT (13) | protection / prevention (10) | #312 |
| 20 | 2133–2234 | 29 | cost modification (15) | until EOT (9) | #313 |
| 21 | 2235–2334 | 37 | cost modification (15) | attachments (10) | #383 |
| 22 | 2335–2435 | 34 | cost modification (10) | until EOT (10) | #384 |
| 23 | 2436–2535 | 21 | cost modification (17) | until EOT (7) | #385 |
| 24 | 2536–2637 | 40 | cost modification (10) | until EOT (9) | #386 |
| 25 | 2638–2737 | 32 | protection / prevention (9) | cost modification (7) | #387 |
| 26 | 2738–2838 | 35 | until EOT (8) | protection / prevention (8) | #388 |
| 27 | 2839–2939 | 29 | cost modification (17) | until EOT (12) | #389 |
| 28 | 2940–3040 | 37 | cost modification (11) | protection / prevention (7) | #390 |
| 29 | 3041–3144 | 33 | until EOT (16) | cost modification (10) | #391 |
| 30 | 3145–3244 | 32 | until EOT (12) | cost modification (11) | #393 |
| 31 | 3245–3345 | 24 | until EOT (16) | protection / prevention (10) | #394 |
| 32 | 3346–3446 | 30 | cost modification (14) | protection / prevention (10) | #395 |
| 33 | 3447–3548 | 34 | cost modification (18) | until EOT (8) | #396 |
| 34 | 3550–3649 | 42 | cost modification (9) | protection / prevention (8) | #397 |
| 35 | 3650–3752 | 28 | cost modification (13) | until EOT (11) | #398 |
| 36 | 3753–3852 | 33 | cost modification (8) | until EOT (8) | #399 |
| 37 | 3853–3952 | 40 | until EOT (13) | cost modification (10) | #400 |
| 38 | 3953–4052 | 36 | cost modification (10) | protection / prevention (9) | #401 |
| 39 | 4053–4153 | 35 | cost modification (11) | protection / prevention (10) | #402 |
| 40 | 4154–4253 | 28 | cost modification (13) | card types (9) | #403 |
| 41 | 4254–4353 | 37 | cost modification (9) | attachments (8) | #448 |
| 42 | 4354–4455 | 38 | until EOT (13) | cost modification (12) | #449 |
| 43 | 4456–4555 | 37 | cost modification (12) | until EOT (11) | #450 |
| 44 | 4556–4657 | 39 | until EOT (15) | protection / prevention (9) | #451 |
| 45 | 4658–4758 | 28 | cost modification (14) | until EOT (6) | #452 |
| 46 | 4759–4861 | 30 | until EOT (14) | cost modification (13) | #453 |
| 47 | 4862–4962 | 31 | until EOT (13) | card types (7) | #454 |
| 48 | 4963–5066 | 30 | cost modification (11) | protection / prevention (7) | #455 |
| 49 | 5067–5167 | 37 | until EOT (11) | cost modification (11) | #456 |
| 50 | 5168–5271 | 30 | cost modification (14) | until EOT (11) | #457 |
| 51 | 5272–5372 | 37 | until EOT (10) | deferred keywords (8) | #458 |
| 52 | 5373–5472 | 34 | cost modification (13) | attachments (8) | #459 |
| 53 | 5473–5573 | 30 | cost modification (11) | until EOT (10) | #460 |
| 54 | 5574–5676 | 31 | cost modification (15) | until EOT (8) | #461 |
| 55 | 5677–5777 | 39 | cost modification (9) | until EOT (8) | #462 |
| 56 | 5778–5879 | 37 | cost modification (8) | until EOT (7) | #463 |
| 57 | 5880–5981 | 36 | cost modification (14) | card types (7) | #464 |
| 58 | 5982–6083 | 27 | cost modification (15) | until EOT (9) | #465 |
| 59 | 6084–6184 | 25 | cost modification (14) | other-zone casting (10) | #466 |
| 60 | 6185–6289 | 40 | cost modification (15) | until EOT (11) | #467 |

Batches 21–60 are **not** sub-issues of #293. The tracking issue's
sub-issue list is a field on #293 itself, and the sessions that created
these forty were scoped to creating issues only — the links are a
one-line `gh` call for whoever owns the umbrella.

## What "done" means for a batch

Every card in a batch's table is either **registered** in
`server/internal/cards/effects/` with a test, or **declared skipped** with
the reason on that batch's issue. Simplifications go on the card file as
comments. Batches need not land in one PR, and within a batch the "no new
machinery" group is the part that can start immediately.
