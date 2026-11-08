package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Mary Janes — Legendary Creature — Human Bard Performer {3}{R},
// 3/3 (Marvel's Spider-Man, #4460):
//
//	"This spell costs {1} less to cast for each creature that attacked
//	 this turn.
//	 Menace"
//
// The discount counts distinct creatures declared as attackers this turn,
// anyone's, through the one pricer (CR 601.2f). It spends generic mana
// only, so the spell never costs less than {R}.
func init() {
	Register(Spec{
		OracleID:        "f2f4e14d-1ce8-41b3-b907-03ff561b42f7",
		Name:            "The Mary Janes",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(creaturesThatAttackedThisTurn,
				"This spell costs {1} less to cast for each creature that attacked this turn"),
		},
	})
}

// creaturesThatAttackedThisTurn is the count behind "This spell costs {1}
// less to cast for each creature that attacked this turn" (The Mary Janes,
// Rowdy Research): distinct creatures declared as attackers this turn,
// anyone's.
func creaturesThatAttackedThisTurn(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	return q.Game.CreaturesAttackedThisTurn()
}
