package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corridor Monitor — Artifact Creature — Construct {1}{U}, 1/4 (Throne of Eldraine, #5516):
//
//	"When this creature enters, untap target artifact or creature you control."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and
// re-read at resolution; the untap is Aphetto Alchemist's untapTheTarget.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07b6ee55-a77e-4ba7-9ab7-8a8375c7acc9",
		Name:         "Corridor Monitor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Corridor Monitor — untap target artifact or creature you control", effects.UntapTheTarget),
				effects.TargetPermanent("target artifact or creature you control", effects.Or(effects.Artifact(), effects.Creature()), effects.YouControl())),
		},
	})
}
