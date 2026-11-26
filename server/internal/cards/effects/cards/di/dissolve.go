package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dissolve — Instant {1}{U}{U} (Theros):
//
//	"Counter target spell. Scry 1."
//
// Counterspell's counter, then the scry, which happens whenever the spell
// resolves (Jace's Defeat's shape). A target spell that left the stack makes
// the whole spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f0741f9-802b-4761-8094-7b1f3ea1a424",
		Name:         "Dissolve",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.CounterTheTargetSpell(item, ctx); err != nil {
				return err
			}
			return effects.Scry{N: 1}.Apply(ctx)
		},
	})
}
