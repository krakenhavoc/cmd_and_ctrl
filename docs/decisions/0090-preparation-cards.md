# ADR 0090 — Preparation cards: the prepared designation and casting the prepare spell as a copy

**Status:** Accepted · 2026-09-23 · S46 — permanents that change what they are
**Issues:** [#1328](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1328) (this seam),
[#1306](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1306) (deck tracker "Aang is so flashy" — Skycoach Conductor // All Aboard)
**Trackers:** [#889](https://github.com/krakenhavoc/cmd_and_ctrl/issues/889) (S46 — permanents that change what they are),
[#885](https://github.com/krakenhavoc/cmd_and_ctrl/issues/885) (S42 — casting from non-hand zones, where #1328 was filed)
**Numbering:** swept with the AGENTS.md §4 loop on 2026-09-23 — `git fetch origin --prune`, then
`git ls-tree --name-only <ref> docs/decisions/` over every remote head (413 refs) plus the recent
issue claim comments. The highest number present anywhere is **0089** (`0089-gift.md`, on
`feat/gift`, reserved on #1267); `0090` appears on no branch, and the claim comment on #1328
reserves it.

**Related:** [ADR 0071](0071-designations-that-switch-abilities-on.md) (the designation family this
joins), [ADR 0066](0066-granted-cast-and-play-permissions.md) (cast permissions, and why standing
ones are derived), [ADR 0043](0043-copy-effects.md) (copiable values), [ADR 0034](0034-multi-face-cards.md)
(the multi-face model, which recorded prepare as out of scope), [ADR 0084](0084-phasing.md) (phasing
keeps designations).

---

## Context

A preparation card (CR 722) is a permanent card with a second set of characteristics printed in an
inset frame: its **prepare spell** (CR 722.2a). Skycoach Conductor is a {2}{U} 2/3 flying
vigilance Bird Pilot with flash; All Aboard, printed beside it, is a {U} instant. The rules, checked
against the pinned edition (`MagicCompRules 20260819.txt`, effective August 7, 2026):

- **CR 722.3** — the card is never cast using the prepare spell's characteristics; they "define
  characteristics of copies which may be cast".
- **CR 722.3a** — some effects make a permanent with a prepare spell "become prepared", or say it
  "enters prepared". Prepared is a designation. A permanent without a prepare spell can't gain it,
  and a permanent that already has it can't gain it again.
- **CR 722.3b** — "unprepared" removes it.
- **CR 722.3c** — as a permanent gains the designation *or phases in prepared*, its controller
  creates a copy of it in exile that has **only** the prepare spell's characteristics, which become
  the copy's normal characteristics. The copy stays in exile for as long as the permanent stays on
  the battlefield with the designation — an exception to CR 704.5e. While the copy is there, the
  permanent's controller may cast it, and the permanent loses the designation as the spell becomes
  cast (CR 601.2i).
- **CR 722.2b** — the prepare spell is part of the object's copiable values.

What the engine had: ADR 0034 imports both faces, and `CastableFaces` offers a `prepare` card's
front face only — which is right under CR 722.3, so ADR 0034's "cast as its creature half only"
was never a simplification of the CAST. What was missing was everything else: no designation, no
copy in exile, and — the harder part, recorded by the Paradigm row in `engine-seams.md` — "every
cast permission in the engine opens the CARD, and 'cast a copy of a card in exile' exists nowhere".

## Decision 1 — Prepared is a designation on the Card, beside ADR 0071's

`Card.Prepared` sits with `ClassLevel` and `Solved`, and follows their rules: battlefield state,
not a characteristic and not copiable; cleared by CR 400.7 in `MoveCard`'s battlefield-exit block
and in the exile return's new-object reset; kept across phasing (CR 702.26d) because phasing does
not travel through either; carried by the snapshot.

It is **not** a new `DesignationKind`. ADR 0071's gate exists to switch a permanent's own printed
abilities on, and no printed preparation card has an ability "as long as it is prepared". What the
designation switches on is a *cast*, which is Decision 3's business. Adding a gate kind nobody
declares would be the reserved-but-unbuilt shape ADR 0071 already carries once (the Room door),
with no card to build it for.

`becomePreparedLocked` is the one writer. It refuses both things CR 722.3a refuses — a permanent
with no prepare spell (`HasPrepareSpell`: layout `prepare`, two faces, and not a face-down
CR 708.2 object) and one already prepared — as a quiet no-op rather than an error, because the
rule says the permanent simply doesn't gain it: Skycoach Waypoint pointed at a Grizzly Bears
resolves and does nothing.

## Decision 2 — The copy is a real object in exile, and it is marked as not a card

`createPrepareCopyLocked` builds the copy from the permanent's **copiable values**
(`CopiableValuesOf`, CR 707.2) — so a Clone that copied a Skycoach Conductor has All Aboard's copy
to cast, which is CR 722.2b — then makes face 1 active. It keeps the oracle ID and the face list,
so `CatalogKey` answers `"<oracle>#1"` and the copy resolves through the prepare spell's own
catalog entry with no new lookup. It drops the three card-level fields that are not the prepare
spell's: `Keywords` (the importer stamps the front face's printed keywords at card level, and a
flash creature's copy of its sorcery would otherwise be castable at instant speed),
`GrantedAbilities` and `ProducedMana` ("ignoring other exceptions to the copying process").

Two new fields on the copy:

- `PrepareCopy bool` — this object is not a card (CR 707.10). It is what the sweep in Decision 5
  reads, what `MoveCard`'s CR 712.8a front-face reset now skips (CR 722.3c makes the prepare spell
  the copy's *normal* characteristics, so a countered All Aboard is never, even for a moment, a
  Skycoach Conductor in a graveyard), and what `CastSpell` reads to cast it as a copy.
- `PreparedBy PermissionCardRef` — the permanent **object** (instance and CR 400.7 epoch) that keeps
  it castable. `MoveCard` clears it on every move, so a copy that leaves exile by any route — the
  cast, a counter, an effect — names nothing ever again, even if it comes back to exile while the
  same permanent is prepared once more.

The copy is created by, owned by and controlled by the permanent's controller, is public, and is
created in place (`PushTop`), not moved there: it emits no zone move, so nothing watching "a card is
put into exile" sees it arrive.

## Decision 3 — The cast permission is derived, never stored

`prepareCopyPermissionLocked` answers the permission over a copy from the live permanent on every
query: a ScopeCards permission for the permanent's **current** controller, opening face 1 only,
paying the printed cost, with no timing of its own. `CastPermissionForLocked` (the cast path and the
enumerator), `CastPermissionOnCardForEffect` (the wire's `exile_play`) and
`AnyCastPermissionsForEffect` (both callers' fast negative) consult it, so all three see one answer.

Derived for ADR 0066's own reason for deriving standing permissions. CR 722.3c's "for as long as the
prepared permanent remains on the battlefield and has the prepared designation" is then true **by
construction**: the permission follows a change of control with no bookkeeping, and a copy whose
permanent has died, been flickered or been unprepared is uncastable in the same instant, before any
sweep has run. A stored permission would need a writer at each of those moments, and a missed one
is a castable copy with no permanent behind it.

And a prepare copy answers **only** through this permission: `CastPermissionForLocked` returns the
derived answer or nil for a `PrepareCopy`, and never falls through to stored or standing
permissions, because an impulse grant or a standing rule over exile reaching the copy would cast an
object CR 722.3c keeps for one player.

## Decision 4 — Casting it is the ordinary cast out of exile, plus two lines

The copy is cast through `CastSpell` from `exile` like any granted card, so every announce gate —
timing, targets, the cast gate, costs — reads it unchanged. Two lines are new:

- `StackItem.IsCopy` is set from `PrepareCopy`. The spell is then a CR 707.10 copy on the stack, and
  the resolution frame's existing branches end it (it ceases to exist rather than going to a
  graveyard) — the path Twincast's copies already take. It is still a **cast**: EventCast fires, the
  cast tally counts it and "whenever you cast a spell from exile" (Appa) sees it, all of which CR
  707.12 says ("casting a copy of an object follows steps 601.2a–h … then the copy becomes cast").
- The permanent is unprepared "at the time the spell becomes cast" (CR 601.2i) — after every cost is
  paid and just before `EventCast`. The value copy of the card read out of exile still names the
  permanent; the copy itself lost the link to `MoveCard` a moment earlier.

## Decision 5 — One sweep: CR 704.5e with CR 722.3c's exception

`prepareCopySweepSBALocked` runs beside CR 704.5d's token sweep: a `PrepareCopy` in any zone but the
stack ceases to exist, unless it is in exile and `PreparedBy` names a permanent that is still that
object, on the battlefield and prepared. This one rule clears the copy when its permanent dies, is
bounced or flickered (a new object — Skycoach Conductor blinked by another effect enters prepared
again with a fresh copy), phases out, or when the cast copy is countered into a graveyard or exile.
An effect that unprepares the permanent (CR 722.3b) removes the copy immediately as well, since
that is the one exit with no zone change for the sweep to follow.

Phasing needs one more line: the designation rides through the phase-out, the copy does not (the
permanent is treated as though it does not exist, so the sweep removes it), and `phaseInLocked`
makes a fresh copy for a permanent that "phases in prepared".

## Decision 6 — "Enters prepared" is a CR 614.1d replacement

"This creature enters prepared" is `effects.SelfEntersPrepared()`, the `SelfEntersTapped` shape: it
sets `ReplacementEvent.EntersPrepared`, and every battlefield landing — the entry finisher,
`putOntoBattlefieldFromZoneLocked` and the sandbox move — calls `applyEntersPreparedLocked` after
the copy effect and the counters and before `EventETB`. So the permanent is never on the battlefield
unprepared, an ETB trigger finds it prepared, and a Clone copying a preparation card prepares the
prepare spell it copied. "Becomes prepared" from an effect is the `effects.BecomePrepared` primitive.

## Cards

| Card | Shape | Completeness |
|---|---|---|
| Skycoach Conductor // All Aboard | enters prepared; instant copy; the blink (#1306's card) | full |
| Landscape Painter // Vibrant Idea | enters prepared; SORCERY copy keeps sorcery timing | full |
| Encouraging Aviator // Jump | "whenever this creature attacks, it becomes prepared" — a trigger on the stack | full |
| Skycoach Waypoint | "{3}, {T}: target creature becomes prepared" — targets any creature, CR 722.3a's refusals | full |

None of the four is on the card-coverage roadmap's batch lists (the roadmap predates the set); they
are the cheapest preparation cards that each exercise one path the designation can arrive by.

`deck/validate.go`'s `layoutSimplifications` loses its `prepare` banner: the card casts as its
creature half because CR 722.3 says so, and an uncatalogued preparation card that never becomes
prepared is the ordinary `unimplemented` badge's job, not a layout warning.

## Out of scope

- **Enters prepared through a copy effect's entry.** A Clone entering as a copy of a Skycoach
  Conductor has the prepare spell (CR 722.2b) but not the Conductor's "enters prepared"
  replacement, which the pipeline gathers from the entering card before the copy is applied. The
  same timing question every "as a copy, it enters with…" interaction has; no catalogued card
  reaches it.
- **CR 722.3d** — a copy of a prepare **spell** (Twincast on a cast All Aboard) "is also a prepare
  spell". Nothing in the engine asks whether a spell is a prepare spell yet.
- **CR 722.5** — naming a preparation card's alternative name. No name-a-card effect offers
  prepare-spell names today.
- **The spell copies CR 704.5e also covers.** A copy made by `CopySpellForEffect` that is countered
  into a graveyard still lands there as a card ([#1340](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1340)). `PrepareCopy` is deliberately not
  widened to them in this change.

## Consequences

- The Paradigm row's "cast a copy of a card in exile exists nowhere" is half answered: the engine
  can now cast a copy out of exile through the ordinary cast path. Paradigm's own copy is made at a
  different moment (a recurring first-main-phase offer) and still has no shape.
- `Card` grows three fields, all carried by the snapshot and pinned by the drift test.
- The wire grows one field, `CardView.prepared`; the copy itself rides the existing `exile_play`
  stamp, so the client's exile button casts it with no new code, and the designation badge gains a
  third tenant ("PREPARED").

## Test plan

- `game/prepare_test.go` — the copy's characteristics; CR 722.3a's two refusals; the controller-only
  derived permission; the cast (copy on the stack, unprepared at 601.2i, ceases to exist); the sweep
  after the permanent dies; unprepare; a countered copy; phasing in prepared; a snapshot round trip;
  a copy countered into exile is not castable again.
- `cards/effects/prepare_cards_test.go` — one test per card above (two for the Aviator).
- `legal/prepare_moves_test.go` — the enumerator offers the copy's face 1 to the controller and to
  nobody else, and every offered move dispatches.
- `protocol/prepare_view_test.go` — `prepared` on the permanent and `exile_play` with `faces: [1]`
  on the copy, for both seats.
- Client: `designationBadge.render.test.ts` renders "PREPARED".
