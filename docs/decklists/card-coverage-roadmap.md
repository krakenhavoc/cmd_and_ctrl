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

## The audited catalog count — 288

Not 289, and not the number `grep -c 'OracleID:'` reports.

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
