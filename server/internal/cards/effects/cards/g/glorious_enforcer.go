package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glorious Enforcer — Creature — Angel {5}{W}{W}, 5/5 (Modern Horizons 2):
//
//	"Flying, lifelink
//	 At the beginning of each combat, if you have more life than an opponent,
//	 this creature gains double strike until end of turn."
//
// Every player's beginning of combat, with an intervening "if" (CR 603.4):
// more life than at least one opponent still in the game, checked as the step
// begins and again as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3dcbbbeb-4efe-4363-bfb4-b7d7f803bafa",
		Name:            "Glorious Enforcer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you have more life than an opponent", effects.OldB034YouHaveMoreLifeThanAnOpponent,
				effects.AtBeginningOfEachCombat("Glorious Enforcer — this creature gains double strike until end of turn",
					effects.ThisGainsKeywordUntilEndOfTurn("double strike", "Glorious Enforcer — double strike until end of turn"))),
		},
	})
}
