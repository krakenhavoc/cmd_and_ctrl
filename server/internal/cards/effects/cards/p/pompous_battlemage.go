package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pompous Battlemage // Improvised Act — Creature — Goblin Sorcerer {R},
// 1/1 // Sorcery {R} (preparation card, CR 722):
//
//	"Prowess
//	 This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Improvised Act — "You may discard a card. If you do, draw a card."
//
// No simplification.
func init() {
	const id = "33e3793c-098f-4897-88f3-9f9ce9081e0f"
	const question = "Improvised Act — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:        id,
		Name:            "Pompous Battlemage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Improvised Act",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B39MayDiscardThenDraw(1, false, question, effects.DrawAsManyAsDiscarded)(ctx)
		},
	})
}
