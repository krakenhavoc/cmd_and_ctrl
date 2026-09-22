# ADR 0081 — Earthbend: a counted keyword action that animates a target, and a delayed trigger keyed to an object

**Status:** Accepted · 2026-09-21 · S46 — permanents that change what they are
**Issues:** [#1178](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1178) (the keyword), [#1179](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1179) (the caveat this clears), [#343](https://github.com/krakenhavoc/cmd_and_ctrl/issues/343) (the Avatar set)
**Numbering:** swept with the AGENTS.md §4 loop on 2026-09-21 — `git fetch origin`,
then `git log --all --diff-filter=A --name-only -- 'docs/decisions/*.md'` over every
fetched ref, plus a per-branch `git ls-tree docs/decisions/` across all 406 remote
heads. The highest number present anywhere was **0080** (`0080-attack-taxes.md`);
`0081` was free, and the sweep was repeated immediately before the push. `0005`,
`0024`, `0029` and `0030` stay permanently unused.

**Related:** [ADR 0013](0013-replacement-effects.md) §5s (the counted-keyword-action
model this extends), [ADR 0026](0026-delayed-triggers.md) and its 2026-09-18
amendment (`DelayedTrigger.On`, the event condition this keys on),
[ADR 0063](0063-durations-and-control.md) (the duration model and the pin),
[ADR 0067](0067-layer-dependency-ordering.md) (layers), [ADR 0035](0035-until-end-of-turn-effects.md)
(the scoped-static registry).

## Context

Twelve cards in *Avatar: The Last Airbender* print one keyword action
and all twelve print the same reminder text:

> **Earthbend N.** *(Target land you control becomes a 0/0 creature
> with haste that's still a land. Put N +1/+1 counters on it. When it
> dies or is exiled, return it to the battlefield tapped.)*

Bitter Work, Rockalanche, Earth Rumble, Bumi King of Three Trials, Ba
Sing Se, Sandbenders' Storm, Cracked Earth Technique, Earthshape, Dai
Li Indoctrination, Earthbending Lesson, The Cave of Two Lovers, and The
Legend of Kyoshi's chapter II.

#1178 filed it as four gaps rather than one, and that framing is the
first thing this ADR rejects:

1. a one-shot type-add on a **targeted** land — "still a land" plus
   Creature, base P/T 0/0. `StaticForDuration` + `SnapshotAffected`
   compose for it, but every existing caller (`coercive_recruiter.go`,
   `BecomeCreatureUntilEOT`) is about the SOURCE's own controller
   acting on something already resolved, and `BecomeCreatureUntilEOT`
   in particular cannot express a 0/0 at all — its own comment says a
   zero P/T means "leave the printed values alone" and that "a card
   that really wants a 0/0 is not expressible and does not exist";
2. haste, with **no stated duration** — every keyword grant in the
   catalog is `GrantKeywordUntilEOT`;
3. N +1/+1 counters — the existing primitive, with nothing missing;
4. "when it dies or is exiled, return it to the battlefield tapped" —
   nothing in the catalog does this. `ScheduleDelayedTrigger` fires at
   a STEP, and `DelayedTrigger.On` (#663) fires on an EVENT but has
   only ever watched a cast.

The Legend of Kyoshi shipped in #1179 with chapter II reduced to its
second independent clause ("that land becomes an Island in addition to
its other types") and a declared caveat naming this issue.

## Decision 1 — Earthbend is a counted keyword action, and its count means a third thing

`KeywordActionEarthbend` joins proliferate, scry and surveil in the
`KeywordAction` enum, and `Game.EarthbendForEffect(actor, source, land, n)`
opens `RepEventKeywordAction` exactly as `ProliferateForEffect` does.
One arm of `applyResolvedKeywordActionLocked` takes the settled action.

This is ADR 0013 §5s applied unchanged, and §5s's own argument is why
it is right here: the number is the only thing that varies across the
twelve cards, which is precisely the shape a CR 614 window is for.
Nothing printed replaces an earthbend today. Opening the window costs
one `switch` arm now and is the difference between a seam and a rewrite
the day a "whenever you would earthbend, earthbend twice that much
instead" is printed — that card is then a card-side `ReplacementEffect`
narrowing on the action, with no engine change at all.

**The count means a third thing, and §5s already said that is the
rules' own distinction.** Proliferate counts TIMES (the whole action,
CR 701.34), scry and surveil count CARDS, earthbend counts
**+1/+1 COUNTERS**. Collapsing them would make "earthbend 4" take the
action four times, which is four animations and four delayed returns.

**The target rides the tail, not the count.** `keywordActionTail.land`
carries the chosen land across a CR 616 pause. A replacement of this
family rewrites the COUNT and nothing else — "earthbend a different
land instead" is not a shape the rules have — so the field is data the
resume reads, not a second thing a card can touch.

## Decision 2 — Earthbend is the first action in the family that acts at a count of zero

`applyResolvedKeywordActionLocked` used to return early on `n <= 0`,
because a proliferate taken zero TIMES is not taken and a scry of zero
CARDS looks at nothing. Earthbend breaks that: **"earthbend 0" is a
printed instruction** — Rockalanche's "earthbend X, where X is the
number of Forests you control" with no Forest on the battlefield — and
it is not a no-op. The land becomes a 0/0 creature with haste, gets no
counters, dies to the CR 704.5f toughness state-based action the next
time anybody would get priority, and comes back tapped.

So the guard is now `n <= 0 && !ev.KeywordAction.actsAtZeroCount()`,
with one predicate carrying the fact rather than the arm's body
carrying it. A named method rather than an inline `!= earthbend`
because the question ("does the rest of this verb depend on its
count?") is a property of each action and a fifth one will have to
answer it.

`abandonKeywordActionLocked`'s body is extracted as
`runKeywordActionThenLocked`, and the settled earthbend arm runs the
continuation through it. That is one implementation of "the rest of
the sentence" shared by the happened and the replaced-away paths, so
they cannot disagree about whether Earthshape's hexproof sweep ran.

## Decision 3 — The animation is three continuous effects with one timestamp and one duration

`animateEarthbentLandLocked` registers three `ScopedStatic`s against
one snapshotted affected set — `{instance, EnteredBattlefieldAt}`,
CR 611.2c keyed the way every one-shot in the engine is keyed:

| Layer | Effect | Why there |
|---|---|---|
| 4 | add `Creature` | CR 613.1d. Land is **not** touched — "that's still a land" is the reminder text spelling out what adding a type already means, and it is load-bearing: the land keeps its mana ability and still counts for landfall and "lands you control" |
| 7b | base power and toughness ← 0/0 | CR 613.4b: an effect from a resolved spell SETTING a specific value. Not 7a, which is for characteristic-defining abilities; `BecomeCreatureUntilEOT` puts the same clause in the same sub-layer. The +1/+1 counters apply at 7d over the top, which is what makes an earthbend 4 a 4/4 |
| 6 | grant `haste` | CR 613.1f |

**#1178 says "layer 4 + 7a"; it is 4 + 7b**, and the difference is not
cosmetic: 7a would sort before every other "base P/T becomes" effect
and would be claiming this is a characteristic-defining ability, which
it is not.

**Why 7b at all, when a Forest already prints 0/0.** Two reasons. A
land that prints a real body — Dryad Arbor is a 1/1 Land Creature — has
that body REPLACED, so an earthbend 2 on it is a 2/2 and not a 3/3.
And the 7b bucket is what sets `Characteristic.PTDefined` (#690), which
is what makes the animated land's toughness KNOWN and therefore what
lets CR 704.5f kill an earthbend-0. Without it the state-based action
skips the permanent whole and "earthbend 0" silently does nothing.

**One timestamp, one duration, all three.** They are one printed
sentence, so they are one continuous effect for CR 613.7 — the shape
`ExchangeControlForEffect` already uses via `registerScopedStaticLocked`.
That is also what makes "the haste and the animation end together" a
fact about the data rather than a coincidence a test has to re-check.

**The duration is `IndefiniteDuration()` pinned to the object.**
Earthbend states no duration, so CR 611.2a says the effect lasts until
the game ends. CR 400.7 says the permanent that comes back after it
dies is a NEW object, and the effect names the old one — which the
affected set already enforces, since the entry stamp is re-minted on
every entry. `Game.PinnedTo` adds the garbage collection: without it,
three dead registry entries per earthbend would sit in `ScopedStatics`,
and in the snapshot census keeping the game off the full-restore path,
for the rest of the game.

**Haste is NOT until end of turn**, which is the one place every
existing keyword-grant caller would have led us wrong. A land earthbent
on turn three is still hasty on turn nine. It matters less often than
it looks — CR 302.6 runs on continuous CONTROL since the turn began,
not on when the permanent became a creature, so a land that has been
around since last turn could attack the moment it is animated anyway —
but it is exactly what lets a land played THIS turn be earthbent and
swing, which is the play the cards are designed around.

## Decision 4 — The counters go through the PLACEMENT window, and they go last

`AddCounterByForEffect(actor, land, "+1/+1", n)`. The land is already
on the battlefield, so these are placed counters and not entry counters:
Hardened Scales, Doubling Season and Vorinclex all see them, the CR 614
ENTRY pipeline is not involved, and "who put them" is the earthbending
player (CR 120.3d).

They are **last** in `applyEarthbendLocked`, ahead of the printed order,
because the placement is the only part that can PAUSE — a CR 616
ordering prompt between two counter replacements — and
`AddCounterByForEffect` returns nil on that pause with the placement
owed to the resume. Everything the earthbend owes has to be registered
before then, or a paused Doubling Season prompt would leave a land that
is not a creature and has no delayed return. That is the ordering
argument `enterBattlefieldThroughPipelineLocked`'s tail makes, applied
to a verb.

Printed order is not otherwise observable: state-based actions run when
a player would receive priority (CR 704.3) and no player does in the
middle of one resolution, so the 0/0 the animation makes cannot die
before the counters land.

## Decision 5 — The return is a delayed trigger keyed to the OBJECT, with a derived identity

This is the decision with nothing behind it, and the reason this is an
ADR rather than an amendment.

**It is a delayed TRIGGERED ability, not a replacement and not a card
hook.** CR 603.7b: it uses the stack, so every player gets a response
window before the land comes back. It belongs to no permanent, because
the permanent it is about is the thing that just left. And it is not a
replacement effect, which matters: the land really DIES, so Blood
Artist, "whenever a creature you control dies" and a graveyard count
all see it. Only what happens next is added.

**The condition is an EVENT, and the event is `EventLTB` narrowed by
its destination.** `DelayedTrigger.On` + `AppliesTo` (#663) is the
shape; `earthbendReturnMatches` is the predicate — this object, leaving
the battlefield, `NewZone` a graveyard or exile. It fires ONCE and
ceases to exist, which `fireEventDelayedTriggersLocked` gives for free
by removing a matched trigger from the queue before dispatching.

**`NewZone` is the right reading of "dies or is exiled."** It is where
the card ACTUALLY went, after the CR 614 window on the move settled. So
a Rest in Peace-style replacement that exiles the land instead of
letting it die still satisfies the trigger — and satisfies the printed
text, which names both destinations. A bounce or a tuck satisfies
neither.

**Identity is DERIVED from the object, not stored.**
`earthbendReturnTriggerID(land, enteredAt)` is `uuid.NewSHA1` over a
fixed namespace plus the instance ID and the battlefield-entry stamp.
A second earthbend on the same object computes the same ID, finds it
already queued and adds nothing; the same land earthbent again after it
has died and come back computes a different one, because the stamp is
re-minted on every entry (CR 400.7).

A derived ID rather than a new field on `DelayedTrigger`: the queue is
plain data that clones and snapshots by value, and this needs no schema
change to answer "is one already watching this object".

**DECLARED DIVERGENCE, one line wide.** In paper two earthbends really
do make two delayed abilities, both triggering; the first returns the
land and the second finds a card that is no longer where it was left —
a new object on the battlefield — and does nothing. One queue entry is
the same game state with less bookkeeping. The one board where they
differ is a FIRST return that is countered: paper would still have the
second, and this would not. Nothing in the catalog counters a triggered
ability (it is an open seam in `docs/engine-seams.md`), and when one
lands this is the place that has to grow a real per-instance queue.

**The duration is the garbage collector, again.** Indefinite, pinned to
the object. Nothing ends an earthbend return on a turn boundary — the
land can sit there for ten turns and still come back when it dies — but
a land that leaves the battlefield any OTHER way can never satisfy the
trigger again, because the returning card is a new object. The pin
drops the dead entry at `clearExpiredDelayedTriggersLocked`'s next
sweep. This is the first `DelayedTrigger` in the engine with an
explicit non-`UntilEndOfTurn` duration; `ScheduleDelayedTriggerForEffect`
stamps `UntilEndOfTurn` on an event-conditioned trigger that names
none, because every printed one before this said "this turn".

## Decision 6 — One entry ROUTER, because the trigger does not know which zone it is looking in

"Return it to the battlefield tapped" has to work from a graveyard or
from exile, and the engine had one door for each with different
signatures: `ReturnFromExileToBattlefieldForEffect` takes `tapped` and
returns the entered ID, `ReturnFromGraveyardUnderControlForEffect`
takes neither.

`Game.ReturnToBattlefieldForEffect(cardID, controller, tapped)` is the
router over the two, dispatching on where the card actually is.
Anywhere else is `ErrCardNotFound` — the object the effect named is not
where it was left, which is CR 608.2b's posture and the caller's to
swallow. That covers a land somebody reanimated first and a graveyard
that was exiled wholesale.

Not a fourth entry primitive: each of the two has its own CR 400.7
bookkeeping (the exile return re-mints the instance ID, the graveyard
return leaves it and relies on the re-minted entry stamp), and
flattening them would mean choosing one of those behaviours for both.
`ReturnFromGraveyardUnderControlForEffect`'s body is extracted as
`returnFromGraveyardLocked` with the `tapped` parameter, and `tapped`
rides onto the CR 614 event rather than being OR-ed in afterwards, for
the reason `SearchLibrarySpec.TappedOnEntry` gives: a resume reads
`ev.EntersTapped` and has no idea what effect sent the card.

The return happens under the card's **OWNER's** control (CR 400.3 — the
printed text names no controller), so a land somebody stole and killed
goes home.

## Decision 7 — The card side is one clause

`effects.EarthbendTargets()` is the target clause all twelve cards
share ("target land you control"), and `EarthbendFirstTarget(count)` is
the whole body of a card whose only sentence is the keyword. Two
registered cards prove the two shapes and a third clears the caveat:

- **Earthbending Lesson** ({3}{G} Sorcery — Lesson, "Earthbend 4.") —
  the keyword and nothing else.
- **Rockalanche** ({2}{G} Sorcery — Lesson, "Earthbend X, where X is
  the number of Forests you control." + Flashback {5}{G}) — a count
  read at RESOLUTION, the earthbend-0 board, and the keyword composing
  with an existing one.
- **The Legend of Kyoshi** chapter II — the keyword beside a second
  independent clause about the same land ("that land becomes an Island
  in addition to its other types"), registered as two continuous
  effects because they are two sentences with the same pin. The face
  goes from `caveats` to **full** and #1179's caveat is deleted.

A card with a second clause about the same land calls `Earthbend`
directly and keeps its own target read, because it needs the ID twice.

## Out of scope (explicit deferrals)

- **The other nine earthbend cards.** Each is now one clause, but each
  also needs something else this branch does not build: Bitter Work and
  Ba Sing Se want **exhaust** ("activate each exhaust ability only
  once"), which nothing in the engine has — filed as
  [#1181](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1181), where
  the shape is named; Earth Rumble wants a
  reflexive "when you do" trigger around the verb; Cracked Earth
  Technique wants two target clauses; Sandbenders' Storm and Dai Li
  Indoctrination are modal with per-mode targets; Earthshape reads the
  animated land's power afterwards. They are ordinary card work now.
- **A per-instance return queue.** See Decision 5's declared
  divergence. It costs a field on `DelayedTrigger` and buys nothing
  until a "counter target triggered ability" card exists.
- **A printed earthbend replacement.** The window is open and
  `keywordCountReplacement`-shaped cards work; none is printed.
- **Earthbend on a permanent that is not a land.** The keyword's own
  target clause forbids it, and no card widens it.

## Consequences

- `game.EarthbendForEffect` is the only door. A card cannot get the
  animation without the delayed return, or the counters without the
  haste, which is the whole reason the four parts are one verb: twelve
  cards is twelve chances to get the layer, the duration or the
  trigger's key wrong.
- `applyResolvedKeywordActionLocked`'s zero guard is now per-action.
  A fifth counted keyword action has to answer `actsAtZeroCount`.
- The delayed-trigger queue can now hold an entry with an indefinite
  duration. `clearExpiredDelayedTriggersLocked` already handles it
  (the pin is checked first for every kind), and the snapshot census
  is unchanged — a `DelayedTrigger` was already two closures.
- `ReturnToBattlefieldForEffect` is available to the next card that
  wants "return it to the battlefield" without knowing the zone.

## Test plan

Engine (`server/internal/game/earthbend_test.go`): the land is a 0/0
Land Creature with haste and N counters; a printed body is overwritten,
not added to; earthbend 0 animates, dies and returns; a land played
this turn can attack once earthbent and cannot without the haste; the
three statics share one timestamp and one duration and are swept
together; Hardened Scales sees the placement; dying returns it tapped
as a new object with the animation and the counters gone; exile returns
it; a CR 614 replacement that exiles instead of killing still returns
it; a bounce does not, and the pin sweeps the trigger; a second
earthbend adds counters and queues no second return, while earthbending
the RETURNED land does; the keyword-action window rewrites N; a
proliferate doubler does not see an earthbend; a cancelled earthbend
does nothing at all; earthbending a land that has left is a no-op.

Cards (`server/internal/cards/effects/earthbend_cards_test.go`):
Earthbending Lesson animates and returns; Rockalanche counts Forests at
resolution (a Stomping Ground counts, an opponent's Forest does not)
and its zero case animates-and-returns; its flashback is declared;
Kyoshi's chapter II does both clauses and the face carries no caveat.
