package game

import (
	"errors"

	"github.com/google/uuid"
)

// mill.go is the CR 614 window on the MILL AMOUNT (#569).
//
// # What was missing, precisely
//
// Not the per-card window. A claim that "MillToZoneForEffect skips the
// CR 614 pipeline" circulated on three status comments and a PR body
// and it was never true: every milled card goes through
// routeCardToZoneLocked, which builds a RepEventMove and runs
// applyReplacementsLocked, so Leyline of the Void and Stone of Erech
// see each card and a milled commander is offered the command zone
// (CR 903.9, #539). #893 went further and chose the batch up front so
// one card's prompt does not shorten the rest of the run.
//
// What was missing is narrower and differently shaped. "If an opponent
// would mill one or more cards, they mill twice that many cards
// instead" (Bruvac the Grandiloquent) is not a per-card zone move at
// all. It replaces the NUMBER, once, before anything leaves the
// library — so it needs its own event, opened once per instruction,
// exactly as RepEventCreateTokens is opened once per creation
// (ADR 0061 decision 1) and RepEventKeywordAction once per action
// (ADR 0013 §5s).
//
// # When the window opens
//
// Only for something the rules call a mill, which is narrower than
// what the mill helpers express:
//
//	graveyard destination   CR 701.13a defines the keyword action by
//	                        where the cards go. "Exile the top four
//	                        cards of your library" uses the same
//	                        helper and is not a mill, so no window.
//	a positive count        There is nothing to replace about milling
//	                        nothing, and an unbounded `until` run (Helm
//	                        of Obedience, n <= 0 with a predicate)
//	                        names no number to double.
//
// The count is the number the INSTRUCTION asked for, not the number
// the library can supply. CR 701.13b makes a player told to mill more
// cards than they have mill as many as possible, and that clamp is
// millPlanLocked's, applied after this window settles.
//
// # What it is not
//
//   - A surveil's graveyard leg is not a mill (CR 701.14a is its own
//     keyword action). ResolveSurveil routes those cards with
//     millRoute so that the per-card graveyard window and the mill
//     PAYOFFS see them — a "whenever a card is put into your graveyard
//     from your library" trigger must not care how it got there — but
//     no mill instruction was given, so Bruvac does not double it.
//   - The per-card window is untouched. A mill therefore opens two
//     kinds of window in sequence: one RepEventMill for the amount,
//     then one RepEventMove per card the settled amount asks for.
//
// # Pausing
//
// Both mill forms can now pause BEFORE any card moves, which neither
// could before: two mill-amount replacements in one window is a
// CR 616.1 ordering prompt. So the destination, the `until` predicate
// and the caller's continuation ride the event on millTail, and the
// resume plans and routes through exactly the function the unpaused
// path runs.
//
// MillToZoneForEffect's returned slice is empty on such a pause — the
// contract CreateTokensForEffect's empty ID slice and
// ScryThenForEffect's zero already carry, and the reason #893's doc
// already told a caller that reads the list to use the Then form.
//
// # Where an "until" run ends (#1159)
//
// On the card that ARRIVED in dest, not on the card that came off the
// library. The clause is CR 400.7's, the same reading the landed list
// already gives the Then continuation: a commander whose owner takes
// the command zone was never put into that graveyard, so it is not one
// of the cards "until a creature card is put into their graveyard"
// counts, and the run carries on.
//
// The planning stays where #529 put it — the whole candidate set is
// chosen top-down before anything moves, so a leg paused on CR 903.9
// does not stall the run and does not shorten it. What moved is the
// VERDICT: millPlanLocked no longer truncates the plan, and the clause
// rides the routing loop as a stop predicate
// (routeAllLandedUntilLocked, routeAllThenUntilLocked) consulted only
// for legs that landed.
//
// The clause is therefore typed as a function of the whole LANDED LIST
// rather than of one card. That is not decoration: the loops carry the
// landed list forward by value across a pause, so a predicate over it
// is pure and an undo that rewinds into an open prompt asks the same
// question and gets the same answer. A per-card predicate accumulating
// state (Improvisation Capstone's running mana-value total) would be
// consumed by the first run and wrong on the replay.

// millTail is what a mill instruction still owes once the amount
// settles — the mill's sibling of zoneRoute, tokenTail and
// keywordActionTail, and it exists for the same reason they do: the
// window can pause, and the resume has to finish the mill exactly as
// the caller asked for it.
//
// Unexported engine plumbing. The catalog never sets or reads it; a
// replacement rewrites MillCount on the event and nothing else.
type millTail struct {
	// dest is where the cards were asked to go — a graveyard for
	// CR 701.13a's mill, exile for "exile the top N cards of your
	// library". It is also what landedInZoneLocked measures "milled
	// this way" against (CR 400.7, ADR 0013 §5l).
	dest ZoneKind

	// until, when non-nil, stops the run AFTER the first card that
	// LANDS in dest and makes it true — Helm of Obedience's "until a
	// creature card is put into their graveyard".
	//
	// #1159: it is answered in the ROUTING loop, against the cards
	// that arrived (CR 400.7, landedInZoneLocked), not in
	// millPlanLocked against the cards that came off the library. A
	// card the CR 614 window diverted was never put into that
	// graveyard, so it is not one of the cards the clause counts and
	// the run carries on past it. The plan is still a flat list of IDs
	// chosen before anything moves, so the batch body still proceeds
	// AROUND a leg paused on CR 903.9 (#529).
	//
	// It takes the whole landed list rather than one card so that it
	// is pure: an undo that rewinds into an open prompt replays the
	// answer and must ask the same question. millStopForLocked binds
	// it to the plan's pre-move copies.
	until func([]Card) bool

	// then is the continuation form's callback, run with the cards
	// that LANDED in dest. nil for the fire-and-forget form, whose
	// caller reads the returned slice instead.
	then func(g *Game, milled []uuid.UUID) error
}

