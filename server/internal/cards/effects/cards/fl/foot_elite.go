package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foot Elite — {2}{W/B} Creature — Human Ninja 2/4 (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Whenever this creature attacks, another target creature you control
//	 gets +1/+0 and gains indestructible until end of turn."
//
// Gruul Beastmaster's "another target creature you control" attack
// trigger, with the shared pumpAndGrantFirstTarget body: the still-legal
// target (CR 608.2b) gets +1/+0 and indestructible, nothing if it left.
//
// No simplification.
func init() {
	const label = "Foot Elite — +1/+0 and indestructible until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "b6244b6c-664d-417a-bd58-ace5c7a995a2",
		Name:         "Foot Elite",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks(label, effects.PumpAndGrantFirstTarget(1, 0, label, "indestructible")),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
