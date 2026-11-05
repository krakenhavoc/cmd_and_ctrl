package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Utrom Scientists — {2}{U} Artifact Creature — Utrom Robot Scientist 2/2
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"When this creature enters, tap up to one target creature and put a
//	 stun counter on it."
//
// Fractal Mascot's trigger with an "up to one" target on any creature: the
// shared tap-and-stun body (tapTargetAndStunIt). With no target chosen, or
// one that is gone by resolution, nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89793c8c-98a3-4621-ad3d-cfc5949c65da",
		Name:         "Utrom Scientists",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Utrom Scientists — tap up to one target creature and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}
