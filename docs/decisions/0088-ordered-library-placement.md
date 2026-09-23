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

## Out of scope (explicit deferrals)

- **A look at ANOTHER player's library top** (#996 item 5 — Jace, the Mind Sculptor's +2, Chaos
  Wand's look). `put_in_library` can order cards in anyone's library, but the LOOK that precedes it
  (`lookAtTopForEffect`) still takes the looker and the library owner as one player.
- **"Second from the top OR on the bottom"** (Temporal Cleansing, Lost Days, Wan Shi Tong) — a
  two-way choice whose top lane is a depth. It is a `ConfirmPrompt` over `PutIntoLibrary`, card
  work on today's pieces, not a new shape.
- **An exact count on the top lane** ("put one of them on top of your library and the rest on the
  bottom in any order" — Cream of the Crop). The placement would need a lane count; nobody in the
  catalog prints it.
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
