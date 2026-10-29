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
//	graveyard destination   CR 701.17a defines the keyword action by
//	                        where the cards go. "Exile the top four
//	                        cards of your library" uses the same
//	                        helper and is not a mill, so no window.
//	a positive count        There is nothing to replace about milling
//	                        nothing, and an unbounded `until` run (Helm
//	                        of Obedience, n <= 0 with a predicate)
//	                        names no number to double.
//
// The count is the number the INSTRUCTION asked for, not the number
// the library can supply. CR 701.17b makes a player told to mill more
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
// ScryKeyedForEffect's zero already carry, and the reason #893's doc
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
// The planning stays where #529 put it — the candidate set is chosen
// top-down before anything moves, so a leg paused on CR 903.9 does not
// stall the instruction and does not shorten it.
//
// # And where its BOUND ends it (#1161)
//
// In the same clause. "Until a creature card or X cards have been put
// into their graveyard this way, whichever comes first" is one
// sentence with two stop conditions, and both of them count ARRIVALS
// — so X is not the mill's amount, it is the other half of the clause
// (effects.UntilAny(UntilCard(...), UntilCount(x))). That is why a
// diverted card does not use up one of the X and why under Rest in
// Peace the Helm mills the whole library.
//
// The amount and the bound are different rules and this is the line
// between them. CR 701.17b's number is the one the INSTRUCTION names —
// "mill three" — and it is the number a mill-amount replacement
// doubles; a bound that counts arrivals is not one, and Bruvac the
// Grandiloquent does not double it.
//
// The clause is typed as a function of the whole LANDED LIST rather
// than of one card. That is not decoration: the run carries the landed
// list forward by value across a pause, so a predicate over it is pure
// and an undo that rewinds into an open prompt asks the same question
// and gets the same answer. A per-card predicate accumulating state
// (Improvisation Capstone's running mana-value total) would be
// consumed by the first run and wrong on the replay.
//
// # But the run is a SEQUENCE of instructions (#1176)
//
// "Target opponent MILLS A CARD, then repeats this process until …"
// gives one instruction and repeats it. Modelling the whole run as one
// instruction that names no number got the bound right and the AMOUNT
// wrong: there was no number for a mill-amount replacement to double,
// so Helm of Obedience + Bruvac at X=3 milled three cards where paper
// mills four (Helm's ruling: each repetition is its own mill, Bruvac
// replaces each of them, and the bound is checked after each replaced
// mill, which can overshoot it by a card).
//
// So each repetition is its own one-card mill instruction
// (millUntilRunKeyedLocked): its own RepEventMill window, its own plan, its
// own sequenced routing, and then the clause, asked with everything
// the run has landed. Both cards of a doubled repetition move — one
// instruction, one batch — which is where the overshoot comes from.
//
//	Helm X=1 + Bruvac   2 cards (one doubled repetition)
//	Helm X=2 + Bruvac   2 cards (the first repetition reaches the bound)
//	Helm X=3 + Bruvac   4 cards (2 + 2)
//
// One thing the loop needs that a single instruction did not: a
// TERMINATION guard. A repetition is a real instruction, so it can be
// replaced away — CR 614.10's null replacement cancels a mill, and a
// replacement that halves an amount turns a mill of one into a mill of
// none — and a run that asks for a repetition which moves nothing would
// ask for it forever. millUntilRunKeyedLocked ends the run when a repetition
// leaves the library no shorter than it found it. The measure is the
// library's DEPTH and not the landed list, because a run under Rest in
// Peace lands nothing and is still making progress.
//
// #1177's structure survives it and is strengthened. The over-mill it
// closed by construction — a landed-count bound only ends a run when
// the engine WAITS for each leg to land, and the fire-and-forget entry
// point does not wait — is still closed the same way:
// MillToZoneForEffect takes no `until`, every repetition goes through
// the sequencing form, and the next repetition starts from the
// previous one's continuation, so a leg paused on CR 903.9 holds the
// whole run. What changed is that no ROUTING LOOP is handed a clause
// either: the verdict is asked between repetitions now, not between
// legs, and routeAllThenUntilLocked is gone.

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
	// CR 701.17a's mill, exile for "exile the top N cards of your
	// library". It is also what landedInZoneLocked measures "milled
	// this way" against (CR 400.7, ADR 0013 §5l).
	dest ZoneKind

	// report is a synchronous door's result (payMillCostLocked): every
	// card the batch took out of the library, written as the mill
	// finishes and read on the door's next line, the way tokenTail.report
	// is. A mill that paused has returned by then, so a restore point
	// drops it. Not a continuation: #4361 S9 deleted the closure `then`.
	report *[]uuid.UUID

	// allMoved changes what the tail is handed: not the cards that LANDED
	// in dest, but every card the mill took out of the library, wherever
	// it ended up (CR 701.17c, #3846). Rest in Peace sends a milled card
	// to exile, where it did not "land" in a graveyard and is still a
	// card that was milled -- Locke, Treasure Hunter's "a spell from
	// among those cards". The caller reads each card's zone itself. Plain
	// data, so the tail stays a restore point; false for every caller but
	// MillToZoneKeyedMovedForEffect and the mill cost.
	allMoved bool

	// keyed is the caller's continuation as data (#4253 slice 9,
	// mill_keyed.go): non-nil for the keyed doors, which always sequence,
	// even with a zero tail. run is a keyed "until" run's repetition in
	// flight. Cleared through the pointer as they run.
	keyed *KeyedTail
	run   *millUntilRun
}

