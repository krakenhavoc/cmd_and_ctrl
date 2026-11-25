package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exquisite Archangel — Creature — Angel {5}{W}{W}, 5/5 (Aether Revolt, S82
// older sets, #5536):
//
//	"Flying
//	 If you would lose the game, instead exile this creature and your life
//	 total becomes equal to your starting life total."
//
// The Golden Throne's game-action replacement (CR 614.1a) with the total
// read from the table's starting life as the replacement applies (40 in
// Commander): "you" is the Angel's controller, the exile goes through the
// ordinary exile route, and the life change is a gain or a loss of the
// difference (CR 119.5). A second Angel replaces the next loss in turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "472efb69-47c8-4c93-b60a-f6dba4364b8c",
		Name:            "Exquisite Archangel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		GameActionReplacements: []game.GameActionReplacement{
			{Kind: game.ReplaceLoseGameExileThisLifeBecomes, StartingLife: true},
		},
	})
}
