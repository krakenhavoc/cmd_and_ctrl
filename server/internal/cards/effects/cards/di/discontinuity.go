package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Discontinuity — Instant {3}{U}{U}{U} (Core Set 2021, #5511):
//
//	"During your turn, this spell costs {2}{U}{U} less to cast.
//	 End the turn. (Exile all spells and abilities from the stack, including
//	 this card. The player whose turn it is discards down to their maximum
//	 hand size. Damage wears off, and "this turn" and "until end of turn"
//	 effects end.)"
//
// Time Stop's EndTheTurn (CR 724.1) with a self cost reduction whose unit is
// the whole {2}{U}{U} (CR 601.2f): it applies only while it is the caster's
// own turn, so a cast on an opponent's turn costs the full {3}{U}{U}{U}.
//
// No simplification.
func init() {
	unit, err := game.ParseCost("{2}{U}{U}")
	if err != nil {
		panic("effects.Discontinuity: " + err.Error())
	}
	reduction := effects.CostsLess(1, "During your turn, this spell costs {2}{U}{U} less to cast.",
		func(q game.CostQuery) bool {
			return q.Game != nil && effects.ActivePlayerIDOf(q.Game) == q.Controller
		})
	reduction.Unit = &unit
	effects.Register(effects.Spec{
		OracleID:          "5cfd73ab-8509-4e4b-9704-68d8dad3b301",
		Name:              "Discontinuity",
		Completeness:      effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{reduction},
		OnResolve:         effects.EndTheTurnOnResolve,
	})
}
