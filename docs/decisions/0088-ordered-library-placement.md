# ADR 0088 — Ordered library placement: one prompt for "put them on top / on the bottom in any order"

**Status:** Accepted · 2026-09-23 · S45 — modal spells, multi-target clauses and copy effects
**Issues:** [#996](https://github.com/krakenhavoc/cmd_and_ctrl/issues/996) (the library-top row's tracker)
**Trackers:** [#888](https://github.com/krakenhavoc/cmd_and_ctrl/issues/888) — S45, and
[#879](https://github.com/krakenhavoc/cmd_and_ctrl/issues/879) — S36, "tables that wedge" (a new
prompt kind is a new way to owe the table an answer)
**Numbering:** swept with the AGENTS.md §4 loop on 2026-09-23 — `git fetch origin`, then
`git ls-tree --name-only <ref> docs/decisions/` over every remote head. The highest number present
anywhere is **0087** (`0087-amass.md`); `0088` appears on no branch, no open issue reserves it, and
it was reserved on #996 when the issue was claimed. `0005`, `0024`, `0029` and `0030` stay
permanently unused per AGENTS.md §4.

**Related:** [ADR 0013](0013-replacement-effects.md) (the CR 614 window every library arrival goes
through, and CR 903.9's offer on the way), [ADR 0060](0060-leaving-the-game.md) (what happens to a
prompt whose chooser leaves), [ADR 0066](0066-granted-cast-and-play-permissions.md) (CR 401.5
visibility kept on the POSITION rather than the card — the rule this ADR's knower decision has to
agree with).

---

## Context

The roadmap's **"Library-top placement and ordered look (Brainstorm / tutors / Ponder)"** row is
rank 7 / 9 / 7 across the three 2000-card passes — 45 / 31 / 37 sole unlocks — and #996 is its
tracker. By September most of the row had shipped piecemeal: scry, surveil and "look at the top N,
then put them back in any order" share one prompt body (`lookAtTopForEffect`); #744 gave the engine
a keyed random-order bottom; #952 the take-to-hand door; #783 a depth on the tuck route; #989 and
#1035 CR 401.5's standing visibility. What was left is the **placement** half, and #996's own list
is accurate:

1. **An ORDERED bottom.** "Put the rest on the bottom of your library **in any order**" is printed
   on at least forty Commander-legal cards (Impulse, Dig Through Time, Stock Up, Anticipate, Augur
   of Bolas, Goblin Ringleader, Sylvan Messenger, Garruk Caller of Beasts, Collected Company…). The
   engine had only the RANDOM bottom, so Goblin Ringleader shipped with a caveat and the rest did
   not ship.
2. **Brainstorm's "in any order".** Brainstorm was in the catalog as `full`, and its put-back was
   ordered by *pick order*: the `choose_cards` prompt's first pick went on top. That is a convention
   in a question string, not a decision the client ever shows the player — the card grid returns a
   set, and whether its iteration order is click order is an accident of the client's `Set`.
3. **A depth placement** other than top or bottom exists on the route (`zoneRoute.Depth`) but has no
   card-side sentence helper, and the search path — "search your library for a card, then shuffle
   and put that card **third from the top**" (Long-Term Plans) — has no depth at all.
4. **"Its owner puts it on their choice of the top or bottom of their library."** Aetherspouts is in
   the catalog, and expresses the choice by tucking every attacker to the top and then calling
   **Scry** for each owner. That is a keyword action the card never prints: it emits `EventScry`, so
   every "whenever you scry" payoff triggers off an Aetherspouts, and it opens the CR 614
   keyword-action window, so an "if you would scry N, scry N+1 instead" reaches into the library
   and looks at a card nobody tucked.

The rule every one of these is an instance of is **CR 401.4**:

> *If an effect puts two or more cards in a specific position in a library at the same time, the
> owner of those cards may arrange them in any order. That library's owner doesn't reveal the order
> in which the cards go into the library.*

"In any order" on a card is that rule with the chooser named explicitly (usually "you").

---

## Decision 1 — One prompt kind, `put_in_library`, with a placement

`PendingChoicePutInLibrary` (`"put_in_library"`) is "put these cards into their owners' libraries,
in the order you choose". It carries the cards in `ScryCards` and a **placement**:

| `LibraryPlacement` | printed as | answer |
|---|---|---|
| `top` | "put them back on top in any order", "put two cards from your hand on top of your library in any order" | `{top_order}` |
| `bottom` | "put the rest on the bottom of your library in any order" | `{bottom}` |
| `top_or_bottom` | "its owner puts it on their choice of the top or bottom of their library" | `{top_order, bottom}` |

Both lists are **top-first** — the first entry of `bottom` is the one nearest the top of the pile
that goes under the library, and the last is the bottom card — and every card must appear in
exactly one list. That is scry's answer, and deliberately so: `top_order` and `bottom` already mean
exactly this on a `scry`, the engine already applied a scry's `bottom` in the order given, and the
dispatcher already routes the family by kind rather than by which keys are present. A lane the
placement does not open must be empty; an answer that puts a card in the other lane is refused
rather than silently moved.

**A fourth member of the scry family, not a flag on it.** `IsLookAtTopKind` now includes the kind,
so the view projects its cards as `options[]`, the dispatcher routes `top_order` / `bottom` to its
resolver, and the client opens the same dialog. It is a separate kind rather than a
`look_at_top` with a bottom lane because it is not a look at the top of a library: the cards may be
in a hand (Brainstorm), on the battlefield or the stack (Aetherspouts, Hinder), or already in a
library but not at its top. `look_at_top` promises the chooser's own library top, and its resolver
checks for it.

**Why not a `choose_cards` whose pick order is the answer (Brainstorm's old shape).** The pick and
the order are two decisions, and folding them makes one of them invisible. Brainstorm now asks
twice: which two cards (a `choose_cards` over the hand), then in what order (a `put_in_library`
with `top`). A pile of one has no order, so no second prompt is raised for it.

## Decision 2 — The move is the ordinary route, one card at a time, in pile order

`Game.PutInLibraryInChosenOrderThenForEffect(spec)` queues the prompt; `ResolvePutInLibrary`
validates the answer and places the cards. Each card goes to its **owner's** library (CR 400.3 —
every printed variant says "its owner's library" or means it).

- A card **already in its owner's library** is a reorder, not a zone change: it is lifted out and
  pushed at its position, no CR 614 window opens and no zone-move event fires, because nothing
  changed zones. This is Goblin Ringleader's "the rest" and Aetherspouts' tucked attackers.
- A card **anywhere else** goes through `tuckRoute` — `TuckToLibraryThenForEffect` with `ToBottom`
  — so the CR 614 window opens and a commander is offered the command zone (CR 903.9). A leg that
  pauses on that offer pauses the pile; the next card is placed from its continuation, which is what
  keeps the order right on either side of the prompt.

The legs go **in pile order**: the bottom pile first card first (each `PushBottom` goes under the
one before, so the first entry ends nearest the top), then the top pile last card first (each
`PushTop` lands above, so the first entry ends on top).

**`From` is CR 400.7.** The spec names the zone kind the cards are in when the prompt is raised, and
a card that is no longer in a zone of that kind when the answer arrives is skipped — it is a new
object the instruction no longer refers to. Without it a stale ID pulls a card back out of a hand or
off the battlefield, which is the bug #744 fixed for the random bottom.

**No prompt when there is no choice.** An empty pile runs `Then` at once; a pile of one with a
one-lane placement is placed without asking. `top_or_bottom` always asks, even for one card —
that is a real two-way choice (Hinder).

## Decision 3 — Visibility: "look at", and CR 401.4 keeps the order private

The chooser is made a knower of every card in the prompt when it is queued: ordering cards you
cannot see is not a choice. The view projects the options exactly as the rest of the scry family
does, so every other seat sees only the cards it already knew.

**After the move**, a card placed in a library keeps knowledge by one rule, CR 401.4's last
sentence: *the order is not revealed*.

- A card placed **alone in its lane** keeps every knower it had before it moved, plus the chooser.
  It has no order to hide: a Hinder'd spell put on top of its owner's library was seen going there
  by the whole table.
- Cards placed **two or more to a lane** are known afterwards by the chooser alone. The other seats
  saw them (Goblin Ringleader reveals four) but not the arrangement, and a knower mark is per
  card-at-a-position — leaving the table marked on three cards in a pile of three would be telling
  it the order.

The tuck route's own rule — a card entering a library loses its knowers — still applies to every
other tuck. It is right for a Sensei's Divining Top that nobody arranged; it is wrong for a card a
player just put somewhere on purpose.

## Decision 4 — Depth is a sentence, not a prompt

"Put it into its owner's library second / third / seventh from the top" has no decision in it, so
it is not this prompt. Two pieces close the row's depth half:

- `effects.PutIntoLibrary{Card, Depth, ToBottom}` — the card-side sentence over
  `TuckToLibraryThenForEffect(… TuckOptions{Depth})`, with a `Then`. The engine had the position
  (`TuckToLibraryAtDepthForEffect`, #783) and three catalog cards calling it by hand; nothing a card
  author would find by reading `primitives.go`.
- `SearchLibrarySpec.Depth` / `effects.SearchLibrary.Depth` — "then shuffle and put that card
  **third** from the top". It rides the existing `ToTop` placement after the shuffle, through
  `Zone.InsertFromTop`, so a library shorter than the depth takes the card on the bottom (the
  closest position the zone has).

## Decision 5 — Table obligations

- **Gate:** blocks the table (`choiceGateDecisions`), for every resolution-time prompt's reason —
  the effect that asked is paused mid-resolution with the cards still where they were.
- **Leaving the game:** dropped with the chooser (`{}` in `choiceDepartureDecisions`), like the rest
  of the family. Every catalogued chooser owns the cards, and CR 800.4a takes those out of the game
  with them. A future card whose chooser is not the owner (Hinder: "your choice of the top or bottom
  of its OWNER's library") leaves the cards where the effect had already put them — on top — which is
  a legal position for the instruction.
- **Enumerator:** `internal/legal` offers the canonical answers rather than N! permutations —
  leave the order alone, and pull each card to the front of its lane; for `top_or_bottom`, keep all
  on top, bottom all, and bottom each one. Every one is a legal answer, and "leave it alone" is
  always among them.
- **Bot:** the heuristic scores a `put_in_library` answer by the card it leaves on top of the bot's
  own library, and — on `top_or_bottom` — by what it buries: its own worst cards, and every card an
  opponent owns.
- **Undo:** the continuation (`PendingChoice.libraryOrderResume`) captures only IDs and scalars and
  is carried by `Clone` like every other resume frame; the resolver hands it copies of the answer, so
  "undo the answer, answer again" lands where answering once would. It is counted in the snapshot
  census, so a restore point is never taken with one open, and `LibraryPlacement` is plain data the
  snapshot carries.

## Cards

| Card | Before | After |
|---|---|---|
| Goblin Ringleader | caveat: the rest go to the bottom in a random order | `full` — `bottom`, any order |
| Brainstorm | `full`, but the order was pick order in a question string | `full` — pick, then a `top` order prompt |
| Aetherspouts | `full`, but the choice was a Scry (fires "whenever you scry", opens the scry window) | `full` — `top_or_bottom` per owner, no keyword action |
| Impulse | not in the catalog | `full` — look at four, take one, the rest on the bottom in any order |
| Oust | not in the catalog | `full` — `PutIntoLibrary{Depth: 2}` |
| Long-Term Plans | not in the catalog | `full` — `SearchLibrary{ToTop, Depth: 3}` |
| Vampiric Tutor | caveat "goes straight to your hand" — stale since S22, when `ToTop` shipped | `full` |

`TakeRestOnBottomInAnyOrder` is the `Then` for `TakeFromLibraryToHand` that every "look at the top N,
put one into your hand and the rest on the bottom of your library in any order" card now uses;
`RevealTopThenTakeToHand` uses it too, because the card it was written for (Goblin Ringleader)
prints "in any order".

## Out of scope (explicit deferrals) — tracked on #1298

> The first four of these shipped on 2026-09-23 — see the
> [amendment](#amendment-2026-09-23-1298-the-library-positions-put_in_library-did-not-cover) at
> the end. Library of Leng's batch order is still out of scope.

- **A look at ANOTHER player's library top** (#996 item 5 — Jace, the Mind Sculptor's +2, Chaos
  Wand's look). `put_in_library` can order cards in anyone's library, but the LOOK that precedes it
  (`lookAtTopForEffect`) still takes the looker and the library owner as one player.
- **"Second from the top OR on the bottom"** (Temporal Cleansing, Lost Days, Wan Shi Tong) — a
  two-way choice whose top lane is a depth. It is a `ConfirmPrompt` over `PutIntoLibrary`, card
  work on today's pieces, not a new shape.
- **An exact count on the top lane** ("put one of them on top of your library and the rest on the
  bottom in any order" — Cream of the Crop). The placement would need a lane count; nobody in the
  catalog prints it.
- **Hinder / Spell Crumple** ("put that card on your choice of the top or bottom of its owner's
  library instead of into that player's graveyard") — a counter-to-library followed by a
  `top_or_bottom` prompt whose chooser is not the owner. The counter's CR 903.9 leg has no
  continuation to hang the choice off, so the card waits for one.
- **Library of Leng's** multi-card discard redirect still lands in the order the discard names; it
  is a replacement on a batch, not an instruction with a chooser, and CR 401.4's owner arrangement
  there is a replacement-pipeline question.

## Consequences

- Forty-odd "rest on the bottom in any order" cards are writable on one `Then`.
- The scry dialog grows one control — reorder within the bottom lane — which scry itself also
  wants: CR 701.22a puts the bottom cards "in any order" too, and the engine always applied the
  `bottom` list in the order given.
- A client that sends a scry-shaped answer to a `put_in_library` prompt is routed correctly: the
  dispatcher reads the kind.

## Test plan

- `game/library_order_test.go` — each placement from a library, from a hand and from the
  battlefield; top-first order in both lanes; the lane a placement does not open is refused; a
  partial, duplicated or foreign answer is refused; a card that left its `From` zone is skipped; a
  pile of one raises no prompt; knowers after the move (alone vs. two or more to a lane); a commander
  leg pausing on CR 903.9 keeps the pile order; undo across the prompt replays identically; the
  snapshot census counts the continuation.
- `legal/put_in_library_test.go` — every placement enumerates, every answer offered dispatches, and
  no other seat is offered anything while it is open.
- `protocol/put_in_library_view_test.go` — `placement` and the options on the chooser's wire, and
  no handle on the cards on anyone else's.
- `aiseat/heuristic` — the bot buries an opponent's card on top-or-bottom and puts its best card on
  top.
- `client/src/lib/putInLibraryModal.render.test.ts` — the three placements' lanes and payloads.
- One card test per proof card (`cards/effects/library_order_test.go`, the updated Aetherspouts and
  Brainstorm tests, the existing Vampiric Tutor test).

---

## Amendment (2026-09-23, #1298): the library positions put_in_library did not cover

**Trackers:** [#888](https://github.com/krakenhavoc/cmd_and_ctrl/issues/888) (S45) and
[#879](https://github.com/krakenhavoc/cmd_and_ctrl/issues/879) (S36: the prompt still owes the
table an answer, now sometimes from a player who does not own the cards). No new ADR number. The
prompt kind, its wire keys and its table obligations are unchanged. What changes is who may
choose, which positions the top lane can mean, and what one leg of the move is.

The first four "Out of scope" items above, in order.

### A1 — The looker and the library's owner are two players

"Look at the top card of target player's library. You may put that card on the bottom of that
player's library" (Jace, the Mind Sculptor's +2). "Look at the top three cards of target player's
library, then put them back in any order" (Portent). The placement half already worked: a
put_in_library places each card in its OWNER's library, and a card already there is reordered in
place. The look is what was missing. `lookAtTopForEffect` treats the looker and the owner as one
player, and so did `LookAtTopOfLibraryForEffect`.

`Game.LookAtTopOfPlayersLibraryForEffect(looker, owner, n)` (`game/random_bottom.go`) is the look
with the two split. `LookAtTopOfLibraryForEffect(p, n)` is now `(p, p, n)`. Only the looker is
marked a knower. The library's owner learns nothing: CR 401.2 keeps a library face down to its
owner too, and a look (CR 701.20) is not a reveal. The card side is
`effects.LookAtLibraryThenPlace{Owner, N, Placement}`, which is that look followed by a
put_in_library whose chooser is the looker.

`lookAtTopForEffect` itself is **not** split. Scry, surveil and "look at the top N of your library,
then put them back" are the chooser's own library by definition (the keyword or the printed "your"
says so). Their resolvers check that the chooser owns the cards, and that check is correct.
put_in_library was already the family member that makes no such claim.

**Visibility after the move** is Decision 3 unchanged. A card alone in its lane keeps its knowers.
Jace's buried card was known to Jace's controller and stays known only to them. A lane of two or
more is known to the chooser alone. Portent's three cards are known only to the caster afterwards,
even when their owner had scryed them first, because the order is new and CR 401.4 does not reveal
it.

### A2 — A top lane at a depth, chosen by the card's owner

"The owner of target nonland permanent puts it into their library second from the top or on the
bottom" (Temporal Cleansing, Lost Days, Wan Shi Tong, All-Knowing). The out-of-scope note guessed
at a `ConfirmPrompt` over `PutIntoLibrary`. We did not do that, because a confirm has no library
position on the wire, the enumerator and the bot would score it as a bare yes/no, and the client
would render it without the card. The choice is top-or-bottom with the top lane moved down.

`PutInLibrarySpec.TopDepth` (`PendingChoice.LibraryTopDepth`) is where the top lane starts, counted
from the top. 2 means "second from the top", and 0 or 1 means the top. The lane stays top-first:
its legs run last-first, each inserted at the depth, so the first card of the answer ends at the
depth. A reorder uses `Zone.InsertFromTop`, and the tuck route uses `TuckOptions.Depth`. As on
the tuck route, a library shorter than the depth takes the card on the bottom.

The chooser is the permanent's owner. `effects.PutIntoLibraryAtDepthOrBottom{Card, Depth}` defaults
`Chooser` to the owner, because every printed variant names the owner.

### A3 — An exact count on the top lane

"Put one of those cards on top of your library and the rest on the bottom of your library in any
order" (Cream of the Crop). `PutInLibrarySpec.TopCount` (`PendingChoice.LibraryTopCount`) is
exactly how many cards the answer's `top_order` must hold. It is allowed only with
`top_or_bottom`, and `ResolvePutInLibrary` refuses any other count. A pile no bigger than the
count has no lane left to choose: it is placed as `top`, is asked only for its order, and is not
asked at all when it is one card. That is Cream of the Crop off a 1-power creature.

### A4 — A counter to a position, chosen by the counterer

"Counter target spell. If that spell is countered this way, put that card on your choice of the
top or bottom of its owner's library" (Hinder). "…put it on the bottom of its owner's library"
(Spell Crumple). The out-of-scope note had the order backwards. It assumed counter first, then
ask, and found no continuation to ask from. The card resolves the other way round: the counterer
picks the end, then the spell is countered to it. Asking first also means:

- the spell is still on the stack while its fate is decided, so nothing is momentarily on top of
  a library that the answer then moves;
- a commander's owner is offered CR 903.9 knowing which end the card was headed for;
- a spell that can't be countered is never asked about. It was not "countered this way", so it
  stays on the stack and the effect goes on.

`PutInLibrarySpec.Counter` makes each leg a counter. `From` is forced to the stack, a card is live
only while it is a spell a counter would counter (`counterableSpellOnStackLocked`), and the leg is
`Game.CounterSpellToLibraryThenForEffect(stackID, TuckOptions, then)` (`game/effect_api.go`). That
is `CounterTargetToZoneForEffect` with a library position. Underneath it,
`exitSpellFromStackAtLocked` is the shared stack exit with the position riding the route, so
flashback's CR 702.34a exile still wins, EventCounterSpell fires, and the stack record retires. A
countered ability ceases to exist (CR 701.6b) and is never offered. The card side is
`effects.CounterToLibrary{StackID, Placement}`. Spell Crumple uses `bottom` with no choice, so it
raises no prompt.

**Visibility.** A card that was in a public zone (the stack, the battlefield) was watched by the
table, whether or not anything had stamped a knower mark on it. `libraryOrderKnowers` now counts
every seat as a prior knower of a card whose source zone is public (`publicKnowersOfLocked`,
CR 400.2). Placed alone in its lane, a Hinder'd spell or a Temporal Cleansing'd permanent is known
to the whole table on top of the library. That is what Decision 3 already said it should be. A
pile of two or more is still the chooser's alone.

### A5 — Leaving the game

The departure row stays `{}` (dropped, and the drop does nothing more). Decision 5 assumed every
chooser owns the pile. That is no longer true, but the drop is still right in each new case:

- **Jace / Portent.** The looked-at cards stay on top of the other player's library, untouched,
  and the looker's knowledge leaves with the looker.
- **Hinder.** Because the choice comes before the counter, a Hinder whose controller leaves
  mid-prompt countered nothing, and the spell stays on the stack. This replaces Decision 5's guess
  that the card would be "where the effect had already put them — on top".
- **Temporal Cleansing.** The chooser is the owner, and CR 800.4a took the permanent with them.

### A6 — Enumerator, bot, wire, client

- **Enumerator** (`legal/choices.go`). With a `TopCount` there is no "leave it alone" answer. The
  canonical set is one answer per card, with that card first on top, the next `TopCount-1` beside
  it and the rest under. Every answer holds exactly the count, and every one dispatches. The other
  new shapes need nothing new: a depth changes where the top lane lands, not which answers are
  legal.
- **Bot** (`aiseat/heuristic/choices.go`). No new branch. With an exact count every answer holds
  the same number on top, so the existing terms rank the answers by the card left on top. An
  opponent's card on a top-or-bottom (Jace, Hinder) is buried by the existing opponent term.
- **Wire.** `PendingChoiceView.TopCount` / `TopDepth` (`top_count`, `top_depth`) are public, like
  `placement`. Options still reach the chooser alone.
- **Client** (`ChoicePromptModal.svelte`, the existing reorder dialog). With a count, the seed is
  already a legal answer (the first N on top), the top lane reads "On top (k of N)", and Done is
  disabled until the count holds. With a depth, the top lane is named "Second from the top". When
  the cards are not the viewer's (Jace, Portent, Hinder), the hint no longer calls the top card
  "your next draw".

### Cards

| Card | Shape | Completeness |
|---|---|---|
| Jace, the Mind Sculptor | A1 (+2). The 0 is Brainstorm's sentence, the −1 a bounce, the −12 exile-library then tuck-hand-and-shuffle | `full` (all four abilities) |
| Portent | A1 with an ordered top, then "you may have that player shuffle", then a CR 603.7 draw at the next upkeep | `full` |
| Temporal Cleansing | A2 (convoke is `Spec.TapCost`) | `full` |
| Cream of the Crop | A3 | `caveats`: X falls back to the creature's power when the ability triggered if the creature has left by resolution. The engine keeps no last-known power past the trigger's own dispatch. |
| Hinder | A4, top or bottom | `full` |
| Spell Crumple | A4, bottom, then the card's own "put Spell Crumple on the bottom" | `full` |

Spell Crumple's second sentence runs on the line after the counter, not in the counter's `Then`.
The resolution frame routes a still-resolving spell to the graveyard as soon as `OnResolve`
returns, so a `Then` that ran later would tuck the card back out of the graveyard. That would
happen after a countered commander's owner answered CR 903.9. The cost is one ordering corner: if
you counter your *own* commander with Spell Crumple and decline the command zone, the commander
ends up under Spell Crumple instead of over it.

### Still out of scope

- Library of Leng's multi-card discard order (unchanged, above).
- Lost Days and Wan Shi Tong, All-Knowing are A2 card work. Wan Shi Tong's "whenever one or more
  cards are put into a library from anywhere" has no batched event yet.

### Test plan (amendment)

- `game/library_order_leftovers_test.go`:
  - the looker is the only knower;
  - another player's library is reordered in place, and its owner does not learn the order;
  - depth through the tuck route, as a pile, and as a reorder;
  - the exact count is enforced, and a count that fits the pile becomes a top placement;
  - a malformed spec is refused;
  - a counter to each end: EventCounterSpell fires, the card lands in its owner's library, and
    the table knows it;
  - an uncounterable spell is not asked about;
  - a countered commander is offered CR 903.9 and, on decline, lands at the bottom.
- `legal/put_in_library_leftovers_test.go` — the exact-count answers, and a chooser who is not
  the owner (a look, a counter): offered answers dispatch, and nobody else is offered anything.
- `protocol/put_in_library_leftovers_view_test.go` — a look at another player's library reaches
  the looker's wire only, and `top_count` / `top_depth` are public.
- `aiseat/heuristic/choices_test.go` — the exact count keeps the best card on top, and the top of
  an opponent's library is buried.
- `client/src/lib/putInLibraryModal.render.test.ts` — the count seeds a legal answer and holds
  Done, the depth names the lane, and another player's library is not "your next draw".
- `cards/effects/library_top_leftovers_test.go` — one test per proof card, and all four of Jace's
  abilities.
