package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mouse Trapper — Creature — Mouse Soldier {2}{W}, 3/2 (Bloomburrow, #5230):
//
//	"Flash
//	 Valiant — Whenever this creature becomes the target of a spell or
//	 ability you control for the first time each turn, tap target creature
//	 an opponent controls."
//
// The shared valiant trigger (valiant.go) with a target clause, chosen as
// the trigger goes on the stack (CR 603.3d): with no creature an opponent
// controls it is removed, and the Trapper has still been targeted for the
// turn. The tap is Harrier Strix's, on the target if it is still legal (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5b0a4530-6e19-41a2-9ea6-e4c8cf85b6c4",
		Name:            "Mouse Trapper",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Answers: game.AnswerMakesBlocker}, // a creature with flash is a blocker
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(Valiant("Mouse Trapper — Valiant: tap target creature an opponent controls", tapTheFirstLegalTarget),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
