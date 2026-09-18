package game

import (
	"errors"

	"github.com/google/uuid"
)

// choice_gate.go holds one rule (#730): while a prompt nobody has
// answered is sitting in Game.PendingChoices, the table does not move
// on. The turn-structure verbs — advance_step, pass_priority,
// pass_turn — are refused, and the prompt's chooser answers it first.
//
// Why the engine and not just the client. `internal/legal` already
// models this for bots: a seat that owes a choice is offered that
// choice's answers and nothing else, and while a BLOCKING choice is
// open no seat is offered anything (legal.anyBlockingChoiceOpen, which
// asks ChoiceBlocksTable below — #794). The client mirrors it in its
// own way. Neither is enforcement — a hand-built frame, a
// stale tab, or an admin clicking "next step" walks the cursor past
// the prompt, and the continuation then resumes against a game that
// has moved on. #701 and #725 each had to teach one resume path to
// recognise its own stale frame and drop it; #651 reports the same
// shape for effect discards. Those tolerances are still worth having
// — a prompt can go stale without anyone advancing the step — but
// they are the second line, not the first.
//
// Deny by default. A pending choice blocks unless choiceGateDecisions
// below classifies its kind as non-blocking. That direction is
// deliberate: a kind added later (choose_color arrived in #742 while
// this was being written) blocks without its author having to find
// this file, and the failure mode of the wrong default is a table that
// waits, not a table that silently loses a decision.

// ErrChoicePending is the sentinel behind every refusal from this
// gate. Callers discriminate with errors.Is; the concrete
// *ChoicePendingError carries which prompt is owed.
var ErrChoicePending = errors.New("game: an unanswered prompt is blocking the table")

// ChoicePendingError is what the gated verbs actually return: the
// sentinel plus enough of the outstanding choice for the wire message
// to say what is owed and for a client to match it against the
// prompt it is already showing.
type ChoicePendingError struct {
	// ChoiceID is the PendingChoice.ID the table is waiting on.
	ChoiceID uuid.UUID
	// Kind is that choice's kind ("sacrifice_choice", "scry", ...).
	Kind PendingChoiceKind
	// Chooser is the seat that owes the answer.
	Chooser uuid.UUID
	// Reason is the prompt's own header text ("Braids, Cabal
	// Minion"), empty for choices queued without one.
	Reason string
}

func (e *ChoicePendingError) Error() string {
	label := string(e.Kind)
	if e.Reason != "" {
		label = e.Reason + " (" + string(e.Kind) + ")"
	}
	// The "game: " prefix is stripped by ws.classifyActionError, so
	// what the player reads is the sentence after it.
	return "game: waiting on an unanswered prompt — " + label + " [" + e.ChoiceID.String() + "]"
}

// Unwrap lets errors.Is(err, ErrChoicePending) match.
func (e *ChoicePendingError) Unwrap() error { return ErrChoicePending }

