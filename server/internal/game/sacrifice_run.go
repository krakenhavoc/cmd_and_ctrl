package game

import (
	"slices"

	"github.com/google/uuid"
)

// sacrifice_run.go — the CONTINUATION of a prompted sacrifice
// (#1019, CR 701.17a, ADR 0013 §5x).
//
// PlayerSacrificesForEffect and EachPlayerSacrificesForEffect do not
// sacrifice anything. They queue a QUESTION per seat and return how
// many seats were asked; the permanent leaves the battlefield when a
// player answers, which is one or more actions later. So every clause
// a card wrote after one of them was a payout on a move that had not
// happened — the #911 / #993 shape arriving through a prompt rather
// than through a fire-and-forget exit, which is exactly why ADR 0013
// §5v left both entry points out of the payout lint's tables: the
// message a lint prints has to name a fix, and there was no
// continuation form to name.
//
// A RUN is that form. One run is one printed instruction — "each
// player sacrifices a creature of their choice", "sacrifice two
// lands" — however many prompts it takes to ask it. Its continuation
// runs ONCE, after the LAST of those prompts has settled, and is told
// per seat which permanents really left the battlefield.
//
// Three properties, and each of them is a bug that was live before:
//
//   - It waits for every asked seat. Rise of the Witch-king's "if you
//     sacrificed a creature this way" gated on "was the controller
//     handed a prompt", so it returned the permanent before anybody
//     had chosen anything.
//   - It waits for the MOVE, not for the answer. The picked permanent
//     goes through sacrificeRoute (#910/#964) like every other
//     sacrifice, so a sacrificed commander's CR 903.9 prompt pauses
//     the leg and the run waits for that too. A run leg that is still
//     paused is a run that has not finished.
//   - A prompt the engine WITHDRAWS settles the leg with nothing
//     sacrificed rather than stranding the run. That is #1016's
//     dropDefault rule at a second kind: the question ends, the rest
//     of the card does not (settleSacrificeRunLegLocked, reached
//     through the departure table).
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

// SeatSacrifice is what ONE seat asked by a prompted sacrifice run
// actually sacrificed: the permanents that LEFT THE BATTLEFIELD, in
// the order that seat answered for them.
//
// `Cards` is empty for a seat that was asked and sacrificed nothing —
// its prompt was withdrawn because the board emptied under it, its
// only legal permanent left while another seat was being asked, or
// the CR 614 window cancelled the move outright. Being told about a
// seat that sacrificed nothing is the point: "if you sacrificed a
// creature this way" needs the difference between "asked and did not"
// and "never asked".
type SeatSacrifice struct {
	Seat  uuid.UUID
	Cards []uuid.UUID
}

// PromptedSacrifices is what a prompted sacrifice's continuation is
// handed: one entry per seat the run ASKED, in the order the seats
// were asked (APNAP from the active player for a fan-out).
//
// A seat that was skipped at queue time — it controlled no permanent
// the effect's spec admits, so CR 701.21a's "if you can" excused it —
// has no entry at all. `Sacrificed` reads the same for it as for a
// seat that was asked and answered with nothing, which is right: both
// sacrificed nothing. A caller that needs to tell them apart is
// asking about the QUESTION rather than about the sacrifice, and
// should not be reading this.
type PromptedSacrifices []SeatSacrifice

// By returns the permanents `seat` sacrificed this way, or nil.
func (s PromptedSacrifices) By(seat uuid.UUID) []uuid.UUID {
	for _, e := range s {
		if e.Seat == seat {
			return e.Cards
		}
	}
	return nil
}

// Sacrificed reports whether `seat` sacrificed at least one permanent
// this way — Rise of the Witch-king's "if you sacrificed a creature
// this way", written as the card prints it.
func (s PromptedSacrifices) Sacrificed(seat uuid.UUID) bool {
	return len(s.By(seat)) > 0
}

// Cards flattens the run into every permanent sacrificed this way, in
// ask order — "for each permanent sacrificed this way".
func (s PromptedSacrifices) Cards() []uuid.UUID {
	var out []uuid.UUID
	for _, e := range s {
		out = append(out, e.Cards...)
	}
	return out
}

