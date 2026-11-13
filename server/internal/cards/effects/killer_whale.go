package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Killer Whale — Creature — Whale {3}{U}{U}, 3/5 (Exodus, #5626):
//
//	"{U}: This creature gains flying until end of turn."
//
// Vodalian Knights' {U} row through the shared
// thisGainsKeywordUntilEndOfTurn body: a layer-6 grant pinned to the
// whale that ends in the cleanup step (CR 514.2). Activated after
// blockers are declared, it doesn't undo a block already made
// (CR 509.1h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ace3817-5a98-4eee-b6b9-8c105cea302f",
		Name:         "Killer Whale",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{U}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Killer Whale — flying"),
		}},
	})
}
