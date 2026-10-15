package game

import (
	"errors"

	"github.com/google/uuid"
)

// turn_skip_order.go — asking which skip takes an extra turn (ADR 0188
// Decision 12, CR 616.1, #3561, tracker #753).
//
// A player who owes "skip your next turn" (turn_skip.go) skips the next
// turn they would begin. When that turn is an extra turn and an
// extra-turn skip (Trouble in Pairs, Ugin's Nexus, Gerrard's Hourglass
// Pendant) applies to it as well, both are replacement effects on one
// event, and CR 616.1 lets the affected player choose which applies.
// Every answer skips the turn. The question is only which effect is used
// up, and so whether the player still owes a skip afterwards.
//
// It is asked only when both apply (popExtraTurnLocked, extra_turns.go),
// and it is the one place the rotation seam can pause:
//
//   - The old turn has closed (closeTurnLocked) and the candidate extra
//     turn is back on top of Game.ExtraTurns. Game.TurnChangePending is
//     set, the cursor stays where the old turn left it with nobody
//     holding priority, and nothing has begun: no Seq, no untap, no
//     triggers.
//   - The answer is the resume (ResolveTurnSkipOrder). It settles the
//     candidate, clears the flag and chooses again, which may pause once
//     more for the next candidate (a second Time Stretch turn), and runs
//     the step entry hooks when it does not.
//   - The prompt is plain data, so a table waiting on it is a restore
//     point, an undo of the answer reopens it, and an older binary
//     refuses the file (unknown pending-choice kind).
//
// "Use my skip" (answer 0, `apply: true` on the wire) is recommended.
// The heuristic, the MCP seat's suggestion, an ADR 0127 standing
// "Always" and a seat that leaves all take it.

// PendingChoiceTurnSkipOrder is the which-skip question (ADR 0188
// Decision 12). Chooser is the extra turn's player (CR 616.1); the
// details ride PendingChoice.TurnSkipOrder. Answered by
// ResolveTurnSkipOrder: true uses the player's own skip, false lets the
// extra-turn skip take the turn.
const PendingChoiceTurnSkipOrder PendingChoiceKind = "turn_skip_order"

// TurnSkipOrderAutoAnswerKey is the fixed ADR 0127 key the question is
// filed under: the one exception to §2's "no catalog row, no key",
// because the question is the engine's, not a card's. Always is answer
// 0 (use my skip), Never answer 1.
const TurnSkipOrderAutoAnswerKey = "engine|turn_skip_order"

// ErrTurnChangePending refuses a cursor verb (pass_turn, advance_step,
// pass_priority, end_turn) while the turn change waits on the which-skip
// question: there is no turn in progress to move through.
var ErrTurnChangePending = errors.New("game: the next turn is waiting on a question")

// TurnSkipOrder is the which-skip question's plain data
// (PendingChoice.TurnSkipOrder).
type TurnSkipOrder struct {
	// ExtraTurnRef is the extra turn the question is about: the entry on
	// top of Game.ExtraTurns while it is open.
	ExtraTurnRef int `json:"extraTurnRef"`
	// SkipSource is the source of the player's oldest skip record, the
	// one answer 0 spends, and SkipSeq that record's Seq.
	SkipSource uuid.UUID `json:"skipSource"`
	SkipSeq    int64     `json:"skipSeq"`
	// SkipName is SkipSource's name when the question was asked, for the
	// prompt's words: the card may have left by the time it is read.
	SkipName string `json:"skipName,omitempty"`
	// Others names the extra-turn skips that apply (their sources'
	// names), for answer 1's words.
	Others []string `json:"others,omitempty"`
}

func cloneTurnSkipOrder(in *TurnSkipOrder) *TurnSkipOrder {
	if in == nil {
		return nil
	}
	out := *in
	out.Others = append([]string(nil), in.Others...)
	return &out
}

// oldestTurnSkipLocked is the record consumeTurnSkipLocked would spend
// for `seat`, without spending it. ok is false when the seat owes none.
//
// Caller must hold g.mu.
func (g *Game) oldestTurnSkipLocked(seat int) (seq int64, source uuid.UUID, ok bool) {
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return 0, uuid.Nil, false
	}
	id := g.Seats[seat].ID
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModSkipTurns || m.Player != id || m.Amount < 1 {
				continue
			}
			if !ok || e.Seq < seq {
				seq, source, ok = e.Seq, e.Source.ID, true
			}
			break
		}
	}
	return seq, source, ok
}