// Count is how many permanents were sacrificed this way, across every
// seat. Lich-Knights' Conquest returns that many creature cards.
func (s PromptedSacrifices) Count() int {
	n := 0
	for _, e := range s {
		n += len(e.Cards)
	}
	return n
}

// sacrificeRun is one printed "sacrifice" instruction in flight: the
// prompts still outstanding, what has landed so far, and the rest of
// the card.
//
// `then` is a closure and is SHARED with a clone like every other
// continuation in the engine — a resume reads it and never writes it.
// Everything else is deep-copied (cloneSacrificeRuns), because
// everything else is what an answer changes.
type sacrificeRun struct {
	// outstanding is how many queued prompts have not settled yet.
	// The run finishes at zero; it is never re-armed, because a
	// continuation that queued more sacrifices starts a run of its
	// own.
	outstanding int

	// seats is the DISTINCT seats this run asked, in ask order. It is
	// what makes the answer deterministic: `landed` is a map, and a
	// continuation that walked it would decide differently on two
	// runs of the same game.
	seats []uuid.UUID

	// landed is what each seat sacrificed, accumulated as the legs
	// settle.
	landed map[uuid.UUID][]uuid.UUID

	// then is the rest of the card. Nil is a run nobody is waiting on,
	// which is every caller that used the fire-and-forget entry points.
	then func(g *Game, sacrificed PromptedSacrifices) error
}

// answer builds the continuation's argument: the ask order, with each
// seat's landed list.
func (r *sacrificeRun) answer() PromptedSacrifices {
	out := make(PromptedSacrifices, 0, len(r.seats))
	for _, seat := range r.seats {
		out = append(out, SeatSacrifice{Seat: seat, Cards: r.landed[seat]})
	}
	return out
}

// EachPlayerSacrificesThenForEffect is EachPlayerSacrificesForEffect
// with the rest of the card attached: "each player sacrifices a
// creature of their choice. If you sacrificed a creature this way, …"
// (Rise of the Witch-king).
//
// The prompts are queued exactly as the fire-and-forget form queues
// them — one per affected seat, APNAP from the active player, a seat
// with no legal permanent skipped — and `then` runs once every one of
// them has settled AND the permanents they named have finished moving.
//
// `then` runs even when nothing was sacrificed, including when nobody
// could be asked: a continuation is the rest of a card that is paused
// mid-resolution (#544, #1006), and one that is silently never called
// is a card that stops halfway. It is handed a PromptedSacrifices
// whose entries are the seats that were ASKED.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) EachPlayerSacrificesThenForEffect(
	source, except uuid.UUID,
	spec *TargetSpec,
	reason string,
	then func(g *Game, sacrificed PromptedSacrifices) error,
) error {
	var seats []uuid.UUID
	if n := len(g.Seats); n > 0 {
		start := g.Turn.ActiveSeat
		for i := 0; i < n; i++ {
			p := g.Seats[(start+i)%n]
			if p == nil || p.Eliminated || p.ID == except {
				continue
			}
			seats = append(seats, p.ID)
		}
	}
	_, err := g.sacrificeRunLocked(source, seats, spec, reason, then)
	return err
}

// PlayersSacrificeThenForEffect is the GENERAL form: one prompt per
// entry in `players`, in the order given, with one continuation over
// the lot — "each of two target opponents loses 2 life and sacrifices
// a creature. You add {B}{B} and draw a card" (Priest of Forgotten
// Gods), where the seats are the spell's targets rather than the whole
// table.
//
// The other two entry points are this one with their seat list worked
// out: EachPlayerSacrificesThenForEffect walks APNAP,
// PlayerSacrificesThenForEffect repeats one seat. A caller that names
// the same seat twice gets two prompts and one entry in the answer.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) PlayersSacrificeThenForEffect(
	source uuid.UUID,
	players []uuid.UUID,
	spec *TargetSpec,
	reason string,
	then func(g *Game, sacrificed PromptedSacrifices) error,
) error {
	_, err := g.sacrificeRunLocked(source, players, spec, reason, then)
	return err
}

