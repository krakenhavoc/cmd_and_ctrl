package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raging Regisaur — Creature — Dinosaur {2}{R}{G}, 4/4 (Lost Caverns of
// Ixalan Commander):
//
//	"Whenever this creature attacks, it deals 1 damage to any target."
//
// An attack trigger with a target chosen as it goes on the stack (CR
// 603.3d); the damage comes from the Regisaur, and a target that left in
// response leaves the ability to do nothing (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "3d563149-7d4e-43e5-b701-28e3212051ec",
		Name:         "Raging Regisaur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Raging Regisaur — 1 damage to any target", DealDamageToTheTarget(1)),
				TargetAny()),
		},
	})
}
