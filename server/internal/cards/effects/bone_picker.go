package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "2edf106d-f504-430e-ad9c-50fc95b0ad9c",
		Name:            "Bone Picker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(3, "This spell costs {3} less to cast if a creature died this turn.",
				func(q game.CostQuery) bool { return q.Game != nil && b11CreaturesDiedThisTurn(q.Game) > 0 }),
		},
	})
}
