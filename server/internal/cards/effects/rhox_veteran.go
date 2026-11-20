package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rhox Veteran — Creature — Rhino Soldier {3}{W}, 2/4 (Modern Horizons, #5519):
//
//	"Battle cry (Whenever this creature attacks, each other attacking
//	 creature gets +1/+0 until end of turn.)
//	 Whenever this creature attacks, tap target creature an opponent
//	 controls."
//
// The shared battle cry trigger and Tidus's attack trigger that taps a
// creature an opponent controls; the two are separate triggers its
// controller orders.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a681c6ea-b10a-4061-abf4-7b45e06ccebd",
		Name:         "Rhox Veteran",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			BattleCry("Rhox Veteran"),
			Targeting(
				WheneverThisAttacks("Rhox Veteran — tap target creature an opponent controls",
					tapTheFirstLegalTarget),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