// millThroughReplacementsTailLocked is the one body every mill entry
// point goes through: it validates the instruction, opens the CR 614
// window on the amount when the instruction is a mill the rules can
// replace, and plans and routes what the window settles on.
//
// The returned slice is the cards that LANDED in dest, and it is empty
// for a sequenced tail (a keyed door, which reports through its tail)
// and for a mill that PAUSED on a CR 616 prompt (nothing has moved yet;
// the resume mills when the prompt is answered).
//
// An "until" run is not one instruction: see millUntilRunKeyedLocked.
//
// Caller must hold g.mu.
func (g *Game) millThroughReplacementsTailLocked(
	playerID uuid.UUID,
	n int,
	dest ZoneKind,
	tail *millTail,
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
		mill:       tail,
	}
	if !millAmountIsReplaceable(dest, n) {
		// Not a mill of a number: an exile of the top N, or a mill of
		// nothing. Straight to the plan, through the same body a
		// settled window reaches.
		return g.applyResolvedMillLocked(ev)
	}
	return g.runMillLocked(ev)
}

// millAmountIsReplaceable reports whether an instruction is a mill
// CR 614 can replace the amount of. See the file comment: a graveyard
// destination and a count somebody could double.
//
// #1176: every repetition of an `until` run reaches this with n == 1,
// because a repetition is an ordinary one-card mill. The `n <= 0`
// arm is now only "mill nothing".
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
	if ev.MillCount <= 0 {
		// Replaced down to nothing, or asked for nothing. Not a
		// cancellation — the instruction stands, it simply has no count
		// left — but there is nothing to move and the caller still has
		// to be told.
		//
		// #1176: there is no "unbounded" arm here any more. An `until`
		// run never reaches this function as a run; it reaches it one
		// repetition at a time, each of them a mill of exactly one.
		return nil, g.abandonMillLocked(ev)
	}
	plan, err := g.millPlanLocked(ev.MillPlayer, ev.MillCount, tail.dest)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(plan))
	for _, c := range plan {
		ids = append(ids, c.InstanceID)
	}
	r := millRoute(ev.MillPlayer, tail.dest)
	if run := tail.run; run != nil {
		// #4253 slice 9: a keyed run's repetition; the run rides the
		// batch, and the batch's end asks the clause.
		tail.run = nil
		rep := cloneMillUntilRun(run)
		for _, c := range plan {
			var scratch ContinuationCensus
			rep.Planned = append(rep.Planned, snapshotCard(c, &scratch))
		}
		return nil, g.routeAllRunLocked(r, ids, &routeBatchRun{Mill: rep})
	}
	if k := tail.keyed; k != nil {
		tail.keyed = nil
		run := &routeBatchRun{Then: snapshotKeyedTail(*k), AllMoved: tail.allMoved}
		if tail.allMoved {
			run.Planned = copyUUIDs(ids)
		}
		err := g.routeAllRunLocked(r, ids, run)
		if rep := tail.report; rep != nil {
			// The mill cost's read-back (payMillCostLocked): every card
			// the batch has taken out of the library by now, wherever it
			// went (CR 701.17c).
			tail.report = nil
			*rep = g.leftLibraryLocked(ids)
		}
		return nil, err
	}
	// Fire-and-forget: every leg is routed on this line and one that
	// pauses on CR 903.9 lands later without holding the rest of the
	// mill up (#529).
	return g.routeAllLandedLocked(r, ids), nil
}

// leftLibraryLocked is the subset of `ids`, in order, that is no longer
// in a library -- the mill took it out and it is somewhere else now: a
// graveyard, exile (Rest in Peace), the command zone. A card a CR 614
// window left where it was is not in it. Read off the live board, like
// landedInZoneLocked, so it is true after an undo replays an open prompt.
//
// Caller must hold g.mu.
func (g *Game) leftLibraryLocked(ids []uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range ids {
		if z := g.findCardZoneLocked(id); z != nil && z.Kind != ZoneLibrary {
			out = append(out, id)
		}
	}
	return out
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
	if ev != nil && ev.mill != nil && ev.mill.run != nil {
		run := ev.mill.run
		ev.mill.run = nil
		return g.continueMillUntilRunKeyedLocked(run, nil)
	}
	if ev != nil && ev.mill != nil && ev.mill.keyed != nil {
		k := *ev.mill.keyed
		ev.mill.keyed = nil
		return g.runKeyedTailLocked(k, nil, 0)
	}
	return nil
}
