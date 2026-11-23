package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fate Forgotten — Instant {2}{W} (Dragons of Tarkir,
// #5551):
//
//	"Exile target artifact or enchantment."
//
// Altar's Light's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f591bfb-a241-4980-9cb3-fb5af3355f6b",
		Name:         "Fate Forgotten",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstTarget(ctx.Game, item)
		},
	})
}
