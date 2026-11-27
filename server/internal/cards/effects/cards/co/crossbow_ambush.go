package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crossbow Ambush — Instant {G} (Stronghold, S82 OLD b749, #5627):
//
//	"Creatures you control gain reach until end of turn."
//
// Aerie Mystics' group grant at reach: the creatures you control as the
// spell resolves gain it (CR 611.2c) until the cleanup step (CR 514.2); a
// creature that arrives later does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c10fd7a6-360f-43d6-a4ef-d990c97cb2f6",
		Name:         "Crossbow Ambush",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"reach"},
				Label:    "Crossbow Ambush — reach",
			}.Apply(ctx)
		},
	})
}
