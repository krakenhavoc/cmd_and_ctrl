package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyrokinesis — Instant {4}{R}{R} (#1658, unblocked by #1656):
//
//	"You may exile a red card from your hand rather than pay this
//	 spell's mana cost.
//	 Pyrokinesis deals 4 damage divided as you choose among any
//	 number of target creatures."
//
// The alternative cost is the Force of Will/Fury shape with no life
// component (Pitch's `life` parameter left at 0 — see Force of
// Negation for the same zero-life pattern), so it needs no new
// machinery: exile a red card from hand rather than pay the mana cost,
// paid with the spell already on the stack (CR 601.2a before 601.2h).
// The division is Fury's fixed-4 "any number" shape, restricted to
// creatures.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d001febc-d511-4e65-a631-91dc9415056b",
		Name:         "Pyrokinesis",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Pitch(
				"Exile a red card from your hand rather than pay this spell's mana cost.",
				0,
				effects.CardInYourHand("a red card from your hand", effects.OfColor("R")),
				"a red card from your hand",
			),
		},
		Targets: effects.TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(effects.Divide(4)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
