package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daybreak Chimera — Creature — Chimera {3}{W}{W}, 3/3 (Theros Beyond Death, #5514):
//
//	"This spell costs {X} less to cast, where X is your devotion to white.
//	 Flying"
//
// A cost reduction on the spell itself (CR 601.2f), X your devotion to white as it is
// cast; the Chimera is on the stack, so its own {W}{W} is not counted. It reduces
// the generic part only (CR 118.7) and never below {W}{W}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a2ea435c-5a61-4c98-bb76-8ff4ef0a0da9",
		Name:            "Daybreak Chimera",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int {
				return effects.DevotionTo(q.Game, q.Controller, q.Card.TextColor("W"))
			}, "This spell costs {X} less to cast, where X is your devotion to white."),
		},
	})
}
