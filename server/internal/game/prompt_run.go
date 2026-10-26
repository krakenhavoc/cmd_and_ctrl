package game

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// prompt_run.go — the SHARED body of a prompted run (#1019, #1027,
// ADR 0013 §5x / §5y).
//
// A RUN is one printed instruction in flight, however many prompts it
// takes to ask it: "each player sacrifices a creature of their
// choice" over four seats, "sacrifice two lands" over one seat twice,
// "each opponent discards a card" over three. Its continuation runs
// ONCE, after the LAST of those prompts has settled AND the cards
// they named have finished moving, and is handed one entry per seat
// the run ASKED.
//
// TWO VERBS, ONE BODY. #1019 built the shape for a prompted
// SACRIFICE; #1027 needed exactly the same thing for a prompted
// DISCARD — Archon of Cruelty's "sacrifices …, discards a card, and
// loses 3 life" is one nested chain, and the discard half had no
// continuation to nest into. Everything about a run except WHICH
// prompt it queues and what its answer is CALLED is the same in both
// verbs: the outstanding counter, the ask order, the per-seat landed
// lists, the settle-once rule, the deep copy an undo needs. So it is
// one struct and one settle rather than two near-identical files —
// the catalog's clone gate exists because the second copy is how the
// first one rots.
//
// What each verb keeps of its own is only the vocabulary:
// sacrifice_run.go holds the four sacrifice entry points and
// PromptedSacrifices, discard_run.go holds the three discard entry
// points and PromptedDiscards. Both are thin.
//
// WHY THE STATE LIVES ON THE GAME rather than in a frame on the
// prompt. The prompts of one run are answered in any order, so the
// run has to be SHARED by all of them and MUTATED as each settles —
// and a server-only frame shared by pointer is copied per prompt by
// cloneLocked, which would give an undo snapshot as many half-finished
// runs as it had prompts. Keyed by a plain uuid on the choice instead,
// the link survives a value copy for free and cloneLocked deep-copies
// the runs once: the counter and the landed lists rewind together with
// the queue, so an undone answer replays to the same place. It is the
// shape Game.replacementsAppliedThisEvent already has, for the same
// reason (#808).

// SeatCards is what ONE seat asked by a prompted run actually moved:
// the cards that really left the zone the instruction named, in the
// order that seat answered for them.
//
// `Cards` is empty for a seat that was asked and moved nothing — its
// prompt was withdrawn because its material vanished under it, or the
// CR 614 window cancelled the move outright. Being told about a seat
// that moved nothing is the point: "if you sacrificed a creature this
// way" needs the difference between "asked and did not" and "never
// asked".
type SeatCards struct {
	Seat  uuid.UUID
	Cards []uuid.UUID
}

// seatCardsBy returns the cards `seat` moved this way, or nil. The one
// body behind PromptedSacrifices.By and PromptedDiscards.By.
func seatCardsBy(in []SeatCards, seat uuid.UUID) []uuid.UUID {
	for _, e := range in {
		if e.Seat == seat {
			return e.Cards
		}
	}
	return nil
}

// seatCardsFlat is every card moved this way, in ask order — "for each
// permanent sacrificed this way", "for each card discarded this way".
func seatCardsFlat(in []SeatCards) []uuid.UUID {
	var out []uuid.UUID
	for _, e := range in {
		out = append(out, e.Cards...)
	}
	return out
}

// seatCardsCount is how many cards moved this way, across every seat.
func seatCardsCount(in []SeatCards) int {
	n := 0
	for _, e := range in {
		n += len(e.Cards)
	}
	return n
}

