package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marshmist Titan — {6}{B} Creature — Giant 4/5 (Born of the Gods):
//
//	"This spell costs {X} less to cast, where X is your devotion to black. (Each {B} in the
//	 mana costs of permanents you control counts toward your devotion to black.)"
//
// A self cost reducer (CR 601.2f, generic mana only), the devotion read as the spell is
// priced (CR 700.5).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8b1ac76-2f9b-454e-a20a-1cb79c23f414",
		Name:         "Marshmist Titan",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int { return effects.DevotionToColors(q.Game, q.Controller, "B") },
				"This spell costs {X} less to cast, where X is your devotion to black."),
		},
	})
}
