package aiseat

import (
	"context"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// attack_targets.go — a bot whose attackers another player chose
// (#4154, Master Warcraft: "you choose which creatures attack this
// turn").
//
// The declarer chooses the attacking creatures; the active player then
// chooses what each attacks and whether to pay its costs (the card's
// ruling, and CR 508.1b). A bot in that half is offered the re-points,
// the confirmation that pays the costs, and the pass that accepts the
// targets as they stand and declines any cost. attackTargetsPolicy
// never asks a model: it accepts when it may, pays when the engine
// needs an answer and the bot can, and otherwise takes a creature's
// free target. That is legal and ends the half in one or two moves: it
// keeps the bot's mana, and it never sends a creature it did not choose
// to attack into a cost.

// attackTargetsPolicy decides the active player's half of a chosen
// attack declaration.
type attackTargetsPolicy struct {
	inner Policy
}

func (p attackTargetsPolicy) Name() string {
	if p.inner == nil {
		return "attack-targets"
	}
	return p.inner.Name() + "/attack-targets"
}

// Unwrap names the seat's own policy (capability.go).
func (p attackTargetsPolicy) Unwrap() Policy { return p.inner }

func (p attackTargetsPolicy) Decide(_ context.Context, in Input) (Decision, error) {
	if len(in.Moves) == 0 {
		return Decision{}, ErrNoMoves
	}
	if i := PassIndex(in.Moves); i >= 0 {
		return Decision{Index: i, Reason: "chosen attackers: accept the targets"}, nil
	}
	for i, m := range in.Moves {
		if m.Type == legal.TypeDeclareAttackers {
			return Decision{Index: i, Reason: "chosen attackers: confirm and pay"}, nil
		}
	}
	for i, m := range in.Moves {
		if m.Type == legal.TypeDeclareAttacker && (m.Cost == nil || m.Cost.Mana == "") {
			return Decision{Index: i, Reason: "chosen attackers: a free target"}, nil
		}
	}
	for i, m := range in.Moves {
		if m.Kind == legal.KindAttack {
			return Decision{Index: i, Reason: "chosen attackers: a target"}, nil
		}
	}
	return Decision{Index: 0, Reason: "chosen attackers: the first answer"}, nil
}
