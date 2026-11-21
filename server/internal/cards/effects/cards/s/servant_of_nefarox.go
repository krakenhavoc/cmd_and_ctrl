package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Servant of Nefarox — Creature — Human Cleric {2}{B}, 3/1 (Magic 2013, #5568):
//
//	"Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)"
//
// Akrasan Squire's exalted (CR 702.83a): the engine derives the trigger
// from the printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8740b9be-3ab4-4bd1-8a89-a59397c59e82",
		Name:            "Servant of Nefarox",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
	})
}
