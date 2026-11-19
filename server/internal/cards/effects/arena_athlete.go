package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arena Athlete — Creature — Human {1}{R}, 2/1 (Foundations Jumpstart,
// #5929):
//
//	"Heroic — Whenever you cast a spell that targets this creature, target
//	 creature an opponent controls can't block this turn."
//
// The shared heroic trigger (CR 207.2c) with a target of its own, chosen as
// the trigger goes on the stack; with no creature an opponent controls the
// trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4dfea3bf-dfd1-4abd-a17e-f3a07488a069",
		Name:         "Arena Athlete",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(heroic("Arena Athlete — target creature an opponent controls can't block this turn (heroic)",
				targetCreatureCantBlockThisTurn("Arena Athlete — target creature can't block this turn")), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
