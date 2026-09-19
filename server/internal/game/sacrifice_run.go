package game

import (
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
// A RUN is that form, and since #1027 the run itself is shared with
// the prompted DISCARD (prompt_run.go, ADR 0013 §5y) — this file is
// the four sacrifice entry points, the sacrifice's own "this way"
// reading, and the vocabulary a card writes its clause in. One run is
// one printed instruction — "each player sacrifices a creature of
// their choice", "sacrifice two lands" — however many prompts it
// takes to ask it. Its continuation runs ONCE, after the LAST of
// those prompts has settled, and is told per seat which permanents
// really left the battlefield.
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
//     of the card does not (settleRunLegLocked, reached through the
//     departure table).

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
type PromptedSacrifices []SeatCards

// By returns the permanents `seat` sacrificed this way, or nil.
func (s PromptedSacrifices) By(seat uuid.UUID) []uuid.UUID { return seatCardsBy(s, seat) }

// Sacrificed reports whether `seat` sacrificed at least one permanent
// this way — Rise of the Witch-king's "if you sacrificed a creature
// this way", written as the card prints it.
func (s PromptedSacrifices) Sacrificed(seat uuid.UUID) bool { return len(seatCardsBy(s, seat)) > 0 }

// Cards flattens the run into every permanent sacrificed this way, in
// ask order — "for each permanent sacrificed this way".
func (s PromptedSacrifices) Cards() []uuid.UUID { return seatCardsFlat(s) }

// Count is how many permanents were sacrificed this way, across every
// seat. Lich-Knights' Conquest returns that many creature cards.
func (s PromptedSacrifices) Count() int { return seatCardsCount(s) }

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

// sacrificeRunLocked is the sacrifice verb's half of a prompted run:
// the shared body (runPromptsLocked) with the sacrifice prompt as its
// queue step and PromptedSacrifices as the name its continuation reads
// the answer under.
//
// Caller must hold g.mu.
func (g *Game) sacrificeRunLocked(
	source uuid.UUID,
	asks []uuid.UUID,
	spec *TargetSpec,
	reason string,
	then func(g *Game, sacrificed PromptedSacrifices) error,
) (int, error) {
	var landedThen func(*Game, []SeatCards) error
	if then != nil {
		landedThen = func(g *Game, landed []SeatCards) error {
			return then(g, PromptedSacrifices(landed))
		}
	}
	return g.runPromptsLocked(asks, landedThen, func(seat, run uuid.UUID) bool {
		// CR 701.21a's "if you can": a seat that controls nothing the
		// effect admits is skipped rather than handed an empty list —
		// which is the fire-and-forget form's rule, unchanged. A
		// repeated ask of the same seat fails here for the same
		// reason every time, since nothing moves between two queued
		// prompts.
		return g.queueSacrificePromptLocked(source, seat, spec, reason, run)
	})
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
		promptRun:        run,
	}) != uuid.Nil
}
