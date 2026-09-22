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
// between them. CR 701.13b's number is the one the INSTRUCTION names —
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
// (millUntilRunLocked): its own RepEventMill window, its own plan, its
// own sequenced routing, and then the clause, asked with everything
// the run has landed. Both cards of a doubled repetition move — one
// instruction, one batch — which is where the overshoot comes from.
//
//	Helm X=1 + Bruvac   2 cards (one doubled repetition)
//	Helm X=2 + Bruvac   2 cards (the first repetition reaches the bound)
//	Helm X=3 + Bruvac   4 cards (2 + 2)
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
	// CR 701.13a's mill, exile for "exile the top N cards of your
	// library". It is also what landedInZoneLocked measures "milled
	// this way" against (CR 400.7, ADR 0013 §5l).
	dest ZoneKind

	// then is the continuation form's callback, run with the cards
	// that LANDED in dest. nil for the fire-and-forget form, whose
	// caller reads the returned slice instead.
	//
	// #1176: for one repetition of an `until` run this is the run's
	// own continuation, which appends what landed, asks the clause,
	// and either finishes the run or starts the next repetition. The
	// tail carries no `until` of its own any more — an instruction is
	// an instruction, and the run is a loop over instructions.
	then func(g *Game, milled []uuid.UUID) error
}

// millRun is one "mills a card, then repeats this process until …"
// run (#1176): the state carried from each repetition to the next.
//
// Every field is treated as IMMUTABLE once the value is built.
// A repetition's continuation builds a FRESH millRun with fresh
// slices rather than appending in place, which is the same property
// routeEachStepLocked's landed list has and for the same reason: an
// undo that rewinds into an open CR 903.9 prompt and replays the
// answer must ask the clause the same question and get the same
// answer.
type millRun struct {
	player uuid.UUID
	dest   ZoneKind

	// until is the stop clause, asked with the cards that have LANDED
	// in dest so far, after each REPETITION. The run ends on the first
	// list it accepts.
	until func([]Card) bool

	// then is the caller's continuation, run once, with everything the
	// whole run landed.
	then func(g *Game, milled []uuid.UUID) error

	// repetitions caps how many times the process repeats, for a run
	// that named a number as well as a clause (n > 0 with an Until).
	// Zero — every card in the catalog — is "no limit but the
	// library".
	//
	// Repetitions and not cards, because a repetition is the
	// instruction now and a mill-amount replacement can make one of
	// them move two cards. No printed card uses it; a bound a card
	// prints is a clause (effects.UntilCount), not this.
	repetitions int

	// done counts the repetitions that have run.
	done int

	// landed is every card that has reached dest across the whole run,
	// in the order it arrived.
	landed []uuid.UUID

	// before is the run's immutable view of the cards it can mill:
	// every card in the library as the run BEGAN, by instance ID.
	//
	// It is what turns `landed` back into the []Card the clause takes,
	// and taking it once, up front, is what keeps the clause pure —
	// a live lookup would read a card that a later leg may have moved
	// again, and the replayed answer could differ from the first one.
	before map[uuid.UUID]Card
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
// #1176: with an `until` clause it is not one instruction at all. See
// millUntilRunLocked.
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
	if until != nil {
		// #1176: a run, not an instruction. It reports through `then`,
		// so the returned slice is empty exactly as the continuation
		// form's already is.
		return nil, g.millUntilRunLocked(&millRun{
			player:      playerID,
			dest:        dest,
			until:       until,
			then:        then,
			repetitions: n,
		})
	}
	ev := &ReplacementEvent{
		Kind:       RepEventMill,
		Actor:      playerID,
		MillPlayer: playerID,
		MillCount:  n,
		mill:       &millTail{dest: dest, then: then},
	}
	if !millAmountIsReplaceable(dest, n) {
		// Not a mill of a number: an exile of the top N, or a mill of
		// nothing. Straight to the plan, through the same body a
		// settled window reaches.
		return g.applyResolvedMillLocked(ev)
	}
	return g.runMillLocked(ev)
}

