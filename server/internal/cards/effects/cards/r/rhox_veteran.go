package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a681c6ea-b10a-4061-abf4-7b45e06ccebd",
		Name:         "Rhox Veteran",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.BattleCry("Rhox Veteran"),
			effects.Targeting(
				effects.WheneverThisAttacks("Rhox Veteran — tap target creature an opponent controls",
					effects.TapTheFirstLegalTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
