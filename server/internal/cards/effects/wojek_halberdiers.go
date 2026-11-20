package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wojek Halberdiers — Creature — Human Soldier {R}{W}, 3/2 (Gatecrash, #5565):
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, this creature gains first strike until end of turn."
//
// Battalion is Boros Strike-Captain's group test (gtcBattalion); the
// bonus is pinned to this creature and ends in the cleanup step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "96d6a400-f1bc-40f8-91ef-1d3cb96926f1",
		Name:         "Wojek Halberdiers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			gtcBattalion("Wojek Halberdiers — it gains first strike until end of turn (battalion)",
				thisGainsKeywordUntilEndOfTurn("first strike", "Wojek Halberdiers — first strike until end of turn")),
		},
	})
}
