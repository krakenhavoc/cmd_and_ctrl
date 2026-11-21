package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venerable Knight — Creature — Human Knight {W}, 2/1 (Throne of Eldraine, #5516):
//
//	"When this creature dies, put a +1/+1 counter on target Knight you control."
//
// A targeted dies trigger; with no Knight left it is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef027846-be81-4959-a6b5-56bd01b1e68a",
		Name:         "Venerable Knight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Venerable Knight — put a +1/+1 counter on target Knight you control", effects.RfCreatureFCountersOnTarget(1)),
				effects.TargetCreature("target Knight you control", effects.HasSubtype("Knight"), effects.YouControl())),
		},
	})
}