// promptRun is one printed instruction in flight: the prompts still
// outstanding, what has landed so far, and the rest of the card.
//
// `then` is a closure and is SHARED with a clone like every other
// continuation in the engine — a resume reads it and never writes it.
// Everything else is deep-copied (clonePromptRuns), because everything
// else is what an answer changes.
type promptRun struct {
	// outstanding is how many queued prompts have not settled yet.
	// The run finishes at zero; it is never re-armed, because a
	// continuation that queued more prompts starts a run of its own.
	outstanding int

	// seats is the DISTINCT seats this run asked, in ask order. It is
	// what makes the answer deterministic: `landed` is a map, and a
	// continuation that walked it would decide differently on two
	// runs of the same game.
	seats []uuid.UUID

	// landed is what each seat moved, accumulated as the legs settle.
	landed map[uuid.UUID][]uuid.UUID

	// then is the rest of the card. Nil is a run nobody is waiting on,
	// which is every caller that used the fire-and-forget entry points.
	then func(g *Game, landed []SeatCards) error

	// thenKey is the rest of the card as a REGISTERED continuation
	// (RegisterPromptRunThen, prompt_run_keyed.go, #4290) in place of
	// `then`, and carry and source are the plain data it is handed. A
	// run with a key holds no closure, so it is written to a restore
	// point (GameSnapshot.PromptRuns) and the table waiting on its legs
	// is one. At most one of `then` and `thenKey` is set.
	thenKey string
	carry   []uuid.UUID
	source  uuid.UUID
	// nums and asked are the rest of a keyed run's plain data (RunData):
	// the numbers the continuation reads, and the stack item that asked as
	// data (NewAskedItem), for a continuation that reads its targets, its X
	// or its text.
	nums  []int
	asked *AskedItem
}

// answer builds the continuation's argument: the ask order, with each
// seat's landed list.
func (r *promptRun) answer() []SeatCards {
	out := make([]SeatCards, 0, len(r.seats))
	for _, seat := range r.seats {
		out = append(out, SeatCards{Seat: seat, Cards: r.landed[seat]})
	}
	return out
}

// runPromptsLocked is the ONE body behind every prompted-run entry
// point of every verb: queue a prompt per entry in `asks` (the same
// seat may appear more than once), and remember what the run still
// owes.
//
// `queue` puts ONE prompt up for one seat, stamped with the run id it
// belongs to, and reports whether it went up at all — a seat with
// nothing the instruction can take is skipped rather than handed an
// empty question (CR 701.21a's "if you can" for a sacrifice, CR
// 701.9a's "as many as you can" for a discard). It is the only thing
// the two verbs do differently at queue time.
//
// It reports how many prompts it queued, which is what the
// fire-and-forget forms return, and runs `then` inline when it queued
// none — there is nothing to wait for and the rest of the card is
// still owed.
//
// Caller must hold g.mu.
func (g *Game) runPromptsLocked(
	asks []uuid.UUID,
	then func(g *Game, landed []SeatCards) error,
	queue func(seat, run uuid.UUID) bool,
) (int, error) {
	return g.startRunLocked(asks, &promptRun{then: then}, queue)
}

// runPromptsKeyedLocked is runPromptsLocked with the rest of the card
// as a registered continuation (#4290): `then` is looked up by its key
// when the last leg settles, and handed `source` and `carry` with the
// answer. The run is plain data, so a table waiting on its legs is a
// restore point when the legs are too.
//
// Caller must hold g.mu.
func (g *Game) runPromptsKeyedLocked(
	asks []uuid.UUID,
	then PromptRunThen,
	source uuid.UUID,
	data RunData,
	queue func(seat, run uuid.UUID) bool,
) (int, error) {
	if then.key == "" {
		return 0, fmt.Errorf("%w: keyed prompted run with no continuation", ErrInvalidParam)
	}
	return g.startRunLocked(asks, &promptRun{
		thenKey: then.key, source: source,
		carry: copyUUIDs(data.Carry), nums: copyInts(data.Nums), asked: copyAskedItem(data.Asked),
	}, queue)
}

// startRunLocked is the body of both: `run` arrives with its
// continuation (a closure, a key, or neither) and leaves with its legs.
//
// Caller must hold g.mu.
func (g *Game) startRunLocked(asks []uuid.UUID, run *promptRun, queue func(seat, run uuid.UUID) bool) (int, error) {
	run.landed = map[uuid.UUID][]uuid.UUID{}
	id := uuid.New()
	if !run.waited() {
		// Nothing is waiting, so there is no run to keep: the prompts
		// go up unlinked and settle into nobody. Keeping one would be
		// bookkeeping for a question nobody asked about, and it would
		// put a continuation frame on the census for every Grave Pact.
		id = uuid.Nil
	}
	for _, seat := range asks {
		if !queue(seat, id) {
			continue
		}
		if !slices.Contains(run.seats, seat) {
			run.seats = append(run.seats, seat)
		}
		run.outstanding++
	}
	if run.outstanding == 0 {
		return 0, g.finishRunLocked(run)
	}
	if id == uuid.Nil {
		return run.outstanding, nil
	}
	if g.promptRuns == nil {
		g.promptRuns = map[uuid.UUID]*promptRun{}
	}
	g.promptRuns[id] = run
	return run.outstanding, nil
}

