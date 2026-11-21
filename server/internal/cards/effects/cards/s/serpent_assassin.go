package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serpent Assassin — Creature — Snake Assassin {3}{B}{B}, 2/2 (Foundations
// Jumpstart, S82 BNS):
//
//	"When this creature enters, you may destroy target nonblack creature."
//
// Dark Hatchling's target and destroy (regeneration allowed, as the card
// does not forbid it), with "you may" asked as it resolves. The target is
// judged as the trigger goes on the stack and again as it resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bc7d3a7-ddb4-4eaa-882e-404d8f2926fb",
		Name:         "Serpent Assassin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Serpent Assassin — you may destroy target nonblack creature", func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
				}),
				effects.TargetCreature("target nonblack creature", effects.NonBlack())),
				"Serpent Assassin — destroy target nonblack creature?"),
		},
	})
}
