package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Glasswright // Craft with Pride — a preparation card (CR 722, ADR
// 0090), Secrets of Strixhaven (#4456):
//
//	Goblin Glasswright {1}{R}, Creature — Goblin Sorcerer 2/2:
//	"This creature enters prepared. (While it's prepared, you may cast a copy
//	 of its spell. Doing so unprepares it.)"
//	Craft with Pride {R}, Sorcery:
//	"Create a Treasure token."
//
// "Enters prepared" is SelfEntersPrepared; the engine makes, casts and
// unprepares the copy. Craft with Pride is one Treasure for its controller.
//
// No simplification.
func init() {
	const id = "23c3de07-8aa2-48e9-848c-50b5a40bd133"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Goblin Glasswright",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Craft with Pride",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
	})
}
