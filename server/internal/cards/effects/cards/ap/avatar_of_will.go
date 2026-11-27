package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avatar of Will — Creature — Avatar {6}{U}{U}, 5/6 (Prophecy, S82 OLD batch
// 673, #5617):
//
//	"If an opponent has no cards in hand, this spell costs {6} less to cast.
//	 Flying"
//
// A SelfCostModifiers entry (CR 601.2f): {6} less when any opponent of the
// caster still in the game has an empty hand as the spell is priced. The
// generic part is reduced; the {U}{U} stays. Flying is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "13aa88a0-4b47-4012-ba50-002e51e637cf",
		Name:            "Avatar of Will",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(6, "If an opponent has no cards in hand, this spell costs {6} less to cast.",
				func(q game.CostQuery) bool {
					if q.Game == nil {
						return false
					}
					for _, p := range q.Game.Seats {
						if p != nil && p.ID != q.Controller && effects.OldB529HandIsEmpty(q.Game, p.ID) {
							return true
						}
					}
					return false
				}),
		},
	})
}
