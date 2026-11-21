package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5ace3817-5a98-4eee-b6b9-8c105cea302f",
		Name:         "Killer Whale",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Killer Whale — flying"),
		}},
	})
}
