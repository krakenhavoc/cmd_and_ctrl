package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grand Arbiter Augustin IV — Legendary Creature — Human Advisor {2}{W}{U}, 2/3 (S82 GAP, #5948):
//
//	"White spells you cast cost {1} less to cast.
//	 Blue spells you cast cost {1} less to cast.
//	 Spells your opponents cast cost {1} more to cast."
//
// Three cost modifiers (CR 601.2f). A white-and-blue spell gets both discounts.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f8d4d5f-e82f-45f3-823e-1bb6b536eb18",
		Name:         "Grand Arbiter Augustin IV",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "White spells you cast cost {1} less to cast.", effects.YourSpell(), effects.ColoredSpell("W")),
			effects.CostsLess(1, "Blue spells you cast cost {1} less to cast.", effects.YourSpell(), effects.ColoredSpell("U")),
			effects.CostsMore(1, "Spells your opponents cast cost {1} more to cast.", effects.OpponentsSpell()),
		},
	})
}
