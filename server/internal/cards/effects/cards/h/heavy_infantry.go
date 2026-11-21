package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heavy Infantry — Creature — Human Soldier {4}{W}, 3/4 (Magic Origins, #5549):
//
//	"When this creature enters, tap target creature an opponent controls."
//
// Vexing Drone's entry: the opponent's creature is targeted as the
// trigger goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d07707ef-fa4e-423f-8a1f-095d19100321",
		Name:         "Heavy Infantry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Heavy Infantry — tap target creature an opponent controls", effects.TapTheFirstLegalTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