// consumeTurnSkipSeqLocked spends one charge of the skip record `seq`
// for `seat`, or of the oldest record when that one is gone (ADR 0188
// Decision 12's answer 0).
//
// Caller must hold g.mu (write).
func (g *Game) consumeTurnSkipSeqLocked(seat int, seq int64) (uuid.UUID, bool) {
	if seat >= 0 && seat < len(g.Seats) && g.Seats[seat] != nil {
		id := g.Seats[seat].ID
		for i := range g.ScopedEffects {
			e := g.ScopedEffects[i]
			if e.Seq != seq {
				continue
			}
			for j, m := range e.Mods {
				if m.Kind == ModSkipTurns && m.Player == id && m.Amount > 0 {
					g.setShieldChargeLocked(e.Seq, j, m.Amount-1)
					return e.Source.ID, true
				}
			}
		}
	}
	return g.consumeTurnSkipLocked(seat)
}

// extraTurnSkipAppliesLocked reports which extra-turn skips would apply
// to `et` as it began, by building the RepEventExtraTurn event the
// window would and gathering the replacements that apply to it, applying
// nothing (ADR 0188 Decision 12). The result names each one by its
// source, or by its label for a source-less effect. Empty when none
// applies.
//
// Caller must hold g.mu.
func (g *Game) extraTurnSkipAppliesLocked(et ExtraTurn) []string {
	if et.Seat < 0 || et.Seat >= len(g.Seats) || g.Seats[et.Seat] == nil {
		return nil
	}
	ev := &ReplacementEvent{
		Kind:          RepEventExtraTurn,
		Actor:         g.Seats[et.Seat].ID,
		Source:        et.Source,
		ExtraTurnSeat: et.Seat,
		ExtraTurnRef:  et.Ref,
		mustSettleNow: true,
	}
	var names []string
	for _, ar := range g.gatherActiveReplacementsLocked(ev) {
		name := ar.effect.Label
		if ar.source != nil && ar.source.Name != "" {
			name = ar.source.Name
		}
		names = append(names, name)
	}
	return names
}

// pauseForTurnSkipOrderLocked pauses the turn change over `et` (ADR 0188
// Decision 12, case 3): the entry goes back on top of the queue, the
// flag is set, and the question is queued to the turn's player. Reports
// false, changing nothing, when the question cannot be asked — the
// player owes no skip after all, or the queue refused the chooser — and
// the caller then takes the recommended answer without asking.
//
// Never reached inside a resolution: the seam runs at action boundaries
// only (rotation.go's file comment).
//
// Caller must hold g.mu (write).
func (g *Game) pauseForTurnSkipOrderLocked(et ExtraTurn, others []string) bool {
	seq, source, ok := g.oldestTurnSkipLocked(et.Seat)
	if !ok {
		return false
	}
	chooser := g.Seats[et.Seat].ID
	name := ""
	if source != uuid.Nil {
		if c := g.findCardByIDLocked(source); c != nil {
			name = c.Name
		}
	}
	order := &TurnSkipOrder{
		ExtraTurnRef: et.Ref,
		SkipSource:   source,
		SkipSeq:      seq,
		SkipName:     name,
		Others:       append([]string(nil), others...),
	}
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:             PendingChoiceTurnSkipOrder,
		Chooser:          chooser,
		FromPlayer:       chooser,
		Count:            1,
		Source:           source,
		Reason:           "Your extra turn is skipped either way. Use it to pay off your own \"skip your next turn\"?",
		AcceptLabel:      useMySkipLabel(name),
		DeclineLabel:     letOthersSkipLabel(others),
		TurnSkipOrder:    order,
		AutoAnswerKey:    TurnSkipOrderAutoAnswerKey,
		AutoAnswerCard:   name,
		AutoAnswerPrompt: "Use your own skip on an extra turn another effect would skip?",
	})
	if id == uuid.Nil {
		return false
	}
	// A fresh slice, never an in-place append: an undo snapshot may share
	// the backing array (TakeExtraTurnsForEffect's rule).
	queue := make([]ExtraTurn, len(g.ExtraTurns), len(g.ExtraTurns)+1)
	copy(queue, g.ExtraTurns)
	g.ExtraTurns = append(queue, et)
	g.TurnChangePending = true
	g.Turn.PriorityHolder = NoPriority
	return true
}

// useMySkipLabel is answer 0's words.
func useMySkipLabel(name string) string {
	if name == "" {
		return "Use my skip"
	}
	return "Use my skip (" + name + ")"
}

