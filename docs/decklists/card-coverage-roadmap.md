# The card-coverage roadmap — the next 2000 Commander cards

Source: **`edhrec_rank`** on every card in the local Scryfall bulk dump
(`data/scryfall/default-cards.json`, refreshed 2026-09-09). Lower rank
means more played. Nothing here was scraped, recalled, or guessed — one
pass over the dump and one probe of the live registry produced every
number below.

This is the whole-format sequel to
[the top-100 staples triage](top-100-commander-staples.md): that file
ranked the first 100 cards of the gap and shipped 32 of them, this one
ranks the next **2000** and splits them into 20 tracked batches of 100.
Tracking issue: **#293**.

## Progress

The map is being walked. This section is the running tally; the
per-card detail lives on each batch's issue, and every simplification
is a comment on its card file.

| Batch | Issue | Registered | Skipped (declared) | Still blocked | Notes |
|---|---|---:|---:|---:|---|
| 01 | #294 | **29** | 2 | 69 | first pass — the "no new machinery" group, plus 3 cards two engine changes unblocked |
| 02–20 | #295–#313 | 0 | 0 | — | not started |

**Catalog: 288 → 319.** The 288 the audit below counted, plus Giant
Growth and Overrun from the until-end-of-turn work (#314), plus the 29
here. (The test binary reports 320 — the flicker probe, as explained
below.) Batch 01 alone moves the play-rate coverage by **+2 in the top
100** (Dark Ritual, Arcane Denial), **+17 in the top 200**, **+29 in
the top 300**.

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

### 3. The mechanic triage

Each of the 2000 is checked against the primitive set that actually exists
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

## What the catalog covers today

288 cards, by shape:

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

Across all 2000 cards. **"Unlocks alone"** counts cards where the named
mechanic is the *only* missing piece — build it and those cards become
writable that day. **"Appears in"** counts every card that needs it at
all, whether or not something else also blocks. **"Dominant blocker for"**
is the grouping the batch issues use.

Ordered by "unlocks alone", because that is the column that answers
"what should we build next".

| Rank | Missing mechanic | Unlocks alone | Appears in | Dominant blocker for | Tracking |
|---:|---|---:|---:|---:|---|
| 1 | Cost modification, alternative casts and costs computed at activation | **120** | 241 | 205 | #93 |
| 2 | Mana pipeline — restricted / derived mana, mana from a spell, gated or scaled mana abilities | **98** | 183 | 115 | — |
| 3 | Protection / hexproof / ward / indestructible / shroud, damage prevention, copying | **78** | 242 | 119 | #95 / #176 |
| 4 | Until-end-of-turn continuous effects (turn-scoped statics) | **59** | 232 | 171 | #279 |
| 5 | Casting and playing from zones other than hand (flashback, escape, cycling, foretell, impulse) | **59** | 106 | 95 | — |
| 6 | Attachments — Equipment and Auras | **47** | 105 | 104 | #280 |
| 7 | Library-top placement and ordered look (Brainstorm / tutors / Ponder) | **45** | 98 | 61 | — |
| 8 | Player-scoped and game-rule effects (hand size, extra turns / combats / land drops, command zone) | **39** | 95 | 45 | — |
| 9 | Deferred combat keywords (infect, persist, undying, exalted, landwalk, changeling…) | **28** | 96 | 28 | #176 |
| 10 | Keyword actions with no primitive (proliferate, surveil, explore, connive, amass…) | **24** | 54 | 28 | — |
| 11 | Exile-and-return (blink) and exile-until-leaves | **24** | 52 | 29 | — |
| 12 | "As this enters, choose …" — creature type / colour / name a card | **23** | 47 | 24 | — |
| 13 | Attack / block restrictions and taxes (can't be blocked, attack taxes, must attack) | **21** | 77 | 32 | — |
| 14 | Multi-face cards (MDFC / transform / adventure / split / class / case) | **18** | 69 | 69 | #278 |
| 15 | Card-type completeness — planeswalkers, sagas, vehicles, battles, classes | **13** | 49 | 44 | #92 |
| 16 | Layer-4 type-changing statics feeding mana derivation (Urborg / Yavimaya / Blood Moon) | **9** | 29 | 14 | — |
| 17 | Per-player / per-turn tallies (storm, second-spell, cast counts, lifegain counts) | **9** | 23 | 9 | — |
| 18 | Change of control (gain control, exchange control) | **7** | 16 | 8 | #76 |
| 19 | Shuffle a card or permanent into a library | **5** | 10 | 6 | — |
| 20 | Recurring self-drawbacks (cumulative upkeep, echo, fading, vanishing, doesn't untap) | **4** | 9 | 4 | — |
| 21 | Table-state mechanics (monarch, initiative, day/night, The Ring, dungeons, speed) | **4** | 6 | 4 | — |
| 22 | Face-down permanents (morph, manifest, disguise, cloak, mutate) | **3** | 14 | 13 | #95 |
| 23 | Regeneration, phasing, totem armor | **2** | 14 | 3 | #176 |
| 24 | Counters on players (energy, experience, poison, rad, ticket) | **0** | 10 | 1 | — |

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
- **The mana pipeline is second and has no tracking issue.** 98 cards on
  its own. Four sub-gaps, same as the top-100 triage found: colours
  derived from the board (Exotic Orchard, Reflecting Pool, Chrome Mox,
  Mox Amber), spend restrictions (Cavern of Souls, Delighted Halfling),
  an activation gate (`Temple of the False God`, Mox Opal), and a mana
  component inside a *mana* ability's cost — which is the entire Signet
  cycle, five cards that look trivial and are currently unwritable.
- **Until-end-of-turn (#279) is the clearest case of leverage over
  count.** Only 59 cards name it as their sole blocker, but it appears in
  **232** — nearly one card in eight. It is in flight in S32 and the
  ranking supports that.

## Known traps across the 2000

Orthogonal to the blockers: a card can be implementable today and still be
one of these.

| Trap | Cards | What it means |
|---|---:|---|
| `search_chooser` | 138 | rides the catalog-wide **deterministic search pick** until a search chooser lands |
| `opponent_paid` | 106 | needs an **opponent-paid cost or an opponent's choice** (PayUnless / MayPay plumbing) |
| `intervening_if` | 70 | carries an **intervening-if clause** — checked on announce, not re-checked on resolution |
| `per_player_tally` | 68 | needs a **per-player tally** kept across the turn (cast counts, life gained, cards drawn, deaths) |
| `stronger_than_printed` | 27 | would ship **stronger than printed** — the drawback half has no seam (the #259 rule) |

`stronger_than_printed` is the one that changes decisions rather than
effort. The rule from #259: a simplification that makes a card **weaker**
than printed is acceptable and must be declared; one that makes it
**stronger** is not, and the card gets skipped instead.

## The 20 batches

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

## What "done" means for a batch

Every card in a batch's table is either **registered** in
`server/internal/cards/effects/` with a test, or **declared skipped** with
the reason on that batch's issue. Simplifications go on the card file as
comments. Batches need not land in one PR, and within a batch the "no new
machinery" group is the part that can start immediately.
