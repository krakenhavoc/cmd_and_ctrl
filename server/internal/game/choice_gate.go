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
// choice's answers and nothing else, and while ANY choice is open no
// seat is offered anything (legal.anyChoiceOpen). The client mirrors
// it in its own way. Neither is enforcement — a hand-built frame, a
// stale tab, or an admin clicking "next step" walks the cursor past
// the prompt, and the continuation then resumes against a game that
// has moved on. #701 and #725 each had to teach one resume path to
// recognise its own stale frame and drop it; #651 reports the same
// shape for effect discards. Those tolerances are still worth having
// — a prompt can go stale without anyone advancing the step — but
// they are the second line, not the first.
//
// Deny by default. A pending choice blocks unless its kind is on
// nonBlockingChoiceKinds below. That direction is deliberate: a kind
// added later (choose_color arrived in #742 while this was being
// written) blocks without its author having to find this file, and
// the failure mode of the wrong default is a table that waits, not a
// table that silently loses a decision.

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
// battle protector, entry pay-life — blocks, including any kind added
// after this comment was written.
//
// One entry today:
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
// mana_pick is deliberately NOT here, though it looks like a
// candidate. The engine does not treat it as a background decision:
// legal offers nothing else while it is open, and the auto-tapper
// skips multi-option slots precisely because its contract is "no
// further player decisions required" (materializePlanLocked). An
// unanswered mana_pick is mana that has not entered the pool yet, so
// advancing past it would quietly discard the ability the player just
// activated.
var nonBlockingChoiceKinds = map[PendingChoiceKind]struct{}{
	PendingChoicePayUnless: {},
}

// blockingChoiceLocked returns the first outstanding choice whose
// kind is not on the allowlist, or nil when the table is free to
// move. Queue order, so the error names the prompt that has been
// waiting longest. Caller must hold g.mu.
func (g *Game) blockingChoiceLocked() *PendingChoice {
	for _, c := range g.PendingChoices {
		if c == nil {
			continue
		}
		if _, ok := nonBlockingChoiceKinds[c.Kind]; ok {
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
