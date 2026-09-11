# ADR 0012 — Continuous effects + layer system (S16)

**Status:** Accepted · 2026-04-22 · Sprint S16

## Context

S15 closed out the mana economy; S16 is the rules-engine arc's
hard sprint — continuous effects per CR 613. After S16 the catalog
can include cards whose effects exist *while a permanent is on the
battlefield* rather than just on resolution: anthems, type-changers,
keyword-granters, characteristic-defining abilities. The 4 starter
cards exercise four of the in-scope layers:

- **Glorious Anthem** — Layer 7c (modify P/T)
- **Mycosynth Lattice** — Layer 4 (type-add)
- **Lord of Atlantis** — Layer 6 (ability grant) + Layer 7c
- **Tarmogoyf** — Layer 7a (CDA — characteristic-defining ability)

Continuous effects are the second-most complex MTG subsystem after
the stack. The naive approach — "apply effects in chronological
order" — is provably wrong for many real cards. CR 613 fixes this
with a 7-stage pipeline that recomputes characteristics from
scratch on every state change. S16 ships the 7-layer skeleton with
timestamp-only ordering. Dependency detection (CR 613.8 — the
Opalescence + Humility class of pathological cycle) is
intentionally deferred to S16.5; pure timestamp ordering covers
~95% of real cards and never matters in casual EDH.

## Decisions

### 1. Recompute-from-scratch with a version counter, not
       incremental diffs

