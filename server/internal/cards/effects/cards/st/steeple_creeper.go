package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steeple Creeper — Creature — Frog Snake {2}{G}, 4/2 (Ravnica: Clue Edition, #5929): "{3}{U}: This creature gains flying until end of turn."
//
// Killer Whale's shape with a {3}{U} cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f53f33ed-5c8c-4835-bc5f-65a8acfa3f5b",
		Name:         "Steeple Creeper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{3}{U}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Steeple Creeper — flying"),
		}},
	})
}
