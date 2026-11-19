package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goma Fada Vanguard — Creature — Human Warrior {1}{R}, 2/2 (Foundations
// Jumpstart):
//
//	"Whenever this creature attacks, target creature an opponent controls
//	 with power less than or equal to the number of Warriors you control
//	 can't block this turn."
//
// Clamor Shaman's attack trigger (Ahn-Crop Crasher's targetCreatureCantBlockThisTurn) with a target clause counted in Warriors
// (PowerAtMostWarriorsYouControl, which counts the Vanguard itself). No
// simplification.
func init() {
	const label = "Goma Fada Vanguard — target creature an opponent controls with power less than or equal to the number of Warriors you control can't block this turn"
	Register(Spec{
		OracleID:     "74027d31-6efb-471d-9499-cdb1cefc9eae",
		Name:         "Goma Fada Vanguard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks(label, targetCreatureCantBlockThisTurn(label)), TargetCreature("target creature an opponent controls with power less than or equal to the number of Warriors you control",
				OpponentControls(), PowerAtMostWarriorsYouControl())),
		},
	})
}
