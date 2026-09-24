# ADR 0079 — Transforming a permanent: turning a card over is not a zone change (CR 701.27, CR 712)

**Status:** Accepted · 2026-09-19 · S46 — Permanents that change what they are
**Issues:** [#343](https://github.com/krakenhavoc/cmd_and_ctrl/issues/343) (tracker —
"Aang swift savior transform missing"), driven by
[#1112](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1112) (Storm the Vault,
Fable of the Mirror-Breaker, Brass's Tunnel-Grinder)
**Numbering:** originally written as 0078. On 2026-09-19, after `git fetch --all
--prune`, all 385 remote branch heads and all 259 local branch heads were listed with
`git ls-tree --name-only <ref> docs/decisions/`, every commit reachable from any ref
was swept with `git log --all --name-only -- docs/decisions/`, and the ~90 sibling
worktrees on this machine were listed for uncommitted ADR files. The highest number
present anywhere was **0077** (`0077-opponent-board-summary.md`), and nothing held a
`0078-*` file.

`docs/adr-0078-token-art` was pushed four minutes before this branch and took the
same number — the collision AGENTS.md §4 describes, which no sweep can prevent
because the other branch did not exist when this one looked. That PR was opened
first, so per §4 it keeps 0078 and this file moved to **0079**; the filename and the
H1 above were both changed, and nothing links to this ADR yet. 0005, 0024, 0029 and
0030 stay permanently unused per the same section, and 0078 is now spoken for rather
than free.

**Related:** [ADR 0034](0034-multi-face-cards.md) (the face model this builds the verb
on — `Card.Faces`, `Card.ActiveFace`, `Card.SetFace`, and the CR 712.8a reset in
`MoveCard`), [ADR 0039](0039-layer-4-authoritative.md) (why a printed-characteristic
change has to invalidate the layer cache rather than patch `Characteristic`),
[ADR 0018](0018-triggers-on-the-stack.md) (the harvester that makes `EventTransform`
writable with no new machinery), [ADR 0010](0010-card-effect-catalog.md) (the `Spec`
slots a transforming card declares through)

**A note on rule numbers.** Transform is **CR 701.27** in the pinned August 7, 2026
edition, not 701.28 — 701.28 is *Convert*, the Omen-card sibling that follows
701.27a–f by reference. The seam registry and #1112's write-up both say 701.28,
which was correct in a pre-2025 edition and moved in the re-sort AGENTS.md §6 warns
about. Saga cards are **CR 714** here (715 is Adventurer cards).

---

## Context

The seam registry has called this the largest open engine seam since the
2026-09-17 audit, and the audit's figure is the reason: **47 cards** in the sampled
5,084 have transform as their *only* blocker and 72 have it as *a* blocker. Every
one of them is stuck behind a single missing verb.

The face model itself is not missing. [ADR 0034](0034-multi-face-cards.md) already
built:

- `Card.Faces []Face` and `Card.ActiveFace int`, with the flat printed fields
  (`Name`, `TypeLine`, `ManaCost`, `Colors`, `Power`, `Toughness`,
  `VariableToughness`, `StartingLoyalty`, `StartingDefense`) as the
  *materialisation* of `Faces[ActiveFace]` rather than a copy of Scryfall's
  top-level record;
- `Card.SetFace(i)`, the one writer of `ActiveFace`, which re-materialises those
  fields — and whose doc comment already says a battlefield caller owes a
  layer-staleness mark;
- `CatalogKey`, which appends `#<face>` for a non-zero active face, so a back face
  registered as `"<oracle_id>#1"` gets its triggers, statics, replacements, mana
  abilities and activated abilities **for free** the instant the face flips;
- `MoveCard`'s CR 712.8a reset — a card whose destination is neither the battlefield
  nor the stack is put back to face 0 — so a transformed permanent that dies is
  already a front-face card in the graveyard;
- the wire: `CardView.layout`, `CardView.faces[]` (name, type line, mana cost,
  oracle text, P/T and a per-face `/cards/{id}/image?face=N` URL) and
  `CardView.active_face`, with `client/src/lib/cardImage.ts` defaulting its face
  argument to `card.active_face`;
- the snapshot: `cardSnapshot.Layout` / `.Faces` / `.activeFace`, all three
  classified `carried` in `snapshot_drift_test.go`;
- `cloneCard`'s deep copy of `Faces`, whose comment says in so many words that
  "a later transform effect that rewrites a face in place would make it
  load-bearing".

What is missing is one verb. Nothing in the engine turns a permanent over. The only
route onto a back face today is S32's Siege grant, which is a **cast** of the back
face out of exile (`CastPermission.Faces`, `faceOnResolve`) — a different thing that
happens to land on the same face.

Two shapes of card need two different verbs, and conflating them is the trap this
ADR exists to avoid:

| Printed text | What happens |
|---|---|
| "transform Storm the Vault" | the permanent turns over **in place** — same object |
| "Exile this Saga, then return it to the battlefield transformed" | two zone changes — a **new object** |

---

## Decisions

### 1. Transform is not a zone change, and the verb never touches `MoveCard`

CR 712.18 is one sentence and it is the whole design:

> When a double-faced permanent transforms or converts, it doesn't become a new
> object. Any effects that applied to that permanent will continue to apply to it.

So `TransformPermanentForEffect` mutates the card **in its slot on the
battlefield** and touches nothing that CR 400.7 would take away.

**Preserved, because the object never ended:**

`InstanceID` · `ObjectEpoch` · `Controller` and `BaseController` · `Owner` ·
`Tapped` · all `Counters` (including `+1/+1` and lore) · marked damage and the
deathtouch flag · `RegenerationShields` · `AttachedTo` / `AttachedAt` (an Equipment
stays equipped, an Aura stays attached) · `EnteredBattlefieldAt` — the CR 613.7
timestamp — and `SummonedThisTurn`, so a creature that was summoning sick stays
summoning sick and one that was not does not become so · `NamedTribe`,
`ChosenColor`, `ChosenPlayer` · `Provenance` · `ClassLevel` / `Solved` ·
`ProtectorPlayerID` · `GoadedBy` · `AttackingTarget` / `BlockingTarget`, so a
creature that transforms mid-combat stays in combat · every `Game`-side per-object
registry keyed by instance ID (`LoyaltyActivatedThisTurn`, `announcedAttacks`,
`announcedBlocks`, `blockedAttackers`, the turn tally's per-object cells, the
CR 726 loop breaker's run count) · every continuous effect and duration scoped to
the permanent.

**Not preserved, and each has a rule behind it:**

- **The printed characteristics.** That is the point: `Name`, `TypeLine`,
  `ManaCost`, `Colors`, `Power`, `Toughness`, `VariableToughness`,
  `StartingLoyalty` and `StartingDefense` are all re-materialised from the new face
  (CR 712.8d–e). A Storm the Vault that transforms *stops being an enchantment and
  starts being a land*, with everything that follows.
- **The cached `Characteristic`.** `Card.effective` is set to nil and
  `g.layerVersion` is bumped — see decision 3.
- **The catalog entry.** `CatalogKey` now answers `"<oracle_id>#1"`, so the front
  face's triggers, statics and abilities stop applying and the back face's start.
  Nothing had to be written for this; it is what the `#N` key was for.
- **Nothing else.** In particular the verb does **not** mint a new `InstanceID`,
  does **not** bump `ObjectEpoch`, does **not** re-stamp `EnteredBattlefieldAt`
  and does **not** emit `EventZoneMove`, `EventETB` or `EventLTB`. A card whose
  back face has an "enters the battlefield" trigger does not get it from
  transforming, which is correct: nothing entered.

One consequence worth stating because it is easy to get backwards: **a planeswalker
back face does not get its starting loyalty stamped.** CR 306.5b's stamp is an
entry rule and nothing entered; a permanent that transforms into a planeswalker
with no loyalty counters dies to the CR 704.5i state-based action, which is what
the rules say and what paper does. The printed cards that transform into
planeswalkers all put the loyalty on with a replacement or an "as this transforms"
clause, and those are card text, not engine behaviour.

### 2. The verb: `Game.TransformPermanentForEffect`, and one `effects` primitive

`server/internal/game/transform.go`:

```go
// CanTransform reports whether this permanent is one CR 712.9 allows to
// turn over at all.
func CanTransform(c Card) bool

// TransformPermanentForEffect turns the named battlefield permanent
// over (CR 701.27a). Caller must hold g.mu.
func (g *Game) TransformPermanentForEffect(cardID uuid.UUID) error
```

Named and shaped like the rest of `effect_api.go`: a `*Game` method, `…ForEffect`,
under the lock the caller already holds, taking an instance ID and returning an
error. It lives in its own file rather than in `effect_api.go` because
`CanTransform`, the CR 712.10 instant/sorcery guard and the "exile then return
transformed" sibling belong beside it, and `effect_api.go` is already 3,000 lines.

Card-side, in `server/internal/cards/effects/transform.go`:

```go
Transform{Target: id}.Apply(ctx)   // "transform target Vehicle"
TransformThis{}.Apply(ctx)         // "transform Storm the Vault" — the common case
```

`TransformThis` reads the source off the `Context` rather than capturing a
`*game.Card`, for the reason every trigger `Effect` does (undo restores a clone).

**Which face?** "Turn it over so that its other face is up" (CR 701.27a). With
exactly two faces that is `1 - ActiveFace`, and it is symmetric: a card that
transforms twice is back where it started. No printed card in the sampled pool has
more than two faces, and `CanTransform` refuses anything that does, so the verb
never has to guess.

### 3. Transforming emits `EventTransform`, and the layer version bumps

Two reasons, and the second is not optional.

**The template exists.** "Whenever this creature transforms into <back face>"
(Tovolar, Dire Overlord; the Innistrad werewolf cycle; Kefka, Court Mage) and
"whenever a permanent you control transforms" are printed. CR 701.27e even defines
when such a trigger fires. Adding the kind *now* costs one constant; retrofitting it
later means finding every mutation site again.

`EventTransform EventKind = "transform"`, carrying:

| field | value |
|---|---|
| `CardID` | the permanent that turned over |
| `Actor` | its controller |
| `Amount` | the face index it turned **to** |
| `Label` | the name of the face it turned **from** |

No new trigger constructor ships with it. The harvester already watches any
`EventKind` a `TriggeredAbility` names, so `On(game.EventTransform, Self, label,
effect)` is a working "whenever this transforms" today; a named
`WhenThisTransforms` shape goes into `triggers_common.go` in the same PR as the
first card that prints one, per the append-only rule.

**The layer cache is stale the instant the face lands.** A face change is a
*printed-value* change — the base P/T, the type line, the colours and the ability
list all move at once — and [ADR 0039](0039-layer-4-authoritative.md) made the
layered `Characteristic` the authority for every one of those. `Card.Effective()`
returns `*c.effective` verbatim when the cache is warm, so without invalidation a
transformed Storm the Vault would keep reporting "Legendary Enchantment" to the
type predicate, the combat engine and the wire until some unrelated event happened
to bump the version.

So `TransformPermanentForEffect` does both: it nils `Card.effective` on the card it
just turned over (the same belt-and-braces `stampBattlefieldEntryLocked` does) and
`layerVersionBump.OnEvent` gains an `EventTransform` arm. The bump goes in the
listener rather than inline for the reason every other arm is there: the listener is
the one place that lists what invalidates the layer engine, and an inline bump is a
line the next reader of that file will not find.

**The public log gets a line.** A new `EventKind` must either be narrated by
`projectEvent` or listed in `silentEventKinds` with a reason
(`TestEveryEventKindIsNarratedOrDeliberatelySilent`, #984). A transform is narrated:
`LogTransform`, rendering "Storm the Vault transformed into Vault of Catlacan". It
is the clearest case there is of a thing a player announces out loud — the card
physically turns over — and the board-state-is-visible silence does not apply,
because a reader scrolling back wants to know *when* it happened.

### 4. A permanent that can't transform does nothing, and that is not an error

CR 701.27c and CR 712.9: an instruction to transform something that isn't a
double-faced card or a double-faced token **does nothing**. CR 701.27d and
CR 712.10: so does an instruction whose target face would be an instant or sorcery.
CR 712.4c: meld cards cannot transform. CR 712.15a: a face-down permanent cannot.

`TransformPermanentForEffect` returns `nil` — not an error — for every one of them,
and emits no event. The predicate is:

```go
func CanTransform(c Card) bool {
    if c.FaceDownIsPermanent() { return false }              // CR 712.15a
    if len(c.Faces) != 2 { return false }                    // CR 712.9
    switch c.Layout {
    case LayoutTransform, LayoutModalDFC:                    // CR 712.2, 712.3
    default: return false                                    // adventure, split, prepare, meld, normal
    }
    other := c.Faces[1-c.ActiveFace]
    return !faceIsInstantOrSorcery(other)                    // CR 701.27d / 712.10
}
```

The layout allowlist is the load-bearing line. An `adventure` card has two `Faces`
and a `split` card has two `Faces`, and neither is a double-faced card — turning
Foulmire Knight into Profane Insight on the battlefield is not a thing that can
happen. Keying on `len(Faces) == 2` alone would have made it one.

`ErrCardNotFound` is still returned for an ID that names no battlefield permanent,
because that is a caller bug rather than a rules outcome.

`nil` and no event is deliberately **not** `EventEffectError`: the catalog soak
(`AISEAT_CATALOG_GAMES`) fails on any `EventEffectError`, and "Vandalblast
transformed nothing" is a legal resolution, not a card that threw.

### 5. "Exile it, then return it transformed" is a **second, separate verb**

Fable of the Mirror-Breaker's chapter III reads *"Exile this Saga, then return it to
the battlefield transformed under your control."* That is two zone changes. CR 400.7
applies in full: the permanent that comes back is a **new object** with a new
`InstanceID`, no counters, no damage, no attachments, a fresh CR 613.7 timestamp,
summoning sick again, and it triggers every "enters the battlefield" ability it has
on its new face.

```go
// ExileAndReturnTransformedForEffect — CR 712.14a. Caller must hold g.mu.
func (g *Game) ExileAndReturnTransformedForEffect(cardID, controller uuid.UUID) error
```

It is built out of the two moves that already exist rather than out of a bespoke
route:

1. `ExileCardThenForEffect` — the shared exit primitive, so a commander that would
   be exiled is offered the command zone (CR 903.9) and the continuation waits for
   the answer instead of firing into a paused move.
2. In the continuation, set the face on the card **while it is still in exile**,
   then `ReturnFromExileToBattlefieldForEffect`, which mints the new object and runs
   the CR 614 entry pipeline.

The "set the face in the source zone first" step is the same move `CastSpell` makes
through `setFaceInZoneLocked`, and for the same reason: the CR 614 replacement
pipeline resolves the entering card by ID out of its source zone, so a
self-replacement on the back face (an "enters tapped" clause) has to be findable
under the back face's catalog key before the pipeline runs. It puts a card in exile
momentarily on its back face, which CR 712.8a says should not happen; the window is
inside one locked mutation, nothing observes it, and the alternative — threading a
face through `ReplacementEvent` and the resumable entry tail — is a much larger
change for a state no viewer can see. Recorded here so the next reader knows it was
a choice.

**If the exile is replaced away** (the commander took the command zone, a
replacement redirected it), the continuation is handed `exiled == false` and returns
without returning anything. The Saga stays where the replacement put it. That is
right: the printed sentence's second half is conditional on its first.

Keeping the two verbs apart is the whole of decision 5. A single
"transform, maybe by blinking" verb would have given Fable's chapter III a Saga that
kept its lore counters and its timestamp, and given Storm the Vault a fresh object
that re-triggered its own ETBs. Both are wrong, in opposite directions.

### 6. Transforming Sagas need **no** engine change, and that is worth writing down

The worry was CR 714.4's state-based action:

> If the number of lore counters on a Saga permanent **with one or more chapter
> abilities** is greater than or equal to its final chapter number, and it isn't the
> source of a chapter ability that has triggered but not yet left the stack, that
> Saga's controller sacrifices it.

Both transforming shapes fall out of the existing implementation
(`sagasReadyToSacrificeLocked`, `server/internal/game/sagas.go`) with nothing added:

- **`IsSaga` reads the *effective* subtypes.** A permanent that transformed in place
  into "Enchantment Creature — Goblin Shaman" is not a Saga, so the SBA skips it
  entirely. Its lore counters stay on it (CR 712.18 — same object) and nothing
  reads them, which is exactly paper.
- **`SagaFinalChapter` derives from the declared chapter triggers**, which come from
  `CatalogKey` — now the back face's key — and the back face declares none. `final`
  is 0, and the SBA already declines a Saga whose final chapter is 0.
- **The stack clause holds the door open in the meantime.** `sagaHasChapterOnStackLocked`
  keeps the SBA off a Saga while its own chapter ability is still resolving, so
  chapter III gets to *run* — and by the time it has run, the Saga has either
  transformed (not a Saga any more) or exiled itself (not on the battlefield).

So: **a transforming Saga is never sacrificed, because by the time CR 714.4 is
checked there is nothing there that satisfies it.** No guard, no flag, no
"transformed this turn" bit. The one thing a card author must not do is write
chapter III as an in-place transform *and* leave a chapter ability declared on the
back face — that would give the back face a final chapter number and the lore
counters are still on it. No printed card does this and `CanTransform` cannot know
about it, so it is a review note rather than a check.

### 7. The wire needs nothing, and neither do snapshots, clone or undo

This is the happy consequence of ADR 0034 having built the representation first.

- **View.** `viewOfCard` already publishes `layout`, `active_face` and the full
  `faces[]` array with a per-face image URL, on every snapshot. A transform is a
  mutation of `ActiveFace` on a battlefield card, so the next broadcast carries it
  to every viewer, and `cardImageURL` already defaults its face argument to
  `card.active_face`. **No protocol change, no client change.** (`client/src/lib/faces.ts`'s
  `needsFacePicker` correctly excludes `transform` — the picker is an *announce-time*
  choice and a transform is not one.)
- **Redaction.** `FilterViewFor` clears `faces` / `layout` / `active_face` for a
  non-knower, unchanged. A battlefield permanent is public (CR 400.2), so every seat
  sees the new face.
- **Snapshot.** `Layout`, `Faces` and `ActiveFace` are already `carried` in
  `cardSnapshot`, `snapshotCard` and `restoreCard`, and already classified in
  `snapshot_drift_test.go`'s `cardFields` plan. **No new per-instance field is
  added to `game.Card` by this work**, which is the deliberate part: a "has
  transformed" flag would need a plan entry, a probe in `snapshot_carried_test.go`,
  and a decision about what its zero value means for every snapshot written before
  today. There is nothing to store — "this permanent is transformed" is
  `ActiveFace != 0` (CR 701.27g reads exactly that way), and it is already on disk.
- **Clone / undo.** `cloneCard` deep-copies `Faces` per face including each face's
  `Colors`, and `ActiveFace` rides the value copy. The verb only calls `SetFace`,
  which writes `ActiveFace` and the flat scalars; it never mutates a `Face` entry in
  place, so the existing deep copy stays sufficient and the aliasing its comment
  warns about cannot happen.

### 8. CR 712.8e: a nonmodal back face's mana value comes from the **front** face

> While a nonmodal double-faced permanent has its back face up, it has only the
> characteristics of its back face. **However, its mana value is calculated using the
> mana cost of its front face.**

`Card.ParsedManaValue` reads `c.ManaCost`, which after a transform is the back
face's printed cost — and a `transform` back face prints no cost at all, so
`ParseCost("")` succeeds and every transformed permanent in the game would have mana
value 0. Reflection of Kiki-Jiki is mana value 3, not 0, and "destroy target
permanent with mana value 3 or less" has to agree.

So `ParsedManaValue` gains one branch: for `Layout == LayoutTransform` with
`ActiveFace != 0`, read `Faces[0].ManaCost`. **Only for `transform`** — CR 712.8f
gives a modal DFC each face's own characteristics, cost included, and Sea Gate,
Reborn really is mana value 0.

It ships here rather than as its own fix because it is unobservable until something
can transform, and because the branch belongs next to the reason for it.

---

## Out of scope, stated explicitly

Each of these stays on the seam registry's transform row after this ADR lands.

1. **Incubator tokens.** A double-faced *token* (CR 111.9, `//` Phyrexian) needs
   `game.Card` to carry two faces on an object with no oracle ID, which is the same
   gap as "Triggered and static abilities on non-copy tokens" (#521) and should be
   closed with it, not here. `CanTransform` refuses one today because a token has
   no `Layout`.

   That neighbouring gap bites one of this ADR's own proof cards, which is worth
   naming: Fable of the Mirror-Breaker's chapter I makes *"a 2/2 red Goblin Shaman
   creature token with 'Whenever this creature attacks, create a Treasure token'"*.
   `TriggersForCard` returns nil for an empty catalog key and a plain token has one,
   so the token ships without its attack trigger and Fable ships
   `CompletenessCaveats`. It is weaker than printed, which is the only direction
   allowed (#259). The `GrantedAbilities` composite key would technically carry a
   trigger onto a minted token, but no card exercises that path and inventing the
   first use of it inside a transform PR is how a seam gets closed by accident and
   half-tested; it belongs to #521.
2. **Meld** (CR 701.42, 712.4–5). Two cards becoming one object is a different
   representation problem — `Card` is one card — and CR 712.4c explicitly says meld
   cards cannot transform, so the verb is already correct about them.
3. **Day and night** (CR 730). The designation, its turn-based transitions and the
   daybound/nightbound keywords are a turn-machinery change, not a transform one.
   The werewolves wait on it, not on this.
4. **The Siege grant's existing path is untouched.** `CastPermission.Faces`,
   `faceForCastLocked` and `faceOnResolve` keep doing exactly what S32 built. A cast
   of a back face and a transform of a permanent reach the same face by different
   routes and neither now goes through the other.
5. **"Enters the battlefield transformed" from a zone other than the stack**
   (CR 712.14a) beyond the one use in decision 5. `ReturnFromExileToBattlefieldForEffect`
   is the only entry site given a face; a search, a reanimation or a
   put-from-hand that wanted one would need the same two lines, and gets them when a
   card asks.
6. **CR 701.27f** — "the permanent transforms only if it hasn't transformed since
   the ability was put onto the stack". This is the anti-loop rule for a permanent
   with a transform ability on both faces. Nothing tracks "has transformed since"
   and nothing needs to yet; ADR 0055's loop breaker catches the runaway case at a
   coarser grain. Written down because the day a card wants it, the state it needs
   is a stack-item stamp, not a card field.
7. **"As this permanent transforms …"** (CR 712.20). No `AsTransforms` slot on
   `Spec`; the printed cards that use it are all day/night or meld.
   **Closed by the 2026-09-24 amendment below** (#1574): Sephiroth, One-Winged
   Angel prints one, and `Spec.AsTransformsInto` now exists.
8. **A named `WhenThisTransforms` trigger constructor.** The event is emitted and
   harvestable; the constructor lands with the first card that prints the trigger,
   per the append-only rule for `triggers_common.go`.
9. **Brass's Tunnel-Grinder // Tecutlan, the Searing Rift**, the third #1112 card on
   this row. Its front face needs **discover** (CR 701.57) and its back face needs
   **descend** — neither exists — so the transform verb unblocks only a third of it.
   Recorded on the seam registry against both of those, not against this row.

---

## Consequences

**What this unblocks.** Every card whose only blocker was the verb: the eight named
on the seam row (Legion's Landing, Treasure Map, Growing Rites of Itlimoc, Dowsing
Dagger, Kefka Court Mage, Search for Azcanta, Edgar Charmed Groom, Vincent
Valentine), the two #1112 cards this ADR is proved with, Ojer Axonil (claimed by
#1107), and Aang, Swift Savior — the card #343 was filed about.

**What it costs.** One new file in `internal/game` (~120 lines with the doc
comments), one new `EventKind` with its listener arm and its log arm, one new
`LogKind`, one branch in `ParsedManaValue`, and one card-side primitive file. No new
`Card` field, no protocol change, no client change, no snapshot change.

**The thing most likely to be got wrong later.** Somebody will reach for
`TransformPermanentForEffect` to implement "enters the battlefield transformed" or
"put it onto the battlefield transformed", and it will *appear* to work: the
permanent enters front-up, then turns over. It is wrong in two observable ways — the
front face's ETB triggers fire, and the back face's do not — and it is exactly the
"stronger or weaker than printed for no stated reason" failure #259 is about. Those
are entry-time face choices and belong with decision 5's mechanism, not with the
in-place verb.

---

## Amendment 2026-09-24 (#1574): "As this permanent transforms into …" is a face hook the verb runs

Issue [#1574](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1574),
found by slice E2 of the Edea deck ([#1565](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1565),
PR [#1577](https://github.com/krakenhavoc/cmd_and_ctrl/pull/1577)). Out of
scope item 7 above said no card needed this slot. Sephiroth, One-Winged Angel
needs it:

> Super Nova — As this creature transforms into Sephiroth, One-Winged Angel,
> you get an emblem with "Whenever a creature dies, target opponent loses 1
> life and you gain 1 life."

E2 made the emblem in the front face's own transform clause, the fourth drain,
so a Sephiroth turned over by anything else (Moonmist transforms every Human,
and the front face is a Human) made no emblem. The clause belongs to the back
FACE, not to the effect that turns the card over.

### Decision 9. `CardDef.AsTransformsInto`, run by `TransformPermanentForEffect` for the face now up

- **The declaration** is `Spec.AsTransformsInto func(card *game.Card, ctx
  *Context) error`, on the Spec of the face being turned TO. For a back face
  that is its `"<oracle_id>#1"` entry. It has AsEnters's contract: off the
  stack, `ctx.Item` nil, controller read off the card. Both slots share one
  live-card wrapper in `effects/carddef.go` (`liveCardHook`).
- **The engine runs it in the verb**, `runAsTransformsIntoLocked`
  (`game/transform.go`), right after `EventTransform`. It is the one in-place
  transform in the engine, so every card that transforms a permanent (its own
  clause, `Transform{Target}`, `TransformThis`, and anything later) reaches it.
  After the event rather than before it, for the reason `AsEnters` runs after
  `EventETB`: the event bumps the layer version, and the lookup has to see the
  new face.
- **It is a static ability, not a trigger.** Nothing goes on the stack and
  nobody gets a response window. A "Whenever this transforms" trigger already
  works through `EventTransform` (decision 3) and is a different sentence.
- **The key is `CatalogAbilityKey`, read after `RecomputeLayersIfStaleLocked`.**
  CR 712.18 keeps every effect that applied to the permanent applying after it
  turns over, so a Darksteel Mutation that removed its abilities has removed
  the new face's clause too. The `AsEnters` hook reads `CatalogKey` instead,
  because at entry there is no layered characteristic yet. A transforming
  permanent is on the battlefield the whole time, so it has one.
- **An error is published as `EventEffectError`**, as `fireETBHookLocked` does
  for `AsEnters`. The transform has already happened and cannot be unwound.

### What does not run it

`ExileAndReturnTransformedForEffect` (decision 5). That permanent ENTERS on its
back face, and a permanent that enters transformed did not transform. It is a
new object that was never on its front face, so an "as this transforms" clause
has nothing to apply to. A back face that wants something as it enters
declares `AsEnters`, as any other face does.

### Snapshots and undo

Nothing is stored. The hook is catalog data found by key, and what it does is
ordinary state (Sephiroth's emblem is a card in the command zone, carried as
every emblem is), so `Clone` / `RestoreFrom` and the snapshot need no change.

### Cards

Sephiroth, One-Winged Angel comes off its caveat and ships `full`. Its front
face's fourth-drain clause now just transforms, and the emblem comes from the
back face's hook. The tests are in
`server/internal/cards/effects/gogo_sephiroth_hooks_test.go`.