// millUntilRunLocked runs one repetition of an `until` run and hangs
// the next one off its continuation (#1176).
//
// # Why the run is a loop and not an instruction
//
// "Target opponent MILLS A CARD, then repeats this process until a
// creature card or X cards have been put into their graveyard this
// way, whichever comes first" (Helm of Obedience). The instruction the
// card gives is "mill a card", once, and the sentence repeats it. The
// engine used to model the whole run as one instruction that named no
// number, which made the bound right (Bruvac the Grandiloquent doubles
// mills, not bounds — #1161 pins it) and made the AMOUNT invisible:
// there was no number for a mill-amount replacement to double, so
// Helm + Bruvac at X=3 milled three cards where paper mills four.
//
// Each repetition is now its own one-card mill instruction, so it
// opens its own RepEventMill window and a mill-amount replacement
// rewrites it (CR 701.13b). Both cards of a doubled repetition are
// milled — one instruction, one simultaneous batch — and the clause is
// asked AFTER the repetition, which is where the card asks it. That is
// the whole of the paper numbers:
//
//	Helm X=1 + Bruvac   2 cards (one doubled repetition)
//	Helm X=2 + Bruvac   2 cards (the first repetition already reaches X)
//	Helm X=3 + Bruvac   4 cards (2 + 2, overshooting the bound by one)
//
// # What it does NOT reintroduce
//
// The over-mill #1177 closed by construction. A run that ends on what
// ARRIVED has to WAIT for each leg to arrive, and every repetition
// here goes through MillToZoneThenForEffect's sequencing body: a leg
// paused on the CR 903.9 prompt holds the rest of its own repetition
// and the whole run behind it, because the next repetition is started
// from the previous one's continuation and not from a loop that walks
// past it. The fire-and-forget entry point still cannot express an
// `until` at all, and now neither can a routing loop — the verdict is
// not inside one any more.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) millUntilRunLocked(run *millRun) error {
	p := g.playerByIDLocked(run.player)
	if p == nil {
		return ErrPlayerNotFound
	}
	if run.repetitions > 0 && run.done >= run.repetitions {
		return run.finish(g)
	}
	if len(p.Library.Cards) == 0 {
		// CR 701.13b: the run ends when the library does, with no error
		// and no loss. The caller's continuation still owes an answer.
		return run.finish(g)
	}
	if run.before == nil {
		run.before = make(map[uuid.UUID]Card, len(p.Library.Cards))
		for _, c := range p.Library.Cards {
			run.before[c.InstanceID] = c
		}
	}
	// One repetition IS an ordinary one-card mill: the same entry, the
	// same CR 614 window on its amount, the same plan, the same
	// sequencing routing loop. Nothing about a repetition knows it is
	// part of a run.
	_, err := g.millThroughReplacementsLocked(run.player, 1, run.dest, nil,
		func(g *Game, milled []uuid.UUID) error {
			next := run.next(milled)
			if next.stops() {
				return next.finish(g)
			}
			return g.millUntilRunLocked(next)
		})
	return err
}

// next is the run one repetition further on, as a NEW value with a new
// landed slice. Two runs of the same continuation — an undo, then the
// same answer again — must not see each other's entries.
func (r *millRun) next(milled []uuid.UUID) *millRun {
	out := *r
	out.done = r.done + 1
	out.landed = append(append(make([]uuid.UUID, 0, len(r.landed)+len(milled)), r.landed...), milled...)
	return &out
}

// stops asks the clause about everything that has LANDED so far
// (CR 400.7, landedInZoneLocked): a card the CR 614 window diverted —
// a commander taking the command zone, "if a card would be put into a
// graveyard from anywhere, exile it instead" — was never put into that
// graveyard, so it is not in the list and does not end the run.
//
// Pure: it reads its own immutable fields and nothing else.
func (r *millRun) stops() bool {
	if r.until == nil || len(r.landed) == 0 {
		return false
	}
	cards := make([]Card, 0, len(r.landed))
	for _, id := range r.landed {
		if c, ok := r.before[id]; ok {
			cards = append(cards, c)
		}
	}
	return len(cards) > 0 && r.until(cards)
}

// finish runs the caller's continuation with everything the run
// landed, once. A run that milled nothing still reports — a caller
// sequencing work behind it has to be told even when the answer is
// "none", which is the rule every terminal outcome of a routed move
// follows.
func (r *millRun) finish(g *Game) error {
	if r.then == nil {
		return nil
	}
	then := r.then
	r.then = nil
	return then(g, r.landed)
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
	if tail.then == nil {
		// Fire-and-forget: every leg is routed on this line and one
		// that pauses on CR 903.9 lands later without holding the rest
		// of the mill up (#529).
		return g.routeAllLandedLocked(r, ids), nil
	}
	// Cleared THROUGH the pointer, for the reason the token and
	// keyword-action tails are: a continuation that re-enters the
	// pipeline on the same tail value must not run itself twice, and an
	// undo snapshot therefore needs its own copy of the tail (which
	// cloneReplacementResume gives it).
	then := tail.then
	tail.then = nil
	return nil, g.routeAllThenLocked(r, ids, then)
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
