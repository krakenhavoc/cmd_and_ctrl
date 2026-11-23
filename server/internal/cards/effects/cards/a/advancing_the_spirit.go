package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Advancing the Spirit — {2}{G} Enchantment:
//
//	"When this enchantment enters, draw a card.
//	 You may pay {0} rather than pay the power-up cost of the first
//	 power-up ability you activate during each of your turns."
//
// The second clause is a CR 118.9 alternative cost on an activation
// (game.CostZeroAlternative): during your turn, until you have activated
// a power-up ability that turn, each power-up ability you activate is
// priced from {0}, and increases and reductions still apply to it
// (CR 118.9d). The engine takes the alternative for you, because paying
// the printed cost instead never leaves you better off: the first
// power-up of the turn uses the clause up whichever cost it paid. The
// count is written at the announce, so a countered power-up was still
// the first.
//
// No simplification. Marvel Super Heroes (#4455, seam #4527).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9949a5a-13d4-41ca-80fe-902b0ae71204",
		Name:         "Advancing the Spirit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Advancing the Spirit — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		CostModifiers: []game.CostModifier{
			effects.PayZeroRatherThanThePowerUpCost("You may pay {0} rather than pay the power-up cost of the first power-up ability you activate during each of your turns."),
		},
	})
}
