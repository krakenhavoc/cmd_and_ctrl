package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chronostutter — Instant {5}{U} (Magic 2015, #5556):
//
//	"Put target creature into its owner's library second from the top."
//
// Oust's tuck without the life: PutIntoLibrary with Depth 2 (ADR 0088), so a
// commander whose owner declines the command zone still lands second from the
// top.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7c46e5f-6e3a-405b-86fa-404d3b6971db",
		Name:         "Chronostutter",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.PutIntoLibrary{Card: item.Targets[0].ID, Depth: 2}.Apply(ctx)
		},
	})
}