// PlayerSacrificesThenForEffect is PlayerSacrificesForEffect with the
// rest of the card attached, and with a COUNT: "sacrifice two lands"
// (Lotus Field), "sacrifice that many permanents of their choice"
// (Phyrexian Obliterator), "sacrifice an artifact, enchantment or
// token" N times (Lich-Knights' Conquest).
//
// The count is here rather than in the caller's loop because the
// continuation belongs to the whole instruction: N prompts are ONE
// run, and `then` runs after the last of them. A caller that looped
// this entry point N times would start N runs and pay out N times.
//
// `count` below 1 is treated as 1. A seat that has nothing left to
// sacrifice stops the queueing early, which is what the fire-and-
// forget form's `== 0` return meant; a seat with fewer permanents
// than owed is still asked for all of them, and the surplus prompts
// are withdrawn by pruneSacrificeChoicesLocked as the board empties —
// each withdrawal settling its leg with nothing.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) PlayerSacrificesThenForEffect(
	source, playerID uuid.UUID,
	spec *TargetSpec,
	reason string,
	count int,
	then func(g *Game, sacrificed PromptedSacrifices) error,
) error {
	_, err := g.sacrificeRunLocked(source, repeatSeat(playerID, count), spec, reason, then)
	return err
}

// PlayerSacrificesNForEffect queues `count` sacrifice prompts for one
// player with NOTHING waiting on them — Phyrexian Obliterator's "that
// many permanents of their choice", Lotus Field's two lands,
// Archfiend of Depravity's "sacrifice all but two".
//
// PlayerSacrificesThenForEffect without the continuation, and the loop
// those three helpers used to write out: ask `count` times, stop when
// the seat has nothing the spec admits. The stop rule is here rather
// than in three callers because it is one rule (CR 701.21a's "if you
// can"), and because a caller writing the loop by hand has to read the
// count of QUESTIONS to know when to break — which is the one thing no
// clause may be gated on (#1019).
//
// Returns the number of prompts queued.
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) PlayerSacrificesNForEffect(source, playerID uuid.UUID, spec *TargetSpec, reason string, count int) int {
	queued, _ := g.sacrificeRunLocked(source, repeatSeat(playerID, count), spec, reason, nil)
	return queued
}

// repeatSeat is one seat asked `count` times, with `count` below 1
// read as once — the ask list both single-seat entry points build.
func repeatSeat(playerID uuid.UUID, count int) []uuid.UUID {
	if count < 1 {
		count = 1
	}
	seats := make([]uuid.UUID, 0, count)
	for i := 0; i < count; i++ {
		seats = append(seats, playerID)
	}
	return seats
}

// sacrificeRunLocked is the ONE body behind all four entry points:
// queue a prompt per entry in `asks` (the same seat may appear more
// than once), and remember what the run still owes.
//
// It reports how many prompts it queued, which is what the
// fire-and-forget forms return, and runs `then` inline when it queued
// none — there is nothing to wait for and the rest of the card is
// still owed.
//
// Caller must hold g.mu.
func (g *Game) sacrificeRunLocked(
	source uuid.UUID,
	asks []uuid.UUID,
	spec *TargetSpec,
	reason string,
	then func(g *Game, sacrificed PromptedSacrifices) error,
) (int, error) {
	run := &sacrificeRun{landed: map[uuid.UUID][]uuid.UUID{}, then: then}
	id := uuid.New()
	if then == nil {
		// Nothing is waiting, so there is no run to keep: the prompts
		// go up unlinked and settle into nobody. Keeping one would be
		// bookkeeping for a question nobody asked about, and it would
		// put a continuation frame on the census for every Grave Pact.
		id = uuid.Nil
	}
	for _, seat := range asks {
		if !g.queueSacrificePromptLocked(source, seat, spec, reason, id) {
			// CR 701.21a's "if you can": this seat controls nothing
			// the effect admits, so it is skipped rather than handed
			// an empty list — which is the fire-and-forget form's
			// rule, unchanged. A repeated ask of the same seat fails
			// here for the same reason every time, since nothing
			// moves between two queued prompts.
			continue
		}
		if !slices.Contains(run.seats, seat) {
			run.seats = append(run.seats, seat)
		}
		run.outstanding++
	}
	if run.outstanding == 0 {
		if then == nil {
			return 0, nil
		}
		return 0, then(g, nil)
	}
	if id == uuid.Nil {
		return run.outstanding, nil
	}
	if g.sacrificeRuns == nil {
		g.sacrificeRuns = map[uuid.UUID]*sacrificeRun{}
	}
	g.sacrificeRuns[id] = run
	return run.outstanding, nil
}

