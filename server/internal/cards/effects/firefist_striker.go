package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Firefist Striker — Creature — Human Soldier {1}{R}, 2/1 (Ravnica: Clue
// Edition):
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, target creature can't block this turn."
//
// Battalion is Boros Strike-Captain's group test (battalionAttack); the
// restriction is Ahn-Crop Crasher's, ending in the cleanup step.
//
// No simplification.
func init() {
	const label = "Firefist Striker — target creature can't block this turn (battalion)"
	Register(Spec{
		OracleID:     "a017ab43-7d73-441a-8931-2a038306083f",
		Name:         "Firefist Striker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventAttack, battalionAttack, label, targetCreatureCantBlockThisTurn(label)),
				TargetCreature("target creature")),
		},
	})
}
