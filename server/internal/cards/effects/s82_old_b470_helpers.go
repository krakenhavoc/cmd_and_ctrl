package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b470_helpers.go — Scars of Mirrodin, S82 older sets batch 470
// (#5577): the Trigon cycle's shared shape and the charge-counter artifacts'
// small bodies. Append-only.

// OldB470Trigon fills in a Trigon's two activated abilities and its entry
// counters on the given Spec (OracleID, Name, Completeness): "This artifact
// enters with three charge counters on it. <colour cost>, {T}: Put a charge
// counter on this artifact. {2}, {T}, Remove a charge counter from this
// artifact: <effect>." The counter is removed as the second ability is
// activated (CR 602.2b), so a response cannot spend it twice; the first
// ability does nothing if the Trigon has left (CR 400.7).
func OldB470Trigon(spec Spec, refill, label string, targets *game.TargetSpec, purpose game.Purpose, effect Effect) Spec {
	spec.Replacements = []game.ReplacementEffect{
		B10EntersWithCounters(game.CounterCharge, 3, spec.Name+": enters with three charge counters"),
	}
	spec.Activated = []ActivatedAbility{
		{
			Label:   refill + ", {T}: Put a charge counter on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost(refill), TapCost()),
			Effect:  B33PutChargeCounterOnSelf,
		},
		{
			Label:   "{2}, {T}, Remove a charge counter from this artifact: " + label,
			Purpose: purpose,
			Cost:    Plus(ManaCost("{2}"), TapCost(), RemoveCountersFromThis(game.CounterCharge, 1)),
			Targets: targets,
			Effect:  effect,
		},
	}
	return spec
}

// OldB470TargetPlayerLosesLife is "target player loses N life" as a
// triggered or activated body: the first legal player target, gone if it has
// left the game (CR 608.2b).
func OldB470TargetPlayerLosesLife(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		p, ok := FirstLegalPlayerTarget(ctx)
		if !ok {
			return nil
		}
		return GainLife{Player: p, Amount: -n}.Apply(ctx)
	}
}
