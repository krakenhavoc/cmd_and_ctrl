package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Festering Mummy — Creature — Zombie {B}, 1/1
// (Amonkhet, #5535):
//
//	"When this creature dies, you may put a -1/-1 counter on target creature."
//
// Necropede's dies trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bdf78365-5887-497e-af25-ca0cb731f35a",
		Name:         "Festering Mummy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WhenThisDies("Festering Mummy — you may put a -1/-1 counter on target creature",
					putACounterOnTheTarget(game.CounterMinusOne)),
				Harmful(TargetCreature("target creature"))),
				"Festering Mummy — put a -1/-1 counter on the target creature?"),
		},
	})
}
