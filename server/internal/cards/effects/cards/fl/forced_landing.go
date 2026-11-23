package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forced Landing — Instant {1}{G} (Ravnica Remastered):
//
//	"Put target creature with flying on the bottom of its owner's library."
//
// The flying is read through the layered view when the target is chosen and
// again on resolution (CR 608.2b); a creature that lost it is not tucked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a787b18d-d33f-41ea-a0f8-f05e272d06eb",
		Name:         "Forced Landing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return ctx.Game.TuckToLibraryForEffect(t.ID, true)
				}
			}
			return nil
		},
	})
}
