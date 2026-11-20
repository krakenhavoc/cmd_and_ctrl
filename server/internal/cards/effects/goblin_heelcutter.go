package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Heelcutter — Creature — Goblin Berserker {3}{R}, 3/2 (Fate Reforged,
// #5552):
//
//	"Whenever this creature attacks, target creature can't block this
//	 turn.
//	 Dash {2}{R} (You may cast this spell for its dash cost. If you do,
//	 it gains haste, and it's returned from the battlefield to its
//	 owner's hand at the beginning of the next end step.)"
//
// The attack trigger is Ahn-Crop Crasher's restriction: any creature is a
// legal target, and it can't be declared as a blocker until the turn ends;
// the trigger resolves before blockers. With no creature to target the
// trigger is removed (CR 603.3d). Dash is the shared Dash constructor
// (freerunning.go): an alternative cost whose entry clause gives haste and
// returns the creature to its owner's hand at the beginning of the next end
// step (CR 702.109a, game/dash.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "ab52988f-ffd0-4154-9052-593a9d450d98",
		Name:             "Goblin Heelcutter",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Dash("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks(goblinHeelcutterLabel, targetCreatureCantBlockThisTurn(goblinHeelcutterLabel)),
				TargetCreature("target creature")),
		},
	})
}

const goblinHeelcutterLabel = "Goblin Heelcutter — target creature can't block this turn"
