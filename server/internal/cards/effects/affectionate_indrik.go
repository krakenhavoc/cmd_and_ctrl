package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "88aa032d-dc11-4cae-81f3-ce66353963e0",
		Name:         "Affectionate Indrik",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WhenThisEnters("Affectionate Indrik — it fights target creature you don't control",
					b30SourceFightsFirstLegalTarget),
				TargetCreature("target creature you don't control", OpponentControls())),
				"Affectionate Indrik — have it fight target creature you don't control?"),
		},
	})
}