// settleRunLegLocked records what ONE prompt of a run produced and
// runs the run's continuation when it was the last one outstanding.
//
// THE one place a leg is settled, for either verb, reached from
// exactly two kinds of moment: the prompt was ANSWERED and the cards
// it named have finished moving (ResolveSacrificeChoice, the discard
// prompt's own continuation), or the prompt was WITHDRAWN unanswered
// and moved nothing (the departure table's dropDefault action,
// runChoiceDropActionLocked).
//
// `landed` is the "this way" answer for the cards the seat named, so a
// sacrificed commander that took the command zone IS in it — it was
// sacrificed, and only where the card went was replaced — a madness
// card that went to exile instead of a graveyard IS in it, and a leg
// the CR 614 window cancelled is not.
//
// A leg whose run has already finished, or that never had one, is a
// no-op: settling twice must not pay out twice, which is what makes
// an undone-then-replayed answer land where answering once would.
//
// Caller must hold g.mu.
func (g *Game) settleRunLegLocked(runID, seat uuid.UUID, landed []uuid.UUID) error {
	run := g.promptRuns[runID]
	if run == nil {
		return nil
	}
	if len(landed) > 0 {
		// A fresh slice rather than an append in place, the reason
		// routeEachStepLocked gives: two runs of the same
		// continuation (an undo, then the same answer again) must not
		// see each other's entry.
		prev := run.landed[seat]
		next := make([]uuid.UUID, 0, len(prev)+len(landed))
		next = append(append(next, prev...), landed...)
		run.landed[seat] = next
	}
	run.outstanding--
	if run.outstanding > 0 {
		return nil
	}
	delete(g.promptRuns, runID)
	if len(g.promptRuns) == 0 {
		g.promptRuns = nil
	}
	return g.finishRunLocked(run)
}

// waited reports whether anything is waiting on the run: a closure or a
// registered continuation.
func (r *promptRun) waited() bool { return r.then != nil || r.thenKey != "" }

// finishRunLocked runs the rest of the card once the run's last leg has
// settled (or when no leg went up): the closure with the answer, or the
// registered continuation with the answer and the run's plain data.
//
// Caller must hold g.mu.
func (g *Game) finishRunLocked(run *promptRun) error {
	// No seat was asked: the continuation is handed nothing, as the
	// closure form always was.
	var answer []SeatCards
	if len(run.seats) > 0 {
		answer = run.answer()
	}
	switch {
	case run.then != nil:
		return run.then(g, answer)
	case run.thenKey != "":
		fn, ok := lookupPromptRunThen(run.thenKey)
		if !ok {
			return fmt.Errorf("%w: prompted-run continuation %q", ErrUnknownEffectKey, run.thenKey)
		}
		return fn(g, RunDone{
			Source: run.source, Landed: answer,
			Carry: copyUUIDs(run.carry), Nums: copyInts(run.nums), Asked: copyAskedItem(run.asked),
		})
	}
	return nil
}

// clonePromptRuns gives an undo snapshot its own copy of every run in
// flight.
//
// Deep in everything an answer writes — the counter, the per-seat
// landed lists, the ask order — and shallow in the one thing it does
// not, the continuation closure. That is the same split
// cloneReplacementResume makes, and it is what makes an undo across a
// half-answered fan-out replay identically: restore the queue and the
// runs together, and the second answer finds exactly the run the
// first one did.
func clonePromptRuns(in map[uuid.UUID]*promptRun) map[uuid.UUID]*promptRun {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]*promptRun, len(in))
	for id, run := range in {
		if run == nil {
			continue
		}
		cp := *run
		cp.seats = append([]uuid.UUID(nil), run.seats...)
		cp.carry = copyUUIDs(run.carry)
		cp.nums = copyInts(run.nums)
		cp.asked = copyAskedItem(run.asked)
		cp.landed = make(map[uuid.UUID][]uuid.UUID, len(run.landed))
		for seat, cards := range run.landed {
			cp.landed[seat] = append([]uuid.UUID(nil), cards...)
		}
		out[id] = &cp
	}
	return out
}
