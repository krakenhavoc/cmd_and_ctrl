package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bone Picker — Creature — Bird {3}{B}, 3/2
// (Amonkhet, #5535):
//
//	"This spell costs {3} less to cast if a creature died this turn.
//	 Flying, deathtouch"
//
// Dreaded Bat-Cloud's cost reduction and keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2edf106d-f504-430e-ad9c-50fc95b0ad9c",
		Name:            "Bone Picker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(3, "This spell costs {3} less to cast if a creature died this turn.",
				func(q game.CostQuery) bool { return q.Game != nil && effects.B11CreaturesDiedThisTurn(q.Game) > 0 }),
		},
	})
}