// millThroughReplacementsLocked is the one body both mill entry points
// go through: it validates the instruction, opens the CR 614 window on
// the amount when the instruction is a mill the rules can replace, and
// plans and routes what the window settles on.
//
// The returned slice is the cards that LANDED in dest, and it is empty
// for the continuation form (which reports through `then` instead) and
// for a mill that PAUSED on a CR 616 prompt (nothing has moved yet;
// the resume mills when the prompt is answered).
//
// Caller must hold g.mu.
func (g *Game) millThroughReplacementsLocked(
	playerID uuid.UUID,
	n int,
	dest ZoneKind,
	until func([]Card) bool,
	then func(g *Game, milled []uuid.UUID) error,
) ([]uuid.UUID, error) {
	// The up-front errors, asked before the window rather than inside
	// millPlanLocked's copy of them: an unseated player and a
	// destination that is neither a graveyard nor exile are not mills
	// that failed, they are calls that were never legal, and opening a
	// replacement window for one would let a card watch an instruction
	// that does not exist.
	if g.playerByIDLocked(playerID) == nil {
		return nil, ErrPlayerNotFound
	}
	switch dest {
	case ZoneGraveyard, ZoneExile:
	default:
		return nil, ErrInvalidParam
	}
	ev := &ReplacementEvent{
		Kind:       RepEventMill,
		Actor:      playerID,
		MillPlayer: playerID,
		MillCount:  n,
		mill:       &millTail{dest: dest, until: until, then: then},
	}
	if !millAmountIsReplaceable(dest, n) {
		// Not a mill of a number: an exile of the top N, or an
		// unbounded `until` run. Straight to the plan, through the same
		// body a settled window reaches.
		return g.applyResolvedMillLocked(ev)
	}
	return g.runMillLocked(ev)
}

// millAmountIsReplaceable reports whether an instruction is a mill
// CR 614 can replace the amount of. See the file comment: a graveyard
// destination and a count somebody could double.
func millAmountIsReplaceable(dest ZoneKind, n int) bool {
	return dest == ZoneGraveyard && n > 0
}

// runMillLocked runs the CR 614 window for ev and, once it settles,
// mills what the window left. Shared by the entry points above and by
// the CR 616 resume (applyResolvedReplacementEventLocked), so a paused
// mill and an unpaused one cannot drift apart.
//
// Caller must hold g.mu.
func (g *Game) runMillLocked(ev *ReplacementEvent) ([]uuid.UUID, error) {
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// A prompt is queued and the resume owns the mill now. NOTHING
		// has happened: no card has left the library, no event was
		// emitted, and the caller's continuation is still owed.
		return nil, nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return nil, err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		// CR 614.10 with a null replacement: no cards are milled. The
		// caller's continuation still runs, with nothing.
		return nil, g.abandonMillLocked(ev)
	}
	return g.applyResolvedMillLocked(out)
}

// applyResolvedMillLocked plans and routes the mill a settled
// RepEventMill describes, with the count the window left on it.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedMillLocked(ev *ReplacementEvent) ([]uuid.UUID, error) {
	tail := ev.mill
	if tail == nil {
		tail = &millTail{dest: ZoneGraveyard}
	}
	unbounded := tail.until != nil && ev.MillCount <= 0
	if ev.MillCount <= 0 && !unbounded {
		// Replaced down to nothing, or asked for nothing. Not a
		// cancellation — the instruction stands, it simply has no count
		// left — but there is nothing to move and the caller still has
		// to be told.
		return nil, g.abandonMillLocked(ev)
	}
	plan, err := g.millPlanLocked(ev.MillPlayer, ev.MillCount, tail.dest, tail.until)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(plan))
	for _, c := range plan {
		ids = append(ids, c.InstanceID)
	}
	// #1159: an `until` clause ends the run on a card that ARRIVED, so
	// it rides the routing loop as a stop predicate rather than
	// truncating the plan. nil when there is no clause.
	stop := millStopForLocked(plan, tail.until)
	r := millRoute(ev.MillPlayer, tail.dest)
	if tail.then == nil {
		// Fire-and-forget: every leg is routed on this line and one
		// that pauses on CR 903.9 lands later without holding the rest
		// of the mill up (#529).
		return g.routeAllLandedUntilLocked(r, ids, stop), nil
	}
	// Cleared THROUGH the pointer, for the reason the token and
	// keyword-action tails are: a continuation that re-enters the
	// pipeline on the same tail value must not run itself twice, and an
	// undo snapshot therefore needs its own copy of the tail (which
	// cloneReplacementResume gives it).
	then := tail.then
	tail.then = nil
	return nil, g.routeAllThenUntilLocked(r, ids, stop, then)
}

// abandonMillLocked is the terminal outcome of a mill that moved
// NOTHING — cancelled by a CR 614.10 null replacement, replaced down to
// a count of zero, or a prompt taken away (§5j). It runs the caller's
// continuation with an empty list.
//
// A caller sequencing work behind the mill has to be told even when the
// answer is "none", or it waits forever; that is the call #808 made for
// the life tail, #853 for the route tail, #762 for the token tail and
// #976 for the keyword action. The fire-and-forget form carries no
// continuation and this is a no-op for it.
//
// Caller must hold g.mu.
func (g *Game) abandonMillLocked(ev *ReplacementEvent) error {
	if ev == nil || ev.mill == nil || ev.mill.then == nil {
		return nil
	}
	then := ev.mill.then
	ev.mill.then = nil
	return then(g, nil)
}
