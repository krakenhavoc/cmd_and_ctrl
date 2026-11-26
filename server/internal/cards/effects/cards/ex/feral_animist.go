package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feral Animist — Creature — Goblin Shaman {1}{R}{G}, 2/1 (Dragon's Maze, S82
// OLD b406, #5564):
//
//	"{3}: This creature gets +X/+0 until end of turn, where X is its power."
//
// +X/+0 with X its own power is a doubling (CR 701.10b: DoublePowerUntilEOT),
// read as the ability resolves, so a second activation doubles the doubled
// power. An Animist that has left the battlefield gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "beebed34-3a9b-4b51-aacb-1e85d76cf326",
		Name:         "Feral Animist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: This creature gets +X/+0 until end of turn, where X is its power.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DoublePowerUntilEOT(effects.NewContext(g, item), item.SourceCardID, "Feral Animist — +X/+0 where X is its power")
			},
		}},
	})
}
