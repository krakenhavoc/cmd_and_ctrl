# ADR 0037 — Telling the player a card's rules aren't implemented

**Status:** Accepted · 2026-09-11 · Sprint S32 · Reports [#321](https://github.com/krakenhavoc/cmd_and_ctrl/issues/321), [#324](https://github.com/krakenhavoc/cmd_and_ctrl/issues/324), [#325](https://github.com/krakenhavoc/cmd_and_ctrl/issues/325), [#332](https://github.com/krakenhavoc/cmd_and_ctrl/issues/332), [#333](https://github.com/krakenhavoc/cmd_and_ctrl/issues/333)

## Context

The effect catalog is opt-in per oracle ID (ADR 0010 §3). 288 cards
have a hand-written Spec; the ~31,500 other Commander-legal cards
resolve with no effect at all and the players move the pieces by
hand, Cockatrice-style. That is a deliberate decision and it is not
changing soon — the catalog grows one card file at a time, and the
coverage roadmap schedules the next 2,000 over twenty batches.

The defect is not the gap. It is the **silence** about the gap.

Five of the twelve reports from the 2026-09-10 playtest are one
player casting five different uncatalogued cards, watching nothing
happen, and reasonably concluding each was broken:

| Report | Card | What the player expected |
|---|---|---|
| #321 | Enduring Curiosity | a draw on each instance of combat damage |
| #324 | Anticausal Vestige | a warp-or-full-cost prompt at cast |
| #325 | Aang, Swift Savior | the ETB trigger to fire |
| #332 | Lotus Field | to enter tapped and prompt for the sacrifice |
| #333 | Fortune Teller's Talent | chapter 1 to do something |

Every one of those is accurate as a description of what happened and
wrong as a diagnosis. Triaging them costs a maintainer a dump lookup
and a registry probe each, and the player learns nothing, so they
file the next five next week. Five missing cards is not the bug. The
bug is that the game has the answer and does not say it.

## Decisions

### 1. The predicate is "prints rules we won't run", not "not in the catalog"

Catalog membership is the tempting signal because it is one function
call away (`effects.Lookup` / `game.IsCatalogCard`, already on the
wire as `CardView.auto`). It is also a lie, and a loud one.

Since #317 / #319 / #320 the engine honours printed keywords on every
card in the dump. An uncatalogued Serra Angel flies, has vigilance,
blocks correctly and deals its damage. An uncatalogued Grizzly Bears
is a complete and correct 2/2. A Forest taps for `{G}` off its type
line alone (CR 305.6). None of those cards is missing anything, and
on a real 100-card deck they are a large minority of the list.

The honest predicate is narrower, and it is two independent facts
joined in exactly one place (`game.Unimplemented`, `coverage.go`):

```go
func Unimplemented(c Card) bool {
	return c.NeedsEffect && !IsAutoCard(c.OracleID)
}
```

`NeedsEffect` is a pure function of the printed Scryfall record,
stamped onto `game.Card` by the deck importer on the road #274 built
for `StartingLoyalty` and #317 built for `Keywords`. It is true when
the card's oracle text has a line that is not a keyword-ability line,
after parenthesised reminder text is removed. Two consequences worth
stating:

- **Keywords the engine does not enforce count as rules.** Ward,
  cycling, equip, flashback — `canonicalKeywords` is a closed set
  precisely because a keyword joins it in the change that teaches the
  engine to honour it (`keywords.go`), so a keyword outside the table
  is by construction a keyword nothing acts on.
- **Multi-faced cards are scanned as the union over faces.** That
  over-counts a DFC with a vanilla front face, which is the right
  answer anyway: multi-face cards are not modelled at all (ADR 0034).

### 2. Lands are handled explicitly, because "not in the catalog" is most wrong about them

A basic land works perfectly with no Spec, so the naive signal is at
its most misleading on the twenty-odd cards every deck runs most
copies of. The text scan gets this right for free — a basic's whole
mana ability is reminder text, so nothing survives the strip.

It gets it right for the wrong reason, though, and there is one card
where that matters. `ManaAbilitiesForCard` synthesises the mana
ability from the type line for the five colours **and nothing else**,
which leaves Wastes: all reminder text, no colour, and no mana
whatsoever without a Spec. `NeedsCatalogEffect` carries an explicit
clause for a basic land the synthetic path does not cover. Two cards
in the format hit it, and getting them right is the difference
between a predicate that is honest about lands and one that merely
looks like it.

Non-basic lands need no special case: every one of them prints text,
and an uncatalogued one really does tap for nothing.

### 3. Three surfaces, chosen as moments rather than as chrome

A board badge was the obvious design and is rejected. Most permanents
on a real battlefield would carry one; a badge on everything is a
badge nobody reads, and it would sit next to the existing gold `AUTO`
pip implying the two are a matched pair of equals when one marks 288
cards and the other marks 31,500.

Instead the signal appears at the three moments a player forms an
expectation, and nowhere else:

| Surface | Moment | Why |
|---|---|---|
| Deck upload summary (`unimplemented` on the upload response, a collapsed `<details>` in `DeckUploadForm`) | before the game | The best moment there is. One honest sentence about a hundred cards, read once, instead of once per mid-combat surprise. Count in the summary, names behind the disclosure. |
| Stack overlay `manual` chip | the spell is resolving | Exactly where all five reports were written. Transient by construction — the chip leaves with the stack item, so it never becomes furniture. |
| Hover / inspect panel line | on demand | The player is already asking what the card does; answering the other half costs nothing and is zero-noise by definition. |

Styling is deliberately quiet in all three: muted greys, a dashed
chip outline rather than a colour. Red means the deck was rejected
and gold means the engine is doing something; this is neither.

### 4. Rejected: a log line at resolution

Suggested and not built. The client has no in-game event feed —
`GameView` does not carry `events` and nothing renders `EventCostWarning`
today — so a new `EventKind` would reach replays and bug-report
artifacts but not the player's eyes. The stack chip covers the same
moment and the player actually sees it. Snapshot-based replays carry
`CardView.unimplemented` already, so triage gets the fact anyway.

Building a log UI to carry one line is the wrong order of work. If
one lands later, the event is a small addition on top of a predicate
that will already exist.

### 5. The signal constrains what may be registered

A card with a Spec loses the flag. That makes partial registrations
actively harmful in a way they were not before: a card shipped with
half its text implemented now reads as **complete**, and the signal
that was supposed to be the honest one starts lying.

So: a card joins the catalog when its whole printed text is carried
out, or it does not join. This is the same rule `canonicalKeywords`
already states for keywords, applied to cards. Where a Spec knowingly
omits a clause, the card comment says so (as `weftstalker_ardent.go`
does for warp) — but that is a card that should not have shipped
under this ADR, and the exceptions should not grow.

## Consequences

- `game.Card` gains one bool. `CardView` gains one omitempty bool.
  The deck-upload response gains one optional string array.
- Cards that never go through deck import — tokens, fixtures, the
  demo seed — carry `NeedsEffect == false` and are never flagged.
  Deliberate: a missed signal costs a player nothing they were not
  going to learn anyway, and a false one costs the signal its
  credibility.
- The predicate is textual and therefore approximate at the edges. It
  errs toward silence everywhere it is uncertain (an unclosed paren
  swallows the rest of the text rather than flagging it), which is
  the direction that keeps it trustworthy.
- The upload summary will say "68 of these cards" for a typical deck
  today and a smaller number every sprint. That number is a coverage
  metric the project did not previously surface to anyone.