// letOthersSkipLabel is answer 1's words.
func letOthersSkipLabel(others []string) string {
	if len(others) == 0 {
		return "Let the other effect skip it"
	}
	label := others[0]
	for _, o := range others[1:] {
		label += " / " + o
	}
	return "Let " + label + " skip it"
}

// turnSkipOrderOpenLocked reports whether a which-skip question is still
// in the queue. Caller must hold g.mu.
func (g *Game) turnSkipOrderOpenLocked() bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceTurnSkipOrder {
			return true
		}
	}
	return false
}

// ResolveTurnSkipOrder answers the which-skip question (ADR 0188
// Decision 12). useSkip true is answer 0: the player's own skip takes
// the turn and the charge is spent. False is answer 1: the extra-turn
// skip takes it (EventExtraTurnSkipped) and the charge is still owed —
// unless that effect no longer applies (the Nexus has left), in which
// case the charge is spent instead, because the turn must still be
// skipped once. Either way the turn change then resumes.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveTurnSkipOrder(choiceID, chooserID uuid.UUID, useSkip bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, c := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if c.Kind != PendingChoiceTurnSkipOrder {
		return ErrInvalidParam
	}
	if c.Chooser != chooserID {
		return ErrNotTheChooser
	}
	g.answerTurnSkipOrderLocked(idx, c, useSkip, false)
	return nil
}

// answerTurnSkipOrderLocked is the answer's body, shared with the ADR
// 0127 automatic answer. auto keeps the answer out of the CR 732 loop
// run (answerChoiceLocked).
//
// Caller must hold g.mu (write).
func (g *Game) answerTurnSkipOrderLocked(idx int, c *PendingChoice, useSkip, auto bool) {
	order := c.TurnSkipOrder
	g.answerChoiceLocked(idx, auto)
	if order != nil {
		g.settleTurnSkipOrderLocked(*order, useSkip)
	}
	if g.resumeTurnChangeLocked() {
		return
	}
	// CR 117.5 / 704.3: the new turn's first priority grant, as after
	// PassTurn.
	g.runStateChecksLocked()
}

// settleTurnSkipOrderLocked takes the questioned extra turn off the top
// of the queue and skips it the way the answer says. A turn that is no
// longer on top (nothing can move it while the question is open, but a
// restore is data) is left for the selection loop to decide afresh.
//
// Caller must hold g.mu (write).
func (g *Game) settleTurnSkipOrderLocked(o TurnSkipOrder, useSkip bool) {
	n := len(g.ExtraTurns)
	if n == 0 || g.ExtraTurns[n-1].Ref != o.ExtraTurnRef {
		return
	}
	et := g.ExtraTurns[n-1]
	// Re-sliced, never written through: popExtraTurnLocked's rule.
	g.ExtraTurns = g.ExtraTurns[: n-1 : n-1]
	if len(g.ExtraTurns) == 0 {
		g.ExtraTurns = nil
	}
	if et.Seat < 0 || et.Seat >= len(g.Seats) || g.Seats[et.Seat] == nil {
		return
	}
	if g.Seats[et.Seat].Eliminated {
		// CR 800.4k: the chooser has left, and their turn doesn't begin.
		g.noteTurnBegunLocked(et.Seat)
		g.dropTurnSkipsOfSeatLocked(et.Seat)
		return
	}
	if !useSkip && g.extraTurnSkippedLocked(et) {
		return
	}
	if source, ok := g.consumeTurnSkipSeqLocked(et.Seat, o.SkipSeq); ok {
		g.emitTurnSkippedLocked(et.Seat, source, et.Ref)
		return
	}
	// Neither skip applies any more: the turn is taken. Back on top, for
	// the selection loop.
	queue := make([]ExtraTurn, len(g.ExtraTurns), len(g.ExtraTurns)+1)
	copy(queue, g.ExtraTurns)
	g.ExtraTurns = append(queue, et)
}

// resumeTurnChangeLocked finishes a paused turn change: the flag is
// cleared and the next turn chosen, which may pause again for the next
// candidate. When it does not, the new turn's step entry hooks run.
// Reports whether it paused again.
//
// Caller must hold g.mu (write).
func (g *Game) resumeTurnChangeLocked() (paused bool) {
	g.TurnChangePending = false
	if g.chooseNextTurnLocked() {
		return true
	}
	g.runStepEntryHooksLocked()
	return false
}
