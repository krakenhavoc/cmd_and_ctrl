package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Affectionate Indrik — Creature — Beast {5}{G}, 4/4 (Foundations,
// #4465):
//
//	"When this creature enters, you may have it fight target creature
//	 you don't control."
//
// Frontier Siege's optional fight on the Indrik's own enters trigger.
// The target is chosen as the trigger goes on the stack (with none, the
// trigger is removed, CR 603.3d); "you may" is asked as it resolves. An
// Indrik no longer on the battlefield, or a target that has left, deals
// and is dealt no damage (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88aa032d-dc11-4cae-81f3-ce66353963e0",
		Name:         "Affectionate Indrik",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Affectionate Indrik — it fights target creature you don't control",
					effects.B30SourceFightsFirstLegalTarget),
				effects.TargetCreature("target creature you don't control", effects.OpponentControls())),
				"Affectionate Indrik — have it fight target creature you don't control?"),
		},
	})
}
