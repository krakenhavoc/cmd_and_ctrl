package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// D'Avenant Trapper — Creature — Human Archer {2}{W}, 3/2 (Dominaria, S82
// OLD, #5530):
//
//	"Whenever you cast a historic spell, tap target creature an opponent
//	 controls. (Artifacts, legendaries, and Sagas are historic.)"
//
// Jhoira's "you cast a historic spell" (CR 700.6) with a target chosen as the
// trigger goes on the stack; the creature is tapped as it resolves if it is
// still a legal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "745d37dd-d18e-47cc-81d0-8e61a07342f2",
		Name:         "D'Avenant Trapper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCast, effects.YouCastAHistoricSpell, "D'Avenant Trapper — tap target creature an opponent controls", effects.TapEachLegalTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
