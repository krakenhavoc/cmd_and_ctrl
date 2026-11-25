package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Spike — Instant {5}{B}{B} (Secret Lair Drop, #5462):
//
//	"You may exile two black cards from your hand rather than pay this
//	 spell's mana cost.
//	 Soul Spike deals 4 damage to any target and you gain 4 life."
//
// Fury of the Horde's two-card pitch over Warleader's Helix's body. The pitch
// is an alternative cost (CR 118.9) with no life: both black cards leave the
// hand as the cost is paid, and the spell itself, already on the stack, can't
// be one. The damage and the life are two steps of one resolution, the life
// after any prompt the damage asks (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be979bfc-998d-43d5-a7a4-44548d9a14c4",
		Name:         "Soul Spike",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Pitch("Exile two black cards from your hand", 0,
				effects.CardInYourHand("two black cards from your hand", effects.OfColor("B")).WithCount(2, 2),
				"two black cards"),
		},
		Targets: effects.TargetAny(),
		Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.S82B48DamageToTheFirstTarget(4)),
			effects.GainLife{Amount: 4},
		),
	})
}