`Game.layerVersion` (atomic.Uint64) is bumped by listener hooks on
every event that could change which static abilities are active or
what they apply to. `Game.lastResolvedVersion` mirrors the most
recent version the recompute pass has caught up to. The snapshot
path's `RecomputeLayersIfStaleLocked` no-ops when they match;
otherwise it runs the body. (As specified it serialised through a
dedicated `recompute.mu` mutex; the S24 amendment to decision 11
replaced that with the game's write lock.)

**Why:** XMage / Forge approach. Maintaining incremental diffs is a
correctness nightmare — every possible mutation has to know which
layer-effects might depend on it. Recompute from scratch is O(N×L)
per pass where N is battlefield size (~40) and L is layer-7 sub-
layer count; sub-millisecond on modern CPUs and trivially correct.
The version counter ensures we only pay the recompute cost when
something actually changed.

**Where the recompute fires:** `ReadSnapshot` resolves a stale
recompute before invoking the projection. (S24 amendment, see
decision 11: it does that by upgrading to the game's write lock,
not by recomputing under the read lock as originally specified.)

**Where it does NOT fire:** inside the SBA loop. SBA mutations
emit zone-change events that bump `layerVersion`; if the recompute
ran inside the loop, it would interleave with SBA evaluation and
mask state-based effects. The snapshot-only entry point is the
safe placement.

### 2. Two characteristic shapes: printed vs. effective

`Card.printedCharacteristic()` is derived from the immutable
printed fields (Power, Toughness, TypeLine, ManaCost). No new
storage on `Card` for printed — the existing fields ARE the
printed values.

`Card.Effective() Characteristic` returns the post-layer-resolution
view, cached in `Card.effective *Characteristic`. Sub-PR 1 ships
the API with a no-op pass (effective == printed); sub-PR 3 makes
the recompute populate `effective` for real.

Wire ships `Effective`; rules logic that needs printed values
(commander tax base, owner) reads `Card` directly.

### 3. Counter math stays in `CurrentPower` / `CurrentToughness`

Layer 7d's contract is "+1/+1 and -1/-1 counter modifications." The
existing `Card.CurrentPower()` / `CurrentToughness()` (from S13.2)
already do this math. S16 keeps them as the canonical counter-
math path; Layer 7d will *call* them rather than reimplement.

**Why:** Minimum blast radius on S13.2's SBA suite. Folding
counter math into Layer 7d would require touching every test that
asserts on `CurrentPower()` (lethal-damage SBA, planeswalker
loyalty 0, battle defense 0). Sub-PR 5 doesn't ship the 7d
delegation hookup yet — Tarmogoyf ships at 7a (sets P/T) without
needing 7d, and Glorious Anthem ships at 7c (modifies P/T) without
needing 7d. When the first card needs explicit counter integration
through the layer engine, the delegation lands as a one-line Apply
in 7d that reads `CurrentPower() - c.Power`.

### 4. Skip dependency detection (CR 613.8) — pure timestamp
       ordering

Effects are sorted by their source's `EnteredBattlefieldAt` Unix-
nano timestamp ascending within each layer (stable sort, so
same-timestamp ties resolve to gather order). The CR 613.8
dependency-detection rule (apply A before B if A's existence
changes whether B applies) is **not** implemented.

**Why:** The pathological case is Opalescence + Humility — niche
combo that essentially never shows up in casual EDH. The
implementation cost of dependency detection is high (every effect
needs to declare its dependency inputs; cycles need detection).
Pure timestamp ordering produces correct results for ~95% of real
cards. S16.5 follow-up if a real card surfaces in playtesting.

### 5. Layer 7 ships full sub-layer support; layers 1, 3, 5 ship
       as stubs

Layer 7 sub-layers (7a CDA, 7b set, 7c modify, 7d counters, 7e
switch) are needed by the in-scope cards: Tarmogoyf at 7a,
anthems at 7c. 7b and 7e ship the bucket in `layerOrder` but no
catalog card exercises them yet.

Layer 1 (copy effects — Clone, Phyrexian Metamorph) — deferred to
S16.5.

Layer 3 (text-changing effects — Mind Bend, Glamerdye) — engine
ships the layer-3 bucket but no in-scope card needs it.

Layer 5 (color-changing effects — Painter's Servant) — engine
ships the bucket; no catalog card. The commander-identity proxy
replacement (sub-PR 5) reads `Effective().Colors` so a future
Layer-5 card lights up identity automatically.

### 6. Static abilities declared on `Spec.Static`

`effects.Spec.Static []game.StaticAbility` is a parallel field to
`OnETB`, `OnResolve`, `ManaAbilities`. Each entry declares its
`Layer` + `SubLayer` + an `AppliesTo` predicate (evaluated per
candidate target on every recompute) + an `Apply` function (mutates
the candidate's `Characteristic` in place).

`game.StaticAbility` lives in the `game` package — card files
reference it via the existing `effects → game` import without
going through a parallel adapter type (the function shapes already
reference `game.Card` / `game.Game` / `game.Characteristic`, so
nothing would survive a separate struct).

A 6th function-var hook `game.CatalogStaticAbilities` joins the
existing five (EffectResolver, ETBEffectHook, IsCatalogCard,
CatalogTargetMode, CatalogManaAbilities). Same dependency-inversion
pattern as S14: `effects/wire.go` populates the hook at init time
as a thin Lookup pass-through.

### 7. `EnteredBattlefieldAt` is the layer ordering key,
       maintained by the listener

`Card.EnteredBattlefieldAt int64` (Unix-nano) is stamped by the
layer-version-bump listener on every `EventZoneMove` with
`NewZone == battlefield`. The listener also clears
`Card.effective` on the battlefield-leave path so the next
recompute rebuilds from the printed baseline.

**Why a listener instead of every entry-site call:** the listener
is single source of truth. Audit of every `g.Battlefield` push
site would have been brittle (~5 callsites to keep in sync); the
listener keys off `EventZoneMove` which is universally emitted.
The same listener bumps `layerVersion` on the same event — one
read, one write, both correctness-critical.

### 8. Wire surface: `CardView.power` / `toughness` / `type_line`
       become *effective*; new `abilities []string` field

`viewOfCard` reads from `Card.Effective()` for Power, Toughness,
Name, and Abilities. The existing `CardView.power`, `toughness`,
`type_line` keys keep their names; values become post-layer.

`CardView.abilities []string` is a NEW field — drives S18's
keyword renderer. Populated by Layer-6 ability grants (Lord of
Atlantis grants `["islandwalk"]` to other Merfolk). S16 just
exposes the keyword grant on the wire; combat behavior of the
keywords lands with S18.

`type_line` reconstruction is via `effectiveTypeLine(c, eff)` —
returns the printed string verbatim when no Layer-4 type change
has occurred (common case, byte-for-byte identical to pre-S16
output), otherwise rebuilds "Supertypes Types — Subtypes" from
`eff.Supertypes/Types/Subtypes`.

**No protocol bump.** Same JSON keys, same shape; values diverge
from printed only when a static ability is active.

### 9. Battlefield-only static abilities (CR 113.6 default)

Continuous effects from non-battlefield zones (Yixlid Jailer's
"cards in graveyards lose all abilities" while in graveyard) are
deferred. The recompute pass walks `g.Battlefield` only when
collecting active abilities. The CR 113.6 default ("an effect
generated by an ability of an object only does anything if that
object is on the battlefield") is the S16 simplification.

When a future card needs a non-battlefield static, the gather pass
can extend to other zones via per-card opt-in. Out of scope for
S16.

### 10. Commander-identity proxy replacement

S15's hand-rolled `distinctColorsInManaCost` was the explicit
S15→S16 hand-off per ADR 0011 decision 6. S16 replaces it:
`commanderIdentityFor` reads `commander.Effective().Colors`.

To make the layer-aware path actually meaningful,
`printedCharacteristic` populates `Colors` from `ManaCost` via
`printedColorsFromCost`. Result: same identity for ~99% of EDH
commanders (whose identity is fully captured by their mana cost),
but a future Layer-5 color-change effect on a commander mutates
`Effective().Colors` and the identity computation picks the
change up automatically.

Falls back to the printed-cost proxy when `Effective().Colors` is
empty — covers placeholder commanders with no parsed cost (lobby-
time pre-effects-init path).

### 11. Lock semantics: atomic version + dedicated recompute mutex

> **Superseded in S24.** This decision was wrong and shipped a data
> race. See the amendment below; the current contract is that the
> recompute requires the game's WRITE lock.

Reading `layerVersion` and `lastResolvedVersion` is atomic. Bumps
happen under the game write lock (the EmitEvent path) — no race
with snapshot readers. The recompute body runs under
`g.recompute.mu` (separate from `g.mu`) so multiple readers can
contend without promoting the read lock. Double-check the version
after acquire to avoid duplicate work when readers race.

Listener bumps + recompute reads are the only writes to the
atomic version counters. The recompute writes `Card.effective`
under the recompute mutex; concurrent readers may see either the
old or new pointer (atomic pointer assignment in Go), which is
fine — both are valid Characteristic snapshots.

#### Amendment (S24): the recompute takes the write lock

Two things above are false.

**"Multiple readers can contend without promoting the read lock"**
confuses two questions. The recompute mutex serialises recomputes
against **each other**; it does nothing about a goroutine holding
only `g.mu` in read mode, and `sync.RWMutex` admits any number of
those simultaneously. Every concurrent read-lock holder raced the
pass. `Game.AutoTapForCostExcluding` (the lobby `/autotap` preview,
on a net/http goroutine) and `Game.ControllerOfCard` (the actions
package's authorization gate) both walk the battlefield under the
read lock while `protocol.ViewOfGame` recomputes on another.

**"Atomic pointer assignment in Go, which is fine"** is not a
guarantee the language makes. An unsynchronised pointer write
concurrent with a read is a data race under the Go memory model,
with no promise the reader sees either the old or the new value;
`go test -race` reports it as one. The only reason CI stayed green
is that no test exercised a concurrent reader against a recompute —
absence of a race report from a suite with no concurrency in it is
not evidence of absence.

S24's layer 2 raised the stakes rather than creating them: the
materialisation pass writes `Card.Controller`, a 16-byte
`uuid.UUID`, so a torn read silently attributes a permanent to the
wrong seat instead of crashing on a bad pointer.

**Current contract.** `RecomputeLayersIfStaleLocked` requires
`g.mu` in **write** mode. Every mutator that calls it already held
it. `ReadSnapshot` — the one caller that did not — upgrades: it
checks the version counters under the read lock, and when stale
releases it, retakes `g.mu` in write mode for the recompute alone,
and re-enters read mode, looping because a mutation can land in the
gap and `fn()` must see fresh layers. `g.recompute.mu` is deleted;
the write lock is the serialiser.

Cost, measured on a 40-permanent board: the fast path — two atomic
loads, which is what almost every broadcast hits, because every
version bump happens under the write lock and thirteen mutators
recompute before releasing — is unchanged at ~4ns. A stale
recompute goes 6.5µs → 7.3µs, under a percent of `ViewOfGame`'s
hundreds of microseconds. The contended case is flat and less
variable. `server/internal/game/layers_concurrency_test.go` pins
the three pairings under `-race`.

## Consequences

- **4 catalog cards** ship with static abilities: Glorious Anthem
  (7c), Mycosynth Lattice (4), Lord of Atlantis (6 + 7c),
  Tarmogoyf (7a).
- **Sixth catalog hook** `CatalogStaticAbilities` joins the
  S14/S15 five.
- **`CardView.power`, `toughness`, `type_line`** are now post-
  layer effective; the wire JSON keys are unchanged.
- **`CardView.abilities []string`** new field — drives S18
  keyword renderer; populated for Lord-pumped Merfolk today.
- **S15's commander-identity proxy** replaced with the layer-aware
  computation; behavior unchanged for cards in S16 scope, but the
  path is correct for future Layer-5 color effects.
- **`EventLTB`** is now consumed by the layer listener (audit
  showed it was already emitted at all battlefield-leave sites
  pre-S16; no new emission sites needed).
- **`game.ParseTypeLine`** exported so the protocol layer can
  round-trip the printed-vs-effective comparison without
  re-implementing the parser.
- **Test patterns**: catalog cards push to battlefield via
  `pushBattlefieldCardWithTimestamp` (fires `EventZoneMove` so the
  listener stamps EBA + bumps layerVersion); read effective
  characteristics via `effectivePower` / `effectiveToughness` /
  `effectiveTypes` / `effectiveAbilities` helpers in
  `effects/anthem_test.go` and friends.
- **S17** picks up replacement effects (CR 614). Mind Control's
  control-change is the canonical S17 entry (deferred from S16
  scope due to aura-attachment overlap).
- **S18** picks up combat keywords. Lord of Atlantis already
  exposes `["islandwalk"]` on the wire; S18 reads that list to
  drive combat behaviour.
- **S19** picks up triggered abilities. The event log + listener
  registry are the integration point.
- **Dependency detection (CR 613.8)** deferred to S16.5 if a real
  card surfaces in playtesting.
- **Decision 2's "wire ships `Effective`" was only half the
  hand-off**, and the other half sat unbuilt for four sprints:
  `Card.IsCreature` / `IsLand` / `IsArtifact` kept reading the
  printed `TypeLine`, so a Layer-4 type change reached the client
  and nothing else. **ADR 0039** routes the type predicates through
  `Effective()` and splits off an explicit `PrintedIs*` surface for
  callers that want the printed values decision 2 reserves for
  them.
