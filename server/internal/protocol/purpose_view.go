package protocol

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// purpose_view.go — ADR 0126 §6: game.Purpose on the wire.
//
// A purpose is catalog data, declared on the card file, never derived:
// the printed amounts a spell, a mode, an alternative cost or an
// ability declares it does. The bot reads it from here and nowhere else
// (ADR 0033 §3: a policy reads the view, not the catalog).
//
// Where it rides, and who sees it:
//
//   - CardView.purpose — the spell, or a permanent's enters effect. On
//     the hand, the battlefield and the stack, the zones ability_rows
//     and the cast surface use. Cleared with ability_rows for a viewer
//     who cannot see the card (redactCardForViewer), so an opponent's
//     hand card and a face-down object never carry one.
//   - ModeOptionView.purpose and AlternativeCostView.purpose — beside
//     the bullet's label and the offer's label, which say the same
//     thing in words, and hidden exactly when they are.
//   - ActivatedAbilityView.purpose and AbilityRowView.purpose — the
//     row's own, public with the row.

// PurposeView is game.Purpose on the wire. Every field is omitted when
// zero, and a purpose with nothing declared is not sent at all.
type PurposeView struct {
	// Draws is the cards its controller draws (on an any-player row,
	// the activator).
	Draws int `json:"draws,omitempty"`
	// ControllerLosesLife is the life the SOURCE'S controller loses.
	// Any-player activated rows only (ADR 0106 §1 decision 8).
	ControllerLosesLife int `json:"controller_loses_life,omitempty"`
	// Discards is the cards its controller discards on resolution.
	Discards int `json:"discards,omitempty"`
	// Lands is the land cards it puts onto the battlefield.
	Lands int `json:"lands,omitempty"`
	// Tutors is the cards it searches out to hand or to the top of the
	// library.
	Tutors int `json:"tutors,omitempty"`
	// SelfMillTutor is the cards it searches out into the graveyard.
	SelfMillTutor int `json:"self_mill_tutor,omitempty"`
	// Tokens is the tokens it creates for its controller.
	Tokens int `json:"tokens,omitempty"`
	// Sweep is present on a board wipe.
	Sweep *SweepView `json:"sweep,omitempty"`
	// DeathPayoff is set on a triggered row that pays out whenever a
	// creature its controller controls dies.
	DeathPayoff bool `json:"death_payoff,omitempty"`
	// DiscardPayoff is present on a triggered row that pays out
	// whenever its controller discards a card it matches (ADR 0126's
	// amendment of 2026-10-06).
	DiscardPayoff *DiscardPayoffView `json:"discard_payoff,omitempty"`
}

// DiscardPayoffView is game.DiscardPayoff on the wire: which discarded
// cards the row pays on, and what it pays for each one.
type DiscardPayoffView struct {
	// Any: every card its controller discards pays.
	Any bool `json:"any,omitempty"`
	// Types: otherwise, the lowercase card types and subtypes it pays
	// on; a card with any one of them on its type line matches.
	Types []string `json:"types,omitempty"`
	// Tokens is the tokens it creates for its controller per card.
	Tokens int `json:"tokens,omitempty"`
	// Counters is the +1/+1 counters it puts on its source per card.
	Counters int `json:"counters,omitempty"`
	// DamageEachOpponent is the damage its source deals to each
	// opponent per card.
	DamageEachOpponent int `json:"damage_each_opponent,omitempty"`
}

// ActivationPurposeView is ADR 0106's name for PurposeView, from when a
// purpose was declared on any-player rows alone. Kept so a reader of
// that ADR finds the type.
type ActivationPurposeView = PurposeView

// SweepView is game.Sweep on the wire.
type SweepView struct {
	// Matches is the class of permanents it removes (game.SweepMatches).
	Matches string `json:"matches"`
	// How is destroy, exile, bounce, damage, minus or sacrifice.
	How string `json:"how"`
	// Amount is the damage or the N of -N/-N; absent otherwise and
	// when AmountIsX.
	Amount int `json:"amount,omitempty"`
	// AmountIsX: the amount is the spell's or ability's X.
	AmountIsX bool `json:"amount_is_x,omitempty"`
	// OpponentsOnly: only permanents its controller's opponents
	// control are removed.
	OpponentsOnly bool `json:"opponents_only,omitempty"`
	// Partial: some permanents of the class are spared by a condition
	// the view does not name, so Matches is an upper bound.
	Partial bool `json:"partial,omitempty"`
}

// viewOfPurpose projects a declared purpose, nil when none is declared.
func viewOfPurpose(p game.Purpose) *PurposeView {
	if p.IsZero() {
		return nil
	}
	v := &PurposeView{
		Draws:               p.Draws,
		ControllerLosesLife: p.ControllerLosesLife,
		Discards:            p.Discards,
		Lands:               p.Lands,
		Tutors:              p.Tutors,
		SelfMillTutor:       p.SelfMillTutor,
		Tokens:              p.Tokens,
		DeathPayoff:         p.DeathPayoff,
	}
	if d := p.DiscardPayoff; d != nil {
		v.DiscardPayoff = &DiscardPayoffView{
			Any:                d.Any,
			Types:              append([]string(nil), d.Types...),
			Tokens:             d.Tokens,
			Counters:           d.Counters,
			DamageEachOpponent: d.DamageEachOpponent,
		}
	}
	if s := p.Sweep; !s.IsZero() {
		v.Sweep = &SweepView{
			Matches:       string(s.Matches),
			How:           string(s.How),
			Amount:        s.Amount,
			AmountIsX:     s.AmountIsX,
			OpponentsOnly: s.OpponentsOnly,
			Partial:       s.Partial,
		}
	}
	return v
}