// nonBlockingChoiceKinds is the allowlist: the ONLY kinds a table may
// walk past. Everything else — sacrifice, discard, pick-target,
// trigger prompts and ordering, search / choose-cards, scry, surveil,
// look-at-top, replacement order and optional replacements, creature
// type, colour choice, copy target, may-cast, mana pick, legend rule,
// battle protector, entry pay-life, the CR 726 loop shortcut — blocks,
// including any kind added after this comment was written.
//
// One kind is classified `false` today:
//
//	pay_unless — ADR 0018 §6 accepted the looseness on purpose.
//	Rhystic Study asks a DIFFERENT player a question after the
//	trigger has already resolved and left the stack; the table plays
//	it the way paper does ("you paying for that?" while the next
//	spell is already being cast), and a modal lockstep across four
//	browsers is worse. Nothing downstream of the prompt depends on
//	the cursor: the answer spends from the chooser's pool or runs
//	OnDecline, neither of which reads the step.
//
// mana_pick is deliberately NOT one of them, though it looks like a
// candidate. The engine does not treat it as a background decision:
// legal offers nothing else while it is open, and the auto-tapper
// skips multi-option slots precisely because its contract is "no
// further player decisions required" (materializePlanLocked). An
// unanswered mana_pick is mana that has not entered the pool yet, so
// advancing past it would quietly discard the ability the player just
// activated.
//
// #794: this table used to be a bare allowlist read only from inside
// this file, and `internal/legal` kept its own, stricter copy — a seat
// was offered nothing at all while ANY prompt was open, `pay_unless`
// included. The two disagreed for the whole life of the allowlist, so
// a bot idled until a human answered a Rhystic tax, which is the
// opposite of what §6 decided. The map is now a full classification —
// every kind, an explicit true or false — read through
// ChoiceBlocksTable by the engine AND by the enumerator, so there is
// one answer to "does this prompt stop the table" and no second list
// to drift from.
var choiceGateDecisions = map[PendingChoiceKind]bool{
	// The ADR 0018 §6 allowlist, entire.
	PendingChoicePayUnless: false,

	// Everything else stops the table.
	PendingChoiceDiscardFromHand:     true,
	PendingChoiceMana:                true,
	PendingChoiceReplacementOrder:    true,
	PendingChoiceOptionalReplacement: true,
	PendingChoiceDamageAssignment:    true,
	PendingChoiceTriggerPrompt:       true,
	PendingChoiceTriggerOrder:        true,
	// #764 CR 603.3c: a trigger's mode is chosen as it is put on the
	// stack. Nothing may happen until it is — the ability is not on
	// the stack yet, and the targets it asks for next depend on the
	// answer.
	PendingChoiceModePick: true,

	PendingChoicePickTarget:      true,
	PendingChoiceSacrifice:       true,
	PendingChoiceScry:            true,
	PendingChoiceSurveil:         true,
	PendingChoiceLookAtTop:       true,
	PendingChoiceSearchLibrary:   true,
	PendingChoiceMayCast:         true,
	PendingChoiceCoinCall:        true,
	PendingChoiceChooseProtector: true,
	PendingChoiceLegendRule:      true,
	PendingChoiceColor:           true,
	PendingChoiceConfirm:         true,
	PendingChoiceChooseCards:     true,
	PendingChoiceEntryPayLife:    true,
	PendingChoiceCopyTarget:      true,
	PendingChoiceCreatureType:    true,
	// #568. "Choose one of the following", addressed to any seat —
	// Torment of Hailfire's three-way question, and the second half
	// of a Fact or Fiction pile split. It blocks for the reason every
	// resolution-time prompt does: the effect that asked it is paused
	// mid-resolution and its continuation is the rest of the card.
	PendingChoiceOptionPick: true,
	// #804, CR 726. The one kind whose blocking is worth arguing
	// about, since ADR 0055 §4 was careful that the loop breaker
	// refuse no passes. It blocks: the shortcut is proposed while the
	// loop's trigger is still on the stack, and what the answer
	// decides is how many times that trigger resolves next. A table
	// that could pass through the question would be answering it by
	// doing. The prompt is never queued to a seat that has left, so
	// blocking cannot wedge a table (queueLoopShortcutLocked).
	PendingChoiceLoopShortcut: true,
}

// ChoiceBlocksTable is THE question "does an unanswered prompt of this
// kind stop the table?", and the only place it is answered. The engine
// gates advance_step / pass_priority / pass_turn on it
// (blockingChoiceLocked, below) and `internal/legal` decides whether to
// enumerate a seat's ordinary moves with it
// (legal.anyBlockingChoiceOpen). A second copy of this judgement is
// what #794 was.
//
// Deny by default: a kind missing from choiceGateDecisions blocks. The
// failure mode of the wrong default is a table that waits, not a table
// that silently loses a decision — and
// TestEveryChoiceKindIsClassifiedAndEnumerated (internal/legal) fails
// until the new kind is given a row here anyway.
func ChoiceBlocksTable(kind PendingChoiceKind) bool {
	blocks, ok := choiceGateDecisions[kind]
	return !ok || blocks
}

// ClassifiedChoiceKinds lists every kind the gate has an explicit
// decision for, so a test can hold that list against the kinds
// declared in this package and fail on one that was never classified.
// Order is unspecified.
func ClassifiedChoiceKinds() []PendingChoiceKind {
	out := make([]PendingChoiceKind, 0, len(choiceGateDecisions))
	for kind := range choiceGateDecisions {
		out = append(out, kind)
	}
	return out
}

// blockingChoiceLocked returns the first outstanding choice whose kind
// blocks the table, or nil when the table is free to move. Queue
// order, so the error names the prompt that has been waiting longest.
// Caller must hold g.mu.
func (g *Game) blockingChoiceLocked() *PendingChoice {
	for _, c := range g.PendingChoices {
		if c == nil {
			continue
		}
		if !ChoiceBlocksTable(c.Kind) {
			continue
		}
		return c
	}
	return nil
}

// choicePendingErrorLocked builds the refusal for an outstanding
// choice. Caller must hold g.mu (it reads the choice in place).
func choicePendingErrorLocked(c *PendingChoice) error {
	return &ChoicePendingError{
		ChoiceID: c.ID,
		Kind:     c.Kind,
		Chooser:  c.Chooser,
		Reason:   c.Reason,
	}
}