// queueSacrificePromptLocked queues ONE sacrifice prompt and reports
// whether it went up. `run` is the run the prompt belongs to, or
// uuid.Nil for a prompt nothing is waiting on.
//
// THE one place a PendingChoiceSacrifice is created, so the run link
// cannot be forgotten by a new caller.
//
// Caller must hold g.mu.
func (g *Game) queueSacrificePromptLocked(source, playerID uuid.UUID, spec *TargetSpec, reason string, run uuid.UUID) bool {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Eliminated {
		return false
	}
	options := g.sacrificeCandidatesLocked(playerID, spec)
	if len(options) == 0 {
		return false
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:             PendingChoiceSacrifice,
		Chooser:          playerID,
		FromPlayer:       playerID,
		Count:            1,
		Source:           source,
		Reason:           reason,
		SacrificeOptions: options,
		sacrificeRun:     run,
	}) != uuid.Nil
}

// settleSacrificeRunLegLocked records what ONE prompt of a run
// produced and runs the run's continuation when it was the last one
// outstanding.
//
// THE one place a leg is settled, reached from exactly two kinds of
// moment: the prompt was ANSWERED and the permanent it named has
// finished moving (ResolveSacrificeChoice), or the prompt was
// WITHDRAWN unanswered and sacrificed nothing (the departure table's
// dropDefault action, runChoiceDropActionLocked).
//
// `landed` is sacrificedThisWayLocked's answer for the permanent the
// seat named, so a commander that took the command zone IS in it —
// it was sacrificed, and only where the card went was replaced — and
// a leg the CR 614 window cancelled is not.
//
// A leg whose run has already finished, or that never had one, is a
// no-op: settling twice must not pay out twice, which is what makes
// an undone-then-replayed answer land where answering once would.
//
// Caller must hold g.mu.
func (g *Game) settleSacrificeRunLegLocked(runID, seat uuid.UUID, landed []uuid.UUID) error {
	run := g.sacrificeRuns[runID]
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
	delete(g.sacrificeRuns, runID)
	if len(g.sacrificeRuns) == 0 {
		g.sacrificeRuns = nil
	}
	if run.then == nil {
		return nil
	}
	return run.then(g, run.answer())
}

// cloneSacrificeRuns gives an undo snapshot its own copy of every run
// in flight.
//
// Deep in everything an answer writes — the counter, the per-seat
// landed lists, the ask order — and shallow in the one thing it does
// not, the continuation closure. That is the same split
// cloneReplacementResume makes, and it is what makes an undo across a
// half-answered fan-out replay identically: restore the queue and the
// runs together, and the second answer finds exactly the run the
// first one did.
func cloneSacrificeRuns(in map[uuid.UUID]*sacrificeRun) map[uuid.UUID]*sacrificeRun {
	if len(in) == 0 {
		return nil
	}
	out := make(map[uuid.UUID]*sacrificeRun, len(in))
	for id, run := range in {
		if run == nil {
			continue
		}
		cp := *run
		cp.seats = append([]uuid.UUID(nil), run.seats...)
		cp.landed = make(map[uuid.UUID][]uuid.UUID, len(run.landed))
		for seat, cards := range run.landed {
			cp.landed[seat] = append([]uuid.UUID(nil), cards...)
		}
		out[id] = &cp
	}
	return out
}
