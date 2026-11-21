package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strength of the Pack — Sorcery {4}{G}{G} (Rivals of Ixalan, #5531):
//
//	"Put two +1/+1 counters on each creature you control."
//
// Germination Practicum's body without the paradigm: the creatures are
// read once as the spell resolves, and an opponent's get none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be78a76c-e126-4f09-9e8e-b773a437cfeb",
		Name:         "Strength of the Pack",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B11PutCountersOnEachCreatureYouControl(ctx.Game, item, 2)
		},
	})
}
